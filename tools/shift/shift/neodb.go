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
	"net/http"
	"strings"
	"time"

	"github.com/sealdb/neodb/tools/shift/xbase"

	"github.com/juju/errors"
)

func (shift *Shift) setNeoDBReadOnly(v bool) error {
	log := shift.log
	cfg := shift.cfg
	path := cfg.NeoDBURL + "/v1/neodb/readonly"

	type request struct {
		Readonly bool `json:"readonly"`
	}
	req := &request{
		Readonly: v,
	}
	log.Info("shift.set.neodb[%s].readonly.req[%+v]", path, req)

	resp, cleanup, err := xbase.HTTPPut(path, req)
	defer cleanup()
	if err != nil {
		return errors.Trace(err)
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		return errors.Trace(errors.Errorf("shift.set.neodb.readonly[%s].response.error:%+s", path, xbase.HTTPReadBody(resp)))
	}
	return nil
}

func (shift *Shift) setNeoDBRule() error {
	log := shift.log
	cfg := shift.cfg
	path := cfg.NeoDBURL + "/v1/shard/shift"

	if _, isSystem := sysDatabases[strings.ToLower(shift.cfg.FromDatabase)]; isSystem {
		log.Info("shift.set.neodb.rune.skip.system.table:[%s.%s]", shift.cfg.FromDatabase, shift.cfg.FromTable)
		return nil
	}

	type request struct {
		Database    string `json:"database"`
		Table       string `json:"table"`
		FromAddress string `json:"from-address"`
		ToAddress   string `json:"to-address"`
	}
	req := &request{
		Database:    cfg.FromDatabase,
		Table:       cfg.FromTable,
		FromAddress: cfg.From,
		ToAddress:   cfg.To,
	}
	log.Info("shift.set.neodb[%s].rule.req[%+v]", path, req)

	resp, cleanup, err := xbase.HTTPPost(path, req)
	defer cleanup()
	if err != nil {
		return errors.Trace(err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		return errors.Trace(errors.Errorf("shift.set.neodb.shard.rule[%s].response.error:%+s", path, xbase.HTTPReadBody(resp)))
	}
	return nil
}

var (
	neodb_limits_min = 500
	neodb_limits_max = 10000
)

func (shift *Shift) setNeoDBThrottle(factor float32) error {
	log := shift.log
	cfg := shift.cfg
	path := cfg.NeoDBURL + "/v1/neodb/throttle"

	type request struct {
		Limits int `json:"limits"`
	}

	// limits =0 means unlimits.
	limits := int(float32(neodb_limits_max) * factor)
	if limits != 0 && limits < neodb_limits_min {
		limits = neodb_limits_min
	}
	req := &request{
		Limits: limits,
	}
	log.Info("shift.set.neodb[%s].throttle.to.req[%+v].by.factor[%v].limits[%v]", path, req, factor, limits)

	resp, cleanup, err := xbase.HTTPPut(path, req)
	defer cleanup()
	if err != nil {
		return errors.Trace(err)
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		return errors.Trace(errors.Errorf("shift.set.neodb.throttle[%s].response.error:%+s", path, xbase.HTTPReadBody(resp)))
	}
	return nil
}

func (shift *Shift) setNeoDB() error {
	log := shift.log

	// 1. WaitUntilPos
	{
		masterPos, err := shift.masterPosition()
		if err != nil {
			return errors.Trace(err)
		}
		log.Info("shift.wait.until.pos[%#v]...", masterPos)
		if err := shift.canal.WaitUntilPos(*masterPos, time.Hour*12); err != nil {
			log.Error("shift.set.neodb.wait.until.pos[%#v].error", masterPos)
			return errors.Trace(err)
		}
		log.Info("shift.wait.until.pos.done...")
	}

	// 2. Set neodb to readonly.
	{
		log.Info("shift.set.neodb.readonly...")
		if err := shift.setNeoDBReadOnly(true); err != nil {
			log.Error("shift.set.neodb.readonly.error")
			return errors.Trace(err)
		}
		log.Info("shift.set.neodb.readonly.done...")
	}

	// 3. Wait again.
	{
		masterPos, err := shift.masterPosition()
		if err != nil {
			return errors.Trace(err)
		}
		log.Info("shift.wait.until.pos.again[%#v]...", masterPos)
		if err := shift.canal.WaitUntilPos(*masterPos, time.Second*300); err != nil {
			log.Error("shift.wait.until.pos.again[%#v].error", masterPos)
			return errors.Trace(err)
		}
		log.Info("shift.wait.until.pos.again.done...")
	}

	// 4. Checksum table.
	if shift.cfg.Checksum {
		log.Info("shift.checksum.table...")
		if err := shift.ChecksumTable(); err != nil {
			log.Error("shift.checksum.table.error")
			return errors.Trace(err)
		}
		log.Info("shift.checksum.table.done...")
	}

	// 5. Rename ToTable.
	{
		log.Info("shift.rename.totable...")
		if err := shift.renameToTable(); err != nil {
			log.Error("shift.rename.totable.error")
			return errors.Trace(err)
		}
		log.Info("shift.rename.totable.done...")
	}

	// 6. Set neodb rule.
	{
		if shift.cfg.ToFlavor == ToMySQLFlavor || shift.cfg.ToFlavor == ToMariaDBFlavor {
			log.Info("shift.set.neodb.rule...")
			if err := shift.setNeoDBRule(); err != nil {
				log.Error("shift.set.neodb.rule.error")
				return errors.Trace(err)
			}
			log.Info("shift.set.neodb.rule.done...")
		}
	}

	// 7. Set neodb to read/write.
	{
		log.Info("shift.set.neodb.to.write...")
		if err := shift.setNeoDBReadOnly(false); err != nil {
			log.Error("shift.set.neodb.to.write.error")
			return errors.Trace(err)
		}
		log.Info("shift.set.neodb.to.write.done...")
	}

	// 8. Set neodb throttle to unlimits.
	{
		log.Info("shift.set.neodb.throttle.to.unlimits...")
		if err := shift.setNeoDBThrottle(0); err != nil {
			log.Error("shift.set.neodb.throttle.to.unlimits.error")
			return errors.Trace(err)
		}
		log.Info("shift.set.neodb.throttle.to.unlimits.done...")
	}

	// 9. Good, we have all done.
	{
		shift.done <- true
		shift.allDone.Set(true)
		log.Info("shift.all.done...")
	}
	return nil
}
