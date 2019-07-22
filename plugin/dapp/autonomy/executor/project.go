// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

func (a *Autonomy) execLocalProject(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	var set []*types.KeyValue
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case auty.TyLogPropProject,
			auty.TyLogRvkPropProject,
			auty.TyLogVotePropProject,
			auty.TyLogPubVotePropProject,
			auty.TyLogTmintPropProject:
			{
				var receipt auty.ReceiptProposalProject
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv := saveProjectHeightIndex(&receipt)
				set = append(set, kv...)
			}
		default:
			break
		}
	}
	dbSet.KV = append(dbSet.KV, set...)
	return dbSet, nil
}

func saveProjectHeightIndex(res *auty.ReceiptProposalProject) (kvs []*types.KeyValue) {
	// 先将之前的状态删除掉，再做更新
	if res.Current.Status > 1 {
		kv := &types.KeyValue{}
		kv.Key = calcProjectKey4StatusHeight(res.Prev.Status, dapp.HeightIndexStr(res.Prev.Height, int64(res.Prev.Index)))
		kv.Value = nil
		kvs = append(kvs, kv)
	}

	kv := &types.KeyValue{}
	kv.Key = calcProjectKey4StatusHeight(res.Current.Status, dapp.HeightIndexStr(res.Current.Height, int64(res.Current.Index)))
	kv.Value = types.Encode(res.Current)
	kvs = append(kvs, kv)
	return kvs
}

func (a *Autonomy) execDelLocalProject(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	var set []*types.KeyValue
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case auty.TyLogPropProject,
			auty.TyLogRvkPropProject,
			auty.TyLogVotePropProject,
			auty.TyLogPubVotePropProject,
			auty.TyLogTmintPropProject:
			{
				var receipt auty.ReceiptProposalProject
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv := delProjectHeightIndex(&receipt)
				set = append(set, kv...)
			}
		default:
			break
		}
	}
	dbSet.KV = append(dbSet.KV, set...)
	return dbSet, nil
}

func delProjectHeightIndex(res *auty.ReceiptProposalProject) (kvs []*types.KeyValue) {
	kv := &types.KeyValue{}
	kv.Key = calcProjectKey4StatusHeight(res.Current.Status, dapp.HeightIndexStr(res.Current.Height, int64(res.Current.Index)))
	kv.Value = nil
	kvs = append(kvs, kv)

	if res.Current.Status > 1 {
		kv := &types.KeyValue{}
		kv.Key = calcProjectKey4StatusHeight(res.Prev.Status, dapp.HeightIndexStr(res.Prev.Height, int64(res.Prev.Index)))
		kv.Value = types.Encode(res.Prev)
		kvs = append(kvs, kv)
	}
	return kvs
}

func (a *Autonomy) getProposalProject(req *types.ReqString) (types.Message, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}
	value, err := a.GetStateDB().Get(propProjectID(req.Data))
	if err != nil {
		return nil, err
	}
	prop := &auty.AutonomyProposalProject{}
	err = types.Decode(value, prop)
	if err != nil {
		return nil, err
	}
	rep := &auty.ReplyQueryProposalProject{}
	rep.PropProjects = append(rep.PropProjects, prop)
	return rep, nil
}

func (a *Autonomy) listProposalProject(req *auty.ReqQueryProposalProject) (types.Message, error) {
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
		key    = calcProjectKey4StatusHeight(req.Status, heightstr)
	}
	prefix := calcProjectKey4StatusHeight(req.Status, "")
	values, err = localDb.List(prefix, key, req.Count, req.GetDirection())
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, types.ErrNotFound
	}

	var rep auty.ReplyQueryProposalProject
	for _, value := range values {
		prop := &auty.AutonomyProposalProject{}
		err = types.Decode(value, prop)
		if err != nil {
			return nil, err
		}
		rep.PropProjects = append(rep.PropProjects, prop)
	}
	return &rep, nil
}