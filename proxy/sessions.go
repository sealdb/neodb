/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package proxy

import (
	"sort"
	"sync"
	"time"

	"github.com/sealdb/neodb/backend"

	"github.com/sealdb/mysqlstack/driver"
	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/xlog"
)

const (
	sessionStateInTransaction = "In transaction"
)

// Sessions tuple.
type Sessions struct {
	log *xlog.Log
	mu  sync.RWMutex
	// Key is session ID.
	sessions map[uint32]*session
}

// NewSessions creates new session.
func NewSessions(log *xlog.Log) *Sessions {
	return &Sessions{
		log:      log,
		sessions: make(map[uint32]*session),
	}
}

// Add used to add the session to map when session created.
func (ss *Sessions) Add(s *driver.Session) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sessions[s.ID()] = newSession(ss.log, s)
}

// Remove used to remove the session from the map when session exit.
func (ss *Sessions) Remove(s *driver.Session) {
	ss.mu.Lock()
	session, ok := ss.sessions[s.ID()]
	if !ok {
		ss.mu.Unlock()
		return
	}
	delete(ss.sessions, s.ID())
	ss.mu.Unlock()

	session.close()
}

// Kill used to kill a live session.
// 1. remove from sessions list.
// 2. close the session from the server side.
// 3. abort the session's txn.
func (ss *Sessions) Kill(id uint32, reason string) {
	log := ss.log
	ss.mu.Lock()
	session, ok := ss.sessions[id]
	if !ok {
		ss.mu.Unlock()
		return
	}
	delete(ss.sessions, id)
	ss.mu.Unlock()
	log.Warning("session.id[%v].killed.reason:%s", id, reason)

	session.close()
}

// Reaches used to check whether the sessions count reaches(>=) the quota.
func (ss *Sessions) Reaches(quota int) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return (len(ss.sessions) >= quota)
}

// getTxnSession used to get current connection session.
func (ss *Sessions) getTxnSession(session *driver.Session) *session {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.sessions[session.ID()]
}

// getSession used to get current connection session.
func (ss *Sessions) getSession(id uint32) *session {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.sessions[id]
}

// TxnBinding used to bind txn to the session.
func (ss *Sessions) TxnBinding(s *driver.Session, txn backend.Transaction, node sqlparser.Statement, query string) {

	ss.mu.RLock()
	session, ok := ss.sessions[s.ID()]
	if !ok {
		ss.mu.RUnlock()
		return
	}
	ss.mu.RUnlock()

	session.mu.Lock()
	defer session.mu.Unlock()
	q := query
	if len(query) > 128 {
		q = query[:128]
	}
	session.query = q
	session.node = node

	// Bind sid to txn.
	txn.SetSessionID(s.ID())
	session.transaction = txn
	session.timestamp = time.Now().Unix()
}

// TxnUnBinding used to set transaction and node to nil.
func (ss *Sessions) TxnUnBinding(s *driver.Session) {
	ss.mu.RLock()
	session, ok := ss.sessions[s.ID()]
	if !ok {
		ss.mu.RUnlock()
		return
	}
	ss.mu.RUnlock()

	session.mu.Lock()
	defer session.mu.Unlock()
	session.node = nil
	session.query = ""
	session.transaction = nil
	session.timestamp = time.Now().Unix()
}

// MultiStmtTxnBinding used to bind txn, node, query to the session
func (ss *Sessions) MultiStmtTxnBinding(s *driver.Session, txn backend.Transaction, node sqlparser.Statement, query string) {

	ss.mu.RLock()
	session, ok := ss.sessions[s.ID()]
	if !ok {
		ss.mu.RUnlock()
		return
	}
	ss.mu.RUnlock()

	session.mu.Lock()
	defer session.mu.Unlock()
	q := query
	if len(query) > 128 {
		q = query[:128]
	}
	session.query = q
	session.node = node
	// txn should not be nil when "begin" or "start transaction" is executed, to be set just once during the trans.
	if txn != nil {
		// Bind sid to txn.
		txn.SetSessionID(s.ID())
		session.transaction = txn
	}
	session.timestamp = time.Now().Unix()
}

// MultiStmtTxnUnBinding used to set transaction by isEnd
func (ss *Sessions) MultiStmtTxnUnBinding(s *driver.Session, isEnd bool) {
	ss.mu.RLock()
	session, ok := ss.sessions[s.ID()]
	if !ok {
		ss.mu.RUnlock()
		return
	}
	ss.mu.RUnlock()

	session.mu.Lock()
	defer session.mu.Unlock()
	session.node = nil
	session.query = ""
	// If multiple-statement transaction is end or some errors happen, set transaction to be nil
	if isEnd {
		session.transaction = nil
	}
	session.timestamp = time.Now().Unix()
}

// Close used to close all sessions.
func (ss *Sessions) Close() {
	i := 0
	for {
		ss.mu.Lock()
		for _, v := range ss.sessions {
			v.close()
		}
		c := len(ss.sessions)
		ss.mu.Unlock()

		if c > 0 {
			ss.log.Warning("session.wait.for.shutdown.live.txn:[%d].wait.seconds:%d", c, i)
			time.Sleep(time.Second)
			i++
		} else {
			break
		}
	}
}

// SessionInfo tuple.
type SessionInfo struct {
	ID           uint32
	User         string
	Host         string
	DB           string
	Command      string
	Time         uint32
	State        string
	Info         string
	RowsSent     uint64
	RowsExamined uint64
}

// Sort by id.
type sessionInfos []SessionInfo

// Len impl.
func (q sessionInfos) Len() int { return len(q) }

// Swap impl.
func (q sessionInfos) Swap(i, j int) { q[i], q[j] = q[j], q[i] }

// Less impl.
func (q sessionInfos) Less(i, j int) bool { return q[i].ID < q[j].ID }

// Snapshot returns all session info.
func (ss *Sessions) Snapshot() []SessionInfo {
	var infos sessionInfos

	now := time.Now().Unix()
	ss.mu.Lock()
	for _, v := range ss.sessions {
		v.mu.Lock()
		info := SessionInfo{
			ID:      v.session.ID(),
			User:    v.session.User(),
			Host:    v.session.Addr(),
			DB:      v.session.Schema(),
			Command: "Sleep",
			Time:    uint32(now - (int64)(v.session.LastQueryTime().Unix())),
		}

		if v.node != nil {
			info.Command = "Query"
			info.Info = v.query
		}

		if v.transaction != nil {
			// https://dev.mysql.com/doc/refman/5.7/en/general-thread-states.html about state.
			info.State = sessionStateInTransaction
		}

		infos = append(infos, info)
		v.mu.Unlock()
	}
	ss.mu.Unlock()
	sort.Sort(infos)
	return infos
}

// SnapshotTxn returns all sessions info in transaction.
func (ss *Sessions) SnapshotTxn() []SessionInfo {
	var infos sessionInfos

	now := time.Now().Unix()
	ss.mu.Lock()
	for _, v := range ss.sessions {
		v.mu.Lock()
		if v.transaction == nil {
			v.mu.Unlock()
			continue
		}

		info := SessionInfo{
			ID:      v.session.ID(),
			User:    v.session.User(),
			Host:    v.session.Addr(),
			DB:      v.session.Schema(),
			Command: "Sleep",
			Time:    uint32(now - (int64)(v.session.LastQueryTime().Unix())),
		}

		if v.node != nil {
			info.Command = "Query"
			info.Info = v.query
		}
		info.State = sessionStateInTransaction

		infos = append(infos, info)
		v.mu.Unlock()
	}
	ss.mu.Unlock()
	sort.Sort(infos)
	return infos
}

// Snapshot returns all session info about the user.
func (ss *Sessions) SnapshotUser(user string) []SessionInfo {
	var infos sessionInfos

	now := time.Now().Unix()
	ss.mu.Lock()
	for _, v := range ss.sessions {
		if v.session.User() != user {
			continue
		}

		v.mu.Lock()
		info := SessionInfo{
			ID:      v.session.ID(),
			User:    v.session.User(),
			Host:    v.session.Addr(),
			DB:      v.session.Schema(),
			Command: "Sleep",
			Time:    uint32(now - (int64)(v.session.LastQueryTime().Unix())),
		}

		if v.node != nil {
			info.Command = "Query"
			info.Info = v.query
		}

		if v.transaction != nil {
			// https://dev.mysql.com/doc/refman/5.7/en/general-thread-states.html about state.
			info.State = sessionStateInTransaction
		}

		infos = append(infos, info)
		v.mu.Unlock()
	}
	ss.mu.Unlock()
	sort.Sort(infos)
	return infos
}
