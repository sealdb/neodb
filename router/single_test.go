/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package router

import (
	"testing"

	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestSingle(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	single := NewSingle(log, MockTableSConfig())
	{
		err := single.Build()
		assert.Nil(t, err)
		assert.Equal(t, string(single.Type()), MethodTypeSingle)
	}

	{
		parts, err := single.Lookup(nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(parts))
	}
}
