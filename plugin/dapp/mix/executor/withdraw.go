// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/pkg/errors"
)

func spendVerify(db dbm.KV, exec, symbol string, treeRootHash, nulliferHash, authorizeSpendHash string) error {
	//zk-proof校验
	//check tree rootHash exist
	exist, err := checkTreeRootHashExist(db, exec, symbol, mixTy.Str2Byte(treeRootHash))
	if err != nil {
		return errors.Wrapf(err, "roothash=%s not found", treeRootHash)
	}
	if !exist {
		return errors.Wrapf(mixTy.ErrTreeRootHashNotFound, "roothash=%s", treeRootHash)
	}

	//nullifier should not exist
	nullifierKey := calcNullifierHashKey(nulliferHash)
	_, err = db.Get(nullifierKey)
	if err == nil {
		return errors.Wrapf(mixTy.ErrNulliferHashExist, "nullifier=%s", nulliferHash)
	}
	if !isNotFound(err) {
		return errors.Wrapf(err, "nullifier=%s", nulliferHash)
	}

	// authorize should exist if needed
	if len(authorizeSpendHash) > 1 {
		authKey := calcAuthorizeSpendHashKey(authorizeSpendHash)
		_, err = db.Get(authKey)
		if err != nil {
			return errors.Wrapf(err, "authorize=%s", authorizeSpendHash)
		}
	}

	return nil

}

func (a *action) withdrawVerify(exec, symbol string, proof *mixTy.ZkProofInfo) (*mixTy.WithdrawCircuit, error) {
	var input mixTy.WithdrawCircuit
	err := mixTy.ConstructCircuitPubInput(proof.PublicInput, &input)
	if err != nil {
		return nil, errors.Wrapf(err, "setCircuitPubInput")
	}

	treeRootHash := input.TreeRootHash.GetWitnessValue(ecc.BN254)
	nullifierHash := input.NullifierHash.GetWitnessValue(ecc.BN254)
	authSpendHash := input.AuthorizeSpendHash.GetWitnessValue(ecc.BN254)

	err = spendVerify(a.db, exec, symbol, treeRootHash.String(), nullifierHash.String(), authSpendHash.String())
	if err != nil {
		return nil, err
	}

	err = zkProofVerify(a.db, proof, mixTy.VerifyType_WITHDRAW)
	if err != nil {
		return nil, err
	}

	return &input, nil

}

/*
1. verify(zk-proof, sum commit value)
2. withdraw from mix contract
3. set nullifier exist
*/
func (a *action) Withdraw(withdraw *mixTy.MixWithdrawAction) (*types.Receipt, error) {
	exec, symbol := mixTy.GetAssetExecSymbol(a.api.GetConfig(), withdraw.AssetExec, withdraw.AssetSymbol)

	var nulliferSet []string
	var sumValue uint64
	for _, k := range withdraw.Proofs {
		input, err := a.withdrawVerify(exec, symbol, k)
		if err != nil {
			return nil, err
		}
		v := input.Amount.GetWitnessValue(ecc.BN254)
		sumValue += v.Uint64()
		nullHash := input.NullifierHash.GetWitnessValue(ecc.BN254)
		nulliferSet = append(nulliferSet, nullHash.String())
	}

	if sumValue != withdraw.Amount {
		return nil, errors.Wrapf(mixTy.ErrInputParaNotMatch, "amount:input=%d,proof sum=%d", withdraw.Amount, sumValue)
	}

	//withdraw value
	cfg := a.api.GetConfig()
	accoutDb, err := createAccount(cfg, exec, symbol, a.db)
	if err != nil {
		return nil, err
	}
	//主链上存入toAddr为mix 执行器地址，平行链上为user.p.{}.mix执行器地址,execAddr和toAddr一致
	execAddr := address.ExecAddress(string(a.tx.Execer))
	receipt, err := accoutDb.ExecTransfer(execAddr, a.fromaddr, execAddr, int64(withdraw.Amount))
	if err != nil {
		return nil, err
	}

	//set nullifier
	for _, k := range nulliferSet {
		r := makeNullifierSetReceipt(k, &mixTy.ExistValue{Nullifier: k, Exist: true})
		mergeReceipt(receipt, r)
	}
	return receipt, nil
}
