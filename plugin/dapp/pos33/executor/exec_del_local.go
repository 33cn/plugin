// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

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

// ExecDelLocal_Electe for rollback punish
func (p *Pos33) ExecDelLocal_Electe(act *pt.Pos33ElecteAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	local, err := p.getElecteLocal(act.Height)
	if err != nil {
		return nil, err
	}

	local.Es = local.Es[:len(local.Es)-1]
	kvs := []*types.KeyValue{&types.KeyValue{Key: []byte(fmt.Sprintf("%s%d", keyPos33ElectePrefix, act.Height)), Value: types.Encode(local)}}

	rs := sortition(local.Es)
	comm, err := p.getCommittee(keyPos33Committee)
	if err != nil {
		return nil, err
	}
	comm.Rands = rs
	kvs = append(kvs, &types.KeyValue{Key: []byte(keyPos33Committee), Value: types.Encode(comm)})
	return &types.LocalDBSet{KV: kvs}, nil
}

// roll back local db data
// func (p *Pos33) execDelLocal(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
// 	dbSet := &types.LocalDBSet{}
// 	for _, log := range receiptData.Logs {
// 		switch log.GetTy() {
// 		case pt.TyLogDeposit, pt.TyLogWithdraw, pt.TyLogDelegate, pt.TyLogReword, pt.TyLogPunish:
// 			receipt := &pt.ReceiptPos33{}
// 			if err := types.Decode(log.Log, receipt); err != nil {
// 				return nil, err
// 			}
// 			kv := p.rollbackIndex(receipt)
// 			dbSet.KV = append(dbSet.KV, kv...)
// 		}
// 	}
// 	return dbSet, nil
// }
