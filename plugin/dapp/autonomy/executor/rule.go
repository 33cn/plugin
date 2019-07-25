// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

func (a *Autonomy) execLocalRule(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	var set []*types.KeyValue
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case auty.TyLogPropRule,
			auty.TyLogRvkPropRule,
			auty.TyLogVotePropRule,
			auty.TyLogTmintPropRule:
			{
				var receipt auty.ReceiptProposalRule
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv := saveRuleHeightIndex(&receipt)
				set = append(set, kv...)
			}
		default:
			break
		}
	}
	dbSet.KV = append(dbSet.KV, set...)
	return dbSet, nil
}

func saveRuleHeightIndex(res *auty.ReceiptProposalRule) (kvs []*types.KeyValue) {
	// 先将之前的状态删除掉，再做更新
	if res.Current.Status > 1 {
		kv := &types.KeyValue{}
		kv.Key = calcRuleKey4StatusHeight(res.Prev.Status, dapp.HeightIndexStr(res.Prev.Height, int64(res.Prev.Index)))
		kv.Value = nil
		kvs = append(kvs, kv)
	}

	kv := &types.KeyValue{}
	kv.Key = calcRuleKey4StatusHeight(res.Current.Status, dapp.HeightIndexStr(res.Current.Height, int64(res.Current.Index)))
	kv.Value = types.Encode(res.Current)
	kvs = append(kvs, kv)
	return kvs
}

func (a *Autonomy) execDelLocalRule(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	var set []*types.KeyValue
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case auty.TyLogPropRule,
			auty.TyLogRvkPropRule,
			auty.TyLogVotePropRule,
			auty.TyLogTmintPropRule:
			{
				var receipt auty.ReceiptProposalRule
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv := delRuleHeightIndex(&receipt)
				set = append(set, kv...)
			}
		default:
			break
		}
	}
	dbSet.KV = append(dbSet.KV, set...)
	return dbSet, nil
}

func delRuleHeightIndex(res *auty.ReceiptProposalRule) (kvs []*types.KeyValue) {
	kv := &types.KeyValue{}
	kv.Key = calcRuleKey4StatusHeight(res.Current.Status, dapp.HeightIndexStr(res.Current.Height, int64(res.Current.Index)))
	kv.Value = nil
	kvs = append(kvs, kv)

	if res.Current.Status > 1 {
		kv := &types.KeyValue{}
		kv.Key = calcRuleKey4StatusHeight(res.Prev.Status, dapp.HeightIndexStr(res.Prev.Height, int64(res.Prev.Index)))
		kv.Value = types.Encode(res.Prev)
		kvs = append(kvs, kv)
	}
	return kvs
}

func (a *Autonomy) getProposalRule(req *types.ReqString) (types.Message, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}
	value, err := a.GetStateDB().Get(propRuleID(req.Data))
	if err != nil {
		return nil, err
	}
	prop := &auty.AutonomyProposalRule{}
	err = types.Decode(value, prop)
	if err != nil {
		return nil, err
	}
	rep := &auty.ReplyQueryProposalRule{}
	rep.PropRules = append(rep.PropRules, prop)
	return rep, nil
}

func (a *Autonomy) listProposalRule(req *auty.ReqQueryProposalRule) (types.Message, error) {
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
		key    = calcRuleKey4StatusHeight(req.Status, heightstr)
	}
	prefix := calcRuleKey4StatusHeight(req.Status, "")
	values, err = localDb.List(prefix, key, req.Count, req.GetDirection())
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, types.ErrNotFound
	}

	var rep auty.ReplyQueryProposalRule
	for _, value := range values {
		prop := &auty.AutonomyProposalRule{}
		err = types.Decode(value, prop)
		if err != nil {
			return nil, err
		}
		rep.PropRules = append(rep.PropRules, prop)
	}
	return &rep, nil
}

func (a *Autonomy) execLocalCommentProp(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	var set []*types.KeyValue
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case auty.TyLogCommentProp:
			{
				var receipt auty.ReceiptProposalComment
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv := saveCommentHeightIndex(&receipt)
				set = append(set, kv...)
			}
		default:
			break
		}
	}
	dbSet.KV = append(dbSet.KV, set...)
	return dbSet, nil
}

func saveCommentHeightIndex(res *auty.ReceiptProposalComment) (kvs []*types.KeyValue) {
	kv := &types.KeyValue{}
	kv.Key = calcCommentHeight(res.Cmt.ProposalID, dapp.HeightIndexStr(res.Height, int64(res.Index)))
	kv.Value = types.Encode(&auty.RelationCmt{
		RepCmtHash: res.Cmt.RepCmtHash,
		Comment: res.Cmt.Comment,
		Height: res.Height,
		Index: res.Index,
	})
	kvs = append(kvs, kv)
	return kvs
}

func (a *Autonomy) execDelLocalCommentProp(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	var set []*types.KeyValue
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case auty.TyLogCommentProp:
			{
				var receipt auty.ReceiptProposalComment
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv := delCommentHeightIndex(&receipt)
				set = append(set, kv...)
			}
		default:
			break
		}
	}
	dbSet.KV = append(dbSet.KV, set...)
	return dbSet, nil
}

func delCommentHeightIndex(res *auty.ReceiptProposalComment) (kvs []*types.KeyValue) {
	kv := &types.KeyValue{}
	kv.Key = calcCommentHeight(res.Cmt.ProposalID, dapp.HeightIndexStr(res.Height, int64(res.Index)))
	kv.Value = nil
	kvs = append(kvs, kv)
	return kvs
}

func (a *Autonomy) listProposalComment(req *auty.ReqQueryProposalComment) (types.Message, error) {
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
		key    = calcCommentHeight(req.ProposalID, heightstr)
	}
	prefix := calcCommentHeight(req.ProposalID, "")
	values, err = localDb.List(prefix, key, req.Count, req.GetDirection())
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, types.ErrNotFound
	}

	var rep auty.ReplyQueryProposalComment
	for _, value := range values {
		cmt := &auty.RelationCmt{}
		err = types.Decode(value, cmt)
		if err != nil {
			return nil, err
		}
		rep.RltCmt = append(rep.RltCmt, cmt)
	}
	return &rep, nil
}