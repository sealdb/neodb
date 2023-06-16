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

// Operator interface.
type Operator interface {
	Execute(*xcontext.ResultContext) error
}

// ExecSubPlan used to execute all the children plan.
func ExecSubPlan(log *xlog.Log, node builder.PlanNode, ctx *xcontext.ResultContext) error {
	subPlanTree := node.Children()
	if subPlanTree != nil {
		for _, subPlan := range subPlanTree {
			switch subPlan.Type() {
			case builder.ChildTypeAggregate:
				aggrOperator := NewAggregateOperator(log, subPlan)
				if err := aggrOperator.Execute(ctx); err != nil {
					return err
				}
			case builder.ChildTypeOrderby:
				orderByOperator := NewOrderByOperator(log, subPlan)
				if err := orderByOperator.Execute(ctx); err != nil {
					return err
				}
			case builder.ChildTypeLimit:
				limitOperator := NewLimitOperator(log, subPlan)
				if err := limitOperator.Execute(ctx); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
