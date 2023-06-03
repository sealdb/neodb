/*
 * NeoDB
 *
 * Copyright 2020 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package proxy

import (
	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
)

// handleDo used to handle the DO command.
func (spanner *Spanner) handleDo(session *driver.Session, query string, node sqlparser.Statement) (*sqltypes.Result, error) {
	return spanner.ExecuteSingle(query)
}
