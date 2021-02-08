// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/33cn/chain33/common/address"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"

	fr_bn256 "github.com/consensys/gurvy/bn256/fr"
)

type DepositInput struct {
	//public
	NoteHash string `tag:"public"`
	Amount   string `tag:"public"`

	//secret
	ReceiverPubKey  string `tag:"secret"`
	ReturnPubKey    string `tag:"secret"`
	AuthorizePubKey string `tag:"secret"`
	NoteRandom      string `tag:"secret"`
}

func (policy *mixPolicy) depositParams(req *mixTy.DepositInfo) (*mixTy.DepositProofResp, error) {
	if req == nil || len(req.Addr) <= 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "paymentAddr is nil")
	}
	if req.Amount <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "deposit amount=%d need big than 0", req.Amount)
	}

	var secret mixTy.SecretData
	secret.Amount = strconv.FormatUint(req.Amount, 10)

	//1. nullifier 获取随机值
	var fr fr_bn256.Element
	fr.SetRandom()
	secret.NoteRandom = fr.String()

	// 获取receiving addr对应的paymentKey
	toKey, e := policy.getPaymentKey(req.Addr)
	if e != nil {
		return nil, errors.Wrapf(e, "get payment key for addr = %s", req.Addr)
	}
	secret.ReceiverKey = toKey.ReceiverKey

	//获取return addr对应的key
	var returnKey *mixTy.PaymentKey
	var err error
	//如果Input不填，缺省空为“0”字符串
	secret.ReturnKey = "0"
	if len(req.ReturnAddr) > 0 {
		returnKey, err = policy.getPaymentKey(req.ReturnAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "get payment key for return addr = %s", req.ReturnAddr)
		}
		secret.ReturnKey = returnKey.ReceiverKey
	}

	//获取auth addr对应的key
	var authKey *mixTy.PaymentKey
	secret.AuthorizeKey = "0"
	if len(req.AuthorizeAddr) > 0 {
		authKey, err = policy.getPaymentKey(req.AuthorizeAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "get payment key for authorize addr = %s", req.AuthorizeAddr)
		}
		secret.AuthorizeKey = authKey.ReceiverKey
	}

	//DH加密
	data := types.Encode(&secret)
	var group mixTy.DHSecretGroup

	secretData, err := encryptData(toKey.EncryptKey, data)
	if err != nil {
		return nil, errors.Wrapf(err, "encryptData to addr = %s", req.Addr)
	}
	group.Receiver = hex.EncodeToString(types.Encode(secretData))
	if returnKey != nil {
		secretData, err = encryptData(returnKey.EncryptKey, data)
		if err != nil {
			return nil, errors.Wrapf(err, "encryptData to addr = %s", req.ReturnAddr)
		}
		group.Returner = hex.EncodeToString(types.Encode(secretData))
	}
	if authKey != nil {
		secretData, err = encryptData(authKey.EncryptKey, data)
		if err != nil {
			return nil, errors.Wrapf(err, "encryptData to addr = %s", req.AuthorizeAddr)
		}
		group.Authorize = hex.EncodeToString(types.Encode(secretData))
	}

	var resp mixTy.DepositProofResp
	resp.Proof = &secret
	resp.Secrets = &group

	keys := []string{
		secret.ReceiverKey,
		secret.ReturnKey,
		secret.AuthorizeKey,
		secret.Amount,
		secret.NoteRandom,
	}
	resp.NoteHash = mixTy.Byte2Str(mimcHashString(keys))
	return &resp, nil

}

func (policy *mixPolicy) createDepositTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
	var deposit mixTy.DepositTxReq
	err := types.Decode(req.Data, &deposit)
	if err != nil {
		return nil, errors.Wrap(err, "decode req fail")
	}

	resp, err := policy.depositParams(deposit.Deposit)
	if err != nil {
		return nil, err
	}

	var input DepositInput
	input.NoteHash = resp.NoteHash
	input.Amount = resp.Proof.Amount
	input.ReceiverPubKey = resp.Proof.ReceiverKey
	input.AuthorizePubKey = resp.Proof.AuthorizeKey
	input.ReturnPubKey = resp.Proof.ReturnKey
	input.NoteRandom = resp.Proof.NoteRandom

	proofInfo, err := getZkProofKeys(deposit.ZkPath.Path+mixTy.DepositCircuit, deposit.ZkPath.Path+mixTy.DepositPk, input)
	if err != nil {
		return nil, err
	}

	//线上验证proof,失败的原因有可能circuit,Pk和线上vk不匹配，或不是一起产生的版本
	if err := policy.verifyProofOnChain(mixTy.VerifyType_DEPOSIT, proofInfo, deposit.ZkPath.Path+mixTy.DepositVk); err != nil {
		return nil, errors.Wrap(err, "verifyProof fail")
	}
	fmt.Println("createDepositTx ok")
	proofInfo.Secrets = resp.Secrets
	return policy.getDepositTx(strings.TrimSpace(req.Title+mixTy.MixX), proofInfo)

}

func (policy *mixPolicy) getDepositTx(execName string, proof *mixTy.ZkProofInfo) (*types.Transaction, error) {
	payload := &mixTy.MixDepositAction{}
	payload.Proof = proof

	cfg := policy.getWalletOperate().GetAPI().GetConfig()
	action := &mixTy.MixAction{
		Ty:    mixTy.MixActionDeposit,
		Value: &mixTy.MixAction_Deposit{Deposit: payload},
	}

	tx := &types.Transaction{
		Execer:  []byte(execName),
		Payload: types.Encode(action),
		To:      address.ExecAddress(execName),
		Expire:  types.Now().Unix() + int64(300), //5 min
	}
	fmt.Println("createDepositTx tx")
	return types.FormatTx(cfg, execName, tx)
}
