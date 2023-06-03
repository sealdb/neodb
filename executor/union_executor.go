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
	"github.com/sealdb/neodb/executor/engine"
	"github.com/sealdb/neodb/planner"
	"github.com/sealdb/neodb/xcontext"

	"github.com/sealdb/mysqlstack/xlog"
)

var (
	_ Executor = &UnionExecutor{}
)

// UnionExecutor represents select executor
type UnionExecutor struct {
	log  *xlog.Log
	plan planner.Plan
	txn  backend.Transaction
}

// NewUnionExecutor creates the new select executor.
func NewUnionExecutor(log *xlog.Log, plan planner.Plan, txn backend.Transaction) *UnionExecutor {
	return &UnionExecutor{
		log:  log,
		plan: plan,
		txn:  txn,
	}
}

// Execute used to execute the executor.
func (executor *UnionExecutor) Execute(ctx *xcontext.ResultContext) error {
	log := executor.log
	plan := executor.plan.(*planner.UnionPlan)
	planEngine := engine.BuildEngine(log, plan.Root, executor.txn)
	if err := planEngine.Execute(ctx); err != nil {
		return err
	}
	return nil
}
