// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gnark-crypto/ecc"

	"github.com/pkg/errors"
)

func (a *action) authParamCheck(exec, symbol string, input *mixTy.AuthorizeCircuit) error {
	//check tree rootHash exist
	treeRootHash := input.TreeRootHash.GetWitnessValue(ecc.BN254)
	exist, err := checkTreeRootHashExist(a.db, exec, symbol, mixTy.Str2Byte(treeRootHash.String()))
	if err != nil {
		return errors.Wrapf(err, "roothash=%s not found,exec=%s,symbol=%s", treeRootHash.String(), exec, symbol)
	}
	if !exist {
		return errors.Wrapf(mixTy.ErrTreeRootHashNotFound, "roothash=%s", treeRootHash.String())
	}

	//authorize key should not exist
	authHash := input.AuthorizeHash.GetWitnessValue(ecc.BN254)
	authKey := calcAuthorizeHashKey(authHash.String())
	_, err = a.db.Get(authKey)
	if err == nil {
		return errors.Wrapf(mixTy.ErrAuthorizeHashExist, "auth=%s", authHash.String())
	}
	if !isNotFound(err) {
		return errors.Wrapf(err, "get auth=%s", authHash.String())
	}

	return nil
}

func (a *action) authorizePubInputs(exec, symbol string, proof *mixTy.ZkProofInfo) (*mixTy.AuthorizeCircuit, error) {
	var input mixTy.AuthorizeCircuit
	err := mixTy.ConstructCircuitPubInput(proof.PublicInput, &input)
	if err != nil {
		return nil, errors.Wrapf(err, "setCircuitPubInput")
	}

	err = a.authParamCheck(exec, symbol, &input)
	if err != nil {
		return nil, err
	}

	return &input, nil

}

/*
1. verify(zk-proof)
2, check tree root hash exist
3, check authorize pubkey hash in config or not
4. check authorize hash if exist in authorize pool
5. set authorize hash and authorize_spend hash
*/
func (a *action) Authorize(authorize *mixTy.MixAuthorizeAction) (*types.Receipt, error) {

	execer, symbol := mixTy.GetAssetExecSymbol(a.api.GetConfig(), authorize.AssetExec, authorize.AssetSymbol)
	input, err := a.authorizePubInputs(execer, symbol, authorize.ProofInfo)
	if err != nil {
		return nil, err
	}

	//zk-proof校验
	err = zkProofVerify(a.db, authorize.ProofInfo, mixTy.VerifyType_AUTHORIZE)
	if err != nil {
		return nil, err
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	authNullHash := input.AuthorizeHash.GetWitnessValue(ecc.BN254)
	r := makeReceipt(calcAuthorizeHashKey(authNullHash.String()), mixTy.TyLogAuthorizeSet, &mixTy.ExistValue{Nullifier: authNullHash.String(), Exist: true})
	mergeReceipt(receipt, r)
	authSpendHash := input.AuthorizeSpendHash.GetWitnessValue(ecc.BN254)
	r = makeReceipt(calcAuthorizeSpendHashKey(authSpendHash.String()), mixTy.TyLogAuthorizeSpendSet, &mixTy.ExistValue{Nullifier: authSpendHash.String(), Exist: true})
	mergeReceipt(receipt, r)

	return receipt, nil

}
