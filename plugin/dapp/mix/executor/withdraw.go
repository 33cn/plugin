// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/33cn/chain33/common"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/pkg/errors"
)

func (a *action) spendVerify(treeRootHash, nulliferHash, authorizeHash []byte) error {
	//zk-proof校验
	//check tree rootHash exist
	if !checkTreeRootHashExist(a.db, treeRootHash) {
		return errors.Wrapf(mixTy.ErrTreeRootHashNotFound, "roothash=%s", common.ToHex(treeRootHash))
	}

	//nullifier should not exist
	nullifierKey := calcNullifierHashKey(common.ToHex(nulliferHash))
	_, err := a.db.Get(nullifierKey)
	if err == nil {
		return errors.Wrapf(mixTy.ErrNulliferHashExist, "nullifier=%s", common.ToHex(nulliferHash))
	}
	if !isNotFound(err) {
		return errors.Wrapf(err, "nullifier=%s", common.ToHex(nulliferHash))
	}

	// authorize should exist if needed
	if len(authorizeHash) > 0 {
		authKey := calcAuthorizeHashKey(common.ToHex(authorizeHash))
		_, err = a.db.Get(authKey)
		if err != nil {
			return errors.Wrapf(err, "authorize=%s", common.ToHex(authorizeHash))
		}
	}

	return nil

}

func (a *action) withdrawVerify(proof *mixTy.ZkProofInfo) ([]byte, uint64, error) {
	var input mixTy.WithdrawPublicInput
	data, err := hex.DecodeString(proof.PublicInput)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "decode string=%s", proof.PublicInput)
	}
	err = json.Unmarshal(data, &input)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "unmarshal string=%s", proof.PublicInput)
	}
	val, err := strconv.ParseUint(input.Amount, 10, 64)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "parseUint=%s", input.Amount)
	}

	err = a.spendVerify(input.TreeRootHash, input.NullifierHash, input.AuthorizeHash)
	if err != nil {
		return nil, 0, err
	}

	err = a.zkProofVerify(proof, mixTy.VerifyType_WITHDRAW)
	if err != nil {
		return nil, 0, err
	}

	return input.NullifierHash, val, nil

}

/*
1. verify(zk-proof, sum commit value)
2. withdraw from mix contract
3. set nullifier exist
*/
func (a *action) Withdraw(withdraw *mixTy.MixWithdrawAction) (*types.Receipt, error) {
	var nulliferSet [][]byte
	var sumValue uint64
	for _, k := range withdraw.SpendCommits {
		nulfier, v, err := a.withdrawVerify(k)
		if err != nil {
			return nil, err
		}
		sumValue += v
		nulliferSet = append(nulliferSet, nulfier)
	}

	if sumValue != withdraw.Amount {
		return nil, mixTy.ErrInputParaNotMatch
	}

	//withdraw value
	cfg := a.api.GetConfig()
	accoutDb, err := createAccount(cfg, "", "", a.db)
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
		r := makeNullifierSetReceipt(k, &mixTy.ExistValue{Data: true})
		mergeReceipt(receipt, r)
	}

	return receipt, nil
}
