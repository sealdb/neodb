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
	_ Executor = &DeleteExecutor{}
)

// DeleteExecutor represents delete executor
type DeleteExecutor struct {
	log  *xlog.Log
	plan planner.Plan
	txn  backend.Transaction
}

// NewDeleteExecutor creates new delete executor.
func NewDeleteExecutor(log *xlog.Log, plan planner.Plan, txn backend.Transaction) *DeleteExecutor {
	return &DeleteExecutor{
		log:  log,
		plan: plan,
		txn:  txn,
	}
}

// Execute used to execute the executor.
func (executor *DeleteExecutor) Execute(ctx *xcontext.ResultContext) error {
	plan := executor.plan.(*planner.DeletePlan)
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
