// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/guess/types"
)

func (c *Guess) Exec_Start(payload *pkt.GuessGameStart, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GameStart(payload)
}

func (c *Guess) Exec_Bet(payload *pkt.GuessGameBet, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GameBet(payload)
}

func (c *Guess) Exec_StopBet(payload *pkt.GuessGameStopBet, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GameStopBet(payload)
}

func (c *Guess) Exec_Publish(payload *pkt.GuessGamePublish, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GamePublish(payload)
}

func (c *Guess) Exec_Abort(payload *pkt.GuessGameAbort, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GameAbort(payload)
}
