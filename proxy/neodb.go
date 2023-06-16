/*
 * NeoDB
 *
 * Copyright 2018-2019 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package proxy

import (
	"github.com/pkg/errors"
	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqldb"
	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/sqlparser/depends/common"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
)

// handleNeoDB used to handle the command: neodb attach/detach/attachlist.
func (spanner *Spanner) handleNeoDB(session *driver.Session, query string, node sqlparser.Statement) (*sqltypes.Result, error) {
	var err error
	var qr *sqltypes.Result
	log := spanner.log
	attach := NewAttach(log, spanner.scatter, spanner.router, spanner)

	snode := node.(*sqlparser.NeoDB)
	row := snode.Row
	var attachName string

	if row != nil {
		if len(row) != DetachParamsCount && len(row) != AttachParamsCount {
			return nil, errors.Errorf("spanner.query.execute.neodb.%s.error,", snode.Action)
		}

		if len(row) == DetachParamsCount {
			val, _ := row[0].(*sqlparser.SQLVal)
			attachName = common.BytesToString(val.Val)
		}
	}

	switch snode.Action {
	case sqlparser.AttachStr:
		qr, err = attach.Attach(snode)
	case sqlparser.DetachStr:
		qr, err = attach.Detach(attachName)
	case sqlparser.AttachListStr:
		qr, err = attach.ListAttach()
	case sqlparser.ReshardStr:
		table := snode.Table.Name.String()
		database := session.Schema()
		if !snode.Table.Qualifier.IsEmpty() {
			database = snode.Table.Qualifier.String()
		}

		newTable := snode.NewName.Name.String()
		newDatabase := session.Schema()
		if !snode.NewName.Qualifier.IsEmpty() {
			newDatabase = snode.NewName.Qualifier.String()
		}

		reshard := NewReshard(log, spanner.scatter, spanner.router, spanner, session.User())
		reshard.SetHandle(reshard)
		qr, err = reshard.ReShardTable(database, table, newDatabase, newTable)
	case sqlparser.CleanupStr:
		cleanup := NewCleanup(log, spanner.scatter, spanner.router, spanner)
		qr, err = cleanup.Cleanup()
	case sqlparser.RebalanceStr:
		rebalance := NewRebalance(log, spanner.scatter, spanner.router, spanner, spanner.conf, spanner.plugins)
		qr, err = rebalance.Rebalance()
	case sqlparser.XARecoverStr:
		adminXA := NewAdminXA(log, spanner.scatter, spanner.router, spanner)
		qr, err = adminXA.Recover()
	case sqlparser.XACommitStr:
		adminXA := NewAdminXA(log, spanner.scatter, spanner.router, spanner)
		qr, err = adminXA.Commit()
	case sqlparser.XARollbackStr:
		adminXA := NewAdminXA(log, spanner.scatter, spanner.router, spanner)
		qr, err = adminXA.Rollback()
	default:
		log.Error("proxy.neodb.unsupported[%s]", query)
		err = sqldb.NewSQLErrorf(sqldb.ER_UNKNOWN_ERROR, "unsupported.query: %v", query)
	}
	if err != nil {
		log.Error("proxy.query.neodb.[%s].error:%s", query, err)
		return nil, err
	}
	return qr, err
}
