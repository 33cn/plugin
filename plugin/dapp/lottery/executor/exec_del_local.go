// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	//"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/lottery/types"
)

func (l *Lottery) execDelLocal(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	for _, item := range receiptData.Logs {
		switch item.Ty {
		case pty.TyLogLotteryCreate, pty.TyLogLotteryBuy, pty.TyLogLotteryDraw, pty.TyLogLotteryClose:
			var lotterylog pty.ReceiptLottery
			err := types.Decode(item.Log, &lotterylog)
			if err != nil {
				return nil, err
			}
			kv := l.deleteLottery(&lotterylog)
			set.KV = append(set.KV, kv...)

			if item.Ty == pty.TyLogLotteryBuy {
				kv := l.deleteLotteryBuy(&lotterylog)
				set.KV = append(set.KV, kv...)
			} else if item.Ty == pty.TyLogLotteryDraw {
				kv := l.deleteLotteryDraw(&lotterylog)
				set.KV = append(set.KV, kv...)
				kv = l.updateLotteryBuy(&lotterylog, false)
				set.KV = append(set.KV, kv...)
				kv = l.deleteLotteryGain(&lotterylog)
				set.KV = append(set.KV, kv...)
			}
		}
	}
	return set, nil

}

// ExecDelLocal_Create Action
func (l *Lottery) ExecDelLocal_Create(payload *pty.LotteryCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return l.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Buy Action
func (l *Lottery) ExecDelLocal_Buy(payload *pty.LotteryBuy, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return l.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Draw Action
func (l *Lottery) ExecDelLocal_Draw(payload *pty.LotteryDraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return l.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Close Action
func (l *Lottery) ExecDelLocal_Close(payload *pty.LotteryClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return l.execDelLocal(tx, receiptData)
}
