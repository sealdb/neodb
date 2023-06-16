/*
 * NeoDB
 *
 * Copyright 2018-2020 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package proxy

import (
	"testing"

	"github.com/sealdb/neodb/plugins/shiftmanager"

	"github.com/pkg/errors"
	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqlparser"
	querypb "github.com/sealdb/mysqlstack/sqlparser/depends/query"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

var (
	rdbs = &sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name: "Databases",
				Type: querypb.Type_VARCHAR,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("test")),
			},
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("information_schema")),
			},
		},
	}

	res1 = &sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name: "table_name",
				Type: querypb.Type_VARCHAR,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("t1_cleanup")),
			},
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("t2_cleanup")),
			},
		},
	}

	res2 = &sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name: "table_name",
				Type: querypb.Type_VARCHAR,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("t1_migrate")),
			},
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("t2_migrate")),
			},
		},
	}
)

func TestCleanup(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedb, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()
	router := proxy.Router()
	spanner := proxy.Spanner()

	// fakedbs.
	{
		fakedb.AddQuery("show databases", rdbs)
		fakedb.AddQueryPattern("drop .*", &sqltypes.Result{})
		fakedb.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedb.AddQuery("select table_name from information_schema.tables where table_schema = 'test' and table_name like '%_cleanup'", res1)
		fakedb.AddQuery("select table_name from information_schema.tables where table_schema = 'test' and table_name like '%_migrate'", res2)
	}

	// cleanup database.
	{
		query := "neodb cleanup"
		_, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		c := NewCleanup(log, spanner.scatter, router, spanner)
		_, err = c.Cleanup()
		assert.Nil(t, err)
	}

	// create database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create test.t1_cleanup.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table test.t1_cleanup(id int, b int) global"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create test.t1_migrate.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table test.t1_migrate(id int, b int) global"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// cleanup table.
	{
		query := "neodb cleanup"
		_, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		c := NewCleanup(log, spanner.scatter, router, spanner)
		_, err = c.Cleanup()
		assert.Nil(t, err)
	}

	// mock shift.
	{
		key := "`test`.`t2`_backend1"
		mockshift := shiftmanager.NewMockShift(log)
		shiftMgr := proxy.Plugins().PlugShiftMgr()
		err := shiftMgr.StartShiftInstance(key, mockshift, shiftmanager.ShiftTypeRebalance)
		assert.Nil(t, err)
		assert.Equal(t, shiftmanager.ShiftStatusMigrating, shiftMgr.GetStatus(key))
	}

	// cleanup table.
	{
		query := "neodb cleanup"
		_, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		c := NewCleanup(log, spanner.scatter, router, spanner)
		_, err = c.Cleanup()
		assert.Nil(t, err)
	}
}

func TestCleanupShowDatabaseError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedb, proxy, cleanup := MockProxy(log)
	defer cleanup()
	router := proxy.Router()
	spanner := proxy.Spanner()

	// fakedbs.
	{
		fakedb.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedb.AddQueryErrorPattern("show databases", errors.New("mock.show.databases.error"))
	}

	// show databases error.
	{
		query := "neodb cleanup"
		_, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		c := NewCleanup(log, spanner.scatter, router, spanner)
		_, err = c.Cleanup()
		assert.NotNil(t, err)
	}
}

func TestCleanupSelectCleanupErr(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedb, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()
	router := proxy.Router()
	spanner := proxy.Spanner()

	// fakedbs.
	{
		fakedb.AddQuery("show databases", rdbs)
		fakedb.AddQueryPattern("drop .*", &sqltypes.Result{})
		fakedb.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedb.AddQueryError("select table_name from information_schema.tables where table_schema = 'test' and table_name like '%_cleanup'", errors.New("mock.drop.error"))
	}

	// create database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create test table.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table test.t1_cleanup(id int, b int) global"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// select error.
	{
		query := "neodb cleanup"
		_, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		c := NewCleanup(log, spanner.scatter, router, spanner)
		_, err = c.Cleanup()
		assert.NotNil(t, err)
	}
}

func TestCleanupSelectMigrateErr(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedb, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()
	router := proxy.Router()
	spanner := proxy.Spanner()

	// fakedbs.
	{
		fakedb.AddQuery("show databases", rdbs)
		fakedb.AddQueryPattern("drop .*", &sqltypes.Result{})
		fakedb.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedb.AddQuery("select table_name from information_schema.tables where table_schema = 'test' and table_name like '%_cleanup'", res1)
		fakedb.AddQueryError("select table_name from information_schema.tables where table_schema = 'test' and table_name like '%_migrate'", errors.New("mock.drop.error"))
	}

	// create database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create test table.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table test.t1_cleanup(id int, b int) global"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// select error.
	{
		query := "neodb cleanup"
		_, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		c := NewCleanup(log, spanner.scatter, router, spanner)
		_, err = c.Cleanup()
		assert.NotNil(t, err)
	}
}

func TestCleanupDropErr(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedb, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()
	router := proxy.Router()
	spanner := proxy.Spanner()

	// fakedbs.
	{
		fakedb.AddQuery("show databases", rdbs)
		fakedb.AddQueryErrorPattern("drop .*", errors.New("mock.drop.error"))
		fakedb.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedb.AddQuery("select table_name from information_schema.tables where table_schema = 'test' and table_name like '%_cleanup'", res1)
		fakedb.AddQuery("select table_name from information_schema.tables where table_schema = 'test' and table_name like '%_migrate'", res2)
	}

	// drop database error.
	{
		query := "neodb cleanup"
		_, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		c := NewCleanup(log, spanner.scatter, router, spanner)
		_, err = c.Cleanup()
		assert.NotNil(t, err)
	}

	// create database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create test.t1_cleanup.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table test.t1_cleanup(id int, b int) global"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// drop table with suffix _cleanup error.
	{
		query := "neodb cleanup"
		_, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		c := NewCleanup(log, spanner.scatter, router, spanner)
		_, err = c.Cleanup()
		assert.NotNil(t, err)
	}

	// create test.t2_cleanup.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table test.t2_cleanup(id int, b int) global"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// drop table with suffix _migrate error.
	{
		query := "neodb cleanup"
		_, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		c := NewCleanup(log, spanner.scatter, router, spanner)
		_, err = c.Cleanup()
		assert.NotNil(t, err)
	}
}

func TestCleanupReadOnly(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()

	address := proxy.Address()
	client, err := driver.NewConn("mock", "mock", address, "", "utf8")
	assert.Nil(t, err)

	// set neodb readonly.
	proxy.SetReadOnly(true)
	query := "neodb cleanup"
	fakedbs.AddQuery(query, &sqltypes.Result{})
	_, err = client.FetchAll(query, -1)
	want := "The MySQL server is running with the --read-only option so it cannot execute this statement (errno 1290) (sqlstate 42000)"
	got := err.Error()
	assert.Equal(t, want, got)
}
