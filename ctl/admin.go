/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package ctl

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/sealdb/neodb/proxy"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/sealdb/mysqlstack/xlog"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
}

// Admin tuple.
type Admin struct {
	log    *xlog.Log
	proxy  *proxy.Proxy
	server *http.Server
}

// NewAdmin creates the new admin.
func NewAdmin(log *xlog.Log, proxy *proxy.Proxy) *Admin {
	return &Admin{
		log:   log,
		proxy: proxy,
	}
}

// Start starts http server.
func (admin *Admin) Start() {
	api := rest.NewApi()
	router, err := admin.NewRouter()
	if err != nil {
		panic(err)
	}

	api.SetApp(router)
	handlers := api.MakeHandler()
	admin.server = &http.Server{Addr: admin.proxy.PeerAddress(), Handler: handlers}

	go func() {
		log := admin.log
		log.Info("http.server.start[%v]...", admin.proxy.PeerAddress())
		if err := admin.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Panic("%v", err)
		}
	}()
}

// Stop stops http server.
func (admin *Admin) Stop() {
	log := admin.log
	admin.server.Shutdown(context.Background())
	log.Info("http.server.gracefully.stop")
}
