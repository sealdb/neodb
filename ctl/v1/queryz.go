/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package v1

import (
	"strconv"
	"time"

	"github.com/sealdb/neodb/proxy"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/sealdb/mysqlstack/xlog"
)

// QueryzHandler impl.
func QueryzHandler(log *xlog.Log, proxy *proxy.Proxy) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		queryzHandler(log, proxy, w, r)
	}
	return f
}

func queryzHandler(log *xlog.Log, proxy *proxy.Proxy, w rest.ResponseWriter, r *rest.Request) {
	type query struct {
		ConnID   uint64        `json:"connID"`
		Host     string        `json:"host"`
		Start    time.Time     `json:"start"`
		Duration time.Duration `json:"duration"`
		Color    string        `json:"color"`
		Query    string        `json:"query"`
	}

	limit := 100
	if v, err := strconv.Atoi(r.PathParam("limit")); err == nil {
		limit = v
	}

	var rsp []query
	scatter := proxy.Scatter()
	rows := scatter.Queryz().GetQueryzRows()
	for i, row := range rows {
		if i >= limit {
			break
		}
		r := query{
			ConnID:   uint64(row.ConnID),
			Host:     row.Address,
			Start:    row.Start,
			Duration: row.Duration,
			Color:    row.Color,
			Query:    row.Query,
		}
		rsp = append(rsp, r)
	}
	w.WriteJson(rsp)
}
