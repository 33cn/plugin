// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common/db/table"
	//"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/collateralize/types"
)

func (c *Collateralize) execLocal(tx *types.Transaction, receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	var collTable, recordTable *table.Table

	cfg := c.GetAPI().GetConfig()
	if cfg.IsDappFork(c.GetHeight(), pty.CollateralizeX, pty.ForkCollateralizeTableUpdate) {
		recordTable = pty.NewRecordTable(c.GetLocalDB())
		collTable = pty.NewCollateralizeTable(c.GetLocalDB())
	}

	for _, item := range receipt.Logs {
		if item.Ty >= pty.TyLogCollateralizeCreate && item.Ty <= pty.TyLogCollateralizeRetrieve {
			var collateralizeLog pty.ReceiptCollateralize
			err := types.Decode(item.Log, &collateralizeLog)
			if err != nil {
				return nil, err
			}

			if item.Ty == pty.TyLogCollateralizeCreate || item.Ty == pty.TyLogCollateralizeRetrieve {
				if !cfg.IsDappFork(c.GetHeight(), pty.CollateralizeX, pty.ForkCollateralizeTableUpdate) {
					collTable = pty.NewCollateralizeTable(c.GetLocalDB())
				}
				err = collTable.Replace(&pty.ReceiptCollateralize{CollateralizeId: collateralizeLog.CollateralizeId, Status: collateralizeLog.Status,
					AccountAddr: collateralizeLog.AccountAddr})
				if err != nil {
					return nil, err
				}
			} else {
				if !cfg.IsDappFork(c.GetHeight(), pty.CollateralizeX, pty.ForkCollateralizeTableUpdate) {
					recordTable = pty.NewRecordTable(c.GetLocalDB())
				}
				err = recordTable.Replace(&pty.ReceiptCollateralize{CollateralizeId: collateralizeLog.CollateralizeId, Status: collateralizeLog.Status,
					RecordId: collateralizeLog.RecordId, AccountAddr: collateralizeLog.AccountAddr})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if collTable != nil {
		kvs, err := collTable.Save()
		if err != nil {
			return nil, err
		}
		set.KV = append(set.KV, kvs...)
	}

	if recordTable != nil {
		kvs, err := recordTable.Save()
		if err != nil {
			return nil, err
		}
		set.KV = append(set.KV, kvs...)
	}

	set.KV = c.AddRollbackKV(tx, []byte(pty.CollateralizeX), set.KV)
	return set, nil
}

// ExecLocal_Create Action
func (c *Collateralize) ExecLocal_Create(payload *pty.CollateralizeCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Borrow Action
func (c *Collateralize) ExecLocal_Borrow(payload *pty.CollateralizeBorrow, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Repay Action
func (c *Collateralize) ExecLocal_Repay(payload *pty.CollateralizeRepay, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Append Action
func (c *Collateralize) ExecLocal_Append(payload *pty.CollateralizeAppend, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Feed Action
func (c *Collateralize) ExecLocal_Feed(payload *pty.CollateralizeFeed, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Retrieve Action
func (c *Collateralize) ExecLocal_Retrieve(payload *pty.CollateralizeRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Manage Action
func (c *Collateralize) ExecLocal_Manage(payload *pty.CollateralizeManage, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}
