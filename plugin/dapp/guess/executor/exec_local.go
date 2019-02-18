// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	gty "github.com/33cn/plugin/plugin/dapp/guess/types"
)

func (g *Guess) getUserBet(log *gty.ReceiptGuessGame) (userBet *gty.UserBet) {
	userBet = &gty.UserBet{}
	userBet.StartIndex = log.StartIndex
	userBet.Index = log.Index
	userBet.GameID = log.GameID
	userBet.Addr = log.Addr
	if log.Bet {
		userBet.Option = log.Option
		userBet.BetsNumber = log.BetsNumber
	}

	return userBet
}

func (g *Guess) updateIndex(log *gty.ReceiptGuessGame) (kvs []*types.KeyValue, err error) {
	userTable := gty.NewGuessUserTable(g.GetLocalDB())
	gameTable := gty.NewGuessGameTable(g.GetLocalDB())
	tableJoin, err := table.NewJoinTable(userTable, gameTable, []string{"addr#status"})
	if err != nil {
		return nil, err
	}

	if log.Status == gty.GuessGameStatusStart {
		//新创建游戏,game表新增记录
		game := log.Game
		log.Game = nil

		err = gameTable.Add(game)
		if err != nil {
			return nil, err
		}

		kvs, err = gameTable.Save()
		if err != nil {
			return nil, err
		}
	} else if log.Status == gty.GuessGameStatusBet {
		//用户下注，game表发生更新(game中下注信息有更新)，user表新增下注记录
		game := log.Game
		log.Game = nil
		userBet := g.getUserBet(log)

		err = tableJoin.MustGetTable("game").Replace(game)
		if err != nil {
			return nil, err
		}

		err = tableJoin.MustGetTable("user").Add(userBet)
		if err != nil {
			return nil, err
		}

		kvs, err = tableJoin.Save()
		if err != nil {
			return nil, err
		}
	} else if log.StatusChange {
		//其他状态，游戏状态变化，只需要更新game表
		game := log.Game
		log.Game = nil

		err = tableJoin.MustGetTable("game").Replace(game)
		if err != nil {
			return nil, err
		}

		kvs, err = tableJoin.Save()
		if err != nil {
			return nil, err
		}
	}

	return kvs, nil
}

func (g *Guess) execLocal(receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return dbSet, nil
	}

	for _, item := range receipt.Logs {
		if item.Ty >= gty.TyLogGuessGameStart && item.Ty <= gty.TyLogGuessGameTimeout {
			var gameLog gty.ReceiptGuessGame
			err := types.Decode(item.Log, &gameLog)
			if err != nil {
				return nil, err
			}
			kvs, err := g.updateIndex(&gameLog)
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
