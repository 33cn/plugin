// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

//Exec_Commit consensus commit tx exec process
func (m *Mix) Exec_Config(payload *mixTy.MixConfigAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(m, tx)
	receipt, err := a.Config(payload)
	if err != nil {
		mlog.Error("mix config failed", "error", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	return receipt, nil
}

//Exec_Deposit ...
func (m *Mix) Exec_Deposit(payload *mixTy.MixDepositAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(m, tx)
	receipt, err := a.Deposit(payload)
	if err != nil {
		mlog.Error("mix deposit failed", "error", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	return receipt, nil
}

//Exec_Withdraw ...
func (m *Mix) Exec_Withdraw(payload *mixTy.MixWithdrawAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(m, tx)
	receipt, err := a.Withdraw(payload)
	if err != nil {
		mlog.Error("mix withdraw failed", "error", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	return receipt, nil
}

func (m *Mix) Exec_Transfer(payload *mixTy.MixTransferAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(m, tx)
	receipt, err := a.Transfer(payload)
	if err != nil {
		mlog.Error("mix config failed", "error", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	return receipt, nil
}

func (m *Mix) Exec_Authorize(payload *mixTy.MixAuthorizeAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(m, tx)
	receipt, err := a.Authorize(payload)
	if err != nil {
		mlog.Error("mix config failed", "error", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	return receipt, nil
}
