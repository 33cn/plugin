// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

func (a *Autonomy) execAutoLocalRule(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	set, err := a.execLocalRule(receiptData)
	if err != nil {
		return set, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = a.AddRollbackKV(tx, tx.Execer, set.KV)
	return dbSet, nil
}

func (a *Autonomy) execLocalRule(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	table := NewRuleTable(a.GetLocalDB())
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
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
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

	localDb := a.GetLocalDB()
	query := NewRuleTable(localDb).GetQuery(localDb)
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

	cur := &RuleRow{
		AutonomyProposalRule: &auty.AutonomyProposalRule{},
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

	var rep auty.ReplyQueryProposalRule
	for _, row := range rows {
		r, ok := row.Data.(*auty.AutonomyProposalRule)
		if !ok {
			alog.Error("listProposalRule", "err", "bad row type")
			return nil, types.ErrDecode
		}
		rep.PropRules = append(rep.PropRules, r)
	}
	return &rep, nil
}

func (a *Autonomy) getActiveRule() (types.Message, error) {
	rule := &auty.RuleConfig{}
	value, err := a.GetStateDB().Get(activeRuleID())
	cfg := a.GetAPI().GetConfig()
	autoCfg := GetAutonomyParam(cfg, 0)
	if err == nil {
		err = types.Decode(value, rule)
		if err != nil {
			return nil, err
		}
		if rule.PubApproveRatio <= 0 {
			rule.PubApproveRatio = autoCfg.PubApproveRatio
		}
		if rule.PubAttendRatio <= 0 {
			rule.PubAttendRatio = autoCfg.PubAttendRatio
		}
	} else { // 载入系统默认值

		rule.BoardApproveRatio = autoCfg.BoardApproveRatio
		rule.PubOpposeRatio = autoCfg.PubOpposeRatio
		rule.ProposalAmount = autoCfg.ProposalAmount * cfg.GetCoinPrecision()
		rule.LargeProjectAmount = autoCfg.LargeProjectAmount * cfg.GetCoinPrecision()
		rule.PublicPeriod = autoCfg.PublicPeriod
		rule.PubAttendRatio = autoCfg.PubAttendRatio
		rule.PubApproveRatio = autoCfg.PubApproveRatio
	}
	return rule, nil
}

func (a *Autonomy) execAutoLocalCommentProp(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	set, err := a.execLocalCommentProp(receiptData)
	if err != nil {
		return set, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = a.AddRollbackKV(tx, tx.Execer, set.KV)
	return dbSet, nil
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
		RepHash: res.Cmt.RepHash,
		Comment: res.Cmt.Comment,
		Height:  res.Height,
		Index:   res.Index,
		Hash:    res.Hash,
	})
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
	if req.Height <= 0 {
		key = nil
	} else { //翻页查找指定的txhash列表
		heightstr := dapp.HeightIndexStr(req.Height, int64(req.Index))
		key = calcCommentHeight(req.ProposalID, heightstr)
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
