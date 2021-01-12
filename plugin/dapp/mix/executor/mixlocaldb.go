// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

func (e *Mix) execAutoLocalMix(tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if receiptData.Ty != types.ExecOk {
		return nil, types.ErrInvalidParam
	}
	set, err := e.execLocalMix(tx, receiptData, index)
	if err != nil {
		return set, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = e.AddRollbackKV(tx, tx.Execer, set.KV)
	return dbSet, nil
}

func (e *Mix) execLocalMix(tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	table := NewMixTxTable(e.GetLocalDB())

	r := &mixTy.LocalMixTx{
		Hash:   common.ToHex(tx.Hash()),
		Height: e.GetHeight(),
		Index:  int64(index),
	}
	err := table.Add(r)
	if err != nil {
		return nil, err
	}

	kvs, err := table.Save()
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

func (e *Mix) execAutoDelLocal(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	kvs, err := e.DelRollbackKV(tx, tx.Execer)
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

func (e *Mix) listMixInfos(req *mixTy.MixTxListReq) (types.Message, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}
	localDb := e.GetLocalDB()
	query := NewMixTxTable(localDb).GetQuery(localDb)
	var primary []byte
	if len(req.TxIndex) > 0 {
		primary = []byte(req.TxIndex)
	}

	cur := &MixTxRow{}
	indexName := "height"
	var prefix []byte
	info := &mixTy.LocalMixTx{Height: req.Height, Index: req.Index}
	if len(req.Hash) > 0 {
		info.Hash = req.Hash
		indexName = "hash"
		var err error
		prefix, err = cur.Get(indexName)
		if err != nil {
			mlog.Error("listMixInfos Get", "indexName", indexName, "err", err)
			return nil, err
		}
	}
	cur.SetPayload(info)
	rows, err := query.ListIndex(indexName, prefix, primary, req.Count, req.Direction)
	if err != nil {
		mlog.Error("listMixInfos query failed", "indexName", indexName, "prefix", string(prefix), "key", string(primary), "err", err)
		return nil, err
	}
	if len(rows) == 0 {
		return nil, types.ErrNotFound
	}
	var rep mixTy.MixTxListResp
	for _, row := range rows {
		r, ok := row.Data.(*mixTy.LocalMixTx)
		if !ok {
			mlog.Error("listMixInfos", "err", "bad row type")
			return nil, types.ErrDecode
		}
		rep.Txs = append(rep.Txs, r)
	}
	return &rep, nil
}
