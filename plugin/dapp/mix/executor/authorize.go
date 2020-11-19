// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"encoding/json"

	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"

	"github.com/pkg/errors"
)

func (a *action) authParamCheck(input *mixTy.AuthorizePublicInput) error {
	//check tree rootHash exist
	if !checkTreeRootHashExist(a.db, transferFr2Bytes(input.TreeRootHash)) {
		return errors.Wrapf(mixTy.ErrTreeRootHashNotFound, "roothash=%s", input.TreeRootHash)
	}

	//authorize key should not exist
	authKey := calcAuthorizeHashKey(input.AuthorizeHash)
	_, err := a.db.Get(authKey)
	if err == nil {
		return errors.Wrapf(mixTy.ErrAuthorizeHashExist, "auth=%s", input.AuthorizeHash)
	}
	if !isNotFound(err) {
		return errors.Wrapf(err, "auth=%s", input.AuthorizeHash)
	}

	authPubKeys, err := a.getAuthKeys()
	if err != nil {
		return errors.Wrap(err, "get AuthPubkey")
	}

	//authorize pubkey hash should be configured already
	var found bool
	for _, k := range authPubKeys.Data {
		if input.AuthorizePubKey == k {
			found = true
			break
		}
	}
	if !found {
		return errors.Wrapf(types.ErrNotFound, "authPubkey=%s", input.AuthorizePubKey)
	}
	return nil
}

func (a *action) authorizeVerify(proof *mixTy.ZkProofInfo) (*mixTy.AuthorizePublicInput, error) {
	var input mixTy.AuthorizePublicInput
	data, err := hex.DecodeString(proof.PublicInput)
	if err != nil {
		return nil, errors.Wrapf(err, "decode string=%s", proof.PublicInput)
	}
	err = json.Unmarshal(data, &input)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal string=%s", proof.PublicInput)
	}

	err = a.authParamCheck(&input)
	if err != nil {
		return nil, err
	}

	//zk-proof校验
	err = a.zkProofVerify(proof, mixTy.VerifyType_AUTHORIZE)
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
	var inputs []*mixTy.AuthorizePublicInput
	for _, proof := range authorize.AuthCommits {
		in, err := a.authorizeVerify(proof)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, in)
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	var auths, authSpends []string
	for _, in := range inputs {
		r := makeReceipt(calcAuthorizeHashKey(in.AuthorizeHash), mixTy.TyLogAuthorizeSet, &mixTy.ExistValue{Data: true})
		mergeReceipt(receipt, r)
		r = makeReceipt(calcAuthorizeSpendHashKey(in.AuthorizeSpendHash), mixTy.TyLogAuthorizeSpendSet, &mixTy.ExistValue{Data: true})
		mergeReceipt(receipt, r)
		auths = append(auths, in.AuthorizeHash)
		authSpends = append(authSpends, in.AuthorizeSpendHash)
	}

	return receipt, nil

}
