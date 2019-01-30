// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	gty "github.com/33cn/plugin/plugin/dapp/guess/types"
)

func (g *Guess) rollbackGame(game *gty.GuessGame, log *gty.ReceiptGuessGame) {
	if game == nil || log == nil {
		return
	}

	//如果状态发生了变化，则需要将游戏状态恢复到前一状态
	if log.StatusChange {
		game.Status = log.PreStatus
		game.Index = log.PreIndex

		//玩家信息中的index回滚
		for i := 0; i < len(game.Plays); i++ {
			player := game.Plays[i]
			player.Bet.Index = player.Bet.PreIndex
		}
	}

	//如果下注了，则需要把下注回滚
	if log.Bet {
		//统计信息回滚
		game.BetStat.TotalBetTimes--
		game.BetStat.TotalBetsNumber -= log.BetsNumber
		for i := 0; i < len(game.BetStat.Items); i++ {
			item := game.BetStat.Items[i]
			if item.Option == log.Option {
				item.BetsTimes--
				item.BetsNumber -= log.BetsNumber
				break
			}
		}

		//玩家下注信息回滚
		for i := 0; i < len(game.Plays); i++ {
			player := game.Plays[i]
			if player.Addr == log.Addr && player.Bet.Index == log.Index {
				game.Plays = append(game.Plays[:i], game.Plays[i+1:]...)
				break
			}
		}

	}
}

func (g *Guess) rollbackIndex(log *gty.ReceiptGuessGame) (kvs []*types.KeyValue, err error) {
	userTable := gty.NewGuessUserTable(g.GetLocalDB())
	gameTable := gty.NewGuessGameTable(g.GetLocalDB())

	tableJoin, err := table.NewJoinTable(userTable, gameTable, []string{"addr#status"})
	if err != nil {
		return nil, err
	}

	if log.Status == gty.GuessGameStatusStart {
		//新创建游戏回滚,game表删除记录
		err = gameTable.Del([]byte(fmt.Sprintf("%018d", log.StartIndex)))
		if err != nil {
			return nil, err
		}
		kvs, err = tableJoin.Save()
		return kvs, err
	} else if log.Status == gty.GuessGameStatusBet {
		//下注阶段，需要更新游戏信息，回滚下注信息
		game := log.Game
		log.Game = nil

		//先回滚游戏信息，再进行更新
		g.rollbackGame(game, log)

		err = tableJoin.MustGetTable("game").Replace(game)
		if err != nil {
			return nil, err
		}

		err = tableJoin.MustGetTable("user").Del([]byte(fmt.Sprintf("%018d", log.Index)))
		if err != nil {
			return nil, err
		}

		kvs, err = tableJoin.Save()
		if err != nil {
			return nil, err
		}
	} else if log.StatusChange {
		//如果是其他状态下仅发生了状态变化，则需要恢复游戏状态，并更新游戏记录。
		game := log.Game
		log.Game = nil

		//先回滚游戏信息，再进行更新
		g.rollbackGame(game, log)

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

func (g *Guess) execDelLocal(receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return dbSet, nil
	}

	for _, log := range receipt.Logs {
		switch log.GetTy() {
		case gty.TyLogGuessGameStart, gty.TyLogGuessGameBet, gty.TyLogGuessGameStopBet, gty.TyLogGuessGameAbort, gty.TyLogGuessGamePublish, gty.TyLogGuessGameTimeout:
			receiptGame := &gty.ReceiptGuessGame{}
			if err := types.Decode(log.Log, receiptGame); err != nil {
				return nil, err
			}
			kv, err := g.rollbackIndex(receiptGame)
			if err != nil {
				return nil, err
			}
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
