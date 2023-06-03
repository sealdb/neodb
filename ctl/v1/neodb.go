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

type neoDBParams struct {
	MaxConnections      *int     `json:"max-connections"`
	MaxResultSize       *int     `json:"max-result-size"`
	MaxJoinRows         *int     `json:"max-join-rows"`
	DDLTimeout          *int     `json:"ddl-timeout"`
	QueryTimeout        *int     `json:"query-timeout"`
	TwoPCEnable         *bool    `json:"twopc-enable"`
	LoadBalance         *int     `json:"load-balance"`
	AllowIP             []string `json:"allowip,omitempty"`
	AuditMode           *string  `json:"audit-mode"`
	StreamBufferSize    *int     `json:"stream-buffer-size"`
	Blocks              *int     `json:"blocks-readonly"`
	LowerCaseTableNames *int     `json:"lower-case-table-names"`
}

// NeoDBConfigHandler impl.
func NeoDBConfigHandler(log *xlog.Log, proxy *proxy.Proxy) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		neoDBConfigHandler(log, proxy, w, r)
	}
	return f
}

func neoDBConfigHandler(log *xlog.Log, proxy *proxy.Proxy, w rest.ResponseWriter, r *rest.Request) {
	p := neoDBParams{}
	err := r.DecodeJsonPayload(&p)
	if err != nil {
		log.Error("api.v1.neodb.config.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Warning("api.v1.neodb[from:%v].body:%+v", r.RemoteAddr, p)
	if p.MaxConnections != nil {
		proxy.SetMaxConnections(*p.MaxConnections)
	}
	if p.MaxResultSize != nil {
		proxy.SetMaxResultSize(*p.MaxResultSize)
	}
	if p.MaxJoinRows != nil {
		proxy.SetMaxJoinRows(*p.MaxJoinRows)
	}
	if p.DDLTimeout != nil {
		proxy.SetDDLTimeout(*p.DDLTimeout)
	}
	if p.QueryTimeout != nil {
		proxy.SetQueryTimeout(*p.QueryTimeout)
	}
	if p.TwoPCEnable != nil {
		proxy.SetTwoPC(*p.TwoPCEnable)
	}
	if p.LoadBalance != nil {
		proxy.SetLoadBalance(*p.LoadBalance)
	}
	proxy.SetAllowIP(p.AllowIP)
	if p.AuditMode != nil {
		proxy.SetAuditMode(*p.AuditMode)
	}
	if p.StreamBufferSize != nil {
		proxy.SetStreamBufferSize(*p.StreamBufferSize)
	}
	if p.Blocks != nil {
		proxy.SetBlocks(*p.Blocks)
	}
	if p.LowerCaseTableNames != nil {
		proxy.SetLowerCaseTableNames(*p.LowerCaseTableNames)
	}

	// reset the allow ip table list.
	proxy.IPTable().Refresh()

	// write to file.
	if err := proxy.FlushConfig(); err != nil {
		log.Error("api.v1.neodb.flush.config.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type readonlyParams struct {
	ReadOnly bool `json:"readonly"`
}

// ReadonlyHandler impl.
func ReadonlyHandler(log *xlog.Log, proxy *proxy.Proxy) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		readonlyHandler(log, proxy, w, r)
	}
	return f
}

func readonlyHandler(log *xlog.Log, proxy *proxy.Proxy, w rest.ResponseWriter, r *rest.Request) {
	p := readonlyParams{}
	err := r.DecodeJsonPayload(&p)
	if err != nil {
		log.Error("api.v1.readonly.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Warning("api.v1.readonly[from:%v].body:%+v", r.RemoteAddr, p)
	proxy.SetReadOnly(p.ReadOnly)
}

type twopcParams struct {
	Twopc bool `json:"twopc"`
}

// TwopcHandler impl.
func TwopcHandler(log *xlog.Log, proxy *proxy.Proxy) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		twopcHandler(log, proxy, w, r)
	}
	return f
}

func twopcHandler(log *xlog.Log, proxy *proxy.Proxy, w rest.ResponseWriter, r *rest.Request) {
	p := twopcParams{}
	err := r.DecodeJsonPayload(&p)
	if err != nil {
		log.Error("api.v1.twopc.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Warning("api.v1.twopc[from:%v].body:%+v", r.RemoteAddr, p)
	proxy.SetTwoPC(p.Twopc)
}

type throttleParams struct {
	Limits int `json:"limits"`
}

// ThrottleHandler impl.
func ThrottleHandler(log *xlog.Log, proxy *proxy.Proxy) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		throttleHandler(log, proxy, w, r)
	}
	return f
}

func throttleHandler(log *xlog.Log, proxy *proxy.Proxy, w rest.ResponseWriter, r *rest.Request) {
	p := throttleParams{}
	err := r.DecodeJsonPayload(&p)
	if err != nil {
		log.Error("api.v1.neodb.throttle.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Warning("api.v1.neodb.throttle[from:%v].body:%+v", r.RemoteAddr, p)
	proxy.SetThrottle(p.Limits)
}

// StatusHandler impl.
func StatusHandler(log *xlog.Log, proxy *proxy.Proxy) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		statusHandler(log, proxy, w, r)
	}
	return f
}

func statusHandler(log *xlog.Log, proxy *proxy.Proxy, w rest.ResponseWriter, r *rest.Request) {
	spanner := proxy.Spanner()
	type status struct {
		ReadOnly bool `json:"readonly"`
	}
	statuz := &status{
		ReadOnly: spanner.ReadOnly(),
	}
	w.WriteJson(statuz)
}

// RestAPIAddressHandler impl.
func RestAPIAddressHandler(log *xlog.Log, proxy *proxy.Proxy) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		type resp struct {
			Addr string `json:"address"`
		}
		rsp := &resp{Addr: proxy.PeerAddress()}
		w.WriteJson(rsp)
	}
	return f
}
