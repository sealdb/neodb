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
	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
)

// handleSelect used to handle the select command.
func (spanner *Spanner) handleSelect(session *driver.Session, query string, node sqlparser.Statement) (*sqltypes.Result, error) {
	database := session.Schema()
	return spanner.ExecuteDML(session, database, query, node)
}

func (spanner *Spanner) handleSelectStream(session *driver.Session, query string, node sqlparser.Statement, callback func(qr *sqltypes.Result) error) error {
	database := session.Schema()
	return spanner.ExecuteStreamFetch(session, database, query, node, callback)
}
