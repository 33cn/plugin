// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gty "github.com/33cn/plugin/plugin/dapp/guess/types"
)

func (g *Guess) rollbackIndex(log *gty.ReceiptGuessGame) (kvs []*types.KeyValue) {
	//新创建游戏，将增加的记录都删除掉
	if log.Status == gty.GuessGameStatusStart {
		//kvs = append(kvs, addGuessGameAddrIndexKey(log.Status, log.Addr, log.GameId, log.Index))
		kvs = append(kvs, delGuessGameStatusIndexKey(log.Status, log.Index))
		kvs = append(kvs, delGuessGameAdminIndexKey(log.AdminAddr, log.Index))
		kvs = append(kvs, delGuessGameAdminStatusIndexKey(log.Status, log.AdminAddr, log.Index))
		kvs = append(kvs, delGuessGameCategoryStatusIndexKey(log.Status, log.Category, log.Index))
	} else if log.Status == gty.GuessGameStatusBet {
		//如果是下注状态，则有用户进行了下注操作，对这些记录进行删除
		kvs = append(kvs, delGuessGameAddrIndexKey(log.Addr, log.Index))
		kvs = append(kvs, delGuessGameAddrStatusIndexKey(log.Status, log.Addr, log.Index))

		//如果发生了状态变化，恢复老状态的记录，删除新添加的状态记录
		if log.StatusChange {
			kvs = append(kvs, addGuessGameStatusIndexKey(log.PreStatus, log.GameID, log.PreIndex))
			kvs = append(kvs, addGuessGameAdminStatusIndexKey(log.PreStatus, log.AdminAddr, log.GameID, log.PreIndex))
			kvs = append(kvs, addGuessGameCategoryStatusIndexKey(log.PreStatus, log.Category, log.GameID, log.PreIndex))

			kvs = append(kvs, delGuessGameStatusIndexKey(log.Status, log.Index))
			kvs = append(kvs, delGuessGameAdminStatusIndexKey(log.Status, log.AdminAddr, log.Index))
			kvs = append(kvs, delGuessGameCategoryStatusIndexKey(log.Status, log.Category, log.Index))
		}
	} else if log.StatusChange {
		//其他状态时的状态发生变化的情况,要将老状态对应的记录恢复，同时删除新加的状态记录；对于每个地址的下注记录也需要遍历处理。
		kvs = append(kvs, addGuessGameStatusIndexKey(log.PreStatus, log.GameID, log.PreIndex))
		kvs = append(kvs, addGuessGameAdminStatusIndexKey(log.PreStatus, log.AdminAddr, log.GameID, log.PreIndex))
		kvs = append(kvs, addGuessGameCategoryStatusIndexKey(log.PreStatus, log.Category, log.GameID, log.PreIndex))

		kvs = append(kvs, delGuessGameStatusIndexKey(log.Status, log.Index))
		kvs = append(kvs, delGuessGameAdminStatusIndexKey(log.Status, log.AdminAddr, log.Index))
		kvs = append(kvs, delGuessGameCategoryStatusIndexKey(log.Status, log.Category, log.Index))

		//从game中遍历每个地址的记录进行删除新增记录，回复老记录
		game, err := readGame(g.GetStateDB(), log.GameID)
		if err == nil {
			for i := 0; i < len(game.Plays); i++ {
				player := game.Plays[i]
				kvs = append(kvs, addGuessGameAddrStatusIndexKey(log.PreStatus, player.Addr, log.GameID, player.Bet.PreIndex))
				kvs = append(kvs, delGuessGameAddrStatusIndexKey(log.Status, player.Addr, log.Index))
			}
		}
	}

	return kvs
}

func (g *Guess) execDelLocal(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.GetTy() != types.ExecOk {
		return dbSet, nil
	}

	for _, log := range receiptData.Logs {
		switch log.GetTy() {
		case gty.TyLogGuessGameStart, gty.TyLogGuessGameBet, gty.TyLogGuessGameStopBet, gty.TyLogGuessGameAbort, gty.TyLogGuessGamePublish, gty.TyLogGuessGameTimeout:
			receiptGame := &gty.ReceiptGuessGame{}
			if err := types.Decode(log.Log, receiptGame); err != nil {
				return nil, err
			}
			kv := g.rollbackIndex(receiptGame)
			dbSet.KV = append(dbSet.KV, kv...)
		}
	}
	return dbSet, nil
}

//ExecDelLocal_Start Guess执行器Start交易撤销
func (g *Guess) ExecDelLocal_Start(payload *gty.GuessGameStart, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

//ExecDelLocal_Bet Guess执行器Bet交易撤销
func (g *Guess) ExecDelLocal_Bet(payload *gty.GuessGameBet, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

//ExecDelLocal_Publish Guess执行器Publish交易撤销
func (g *Guess) ExecDelLocal_Publish(payload *gty.GuessGamePublish, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

//ExecDelLocal_Abort Guess执行器Abort交易撤销
func (g *Guess) ExecDelLocal_Abort(payload *gty.GuessGameAbort, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}
