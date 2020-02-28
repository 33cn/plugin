// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common/db/table"
	//"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/issuance/types"
)

func (c *Issuance) execLocal(tx *types.Transaction, receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	var IDtable, recordTable *table.Table

	cfg := c.GetAPI().GetConfig()
	if cfg.IsDappFork(c.GetHeight(), pty.IssuanceX, pty.ForkIssuanceTableUpdate) {
		recordTable = pty.NewRecordTable(c.GetLocalDB())
		IDtable = pty.NewIssuanceTable(c.GetLocalDB())
	}

	for _, item := range receipt.Logs {
		if item.Ty >= pty.TyLogIssuanceCreate && item.Ty <= pty.TyLogIssuanceClose {
			var issuanceLog pty.ReceiptIssuance
			err := types.Decode(item.Log, &issuanceLog)
			if err != nil {
				return nil, err
			}

			if item.Ty == pty.TyLogIssuanceCreate || item.Ty == pty.TyLogIssuanceClose {
				if !cfg.IsDappFork(c.GetHeight(), pty.IssuanceX, pty.ForkIssuanceTableUpdate) {
					IDtable = pty.NewIssuanceTable(c.GetLocalDB())
				}
				err = IDtable.Replace(&pty.ReceiptIssuanceID{IssuanceId: issuanceLog.IssuanceId, Status: issuanceLog.Status})
				if err != nil {
					return nil, err
				}
			} else {
				if !cfg.IsDappFork(c.GetHeight(), pty.IssuanceX, pty.ForkIssuanceTableUpdate) {
					recordTable = pty.NewRecordTable(c.GetLocalDB())
				}
				err = recordTable.Replace(&pty.ReceiptIssuance{IssuanceId: issuanceLog.IssuanceId, Status: issuanceLog.Status,
					DebtId: issuanceLog.DebtId, AccountAddr: issuanceLog.AccountAddr})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if IDtable != nil {
		kvs, err := IDtable.Save()
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

	set.KV = c.AddRollbackKV(tx, []byte(pty.IssuanceX), set.KV)
	return set, nil
}

// ExecLocal_Create Action
func (c *Issuance) ExecLocal_Create(payload *pty.IssuanceCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Debt Action
func (c *Issuance) ExecLocal_Debt(payload *pty.IssuanceDebt, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Repay Action
func (c *Issuance) ExecLocal_Repay(payload *pty.IssuanceRepay, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Feed Action
func (c *Issuance) ExecLocal_Feed(payload *pty.IssuanceFeed, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Close Action
func (c *Issuance) ExecLocal_Close(payload *pty.IssuanceClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Manage Action
func (c *Issuance) ExecLocal_Manage(payload *pty.IssuanceManage, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}
