/*
 * NeoDB
 *
 * Copyright 2018-2019 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package proxy

import (
	"testing"

	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestErrorParams(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// attach.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "neodb attach('attach1', '127.0.0.1:6000', 'root', '123456', 'xxxx')"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}

	// detach.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "neodb detach('attach1','127')"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}

	// reshard.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "neodb reshard db.tb to db2.t2"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}
}

func TestNeoDBCleanup(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedb, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedb.AddQueryPattern("show databases", &sqltypes.Result{})
	}

	// cleanup.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "neodb cleanup"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}
}
