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
	_ Executor = &UpdateExecutor{}
)

// UpdateExecutor represents update executor
type UpdateExecutor struct {
	log  *xlog.Log
	plan planner.Plan
	txn  backend.Transaction
}

// NewUpdateExecutor creates the new update executor.
func NewUpdateExecutor(log *xlog.Log, plan planner.Plan, txn backend.Transaction) *UpdateExecutor {
	return &UpdateExecutor{
		log:  log,
		plan: plan,
		txn:  txn,
	}
}

// Execute used to execute the executor.
func (executor *UpdateExecutor) Execute(ctx *xcontext.ResultContext) error {
	plan := executor.plan.(*planner.UpdatePlan)
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
