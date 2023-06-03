/*
 * NeoDB
 *
 * Copyright 2018-2019 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package plugins

import (
	"testing"

	"github.com/sealdb/neodb/backend"
	"github.com/sealdb/neodb/plugins/privilege"

	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestPlugins(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	//Create scatter and query handler.
	scatter, fakedbs, cleanup := backend.MockScatter(log, 10)
	defer cleanup()

	privilege.MockInitPrivilegeY(fakedbs)

	plugin := NewPlugin(log, nil, nil, scatter)
	err := plugin.Init()
	assert.Nil(t, err)
	defer plugin.Close()

	autoincPlug := plugin.PlugAutoIncrement()
	assert.NotNil(t, autoincPlug)

	privilegePlug := plugin.PlugPrivilege()
	assert.NotNil(t, privilegePlug)

	shiftMgrPlug := plugin.PlugShiftMgr()
	assert.NotNil(t, shiftMgrPlug)
}
