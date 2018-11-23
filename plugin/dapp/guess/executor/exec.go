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

func (c *Guess) Exec_Continue(payload *pkt.PBGameContinue, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GameContinue(payload)
}

func (c *Guess) Exec_Quit(payload *pkt.PBGameQuit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GameQuit(payload)
}