// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/guess/types"
)

func (c *Guess) updateIndex(log *pkt.ReceiptGuessGame) (kvs []*types.KeyValue) {
	//新创建游戏
	if log.Status == pkt.GuessGameStatusStart{
		//kvs = append(kvs, addGuessGameAddrIndexKey(log.Status, log.Addr, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameStatusIndexKey(log.Status, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameAdminIndexKey(log.Status, log.Addr, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameAdminStatusIndexKey(log.Status, log.AdminAddr, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameCategoryStatusIndexKey(log.Status, log.Category, log.GameId, log.Index))
	} else if log.Status == pkt.GuessGameStatusBet {
		//如果是下注状态，则有用户进行了下注操作
		kvs = append(kvs, addGuessGameAddrIndexKey(log.Status, log.Addr, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameAddrStatusIndexKey(log.Status, log.Addr, log.GameId, log.Index))

		kvs = append(kvs, addGuessGameStatusIndexKey(log.Status, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameAdminStatusIndexKey(log.Status, log.AdminAddr, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameAddrStatusIndexKey(log.Status, log.Addr, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameCategoryStatusIndexKey(log.Status, log.Category, log.GameId, log.Index))
		//对于老状态进行删除
		kvs = append(kvs, delGuessGameStatusIndexKey(log.PreStatus, log.PreIndex))
		kvs = append(kvs, delGuessGameAdminStatusIndexKey(log.PreStatus, log.AdminAddr, log.PreIndex))
		kvs = append(kvs, delGuessGameCategoryStatusIndexKey(log.PreStatus, log.Category, log.PreIndex))
	}else if log.StatusChange {
		//其他状态时的状态发生变化,要将老状态对应的记录删除，同时加入新状态记录；对于每个地址的下注记录也需要遍历处理。
		kvs = append(kvs, addGuessGameStatusIndexKey(log.Status, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameAdminStatusIndexKey(log.Status, log.AdminAddr, log.GameId, log.Index))
		kvs = append(kvs, addGuessGameCategoryStatusIndexKey(log.Status, log.Category, log.GameId, log.Index))

		kvs = append(kvs, delGuessGameStatusIndexKey(log.PreStatus, log.PreIndex))
		kvs = append(kvs, delGuessGameAdminStatusIndexKey(log.PreStatus, log.AdminAddr, log.PreIndex))
		kvs = append(kvs, delGuessGameCategoryStatusIndexKey(log.PreStatus, log.Category, log.PreIndex))

		
		//从game中遍历每个地址的记录进行新增和删除
		kvs = append(kvs, addGuessGameAddrStatusIndexKey(log.Status, log.Addr, log.GameId, log.Index))
		kvs = append(kvs, delGuessGameAddrStatusIndexKey(log.Status, log.Addr, log.Index))
	}


	return kvs
}

func (c *Guess) execLocal(receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return dbSet, nil
	}
	for i := 0; i < len(receipt.Logs); i++ {
		item := receipt.Logs[i]
		if item.Ty >= pkt.TyLogGuessGameStart && item.Ty <= pkt.TyLogGuessGameTimeout {
			var Gamelog pkt.ReceiptGuessGame
			err := types.Decode(item.Log, &Gamelog)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := c.updateIndex(&Gamelog)
			dbSet.KV = append(dbSet.KV, kv...)
		}
	}
	return dbSet, nil
}

func (c *Guess) ExecLocal_Start(payload *pkt.PBGameStart, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(receiptData)
}

func (c *Guess) ExecLocal_Continue(payload *pkt.PBGameContinue, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(receiptData)
}

func (c *Guess) ExecLocal_Quit(payload *pkt.PBGameQuit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(receiptData)
}
