/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sealdb/neodb/config"

	"github.com/pkg/errors"
	"github.com/sealdb/mysqlstack/xlog"
)

const (
	xacheckJSONFile         = "xacheck.json"
	xacheckTimesOutJSONFile = "xacheck_timesout.json"
)

const (
	txnXACommitErrStateCommit   = "commit"
	txnXACommitErrStateRollback = "rollback"
)

// XaCommitErr tuple.
type XaCommitErr struct {
	Time  string `json:"time"`
	Xaid  string `json:"xaid"`
	State string `json:"state"`
	Times int    `json:"times"`
}

// XaCommitErrs tuple
type XaCommitErrs struct {
	Logs []*XaCommitErr `json:"xacommit-errs"`
}

// XaCheck tuple.
type XaCheck struct {
	log     *xlog.Log
	dir     string
	times   int
	scatter *Scatter
	retrys  map[string]*XaCommitErr
	done    chan bool
	ticker  *time.Ticker
	wg      sync.WaitGroup
	mu      sync.RWMutex
}

// NewXaCheck creates the XaCheck tuple.
func NewXaCheck(scatter *Scatter, conf *config.ScatterConfig) *XaCheck {
	return &XaCheck{
		log:     scatter.log,
		dir:     conf.XaCheckDir,
		times:   conf.XaCheckRetrys,
		scatter: scatter,
		retrys:  make(map[string]*XaCommitErr),
		done:    make(chan bool),
		ticker:  time.NewTicker(time.Duration(time.Second * time.Duration(conf.XaCheckInterval))),
	}
}

// Init used to init xa check goroutine.
func (xc *XaCheck) Init() error {
	log := xc.log

	// If the xc.dir is already a directory, MkdirAll does nothing
	// if the dir is one file, return err
	if err := os.MkdirAll(xc.dir, 0744); err != nil {
		return err
	}

	if err := xc.LoadXaCommitErrLogs(); err != nil {
		return err
	}

	xc.wg.Add(1)
	go func(dc *XaCheck) {
		defer dc.wg.Done()
		dc.xaCommitCheck()
	}(xc)

	log.Info("xacheck.init.done")
	return nil
}

func writeAppendFile(file string, data []byte) error {
	flag := os.O_RDWR | os.O_APPEND
	if _, err := os.Stat(file); os.IsNotExist(err) {
		flag |= os.O_CREATE
	}
	f, err := os.OpenFile(file, flag, 0644)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	n, err := f.Write(data)
	if err != nil {
		return errors.WithStack(err)
	}
	if n != len(data) {
		return errors.WithStack(io.ErrShortWrite)
	}
	return f.Sync()
}

func (xc *XaCheck) addXaCommitErrLog(new *XaCommitErr) error {
	log := xc.log
	log.Info("xc.addXaCommitErrLog.add:+%v", new)

	if _, ok := xc.retrys[new.Xaid]; ok {
		log.Error("xacheck.addXACommitErrLog.xaid[%v].duplicate", new)
		return errors.Errorf("xacheck.addXACommitErrLog.xaid[%v].duplicate", new.Xaid)
	}

	xc.retrys[new.Xaid] = new
	return nil
}

// flushXaCommitErrLog is used to write the xaCommitErrlogs to the file.
func (xc *XaCheck) flushXaCommitErrLog() error {
	log := xc.log
	file := path.Join(xc.dir, xacheckJSONFile)

	// stored in the way of the array
	var xaCommitErrs XaCommitErrs
	for _, v := range xc.retrys {
		xaCommitErrs.Logs = append(xaCommitErrs.Logs, v)
	}

	log.Info("xacheck.flush.to.file[%v].xaCommitErrs:%+v", file, xaCommitErrs)
	if err := config.WriteConfig(file, xaCommitErrs); err != nil {
		log.Panicf("xacheck.flush.config.to.file[%v].error:%v", file, err)
		return err
	}
	return nil
}

// WriteXaCommitErrLog is used to write the xaCommitErrLog into the xacheck file.
func (xc *XaCheck) WriteXaCommitErrLog(txn *Txn, state string) error {
	xaCommitErr := &XaCommitErr{
		Time:  time.Now().Format("20060102150405"),
		Xaid:  txn.xid,
		State: state,
		Times: xc.times,
	}

	xc.mu.Lock()
	defer xc.mu.Unlock()
	// add the xaCommitErrLog to xacheck
	if err := xc.addXaCommitErrLog(xaCommitErr); err != nil {
		return errors.WithStack(err)
	}

	// TODO: if the NeoDB crash at the moment.
	// flush the mem to the file
	if err := xc.flushXaCommitErrLog(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// commitRetryBackends
// 1st. xa recover to
//  1. check all backends are ready, or else return err
//  2. check the xid is valid to some backends which need the 2nd command, and set needCommitBackends
//
// 2nd. xa commit/rollback 'xaid' to the Backends which need commit
func (xc *XaCheck) commitRetryBackends(retry *XaCommitErr, scatter *Scatter) (bool, error) {
	backends := scatter.AllBackends()
	log := xc.log

	// if the backend is empty, output error log.
	if len(backends) == 0 {
		log.Error("xacheck.commitRetryBackends.backend.empty.")
		return false, errors.New("xacheck.backend.empty")
	}

	txn, err := scatter.CreateTransaction()
	if err != nil {
		log.Error("xacheck.commitRetryBackends.create.transaction.error:[%v]", err)
		return false, err
	}
	defer txn.Finish()

	query := fmt.Sprintf("xa %s '%s' ", retry.State, retry.Xaid)
	// the 1st stage: xa recover
	xaRecoverQuery := "xa recover"
	var needCommitBackends []string
	// if one backend return the err, the backend may not be ready, will return,
	// because all backends are ready，the needCommitBackends is valuable, or else it is misleading.
	for _, backend := range backends {
		result, err := txn.ExecuteOnThisBackend(backend, xaRecoverQuery)
		if err != nil {
			log.Warning("xacheck.commitRetryBackends.xa.recover.error:[%v]", err)
			return false, err
		}

		if result != nil && result.RowsAffected > 0 && len(result.Fields) == 4 {
			for _, row := range result.Rows {
				// just find the xaid in the row from the cmd of 'xa recover'
				valStr := string(row[3].Raw())
				if strings.EqualFold(valStr, retry.Xaid) {
					log.Info("xacheck.commitRetryBackends.recover.query[%v].needCommitBackend[%v]", query, backend)
					needCommitBackends = append(needCommitBackends, backend)
				}
			}
		}
	}

	if len(needCommitBackends) == 0 {
		log.Info("xacheck.commitRetryBackends.recover.query[%v].find.no.Backends.need.retry.[%v]", query, retry.Times)
	}

	// the 2nd stage: xa commit/rollback '$xid'
	count := 0
	for _, backend := range needCommitBackends {
		_, err = txn.ExecuteOnThisBackend(backend, query)
		if err == nil {
			log.Info("xacheck.commitRetryBackends.query[%v].success.backend[%v]", query, backend)
			count++
		} else {
			log.Warning("xacheck.commitRetryBackends.query[%v].backend[%v].error[%T]:%+v", query, backend, err, err)
		}
	}

	// if there are XA need to retry, and the count committed is equal to the needCommitBackends, return true.
	if count > 0 && count == len(needCommitBackends) {
		return true, nil
	}

	if retry.Times <= 0 {
		log.Warning("xacheck.commitRetryBackends.query[%v].retry.times.out[%v]", query, retry.Times)
		data, err := json.Marshal(retry)
		if err != nil {
			log.Error("xacheck.Marshal.error:[%v]", err)
			return false, err
		}

		file := path.Join(xc.dir, xacheckTimesOutJSONFile)
		if err := writeAppendFile(file, data); err != nil {
			log.Panicf("xacheck.flush.config.to.file[%v].error:%v", file, err)
			return false, err
		}
		return true, nil
	} else {
		retry.Times--
	}
	return false, nil
}

// the retrys map maybe subtract the retry which is committed when all backends are ready
// or has retried for the number of times but unsuccessfully.
func (xc *XaCheck) xaCommitsRetryMain() error {
	log := xc.log
	retrys := xc.retrys
	if xc.getRetrysLen() > 0 {
		log.Info("xacheck.commit.retry %v.", retrys)
	}

	for _, retry := range retrys {
		committedOrTimesout, err := xc.commitRetryBackends(retry, xc.scatter)
		if err != nil {
			log.Warning("xacheck.commits.retry failed.")
			return err
		}

		if committedOrTimesout {
			// every retry is committed, update the mem and flush to the file
			delete(xc.retrys, retry.Xaid)
			if err := xc.flushXaCommitErrLog(); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

func (xc *XaCheck) xaCommitsRetry() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()

	// xaCommitsRetryMain
	if err := xc.xaCommitsRetryMain(); err != nil {
		return err
	}
	return nil
}

func (xc *XaCheck) xaCommitCheck() {
	defer xc.ticker.Stop()
	for {
		select {
		case <-xc.ticker.C:
			xc.xaCommitsRetry()
		case <-xc.done:
			return
		}
	}
}

// ReadXaCommitErrLogs is used to read the Xaredologs config from the data.
func (xc *XaCheck) ReadXaCommitErrLogs(data string) (*XaCommitErrs, error) {
	s := &XaCommitErrs{}
	if err := json.Unmarshal([]byte(data), s); err != nil {
		return nil, errors.WithStack(err)
	}

	return s, nil
}

// LoadXaCommitErrLogs is used to load all XaCommitErr from metadir/xacheck.json file.
func (xc *XaCheck) LoadXaCommitErrLogs() error {
	log := xc.log
	metadir := xc.dir
	file := path.Join(metadir, xacheckJSONFile)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		// not Creating it if the xacheck log doesn't exist.
		// to avoid creating the empty file xaredolog.json about "xaredologs": null
		// the xaredolog.json will be created when the xaredolog are generated.
		return nil
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error("xacheck.LoadXaCommitErrLogs.readfile[%v].error:%v", file, err)
		return err
	}

	retrys, err := xc.ReadXaCommitErrLogs(string(data))
	if err != nil {
		log.Error("xacheck.LoadXaCommitErrLogs.readfile.to.xacheck[%v].error:%v", file, err)
		return err
	}

	for _, new := range retrys.Logs {
		if err := xc.addXaCommitErrLog(new); err != nil {
			return err
		}

		log.Info("xacheck.load.xaid:%+v", new.Xaid)
	}
	return nil
}

// Close is used to close the xacheck goroutine
func (xc *XaCheck) Close() {
	close(xc.done)
	xc.wg.Wait()
}

// GetXaCheckFile get the XaCheck log file
func (xc *XaCheck) GetXaCheckFile() string {
	return path.Join(xc.dir, xacheckJSONFile)
}

// RemoveXaCommitErrLogs is only used to test to avoid the noise,
// XaCommitErrLogs can not be removed in the production environment, it is so important.
func (xc *XaCheck) RemoveXaCommitErrLogs() error {
	return os.RemoveAll(xc.dir)
}

func (xc *XaCheck) getRetrysLen() int {
	return len(xc.retrys)
}

// GetRetrysLen return the retrys num
func (xc *XaCheck) GetRetrysLen() int {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	return len(xc.retrys)
}
