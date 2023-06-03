/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package backend

import (
	"testing"

	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	scatter := NewScatter(log, "")
	// Others.
	{
		assert.NotNil(t, scatter.Queryz())
		assert.NotNil(t, scatter.Txnz())

		assert.NotNil(t, scatter.MySQLStats())
		log.Debug(scatter.MySQLStats().String())

		assert.NotNil(t, scatter.QueryStats())
		log.Debug(scatter.QueryStats().String())

		assert.NotNil(t, scatter.QueryRates())
		log.Debug(scatter.QueryRates().String())

		assert.NotNil(t, scatter.TxnCounters())
		log.Debug(scatter.TxnCounters().String())
	}
}
