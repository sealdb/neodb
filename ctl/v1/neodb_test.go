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
	"testing"

	"github.com/sealdb/neodb/proxy"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestCtlV1NeoDBConfig(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	{
		// server
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Put("/v1/neodb/config", NeoDBConfigHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		type neoDBParams1 struct {
			MaxConnections      int      `json:"max-connections"`
			MaxResultSize       int      `json:"max-result-size"`
			MaxJoinRows         int      `json:"max-join-rows"`
			DDLTimeout          int      `json:"ddl-timeout"`
			QueryTimeout        int      `json:"query-timeout"`
			TwoPCEnable         bool     `json:"twopc-enable"`
			LoadBalance         int      `json:"load-balance"`
			AllowIP             []string `json:"allowip,omitempty"`
			AuditMode           string   `json:"audit-mode"`
			StreamBufferSize    int      `json:"stream-buffer-size"`
			Blocks              int      `json:"blocks-readonly"`
			LowerCaseTableNames int      `json:"lower-case-table-names"`
		}

		// 200.
		{
			// client
			p := &neoDBParams1{
				MaxConnections:      1023,
				MaxResultSize:       1073741823,
				MaxJoinRows:         32767,
				QueryTimeout:        33,
				TwoPCEnable:         true,
				LoadBalance:         1,
				AllowIP:             []string{"127.0.0.1", "127.0.0.2"},
				AuditMode:           "A",
				StreamBufferSize:    16777216,
				Blocks:              128,
				LowerCaseTableNames: 1,
			}
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "http://localhost/v1/neodb/config", p))
			recorded.CodeIs(200)

			neoDBConf := proxy.Config()
			assert.Equal(t, 1023, neoDBConf.Proxy.MaxConnections)
			assert.Equal(t, 1073741823, neoDBConf.Proxy.MaxResultSize)
			assert.Equal(t, 32767, neoDBConf.Proxy.MaxJoinRows)
			assert.Equal(t, 0, neoDBConf.Proxy.DDLTimeout)
			assert.Equal(t, 33, neoDBConf.Proxy.QueryTimeout)
			assert.Equal(t, true, neoDBConf.Proxy.TwopcEnable)
			assert.Equal(t, 1, neoDBConf.Proxy.LoadBalance)
			assert.Equal(t, []string{"127.0.0.1", "127.0.0.2"}, neoDBConf.Proxy.IPS)
			assert.Equal(t, "A", neoDBConf.Audit.Mode)
			assert.Equal(t, 16777216, neoDBConf.Proxy.StreamBufferSize)
			assert.Equal(t, 128, neoDBConf.Router.Blocks)
			assert.Equal(t, 1, neoDBConf.Proxy.LowerCaseTableNames)
		}

		// Unset AllowIP.
		{
			// client
			p := &neoDBParams1{
				MaxConnections:   1023,
				MaxResultSize:    1073741824,
				MaxJoinRows:      32768,
				QueryTimeout:     33,
				TwoPCEnable:      true,
				AuditMode:        "A",
				StreamBufferSize: 67108864,
			}
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "http://localhost/v1/neodb/config", p))
			recorded.CodeIs(200)

			neoDBConf := proxy.Config()
			assert.Equal(t, 1023, neoDBConf.Proxy.MaxConnections)
			assert.Equal(t, 1073741824, neoDBConf.Proxy.MaxResultSize)
			assert.Equal(t, 32768, neoDBConf.Proxy.MaxJoinRows)
			assert.Equal(t, 0, neoDBConf.Proxy.DDLTimeout)
			assert.Equal(t, 33, neoDBConf.Proxy.QueryTimeout)
			assert.Equal(t, true, neoDBConf.Proxy.TwopcEnable)
			assert.Nil(t, neoDBConf.Proxy.IPS)
			assert.Equal(t, "A", neoDBConf.Audit.Mode)
			assert.Equal(t, 67108864, neoDBConf.Proxy.StreamBufferSize)
		}
	}
}

func TestCtlV1NeoDBConfigError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	{
		// server
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Put("/v1/neodb/config", NeoDBConfigHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		// 405.
		{
			p := &neoDBParams{}
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/neodb/config", p))
			recorded.CodeIs(405)
		}

		// 500.
		{
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "http://localhost/v1/neodb/config", nil))
			recorded.CodeIs(500)
		}
	}
}

func TestCtlV1NeoDBReadOnly(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	{
		// server
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Put("/v1/neodb/readonly", ReadonlyHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		// 200.
		{
			// client
			p := &readonlyParams{
				ReadOnly: true,
			}
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "http://localhost/v1/neodb/readonly", p))
			recorded.CodeIs(200)
		}
	}
}

func TestCtlV1ReadOnlyError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	{
		// server
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Put("/v1/neodb/readonly", ReadonlyHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		// 405.
		{
			p := &readonlyParams{}
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/neodb/readonly", p))
			recorded.CodeIs(405)
		}

		// 500.
		{
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "http://localhost/v1/neodb/readonly", nil))
			recorded.CodeIs(500)
		}
	}
}

func TestCtlV1NeoDBTwopc(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	{
		// server
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Put("/v1/neodb/twopc", TwopcHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		// 200.
		{
			// client
			p := &twopcParams{
				Twopc: true,
			}
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "http://localhost/v1/neodb/twopc", p))
			recorded.CodeIs(200)
		}
	}
}

func TestCtlV1TwopcError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	{
		// server
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Put("/v1/neodb/twopc", ReadonlyHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		// 405.
		{
			p := &twopcParams{}
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/neodb/twopc", p))
			recorded.CodeIs(405)
		}

		// 500.
		{
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "http://localhost/v1/neodb/twopc", nil))
			recorded.CodeIs(500)
		}
	}
}

func TestCtlV1NeoDBThrottle(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	{
		// server
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Put("/v1/neodb/throttle", ThrottleHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		// 200.
		{
			// client
			p := &throttleParams{
				Limits: 100,
			}
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "http://localhost/v1/neodb/throttle", p))
			recorded.CodeIs(200)
		}
	}
}

func TestCtlV1NeoDBThrottleError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	{
		// server
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Put("/v1/neodb/throttle", ThrottleHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		// 405.
		{
			p := &throttleParams{}
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/neodb/throttle", p))
			recorded.CodeIs(405)
		}

		// 500.
		{
			recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("PUT", "http://localhost/v1/neodb/throttle", nil))
			recorded.CodeIs(500)
		}
	}
}

func TestCtlV1NeoDBStatus(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()
	address := proxy.Address()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("create .*", &sqltypes.Result{})
	}

	// create database.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create database test"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	// create test table.
	{
		client, err := driver.NewConn("mock", "mock", address, "", "utf8")
		assert.Nil(t, err)
		query := "create table test.t1(id int, b int) partition by hash(id)"
		_, err = client.FetchAll(query, -1)
		assert.Nil(t, err)
	}

	{
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Get("/v1/neodb/status", StatusHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/v1/neodb/status", nil))
		recorded.CodeIs(200)

		want := "{\"readonly\":false}"
		got := recorded.Recorder.Body.String()
		assert.Equal(t, want, got)
	}
}

func TestCtlV1NeoDBApiAddress(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	{
		api := rest.NewApi()
		router, _ := rest.MakeRouter(
			rest.Get("/v1/neodb/restapiaddress", RestAPIAddressHandler(log, proxy)),
		)
		api.SetApp(router)
		handler := api.MakeHandler()

		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/v1/neodb/restapiaddress", nil))
		recorded.CodeIs(200)
	}
}
