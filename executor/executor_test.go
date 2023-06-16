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

	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestExecutor1(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	// Create scatter and query handler.
	scatter, fakedbs, cleanup := backend.MockScatter(log, 10)
	defer cleanup()
	fakedbs.AddQueryPattern("create table sbtest.A.*", fakedb.Result3)
	fakedbs.AddQueryPattern("create database.*", fakedb.Result3)
	fakedbs.AddQueryPattern("select.*", fakedb.Result3)
	fakedbs.AddQueryPattern("checksum.*", fakedb.Result3)

	database := "sbtest"
	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableAConfig(), router.MockTableBConfig())
	assert.Nil(t, err)

	planTree := planner.NewPlanTree()

	// DDL
	{
		query := "create table A(a int)"
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)

		plan := planner.NewDDLPlan(log, database, query, node.(*sqlparser.DDL), route)
		err = planTree.Add(plan)
		assert.Nil(t, err)
	}

	// insert
	{
		query := "insert into A(a) values(1)"
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)

		plan := planner.NewInsertPlan(log, database, query, node.(*sqlparser.Insert), route)
		err = planTree.Add(plan)
		assert.Nil(t, err)
	}

	// delete
	{
		query := "delete from A where a=2"
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)

		plan := planner.NewDeletePlan(log, database, query, node.(*sqlparser.Delete), route)
		err = planTree.Add(plan)
		assert.Nil(t, err)
	}

	// update
	{
		query := "update A set a=3 where a=2"
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)

		plan := planner.NewUpdatePlan(log, database, query, node.(*sqlparser.Update), route)
		err = planTree.Add(plan)
		assert.Nil(t, err)
	}

	// select
	{
		query := "select * from A  where a=2"
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)

		plan := planner.NewSelectPlan(log, database, query, node.(*sqlparser.Select), route)
		err = plan.Build()
		assert.Nil(t, err)
		err = planTree.Add(plan)
		assert.Nil(t, err)
	}

	// union
	{
		query := "select a from A where a=2 union select b from B"
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)

		plan := planner.NewUnionPlan(log, database, query, node.(*sqlparser.Union), route)
		err = plan.Build()
		assert.Nil(t, err)
		err = planTree.Add(plan)
		assert.Nil(t, err)
	}

	// others
	{
		query := "checksum table A"
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)

		plan := planner.NewOthersPlan(log, database, query, node, route)
		err = plan.Build()
		assert.Nil(t, err)
		err = planTree.Add(plan)
		assert.Nil(t, err)
	}

	// Execute.
	txn, err := scatter.CreateTransaction()
	assert.Nil(t, err)
	defer txn.Finish()
	executorTree := NewTree(log, planTree, txn)
	qr, err := executorTree.Execute()
	assert.Nil(t, err)
	assert.Equal(t, fakedb.Result3, qr)
}
