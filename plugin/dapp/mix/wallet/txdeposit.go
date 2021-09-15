// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/hex"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/33cn/chain33/common/address"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func (p *mixPolicy) depositParams(exec, symbol, receiver, returner, auth, amount string) (*mixTy.DepositProofResp, error) {
	if len(receiver) > 0 && len(returner) > 0 && (receiver == returner || receiver == auth || returner == auth) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "addrs should not be same to receiver=%s,return=%s,auth=%s",
			receiver, returner, auth)
	}

	//deposit 产生的secret需要确定的资产符号
	if len(exec) == 0 || len(symbol) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "asset exec=%s or symbol=%s not filled", exec, symbol)
	}

	if len(receiver) <= 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "receiver is nil")
	}

	_, e := strconv.ParseUint(amount, 0, 0)
	if e != nil {
		return nil, errors.Wrapf(e, "deposit amount=%s", amount)
	}

	var secret mixTy.SecretData
	secret.Amount = amount
	secret.AssetExec = exec
	secret.AssetSymbol = symbol

	//1. nullifier 获取随机值
	var r fr.Element
	_, err := r.SetRandom()
	if err != nil {
		return nil, errors.Wrapf(err, "getRandom")
	}
	secret.NoteRandom = r.String()
	//TODO 线上检查是否随机值在nullifer里面

	// 获取receiving addr对应的paymentKey
	receiverKey, e := p.getPaymentKey(receiver)
	if e != nil {
		return nil, errors.Wrapf(e, "get payment key for addr = %s", receiver)
	}
	secret.ReceiverKey = receiverKey.NoteReceiveAddr

	//获取return addr对应的key
	var returnKey *mixTy.NoteAccountKey

	//如果Input不填，缺省空为“0”字符串
	secret.ReturnKey = "0"
	if len(returner) > 0 {
		returnKey, err = p.getPaymentKey(returner)
		if err != nil {
			return nil, errors.Wrapf(err, "get payment key for return addr = %s", returner)
		}
		secret.ReturnKey = returnKey.NoteReceiveAddr
	}

	//获取auth addr对应的key
	var authKey *mixTy.NoteAccountKey
	secret.AuthorizeKey = "0"
	if len(auth) > 0 {
		authKey, err = p.getPaymentKey(auth)
		if err != nil {
			return nil, errors.Wrapf(err, "get payment key for authorize addr = %s", auth)
		}
		secret.AuthorizeKey = authKey.NoteReceiveAddr
	}

	//DH加密
	data := types.Encode(&secret)
	var group mixTy.DHSecretGroup

	secretData, err := encryptData(receiverKey.SecretReceiveKey, data)
	if err != nil {
		return nil, errors.Wrapf(err, "encryptData to addr = %s", receiver)
	}
	group.Receiver = hex.EncodeToString(types.Encode(secretData))
	if returnKey != nil {
		secretData, err = encryptData(returnKey.SecretReceiveKey, data)
		if err != nil {
			return nil, errors.Wrapf(err, "encryptData to addr = %s", returner)
		}
		group.Returner = hex.EncodeToString(types.Encode(secretData))
	}
	if authKey != nil {
		secretData, err = encryptData(authKey.SecretReceiveKey, data)
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
	//exec,symbol没有计算进去，不然在电路中也要引入资产，noteRandom应该足够随机区分一个note了，不然Nullifier也要区分资产
	resp.NoteHash = mixTy.Byte2Str(mimcHashString(keys))
	return &resp, nil

}

func (p *mixPolicy) getDepositProof(exec, symbol, receiver, returner, auth, amount, zkPath string, verifyOnChain bool) (*mixTy.ZkProofInfo, error) {

	resp, err := p.depositParams(exec, symbol, receiver, returner, auth, amount)
	if err != nil {
		return nil, err
	}

	var input mixTy.DepositCircuit
	input.NoteHash.Assign(resp.NoteHash)
	input.Amount.Assign(resp.Proof.Amount)
	input.ReceiverPubKey.Assign(resp.Proof.ReceiverKey)
	input.AuthorizePubKey.Assign(resp.Proof.AuthorizeKey)
	input.ReturnPubKey.Assign(resp.Proof.ReturnKey)
	input.NoteRandom.Assign(resp.Proof.NoteRandom)

	proofInfo, err := getZkProofKeys(mixTy.VerifyType_DEPOSIT, zkPath, mixTy.DepositPk, &input)
	if err != nil {
		return nil, err
	}
	//线上验证proof,失败的原因有可能circuit,Pk和线上vk不匹配，或不是一起产生的版本
	vkFile := filepath.Join(zkPath, mixTy.DepositVk)
	if err := p.verifyProofOnChain(mixTy.VerifyType_DEPOSIT, proofInfo, vkFile, verifyOnChain); err != nil {
		return nil, errors.Wrap(err, "verifyProof fail")
	}
	proofInfo.Secrets = resp.Secrets
	return proofInfo, nil
}

func (p *mixPolicy) getDepositTx(execName string, assetExec, assetSymbol string, proofs []*mixTy.ZkProofInfo) (*types.Transaction, error) {
	payload := &mixTy.MixDepositAction{}
	payload.Proofs = proofs
	payload.AssetExec = assetExec
	payload.AssetSymbol = assetSymbol

	cfg := p.getWalletOperate().GetAPI().GetConfig()
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

func (p *mixPolicy) createDepositTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
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
		p, err := p.getDepositProof(req.AssetExec, req.AssetSymbol, rcv, deposit.Deposit.ReturnAddr, deposit.Deposit.AuthorizeAddr, amounts[i], deposit.ZkPath, req.VerifyOnChain)
		if err != nil {
			return nil, errors.Wrapf(err, "get Deposit proof for=%s", rcv)
		}
		proofs = append(proofs, p)
	}
	return p.getDepositTx(strings.TrimSpace(req.Title+mixTy.MixX), req.AssetExec, req.AssetSymbol, proofs)

}
