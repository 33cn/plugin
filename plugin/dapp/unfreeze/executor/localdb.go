// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

/*
 1. 可以用创建者和收益者进行列表
 1. 按 txIndex 排序
*/

var opt_addr_table = &table.Option{
	Prefix:  "LODB-unfreeze",
	Name:    "addr",
	Primary: "txIndex",
	Index: []string{
		"beneficiary",
		"init",
		"id",
	},
}

// AddrRow order row
type AddrRow struct {
	*pty.LocalUnfreeze
}

// NewAddrRow create row
func NewAddrRow() *AddrRow {
	return &AddrRow{LocalUnfreeze: nil}
}

// CreateRow create row
func (r *AddrRow) CreateRow() *table.Row {
	return &table.Row{Data: &pty.LocalUnfreeze{}}
}

// SetPayload set payload
func (r *AddrRow) SetPayload(data types.Message) error {
	if d, ok := data.(*pty.LocalUnfreeze); ok {
		r.LocalUnfreeze = d
		return nil
	}
	return types.ErrTypeAsset
}

// Get get index key
func (r *AddrRow) Get(key string) ([]byte, error) {
	switch key {
	case "txIndex":
		return []byte(r.TxIndex), nil
	case "init":
		return []byte(r.Unfreeze.Initiator), nil
	case "beneficiary":
		return []byte(r.Unfreeze.Beneficiary), nil
	case "id":
		return []byte(r.Unfreeze.UnfreezeID), nil
	default:
		return nil, types.ErrNotFound
	}
}

// NewAddrTable create order table
func NewAddrTable(kvdb dbm.KV) *table.Table {
	rowMeta := NewAddrRow()
	rowMeta.SetPayload(&pty.LocalUnfreeze{})
	t, err := table.NewTable(rowMeta, kvdb, opt_addr_table)
	if err != nil {
		panic(err)
	}
	return t
}

func update(ldb *table.Table, unfreeze *pty.Unfreeze) error {
	xs, err := ldb.ListIndex("id", []byte(unfreeze.UnfreezeID), nil, 1, 0)
	if err != nil || len(xs) != 1 {
		uflog.Error("update query List failed", "key", unfreeze.UnfreezeID, "err", err, "len", len(xs))
		return nil
	}
	u, ok := xs[0].Data.(*pty.LocalUnfreeze)
	if !ok {
		uflog.Error("update decode failed", "data", xs[0].Data)
		return nil

	}
	u.Unfreeze = unfreeze
	return ldb.Update([]byte(u.TxIndex), u)
}

func list(db dbm.KVDB, indexName string, data *pty.LocalUnfreeze, count, direction int32) ([]*table.Row, error) {
	query := NewAddrTable(db).GetQuery(db)
	var primary []byte
	if len(data.TxIndex) > 0 {
		primary = []byte(data.TxIndex)
	}

	cur := &AddrRow{LocalUnfreeze: data}
	index, err := cur.Get(indexName)
	if err != nil {
		uflog.Error("query List failed", "key", string(primary), "param", data, "err", err)
		return nil, err
	}
	uflog.Debug("query List dbg", "indexName", indexName, "index", string(index), "primary", primary, "count", count, "direction", direction)
	rows, err := query.ListIndex(indexName, index, primary, count, direction)
	if err != nil {
		uflog.Error("query List failed", "key", string(primary), "param", data, "err", err)
		return nil, err
	}
	if len(rows) == 0 {
		return nil, types.ErrNotFound
	}
	return rows, nil
}
