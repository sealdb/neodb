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
	"errors"
	"sync"
	"sync/atomic"

	"github.com/sealdb/neodb/config"

	"github.com/sealdb/mysqlstack/xlog"
)

// TxnManager tuple.
type TxnManager struct {
	log        *xlog.Log
	xaCheck    *XaCheck
	txnid      uint64
	txnNums    int64
	commitLock sync.RWMutex
}

// NewTxnManager creates new TxnManager.
func NewTxnManager(log *xlog.Log) *TxnManager {
	return &TxnManager{
		log:   log,
		txnid: 0,
	}
}

// Init is used to init the async worker xaCheck.
func (mgr *TxnManager) Init(scatter *Scatter, ScatterConf *config.ScatterConfig) error {
	xaChecker := NewXaCheck(scatter, ScatterConf)
	if err := xaChecker.Init(); err != nil {
		return err
	}
	mgr.xaCheck = xaChecker
	return nil
}

// Close is used to close the async worker xaCheck.
func (mgr *TxnManager) Close() {
	if mgr.xaCheck != nil {
		mgr.xaCheck.Close()
		mgr.xaCheck = nil
	}
}

// GetID returns a new txnid.
func (mgr *TxnManager) GetID() uint64 {
	return atomic.AddUint64(&mgr.txnid, 1)
}

// Add used to add a txn to mgr.
func (mgr *TxnManager) Add() error {
	atomic.AddInt64(&mgr.txnNums, 1)
	return nil
}

// Remove used to remove a txn from mgr.
func (mgr *TxnManager) Remove() error {
	atomic.AddInt64(&mgr.txnNums, -1)
	return nil
}

// CreateTxn creates new txn.
func (mgr *TxnManager) CreateTxn(backends map[string]*Poolz) (*Txn, error) {
	if len(backends) == 0 {
		return nil, errors.New("backends.is.NULL")
	}

	txid := mgr.GetID()
	txn, err := NewTxn(mgr.log, txid, mgr, backends)
	if err != nil {
		return nil, err
	}
	mgr.Add()
	return txn, nil
}

// CommitLock used to acquire the commit.
func (mgr *TxnManager) CommitLock() {
	mgr.commitLock.Lock()
}

// CommitUnlock used to release the commit.
func (mgr *TxnManager) CommitUnlock() {
	mgr.commitLock.Unlock()
}

// CommitRLock used to acquire the read lock of commit.
func (mgr *TxnManager) CommitRLock() {
	mgr.commitLock.RLock()
}

// CommitRUnlock used to release the read lock of commit.
func (mgr *TxnManager) CommitRUnlock() {
	mgr.commitLock.RUnlock()
}
