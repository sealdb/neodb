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

func TestGlobal(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	global := NewGlobal(log, MockTableGConfig())
	{
		err := global.Build()
		assert.Nil(t, err)
		assert.Equal(t, string(global.Type()), MethodTypeGlobal)
	}

	{
		parts, err := global.Lookup(nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(parts))
	}

	global = NewGlobal(log, MockTableG1Config())
	{
		err := global.Build()
		assert.Nil(t, err)
		assert.Equal(t, string(global.Type()), MethodTypeGlobal)
	}

	{
		parts, err := global.Lookup(nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(parts))
	}
}
