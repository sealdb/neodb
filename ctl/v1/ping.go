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
	"net/http"

	"github.com/sealdb/neodb/proxy"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/sealdb/mysqlstack/xlog"
)

// PingHandler impl.
func PingHandler(log *xlog.Log, proxy *proxy.Proxy) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		pingHandler(log, proxy, w, r)
	}
	return f
}

func pingHandler(log *xlog.Log, proxy *proxy.Proxy, w rest.ResponseWriter, r *rest.Request) {
	spanner := proxy.Spanner()
	if _, err := spanner.ExecuteScatter("select 1"); err != nil {
		log.Error("api.v1.ping.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
}
