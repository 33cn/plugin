// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/collateralize/types"
)

func (c *Collateralize) execDelLocal(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	for _, item := range receiptData.Logs {
		var collateralizeLog pty.ReceiptCollateralize
		err := types.Decode(item.Log, &collateralizeLog)
		if err != nil {
			return nil, err
		}

		switch item.Ty {
		case pty.TyLogCollateralizeCreate:
			kv := c.deleteCollateralizeStatus(&collateralizeLog)
			set.KV = append(set.KV, kv...)
			break
		case pty.TyLogCollateralizeBorrow:
			set.KV = append(set.KV, c.deleteCollateralizeRecordStatus(&collateralizeLog)...)
			set.KV = append(set.KV, c.deleteCollateralizeAddr(&collateralizeLog)...)
			break
		case pty.TyLogCollateralizeAppend: // append没有状态变化
			break
		case pty.TyLogCollateralizeRepay:
			set.KV = append(set.KV, c.deleteCollateralizeRecordStatus(&collateralizeLog)...)
			set.KV = append(set.KV, c.addCollateralizeAddr(&collateralizeLog)...)
			break
		case pty.TyLogCollateralizeFeed:
			set.KV = append(set.KV, c.deleteCollateralizeRecordStatus(&collateralizeLog)...)
			if collateralizeLog.RecordStatus == pty.CollateralizeUserStatusSystemLiquidate {
				set.KV = append(set.KV, c.addCollateralizeAddr(&collateralizeLog)...)
			}
			break
		case pty.TyLogCollateralizeClose:
			kv := c.addCollateralizeStatus(&collateralizeLog)
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

// ExecDelLocal_Manage Action
func (c *Collateralize) ExecDelLocal_Manage(payload *pty.CollateralizeManage, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}