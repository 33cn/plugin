// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

//ExecLocal_Config asset withdraw local db process
func (m *Mix) ExecDelLocal_Config(payload *mixTy.MixConfigAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

//ExecLocal_Deposit asset withdraw local db process
func (m *Mix) ExecDelLocal_Deposit(payload *mixTy.MixDepositAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return m.execAutoDelLocal(tx, receiptData)
}

//ExecLocal_Withdraw asset withdraw local db process
func (m *Mix) ExecDelLocal_Withdraw(payload *mixTy.MixWithdrawAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return m.execAutoDelLocal(tx, receiptData)
}

// ExecLocal_Transfer asset transfer local db process
func (m *Mix) ExecDelLocal_Transfer(payload *mixTy.MixTransferAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return m.execAutoDelLocal(tx, receiptData)
}

//ExecLocal_Authorize asset withdraw local db process
func (m *Mix) ExecDelLocal_Authorize(payload *mixTy.MixAuthorizeAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return m.execAutoDelLocal(tx, receiptData)
}
