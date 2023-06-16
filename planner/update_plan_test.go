/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package planner

import (
	"github.com/sealdb/neodb/router"
	"testing"

	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePlan(t *testing.T) {
	results := []string{
		`{
	"RawQuery": "update sbtest.A set val = 1 where id = 1",
	"Partitions": [
		{
			"Query": "update sbtest.A6 set val = 1 where id = 1",
			"Backend": "backend6",
			"Range": "[512-4096)"
		}
	]
}`,
		`{
	"RawQuery": "update sbtest.A set val = 1 where id = id2 and id = 1",
	"Partitions": [
		{
			"Query": "update sbtest.A6 set val = 1 where id = id2 and id = 1",
			"Backend": "backend6",
			"Range": "[512-4096)"
		}
	]
}`,
		`{
	"RawQuery": "update sbtest.A set val = 1 where id in (1, 2)",
	"Partitions": [
		{
			"Query": "update sbtest.A6 set val = 1 where id in (1, 2)",
			"Backend": "backend6",
			"Range": "[512-4096)"
		}
	]
}`}
	querys := []string{
		"update sbtest.A set val = 1 where id = 1",
		"update sbtest.A set val = 1 where id = id2 and id = 1",
		"update sbtest.A set val = 1 where id in (1, 2)",
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig())
	assert.Nil(t, err)
	planTree := NewPlanTree()
	for i, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		plan := NewUpdatePlan(log, database, query, node.(*sqlparser.Update), route)

		// plan build
		{
			err := plan.Build()
			assert.Nil(t, err)
			{
				err := planTree.Add(plan)
				assert.Nil(t, err)
			}
			got := plan.JSON()
			want := results[i]
			assert.Equal(t, want, got)
			assert.Equal(t, PlanTypeUpdate, plan.Type())
		}
	}
}
func TestUpdateUnsupportedPlan(t *testing.T) {
	querys := []string{
		"update sbtest.A set a=3",
		"update sbtest.A set id=3 where id=1",
		"update sbtest.A set b=3 where id in (select id from t1)",
	}

	results := []string{
		"unsupported: missing.where.clause.in.DML",
		"unsupported: cannot.update.shard.key",
		"unsupported: subqueries.in.update",
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig())
	assert.Nil(t, err)
	for i, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		plan := NewUpdatePlan(log, database, query, node.(*sqlparser.Update), route)

		// plan build
		{
			err := plan.Build()
			want := results[i]
			got := err.Error()
			assert.Equal(t, want, got)
		}
	}
}

func TestUpdateWithNoDatabase(t *testing.T) {
	query := "update A set b=3 where id in (select id from t1)"

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig())
	assert.Nil(t, err)

	node, err := sqlparser.Parse(query)
	assert.Nil(t, err)

	databaseNull := ""
	plan := NewUpdatePlan(log, databaseNull, query, node.(*sqlparser.Update), route)

	// plan build
	{
		err := plan.Build()
		assert.NotNil(t, err)
	}
}

func TestUpdatePlanError(t *testing.T) {
	query := "update A set b=3"

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig())
	assert.Nil(t, err)

	node, err := sqlparser.Parse(query)
	assert.Nil(t, err)

	databaseNull := ""
	plan := NewUpdatePlan(log, databaseNull, query, node.(*sqlparser.Update), route)

	// plan build
	{
		err := plan.Build()
		assert.NotNil(t, err)
	}
}

func TestUpdateShardKey(t *testing.T) {
	query := "update sbtest.A set id = 1 where id = 2"

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig())
	assert.Nil(t, err)

	node, err := sqlparser.Parse(query)
	assert.Nil(t, err)

	databaseNull := ""
	plan := NewUpdatePlan(log, databaseNull, query, node.(*sqlparser.Update), route)

	// plan build
	{
		err := plan.Build()
		assert.NotNil(t, err)
	}
}

func TestUpdateNoDatabase(t *testing.T) {
	query := "update A set b = 1 where id = 2"

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig())
	assert.Nil(t, err)

	node, err := sqlparser.Parse(query)
	assert.Nil(t, err)

	databaseNull := ""
	plan := NewUpdatePlan(log, databaseNull, query, node.(*sqlparser.Update), route)

	// plan build
	{
		err := plan.Build()
		assert.NotNil(t, err)
	}
}

func TestUpdateDatabaseNotFound(t *testing.T) {
	query := "update sbtest_xxx.A set val = 1 where id = 1"

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig())
	assert.Nil(t, err)

	node, err := sqlparser.Parse(query)
	assert.Nil(t, err)

	plan := NewUpdatePlan(log, database, query, node.(*sqlparser.Update), route)

	// plan build
	{
		err := plan.Build()
		assert.NotNil(t, err)
	}
}
