// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/collateralize/types"
)

// Exec_Create Action
func (c *Collateralize) Exec_Create(payload *pty.CollateralizeCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewCollateralizeAction(c, tx, index)
	return actiondb.CollateralizeCreate(payload)
}

// Exec_Borrow Action
func (c *Collateralize) Exec_Borrow(payload *pty.CollateralizeBorrow, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewCollateralizeAction(c, tx, index)
	return actiondb.CollateralizeBorrow(payload)
}

// Exec_Repay Action
func (c *Collateralize) Exec_Repay(payload *pty.CollateralizeRepay, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewCollateralizeAction(c, tx, index)
	return actiondb.CollateralizeRepay(payload)
}

// Exec_Append Action
func (c *Collateralize) Exec_Append(payload *pty.CollateralizeAppend, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewCollateralizeAction(c, tx, index)
	return actiondb.CollateralizeAppend(payload)
}

// Exec_Feed Action
func (c *Collateralize) Exec_Feed(payload *pty.CollateralizeFeed, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewCollateralizeAction(c, tx, index)
	return actiondb.CollateralizeFeed(payload)
}

// Exec_Retrieve Action
func (c *Collateralize) Exec_Retrieve(payload *pty.CollateralizeRetrieve, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewCollateralizeAction(c, tx, index)
	return actiondb.CollateralizeRetrieve(payload)
}

// Exec_Manage Action
func (c *Collateralize) Exec_Manage(payload *pty.CollateralizeManage, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewCollateralizeAction(c, tx, index)
	return actiondb.CollateralizeManage(payload)
}
