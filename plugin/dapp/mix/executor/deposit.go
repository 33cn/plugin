// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/zksnark"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/golang/protobuf/proto"

	"github.com/pkg/errors"
)

func makeNullifierSetReceipt(hash string, data proto.Message) *types.Receipt {
	return makeReceipt(calcNullifierHashKey(hash), mixTy.TyLogNulliferSet, data)

}

func (a *action) zkProofVerify(proof *mixTy.ZkProofInfo, verifyTy mixTy.VerifyType) error {
	keys, err := a.getVerifyKeys()
	if err != nil {
		return err
	}

	var pass bool
	for _, verifyKey := range keys.Data {
		if verifyKey.Type == verifyTy {
			ok, err := zksnark.Verify(verifyKey.Value, proof.Proof, proof.PublicInput)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			pass = true
			break
		}
	}
	if !pass {
		return errors.Wrap(mixTy.ErrZkVerifyFail, "verify")
	}

	return nil
}

func (a *action) depositVerify(proof *mixTy.ZkProofInfo) ([]byte, uint64, error) {
	var input mixTy.DepositPublicInput
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

	err = a.zkProofVerify(proof, mixTy.VerifyType_DEPOSIT)
	if err != nil {
		return nil, 0, err
	}

	return transferFr2Bytes(input.NoteHash), val, nil

}

/*
1. verify zk-proof
2. verify commit value vs value
3. deposit to mix contract
4. add new commits to merkle tree
*/
func (a *action) Deposit(deposit *mixTy.MixDepositAction) (*types.Receipt, error) {
	//1. zk-proof校验
	var sum uint64
	var commitHashs [][]byte
	for _, v := range deposit.NewCommits {
		hash, val, err := a.depositVerify(v)
		if err != nil {
			return nil, err
		}
		sum += val
		commitHashs = append(commitHashs, hash)
	}
	//校验总存款额
	if sum != deposit.Amount {
		return nil, mixTy.ErrInputParaNotMatch
	}

	//存款
	cfg := a.api.GetConfig()
	accoutDb, err := createAccount(cfg, "", "", a.db)
	if err != nil {
		return nil, errors.Wrapf(err, "createAccount")
	}
	//主链上存入toAddr为mix 执行器地址，平行链上为user.p.{}.mix执行器地址,execAddr和toAddr一致
	execAddr := address.ExecAddress(string(a.tx.Execer))
	receipt, err := accoutDb.ExecTransfer(a.fromaddr, execAddr, execAddr, int64(deposit.Amount))
	if err != nil {
		return nil, errors.Wrapf(err, "ExecTransfer")
	}
	//push new commit to merkle tree
	for _, h := range commitHashs {
		rpt, err := pushTree(a.db, h)
		if err != nil {
			return nil, err
		}
		mergeReceipt(receipt, rpt)
	}

	return receipt, nil

}
