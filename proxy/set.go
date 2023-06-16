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
	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
)

const (
	var_mysql_autocommit      = "autocommit"
	var_neodb_streaming_fetch = "neodb_streaming_fetch"
)

// handleSet used to handle the SET command.
func (spanner *Spanner) handleSet(session *driver.Session, query string, node *sqlparser.Set) (*sqltypes.Result, error) {
	log := spanner.log
	txSession := spanner.sessions.getTxnSession(session)

	for _, expr := range node.Exprs {
		name := expr.Type.Lowered()
		if strings.HasPrefix(name, "@@session.") {
			name = strings.TrimPrefix(name, "@@session.")
		}

		switch name {
		case var_neodb_streaming_fetch:
			switch expr := expr.Val.(*sqlparser.OptVal).Value.(type) {
			case *sqlparser.SQLVal:
				switch expr.Type {
				case sqlparser.StrVal:
					val := strings.ToLower(string(expr.Val))
					switch val {
					case "on":
						txSession.setStreamingFetchVar(true)
					case "off":
						txSession.setStreamingFetchVar(false)
					}
				default:
					return nil, fmt.Errorf("Invalid value type: %v", sqlparser.String(expr))
				}
			case sqlparser.BoolVal:
				if expr {
					txSession.setStreamingFetchVar(true)
				} else {
					txSession.setStreamingFetchVar(false)
				}
			}

		case var_mysql_autocommit:
			var autocommit = true

			switch expr := expr.Val.(*sqlparser.OptVal).Value.(type) {
			case *sqlparser.SQLVal:
				switch expr.Type {
				case sqlparser.IntVal:
					if expr.Val[0] == '0' {
						autocommit = false
					}
				case sqlparser.StrVal:
					if strings.ToLower(string(expr.Val)) == "off" {
						autocommit = false
					}
				}
			case sqlparser.BoolVal:
				if !expr {
					autocommit = false
				}
			}
			if !autocommit && spanner.isAutocommitFalseIsTxn() {
				query := "begin"
				node := &sqlparser.Transaction{
					Action: "begin",
				}
				qr, err := spanner.handleMultiStmtTxn(session, query, node)
				if err != nil {
					log.Error("proxy.transaction[%s](by.autocommit).from.session[%v].error:%+v", query, session.ID(), err)
					return nil, err
				}
				return qr, nil
			}
		default:
			log.Warning("unhandle.set[%v]:%v", name, query)
		}
	}
	qr := &sqltypes.Result{Warnings: 1}
	return qr, nil
}
