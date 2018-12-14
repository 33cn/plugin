// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	//"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/lottery/types"
)

func (l *Lottery) execLocal(tx *types.Transaction, receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	for _, item := range receipt.Logs {
		switch item.Ty {
		case pty.TyLogLotteryCreate, pty.TyLogLotteryBuy, pty.TyLogLotteryDraw, pty.TyLogLotteryClose:
			var lotterylog pty.ReceiptLottery
			err := types.Decode(item.Log, &lotterylog)
			if err != nil {
				return nil, err
			}
			kv := l.saveLottery(&lotterylog)
			set.KV = append(set.KV, kv...)

			if item.Ty == pty.TyLogLotteryBuy {
				kv := l.saveLotteryBuy(&lotterylog)
				set.KV = append(set.KV, kv...)
			} else if item.Ty == pty.TyLogLotteryDraw {
				kv := l.saveLotteryDraw(&lotterylog)
				set.KV = append(set.KV, kv...)
				kv = l.updateLotteryBuy(&lotterylog, true)
				set.KV = append(set.KV, kv...)
				kv = l.saveLotteryGain(&lotterylog)
				set.KV = append(set.KV, kv...)
			}
		}
	}
	return set, nil
}

// ExecLocal_Create Action
func (l *Lottery) ExecLocal_Create(payload *pty.LotteryCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return l.execLocal(tx, receiptData)
}

// ExecLocal_Buy Action
func (l *Lottery) ExecLocal_Buy(payload *pty.LotteryBuy, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return l.execLocal(tx, receiptData)
}

// ExecLocal_Draw Action
func (l *Lottery) ExecLocal_Draw(payload *pty.LotteryDraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return l.execLocal(tx, receiptData)
}

// ExecLocal_Close Action
func (l *Lottery) ExecLocal_Close(payload *pty.LotteryClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return l.execLocal(tx, receiptData)
}
