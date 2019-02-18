// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

func (t *trade) ExecDelLocal_SellLimit(sell *pty.TradeForSell, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.localDelLog(tx, receipt, index, 0)
}

func (t *trade) ExecDelLocal_BuyMarket(buy *pty.TradeForBuy, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.localDelLog(tx, receipt, index, buy.BoardlotCnt)
}

func (t *trade) ExecDelLocal_RevokeSell(revoke *pty.TradeForRevokeSell, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.localDelLog(tx, receipt, index, 0)
}

func (t *trade) ExecDelLocal_BuyLimit(buy *pty.TradeForBuyLimit, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.localDelLog(tx, receipt, index, 0)
}

func (t *trade) ExecDelLocal_SellMarket(sell *pty.TradeForSellMarket, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.localDelLog(tx, receipt, index, sell.BoardlotCnt)
}

func (t *trade) ExecDelLocal_RevokeBuy(revoke *pty.TradeForRevokeBuy, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.localDelLog(tx, receipt, index, 0)
}

func (t *trade) localDelLog(tx *types.Transaction, receipt *types.ReceiptData, index int, tradedBoardlot int64) (*types.LocalDBSet, error) {
	var set types.LocalDBSet
	table := NewOrderTable(t.GetLocalDB())
	txIndex := dapp.HeightIndexStr(t.GetHeight(), int64(index))

	for i := 0; i < len(receipt.Logs); i++ {
		item := receipt.Logs[i]
		if item.Ty == pty.TyLogTradeSellLimit {
			var receipt pty.ReceiptTradeSellLimit
			err := types.Decode(item.Log, &receipt)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := t.deleteSell(receipt.Base, item.Ty, tx, txIndex, table, tradedBoardlot)
			set.KV = append(set.KV, kv...)
		} else if item.Ty == pty.TyLogTradeSellRevoke {
			var receipt pty.ReceiptTradeSellRevoke
			err := types.Decode(item.Log, &receipt)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := t.deleteSell(receipt.Base, item.Ty, tx, txIndex, table, tradedBoardlot)
			set.KV = append(set.KV, kv...)
		} else if item.Ty == pty.TyLogTradeBuyMarket {
			var receipt pty.ReceiptTradeBuyMarket
			err := types.Decode(item.Log, &receipt)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := t.deleteBuy(receipt.Base, txIndex, table)
			set.KV = append(set.KV, kv...)
		} else if item.Ty == pty.TyLogTradeBuyRevoke {
			var receipt pty.ReceiptTradeBuyRevoke
			err := types.Decode(item.Log, &receipt)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := t.deleteBuyLimit(receipt.Base, item.Ty, tx, txIndex, table, tradedBoardlot)
			set.KV = append(set.KV, kv...)
		} else if item.Ty == pty.TyLogTradeBuyLimit {
			var receipt pty.ReceiptTradeBuyLimit
			err := types.Decode(item.Log, &receipt)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := t.deleteBuyLimit(receipt.Base, item.Ty, tx, txIndex, table, tradedBoardlot)
			set.KV = append(set.KV, kv...)
		} else if item.Ty == pty.TyLogTradeSellMarket {
			var receipt pty.ReceiptSellMarket
			err := types.Decode(item.Log, &receipt)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := t.deleteSellMarket(receipt.Base, txIndex, table)
			set.KV = append(set.KV, kv...)
		}
	}
	newKvs, err := table.Save()
	if err != nil {
		tradelog.Error("trade table.Save failed", "error", err)
		return nil, err
	}
	set.KV = append(set.KV, newKvs...)
	for _, kv := range set.KV {
		t.GetLocalDB().Set(kv.Key, kv.Value)
	}

	return &set, nil
}
