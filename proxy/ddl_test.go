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
	"errors"
	"fmt"
	"testing"

	"github.com/sealdb/neodb/fakedb"

	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestProxyDDLDB(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// create database querys
	querys := []string{
		"create database test1 charset default",
		"create database test2 character set default",
		"create database test3 charset utf8mb4",
		"create database test4 character set latin1",
		"create database test5 collate latin1_swedish_ci",
		"create database if not exists test6 collate utf8mb4_unicode_ci charset utf8mb4",
		"create database if not exists test7 collate utf8mb4_unicode_ci charset utf8mb4 charset utf8mb4 collate utf8mb4_unicode_ci",
		"create schema test8 charset default",
		"create schema test9 default encryption = 'N'",
	}

	// fakedbs.
	{
		fakedbs.AddQueryPattern(".* database .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern(".* schema .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("alter database .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
	}

	// create database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		for _, query := range querys {
			_, err = client.FetchAll(query, -1)
			assert.Nil(t, err)
		}
	}

	// create database again.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database if not exists test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// alter database with db not exists.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "alter database dbNotExist collate = utf8mb4_0900_ai_ci read only = default character set = utf8"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}

	// alter database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		// use database
		query := "use test1"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)

		// test alter with default session database test1
		querys := []string{
			"alter database test1 collate = utf8mb4_0900_ai_ci read only = default character set = utf8",
			"alter database collate = utf8mb4_0900_ai_ci read only = default character set = utf8",
		}
		for _, query := range querys {
			_, err = client.FetchAll(query, -1)
			assert.Nil(t, err)
		}
	}

	// alter database error.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		// test alter with no default session database
		query := "alter database collate = utf8mb4_0900_ai_ci read only = default character set = utf8"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}

	// drop database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "drop database test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// drop database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "drop database if exists test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// drop schema.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "drop schema if exists test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// ACL database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database mysql"
		_, err = client.FetchAll(query, -1)
		want := "Access denied; lacking privileges for database mysql (errno 1227) (sqlstate 42000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}
}

func TestProxyDDLResultError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// Add pattern
	{
		fakedbs.AddQueryErrorPattern("alter database .*", errors.New("alter database error"))
		fakedbs.AddQueryPattern(".* database .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
	}

	// create database test.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database if not exists test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// alter database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		// use database
		query := "use test1"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)

		// test alter with default session database test1
		query = "alter database test collate = utf8mb4_0900_ai_ci read only = default character set = utf8"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}

}

func TestProxyDDLDBError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// create database querys, error from mysql
	querysError := []string{
		// In NeoDB, utf8m/latin1 are parsed successfully, we just treat them as
		// normal strings, but in mysql, they are not correct character set, so
		// we still treat the sql next are not correct.
		"create database test3 charset utf8m",  // utf8m-->utf8mb4
		"create database test4 collate latin1", // latin1-->latin1_swedish_ci
		// ERROR 1253 (42000): COLLATION 'latin1_swedish_ci' is not valid for CHARACTER SET 'utf8mb4'
		"create database if not exists test5 character set utf8mb4 charset utf8mb4 collate utf8mb4_unicode_ci",
		"create schema test6 default encryption = 'Y'",
	}

	// fakedbs.
	{
		fakedbs.AddQueryErrorPattern(".* database .*", errors.New("create database error"))
		fakedbs.AddQueryErrorPattern(".* schema .*", errors.New("create database error"))
	}

	// create database and return error.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		for _, query := range querysError {
			_, err = client.FetchAll(query, -1)
			assert.NotNil(t, err)
		}
	}
}

func TestProxyDDLTable(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("show tables from .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("alter table .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("drop table .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("truncate table .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("truncate .*", &sqltypes.Result{})
	}

	// create table without db.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table t1(a int, b int)"
		_, err = client.FetchAll(query, -1)
		want := "No database selected (errno 1046) (sqlstate 3D000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}

	// create database;
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create global table.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "create table t2(a int, b int) GLOBAL"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create global table again.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "create table if not exists t2(a int, b int) GLOBAL"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create global table database error.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "create table if not exists ttt.t2(a int, b int) GLOBAL"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}

	// check test.tables.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "show tables"
		qr, err := client.FetchAll(query, -1)
		assert.Nil(t, err)
		want := "[[t2]]"
		got := fmt.Sprintf("%+v", qr.Rows)
		assert.Equal(t, want, got)
	}

	// drop global table.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "drop table t2"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// drop global table again.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "drop table if exists t2"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create single table.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "create table t3(a int, b int) SINGLE"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create single table again.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "create table if not exists t3(a int, b int) SINGLE"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create single table database error.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "create table if not exists ttt.t3(a int, b int) SINGLE"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}

	// drop single table.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "drop table t3"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create table with table_options
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		querys := []string{
			"create table if not exists t4(a int, b int) comment 'comment test' charset='utf8' SINGLE",
			"create table if not exists t5(a int, b int) default charset utf8 Global",
			"create table if not exists t6(a int key, b int) default character set='utf8' comment 'test' engine innodb",
		}
		for _, query := range querys {
			_, err = client.FetchAll(query, -1)
			assert.Nil(t, err)
		}
	}

	// drop tables.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		querys := []string{
			"drop table t4",
			"drop table t5",
			"drop table t6",
		}
		for _, query := range querys {
			_, err = client.FetchAll(query, -1)
			assert.Nil(t, err)
		}
	}

	// check test.tables.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "show tables"
		qr, err := client.FetchAll(query, -1)
		assert.Nil(t, err)
		want := "[]"
		got := fmt.Sprintf("%+v", qr.Rows)
		assert.Equal(t, want, got)
	}

	// create table(ACL).
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "create table mysql.t2(id int, b int) partition by hash(id)"
		_, err = client.FetchAll(query, -1)
		want := "Access denied; lacking privileges for database mysql (errno 1227) (sqlstate 42000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}

	// create test table.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "create table t1(id int, b int) partition by hash(id)"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create sbtest database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database sbtest"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create sbtest table.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table sbtest.sbt1(id int, b int) partition by hash(id)"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create table with non_reserved_keyword key word.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table sbtest.sbtx(status int, bool int, datetime datetime, enum char) partition by hash(status)"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// drop single table.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "drop table sbtest.sbtx"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// alter test table engine.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "alter table t1 engine=tokudb"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// truncate table.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "truncate table t1"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// truncate without table.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "truncate t1"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create sbtest table mysql internal error.
	{
		fakedbs.AddQueryErrorPattern("create table .*", errors.New("mock.mysql.create.table.error"))

		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table sbtest.sberror2(id int, b int) partition by hash(id)"
		_, err = client.FetchAll(query, -1)
		want := "mock.mysql.create.table.error (errno 1105) (sqlstate HY000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}

	// check sbtest.tables.
	{
		client, err := driver.NewConn("mock", "mock", address, "sbtest", "utf8")
		assert.Nil(t, err)
		query := "show tables"
		qr, err := client.FetchAll(query, -1)
		assert.Nil(t, err)
		want := "[[sbt1]]"
		got := fmt.Sprintf("%+v", qr.Rows)
		assert.Equal(t, want, got)
	}

	// drop sbtest table error.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "drop table sbtest.t1"
		_, err = client.FetchAll(query, -1)
		want := "Table 't1' doesn't exist (errno 1146) (sqlstate 42S02)"
		got := err.Error()
		assert.Equal(t, want, got)
	}

	// drop sbtest1 table error.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "drop table sbtest1.t1"
		_, err = client.FetchAll(query, -1)
		want := "Unknown database 'sbtest1' (errno 1049) (sqlstate 42000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}

	// drop sbtest table.
	{
		client, err := driver.NewConn("mock", "mock", address, "sbtest", "utf8")
		assert.Nil(t, err)
		query := "drop table sbt1"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// check sbtest.tables.
	{
		client, err := driver.NewConn("mock", "mock", address, "sbtest", "utf8")
		assert.Nil(t, err)
		query := "show tables"
		qr, err := client.FetchAll(query, -1)
		assert.Nil(t, err)
		want := "[]"
		got := fmt.Sprintf("%+v", qr.Rows)
		assert.Equal(t, want, got)
	}

	// create sbtest table.
	{
		fakedbs.ResetPatternErrors()
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table sbtest.sbt1(id int, b int) partition by hash(id)"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// drop sbtest table internal error.
	{
		fakedbs.AddQueryErrorPattern("drop table .*", errors.New("mock.mysql.drop.table.error"))
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "drop table sbtest.sbt1"
		_, err = client.FetchAll(query, -1)
		want := "mock.mysql.drop.table.error (errno 1105) (sqlstate HY000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}
}

func TestProxyDropTables(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("show tables from .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("alter table .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("drop table db1.*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("drop table db2.*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("truncate table .*", &sqltypes.Result{})
	}

	// create database;
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database db1"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}
	// create database;
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database db2"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create test table.
	{
		client, err := driver.NewConn("mock", "mock", address, "db1", "utf8")
		assert.Nil(t, err)
		query := "create table t1(id int, b int) partition by hash(id)"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}
	// create global table.
	{
		client, err := driver.NewConn("mock", "mock", address, "db1", "utf8")
		assert.Nil(t, err)
		query := "create table t2(a int, b int) GLOBAL"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}
	// create single table.
	{
		client, err := driver.NewConn("mock", "mock", address, "db2", "utf8")
		assert.Nil(t, err)
		query := "create table t3(a int, b int) SINGLE"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// drop tables.
	{
		client, err := driver.NewConn("mock", "mock", address, "db1", "utf8")
		assert.Nil(t, err)
		query := "drop table db2.t3, t2, db1.t1"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}
}

func TestProxyDropTablesPrivilegeN(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxyPrivilegeN(log, MockDefaultConfig())
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern(".* database .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
	}

	// drop tables.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "drop table test.t1"
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}
}

func TestProxyDDLIndex(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("show tables from .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("show create table .*", fakedb.Result1)
		fakedbs.AddQueryPattern("drop .*", &sqltypes.Result{})
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
	}

	// show create test table.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "show create table t1"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create index.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		querys := []string{
			"create index index1 on t1(a,b)",
			"create index index1 on t1(a,b) lock=shared",
			"create index index1 on t1(a,b) algOrithm=inplace",
			"create index index1 on t1(a,b) lock=none algOrithm=default",
			"create index index1 on t1(a,b) algOrithm=copY Lock = shared",
		}
		for _, query := range querys {
			_, err = client.FetchAll(query, -1)
			assert.Nil(t, err)
		}
	}

	// create index error, unknown database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create index index1 on xx.t1(a,b)"
		_, err = client.FetchAll(query, -1)
		want := "Unknown database 'xx' (errno 1049) (sqlstate 42000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}

	// create index error, wrong type name, duplicate lock or algorithm options.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		querys := []string{
			"create index index1 on t1(a,b) lock=xxx",
			"create index index1 on t1(a,b) lock=default lock=none",
			"create index index1 on t1(a,b) lock=default algorithm=default lock=default",
			"create index index1 on t1(a,b) algorithm=wrong",
			"create index index1 on t1(a,b) algorithm=copy algorithm=default",
			"create index index1 on t1(a,b) lock=default algorithm=copy algorithm=default",
		}
		for _, query := range querys {
			_, err = client.FetchAll(query, -1)
			assert.NotNil(t, err)
		}
	}

	// drop index.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "drop index index1 on t1 lock=default algorithm=default"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// drop index error, wrong type name, duplicate lock or algorithm options.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		querys := []string{
			"drop index index1 on t1 lock=xxx",
			"drop index index1 on t1 lock=default lock=none",
			"drop index index1 on t1 lock=default algorithm=default lock=default",
			"drop index index1 on t1 algorithm=wrong",
			"drop index index1 on t1 algorithm=copy algorithm=default",
			"drop index index1 on t1 lock=default algorithm=copy algorithm=default",
		}
		for _, query := range querys {
			_, err = client.FetchAll(query, -1)
			assert.NotNil(t, err)
		}
	}

	// create fulltext index.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "create fulltext index fts1 on t1(a)"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}
}

func TestProxyDDLColumn(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	querys := []string{
		"create table t1(id int, b int) partition by hash(id)",
		"alter table t1 add column(c1 int, c2 varchar(100))",
		"alter table t1 drop column c2",
		"alter table t1 modify column c2 varchar(1)",
		"create table t2(id int, b int) GLOBAL",
		"alter table t2 add column(c1 int, c2 varchar(100))",
		"alter table t2 drop column c2",
		"alter table t2 modify column c2 varchar(1)",
		"alter table t2 drop column id",
		"alter table t2 modify column id bigint",
		"alter table t2 add column(c3 bigint not null key primary key unique not null key not null comment 'NeoDB', c4 int)",
		"alter table t2 add column(c5 timestamp ON UPDATE CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT 'currenttimestamp' not null key primary key unique not null key not null comment 'NeoDB', c6 int)",
		"alter table t2 add column(status int, bool int, datetime datetime, enum char)",
	}
	queryerr := []string{
		"alter table t1 drop column ID",
		"alter table t1 modify column id bigint",
	}
	wants := []string{
		"unsupported: cannot.drop.the.column.on.shard.key (errno 1105) (sqlstate HY000)",
		"unsupported: cannot.modify.the.column.on.shard.key (errno 1105) (sqlstate HY000)",
	}
	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("alter table .*", &sqltypes.Result{})
	}

	// create database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create test table, t1 hash t2 global.
	// add column.
	// drop column.
	// modify column.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		for _, query := range querys {
			_, err = client.FetchAll(query, -1)
			assert.Nil(t, err)
		}
	}

	// drop column error(drop the shardkey).
	// modify column error(drop the shardkey).
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		for i, query := range queryerr {
			_, err = client.FetchAll(query, -1)
			got := err.Error()
			assert.Equal(t, wants[i], got)
		}
	}
}

func TestProxyDDLUnsupported(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("rename .*", &sqltypes.Result{})
	}

	// rename test table.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "rename table t1 to t2"
		_, err = client.FetchAll(query, -1)
		want := "You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use, syntax error at position 7 near 'rename' (errno 1149) (sqlstate 42000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}
}

func TestProxyDDLCreateTable(t *testing.T) {
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

	querys := []string{
		"create table t1(a int, b int) partition by hash(a)",
		"create table t2(a int, b int) PARTITION BY hash(a)",
		"create table t3(a int, b int)   PARTITION  BY hash(a)  ",
		"create table t4(a int, b int)engine=tokudb PARTITION  BY hash(a)  ",
		"create table t5(a int, b int) default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t6(a int, b int)engine=tokudb auto_increment=10 default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t7(a int collate utf8_bin Collate 'utf8_bin' collate \"utf8_bin\") partition by hash(a)",
		"create table t8(a int, b int) partition by hash(a)",
		"create table t9(a int, b timestamp(5) on update current_timestamp(5) column_format fixed column_format default column_format dynamic) partition by hash(a)",
		"create table t10(a int column_format fixed column_format default column_format dynamic) comment='comment option' engine=tokudb default charset='utf8' avg_row_length=123 checksum=1 collate='utf8_bin' compression='lz4' connection='id' data directory='/data' index directory='/index' delay_key_write=1 encryption='n' insert_method=First key_block_size=1 max_rows=3 min_rows=2 pack_keys=default password='pwd' row_format=dynamic stats_auto_recalc=1 stats_persistent=default stats_sample_pages=65535 tablespace=storage partition by hash(a)",
		"create table t11(a int, b int) partition by hash(A)",
	}

	for _, query := range querys {
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}
}

func TestProxyDDLCreateTableError(t *testing.T) {
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

	querys := []string{
		"create table t2(a int, partition int) PARTiITION BY hash(a)",
		"create table dual(a int) partition by hash(a)",
		"create table t(a int) engine=unkown partition by hash(a)",
	}
	results := []string{
		"You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use, syntax error at position 33 near 'partition' (errno 1149) (sqlstate 42000)",
		"spanner.ddl.check.create.table[dual].error:not support (errno 1105) (sqlstate HY000)",
		"Unknown storage engine 'unkown', currently we only support InnoDB and TokuDB (errno 1286) (sqlstate 42000)",
	}

	for i, query := range querys {
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		_, err = client.FetchAll(query, -1)
		want := results[i]
		got := err.Error()
		assert.Equal(t, want, got)
	}
}

func TestProxyMyLoaderImport(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("show create database .*", &sqltypes.Result{})
		fakedbs.AddQuery("/*show create database sbtest*/", &sqltypes.Result{})
	}

	// create database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	querys := []string{
		"create table t1(a int, b int) partition by hash(a)",
		"show create database sbtest",
		"/*show create database sbtest*/",
		"SET autocommit=0",
		"SET SESSION wait_timeout = 2147483",
	}

	for _, query := range querys {
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}
}

func TestProxyDDLConstraint(t *testing.T) {
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

	querys := []string{
		"CREATE TABLE t0(a int unique,b int ) PARTITION BY HASH(a);",
		"create table t1(a int key, b int) partition by hash(a)",
		"create table t3(a int unique, b int, c int) PARTITION BY hash(a)",
		"create table t4(a int unique key, b int)   PARTITION  BY hash(a)  ",
		"create table t5(a int primary key, b int) partition by hash(a)",
		"create table t9(a int, b int, primary key(a))engine=tokudb PARTITION  BY hash(a)  ",
		"create table t12(a int, b int, primary key(a,b)) default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t15(a int unique, b int, primary key(a,b))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t17(a int unique, b int, primary key(a))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t18(a int unique, b int, key `name` (`a`))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t19(a int unique, b int, index `name` (a))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t20(a int unique, b int, unique index `name` (a))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t21(a int unique, b int, unique key `name` (a))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t22(`a` bigint not null unique default current_timestamp auto_increment unique key key primary key comment 'NeoDB' auto_increment primary key)engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t23(a int unique, b timestamp ON UPDATE CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT 'currenttimestamp' DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'currenttimestamp' ON UPDATE CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP)engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
	}

	for _, query := range querys {
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}
}

func TestProxyDDLConstraintError(t *testing.T) {
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

	querys := []string{
		"create table t1(a int unique index, b int unique) partition by hash(a)",
		"create table t2(a int, b int unique) partition by hash(a)",
		"create table t3(a int unique, b int unique) partition by hash(a)",
		"create table t4(a int, b int primary key) PARTITION BY hash(a)",
		"create table t5(a int unique key, b int primary key)   PARTITION  BY hash(a)  ",
		"create table t6(a int primary key, b int primary key) partition by hash(a)",
		"create table t7(a int, b int unique, primary key(a))engine=tokudb PARTITION  BY hash(a)  ",
		"create table t8(a int, b int unique key, primary key(a))engine=tokudb PARTITION  BY hash(a)  ",
		"create table t9(a int unique key, b int unique key, primary key(a))engine=tokudb PARTITION  BY hash(a)",
		"create table t10(a int unique, b int unique, c int unique, primary key(a,b)) default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t11(a int unique, b int, c int, primary key(b)) default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t12(a int unique, b int, c int, primary key(b, c)) default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t13(a int unique, b int, c int, unique key `name` (`b`)) default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t14(a int unique, b int, c int, unique key `name` (`b`, `c`)) default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t15(a int key, b int key) partition by hash(a)",
		"create table t16(a int unique, b int key) PARTITION BY hash(a)",
		"create table t17(a int unique key, b int key)   PARTITION  BY hash(a)  ",
		"create table t18(a int primary key, b int key) partition by hash(a)",
		"create table t19(a int key, b int key, primary key(a))engine=tokudb PARTITION  BY hash(a)  ",
		"create table t21(a int key, b int key, primary key(a,b)) default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t22(a int unique key, b int key, primary key(a,b))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t23(a int unique, b int key, primary key(a))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t24(a int unique, b int key, key `name` (`a`))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t25(a int unique, b int key, index `name` (a))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t26(a int unique, b int key, unique index `name` (a))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t27(a int unique, b int key, unique key `name` (a))engine=tokudb default charset=utf8  PARTITION  BY hash(a)  ",
		"create table t28(a int primary key, b int unique)",
	}

	results := []string{
		"You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use, syntax error at position 35 near 'index' (errno 1149) (sqlstate 42000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[a] (errno 1105) (sqlstate HY000)",
	}

	for i, query := range querys {
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		_, err = client.FetchAll(query, -1)
		if err != nil {
			want := results[i]
			got := err.Error()
			assert.Equal(t, want, got)
		} else {
			log.Panic("proxy.ddl.constraint.test.case.did.not.return.err")
		}
	}
}

func TestProxyDDLShardKeyCheck(t *testing.T) {
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

	querys := []string{
		"CREATE TABLE t1(a int primary key,b int ) PARTITION BY HASH(`a`);",
		"CREATE TABLE t1(a int,b int ) PARTITION BY HASH(c);",
	}

	results := []string{
		"",
		"Sharding Key column 'c' doesn't exist in table (errno 1105) (sqlstate HY000)",
	}

	for i, query := range querys {
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		_, err = client.FetchAll(query, -1)
		if err != nil {
			want := results[i]
			got := err.Error()
			assert.Equal(t, want, got)
		}
	}
}

func TestProxyDDLAlterCharset(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("show tables from .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("alter table .*", &sqltypes.Result{})
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
	}

	// alter test table charset.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "alter table t1 convert to character set utf8mb"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}
}

func TestProxyDDLUnknowDatabase236(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("drop table .*", &sqltypes.Result{})
	}

	client, err := driver.NewConn("mock", "mock", address, "", "utf8")
	assert.Nil(t, err)
	query := "create database db1"
	_, err = client.FetchAll(query, -1)
	assert.Nil(t, err)

	query = "use db1"
	_, err = client.FetchAll(query, -1)
	assert.Nil(t, err)

	query = "DROP TABLE IF EXISTS `t1`"
	_, err = client.FetchAll(query, -1)
	assert.Nil(t, err)
}

func TestProxyDDLDBPrivilegeN(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxyPrivilegeN(log, MockDefaultConfig())
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern(".* database .*", &sqltypes.Result{})
	}

	// create database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database test"
		_, err = client.FetchAll(query, -1)
		want := "Access denied for user 'mock'@'%' to database 'test' (errno 1045) (sqlstate 28000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}
}

func TestProxyDDLGlobalSingleNormalList(t *testing.T) {
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

	querys := []string{
		"CREATE TABLE t1(a int primary key,b int )",
		"CREATE TABLE t2(a int primary key,b int ) GLOBAL",
		"CREATE TABLE t2(a int primary key,b int ) GLOBAL",
		"CREATE TABLE t4(a int primary key,b int ) partition by hash(a)",
		"CREATE TABLE t4(a int primary key,b int ) partition by hash(a)",
		"CREATE TABLE t3(a int primary key,b int ) SINGLE",
		"CREATE TABLE t3(a int primary key,b int ) SINGLE",
		"CREATE TABLE t1(a int ,b int )",
		"CREATE TABLE t5(a int ,b int, primary key(a))",
		"CREATE TABLE t6(a int ,b int, primary key(a, b))",
		"create table t7(a int, b int unique)",
		"create table `t8/t/t`(a int,b int, primary key(a))",

		// partition list
		"CREATE TABLE l(a int primary key,b int ) partition by list(a)(" +
			"PARTITION backend1 VALUES IN (1)," +
			"PARTITION backend2 VALUES IN (2));",
		"CREATE TABLE l(a int primary key,b int ) partition by list(a)(" +
			"PARTITION backend1 VALUES IN (1)," +
			"PARTITION backend2 VALUES IN (2));",
		"CREATE TABLE l(a int primary key,b int ) partition by list(b)(" +
			"PARTITION backend1 VALUES IN (1)," +
			"PARTITION backend2 VALUES IN (2));",
	}

	results := []string{
		"",
		"",
		"router.add.db[test].table[t2].exists (errno 1105) (sqlstate HY000)",
		"",
		"router.add.db[test].table[t4].exists (errno 1105) (sqlstate HY000)",
		"",
		"router.add.db[test].table[t3].exists (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint shoule be defined or add 'PARTITION BY HASH' to mandatory indication (errno 1105) (sqlstate HY000)",
		"",
		"The unique/primary constraint shoule be defined or add 'PARTITION BY HASH' to mandatory indication (errno 1105) (sqlstate HY000)",
		"",
		"invalid.table.name.currently.not.support.tablename[t8/t/t].contains.with.char:'/' or space ' ' (errno 1105) (sqlstate HY000)",

		// partition list
		"",
		"router.add.db[test].table[l].exists (errno 1105) (sqlstate HY000)",
		"The unique/primary constraint should be only defined on the sharding key column[b] (errno 1105) (sqlstate HY000)",
	}

	for i, query := range querys {
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		_, err = client.FetchAll(query, -1)
		if err != nil {
			want := results[i]
			got := err.Error()
			assert.Equal(t, want, got)
		}
	}
}

func TestProxyDDLAlterRename(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	querys := []string{
		"create table t1(id int, b int) partition by hash(id)",
		"alter table t1 rename t2",
	}

	queryerr := []string{
		"alter table ttt.t3 rename t4",
		"alter table t3 rename t4",
		"alter table t2 rename t2",
		"alter table t2 rename test2.t3",
	}
	wants := []string{
		"Unknown database 'ttt' (errno 1049) (sqlstate 42000)",
		"Table 't3' doesn't exist (errno 1146) (sqlstate 42S02)",
		"Table 't2' already exists (errno 1050) (sqlstate 42S01)",
		"unsupported: Database is not equal[test:test2] (errno 1105) (sqlstate HY000)",
	}
	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("rename table .*", &sqltypes.Result{})
	}

	// create database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create test table, t1 hash t2 global.
	// alter table rename.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		for _, query := range querys {
			_, err = client.FetchAll(query, -1)
			assert.Nil(t, err)
		}
	}

	// alter table t1 rename test2.t2.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		for i, query := range queryerr {
			_, err = client.FetchAll(query, -1)
			got := err.Error()
			assert.Equal(t, wants[i], got)
		}
	}
}

func TestProxyDDLCreateTableDistributed(t *testing.T) {
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

	querys := []string{
		"create table t1(a int, b int) distributed by (backend0)",
	}

	queryErrs := []string{
		"create table t1(a int, b int) distributed by (node0)",
		//querys have created it.
		"create table t1(a int, b int) distributed by (backend0)",
	}

	for _, query := range querys {
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	for _, query := range queryErrs {
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		_, err = client.FetchAll(query, -1)
		assert.NotNil(t, err)
	}
}

func TestProxyDDLUnsupportedAlter(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("alter table .*", &sqltypes.Result{})
	}

	// alter table a ORDER BY i.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		query := "alter table a ORDER BY i"
		_, err = client.FetchAll(query, -1)
		want := "unsupported.query:alter table a ORDER BY i (errno 1105) (sqlstate HY000)"
		got := err.Error()
		assert.Equal(t, want, got)
	}
}
