// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
)

//ExecLocal_Config asset withdraw local db process
func (m *Mix) ExecDelLocal_Config(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

//ExecLocal_Deposit asset withdraw local db process
func (m *Mix) ExecDelLocal_Deposit(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

//ExecLocal_Withdraw asset withdraw local db process
func (m *Mix) ExecDelLocal_Withdraw(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

// ExecLocal_Transfer asset transfer local db process
func (m *Mix) ExecDelLocal_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

//ExecLocal_Authorize asset withdraw local db process
func (m *Mix) ExecDelLocal_Authorize(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}
