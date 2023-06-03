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
	"github.com/sealdb/neodb/backend"
	"github.com/sealdb/neodb/fakedb"
	"github.com/sealdb/neodb/planner"
	"github.com/sealdb/neodb/router"
	"github.com/sealdb/neodb/xcontext"
	"testing"

	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestDeleteExecutor(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	// Create scatter and query handler.
	scatter, fakedbs, cleanup := backend.MockScatter(log, 10)
	defer cleanup()

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableAConfig())
	assert.Nil(t, err)

	// delete.
	querys := []string{
		"delete from sbtest.A where id=1",
		"delete from sbtest.A where id=1 order by xx",
		"delete from sbtest.A where name='xx'",
		"delete from sbtest.A where id in (1, 2,3)",
	}
	// Add querys.
	fakedbs.AddQueryPattern("delete from sbtest..*", fakedb.Result3)

	for _, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)

		plan := planner.NewDeletePlan(log, database, query, node.(*sqlparser.Delete), route)
		err = plan.Build()
		assert.Nil(t, err)

		txn, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn.Finish()
		executor := NewDeleteExecutor(log, plan, txn)
		{
			ctx := xcontext.NewResultContext()
			err := executor.Execute(ctx)
			assert.Nil(t, err)
		}
	}
}
