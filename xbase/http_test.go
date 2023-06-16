/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package xbase

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestHttpGet(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	httpSvr := mockHTTP(log, ":8888")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		httpSvr.Shutdown(ctx)
	}()

	url := "http://127.0.0.1:8888/test/getok"
	body, err := HTTPGet(url)
	assert.Nil(t, err)
	log.Debug("%#v", body)
}

func TestHttpPost(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	httpSvr := mockHTTP(log, ":7888")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		httpSvr.Shutdown(ctx)
	}()

	url := "http://127.0.0.1:7888/test/ok"
	type request struct {
	}
	resp, cleanup, err := HTTPPost(url, &request{})
	assert.Nil(t, err)
	defer cleanup()
	log.Debug("%#v", resp)
}

func TestHttpPostTimeout(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	httpSvr := mockHTTP(log, ":8889")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		httpSvr.Shutdown(ctx)
	}()

	url := "http://127.0.0.1:8889/test/timeout"
	want := "Get \"http://127.0.0.1:8889/test/timeout\": context deadline exceeded"
	_, err := HTTPGet(url)
	got := err.Error()
	assert.Equal(t, want, got)
}

func TestHttpPut(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	httpSvr := mockHTTP(log, ":8888")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		httpSvr.Shutdown(ctx)
	}()

	url := "http://127.0.0.1:8888/test/putok"
	type request struct {
	}
	resp, cleanup, err := HTTPPut(url, &request{})
	assert.Nil(t, err)
	defer cleanup()
	log.Debug("%#v", resp)
}

func mockHTTP(log *xlog.Log, addr string) *http.Server {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := rest.MakeRouter(
		rest.Get("/test/getok", mockOKHandler(log)),
		rest.Get("/test/timeout", mockTimeoutHandler(log)),
		rest.Post("/test/ok", mockOKHandler(log)),
		rest.Put("/test/putok", mockOKHandler(log)),
	)
	if err != nil {
		log.Panicf("mock.rest.make.router.error:%+v", err)
	}
	api.SetApp(router)
	handlers := api.MakeHandler()
	h := &http.Server{Addr: addr, Handler: handlers}
	go func() {
		if err := h.ListenAndServe(); err != nil {
			log.Error("mock.rest.error:%+v", err)
			return
		}
	}()
	time.Sleep(time.Millisecond * 100)
	return h
}

func mockOKHandler(log *xlog.Log) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
	}
	return f
}

func mockTimeoutHandler(log *xlog.Log) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		time.Sleep(time.Second * 20)
	}
	return f
}
