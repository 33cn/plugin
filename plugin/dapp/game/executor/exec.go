// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/game/types"
)

func (g *Game) Exec_Create(payload *gt.GameCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(g, tx, index)
	return action.GameCreate(payload)
}

func (g *Game) Exec_Cancel(payload *gt.GameCancel, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(g, tx, index)
	return action.GameCancel(payload)
}

func (g *Game) Exec_Close(payload *gt.GameClose, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(g, tx, index)
	return action.GameClose(payload)
}

func (g *Game) Exec_Match(payload *gt.GameMatch, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(g, tx, index)
	return action.GameMatch(payload)
}
