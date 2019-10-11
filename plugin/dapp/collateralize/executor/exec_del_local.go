// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	//"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/collateralize/types"
)

func (c *Collateralize) execDelLocal(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	for _, item := range receiptData.Logs {
		var collateralizelog pty.ReceiptCollateralize
		err := types.Decode(item.Log, &collateralizelog)
		if err != nil {
			return nil, err
		}

		switch item.Ty {
		case pty.TyLogCollateralizeCreate:
			kv := c.deleteCollateralizeStatus(&collateralizelog)
			set.KV = append(set.KV, kv...)
			break
		case pty.TyLogCollateralizeBorrow:
			kv := c.deleteCollateralizeAddr(&collateralizelog)
			set.KV = append(set.KV, kv...)
			break
		case pty.TyLogCollateralizeAppend: // append没有状态变化
			break
		case pty.TyLogCollateralizeRepay:
			kv := c.addCollateralizeAddr(&collateralizelog)
			set.KV = append(set.KV, kv...)
			break
		/*case pty.TyLogCollateralizeFeed:*/ // TODO
		case pty.TyLogCollateralizeClose:
			kv := c.addCollateralizeStatus(&collateralizelog)
			set.KV = append(set.KV, kv...)
			break
		}
	}
	return set, nil

}

// ExecDelLocal_Create Action
func (c *Collateralize) ExecDelLocal_Create(payload *pty.CollateralizeCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Borrow Action
func (c *Collateralize) ExecDelLocal_Borrow(payload *pty.CollateralizeBorrow, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Repay Action
func (c *Collateralize) ExecDelLocal_Repay(payload *pty.CollateralizeRepay, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Append Action
func (c *Collateralize) ExecDelLocal_Append(payload *pty.CollateralizeAppend, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Feed Action
func (c *Collateralize) ExecDelLocal_Feed(payload *pty.CollateralizeFeed, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Close Action
func (c *Collateralize) ExecDelLocal_Close(payload *pty.CollateralizeClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}