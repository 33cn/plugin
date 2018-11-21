// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/lottery/types"
)

// Exec_Create Action
func (l *Lottery) Exec_Create(payload *pty.LotteryCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewLotteryAction(l, tx, index)
	return actiondb.LotteryCreate(payload)
}

// Exec_Buy Action
func (l *Lottery) Exec_Buy(payload *pty.LotteryBuy, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewLotteryAction(l, tx, index)
	return actiondb.LotteryBuy(payload)
}

// Exec_Draw Action
func (l *Lottery) Exec_Draw(payload *pty.LotteryDraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewLotteryAction(l, tx, index)
	return actiondb.LotteryDraw(payload)
}

// Exec_Close Action
func (l *Lottery) Exec_Close(payload *pty.LotteryClose, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewLotteryAction(l, tx, index)
	return actiondb.LotteryClose(payload)
}
