// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

func (e *Paracross) execAutoLocalStage(tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := e.execLocalStage(receiptData, index)
	if err != nil {
		return set, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = e.AddRollbackKV(tx, tx.Execer, set.KV)
	return dbSet, nil
}

func (e *Paracross) execLocalStage(receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	table := NewStageTable(e.GetLocalDB())
	txIndex := dapp.HeightIndexStr(e.GetHeight(), int64(index))
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case pt.TyLogParaSelfConsStageConfig:
			{
				var receipt pt.ReceiptSelfConsStageConfig
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				r := &pt.LocalSelfConsStageInfo{
					Stage:   receipt.Current,
					TxIndex: txIndex,
				}
				err = table.Replace(r)
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

func (e *Paracross) execAutoDelLocal(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	kvs, err := e.DelRollbackKV(tx, tx.Execer)
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

func (e *Paracross) listSelfStages(req *pt.ReqQuerySelfStages) (types.Message, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}
	localDb := e.GetLocalDB()
	query := NewStageTable(localDb).GetQuery(localDb)
	var primary []byte
	if req.Height > 0 {
		primary = []byte(dapp.HeightIndexStr(req.Height, int64(req.Index)))
	}
	indexName := ""
	if req.Status > 0 {
		indexName = "status"
	} else if req.Id != "" {
		indexName = "id"
	}

	cur := &StageRow{
		LocalSelfConsStageInfo: &pt.LocalSelfConsStageInfo{},
	}

	cur.Stage.Status = req.Status
	cur.Stage.Id = req.Id
	prefix, err := cur.Get(indexName)
	if err != nil {
		clog.Error("Get", "indexName", indexName, "err", err)
		return nil, err
	}

	rows, err := query.ListIndex(indexName, prefix, primary, req.Count, req.Direction)
	if err != nil {
		clog.Error("query List failed", "indexName", indexName, "prefix", "prefix", "key", string(primary), "err", err)
		return nil, err
	}
	if len(rows) == 0 {
		return nil, types.ErrNotFound
	}

	var rep pt.ReplyQuerySelfStages
	for _, row := range rows {
		r, ok := row.Data.(*pt.LocalSelfConsStageInfo)
		if !ok {
			clog.Error("listProposalProject", "err", "bad row type")
			return nil, types.ErrDecode
		}
		ok, txID, _ := getRealTxHashID(r.Stage.Id)
		if ok {
			r.Stage.Id = txID
		}
		rep.StageInfo = append(rep.StageInfo, r.Stage)
	}
	return &rep, nil
}
