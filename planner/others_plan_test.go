/*
 * NeoDB
 *
 * Copyright 2018-2019 The Radon Authors.
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

func TestOthersPlanChecksumTable(t *testing.T) {
	results := []string{
		`{
	"RawQuery": "checksum table A",
	"Partitions": [
		{
			"Query": "checksum table sbtest.A1",
			"Backend": "backend1",
			"Range": "0-32"
		},
		{
			"Query": "checksum table sbtest.A2",
			"Backend": "backend2",
			"Range": "32-64"
		},
		{
			"Query": "checksum table sbtest.A3",
			"Backend": "backend3",
			"Range": "64-96"
		},
		{
			"Query": "checksum table sbtest.A4",
			"Backend": "backend4",
			"Range": "96-256"
		},
		{
			"Query": "checksum table sbtest.A5",
			"Backend": "backend5",
			"Range": "256-512"
		},
		{
			"Query": "checksum table sbtest.A6",
			"Backend": "backend6",
			"Range": "512-4096"
		}
	]
}`,
		`{
	"RawQuery": "checksum table sbtest.A",
	"Partitions": [
		{
			"Query": "checksum table sbtest.A1",
			"Backend": "backend1",
			"Range": "0-32"
		},
		{
			"Query": "checksum table sbtest.A2",
			"Backend": "backend2",
			"Range": "32-64"
		},
		{
			"Query": "checksum table sbtest.A3",
			"Backend": "backend3",
			"Range": "64-96"
		},
		{
			"Query": "checksum table sbtest.A4",
			"Backend": "backend4",
			"Range": "96-256"
		},
		{
			"Query": "checksum table sbtest.A5",
			"Backend": "backend5",
			"Range": "256-512"
		},
		{
			"Query": "checksum table sbtest.A6",
			"Backend": "backend6",
			"Range": "512-4096"
		}
	]
}`,
		`{
	"RawQuery": "checksum table G",
	"Partitions": [
		{
			"Query": "checksum table sbtest.G",
			"Backend": "backend1",
			"Range": ""
		}
	]
}`,
		`{
	"RawQuery": "checksum table sbtest.S",
	"Partitions": [
		{
			"Query": "checksum table sbtest.S",
			"Backend": "backend1",
			"Range": ""
		}
	]
}`,
	}
	querys := []string{
		"checksum table A",
		"checksum table sbtest.A",
		"checksum table G",
		"checksum table sbtest.S",
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig(), router.MockTableGConfig(), router.MockTableSConfig())
	assert.Nil(t, err)
	planTree := NewPlanTree()
	for i, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		plan := NewOthersPlan(log, database, query, node, route)

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
			assert.Equal(t, PlanTypeOthers, plan.Type())
		}
	}
}

func TestOthersPlanChecksumTableError(t *testing.T) {
	querys := []string{
		"checksum table A",
		"checksum table xx.A",
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig(), router.MockTableGConfig(), router.MockTableSConfig())
	assert.Nil(t, err)
	for _, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		plan := NewOthersPlan(log, "", query, node, route)

		// plan build
		{
			err := plan.Build()
			assert.NotNil(t, err)
		}
	}
}

func TestOthersPlanOptimizeTable(t *testing.T) {
	results := []string{
		`{
	"RawQuery": "optimize table A",
	"Partitions": [
		{
			"Query": "optimize table sbtest.A1",
			"Backend": "backend1",
			"Range": "0-32"
		},
		{
			"Query": "optimize table sbtest.A2",
			"Backend": "backend2",
			"Range": "32-64"
		},
		{
			"Query": "optimize table sbtest.A3",
			"Backend": "backend3",
			"Range": "64-96"
		},
		{
			"Query": "optimize table sbtest.A4",
			"Backend": "backend4",
			"Range": "96-256"
		},
		{
			"Query": "optimize table sbtest.A5",
			"Backend": "backend5",
			"Range": "256-512"
		},
		{
			"Query": "optimize table sbtest.A6",
			"Backend": "backend6",
			"Range": "512-4096"
		}
	]
}`,
		`{
	"RawQuery": "optimize table sbtest.A",
	"Partitions": [
		{
			"Query": "optimize table sbtest.A1",
			"Backend": "backend1",
			"Range": "0-32"
		},
		{
			"Query": "optimize table sbtest.A2",
			"Backend": "backend2",
			"Range": "32-64"
		},
		{
			"Query": "optimize table sbtest.A3",
			"Backend": "backend3",
			"Range": "64-96"
		},
		{
			"Query": "optimize table sbtest.A4",
			"Backend": "backend4",
			"Range": "96-256"
		},
		{
			"Query": "optimize table sbtest.A5",
			"Backend": "backend5",
			"Range": "256-512"
		},
		{
			"Query": "optimize table sbtest.A6",
			"Backend": "backend6",
			"Range": "512-4096"
		}
	]
}`,
		`{
	"RawQuery": "optimize table G",
	"Partitions": [
		{
			"Query": "optimize table sbtest.G",
			"Backend": "backend1",
			"Range": ""
		},
		{
			"Query": "optimize table sbtest.G",
			"Backend": "backend2",
			"Range": ""
		}
	]
}`,
		`{
	"RawQuery": "optimize table sbtest.S",
	"Partitions": [
		{
			"Query": "optimize table sbtest.S",
			"Backend": "backend1",
			"Range": ""
		}
	]
}`,
	}
	querys := []string{
		"optimize table A",
		"optimize table sbtest.A",
		"optimize table G",
		"optimize table sbtest.S",
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig(), router.MockTableGConfig(), router.MockTableSConfig())
	assert.Nil(t, err)
	planTree := NewPlanTree()
	for i, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		plan := NewOthersPlan(log, database, query, node, route)

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
			assert.Equal(t, PlanTypeOthers, plan.Type())
		}
	}
}

func TestOthersPlanOptimizeTableError(t *testing.T) {
	querys := []string{
		"optimize table A",
		"optimize table xx.A",
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig(), router.MockTableGConfig(), router.MockTableSConfig())
	assert.Nil(t, err)
	for _, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		plan := NewOthersPlan(log, "", query, node, route)

		// plan build
		{
			err := plan.Build()
			assert.NotNil(t, err)
		}
	}
}

func TestOthersPlanCheckTable(t *testing.T) {
	results := []string{
		`{
	"RawQuery": "check table A",
	"Partitions": [
		{
			"Query": "check table sbtest.A1",
			"Backend": "backend1",
			"Range": "0-32"
		},
		{
			"Query": "check table sbtest.A2",
			"Backend": "backend2",
			"Range": "32-64"
		},
		{
			"Query": "check table sbtest.A3",
			"Backend": "backend3",
			"Range": "64-96"
		},
		{
			"Query": "check table sbtest.A4",
			"Backend": "backend4",
			"Range": "96-256"
		},
		{
			"Query": "check table sbtest.A5",
			"Backend": "backend5",
			"Range": "256-512"
		},
		{
			"Query": "check table sbtest.A6",
			"Backend": "backend6",
			"Range": "512-4096"
		}
	]
}`,
		`{
	"RawQuery": "check table sbtest.A",
	"Partitions": [
		{
			"Query": "check table sbtest.A1",
			"Backend": "backend1",
			"Range": "0-32"
		},
		{
			"Query": "check table sbtest.A2",
			"Backend": "backend2",
			"Range": "32-64"
		},
		{
			"Query": "check table sbtest.A3",
			"Backend": "backend3",
			"Range": "64-96"
		},
		{
			"Query": "check table sbtest.A4",
			"Backend": "backend4",
			"Range": "96-256"
		},
		{
			"Query": "check table sbtest.A5",
			"Backend": "backend5",
			"Range": "256-512"
		},
		{
			"Query": "check table sbtest.A6",
			"Backend": "backend6",
			"Range": "512-4096"
		}
	]
}`,
		`{
	"RawQuery": "check table G",
	"Partitions": [
		{
			"Query": "check table sbtest.G",
			"Backend": "backend1",
			"Range": ""
		},
		{
			"Query": "check table sbtest.G",
			"Backend": "backend2",
			"Range": ""
		}
	]
}`,
		`{
	"RawQuery": "check table sbtest.S",
	"Partitions": [
		{
			"Query": "check table sbtest.S",
			"Backend": "backend1",
			"Range": ""
		}
	]
}`,
	}
	querys := []string{
		"check table A",
		"check table sbtest.A",
		"check table G",
		"check table sbtest.S",
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig(), router.MockTableGConfig(), router.MockTableSConfig())
	assert.Nil(t, err)
	planTree := NewPlanTree()
	for i, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		plan := NewOthersPlan(log, database, query, node, route)

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
			assert.Equal(t, PlanTypeOthers, plan.Type())
		}
	}
}

func TestOthersPlanCheckTableError(t *testing.T) {
	querys := []string{
		"check table A",
		"check table xx.A",
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig(), router.MockTableGConfig(), router.MockTableSConfig())
	assert.Nil(t, err)
	for _, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		plan := NewOthersPlan(log, "", query, node, route)

		// plan build
		{
			err := plan.Build()
			assert.NotNil(t, err)
		}
	}
}
