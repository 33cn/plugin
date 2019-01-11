// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/pokerbull/types"
)

// Exec_Start 开始游戏交易执行
func (c *PokerBull) Exec_Start(payload *pkt.PBGameStart, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GameStart(payload)
}

// Exec_Continue 继续游戏交易执行
func (c *PokerBull) Exec_Continue(payload *pkt.PBGameContinue, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GameContinue(payload)
}

// Exec_Quit 退出游戏交易执行
func (c *PokerBull) Exec_Quit(payload *pkt.PBGameQuit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GameQuit(payload)
}

// Exec_Play 已匹配玩家直接开始游戏
func (c *PokerBull) Exec_Play(payload *pkt.PBGamePlay, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(c, tx, index)
	return action.GamePlay(payload)
}
