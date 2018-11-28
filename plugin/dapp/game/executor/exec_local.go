// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/game/types"
)

// save receiptData to local db
func (g *Game) execLocal(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
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

// ExecLocal_Create save receiptData for create
func (g *Game) ExecLocal_Create(payload *gt.GameCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

// ExecLocal_Cancel save receiptData for cancel
func (g *Game) ExecLocal_Cancel(payload *gt.GameCancel, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

// ExecLocal_Close save receiptData for close
func (g *Game) ExecLocal_Close(payload *gt.GameClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

// ExecLocal_Match save receiptData for Match
func (g *Game) ExecLocal_Match(payload *gt.GameMatch, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}
