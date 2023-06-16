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
	"sort"

	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
)

// handleOptimizeTable used to handle the 'Optimize TABLE ...' command.
// +--------------+----------+----------+-------------------------------------------------------------------+
// | Table        | Op       | Msg_type | Msg_text                                                          |
// +--------------+----------+----------+-------------------------------------------------------------------+
// | test.t       | optimize | note     | Table does not support optimize, doing recreate + analyze instead |
// | test.t       | optimize | status   | OK                                                                |
// | test.t1_0001 | optimize | status   | OK                                                                |
// | test.t1_0001 | optimize | note     | Table does not support optimize, doing recreate + analyze instead |
// +--------------+----------+----------+-------------------------------------------------------------------+
func (spanner *Spanner) handleOptimizeTable(session *driver.Session, query string, node sqlparser.Statement) (*sqltypes.Result, error) {
	database := session.Schema()
	optimize := node.(*sqlparser.Optimize)
	newqr := &sqltypes.Result{}

	for _, tbl := range optimize.Tables {
		// Construct a new sql with check one table one time, we'll send single table to backends.
		newNode := *optimize
		newNode.Tables = sqlparser.TableNames{tbl}
		qr, err := spanner.ExecuteNormal(session, database, sqlparser.String(&newNode), &newNode)
		if err != nil {
			return nil, err
		}
		newqr.AppendResult(qr)
	}

	// 1. sort by field "Table"
	sort.Slice(newqr.Rows, func(i, j int) bool {
		val := sqltypes.NullsafeCompare(newqr.Rows[i][0], newqr.Rows[j][0])
		return (-1 == val)
	})
	// 2. Formate output to mysql client, status is always displayed first. e.g.:
	// change:
	// | test.t       | optimize | note     | Table does not support optimize, doing recreate + analyze instead |
	// | test.t       | optimize | status   | OK                                                                |
	// to:
	// | test.t       | optimize | status   | OK                                                                |
	// | test.t       | optimize | note     | Table does not support optimize, doing recreate + analyze instead |

	for i := 0; i < len(newqr.Rows); i += 2 {
		j := i + 1
		if -1 == sqltypes.NullsafeCompare(newqr.Rows[i][2], newqr.Rows[j][2]) {
			newqr.Rows[i][2], newqr.Rows[j][2] = newqr.Rows[j][2], newqr.Rows[i][2]
			newqr.Rows[i][3], newqr.Rows[j][3] = newqr.Rows[j][3], newqr.Rows[i][3]
		}
	}
	return newqr, nil
}

// handleCheckTable used to handle the 'Check TABLE ...' command.
// mysql> check table t,t1 quick extended;
// +---------+-------+----------+----------+
// | Table   | Op    | Msg_type | Msg_text |
// +---------+-------+----------+----------+
// | test.t  | check | status   | OK       |
// | test.t1 | check | status   | OK       |
// +---------+-------+----------+----------+
// 2 rows in set (0.00 sec)
func (spanner *Spanner) handleCheckTable(session *driver.Session, query string, node sqlparser.Statement) (*sqltypes.Result, error) {
	database := session.Schema()
	check := node.(*sqlparser.Check)
	newqr := &sqltypes.Result{}

	for _, tbl := range check.Tables {
		// Construct a new sql with check one table one time, we'll send single table to backends.
		newNode := *check
		newNode.Tables = sqlparser.TableNames{tbl}
		qr, err := spanner.ExecuteNormal(session, database, sqlparser.String(&newNode), &newNode)
		if err != nil {
			return nil, err
		}
		newqr.AppendResult(qr)
	}

	// 1. sort by field "Table"
	sort.Slice(newqr.Rows, func(i, j int) bool {
		val := sqltypes.NullsafeCompare(newqr.Rows[i][0], newqr.Rows[j][0])
		return (-1 == val)
	})
	return newqr, nil
}
