// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

//newZkSyncAccount create new zkSync contract account
func newZkSyncAccount(cfg *types.Chain33Config, symbol string, db db.KV) (*account.DB, error) {
	return account.NewAccountDB(cfg, zt.Zksync, symbol, db)
}

func assetDepositBalance(acc *account.DB, addr string, amount int64) (*types.Receipt, error) {
	if !acc.CheckAmount(amount) {
		return nil, types.ErrAmount
	}
	acc1 := acc.LoadAccount(addr)
	copyacc := *acc1
	acc1.Balance += amount
	receiptBalance := &types.ReceiptAccountTransfer{
		Prev:    &copyacc,
		Current: acc1,
	}
	acc.SaveAccount(acc1)
	ty := int32(zt.TyLogContractAssetDeposit)
	log1 := &types.ReceiptLog{
		Ty:  ty,
		Log: types.Encode(receiptBalance),
	}
	kv := acc.GetKVSet(acc1)
	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kv,
		Logs: []*types.ReceiptLog{log1},
	}, nil
}

func assetWithdrawBalance(acc *account.DB, addr string, amount int64) (*types.Receipt, error) {
	if !acc.CheckAmount(amount) {
		return nil, types.ErrAmount
	}
	acc1 := acc.LoadAccount(addr)
	if acc1.Balance-amount < 0 {
		return nil, types.ErrNoBalance
	}
	copyacc := *acc1
	acc1.Balance -= amount
	receiptBalance := &types.ReceiptAccountTransfer{
		Prev:    &copyacc,
		Current: acc1,
	}
	acc.SaveAccount(acc1)
	ty := int32(zt.TyLogContractAssetWithdraw)
	log1 := &types.ReceiptLog{
		Ty:  ty,
		Log: types.Encode(receiptBalance),
	}
	kv := acc.GetKVSet(acc1)
	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kv,
		Logs: []*types.ReceiptLog{log1},
	}, nil
}
