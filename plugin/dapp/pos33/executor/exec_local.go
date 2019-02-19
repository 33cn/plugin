// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

func (p *Pos33) deposit(w int, tx *types.Transaction) (*types.Receipt, error) {
	allw := p.getAllWeight()
	var kvs []*types.KeyValue
	kvs = append(kvs, p.addWeight(tx.From(), w))
	kvs = append(kvs, p.setAllWeight(allw+w))
	return &types.Receipt{Ty: types.ExecOk, KV: kvs}, nil
}

// ExecLocal_Deposit do local deposit
func (p *Pos33) ExecLocal_Deposit(act *pt.Pos33DepositAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	return p.deposit(int(act.W), tx)
}

// ExecLocal_Withdraw do local withdraw
func (p *Pos33) ExecLocal_Withdraw(act *pt.Pos33WithdrawAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	return p.deposit(-int(act.W), tx)
}

// ExecLocal_Delegate do local delegate
func (p *Pos33) ExecLocal_Delegate(act *pt.Pos33DelegateAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	return nil, nil
}

// ExecLocal_Reword do local reword
func (p *Pos33) ExecLocal_Reword(act *pt.Pos33RewordAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	return nil, nil
}

// ExecLocal_Punish do local punish
func (p *Pos33) ExecLocal_Punish(act *pt.Pos33PunishAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	return nil, nil
}
