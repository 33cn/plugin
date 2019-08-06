// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

func (a *Autonomy) execAutoLocalBoard(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	set, err := a.execLocalBoard(receiptData)
	if err != nil {
		return set, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = a.AddRollbackKV(tx, []byte(tx.Execer), set.KV)
	return dbSet, nil
}

func (a *Autonomy) execLocalBoard(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	table := NewBoardTable(a.GetLocalDB())
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case auty.TyLogPropBoard,
			auty.TyLogRvkPropBoard,
			auty.TyLogVotePropBoard,
			auty.TyLogTmintPropBoard:
			{
				var receipt auty.ReceiptProposalBoard
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}

				err = table.Replace(receipt.Current)
				if err != nil {
					return nil, err
				}
			}
		default:
			break
		}
	}
	kvs, err := table.Save()
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

func (a *Autonomy) execAutoDelLocal(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	kvs, err := a.DelRollbackKV(tx, []byte(tx.Execer))
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

func (a *Autonomy) execDelLocalBoard(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	table := NewBoardTable(a.GetLocalDB())
	for _, log := range receiptData.Logs {
		var receipt auty.ReceiptProposalBoard
		err := types.Decode(log.Log, &receipt)
		if err != nil {
			return nil, err
		}
		switch log.Ty {
		case auty.TyLogPropBoard:
			{
				heightIndex := dapp.HeightIndexStr(receipt.Current.Height, int64(receipt.Current.Index))
				err = table.Del([]byte(heightIndex))
				if err != nil {
					return nil, err
				}
			}
		case auty.TyLogRvkPropBoard,
			auty.TyLogVotePropBoard,
			auty.TyLogTmintPropBoard:
			{
				err = table.Replace(receipt.Prev)
				if err != nil {
					return nil, err
				}
			}
		default:
			break
		}
	}
	kvs, err := table.Save()
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

func (a *Autonomy) getProposalBoard(req *types.ReqString) (types.Message, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}
	value, err := a.GetStateDB().Get(propBoardID(req.Data))
	if err != nil {
		return nil, err
	}
	prop := &auty.AutonomyProposalBoard{}
	err = types.Decode(value, prop)
	if err != nil {
		return nil, err
	}
	rep := &auty.ReplyQueryProposalBoard{}
	rep.PropBoards = append(rep.PropBoards, prop)
	return rep, nil
}

func (a *Autonomy) listProposalBoard(req *auty.ReqQueryProposalBoard) (types.Message, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}

	localDb := a.GetLocalDB()
	query := NewBoardTable(localDb).GetQuery(localDb)
	var primary []byte
	if req.Height > 0 {
		primary = []byte(dapp.HeightIndexStr(req.Height, int64(req.Index)))
	}
	indexName := ""
	if req.Status > 0 && req.Addr != "" {
		indexName = "addr_status"
	} else if req.Status > 0 {
		indexName = "status"
	} else if req.Addr != "" {
		indexName = "addr"
	}

	cur := &BoardRow{
		AutonomyProposalBoard: &auty.AutonomyProposalBoard{},
	}
	cur.Address = req.Addr
	cur.Status = req.Status
	cur.Height = req.Height
	cur.Index = req.Index
	prefix, err := cur.Get(indexName)
	if err != nil {
		alog.Error("Get", "indexName", indexName, "err", err)
		return nil, err
	}

	rows, err := query.ListIndex(indexName, prefix, primary, req.Count, req.Direction)
	if err != nil {
		alog.Error("query List failed", "indexName", indexName, "prefix", "prefix", "key", string(primary), "err", err)
		return nil, err
	}
	if len(rows) == 0 {
		return nil, types.ErrNotFound
	}

	var rep auty.ReplyQueryProposalBoard
	for _, row := range rows {
		r, ok := row.Data.(*auty.AutonomyProposalBoard)
		if !ok {
			alog.Error("listProposalBoard", "err", "bad row type")
			return nil, types.ErrDecode
		}
		rep.PropBoards = append(rep.PropBoards, r)
	}
	return &rep, nil
}
