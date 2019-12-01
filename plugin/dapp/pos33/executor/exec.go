// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

// Exec_Genesis exec genesis
func (t *Pos33Ticket) Exec_Genesis(payload *ty.Pos33TicketGenesis, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload.Count <= 0 {
		return nil, ty.ErrPos33TicketCount
	}
	actiondb := NewAction(t, tx)
	return actiondb.GenesisInit(payload)
}

// Exec_Topen exec open
func (t *Pos33Ticket) Exec_Topen(payload *ty.Pos33TicketOpen, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload.Count <= 0 {
		tlog.Error("topen ", "value", payload)
		return nil, ty.ErrPos33TicketCount
	}
	actiondb := NewAction(t, tx)
	return actiondb.Pos33TicketOpen(payload)
}

// Exec_Tbind exec bind
func (t *Pos33Ticket) Exec_Tbind(payload *ty.Pos33TicketBind, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewAction(t, tx)
	return actiondb.Pos33TicketBind(payload)
}

// Exec_Tclose exec close
func (t *Pos33Ticket) Exec_Tclose(payload *ty.Pos33TicketClose, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewAction(t, tx)
	return actiondb.Pos33TicketClose(payload)
}

//Exec_Miner exec miner
func (t *Pos33Ticket) Exec_Pminer(payload *ty.Pos33Miner, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewAction(t, tx)
	//return actiondb.Pos33TicketMiner(payload, index)
	return actiondb.Pos33Miner(payload, index)
}
