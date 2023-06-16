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
	"time"

	"github.com/sealdb/neodb/fakedb"

	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestQueryz(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	// MySQL Server starts...
	fakedb := fakedb.New(log, 1)
	defer fakedb.Close()
	addr := fakedb.Addrs()[0]
	conf := MockBackendConfigDefault(addr, addr)
	pool := NewPool(log, conf, addr)

	querys := []string{
		"SELECT1",
		"SELECT2",
	}

	// conn1
	conn1 := NewConnection(log, pool)
	err := conn1.Dial()
	assert.Nil(t, err)

	// conn2
	conn2 := NewConnection(log, pool)
	err = conn2.Dial()
	assert.Nil(t, err)

	// set conds
	fakedb.AddQueryDelay(querys[0], result1, 200)
	fakedb.AddQueryDelay(querys[1], result1, 205)

	// QueryRows
	{
		e1 := func(q string) {
			_, err := conn1.Execute(q)
			assert.Nil(t, err)
		}

		e2 := func(q string) {
			_, err := conn2.Execute(q)
			assert.Nil(t, err)
		}
		go e1(querys[0])
		time.Sleep(100 * time.Millisecond)
		go e2(querys[1])

		time.Sleep(50 * time.Millisecond)
		rows := qz.GetQueryzRows()
		assert.Equal(t, querys[0], rows[0].Query)
		assert.Equal(t, querys[1], rows[1].Query)
		// Test byStartTime.Swap() funciton to improve test coverage.
		rows.Swap(0, 1)
		assert.Equal(t, querys[0], rows[1].Query)
		assert.Equal(t, querys[1], rows[0].Query)
	}
}
