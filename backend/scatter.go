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
	"io/ioutil"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/sealdb/neodb/config"
	"github.com/sealdb/neodb/monitor"

	"github.com/pkg/errors"
	"github.com/sealdb/mysqlstack/xlog"
)

const (
	backendjson = "backend.json"
)

// Scatter tuple.
type Scatter struct {
	log      *xlog.Log
	mu       sync.RWMutex
	txnMgr   *TxnManager
	metadir  string
	backends map[string]*Poolz
}

// NewScatter creates a new scatter.
func NewScatter(log *xlog.Log, metadir string) *Scatter {
	return &Scatter{
		log:      log,
		txnMgr:   NewTxnManager(log),
		metadir:  metadir,
		backends: make(map[string]*Poolz),
	}
}

// Init is used to init the xaCheck and start the xaCheck thread.
func (scatter *Scatter) Init(scatterConf *config.ScatterConfig) error {
	return scatter.txnMgr.Init(scatter, scatterConf)
}

// Add backend node.
func (scatter *Scatter) add(config *config.BackendConfig) error {
	log := scatter.log
	log.Info("scatter.add:%v", config.Name)

	if _, ok := scatter.backends[config.Name]; ok {
		return errors.Errorf("scatter.backend[%v].duplicate", config.Name)
	}
	// check the address in backendsConfig, fix issue#546.
	for _, pool := range scatter.backends {
		if pool.conf.Address == config.Address {
			return errors.Errorf("scatter.address[%v].already.exists.in.backends", config.Address)
		}
	}

	scatter.backends[config.Name] = NewPoolz(log, config)
	monitor.BackendInc("backend")
	return nil
}

// Add used to add a new backend to scatter.
func (scatter *Scatter) Add(config *config.BackendConfig) error {
	scatter.mu.Lock()
	defer scatter.mu.Unlock()
	return scatter.add(config)
}

func (scatter *Scatter) remove(config *config.BackendConfig) error {
	log := scatter.log
	log.Warning("scatter.remove:%v", config.Name)

	pool, ok := scatter.backends[config.Name]
	if !ok {
		return errors.Errorf("scatter.backend[%v].can.not.be.found", config.Name)
	}
	delete(scatter.backends, config.Name)
	monitor.BackendDec("backend")
	pool.Close()
	return nil
}

// Remove used to remove a backend from the scatter.
func (scatter *Scatter) Remove(config *config.BackendConfig) error {
	scatter.mu.Lock()
	defer scatter.mu.Unlock()
	return scatter.remove(config)
}

// Close used to clean the pools connections.
func (scatter *Scatter) Close() {
	scatter.mu.Lock()
	defer scatter.mu.Unlock()

	log := scatter.log
	log.Info("scatter.prepare.to.close....")
	scatter.clear()
	// to close the xaCheck go
	scatter.txnMgr.Close()
	log.Info("scatter.close.done....")
}

func (scatter *Scatter) clear() {
	for _, v := range scatter.backends {
		v.Close()
	}
	scatter.backends = make(map[string]*Poolz)
}

// FlushConfig used to write the backends to file.
func (scatter *Scatter) FlushConfig() error {
	scatter.mu.Lock()
	defer scatter.mu.Unlock()

	log := scatter.log
	file := path.Join(scatter.metadir, backendjson)

	var backends config.BackendsConfig
	for _, v := range scatter.backends {
		backends.Backends = append(backends.Backends, v.conf)
	}

	log.Warning("scatter.flush.to.file[%v].backends.conf:%+v", file, backends.Backends)
	if err := config.WriteConfig(file, backends); err != nil {
		log.Panicf("scatter.flush.config.to.file[%v].error:%v", file, err)
		return err
	}
	if err := config.UpdateVersion(scatter.metadir); err != nil {
		log.Panicf("scatter.flush.config.update.version.error:%v", err)
		return err
	}
	return nil
}

// LoadConfig used to load all backends from metadir/backend.json file.
func (scatter *Scatter) LoadConfig() error {
	scatter.mu.Lock()
	defer scatter.mu.Unlock()

	// Do clear first.
	scatter.clear()

	log := scatter.log
	metadir := scatter.metadir
	file := path.Join(metadir, backendjson)

	// Create it if the backends config not exists.
	if _, err := os.Stat(file); os.IsNotExist(err) {
		backends := config.BackendsConfig{}
		if err := config.WriteConfig(file, backends); err != nil {
			log.Error("scatter.flush.backends.to.file[%v].error:%v", file, err)
			return err
		}
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error("scatter.load.from.file[%v].error:%v", file, err)
		return err
	}
	conf, err := config.ReadBackendsConfig(string(data))
	if err != nil {
		log.Error("scatter.parse.json.file[%v].error:%v", file, err)
		return err
	}
	monitor.BackendSet("backend", 0)
	for _, backend := range conf.Backends {
		if err := scatter.add(backend); err != nil {
			log.Error("scatter.add.backend[%+v].error:%v", backend.Name, err)
			return err
		}
		log.Info("scatter.load.backend:%+v", backend.Name)
	}
	return nil
}

// AllBackends returns all backends.
func (scatter *Scatter) AllBackends() []string {
	var backends []string
	scatter.mu.RLock()
	defer scatter.mu.RUnlock()
	for k := range scatter.backends {
		backends = append(backends, k)
	}
	sort.Strings(backends)
	return backends
}

// Backends returns all normal backends.
func (scatter *Scatter) Backends() []string {
	var backends []string
	scatter.mu.RLock()
	defer scatter.mu.RUnlock()
	for k, pool := range scatter.backends {
		if pool.conf.Role != config.NormalBackend {
			continue
		}

		backends = append(backends, k)
	}
	sort.Strings(backends)
	return backends
}

func (scatter *Scatter) CheckBackend(backenName string) bool {
	scatter.mu.RLock()
	defer scatter.mu.RUnlock()
	for k, pool := range scatter.backends {
		if pool.conf.Role != config.NormalBackend {
			continue
		}

		if k == backenName {
			return true
		}
	}
	return false
}

// PoolzClone used to copy backends to new map.
func (scatter *Scatter) PoolzClone() map[string]*Poolz {
	poolzMap := make(map[string]*Poolz)
	scatter.mu.RLock()
	defer scatter.mu.RUnlock()
	for k, v := range scatter.backends {
		poolzMap[k] = v
	}
	return poolzMap
}

// BackendConfigsClone used to clone all the backend configs.
func (scatter *Scatter) BackendConfigsClone() []*config.BackendConfig {
	scatter.mu.RLock()
	defer scatter.mu.RUnlock()
	beConfigs := make([]*config.BackendConfig, 0, 16)
	for _, v := range scatter.backends {
		beConfigs = append(beConfigs, v.conf)
	}
	return beConfigs
}

// CreateTransaction used to create a transaction.
func (scatter *Scatter) CreateTransaction() (*Txn, error) {
	return scatter.txnMgr.CreateTxn(scatter.PoolzClone())
}
