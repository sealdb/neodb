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
	"testing"
	"time"

	"github.com/sealdb/neodb/xcontext"

	"github.com/fortytw2/leaktest"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestTxnz(t *testing.T) {
	defer leaktest.Check(t)()
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedb, txnMgr, backends, addrs, cleanup := MockTxnMgr(log, 2)
	defer cleanup()

	querys := []*xcontext.QueryTuple{
		&xcontext.QueryTuple{Query: "select * from node1", Backend: addrs[0]},
		&xcontext.QueryTuple{Query: "select * from node2", Backend: addrs[1]},
	}

	fakedb.AddQueryDelay(querys[0].Query, result2, 10000)
	fakedb.AddQueryDelay(querys[1].Query, result2, 10000)

	{
		txn, err := txnMgr.CreateTxn(backends)
		assert.Nil(t, err)
		defer txn.Finish()

		qzRows := tz.GetTxnzRows()
		assert.NotNil(t, qzRows)

		time.Sleep(30 * time.Millisecond)
		qzRows = tz.GetTxnzRows()
		assert.NotNil(t, qzRows)

		time.Sleep(100 * time.Millisecond)
		qzRows = tz.GetTxnzRows()
		assert.NotNil(t, qzRows)
	}
}
