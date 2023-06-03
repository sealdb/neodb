/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package operator

import (
	"github.com/sealdb/neodb/planner/builder"
	"github.com/sealdb/neodb/xcontext"

	"github.com/sealdb/mysqlstack/xlog"
)

var (
	_ Operator = &LimitOperator{}
)

// LimitOperator represents limit operator.
type LimitOperator struct {
	log  *xlog.Log
	plan builder.ChildPlan
}

// NewLimitOperator creates the new limit operator.
func NewLimitOperator(log *xlog.Log, plan builder.ChildPlan) *LimitOperator {
	return &LimitOperator{
		log:  log,
		plan: plan,
	}
}

// Execute used to execute the operator.
func (operator *LimitOperator) Execute(ctx *xcontext.ResultContext) error {
	rs := ctx.Results
	plan := operator.plan.(*builder.LimitPlan)
	rs.Limit(plan.Offset, plan.Limit)
	return nil
}
