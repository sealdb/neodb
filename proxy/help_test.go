/*
 * NeoDB
 *
 * Copyright 2020 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package proxy

import (
	"errors"
	"testing"

	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestHelpStmt(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fakedbs.AddQueryPattern("help *.*", &sqltypes.Result{})
		fakedbs.AddQueryErrorPattern("help go", errors.New("Nothing found"))
	}

	// normal help stmt.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		defer client.Close()
		querys := []string{
			"help",
			"help index",
			"help content",
			"help 'any string'",
		}
		for _, query := range querys {
			_, err = client.FetchAll(query, -1)
			assert.Nil(t, err)
		}
	}
	// Error help stmt.
	{
		client, err := driver.NewConn("mock", "mock", address, "test", "utf8")
		assert.Nil(t, err)
		defer client.Close()
		query := "help go"
		_, actual := client.FetchAll(query, -1)
		expected := "Nothing found (errno 1105) (sqlstate HY000)"
		assert.EqualValues(t, expected, actual.Error())
	}
}
