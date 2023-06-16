/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
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

func TestProxyUseDatabase(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("use test", &sqltypes.Result{})
	}

	// connection without database.
	{
		_, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
	}

	// lower case.
	{
		session := proxy.sessions.getSession(1).session
		spanner := proxy.Spanner()
		spanner.ComInitDB(session, "TEST")
		assert.Equal(t, "TEST", session.Schema())

		proxy.SetLowerCaseTableNames(1)
		spanner.ComInitDB(session, "TEST")
		assert.Equal(t, "test", session.Schema())
	}

	// use db.
	{
		_, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		//// 'use db' In MySQL client use COM_INIT_DB, but the client.FetchAll use COM_QUERY, so comment the below.
		//query := "use test"
		//_, err = client.FetchAll(query, -1)
		//assert.Nil(t, err)
	}
}

func TestProxyUseDatabasePrivilegeNotSuper(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxyPrivilegeNotSuper(log, MockDefaultConfig())
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("use test1", &sqltypes.Result{})
	}

	// connection without database.
	{
		_, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
	}

	// use db.
	{
		_, err := driver.NewConn("mock", "mock", address, "test1", "utf8")
		assert.Nil(t, err)
	}
}

func TestProxyUseDatabasePrivilegeDB(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxyPrivilegeN(log, MockDefaultConfig())
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("use test1", &sqltypes.Result{})
	}

	// connection without database.
	{
		_, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
	}

	// use db.
	{
		_, err := driver.NewConn("mock", "mock", address, "test1", "utf8")
		assert.NotNil(t, err)
	}
}
