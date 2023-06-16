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
	"time"

	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
)

type mode int

const (
	// R enum.
	R mode = iota
	// W enum.
	W
)

func (spanner *Spanner) auditLog(session *driver.Session, m mode, typ string, query string, qr *sqltypes.Result, status uint16) error {
	adit := spanner.audit
	user := session.User()
	host := session.Addr()
	connID := session.ID()
	affected := uint64(0)
	if qr != nil {
		affected = qr.RowsAffected
	}
	now := time.Now().UTC()
	switch m {
	case R:
		adit.LogReadEvent(typ, user, host, connID, query, status, affected, now)
	case W:
		adit.LogWriteEvent(typ, user, host, connID, query, status, affected, now)
	}
	return nil
}
