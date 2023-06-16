/*
 * NeoDB
 *
 * Copyright 2018-2019 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package router

import (
	"sort"
	"testing"

	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	list := NewList(log, MockTableSConfig())
	{
		err := list.Build()
		assert.Nil(t, err)
		assert.Equal(t, string(list.Type()), MethodTypeList)
	}

	{
		parts, err := list.Lookup(nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(parts))
	}

	{
		err := list.Clear()
		assert.Nil(t, err)
	}
}

func TestListLookup(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	list := NewList(log, MockTableListConfig())
	{
		err := list.Build()
		assert.Nil(t, err)
	}

	// int
	intVal := sqlparser.NewIntVal([]byte("1"))
	{
		parts, err := list.Lookup(intVal, intVal)
		assert.Nil(t, err)
		assert.Equal(t, string(list.Type()), MethodTypeList)
		assert.Equal(t, 1, len(parts))
		assert.Equal(t, "L_0000", parts[0].Table)
		assert.Equal(t, "backend1", parts[0].Backend)
	}
	intVal = sqlparser.NewIntVal([]byte("2"))
	{
		_, err := list.Lookup(intVal, intVal)
		assert.NotNil(t, err)
	}

	// float
	floatVal := sqlparser.NewFloatVal([]byte("65536.99999"))
	{
		_, err := list.Lookup(intVal, floatVal)
		assert.NotNil(t, err)
	}

	// [nil, endKey]
	{
		parts, err := list.Lookup(nil, intVal)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(parts))
	}

	// [nil, nil]
	{
		parts, err := list.Lookup(nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(parts))
	}

	// [start, end)
	{
		s := sqlparser.NewIntVal([]byte("16"))
		e := sqlparser.NewIntVal([]byte("17"))

		parts, err := list.Lookup(s, e)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(parts))
		sort.Sort(Segments(parts))
	}
}
