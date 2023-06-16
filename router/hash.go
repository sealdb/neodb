/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package router

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sealdb/neodb/config"

	jump "github.com/lithammer/go-jump-consistent-hash"
	"github.com/pkg/errors"
	"github.com/sealdb/mysqlstack/sqlparser"
	"github.com/sealdb/mysqlstack/sqlparser/depends/common"
	"github.com/sealdb/mysqlstack/xlog"
)

// HashRange tuple.
// [Start, End)
type HashRange struct {
	Start int
	End   int
}

// String returns start-end info.
func (r *HashRange) String() string {
	return fmt.Sprintf("[%v-%v)", r.Start, r.End)
}

// Less impl.
func (r *HashRange) Less(b KeyRange) bool {
	v := b.(*HashRange)
	return r.Start < v.Start
}

// Hash tuple.
type Hash struct {
	log *xlog.Log

	// hash slots
	slots int

	// hash method
	typ MethodType

	// table config
	conf *config.TableConfig

	// Partition map
	partitions map[int]Segment
	Segments   []Segment `json:",omitempty"`
}

// NewHash creates new hash.
func NewHash(log *xlog.Log, slots int, conf *config.TableConfig) *Hash {
	return &Hash{
		log:        log,
		conf:       conf,
		slots:      slots,
		typ:        MethodTypeHash,
		partitions: make(map[int]Segment),
		Segments:   make([]Segment, 0, 16),
	}
}

// Build used to build hash bitmap from schema config
func (h *Hash) Build() error {
	var err error
	var start, end int

	for _, part := range h.conf.Partitions {
		segments := strings.Split(part.Segment, "-")
		if len(segments) != 2 {
			return errors.Errorf("hash.partition.segment.malformed[%v]", part.Segment)
		}

		// parse partition spec
		if start, err = strconv.Atoi(segments[0]); err != nil {
			return errors.Errorf("hash.partition.segment.malformed[%v].start.can.not.parser.to.int", part.Segment)
		}
		if end, err = strconv.Atoi(segments[1]); err != nil {
			return errors.Errorf("hash.partition.segment.malformed[%v].end.can.not.parser.to.int", part.Segment)
		}
		if end <= start {
			return errors.Errorf("hash.partition.segment.malformed[%v].start[%v]>=end[%v]", part.Segment, start, end)
		}

		partition := Segment{
			Table:   part.Table,
			Backend: part.Backend,
			Range: &HashRange{
				Start: start,
				End:   end,
			},
		}

		// bitmap
		for i := start; i < end; i++ {
			if _, ok := h.partitions[i]; ok {
				return errors.Errorf("hash.partition.segment[%v].overlapped[%v]", part.Segment, i)
			}
			h.partitions[i] = partition
		}

		// Segments
		h.Segments = append(h.Segments, partition)
	}

	if len(h.partitions) != h.slots {
		return errors.Errorf("hash.partition.last.segment[%v].upper.bound.must.be[%v]", len(h.partitions), h.slots)
	}
	sort.Sort(Segments(h.Segments))
	return nil
}

// Clear used to clean hash partitions
func (h *Hash) Clear() error {
	for k := range h.partitions {
		delete(h.partitions, k)
	}
	return nil
}

// Lookup used to lookup partition(s) through the sharding-key range
// Hash.Lookup only supports the type uint64/string
func (h *Hash) Lookup(start *sqlparser.SQLVal, end *sqlparser.SQLVal) ([]Segment, error) {
	// if open interval we returns all partitions
	if start == nil || end == nil {
		return h.Segments, nil
	}

	// Check item types.
	if start.Type != end.Type {
		return nil, errors.Errorf("hash.lookup.key.type.must.be.same:[%v!=%v]", start.Type, end.Type)
	}

	// Hash just handle the equal
	if bytes.Equal(start.Val, end.Val) {
		idx, err := h.GetIndex(start)
		if err != nil {
			return nil, err
		}
		return []Segment{h.partitions[idx]}, nil
	}
	return h.Segments, nil
}

// Type returns the hash type.
func (h *Hash) Type() MethodType {
	return h.typ
}

// GetIndex returns index based on sqlval.
func (h *Hash) GetIndex(sqlval *sqlparser.SQLVal) (int, error) {
	idx := -1
	valStr := common.BytesToString(sqlval.Val)
	switch sqlval.Type {
	case sqlparser.IntVal:
		unsigned, err := strconv.ParseInt(valStr, 0, 64)
		if err != nil {
			return -1, errors.Errorf("hash.getindex.val.key.parser.uint64.error:[%v]", err)
		}
		idx = int(jump.Hash(uint64(unsigned), int32(h.slots)))
	case sqlparser.FloatVal:
		unsigned, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return -1, errors.Errorf("hash.getindex.val.key.parser.float.error:[%v]", err)
		}
		idx = int(jump.Hash(uint64(unsigned), int32(h.slots)))
	case sqlparser.StrVal:
		idx = int(jump.HashString(valStr, int32(h.slots), jump.NewCRC64()))
	default:
		return -1, errors.Errorf("hash.unsupported.key.type:[%v]", sqlval.Type)
	}
	return idx, nil
}

// GetSegments returns Segments based on index.
func (h *Hash) GetSegments() []Segment {
	return h.Segments
}

func (h *Hash) GetSegment(index int) (Segment, error) {
	if index < 0 || index >= h.slots {
		return Segment{}, errors.Errorf("hash.getsegment.index.[%d].out.of.range", index)
	}
	return h.partitions[index], nil
}
