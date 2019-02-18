// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	uf "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

func (u *Unfreeze) execLocal(receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.GetTy() != types.ExecOk {
		return dbSet, nil
	}
	table := NewAddrTable(u.GetLocalDB())
	txIndex := dapp.HeightIndexStr(u.GetHeight(), int64(index))

	for _, log := range receiptData.Logs {
		switch log.Ty {
		case uf.TyLogWithdrawUnfreeze, uf.TyLogTerminateUnfreeze:
			var receipt uf.ReceiptUnfreeze
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				return nil, err
			}
			err = update(table, receipt.Current)
			if err != nil {
				return nil, err
			}
		case uf.TyLogCreateUnfreeze:
			var receipt uf.ReceiptUnfreeze
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				return nil, err
			}
			u := uf.LocalUnfreeze{
				Unfreeze: receipt.Current,
				TxIndex:  txIndex,
			}
			err = table.Add(&u)
			if err != nil {
				return nil, err
			}
		default:
		}
	}
	kv, err := table.Save()
	if err != nil {
		return nil, err
	}
	dbSet.KV = append(dbSet.KV, kv...)
	for _, kv := range dbSet.KV {
		u.GetLocalDB().Set(kv.Key, kv.Value)
	}
	return dbSet, nil
}

// ExecLocal_Create 本地执行创建冻结合约
func (u *Unfreeze) ExecLocal_Create(payload *uf.UnfreezeCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execLocal(receiptData, index)
}

// ExecLocal_Withdraw 本地执行提币
func (u *Unfreeze) ExecLocal_Withdraw(payload *uf.UnfreezeWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execLocal(receiptData, index)
}

// ExecLocal_Terminate 本地执行终止冻结合约
func (u *Unfreeze) ExecLocal_Terminate(payload *uf.UnfreezeTerminate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execLocal(receiptData, index)
}
