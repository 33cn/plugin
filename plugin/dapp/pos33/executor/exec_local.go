// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

func (p *Pos33) deposit(w int, tx *types.Transaction) (*types.LocalDBSet, error) {
	allw := p.getAllWeight()
	var kvs []*types.KeyValue
	kvs = append(kvs, p.addWeight(tx.From(), w))
	kvs = append(kvs, p.setAllWeight(allw+w))
	return &types.LocalDBSet{KV: kvs}, nil
}

// ExecLocal_Deposit do local deposit
func (p *Pos33) ExecLocal_Deposit(act *pt.Pos33DepositAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return p.deposit(int(act.W), tx)
}

// ExecLocal_Withdraw do local withdraw
func (p *Pos33) ExecLocal_Withdraw(act *pt.Pos33WithdrawAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return p.deposit(-int(act.W), tx)
}

// ExecLocal_Delegate do local delegate
func (p *Pos33) ExecLocal_Delegate(act *pt.Pos33DelegateAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

// ExecLocal_Reword do local reword
func (p *Pos33) ExecLocal_Reword(act *pt.Pos33RewordAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

// ExecLocal_Punish do local punish
func (p *Pos33) ExecLocal_Punish(act *pt.Pos33PunishAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

// ExecLocal_Electe do local punish
func (p *Pos33) ExecLocal_Electe(act *pt.Pos33ElecteAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	plog.Info("ExecLocal_Electe", "from", tx.From)
	local, err := p.getElecteLocal(act.Height)
	if err != nil {
		if err == types.ErrNotFound {
			local = new(pt.Pos33ElecteLocal)
		} else {
			return nil, err
		}
	}

	local.Es = append(local.Es, act)
	kvs := []*types.KeyValue{&types.KeyValue{Key: []byte(fmt.Sprintf("%s%d", pt.KeyPos33ElectePrefix, act.Height)), Value: types.Encode(local)}}

	rs := pt.Sortition(local.Es)
	plog.Info("ExecLocal_Electe sortition:", "len(committee)", len(rs.Rands))
	key := fmt.Sprintf("%s%d", pt.KeyPos33CommitteePrefix, act.Height)
	comm := &pt.Pos33Committee{Rands: rs, Height: act.Height}
	kvs = append(kvs, &types.KeyValue{Key: []byte(key), Value: types.Encode(comm)})
	return &types.LocalDBSet{KV: kvs}, nil
}
