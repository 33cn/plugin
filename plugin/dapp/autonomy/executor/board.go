// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"fmt"
)

func (a *Autonomy) execLocalBoard(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	var set []*types.KeyValue
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
				kv := saveBoardHeightIndex(&receipt)
				set = append(set, kv...)
			}
		default:
			break
		}
	}
	dbSet.KV = append(dbSet.KV, set...)
	return dbSet, nil
}

func saveBoardHeightIndex(res *auty.ReceiptProposalBoard) (kvs []*types.KeyValue) {
	// 先将之前的状态删除掉，再做更新
	if res.Current.Status > 1 {
		kv := &types.KeyValue{}
		kv.Key = calcBoardKey4StatusHeight(res.Prev.Status, dapp.HeightIndexStr(res.Prev.Height, int64(res.Prev.Index)))
		kv.Value = nil
		kvs = append(kvs, kv)
	}

	kv := &types.KeyValue{}
	kv.Key = calcBoardKey4StatusHeight(res.Current.Status, dapp.HeightIndexStr(res.Current.Height, int64(res.Current.Index)))
	kv.Value = types.Encode(res.Current)
	kvs = append(kvs, kv)
	return kvs
}

func (a *Autonomy) execDelLocalBoard(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	var set []*types.KeyValue
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
				kv := delBoardHeightIndex(&receipt)
				set = append(set, kv...)
			}
		default:
			break
		}
	}
	dbSet.KV = append(dbSet.KV, set...)
	return dbSet, nil
}

func delBoardHeightIndex(res *auty.ReceiptProposalBoard) (kvs []*types.KeyValue) {
	kv := &types.KeyValue{}
	kv.Key = calcBoardKey4StatusHeight(res.Current.Status, dapp.HeightIndexStr(res.Current.Height, int64(res.Current.Index)))
	kv.Value = nil
	kvs = append(kvs, kv)

	if res.Current.Status > 1 {
		kv := &types.KeyValue{}
		kv.Key = calcBoardKey4StatusHeight(res.Prev.Status, dapp.HeightIndexStr(res.Prev.Height, int64(res.Prev.Index)))
		kv.Value = types.Encode(res.Prev)
		kvs = append(kvs, kv)
	}
	return kvs
}

func (a *Autonomy) getProposalBoard(req *auty.ReqQueryProposalBoard) (types.Message, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}
	var key []byte
	var values [][]byte
	var err error

	localDb := a.GetLocalDB()
	if req.GetIndex() == -1 {
		key = nil
	} else { //翻页查找指定的txhash列表
		heightstr := genHeightIndexStr(req.GetIndex())
		key    = calcBoardKey4StatusHeight(req.Status, heightstr)
	}
	prefix := calcBoardKey4StatusHeight(req.Status, "")
	values, err = localDb.List(prefix, key, req.Count, req.GetDirection())
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, types.ErrNotFound
	}

	var rep auty.ReplyQueryProposalBoard
	for _, value := range values {
		prop := &auty.AutonomyProposalBoard{}
		err = types.Decode(value, prop)
		if err != nil {
			return nil, err
		}
		rep.PropBoards = append(rep.PropBoards, prop)
	}
	return &rep, nil
}

func genHeightIndexStr(index int64) string {
	return fmt.Sprintf("%018d", index)
}