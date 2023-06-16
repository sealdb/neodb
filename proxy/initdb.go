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
	"fmt"
	"strings"

	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqldb"
)

// ComInitDB impl.
// Here, we will send a fake query 'SELECT 1' to the backend and check the 'USE DB'.
func (spanner *Spanner) ComInitDB(session *driver.Session, database string) error {
	router := spanner.router

	// Check the database ACL.
	if err := router.DatabaseACL(database); err != nil {
		return err
	}

	if spanner.isLowerCaseTableNames() {
		database = strings.ToLower(database)
	}

	privilegePlug := spanner.plugins.PlugPrivilege()
	isSet := privilegePlug.CheckUserPrivilegeIsSet(session.User())
	if !isSet {
		isSuper := privilegePlug.IsSuperPriv(session.User())
		if !isSuper {
			if isExist := privilegePlug.CheckDBinUserPrivilege(session.User(), database); !isExist {
				error := sqldb.NewSQLErrorf(sqldb.ER_DBACCESS_DENIED_ERROR, "Access denied for user '%v'@'%%' to database '%v'",
					session.User(), database)
				return error
			}
		}
	}

	query := fmt.Sprintf("use %s", database)
	if _, err := spanner.ExecuteSingle(query); err != nil {
		return err
	}
	session.SetSchema(database)
	return nil
}
