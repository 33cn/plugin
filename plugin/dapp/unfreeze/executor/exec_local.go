// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	uf "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

func (u *Unfreeze) execLocal(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.GetTy() != types.ExecOk {
		return dbSet, nil
	}
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case uf.TyLogCreateUnfreeze, uf.TyLogWithdrawUnfreeze, uf.TyLogTerminateUnfreeze:
			var receipt uf.ReceiptUnfreeze
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				return nil, err
			}
			kv := u.saveUnfreezeCreate(&receipt)
			dbSet.KV = append(dbSet.KV, kv...)
		default:
			return nil, types.ErrNotSupport
		}
	}
	return dbSet, nil
}

func (u *Unfreeze) ExecLocal_Create(payload *uf.UnfreezeCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execLocal(receiptData)
}

func (u *Unfreeze) ExecLocal_Withdraw(payload *uf.UnfreezeWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execLocal(receiptData)
}

func (u *Unfreeze) ExecLocal_Terminate(payload *uf.UnfreezeTerminate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execLocal(receiptData)
}

func localKeys(res *uf.ReceiptUnfreeze, value []byte) (kvs []*types.KeyValue) {
	kvs = append(kvs, &types.KeyValue{Key: initKey(res.Current.Initiator), Value: value})
	kvs = append(kvs, &types.KeyValue{Key: beneficiaryKey(res.Current.Beneficiary), Value: value})
	return
}

func (u *Unfreeze) saveUnfreezeCreate(res *uf.ReceiptUnfreeze) (kvs []*types.KeyValue) {
	kvs = localKeys(res, []byte(res.Current.UnfreezeID))
	return
}

func (u *Unfreeze) rollbackUnfreezeCreate(res *uf.ReceiptUnfreeze) (kvs []*types.KeyValue) {
	kvs = localKeys(res, nil)
	return
}
