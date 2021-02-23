// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/hex"
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

func (policy *mixPolicy) depositParams(receiver, returner, auth, amount string) (*mixTy.DepositProofResp, error) {
	if len(receiver) <= 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "receiver is nil")
	}

	_, e := strconv.ParseUint(amount, 0, 0)
	if e != nil {
		return nil, errors.Wrapf(e, "deposit amount=%s", amount)
	}

	var secret mixTy.SecretData
	secret.Amount = amount

	//1. nullifier 获取随机值
	var fr fr_bn256.Element
	fr.SetRandom()
	secret.NoteRandom = fr.String()

	//TODO 线上检查是否随机值在nullifer里面

	// 获取receiving addr对应的paymentKey
	payKeys, e := policy.getPaymentKey(receiver)
	if e != nil {
		return nil, errors.Wrapf(e, "get payment key for addr = %s", receiver)
	}
	secret.ReceiverKey = payKeys.ReceiverKey

	//获取return addr对应的key
	var returnKey *mixTy.PaymentKey
	var err error
	//如果Input不填，缺省空为“0”字符串
	secret.ReturnKey = "0"
	if len(returner) > 0 {
		returnKey, err = policy.getPaymentKey(returner)
		if err != nil {
			return nil, errors.Wrapf(err, "get payment key for return addr = %s", returner)
		}
		secret.ReturnKey = returnKey.ReceiverKey
	}

	//获取auth addr对应的key
	var authKey *mixTy.PaymentKey
	secret.AuthorizeKey = "0"
	if len(auth) > 0 {
		authKey, err = policy.getPaymentKey(auth)
		if err != nil {
			return nil, errors.Wrapf(err, "get payment key for authorize addr = %s", auth)
		}
		secret.AuthorizeKey = authKey.ReceiverKey
	}

	//DH加密
	data := types.Encode(&secret)
	var group mixTy.DHSecretGroup

	secretData, err := encryptData(payKeys.EncryptKey, data)
	if err != nil {
		return nil, errors.Wrapf(err, "encryptData to addr = %s", receiver)
	}
	group.Receiver = hex.EncodeToString(types.Encode(secretData))
	if returnKey != nil {
		secretData, err = encryptData(returnKey.EncryptKey, data)
		if err != nil {
			return nil, errors.Wrapf(err, "encryptData to addr = %s", returner)
		}
		group.Returner = hex.EncodeToString(types.Encode(secretData))
	}
	if authKey != nil {
		secretData, err = encryptData(authKey.EncryptKey, data)
		if err != nil {
			return nil, errors.Wrapf(err, "encryptData to addr = %s", auth)
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

func (policy *mixPolicy) getDepositProof(receiver, returner, auth, amount, zkPath string) (*mixTy.ZkProofInfo, error) {

	resp, err := policy.depositParams(receiver, returner, auth, amount)
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

	proofInfo, err := getZkProofKeys(zkPath+mixTy.DepositCircuit, zkPath+mixTy.DepositPk, input)
	if err != nil {
		return nil, err
	}

	//线上验证proof,失败的原因有可能circuit,Pk和线上vk不匹配，或不是一起产生的版本
	if err := policy.verifyProofOnChain(mixTy.VerifyType_DEPOSIT, proofInfo, zkPath+mixTy.DepositVk); err != nil {
		return nil, errors.Wrap(err, "verifyProof fail")
	}
	proofInfo.Secrets = resp.Secrets
	return proofInfo, nil
}

func (policy *mixPolicy) createDepositTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
	var deposit mixTy.DepositTxReq
	err := types.Decode(req.Data, &deposit)
	if err != nil {
		return nil, errors.Wrap(err, "decode req fail")
	}

	if deposit.Deposit == nil {
		return nil, errors.Wrap(err, "decode deposit  fail")
	}

	if len(deposit.ZkPath) == 0 {
		deposit.ZkPath = "./"
	}

	//多个receiver
	receivers := strings.Split(deposit.Deposit.ReceiverAddrs, ",")
	amounts := strings.Split(deposit.Deposit.Amounts, ",")
	if len(receivers) != len(amounts) || len(receivers) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "not match receivers=%s and amounts=%s", deposit.Deposit.ReceiverAddrs, deposit.Deposit.Amounts)
	}

	var proofs []*mixTy.ZkProofInfo
	for i, rcv := range receivers {
		p, err := policy.getDepositProof(rcv, deposit.Deposit.ReturnAddr, deposit.Deposit.AuthorizeAddr, amounts[i], deposit.ZkPath)
		if err != nil {
			return nil, errors.Wrapf(err, "get Deposit proof for=%s", rcv)
		}
		proofs = append(proofs, p)
	}

	return policy.getDepositTx(strings.TrimSpace(req.Title+mixTy.MixX), proofs)

}

func (policy *mixPolicy) getDepositTx(execName string, proofs []*mixTy.ZkProofInfo) (*types.Transaction, error) {
	payload := &mixTy.MixDepositAction{}
	payload.Proofs = proofs

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

	return types.FormatTx(cfg, execName, tx)
}
