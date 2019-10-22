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
		var IssuanceLog pty.ReceiptIssuance
		err := types.Decode(item.Log, &IssuanceLog)
		if err != nil {
			return nil, err
		}

		switch item.Ty {
		case pty.TyLogIssuanceCreate:
			set.KV = append(set.KV, c.addIssuanceStatus(&IssuanceLog)...)
			break
		case pty.TyLogIssuanceDebt:
			set.KV = append(set.KV, c.addIssuanceRecordStatus(&IssuanceLog)...)
			set.KV = append(set.KV, c.addIssuanceAddr(&IssuanceLog)...)
			break
		case pty.TyLogIssuanceRepay:
			set.KV = append(set.KV, c.addIssuanceRecordStatus(&IssuanceLog)...)
			set.KV = append(set.KV, c.deleteIssuanceAddr(&IssuanceLog)...)
			break
		case pty.TyLogIssuanceFeed:
			set.KV = append(set.KV, c.addIssuanceRecordStatus(&IssuanceLog)...)
			if IssuanceLog.RecordStatus == pty.IssuanceUserStatusSystemLiquidate {
				set.KV = append(set.KV, c.deleteIssuanceAddr(&IssuanceLog)...)
			}
			break
		case pty.TyLogIssuanceClose:
			set.KV = append(set.KV, c.deleteIssuanceStatus(&IssuanceLog)...)
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
