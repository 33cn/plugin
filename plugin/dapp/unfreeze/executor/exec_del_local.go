// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	uf "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

func (u *Unfreeze) execDelLocal(receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
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
			err = update(table, receipt.Prev)
			if err != nil {
				return nil, err
			}
		case uf.TyLogCreateUnfreeze:
			err := table.Del([]byte(txIndex))
			if err != nil {
				return nil, err
			}
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

// ExecDelLocal_Create 本地撤销执行创建冻结合约
func (u *Unfreeze) ExecDelLocal_Create(payload *uf.UnfreezeCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execDelLocal(receiptData, index)
}

// ExecDelLocal_Withdraw 本地撤销执行冻结合约中提币
func (u *Unfreeze) ExecDelLocal_Withdraw(payload *uf.UnfreezeWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execDelLocal(receiptData, index)
}

// ExecDelLocal_Terminate 本地撤销执行冻结合约的终止
func (u *Unfreeze) ExecDelLocal_Terminate(payload *uf.UnfreezeTerminate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execDelLocal(receiptData, index)
}
