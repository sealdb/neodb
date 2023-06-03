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

// handleUseDB used to handle the UseDB command.
// Here, we will send a fake query 'SELECT 1' to the backend and check the 'USE DB'.
func (spanner *Spanner) handleUseDB(session *driver.Session, query string, node *sqlparser.Use) (*sqltypes.Result, error) {
	usedb := node
	db := usedb.DBName.String()
	router := spanner.router
	// Check the database ACL.
	if err := router.DatabaseACL(db); err != nil {
		return nil, err
	}

	if _, err := spanner.ExecuteSingle(query); err != nil {
		return nil, err
	}
	session.SetSchema(db)
	return &sqltypes.Result{}, nil
}
