// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

func (a *Autonomy) execAutoLocalProject(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	set, err := a.execLocalProject(receiptData)
	if err != nil {
		return set, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = a.AddRollbackKV(tx, tx.Execer, set.KV)
	return dbSet, nil
}

func (a *Autonomy) execLocalProject(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	table := NewProjectTable(a.GetLocalDB())
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
	localDb := a.GetLocalDB()
	query := NewProjectTable(localDb).GetQuery(localDb)
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

	cur := &ProjectRow{
		AutonomyProposalProject: &auty.AutonomyProposalProject{},
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

	var rep auty.ReplyQueryProposalProject
	for _, row := range rows {
		r, ok := row.Data.(*auty.AutonomyProposalProject)
		if !ok {
			alog.Error("listProposalProject", "err", "bad row type")
			return nil, types.ErrDecode
		}
		rep.PropProjects = append(rep.PropProjects, r)
	}
	return &rep, nil
}
