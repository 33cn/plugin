// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"math/rand"
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
)

// NewDepositTx create a deposit tx
func NewDepositTx(w int) (*types.Transaction, error) {
	act := &Pos33Action{
		Value: &Pos33Action_Deposit{Deposit: &Pos33DepositAction{W: int64(w)}},
		Ty:    Pos33ActionDeposit,
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(Pos33X)),
		Payload: types.Encode(act),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(types.ExecName(Pos33X)),
	}
	return tx, nil
}

// NewWithdrawTx create a deposit tx
func NewWithdrawTx(w int) (*types.Transaction, error) {
	act := &Pos33Action{
		Value: &Pos33Action_Withdraw{Withdraw: &Pos33WithdrawAction{W: int64(w)}},
		Ty:    Pos33ActionWithdraw,
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(Pos33X)),
		Payload: types.Encode(act),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(types.ExecName(Pos33X)),
	}
	return tx, nil
}

// NewRewordTx create a deposit tx
func NewRewordTx(votes []*Pos33Vote, randHash []byte) (*types.Transaction, error) {
	act := &Pos33Action{
		Value: &Pos33Action_Reword{Reword: &Pos33RewordAction{Votes: votes, RandHash: randHash}},
		Ty:    Pos33ActionReword,
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(Pos33X)),
		Payload: types.Encode(act),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(types.ExecName(Pos33X)),
	}
	return tx, nil
}

// NewPunishTx create a deposit tx
func NewPunishTx(w int) (*types.Transaction, error) {
	return nil, nil
}

// NewDelegateTx create a deposit tx
func NewDelegateTx(w int) (*types.Transaction, error) {
	return nil, nil
}

// NewElecteTx create a deposit tx
func NewElecteTx(rands []*Pos33Rands, blockHash []byte, blochHeight int64) (*types.Transaction, error) {
	act := &Pos33Action{
		Value: &Pos33Action_Electe{Electe: &Pos33ElecteAction{Rands: rands, Hash: blockHash, Height: blochHeight}},
		Ty:    Pos33ActionElecte,
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(Pos33X)),
		Payload: types.Encode(act),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(types.ExecName(Pos33X)),
		Fee:     1e7,
	}
	return tx, nil
}
