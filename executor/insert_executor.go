/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package executor

import (
	"github.com/sealdb/neodb/backend"
	"github.com/sealdb/neodb/planner"
	"github.com/sealdb/neodb/xcontext"

	"github.com/sealdb/mysqlstack/xlog"
)

var (
	_ Executor = &InsertExecutor{}
)

// InsertExecutor represents insert executor
type InsertExecutor struct {
	log  *xlog.Log
	plan planner.Plan
	txn  backend.Transaction
}

// NewInsertExecutor creates new insert executor.
func NewInsertExecutor(log *xlog.Log, plan planner.Plan, txn backend.Transaction) *InsertExecutor {
	return &InsertExecutor{
		log:  log,
		plan: plan,
		txn:  txn,
	}
}

// Execute used to execute the executor.
func (executor *InsertExecutor) Execute(ctx *xcontext.ResultContext) error {
	plan := executor.plan.(*planner.InsertPlan)
	reqCtx := xcontext.NewRequestContext()
	reqCtx.Mode = plan.ReqMode
	reqCtx.TxnMode = xcontext.TxnWrite
	reqCtx.Querys = plan.Querys
	reqCtx.RawQuery = plan.RawQuery

	rs, err := executor.txn.Execute(reqCtx)
	if err != nil {
		return err
	}
	ctx.Results = rs
	return nil
}
