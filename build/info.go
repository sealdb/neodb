/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package build

import (
	"fmt"
	"runtime"
)

var (
	mysqlVer = "5.7.25"  // the backend MySQL version
	tag      = "unknown" // tag of this build
	git      string      // git hash
	time     string      // build time
	platform = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
)

// Info tuple.
type Info struct {
	Tag       string
	Time      string
	Git       string
	GoVersion string
	Platform  string
}

// GetInfo returns the info.
func GetInfo() Info {
	return Info{
		GoVersion: runtime.Version(),
		Tag:       mysqlVer + "-" + tag,
		Time:      time,
		Git:       git,
		Platform:  platform,
	}
}
