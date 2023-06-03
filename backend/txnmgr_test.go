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

	"github.com/sealdb/neodb/fakedb"

	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestTxnManager(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedb := fakedb.New(log, 2)
	defer fakedb.Close()
	backends := make(map[string]*Poolz)
	addrs := fakedb.Addrs()
	for _, addr := range addrs {
		conf := MockBackendConfigDefault(addr, addr)
		poolz := NewPoolz(log, conf)
		backends[addr] = poolz
	}
	txnmgr := NewTxnManager(log)

	{
		txn, err := txnmgr.CreateTxn(backends)
		assert.Nil(t, err)
		defer txn.Finish()
	}
}
