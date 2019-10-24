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
			set.KV = append(set.KV, c.deleteCollateralizeStatus(collateralizeLog.Status, collateralizeLog.Index)...)
			set.KV = append(set.KV, c.deleteCollateralizeAddr(collateralizeLog.CreateAddr, collateralizeLog.Index)...)
			break
		case pty.TyLogCollateralizeBorrow:
			set.KV = append(set.KV, c.deleteCollateralizeRecordStatus(collateralizeLog.RecordStatus, collateralizeLog.Index)...)
			set.KV = append(set.KV, c.deleteCollateralizeRecordAddr(collateralizeLog.AccountAddr, collateralizeLog.Index)...)
			break
		case pty.TyLogCollateralizeAppend: // append没有状态变化
			break
		case pty.TyLogCollateralizeRepay:
			set.KV = append(set.KV, c.addCollateralizeRecordStatus(collateralizeLog.RecordPreStatus, collateralizeLog.CollateralizeId,
				collateralizeLog.RecordId, collateralizeLog.PreIndex)...)
			set.KV = append(set.KV, c.deleteCollateralizeRecordStatus(collateralizeLog.RecordStatus, collateralizeLog.Index)...)
			set.KV = append(set.KV, c.addCollateralizeRecordAddr(collateralizeLog.AccountAddr, collateralizeLog.CollateralizeId,
				collateralizeLog.RecordId, collateralizeLog.PreIndex)...)
			break
		case pty.TyLogCollateralizeFeed:
			set.KV = append(set.KV, c.addCollateralizeRecordStatus(collateralizeLog.RecordStatus, collateralizeLog.CollateralizeId,
				collateralizeLog.RecordId, collateralizeLog.PreIndex)...)
			set.KV = append(set.KV, c.deleteCollateralizeRecordStatus(collateralizeLog.RecordStatus, collateralizeLog.Index)...)
			set.KV = append(set.KV, c.addCollateralizeRecordAddr(collateralizeLog.AccountAddr, collateralizeLog.CollateralizeId,
				collateralizeLog.RecordId, collateralizeLog.PreIndex)...)
			// 如果没有被清算，需要把地址索引更新
			if collateralizeLog.RecordStatus == pty.CollateralizeUserStatusWarning || collateralizeLog.RecordStatus == pty.CollateralizeUserStatusExpire {
				set.KV = append(set.KV, c.deleteCollateralizeRecordAddr(collateralizeLog.AccountAddr, collateralizeLog.Index)...)
			}
			break
		case pty.TyLogCollateralizeClose:
			set.KV = append(set.KV, c.addCollateralizeStatus(pty.CollateralizeStatusCreated, collateralizeLog.CollateralizeId,
				collateralizeLog.PreIndex)...)
			set.KV = append(set.KV, c.addCollateralizeAddr(collateralizeLog.CreateAddr, collateralizeLog.CollateralizeId,
				collateralizeLog.PreIndex)...)
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