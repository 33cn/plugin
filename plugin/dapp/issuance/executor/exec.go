// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/issuance/types"
)

// Exec_Create Action
func (c *Issuance) Exec_Create(payload *pty.IssuanceCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewIssuanceAction(c, tx, index)
	return actiondb.IssuanceCreate(payload)
}

// Exec_Debt Action
func (c *Issuance) Exec_Debt(payload *pty.IssuanceDebt, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewIssuanceAction(c, tx, index)
	return actiondb.IssuanceDebt(payload)
}

// Exec_Repay Action
func (c *Issuance) Exec_Repay(payload *pty.IssuanceRepay, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewIssuanceAction(c, tx, index)
	return actiondb.IssuanceRepay(payload)
}

// Exec_Feed Action
func (c *Issuance) Exec_Feed(payload *pty.IssuanceFeed, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewIssuanceAction(c, tx, index)
	return actiondb.IssuanceFeed(payload)
}

// Exec_Close Action
func (c *Issuance) Exec_Close(payload *pty.IssuanceClose, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewIssuanceAction(c, tx, index)
	return actiondb.IssuanceClose(payload)
}

// Exec_Manage Action
func (c *Issuance) Exec_Manage(payload *pty.IssuanceManage, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewIssuanceAction(c, tx, index)
	return actiondb.IssuanceManage(payload)
}
