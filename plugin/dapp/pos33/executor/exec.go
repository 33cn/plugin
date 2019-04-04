// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"errors"
	"fmt"

	"github.com/33cn/chain33/common/address"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

// Exec_Deposit do deposit action
func (p *Pos33) Exec_Deposit(act *pt.Pos33DepositAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	r, err := p.GetCoinsAccount().ExecDepositFrozen(tx.From(), drivers.ExecAddress(p.GetDriverName()), pt.Pos33Miner*act.W)
	if err != nil {
		panic(err)
	}
	return r, err
}

// Exec_Withdraw do withdraw action
func (p *Pos33) Exec_Withdraw(act *pt.Pos33WithdrawAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	return p.GetCoinsAccount().ExecActive(tx.From(), drivers.ExecAddress(p.GetDriverName()), pt.Pos33Miner*act.W)
}

// Exec_Delegate do delegate action
func (p *Pos33) Exec_Delegate(act *pt.Pos33DelegateAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	return nil, nil
}

// Exec_Reword do reword action
func (p *Pos33) Exec_Reword(act *pt.Pos33RewordAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	sumw := 0
	for i, v := range act.Votes {
		w := int(v.Weight)
		sumw += w
		if sumw > int(pt.Pos33BlockReword/types.Coin) {
			act.Votes = act.Votes[:i]
			sumw -= w
			break
		}
	}

	const vr = pt.Pos33VoteReword
	bpReword := vr * int64(sumw)
	bp := address.PubKeyToAddress(tx.Signature.Pubkey).String()

	db := p.GetCoinsAccount()
	bpAcc := db.LoadAccount(bp)
	bpAcc.Balance += int64(bpReword)

	var kvs []*types.KeyValue
	for _, v := range act.Votes {
		addr := address.PubKeyToAddress(v.Sig.Pubkey).String()
		acc := db.LoadAccount(addr)
		acc.Balance += vr * int64(v.Weight)
		kvs = append(kvs, db.GetKVSet(acc)...)
		plog.Info("block reword", "voter", addr, "voter reword", vr*int64(v.Weight))
	}
	facc := db.LoadAccount(pt.Pos33FundKeyAddr)
	fr := pt.Pos33BlockReword - types.Coin*int64(sumw)
	facc.Balance += fr
	kvs = append(kvs, db.GetKVSet(facc)...)

	plog.Info("block reword", "bp", bp, "bp-reword", bpReword, "fund-reword", fr)
	return &types.Receipt{Ty: types.ExecOk, KV: kvs}, nil
}

// Exec_Punish do punish action
func (p *Pos33) Exec_Punish(act *pt.Pos33PunishAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	db := p.GetCoinsAccount()
	var kvs []*types.KeyValue
	for who := range act.Punishs {
		frozen := db.LoadAccount(who).Frozen
		_, err := db.ExecActive(who, drivers.ExecAddress(p.GetDriverName()), frozen)
		if err != nil {
			return nil, err
		}
		acc := db.LoadAccount(who)
		acc.Balance -= frozen
		kvs = append(kvs, db.GetKVSet(acc)...)
	}

	return &types.Receipt{Ty: types.ExecOk, KV: kvs}, nil
}

// Exec_Electe do electe action
func (p *Pos33) Exec_Electe(act *pt.Pos33ElecteAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	pub := string(tx.Signature.Pubkey)
	allw := p.getAllWeight()
	addr := address.PubKeyToAddress([]byte(pub)).String()
	plog.Info("Exec_Electe", "addr", addr)
	w := p.getWeight(addr)
	if w < 1 {
		return nil, fmt.Errorf("%s is NOT deposit ycc", addr)
	}

	if p.GetHeight() <= act.Height || p.GetHeight() > act.Height+pt.Pos33CommitteeSize || act.Height%pt.Pos33CommitteeSize != 0 {
		return nil, errors.New("act.Hieght error")
	}

	// TODO: should check act.Hash == Block(act.height).Hash
	// seed, err := p.getBlockSeed(act.Height)
	// if err != nil {
	// 	return nil, err
	// }
	// if string(seed) != string(act.Hash) {
	// 	return nil, fmt.Errorf("block seed error")
	// }
	seed := act.Hash
	err := pt.CheckRands(addr, allw, w, act.Rands, act.Height, seed, act.Sig)
	if err != nil {
		return nil, err
	}
	key := []byte("mavl-" + pt.Pos33X + "-electe")
	value := tx.Hash()
	kvs := []*types.KeyValue{&types.KeyValue{Key: key, Value: value}}
	return &types.Receipt{Ty: types.ExecOk, KV: kvs}, nil
}
