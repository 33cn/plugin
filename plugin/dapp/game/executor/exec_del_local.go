// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/game/types"
)

// roll back local db data
func (g *Game) execDelLocal(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	for _, log := range receiptData.Logs {
		switch log.GetTy() {
		case gt.TyLogCreateGame, gt.TyLogMatchGame, gt.TyLogCloseGame, gt.TyLogCancleGame:
			receiptGame := &gt.ReceiptGame{}
			if err := types.Decode(log.Log, receiptGame); err != nil {
				return nil, err
			}
			kv := g.rollbackIndex(receiptGame)
			dbSet.KV = append(dbSet.KV, kv...)
		}
	}
	return dbSet, nil
}

// ExecDelLocal_Create roll back local db data for create
func (g *Game) ExecDelLocal_Create(payload *gt.GameCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execDelLocal(receiptData)
}

// ExecDelLocal_Cancel roll back local db data for cancel
func (g *Game) ExecDelLocal_Cancel(payload *gt.GameCancel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execDelLocal(receiptData)
}

// ExecDelLocal_Close roll back local db data for close
func (g *Game) ExecDelLocal_Close(payload *gt.GameClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execDelLocal(receiptData)
}

// ExecDelLocal_Match roll back local db data for match
func (g *Game) ExecDelLocal_Match(payload *gt.GameMatch, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execDelLocal(receiptData)
}
