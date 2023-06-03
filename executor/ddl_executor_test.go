/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package executor

import (
	"testing"

	"github.com/sealdb/neodb/backend"
	"github.com/sealdb/neodb/fakedb"
	"github.com/sealdb/neodb/planner"
	"github.com/sealdb/neodb/router"
	"github.com/sealdb/neodb/xcontext"

	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestDDLExecutor(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	// Create scatter and query handler.
	scatter, fakedbs, cleanup := backend.MockScatter(log, 10)
	defer cleanup()
	fakedbs.AddQueryPattern("create table sbtest.A.*", fakedb.Result3)
	fakedbs.AddQueryPattern("create database.*", fakedb.Result3)

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	err = route.AddForTest(database, router.MockTableAConfig())
	assert.Nil(t, err)

	// create table
	{
		query := "create table A(a int)"
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)

		plan := planner.NewDDLPlan(log, database, query, node.(*sqlparser.DDL), route)
		err = plan.Build()
		assert.Nil(t, err)

		txn, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn.Finish()
		executor := NewDDLExecutor(log, plan, txn)
		{
			ctx := xcontext.NewResultContext()
			err := executor.Execute(ctx)
			assert.Nil(t, err)
		}
	}
}
