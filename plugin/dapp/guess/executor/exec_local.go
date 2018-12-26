// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gty "github.com/33cn/plugin/plugin/dapp/guess/types"
)

func (g *Guess) updateIndex(log *gty.ReceiptGuessGame) (kvs []*types.KeyValue) {
	//新创建游戏
	if log.Status == gty.GuessGameStatusStart {
		//kvs = append(kvs, addGuessGameAddrIndexKey(log.Status, log.Addr, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameStatusIndexKey(log.Status, log.GameID, log.Index))
		kvs = append(kvs, addGuessGameAdminIndexKey(log.Status, log.AdminAddr, log.GameID, log.Index))
		kvs = append(kvs, addGuessGameAdminStatusIndexKey(log.Status, log.AdminAddr, log.GameID, log.Index))
		kvs = append(kvs, addGuessGameCategoryStatusIndexKey(log.Status, log.Category, log.GameID, log.Index))
	} else if log.Status == gty.GuessGameStatusBet {
		//如果是下注状态，则有用户进行了下注操作
		kvs = append(kvs, addGuessGameAddrIndexKey(log.Status, log.Addr, log.GameID, log.Index))
		kvs = append(kvs, addGuessGameAddrStatusIndexKey(log.Status, log.Addr, log.GameID, log.Index))
		//如果发生了状态变化，则是从start->bet，对于老状态的记录进行删除操作，并增加新状态记录
		if log.StatusChange {
			kvs = append(kvs, addGuessGameStatusIndexKey(log.Status, log.GameID, log.Index))
			kvs = append(kvs, addGuessGameAdminStatusIndexKey(log.Status, log.AdminAddr, log.GameID, log.Index))
			kvs = append(kvs, addGuessGameCategoryStatusIndexKey(log.Status, log.Category, log.GameID, log.Index))

			kvs = append(kvs, delGuessGameStatusIndexKey(log.PreStatus, log.PreIndex))
			kvs = append(kvs, delGuessGameAdminStatusIndexKey(log.PreStatus, log.AdminAddr, log.PreIndex))
			kvs = append(kvs, delGuessGameCategoryStatusIndexKey(log.PreStatus, log.Category, log.PreIndex))
		}
	} else if log.StatusChange {
		//其他状态时的状态发生变化,要将老状态对应的记录删除，同时加入新状态记录；对于每个地址的下注记录也需要遍历处理。
		kvs = append(kvs, addGuessGameStatusIndexKey(log.Status, log.GameID, log.Index))
		kvs = append(kvs, addGuessGameAdminStatusIndexKey(log.Status, log.AdminAddr, log.GameID, log.Index))
		kvs = append(kvs, addGuessGameCategoryStatusIndexKey(log.Status, log.Category, log.GameID, log.Index))

		kvs = append(kvs, delGuessGameStatusIndexKey(log.PreStatus, log.PreIndex))
		kvs = append(kvs, delGuessGameAdminStatusIndexKey(log.PreStatus, log.AdminAddr, log.PreIndex))
		kvs = append(kvs, delGuessGameCategoryStatusIndexKey(log.PreStatus, log.Category, log.PreIndex))

		//从game中遍历每个地址的记录进行新状态记录的增和老状态记录的删除
		game, err := readGame(g.GetStateDB(), log.GameID)
		if err == nil {
			for i := 0; i < len(game.Plays); i++ {
				player := game.Plays[i]
				kvs = append(kvs, addGuessGameAddrStatusIndexKey(log.Status, player.Addr, log.GameID, log.Index))
				kvs = append(kvs, delGuessGameAddrStatusIndexKey(log.PreStatus, player.Addr, player.Bet.PreIndex))
			}
		}
	}

	return kvs
}

func (g *Guess) execLocal(receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return dbSet, nil
	}

	/*
	for i := 0; i < len(receipt.Logs); i++ {
		item := receipt.Logs[i]
		if item.Ty >= gty.TyLogGuessGameStart && item.Ty <= gty.TyLogGuessGameTimeout {
			var Gamelog gty.ReceiptGuessGame
			err := types.Decode(item.Log, &Gamelog)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := g.updateIndex(&Gamelog)
			dbSet.KV = append(dbSet.KV, kv...)
		}
	}
	*/

	table := gty.NewTable(g.GetLocalDB())
	for _, item := range receipt.Logs {
		if item.Ty >= gty.TyLogGuessGameStart && item.Ty <= gty.TyLogGuessGameTimeout {
			var gameLog gty.ReceiptGuessGame
			err := types.Decode(item.Log, &gameLog)
			if err != nil {
				return nil, err
			}
			err = table.Replace(&gameLog)
			if err != nil {
				return nil, err
			}
			kvs, err := table.Save()
			if err != nil {
				return nil, err
			}
			dbSet.KV = append(dbSet.KV, kvs...)
		}
	}

	return dbSet, nil
}

//ExecLocal_Start method
func (g *Guess) ExecLocal_Start(payload *gty.GuessGameStart, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

//ExecLocal_Bet method
func (g *Guess) ExecLocal_Bet(payload *gty.GuessGameBet, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

//ExecLocal_StopBet method
func (g *Guess) ExecLocal_StopBet(payload *gty.GuessGameStopBet, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

//ExecLocal_Publish method
func (g *Guess) ExecLocal_Publish(payload *gty.GuessGamePublish, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}

//ExecLocal_Abort method
func (g *Guess) ExecLocal_Abort(payload *gty.GuessGameAbort, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return g.execLocal(receiptData)
}
