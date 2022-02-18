// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/golang/protobuf/proto"

	"github.com/pkg/errors"
)

func makeNullifierSetReceipt(hash string, data proto.Message) *types.Receipt {
	return makeReceipt(calcNullifierHashKey(hash), mixTy.TyLogNulliferSet, data)

}

func (a *action) depositVerify(proof *mixTy.ZkProofInfo) (*mixTy.DepositCircuit, error) {
	var input mixTy.DepositCircuit

	err := mixTy.ConstructCircuitPubInput(proof.PublicInput, &input)
	if err != nil {
		return nil, errors.Wrapf(err, "setCircuitPubInput")
	}

	err = zkProofVerify(a.db, proof, mixTy.VerifyType_DEPOSIT)
	if err != nil {
		return nil, errors.Wrapf(err, "verify fail for input=%s", proof.PublicInput)
	}

	return &input, nil

}

/*
1. verify zk-proof
2. verify commit value vs value
3. deposit to mix contract
4. add new commits to merkle tree
*/
func (a *action) Deposit(deposit *mixTy.MixDepositAction) (*types.Receipt, error) {
	var notes []string
	var sum uint64
	//1. zk-proof校验
	for _, p := range deposit.Proofs {
		input, err := a.depositVerify(p)
		if err != nil {
			return nil, errors.Wrap(err, "get pub input")
		}
		v := input.Amount.GetWitnessValue(ecc.BN254)
		sum += v.Uint64()
		noteHash := input.NoteHash.GetWitnessValue(ecc.BN254)
		notes = append(notes, noteHash.String())
	}

	//存款
	cfg := a.api.GetConfig()
	execer, symbol := mixTy.GetAssetExecSymbol(cfg, deposit.AssetExec, deposit.AssetSymbol)
	accoutDb, err := createAccount(cfg, execer, symbol, a.db)
	if err != nil {
		return nil, errors.Wrapf(err, "createAccount,execer=%s,symbol=%s", execer, symbol)
	}
	//主链上存入toAddr为mix 执行器地址，平行链上为user.p.{}.mix执行器地址,execAddr和toAddr一致
	execAddr := address.ExecAddress(string(a.tx.Execer))
	receipt, err := accoutDb.ExecTransfer(a.fromaddr, execAddr, execAddr, int64(sum))
	if err != nil {
		return nil, errors.Wrapf(err, "account save to exec")
	}
	//push new commit to merkle tree
	var leaves [][]byte
	for _, n := range notes {
		leaves = append(leaves, mixTy.Str2Byte(n))
	}
	conf := types.ConfSub(cfg, mixTy.MixX)
	maxTreeLeaves := conf.GInt("maxTreeLeaves")
	rpt, err := pushTree(a.db, execer, symbol, leaves, int32(maxTreeLeaves))
	if err != nil {
		return nil, errors.Wrap(err, "pushTree")
	}
	mergeReceipt(receipt, rpt)

	return receipt, nil

}
