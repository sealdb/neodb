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

func TestProxyExplain(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
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
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "create table t1(id int, b int) partition by hash(id)"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
		client.Quit()
	}

	// explain.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "explain select 1, sum(a),avg(a),a,b from test.t1 as t1 where id>1 group by a,b order by a desc limit 10 offset 100"
		qr, err := client.FetchAll(query, -1)
		assert.Nil(t, err)
		want := `{
	"RawQuery": "explain select 1, sum(a),avg(a),a,b from test.t1 as t1 where id>1 group by a,b order by a desc limit 10 offset 100",
	"Project": "1, sum(a), avg(a), a, b",
	"Partitions": [
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0000 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend0",
			"Range": "[0-128)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0001 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend0",
			"Range": "[128-256)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0002 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend0",
			"Range": "[256-384)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0003 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend0",
			"Range": "[384-512)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0004 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend0",
			"Range": "[512-640)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0005 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend0",
			"Range": "[640-819)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0006 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend1",
			"Range": "[819-947)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0007 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend1",
			"Range": "[947-1075)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0008 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend1",
			"Range": "[1075-1203)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0009 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend1",
			"Range": "[1203-1331)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0010 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend1",
			"Range": "[1331-1459)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0011 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend1",
			"Range": "[1459-1638)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0012 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend2",
			"Range": "[1638-1766)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0013 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend2",
			"Range": "[1766-1894)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0014 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend2",
			"Range": "[1894-2022)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0015 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend2",
			"Range": "[2022-2150)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0016 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend2",
			"Range": "[2150-2278)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0017 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend2",
			"Range": "[2278-2457)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0018 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend3",
			"Range": "[2457-2585)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0019 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend3",
			"Range": "[2585-2713)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0020 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend3",
			"Range": "[2713-2841)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0021 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend3",
			"Range": "[2841-2969)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0022 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend3",
			"Range": "[2969-3097)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0023 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend3",
			"Range": "[3097-3276)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0024 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend4",
			"Range": "[3276-3404)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0025 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend4",
			"Range": "[3404-3532)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0026 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend4",
			"Range": "[3532-3660)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0027 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend4",
			"Range": "[3660-3788)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0028 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend4",
			"Range": "[3788-3916)"
		},
		{
			"Query": "select 1, sum(a), sum(a) as ` + "`avg(a)`" + `, count(a), a, b from test.t1_0029 as t1 where id > 1 group by a, b order by a desc",
			"Backend": "backend4",
			"Range": "[3916-4096)"
		}
	],
	"Aggregate": [
		"sum(a)",
		"avg(a)",
		"sum(a)",
		"count(a)"
	],
	"GatherMerge": [
		"a"
	],
	"HashGroupBy": [
		"a",
		"b"
	],
	"Limit": {
		"Offset": 100,
		"Limit": 10
	}
}`
		got := string(qr.Rows[0][0].Raw())
		log.Info(got)
		assert.Equal(t, want, got)
	}
}

func TestProxyExplainError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create table .*", &sqltypes.Result{})
	}

	// build plan error.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "explain select xx from sdf"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
		want := "Table 'test.sdf' doesn't exist (errno 1146) (sqlstate 42S02)"
		got := err.Error()
		assert.Equal(t, want, got)
	}

	// parse query error.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "explain format = none select xx from sdf"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
		want := "You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use, syntax error at position 22 near 'none' (errno 1149) (sqlstate 42000)"
		assert.Equal(t, want, err.Error())
	}
}

func TestProxyExplainUnsupported(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create table .*", &sqltypes.Result{})
	}

	// parse query error.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "explain create table t1(a int)"
		_, err = client.FetchAll(query, -1)
		want := "You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use, syntax error at position 15 near 'create' (errno 1149) (sqlstate 42000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}
}

func TestProxyExplainPrivilege(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxyPrivilegeN(log, MockDefaultConfig())
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
	}

	// explain.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "explain select 1, sum(a),avg(a),a,b from test.t1 as t1 where id>1 group by a,b order by a desc limit 10 offset 100"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}
}
