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
	"github.com/sealdb/neodb/planner"
	"github.com/sealdb/neodb/router"
	"github.com/sealdb/neodb/xcontext"

	"github.com/pkg/errors"
	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestSelectExecutorErr(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableAConfig(), router.MockTableBConfig())
	assert.Nil(t, err)

	// Create scatter and query handler.
	scatter, fakedbs, cleanup := backend.MockScatter(log, 10)
	defer cleanup()
	// desc
	fakedbs.AddQueryErrorPattern("select .*", errors.New("mock.execute.error"))

	querys := []string{
		"select * from A",
	}

	for _, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)

		plan := planner.NewSelectPlan(log, database, query, node.(*sqlparser.Select), route)
		err = plan.Build()
		assert.Nil(t, err)
		log.Debug("plan:%+v", plan.JSON())

		txn, err := scatter.CreateTransaction()
		assert.Nil(t, err)
		defer txn.Finish()
		executor := NewSelectExecutor(log, plan, txn)
		{
			ctx := xcontext.NewResultContext()
			err := executor.Execute(ctx)
			assert.NotNil(t, err)
		}
	}
}
