// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	//"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/collateralize/types"
)

func (c *Collateralize) execLocal(tx *types.Transaction, receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	for _, item := range receipt.Logs {
		switch item.Ty {
		case pty.TyLogCollateralizeCreate, pty.TyLogCollateralizeBorrow, pty.TyLogCollateralizeRepay, pty.TyLogCollateralizeAppend,
		pty.TyLogCollateralizeFeed, pty.TyLogCollateralizeClose:
			var Collateralizelog pty.ReceiptCollateralize
			err := types.Decode(item.Log, &Collateralizelog)
			if err != nil {
				return nil, err
			}
			kv := c.saveCollateralize(&Collateralizelog)
			set.KV = append(set.KV, kv...)

			if item.Ty == pty.TyLogCollateralizeBorrow {
				kv := c.saveCollateralizeBorrow(&Collateralizelog)
				set.KV = append(set.KV, kv...)
			} else if item.Ty == pty.TyLogCollateralizeRepay {
				kv := c.saveCollateralizeRepay(&Collateralizelog)
				set.KV = append(set.KV, kv...)
			} else {
				//TODO
			}
		}
	}
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

// ExecLocal_Repay Action
func (c *Collateralize) ExecLocal_Append(payload *pty.CollateralizeAppend, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Feed Action
func (c *Collateralize) ExecLocal_Feed(payload *pty.CollateralizeFeed, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}

// ExecLocal_Close Action
func (c *Collateralize) ExecLocal_Close(payload *pty.CollateralizeClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return c.execLocal(tx, receiptData)
}
