/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/sealdb/neodb/build"
	"github.com/sealdb/neodb/config"
	"github.com/sealdb/neodb/ctl"
	"github.com/sealdb/neodb/monitor"
	"github.com/sealdb/neodb/proxy"
	"github.com/sealdb/neodb/version"

	"github.com/sealdb/mysqlstack/xlog"
)

var (
	flagConf string
)

func init() {
	flag.StringVar(&flagConf, "c", "", "neodb config file")
	flag.StringVar(&flagConf, "config", "", "neodb config file")
}

func usage() {
	fmt.Println("Usage: " + os.Args[0] + " [-c|--config] <neodb-config-file>")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))

	fmt.Println(*version.GetBanner())
	fmt.Printf("version: [%+v]\n", version.GetVersion())
	build := build.GetInfo() // TODO: choose version or build info
	fmt.Printf("neodb:[%+v]\n", build)

	// config
	flag.Usage = func() { usage() }
	flag.Parse()
	if flagConf == "" {
		usage()
		os.Exit(0)
	}

	conf, err := config.LoadConfig(flagConf)
	if err != nil {
		log.Panic("neodb.load.config.error[%v]", err)
	}
	log.SetLevel(conf.Log.Level)

	// Monitor
	monitor.Start(log, conf)

	// Proxy.
	proxy := proxy.NewProxy(log, flagConf, build.Tag, conf)
	proxy.Start()

	// Admin portal.
	admin := ctl.NewAdmin(log, proxy)
	admin.Start()

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Info("neodb.signal:%+v", <-ch)

	// Stop the proxy and httpserver.
	proxy.Stop()
	admin.Stop()
}
