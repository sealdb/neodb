/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package proxy

import (
	"testing"

	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestDiskCheck(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	dc := NewDiskCheck(log, "/tmp/")
	err := dc.Init()
	assert.Nil(t, err)
	defer dc.Close()

	dc.doCheck()
	high := dc.HighWater()
	assert.False(t, high)
}
