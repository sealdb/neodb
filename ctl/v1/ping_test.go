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
	"errors"
	"testing"

	"github.com/sealdb/neodb/proxy"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
	"github.com/sealdb/mysqlstack/xlog"
)

func TestCtlV1Ping(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("select .*", &sqltypes.Result{})
	}

	{
		// server
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Get("/v1/neodb/ping", PingHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		// client
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/v1/neodb/ping", nil))
		recorded.CodeIs(200)
	}
}

func TestCtlV1PingError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQueryError("select 1", errors.New("mock.ping.error"))
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Get("/v1/neodb/ping", PingHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	// 405.
	{
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/neodb/ping", nil))
		recorded.CodeIs(405)
	}

	// 503.
	{
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/v1/neodb/ping", nil))
		recorded.CodeIs(503)
	}
}
