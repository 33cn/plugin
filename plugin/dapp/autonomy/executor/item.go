// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/pkg/errors"
)

func (a *Autonomy) execAutoLocalItem(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	set, err := a.execLocalItem(receiptData)
	if err != nil {
		return set, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = a.AddRollbackKV(tx, tx.Execer, set.KV)
	return dbSet, nil
}

func (a *Autonomy) execLocalItem(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	table := NewItemTable(a.GetLocalDB())
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case auty.TyLogPropItem,
			auty.TyLogRvkPropItem,
			auty.TyLogVotePropItem,
			auty.TyLogTmintPropItem:
			{
				var receipt auty.ReceiptProposalItem
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

func (a *Autonomy) listProposalItem(req *auty.ReqQueryProposalItem) (types.Message, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}
	localDb := a.GetLocalDB()
	query := NewItemTable(localDb).GetQuery(localDb)
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

	cur := &ItemRow{
		AutonomyProposalItem: &auty.AutonomyProposalItem{},
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

	var rep auty.ReplyQueryProposalItem
	for _, row := range rows {
		r, ok := row.Data.(*auty.AutonomyProposalItem)
		if !ok {
			alog.Error("listProposalItem", "err", "bad row type")
			return nil, types.ErrDecode
		}
		rep.PropItems = append(rep.PropItems, r)
	}
	return &rep, nil
}

func getProposalItem(db dbm.KV, req *types.ReqString) (*auty.ReplyQueryProposalItem, error) {
	if req == nil || len(req.Data) <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "invalid parameter")
	}
	value, err := db.Get(propItemID(req.Data))
	if err != nil {
		return nil, errors.Wrapf(err, "fail,db.get item id=%s", req.Data)
	}
	prop := &auty.AutonomyProposalItem{}
	err = types.Decode(value, prop)
	if err != nil {
		return nil, errors.Wrapf(err, "decode item fail")
	}
	rep := &auty.ReplyQueryProposalItem{}
	rep.PropItems = append(rep.PropItems, prop)
	return rep, nil
}

// IsAutonomyApprovedItem get 2 parameters: autonomyItemID, applyTxHash
func IsAutonomyApprovedItem(db dbm.KV, req *types.ReqMultiStrings) (types.Message, error) {
	if req == nil {
		return nil, errors.Wrapf(types.ErrInvalidParam, "req is nil")
	}

	if len(req.Datas) < 2 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "req datas less 2 parameters")
	}

	autonomyItemID := req.Datas[0]
	applyTxHash := req.Datas[1]
	res, err := getProposalItem(db, &types.ReqString{Data: autonomyItemID})
	if err != nil {
		return nil, err
	}
	if len(res.GetPropItems()) <= 0 {
		return nil, errors.Wrapf(types.ErrNotFound, "not get item")
	}
	if res.PropItems[0].ProposalID != autonomyItemID {
		return nil, errors.Wrapf(types.ErrInvalidParam, "get prop id=%s not equal req=%s", res.PropItems[0].ProposalID, autonomyItemID)
	}
	if res.PropItems[0].PropItem.ItemTxHash != applyTxHash {
		return nil, errors.Wrapf(types.ErrInvalidParam, "get item id=%s != req=%s", res.PropItems[0].PropItem.ItemTxHash, applyTxHash)
	}

	if res.PropItems[0].Status == auty.AutonomyStatusTmintPropItem && res.PropItems[0].BoardVoteRes.Pass {
		return &types.Reply{IsOk: true}, nil
	}

	if res.PropItems[0].Status != auty.AutonomyStatusTmintPropItem {
		return nil, errors.Wrapf(types.ErrNotAllow, "item status =%d not terminate", res.PropItems[0].Status)
	}

	return nil, errors.Wrapf(types.ErrNotAllow, "item vote status not pass = %v", res.PropItems[0].BoardVoteRes.Pass)
}
