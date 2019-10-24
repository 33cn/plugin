// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	//"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/issuance/types"
)

func (c *Issuance) execLocal(tx *types.Transaction, receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	for _, item := range receipt.Logs {
		var issuanceLog pty.ReceiptIssuance
		err := types.Decode(item.Log, &issuanceLog)
		if err != nil {
			return nil, err
		}

		switch item.Ty {
		case pty.TyLogIssuanceCreate:
			set.KV = append(set.KV, c.addIssuanceStatus(issuanceLog.Status, issuanceLog.Index, issuanceLog.IssuanceId)...)
			break
		case pty.TyLogIssuanceDebt:
			set.KV = append(set.KV, c.addIssuanceRecordStatus(issuanceLog.RecordStatus, issuanceLog.AccountAddr, issuanceLog.Index,
				issuanceLog.DebtId, issuanceLog.IssuanceId)...)
			set.KV = append(set.KV, c.addIssuanceRecordAddr(issuanceLog.AccountAddr, issuanceLog.Index, issuanceLog.DebtId,
				issuanceLog.IssuanceId)...)
			break
		case pty.TyLogIssuanceRepay:
			set.KV = append(set.KV, c.deleteIssuanceRecordStatus(issuanceLog.RecordPreStatus, issuanceLog.PreIndex)...)
			set.KV = append(set.KV, c.addIssuanceRecordStatus(issuanceLog.RecordStatus, issuanceLog.AccountAddr, issuanceLog.Index,
				issuanceLog.DebtId, issuanceLog.IssuanceId)...)
			set.KV = append(set.KV, c.deleteIssuanceRecordAddr(issuanceLog.AccountAddr, issuanceLog.PreIndex)...)
			break
		case pty.TyLogIssuanceFeed:
			set.KV = append(set.KV, c.deleteIssuanceRecordStatus(issuanceLog.RecordPreStatus, issuanceLog.PreIndex)...)
			set.KV = append(set.KV, c.addIssuanceRecordStatus(issuanceLog.RecordStatus, issuanceLog.AccountAddr, issuanceLog.Index,
				issuanceLog.DebtId, issuanceLog.IssuanceId)...)
			set.KV = append(set.KV, c.deleteIssuanceRecordAddr(issuanceLog.AccountAddr, issuanceLog.PreIndex)...)
			// 如果没有被清算，需要把地址索引更新
			if issuanceLog.RecordStatus == pty.IssuanceUserStatusWarning || issuanceLog.RecordStatus == pty.IssuanceUserStatusExpire {
				set.KV = append(set.KV, c.addIssuanceRecordAddr(issuanceLog.AccountAddr, issuanceLog.Index, issuanceLog.DebtId,
					issuanceLog.IssuanceId)...)
			}
			break
		case pty.TyLogIssuanceClose:
			set.KV = append(set.KV, c.deleteIssuanceStatus(issuanceLog.Status, issuanceLog.PreIndex)...)
			break
		}
	}
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
