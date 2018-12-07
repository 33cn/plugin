// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/fingerguessing/types"
)

func (g *Fingerguessing) execLocal(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.GetTy() != types.ExecOk {
		return dbSet, nil
	}
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case gt.TyLogCreateGame, gt.TyLogMatchGame, gt.TyLogCloseGame, gt.TyLogCancleGame:
			receiptGame := &gt.ReceiptGame{}
			if err := types.Decode(log.Log, receiptGame); err != nil {
				return nil, err
			}
			kv := g.updateIndex(receiptGame)
			dbSet.KV = append(dbSet.KV, kv...)
		}
	}
	return dbSet, nil
}

func (g *Fingerguessing) ExecLocal_Create(payload *gt.GameCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

func (g *Fingerguessing) ExecLocal_Cancel(payload *gt.GameCancel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

func (g *Fingerguessing) ExecLocal_Close(payload *gt.GameClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

func (g *Fingerguessing) ExecLocal_Match(payload *gt.GameMatch, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}
