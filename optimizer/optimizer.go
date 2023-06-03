/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package optimizer

import (
	"github.com/sealdb/neodb/planner"
)

// Optimizer interface.
type Optimizer interface {
	BuildPlanTree() (*planner.PlanTree, error)
}
