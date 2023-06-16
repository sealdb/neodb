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
	"testing"

	"github.com/sealdb/neodb/router"

	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestPlanner(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	database := "xx"
	query := "create table A(a int)"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableAConfig())
	assert.Nil(t, err)

	node, err := sqlparser.Parse(query)
	assert.Nil(t, err)
	DDL := NewDDLPlan(log, database, query, node.(*sqlparser.DDL), route)

	{
		planTree := NewPlanTree()
		for i := 0; i < 64; i++ {
			err := planTree.Add(DDL)
			assert.Nil(t, err)
		}
		err := planTree.Build()
		assert.Nil(t, err)
		planSize := planTree.Size()
		log.Info("planSize: %v", planSize)
		len := len(planTree.Plans())
		assert.Equal(t, 64, len)
	}
}

func TestPlannerError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	database := "xx"
	query := "create table A(a int)"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableAConfig())
	assert.Nil(t, err)

	node, err := sqlparser.Parse(query)
	assert.Nil(t, err)
	database1 := ""
	DDL := NewDDLPlan(log, database1, query, node.(*sqlparser.DDL), route)

	{
		planTree := NewPlanTree()
		for i := 0; i < 64; i++ {
			err := planTree.Add(DDL)
			assert.Nil(t, err)
		}
		err := planTree.Build()
		assert.NotNil(t, err)
	}
}
