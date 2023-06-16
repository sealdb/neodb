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
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sealdb/neodb/config"
	"github.com/sealdb/neodb/fakedb"
	"github.com/sealdb/neodb/xcontext"

	"github.com/fortytw2/leaktest"
	"github.com/sealdb/mysqlstack/sqldb"
	querypb "github.com/sealdb/mysqlstack/sqlparser/depends/query"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

var (
	xaRecoverResult1 = &sqltypes.Result{
		RowsAffected: 1,
		Fields: []*querypb.Field{
			{
				Name: "formatID",
				Type: querypb.Type_INT64,
			},
			{
				Name: "gtrid_length",
				Type: querypb.Type_INT64,
			},
			{
				Name: "bqual_length",
				Type: querypb.Type_INT64,
			},
			{
				Name: "data",
				Type: querypb.Type_VARCHAR,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.MakeTrusted(querypb.Type_INT64, []byte("1")),
				sqltypes.MakeTrusted(querypb.Type_INT64, []byte("21")),
				sqltypes.MakeTrusted(querypb.Type_INT64, []byte("0")),
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("RXID-20180903103145-1")),
			},
		},
	}

	xaRecoverResult2 = &sqltypes.Result{
		RowsAffected: 0,
		Fields: []*querypb.Field{
			{
				Name: "formatID",
				Type: querypb.Type_INT64,
			},
			{
				Name: "gtrid_length",
				Type: querypb.Type_INT64,
			},
			{
				Name: "bqual_length",
				Type: querypb.Type_INT64,
			},
			{
				Name: "data",
				Type: querypb.Type_VARCHAR,
			},
		},
	}

	xaRecoverResult3 = &sqltypes.Result{
		RowsAffected: 1,
		Fields: []*querypb.Field{
			{
				Name: "formatID",
				Type: querypb.Type_INT64,
			},
			{
				Name: "gtrid_length",
				Type: querypb.Type_INT64,
			},
			{
				Name: "bqual_length",
				Type: querypb.Type_INT64,
			},
			{
				Name: "data",
				Type: querypb.Type_VARCHAR,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.MakeTrusted(querypb.Type_INT64, []byte("1")),
				sqltypes.MakeTrusted(querypb.Type_INT64, []byte("21")),
				sqltypes.MakeTrusted(querypb.Type_INT64, []byte("0")),
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("RXID-20180903103146-1")),
			},
		},
	}
)

func TestWriteXaCommitErrorLogsAddXidDuplicate(t *testing.T) {

	defer leaktest.Check(t)()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	scatter, _, cleanup := MockScatter(log, 2)
	defer cleanup()

	err := scatter.Init(MockScatterDefault(log))
	assert.Nil(t, err)
	txn1, err := scatter.CreateTransaction()
	assert.Nil(t, err)
	defer txn1.Finish()
	txn1.xid = "RXID-20180903103145-1"
	state := "rollback"
	err = scatter.txnMgr.xaCheck.WriteXaCommitErrLog(txn1, state)
	assert.Nil(t, err)

	//the same xid, it is error
	txn2, err := scatter.CreateTransaction()
	assert.Nil(t, err)
	defer txn2.Finish()
	txn2.xid = "RXID-20180903103145-1"
	state = "commit"
	err = scatter.txnMgr.xaCheck.WriteXaCommitErrLog(txn2, state)
	assert.NotNil(t, err)

	//add another errlog
	txn3, err := scatter.CreateTransaction()
	assert.Nil(t, err)
	defer txn3.Finish()
	txn3.xid = "RXID-20180903103146-1"
	state = "commit"
	err = scatter.txnMgr.xaCheck.WriteXaCommitErrLog(txn3, state)
	assert.Nil(t, err)
}

func TestReadXaCommitErrorLogsWithoutBackend(t *testing.T) {
	defer leaktest.Check(t)()
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	data := `{
    "xacommit-errs": [
        {
            "time": "20180903103145",
            "xaid": "RXID-20180903103145-1",
            "state": "commit",
            "times": 10
        }
    ]
}`

	dir := fakedb.GetTmpDir("/tmp", "xacheck", log)
	file := path.Join(dir, xacheckJSONFile)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	assert.Nil(t, err)
	defer os.RemoveAll(file)

	scatter := NewScatter(log, "")
	err = scatter.Init(MockScatterDefault2(dir))
	assert.Nil(t, err)

	time.Sleep(1 * time.Second)

	scatter.txnMgr.xaCheck.Close()
}

func TestTxnTwoPCExecuteCommitError(t *testing.T) {
	defer leaktest.Check(t)()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	scatter, fakedb1, cleanup1 := MockScatter(log, 2)
	defer cleanup1()

	err := scatter.Init(MockScatterDefault(log))
	assert.Nil(t, err)
	var backend [2]string
	var i int
	for k := range scatter.backends {
		backend[i] = k
		i++
	}

	querys1 := []xcontext.QueryTuple{
		xcontext.QueryTuple{Query: "update", Backend: backend[0]},
		xcontext.QueryTuple{Query: "update", Backend: backend[1]},
	}

	fakedb1.AddQuery(querys1[0].Query, result1)
	fakedb1.AddQueryDelay(querys1[1].Query, result2, 100)

	// Set 2PC conds.
	resetFunc1 := func(txn *Txn) {
		fakedb1.ResetAll()
		fakedb1.AddQuery(querys1[0].Query, result1)
		fakedb1.AddQueryDelay(querys1[1].Query, result2, 100)
		fakedb1.AddQueryPattern("XA .*", result1)
	}

	// XA COMMIT error.
	{
		txn, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn.Finish()

		resetFunc1(txn)
		fakedb1.AddQueryErrorPattern("XA COMMIT .*", sqldb.NewSQLError1(1397, "XAE04", "XAER_NOTA: Unknown XID"))

		err = txn.Begin()
		assert.Nil(t, err)

		rctx := &xcontext.RequestContext{
			Mode:    xcontext.ReqNormal,
			TxnMode: xcontext.TxnWrite,
			Querys:  querys1,
		}
		_, err = txn.Execute(rctx)
		assert.Nil(t, err)
		err = txn.Commit()
		assert.Nil(t, err)
		time.Sleep(2 * time.Second)
	}

	_, err = os.Stat(scatter.txnMgr.xaCheck.GetXaCheckFile())
	assert.Nil(t, err)

	err = scatter.txnMgr.xaCheck.RemoveXaCommitErrLogs()
	assert.Nil(t, err)
}

// the command XA RECOVER is not reponsed
func TestXaCheckFailed(t *testing.T) {

	defer leaktest.Check(t)()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	scatter, fakedb1, cleanup1 := MockScatter(log, 2)
	defer cleanup1()

	err := scatter.Init(MockScatterDefault(log))
	assert.Nil(t, err)
	var backend [2]string
	var i int
	for k := range scatter.backends {
		backend[i] = k
		i++
	}

	querys1 := []xcontext.QueryTuple{
		xcontext.QueryTuple{Query: "update", Backend: backend[0]},
		xcontext.QueryTuple{Query: "update", Backend: backend[1]},
	}

	fakedb1.AddQuery(querys1[0].Query, result1)
	fakedb1.AddQueryDelay(querys1[1].Query, result2, 100)

	// Set 2PC conds.
	resetFunc1 := func(txn *Txn) {
		fakedb1.ResetAll()
		fakedb1.AddQuery(querys1[0].Query, result1)
		fakedb1.AddQueryDelay(querys1[1].Query, result2, 100)
		fakedb1.AddQueryPattern("XA .*", result1)
	}

	// XA COMMIT error.
	{
		txn, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn.Finish()

		resetFunc1(txn)
		fakedb1.AddQueryErrorPattern("XA COMMIT .*", sqldb.NewSQLError1(1397, "XAE04", "XAER_NOTA: Unknown XID"))

		err = txn.Begin()
		assert.Nil(t, err)

		rctx := &xcontext.RequestContext{
			Mode:    xcontext.ReqNormal,
			TxnMode: xcontext.TxnWrite,
			Querys:  querys1,
		}
		_, err = txn.Execute(rctx)
		assert.Nil(t, err)
		err = txn.Commit()
		assert.Nil(t, err)
		//time.Sleep(2 * time.Second)
	}

	// XA COMMIT ok.
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn2.Finish()

		resetFunc1(txn2)
		fakedb1.AddQuery("XA COMMIT .*", result1)

		err = txn2.Begin()
		assert.Nil(t, err)

		rctx := &xcontext.RequestContext{
			Mode:    xcontext.ReqNormal,
			TxnMode: xcontext.TxnWrite,
			Querys:  querys1,
		}
		_, err = txn2.Execute(rctx)
		assert.Nil(t, err)
		err = txn2.Commit()
		assert.Nil(t, err)

		time.Sleep(2 * time.Second)
	}

	// the xacheck is stil exit.
	_, err = os.Stat(scatter.txnMgr.xaCheck.GetXaCheckFile())
	assert.Nil(t, err)
}

func TestXaCheckFromFileCommit(t *testing.T) {
	defer leaktest.Check(t)()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	data := `{
    "xacommit-errs": [
        {
            "time": "20180903103145",
            "xaid": "RXID-20180903103145-1",
            "state": "commit",
            "times": 10
        }
    ]
}`
	dir := fakedb.GetTmpDir("/tmp", "xacheck", log)
	file := path.Join(dir, xacheckJSONFile)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	assert.Nil(t, err)
	// it is no need to remove the file, if it is successful, err item in the file will be removed
	//defer os.RemoveAll(file)

	scatter, fakedb1, cleanup1 := MockScatter(log, 2)
	defer cleanup1()

	err = scatter.Init(MockScatterDefault2(dir))
	assert.Nil(t, err)
	var backend [2]string
	var i int
	for k := range scatter.backends {
		backend[i] = k
		i++
	}

	querys1 := []xcontext.QueryTuple{
		xcontext.QueryTuple{Query: "update", Backend: backend[0]},
		xcontext.QueryTuple{Query: "update", Backend: backend[1]},
	}

	fakedb1.AddQuery(querys1[0].Query, result1)
	fakedb1.AddQueryDelay(querys1[1].Query, result2, 100)

	// Set 2PC conds.
	resetFunc1 := func(txn *Txn) {
		fakedb1.ResetAll()
		fakedb1.AddQuery(querys1[0].Query, result1)
		fakedb1.AddQueryDelay(querys1[1].Query, result2, 100)
		fakedb1.AddQueryPattern("XA .*", result1)
	}

	time.Sleep(1 * time.Second)

	// XA RECOVER and return err
	// XA COMMIT ok
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn2.Finish()

		resetFunc1(txn2)
		fakedb1.AddQueryErrorPattern("XA RECOVER", errors.New("Server maybe lost, please try again"))
		fakedb1.AddQuery("XA COMMIT .*", result1)

		time.Sleep(2 * time.Second)
	}

	// XA RECOVER and return Empty set
	// XA COMMIT ok
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn2.Finish()

		resetFunc1(txn2)
		fakedb1.AddQuery("XA RECOVER", xaRecoverResult2)
		fakedb1.AddQuery("XA COMMIT .*", result1)

		time.Sleep(1 * time.Second)
	}

	// if the commit failed, It will cost long time related to xaMaxRetryNum to wait
	// omit the case
	// XA RECOVER ok
	// XA COMMIT failed

	// XA RECOVER ok
	// XA COMMIT ok
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn2.Finish()

		resetFunc1(txn2)
		fakedb1.AddQuery("XA RECOVER", xaRecoverResult1)
		fakedb1.AddQuery("XA COMMIT .*", result1)

		time.Sleep(1 * time.Second)
	}
}

func TestXaCheckFromFileRollback(t *testing.T) {
	defer leaktest.Check(t)()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	data := `{
    "xacommit-errs": [
        {
            "time": "20180903103145",
            "xaid": "RXID-20180903103145-1",
            "state": "rollback",
            "times": 10
        }
    ]
}`
	dir := fakedb.GetTmpDir("/tmp", "xacheck", log)
	file := path.Join(dir, xacheckJSONFile)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	assert.Nil(t, err)
	// it is no need to remove the file, if it is successful, err item in the file will be removed
	//defer os.RemoveAll(file)

	scatter, fakedb1, cleanup1 := MockScatter(log, 2)
	defer cleanup1()

	err = scatter.Init(MockScatterDefault2(dir))
	assert.Nil(t, err)
	var backend [2]string
	var i int
	for k := range scatter.backends {
		backend[i] = k
		i++
	}

	querys1 := []xcontext.QueryTuple{
		xcontext.QueryTuple{Query: "update", Backend: backend[0]},
		xcontext.QueryTuple{Query: "update", Backend: backend[1]},
	}

	fakedb1.AddQuery(querys1[0].Query, result1)
	fakedb1.AddQueryDelay(querys1[1].Query, result2, 100)

	// Set 2PC conds.
	resetFunc1 := func(txn *Txn) {
		fakedb1.ResetAll()
		fakedb1.AddQuery(querys1[0].Query, result1)
		fakedb1.AddQueryDelay(querys1[1].Query, result2, 100)
		fakedb1.AddQueryPattern("XA .*", result1)
	}

	// XA RECOVER and return err
	// XA ROLLBACK ok
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn2.Finish()

		resetFunc1(txn2)
		fakedb1.AddQueryErrorPattern("XA RECOVER", errors.New("Server maybe lost, please try again"))
		fakedb1.AddQuery("XA ROLLBACK .*", result1)

		time.Sleep(1 * time.Second)
	}

	// XA RECOVER and return Empty set
	// XA ROLLBACK ok
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn2.Finish()

		resetFunc1(txn2)
		fakedb1.AddQuery("XA RECOVER", xaRecoverResult2)
		fakedb1.AddQuery("XA ROLLBACK .*", result1)

		time.Sleep(2 * time.Second)
	}

	// XA RECOVER ok
	// XA ROLLBACK failed
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn2.Finish()

		resetFunc1(txn2)
		fakedb1.AddQuery("XA RECOVER", xaRecoverResult1)
		fakedb1.AddQueryErrorPattern("XA ROLLBACK .*", errors.New("Server maybe lost, please try again"))

		time.Sleep(1 * time.Second)
	}

	// XA RECOVER ok
	// XA ROLLBACK ok
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn2.Finish()

		resetFunc1(txn2)
		fakedb1.AddQuery("XA RECOVER", xaRecoverResult1)
		fakedb1.AddQuery("XA ROLLBACK .*", result1)

		time.Sleep(2 * time.Second)
	}
}

func TestXaCheckFromFileCommitRollbacks(t *testing.T) {
	defer leaktest.Check(t)()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	data := `{
    "xacommit-errs": [
        {
            "time": "20180903103145",
            "xaid": "RXID-20180903103145-1",
            "state": "commit",
            "times": 10
        },
        {
            "time": "20180903103146",
            "xaid": "RXID-20180903103146-1",
            "state": "rollback",
            "times": 10
        }
    ]
}`
	dir := fakedb.GetTmpDir("/tmp", "xacheck", log)
	file := path.Join(dir, xacheckJSONFile)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	assert.Nil(t, err)
	// it is no need to remove the file, if it is successful, err item in the file will be removed
	//defer os.RemoveAll(file)

	scatter, fakedb1, cleanup1 := MockScatter(log, 2)
	defer cleanup1()

	err = scatter.Init(MockScatterDefault2(dir))
	assert.Nil(t, err)
	var backend [2]string
	var i int
	for k := range scatter.backends {
		backend[i] = k
		i++
	}

	querys1 := []xcontext.QueryTuple{
		xcontext.QueryTuple{Query: "update", Backend: backend[0]},
		xcontext.QueryTuple{Query: "update", Backend: backend[1]},
	}

	fakedb1.AddQuery(querys1[0].Query, result1)
	fakedb1.AddQueryDelay(querys1[1].Query, result2, 100)

	// Set 2PC conds.
	resetFunc1 := func(txn *Txn) {
		fakedb1.ResetAll()
		fakedb1.AddQuery(querys1[0].Query, result1)
		fakedb1.AddQueryDelay(querys1[1].Query, result2, 100)
		fakedb1.AddQueryPattern("XA .*", result1)
	}

	// XA RECOVER ok
	// XA COMMIT ok
	{
		txn1, err := scatter.CreateTransaction()
		assert.Nil(t, err)

		resetFunc1(txn1)
		fakedb1.AddQuery("XA RECOVER", xaRecoverResult1)
		fakedb1.AddQuery("XA COMMIT .*", result1)

		err = txn1.Begin()
		assert.Nil(t, err)

		rctx := &xcontext.RequestContext{
			Mode:    xcontext.ReqNormal,
			TxnMode: xcontext.TxnWrite,
			Querys:  querys1,
		}
		_, err = txn1.Execute(rctx)
		assert.Nil(t, err)
		err = txn1.Commit()
		assert.Nil(t, err)

		txn1.Finish()

		time.Sleep(1 * time.Second)
	}

	// XA RECOVER and return err
	// XA ROLLBACK ok
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)

		resetFunc1(txn2)
		fakedb1.AddQueryError("XA RECOVER", errors.New("connection.is.lost"))
		//fakedb1.AddQuery("XA RECOVER", xaRecoverResult2)
		fakedb1.AddQuery("XA ROLLBACK .*", result1)

		err = txn2.Begin()
		assert.Nil(t, err)

		rctx := &xcontext.RequestContext{
			Mode:    xcontext.ReqNormal,
			TxnMode: xcontext.TxnWrite,
			Querys:  querys1,
		}
		_, err = txn2.Execute(rctx)
		assert.Nil(t, err)
		err = txn2.Commit()
		assert.Nil(t, err)

		txn2.Finish()

		time.Sleep(1 * time.Second)
	}

	// XA RECOVER and return nil
	// XA ROLLBACK ok
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)

		resetFunc1(txn2)

		txn2.Finish()

		time.Sleep(1 * time.Second)
	}

	// XA RECOVER ok
	// XA ROLLBACK ok
	{
		//assert.Equal(t, 1, scatter.txnMgr.xaCheck.GetRetrysLen())

		txn3, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn3.Finish()

		resetFunc1(txn3)
		fakedb1.AddQuery("XA RECOVER", xaRecoverResult3)
		fakedb1.AddQuery("XA ROLLBACK .*", result1)

		err = txn3.Begin()
		assert.Nil(t, err)

		rctx := &xcontext.RequestContext{
			Mode:    xcontext.ReqNormal,
			TxnMode: xcontext.TxnWrite,
			Querys:  querys1,
		}
		_, err = txn3.Execute(rctx)
		assert.Nil(t, err)
		err = txn3.Commit()
		assert.Nil(t, err)

		time.Sleep(2 * time.Second)

		assert.Equal(t, 0, scatter.txnMgr.xaCheck.GetRetrysLen())
	}
}

func TestLoadXaCommitRrrLogs(t *testing.T) {
	defer leaktest.Check(t)()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	scatter := NewScatter(log, "")

	xaChecker := NewXaCheck(scatter, MockScatterDefault(log))
	defer os.RemoveAll(xaChecker.dir)
	err := xaChecker.Init()
	assert.Nil(t, err)

	err = xaChecker.LoadXaCommitErrLogs()
	assert.Nil(t, err)

	xaChecker.Close()
}

func TestReadXaCommitRrrLogs1(t *testing.T) {
	defer leaktest.Check(t)()

	data := `{
    "xacommit-errs": [
        {
            "time": "20180903103145",
            "xaid": "RXID-20180903103145-1",
            "state": "rollback",
            "times": 10
        }
    ]
}`

	MockXaredologs := []*XaCommitErr{
		&XaCommitErr{
			Time:  "20180903103145",
			Xaid:  "RXID-20180903103145-1",
			State: txnXACommitErrStateRollback,
			Times: 10,
		},
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	scatter := NewScatter(log, "")

	xaChecker := NewXaCheck(scatter, MockScatterDefault(log))
	err := xaChecker.Init()
	assert.Nil(t, err)

	xaCommitErrLogs, err := xaChecker.ReadXaCommitErrLogs(string(data))
	assert.Nil(t, err)
	want := &XaCommitErrs{Logs: MockXaredologs}
	got := xaCommitErrLogs
	assert.Equal(t, want, got)
	xaChecker.Close()
	err = xaChecker.RemoveXaCommitErrLogs()
	assert.Nil(t, err)
}

func TestReadXaCommitRrrLogsError2(t *testing.T) {
	defer leaktest.Check(t)()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	scatter := NewScatter(log, "")

	xaChecker := NewXaCheck(scatter, MockScatterDefault(log))

	err := xaChecker.Init()
	assert.Nil(t, err)

	data1 := `{
    "xacommit-errs": [
		2
    ]
}`

	file := path.Join(xaChecker.dir, xacheckJSONFile)
	err = ioutil.WriteFile(file, []byte(data1), 0644)
	assert.Nil(t, err)
	defer os.RemoveAll(file)

	err = xaChecker.Init()
	assert.NotNil(t, err)

	xaChecker.Close()
}

func TestReadXaCommitRrrLogsInit(t *testing.T) {
	defer leaktest.Check(t)()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	scatter := NewScatter(log, "")

	MockXaredologs := []*XaCommitErr{
		&XaCommitErr{
			Time:  "20180903103145",
			Xaid:  "RXID-20180903103145-1",
			State: txnXACommitErrStateRollback,
			Times: 10,
		},
	}
	dir := fakedb.GetTmpDir("/tmp", "xacheck", log)
	file := path.Join(dir, xacheckJSONFile)
	err := config.WriteConfig(file, &XaCommitErrs{Logs: MockXaredologs})
	assert.Nil(t, err)
	defer os.RemoveAll(file)

	xaChecker := NewXaCheck(scatter, MockScatterDefault2(dir))

	err = xaChecker.Init()
	assert.Nil(t, err)
	xaChecker.Close()
}

func TestXaCheckTimesOutFromFile(t *testing.T) {
	defer leaktest.Check(t)()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	data := `{
    "xacommit-errs": [
        {
            "time": "20180903103145",
            "xaid": "RXID-20180903103145-1",
            "state": "rollback",
            "times": 1
        }
    ]
}`
	dir := fakedb.GetTmpDir("/tmp", "xacheck", log)
	file := path.Join(dir, xacheckJSONFile)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	assert.Nil(t, err)
	// it is no need to remove the file, if it is successful, err item in the file will be removed
	//defer os.RemoveAll(file)

	scatter, fakedb1, cleanup1 := MockScatter(log, 2)
	defer cleanup1()

	err = scatter.Init(MockScatterDefault2(dir))
	assert.Nil(t, err)
	var backend [2]string
	var i int
	for k := range scatter.backends {
		backend[i] = k
		i++
	}

	fakedb1.AddQueryPattern("XA .*", result1)
	fakedb1.AddQuery("XA RECOVER", xaRecoverResult2)
	fakedb1.AddQuery("XA ROLLBACK .*", result1)

	// XA RECOVER and return Empty set
	// XA ROLLBACK ok
	assert.Equal(t, 1, scatter.txnMgr.xaCheck.GetRetrysLen())
	{
		txn2, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn2.Finish()

		time.Sleep(3 * time.Second)
	}
	assert.Equal(t, 0, scatter.txnMgr.xaCheck.GetRetrysLen())

	info, err := os.Stat(path.Join(scatter.txnMgr.xaCheck.dir, xacheckTimesOutJSONFile))
	assert.Nil(t, err)
	assert.NotNil(t, info)
}
