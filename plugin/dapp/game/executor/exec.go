// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/game/types"
)

// Exec_Create Create game
func (g *Game) Exec_Create(payload *gt.GameCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(g, tx, index)
	return action.GameCreate(payload)
}

// Exec_Cancel Cancel game
func (g *Game) Exec_Cancel(payload *gt.GameCancel, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(g, tx, index)
	return action.GameCancel(payload)
}

// Exec_Close Close game
func (g *Game) Exec_Close(payload *gt.GameClose, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(g, tx, index)
	return action.GameClose(payload)
}

// Exec_Match Match game
func (g *Game) Exec_Match(payload *gt.GameMatch, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(g, tx, index)
	return action.GameMatch(payload)
}
