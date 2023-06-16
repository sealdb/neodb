/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package backend

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sealdb/neodb/config"
	"github.com/sealdb/neodb/xbase/stats"

	"github.com/sealdb/mysqlstack/xlog"
)

var (
	poolCounterPing       = "#pool.ping"
	poolCounterPingBroken = "#pool.ping.broken"
	poolCounterHit        = "#pool.hit"
	poolCounterMiss       = "#pool.miss"
	poolCounterGet        = "#pool.get"
	poolCounterPut        = "#pool.put"
	poolCounterClose      = "#pool.close"

	poolCounterBackendDialError        = "#backend.dial.error"
	poolCounterBackendExecuteTimeout   = "#backend.execute.timeout"
	poolCounterBackendExecuteMaxresult = "#backend.execute.maxresult"
	poolCounterBackendExecuteAllError  = "#backend.execute.all.error"
	poolCounterBackendKilled           = "#backend.killed"
)

var (
	maxIdleTime = 20 // 20s
	errClosed   = errors.New("can't get connection from the closed DB")
)

// Poolz ...
// Add replica and normal pool to distribute SQL between
// read and write in some cases for load-balance.
type Poolz struct {
	log     *xlog.Log
	conf    *config.BackendConfig
	normal  *Pool
	replica *Pool
}

// NewPoolz create the new Poolz.
func NewPoolz(log *xlog.Log, conf *config.BackendConfig) *Poolz {
	return &Poolz{
		log:     log,
		conf:    conf,
		normal:  NewPool(log, conf, conf.Address),
		replica: NewPool(log, conf, conf.Replica),
	}
}

// Close used to close the poolz.
func (p *Poolz) Close() {
	if p.normal != nil {
		p.normal.Close()
	}
	if p.replica != nil {
		p.replica.Close()
	}
}

// JSON returns the available string.
func (p *Poolz) JSON() string {
	str := p.normal.JSON()
	if p.replica != nil {
		str += ", " + p.replica.JSON()
	}
	return str
}

// Pool tuple.
type Pool struct {
	mu          sync.RWMutex
	log         *xlog.Log
	address     string
	conf        *config.BackendConfig
	counters    *stats.Counters
	connections chan Connection

	// If maxIdleTime reached, the connection will be closed by get.
	maxIdleTime int64
}

// NewPool creates the new Pool.
func NewPool(log *xlog.Log, conf *config.BackendConfig, address string) *Pool {
	if address == "" {
		return nil
	}
	return &Pool{
		log:         log,
		address:     address,
		conf:        conf,
		connections: make(chan Connection, conf.MaxConnections),
		counters:    stats.NewCounters(conf.Name + "@" + address),
		maxIdleTime: int64(maxIdleTime),
	}
}

func (p *Pool) reconnect() (Connection, error) {
	log := p.log
	c := NewConnection(log, p)
	if err := c.Dial(); err != nil {
		log.Error("pool.reconnect.dial.error:%+v", err)
		return nil, err
	}
	c.SetTimestamp(time.Now().Unix())
	return c, nil
}

// Get used to get a connection from the pool.
func (p *Pool) Get() (Connection, error) {
	counters := p.counters
	counters.Add(poolCounterGet, 1)

	conns := p.getConns()
	if conns == nil {
		return nil, errClosed
	}

	select {
	case conn, more := <-conns:
		if !more {
			return nil, errClosed
		}
		// If the idle time more than 1s,
		// we will do a ping to check the connection is OK or NOT.
		now := time.Now().Unix()
		elapsed := (now - conn.Timestamp())
		if elapsed > 1 {
			// If elapsed time more than 20s, we create new one.
			if elapsed > atomic.LoadInt64(&p.maxIdleTime) {
				conn.Close()
				return p.reconnect()
			}

			if err := conn.Ping(); err != nil {
				counters.Add(poolCounterPingBroken, 1)
				return p.reconnect()
			}
			counters.Add(poolCounterPing, 1)
		}
		counters.Add(poolCounterHit, 1)
		return conn, nil
	default:
		counters.Add(poolCounterMiss, 1)
		return p.reconnect()
	}
}

// Put used to put a connection to pool.
func (p *Pool) Put(conn Connection) {
	p.put(conn, true)
}

func (p *Pool) put(conn Connection, updateTs bool) {
	p.counters.Add(poolCounterPut, 1)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.connections == nil {
		return
	}

	if updateTs {
		conn.SetTimestamp(time.Now().Unix())
	}
	select {
	case p.connections <- conn:
	default:
		conn.Close()
	}
}

// Close used to close the pool.
func (p *Pool) Close() {
	p.counters.Add(poolCounterClose, 1)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.connections == nil {
		return
	}
	close(p.connections)
	for conn := range p.connections {
		conn.Close()
	}
	p.connections = nil
}

func (p *Pool) getConns() chan Connection {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connections
}

// JSON returns the available string.
// available is the number of currently unused connections.
func (p *Pool) JSON() string {
	b := bytes.NewBuffer(make([]byte, 0, 256))
	fmt.Fprintf(b, "{'name': '%s@%s', 'capacity': %d, 'counters': %s}", p.conf.Name, p.address, p.conf.MaxConnections, p.counters.String())
	return b.String()
}
