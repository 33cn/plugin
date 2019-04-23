// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

// ExecDelLocal_Deposit for rollback deposit
func (p *Pos33) ExecDelLocal_Deposit(act *pt.Pos33DepositAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return p.deposit(-int(act.W), tx)
}

// ExecDelLocal_Withdraw for rollback withdraw
func (p *Pos33) ExecDelLocal_Withdraw(act *pt.Pos33WithdrawAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return p.deposit(-int(act.W), tx)
}

// ExecDelLocal_Delegate for rollback delegate
func (p *Pos33) ExecDelLocal_Delegate(act *pt.Pos33DelegateAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

// ExecDelLocal_Reword for rollback reword
func (p *Pos33) ExecDelLocal_Reword(act *pt.Pos33RewordAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

// ExecDelLocal_Punish for rollback punish
func (p *Pos33) ExecDelLocal_Punish(act *pt.Pos33PunishAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}
