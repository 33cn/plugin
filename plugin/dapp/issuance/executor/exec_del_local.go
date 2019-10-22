// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/issuance/types"
)

func (c *Issuance) execDelLocal(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	for _, item := range receiptData.Logs {
		var IssuanceLog pty.ReceiptIssuance
		err := types.Decode(item.Log, &IssuanceLog)
		if err != nil {
			return nil, err
		}

		switch item.Ty {
		case pty.TyLogIssuanceCreate:
			kv := c.deleteIssuanceStatus(&IssuanceLog)
			set.KV = append(set.KV, kv...)
			break
		case pty.TyLogIssuanceDebt:
			set.KV = append(set.KV, c.deleteIssuanceRecordStatus(&IssuanceLog)...)
			set.KV = append(set.KV, c.deleteIssuanceAddr(&IssuanceLog)...)
			break
		case pty.TyLogIssuanceRepay:
			set.KV = append(set.KV, c.deleteIssuanceRecordStatus(&IssuanceLog)...)
			set.KV = append(set.KV, c.addIssuanceAddr(&IssuanceLog)...)
			break
		case pty.TyLogIssuanceFeed:
			set.KV = append(set.KV, c.deleteIssuanceRecordStatus(&IssuanceLog)...)
			if IssuanceLog.RecordStatus == pty.IssuanceUserStatusSystemLiquidate {
				set.KV = append(set.KV, c.addIssuanceAddr(&IssuanceLog)...)
			}
			break
		case pty.TyLogIssuanceClose:
			kv := c.addIssuanceStatus(&IssuanceLog)
			set.KV = append(set.KV, kv...)
			break
		}
	}
	return set, nil

}

// ExecDelLocal_Create Action
func (c *Issuance) ExecDelLocal_Create(payload *pty.IssuanceCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Debt Action
func (c *Issuance) ExecDelLocal_Debt(payload *pty.IssuanceDebt, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Repay Action
func (c *Issuance) ExecDelLocal_Repay(payload *pty.IssuanceRepay, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Feed Action
func (c *Issuance) ExecDelLocal_Feed(payload *pty.IssuanceFeed, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}

// ExecDelLocal_Close Action
func (c *Issuance) ExecDelLocal_Close(payload *pty.IssuanceClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execDelLocal(tx, receiptData)
}