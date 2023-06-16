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
	querypb "github.com/sealdb/mysqlstack/sqlparser/depends/query"
	"github.com/sealdb/mysqlstack/sqlparser/depends/sqltypes"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestCtlV1CreateUser(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQuery("GRANT select,insert,update,delete ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", &sqltypes.Result{})
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/v1/user/add", CreateUserHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	{
		p := &userParams{
			Databases: "*",
			User:      "mock",
			Password:  "pwd",
			Privilege: "select,insert,update,delete",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(200)
	}
}

func TestCtlV1CreateUserDatabases(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQuery("GRANT ALL ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", &sqltypes.Result{})
		fakedbs.AddQuery("GRANT ALL ON a.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", &sqltypes.Result{})
		fakedbs.AddQuery("GRANT ALL ON b.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", &sqltypes.Result{})
		fakedbs.AddQueryError("GRANT ALL ON c.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", errors.New("mock.create.user.error"))
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/v1/user/add", CreateUserHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	{
		p := &userParams{
			Databases: "",
			User:      "mock",
			Password:  "pwd",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(200)
	}

	{
		p := &userParams{
			Databases: "a,b,",
			User:      "mock",
			Password:  "pwd",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(200)
	}

	{
		p := &userParams{
			Databases: "*,a,b",
			User:      "mock",
			Password:  "pwd",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(200)
	}

	{
		p := &userParams{
			Databases: "*,a,b,c",
			User:      "mock",
			Password:  "pwd",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(503)
	}
}

func TestCtlV1CreateUserError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQueryError("GRANT ALL ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", errors.New("mock.create.user.error"))
		fakedbs.AddQueryError("GRANT privErr ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", errors.New("mock.create.user.error"))
		fakedbs.AddQueryError("GRANT ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", errors.New("mock.create.user.error"))
		fakedbs.AddQueryError("GRANT selec,insert,update ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", errors.New("mock.create.user.error"))
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/v1/user/add", CreateUserHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	{
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", nil))
		recorded.CodeIs(500)
	}

	{
		p := &userParams{
			Databases: "*",
			User:      "mock",
			Password:  "pwd",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(503)
	}

	{
		p := &userParams{
			Databases: "*",
			User:      "mock",
			Password:  "pwd",
			Privilege: "select",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(503)
	}

	{
		p := &userParams{
			Databases: "*,a,b",
			User:      "mock",
			Password:  "pwd",
			Privilege: "selec,insert,update",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(503)
	}

	{
		p := &userParams{
			Databases: "*,a,b",
			User:      "mock",
			Password:  "pwd",
			Privilege: " ",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(503)
	}

	{
		p := &userParams{
			Databases: "*,a,b",
			User:      "mock",
			Password:  "pwd",
			Privilege: "privErr",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(503)
	}
}

func TestCtlV1CreateUserError1(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	_, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/v1/user/add", CreateUserHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	{
		p := &userParams{
			Databases: "",
			User:      "",
			Password:  "pwd",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(204)
	}
}

func TestCtlV1CreateUserPriv(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQuery("GRANT ALL ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", &sqltypes.Result{})
		fakedbs.AddQuery("GRANT select,insert,update,delete ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", &sqltypes.Result{})
		fakedbs.AddQuery("GRANT select ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", &sqltypes.Result{})
		fakedbs.AddQuery("GRANT insert,update,delete ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", &sqltypes.Result{})
		fakedbs.AddQuery("GRANT delete ON *.* TO 'mock'@'%' IDENTIFIED BY 'pwd'", &sqltypes.Result{})
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/v1/user/add", CreateUserHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	{
		p := &userParams{
			Databases: "*",
			User:      "mock",
			Password:  "pwd",
			Privilege: "select,insert,update,delete",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(200)
	}

	{
		p := &userParams{
			Databases: "*",
			User:      "mock",
			Password:  "pwd",
			Privilege: "select",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(200)
	}

	{
		p := &userParams{
			Databases: "*",
			User:      "mock",
			Password:  "pwd",
			Privilege: "delete",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(200)
	}

	{
		p := &userParams{
			Databases: "*",
			User:      "mock",
			Password:  "pwd",
			Privilege: "insert,update,delete",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/add", p))
		recorded.CodeIs(200)
	}
}

func TestCtlV1AlterUser(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQuery("ALTER USER 'mock'@'%' IDENTIFIED BY 'pwd'", &sqltypes.Result{})
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/v1/user/update", AlterUserHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	{
		p := &userParams{
			User:     "mock",
			Password: "pwd",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/update", p))
		recorded.CodeIs(200)
	}
}

func TestCtlV1AlterUserError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQueryError("ALTER USER 'mock'@'%' IDENTIFIED BY 'pwd'", errors.New("mock.alter.user.error"))
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/v1/user/update", AlterUserHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	// 500.
	{
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/update", nil))
		recorded.CodeIs(500)
	}

	// 503.
	{
		p := &userParams{
			User:     "mock",
			Password: "pwd",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/update", p))
		recorded.CodeIs(503)
	}
}

func TestCtlV1DropUser(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQueryPattern("DROP USER 'mock'@'%'", &sqltypes.Result{})
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/v1/user/remove", DropUserHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	{
		p := &userParams{
			User: "mock",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/remove", p))
		recorded.CodeIs(200)
	}
}

func TestCtlV1DropError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQueryErrorPattern("DROP .*", errors.New("mock.drop.user.error"))
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/v1/user/remove", DropUserHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	// 503.
	{
		p := &userParams{
			User: "mock",
		}
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/v1/user/remove", p))
		recorded.CodeIs(503)
	}
}

func TestCtlV1Userz(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	selectResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name: "User",
				Type: querypb.Type_VARCHAR,
			},
			{
				Name: "Host",
				Type: querypb.Type_VARCHAR,
			},
			{
				Name: "Super_priv",
				Type: querypb.Type_VARCHAR,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("test1")),
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("%")),
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("Y")),
			},
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("test2")),
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("%")),
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("N")),
			},
		},
	}

	//fakedbs.
	{
		fakedbs.AddQueryPattern("select .*", selectResult)
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Get("/v1/user/userz", UserzHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	{
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/v1/user/userz", nil))
		recorded.CodeIs(200)

		want := "[{\"User\":\"test1\",\"Host\":\"%\",\"SuperPriv\":\"Y\"},{\"User\":\"test2\",\"Host\":\"%\",\"SuperPriv\":\"N\"}]"
		got := recorded.Recorder.Body.String()
		log.Debug(got)
		assert.Equal(t, want, got)
	}
}

func TestCtlV1UserzError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	fakedbs, proxy, cleanup := proxy.MockProxy(log)
	defer cleanup()

	// fakedbs.
	{
		fakedbs.AddQueryErrorPattern("select .*", errors.New("api.v1.userz.get.mysql.user.error"))
	}

	// server
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Get("/v1/user/userz", UserzHandler(log, proxy)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	{
		recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/v1/user/userz", nil))
		recorded.CodeIs(503)
	}
}
