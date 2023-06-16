/*
 * NeoDB
 *
 * Copyright 2019 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package shift

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sealdb/neodb/tools/shift/xlog"

	"github.com/ant0ine/go-json-rest/rest"
)

var (
	restfulPort = 8181

	// Config for normal shift.
	mockCfg = &Config{
		ToFlavor:     "mysql",
		From:         "127.0.0.1:3306",
		FromUser:     "root",
		FromDatabase: "shift_test_from",
		FromTable:    "t1",

		To:         "127.0.0.1:3307",
		ToUser:     "root",
		ToDatabase: "shift_test_to",
		ToTable:    "t1",

		Rebalance: false,
		Cleanup:   true,
		Threads:   16,
		Behinds:   256,
		MySQLDump: "mysqldump",
		NeoDBURL:  fmt.Sprintf("http://127.0.0.1:%d", restfulPort),
		Checksum:  true,
	}

	// Config for system (mysql) shift.
	mockCfgMysql = &Config{
		ToFlavor:     "mysql",
		From:         "127.0.0.1:3306",
		FromUser:     "root",
		FromDatabase: "mysql",
		FromTable:    "user",

		To:         "127.0.0.1:3307",
		ToUser:     "root",
		ToDatabase: "mysql",
		ToTable:    "userx",

		Rebalance: false,
		Cleanup:   true,
		Threads:   16,
		Behinds:   256,
		MySQLDump: "mysqldump",
		NeoDBURL:  fmt.Sprintf("http://127.0.0.1:%d", restfulPort),
		Checksum:  false,
	}

	// Config for xa shift.
	mockCfgXa = &Config{
		ToFlavor:     "mysql",
		From:         "127.0.0.1:3306",
		FromUser:     "root",
		FromDatabase: "shift_test_from",
		FromTable:    "t1",

		To:         "127.0.0.1:3307",
		ToUser:     "root",
		ToDatabase: "shift_test_to",
		ToTable:    "t1",

		Rebalance: false,
		Cleanup:   true,
		Threads:   16,
		Behinds:   256,
		MySQLDump: "mysqldump",
		NeoDBURL:  fmt.Sprintf("http://127.0.0.1:%d", restfulPort),
		Checksum:  true,
	}

	// Config for ddl shift.
	mockCfgDDL = &Config{
		ToFlavor:     "mysql",
		From:         "127.0.0.1:3306",
		FromUser:     "root",
		FromDatabase: "shift_test_from",
		FromTable:    "t1",

		To:         "127.0.0.1:3306",
		ToUser:     "root",
		ToDatabase: "shift_test_to",
		ToTable:    "t1",

		Rebalance: false,
		Cleanup:   true,
		Threads:   16,
		Behinds:   256,
		MySQLDump: "mysqldump",
		NeoDBURL:  fmt.Sprintf("http://127.0.0.1:%d", restfulPort),
		Checksum:  false,
	}

	// Config for neodb shift.
	mockNeoDBDBCfg = &Config{
		ToFlavor:     "neodb",
		From:         "127.0.0.1:3306",
		FromUser:     "root",
		FromDatabase: "shift_test_from",
		FromTable:    "t1",

		To:         "127.0.0.1:3308",
		ToUser:     "root",
		ToDatabase: "shift_test_to",
		ToTable:    "t1",

		Rebalance: false,
		Cleanup:   true,
		Threads:   16,
		Behinds:   256,
		MySQLDump: "mysqldump",
		NeoDBURL:  fmt.Sprintf("http://127.0.0.1:%d", 8080),
		Checksum:  true,
	}

	// Config for rebalance shift.
	mockCfgRebalance = &Config{
		ToFlavor:     "mysql",
		From:         "127.0.0.1:3306",
		FromUser:     "root",
		FromDatabase: "shift_test_from",
		FromTable:    "t1",

		To:         "127.0.0.1:3307",
		ToUser:     "root",
		ToDatabase: "shift_test_to",
		ToTable:    "t1",

		Rebalance: true,
		Cleanup:   false,
		Threads:   16,
		Behinds:   256,
		MySQLDump: "mysqldump",
		NeoDBURL:  fmt.Sprintf("http://127.0.0.1:%d", restfulPort),
		Checksum:  true,
	}
)

func mockShift(log *xlog.Log, cfg *Config, hasPK bool, initData bool, readonlyHanler mockHandler, shardshiftHandler mockHandler, throttleHandler mockHandler) (*Shift, func()) {
	h := mockHttp(log, restfulPort, readonlyHanler, shardshiftHandler, throttleHandler)
	shift, _ := NewShift(log, cfg).(*Shift)

	// Prepare connections.
	{
		if err := shift.prepareConnection(); err != nil {
			log.Panicf("mock.shift.prepare.connection.error:%+v", err)
		}
	}

	// Prepare the from database and table.
	{
		fromConn := shift.fromPool.Get()
		if fromConn == nil {
			panic("shift.mock.get.from.conn.nil.error")
		}
		defer shift.fromPool.Put(fromConn)
		toConn := shift.toPool.Get()
		if toConn == nil {
			panic("shift.mock.get.to.conn.nil.error")
		}
		defer shift.toPool.Put(toConn)

		// Cleanup To table first.
		{
			sql := fmt.Sprintf("drop table if exists `%s`.`%s`", cfg.ToDatabase, cfg.ToTable)
			if _, err := toConn.Execute(sql); err != nil {
				log.Panicf("mock.drop.to.table.error:%+v", err)
			}
		}

		if _, isSystem := sysDatabases[strings.ToLower(cfg.FromDatabase)]; !isSystem {
			// Cleanup From database first.
			{
				sql := fmt.Sprintf("drop database if exists `%s`", cfg.FromDatabase)
				if _, err := fromConn.Execute(sql); err != nil {
					log.Panicf("mock.shift.drop.from.table.error:%+v", err)
				}
			}

			// Create database on from.
			sql := fmt.Sprintf("create database if not exists `%s`", cfg.FromDatabase)
			if _, err := fromConn.Execute(sql); err != nil {
				log.Panicf("mock.shift.prepare.database.error:%+v", err)
			}

			// Create table on from.
			if hasPK {
				sql = fmt.Sprintf("create table `%s`.`%s`(a int primary key, b int, c varchar(200), d DOUBLE NULL DEFAULT NULL, e json DEFAULT NULL, f INT UNSIGNED DEFAULT NULL, g BIGINT DEFAULT NULL, h BIGINT UNSIGNED DEFAULT NULL, i TINYINT NULL, j TINYINT UNSIGNED DEFAULT NULL, k SMALLINT DEFAULT NULL, l SMALLINT UNSIGNED DEFAULT NULL, m MEDIUMINT DEFAULT NULL, n INT UNSIGNED DEFAULT NULL, o bit(1) default NULL, p text COLLATE utf8_bin, q longblob, r datetime DEFAULT NULL)", cfg.FromDatabase, cfg.FromTable)
			} else {
				sql = fmt.Sprintf("create table `%s`.`%s`(a int, b int, c varchar(200),  d DOUBLE NULL DEFAULT NULL, e json DEFAULT NULL, f INT UNSIGNED DEFAULT NULL, g BIGINT DEFAULT NULL, h BIGINT UNSIGNED DEFAULT NULL, i TINYINT NULL, j TINYINT UNSIGNED DEFAULT NULL, k SMALLINT DEFAULT NULL, l SMALLINT UNSIGNED DEFAULT NULL, m MEDIUMINT DEFAULT NULL, n INT UNSIGNED DEFAULT NULL, o bit(1) default NULL, p text COLLATE utf8_bin, q longblob, r datetime DEFAULT NULL)", cfg.FromDatabase, cfg.FromTable)
			}
			if _, err := fromConn.Execute(sql); err != nil {
				log.Panicf("mock.shift.prepare.database.error:%+v", err)
			}

			if initData {
				for i := 100; i < 108; i++ {
					sql := fmt.Sprintf("insert into `%s`.`%s`(a,b,c,o,p,q,r) values(%d,%d,'%d', B'1', 0x6B313134363020666638303831383135646534373733633031356465343762353138653030303020E799BDE4BAAC2031302E3131362E32352E3137322C31312E312E31302E313420737061636520636F6E66696775726174696F6E207570646174656420737061636573207479706520676C6F62616C207374617475732063757272656E74206E616D65206B65792073706320686F6D65207061676520706167653A20762E31202833323831383229, 0x6B313134363020666638303831383135646534373733633031356465343762353138653030303020E799BDE4BAAC2031302E3131362E32352E3137322C31312E312E31302E313420737061636520636F6E66696775726174696F6E207570646174656420737061636573207479706520676C6F62616C207374617475732063757272656E74206E616D65206B65792073706320686F6D65207061676520706167653A20762E31202833323831383229, '2019-4-19 18:03:43')", shift.cfg.FromDatabase, shift.cfg.FromTable, i, i, i)
					if _, err := fromConn.Execute(sql); err != nil {
						log.Panicf("mock.shift.prepare.datas.error:%+v", err)
					}
				}
			}
		} else {
			// Prepare mysql.userx(fakes for mysql.user) table on TO.
			sql := fmt.Sprintf("show create table `%s`.`%s`", cfg.FromDatabase, cfg.FromTable)
			r, err := fromConn.Execute(sql)
			if err != nil {
				log.Panicf("mock.prepare.mysql.userx.error:%+v", err)
			}
			sql, _ = r.GetString(0, 1)
			sql = strings.Replace(sql, fmt.Sprintf("CREATE TABLE `%s`", cfg.FromTable), fmt.Sprintf("CREATE TABLE `%s`.`%s`", cfg.ToDatabase, cfg.ToTable), 1)
			if _, err = toConn.Execute(sql); err != nil {
				log.Panicf("mock.prepare.mysql.userx.error:%+v", err)
			}

			if initData {
				for i := 100; i < 108; i++ {
					sql := fmt.Sprintf(`insert into %s.%s values("%d", "%d","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","N","","","","",0,0,0,0,"mysql_native_password","*THISISNOTAVALIDPASSWORDTHATCANBEUSEDHERE","N","2017-06-22 17:37:18",NULL,"Y")`, shift.cfg.ToDatabase, shift.cfg.ToTable, i, i)
					if _, err := toConn.Execute(sql); err != nil {
						log.Panicf("mock.shift.prepare.datas.error:%+v", err)
					}
				}
			}
		}
	}

	// Prepare tables.
	{
		if err := shift.prepareTable(); err != nil {
			log.Panicf("mock.shift.prepare.table.error:%+v", err)
		}
	}

	// Prepare canal.
	{
		if err := shift.prepareCanal(); err != nil {
			log.Panicf("mock.shift.prepare.canal.error:%+v", err)
		}
		time.Sleep(time.Millisecond * 100)
	}

	// Prepare nearcheck.
	{
		if err := shift.behindsCheckStart(); err != nil {
			log.Panicf("mock.shift.behinds.check.error:%+v", err)
		}
	}
	return shift, func() {
		shift.close()
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		h.Shutdown(ctx)
		time.Sleep(time.Millisecond * 100)
	}
}

func MockShift(log *xlog.Log, hasPK bool) (*Shift, func()) {
	cfg := mockCfg
	return mockShift(log, cfg, hasPK, false, mockNeoDBReadonly, mockNeoDBShift, mockNeoDBThrottle)
}

func MockShiftWithCleanup(log *xlog.Log, hasPK bool) (*Shift, func()) {
	mockCfg.Cleanup = true
	cfg := mockCfg
	return mockShift(log, cfg, hasPK, false, mockNeoDBReadonly, mockNeoDBShift, mockNeoDBThrottle)
}

func MockShiftWithRebalance(log *xlog.Log, hasPK bool) (*Shift, func()) {
	cfg := mockCfgRebalance
	return mockShift(log, cfg, hasPK, false, mockNeoDBReadonly, mockNeoDBShift, mockNeoDBThrottle)
}

func MockShiftWithData(log *xlog.Log, hasPK bool) (*Shift, func()) {
	cfg := mockCfg
	return mockShift(log, cfg, hasPK, true, mockNeoDBReadonly, mockNeoDBShift, mockNeoDBThrottle)
}

func MockShiftXa(log *xlog.Log, hasPK bool) (*Shift, func()) {
	cfg := mockCfgXa
	return mockShift(log, cfg, hasPK, false, mockNeoDBReadonly, mockNeoDBShift, mockNeoDBThrottle)
}

func MockShiftDDL(log *xlog.Log, hasPK bool) (*Shift, func()) {
	cfg := mockCfgDDL
	return mockShift(log, cfg, hasPK, false, mockNeoDBReadonly, mockNeoDBShift, mockNeoDBThrottle)
}

func MockShiftMysqlTable(log *xlog.Log, hasPK bool) (*Shift, func()) {
	cfg := mockCfgMysql
	return mockShift(log, cfg, hasPK, false, mockNeoDBReadonly, mockNeoDBShift, mockNeoDBThrottle)
}

func MockShiftMysqlTableWithData(log *xlog.Log, hasPK bool) (*Shift, func()) {
	cfg := mockCfgMysql
	return mockShift(log, cfg, hasPK, true, mockNeoDBReadonly, mockNeoDBShift, mockNeoDBThrottle)
}

func MockShiftWithNeoDBReadonlyError(log *xlog.Log, hasPK bool) (*Shift, func()) {
	cfg := mockCfg
	return mockShift(log, cfg, false, false, mockNeoDBReadonlyError, mockNeoDBShift, mockNeoDBThrottle)
}

func MockShiftWithNeoDBShardRuleError(log *xlog.Log, hasPK bool) (*Shift, func()) {
	cfg := mockCfg
	return mockShift(log, cfg, false, false, mockNeoDBReadonly, mockNeoDBShiftError, mockNeoDBThrottle)
}

// RESTful api.
type mockHandler func(log *xlog.Log) rest.HandlerFunc

func mockHttp(log *xlog.Log, port int, readonly mockHandler, shardshift mockHandler, throttle mockHandler) *http.Server {
	httpAddr := fmt.Sprintf(":%d", port)
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := rest.MakeRouter(
		rest.Put("/v1/neodb/readonly", readonly(log)),
		rest.Put("/v1/neodb/throttle", throttle(log)),
		rest.Post("/v1/shard/shift", shardshift(log)),
	)
	if err != nil {
		log.Panicf("mock.shift.rest.make.router.error:%+v", err)
	}
	api.SetApp(router)
	handlers := api.MakeHandler()
	h := &http.Server{Addr: httpAddr, Handler: handlers}
	go func() {
		if err := h.ListenAndServe(); err != nil {
			log.Error("mock.shift.rest.error:%+v", err)
			return
		}
	}()
	time.Sleep(time.Millisecond * 100)
	return h
}

var readonlyLast bool

type readonlyParams struct {
	ReadOnly bool `json:"readonly"`
}

func mockNeoDBReadonly(log *xlog.Log) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		p := readonlyParams{}
		r.DecodeJsonPayload(&p)
		readonlyLast = p.ReadOnly
		log.Info("mock.api.neodb.readonly.call.req:%+v", p)
	}
	return f
}

var throttleLast int

type throttleParams struct {
	Limits int `json:"limits"`
}

func mockNeoDBThrottle(log *xlog.Log) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		p := throttleParams{}
		r.DecodeJsonPayload(&p)
		throttleLast = p.Limits
		log.Info("mock.api.neodb.throttle.call.req:%+v", p)
	}
	return f
}

func mockNeoDBShift(log *xlog.Log) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		log.Info("mock.api.neodb.rule.call")
	}
	return f
}

func mockNeoDBReadonlyError(log *xlog.Log) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		log.Info("mock.api.readonly.error.call")
		readonlyLast = false
		rest.Error(w, "mock.api.readonly.error", http.StatusInternalServerError)
	}
	return f
}

func mockNeoDBShiftError(log *xlog.Log) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		log.Info("mock.api.shift.error.call")
		rest.Error(w, "mock.api.shift.error", http.StatusInternalServerError)
	}
	return f
}

func mockPanicMe(log *xlog.Log, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Info("mock.panicme.fired, msg:%s", msg)
	panic(1)
}

func mockRecoverPanicMe(log *xlog.Log, format string, v ...interface{}) {
	defer func() {
		if x := recover(); x != nil {
			msg := fmt.Sprintf(format, v...)
			log.Info("mock.panicme.fired, msg:%s", msg)
		}
	}()
	panic(1)
}
