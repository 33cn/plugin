// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/hex"
	"math/big"
	"strconv"

	"github.com/pkg/errors"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"

	fr_bn256 "github.com/consensys/gurvy/bn256/fr"
)

//对secretData 编码为string,同时增加随机值
func encodeSecretData(secret *mixTy.SecretData) (*mixTy.EncodedSecretData, error) {
	if secret == nil {
		return nil, errors.Wrap(types.ErrInvalidParam, "para is nil")
	}
	if len(secret.ReceiverPubKey) <= 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "spendPubKey is nil")
	}
	var val big.Int
	ret, succ := val.SetString(secret.Amount, 10)
	if !succ {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong amount = %s", secret.Amount)
	}
	if ret.Sign() <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount = %s, need bigger than 0", secret.Amount)
	}

	//获取随机值
	var fr fr_bn256.Element
	fr.SetRandom()
	secret.NoteRandom = fr.String()
	code := types.Encode(secret)
	var resp mixTy.EncodedSecretData

	resp.Encoded = common.ToHex(code)
	resp.RawData = secret

	return &resp, nil

}

//产生随机秘钥和receivingPk对data DH加密，返回随机秘钥的公钥
func encryptSecretData(req *mixTy.EncryptSecretData) (*mixTy.DHSecret, error) {
	secret, err := common.FromHex(req.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "decode secret")
	}

	return encryptData(req.SecretPubKey, secret), nil
}

func decryptSecretData(req *mixTy.DecryptSecretData) (*mixTy.SecretData, error) {
	secret, err := common.FromHex(req.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "decode req.secret")
	}
	decrypt, err := decryptData(req.SecretPriKey, req.Epk, secret)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt secret")
	}
	var raw mixTy.SecretData
	err = types.Decode(decrypt, &raw)
	if err != nil {
		return nil, errors.Wrap(mixTy.ErrDecryptDataFail, "decode decrypt.secret")
	}
	return &raw, nil
}

func (policy *mixPolicy) getPaymentKey(addr string) (*mixTy.PaymentKey, error) {
	msg, err := policy.walletOperate.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "mix",
		FuncName: "PaymentPubKey",
		Param:    types.Encode(&types.ReqString{Data: addr}),
	})
	if err != nil {
		return nil, err
	}
	return msg.(*mixTy.PaymentKey), err
}

func (policy *mixPolicy) depositProof(req *mixTy.DepositProofReq) (*mixTy.DepositProofResp, error) {
	if req == nil || len(req.ReceiverAddr) <= 0 {
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
	toKey, e := policy.getPaymentKey(req.ReceiverAddr)
	if e != nil {
		return nil, errors.Wrapf(e, "get payment key for addr = %s", req.ReceiverAddr)
	}
	secret.ReceiverPubKey = toKey.ReceiverKey

	//获取return addr对应的key
	var returnKey *mixTy.PaymentKey
	var err error
	//如果Input不填，缺省空为“0”字符串
	secret.ReturnPubKey = "0"
	if len(req.ReturnAddr) > 0 {
		returnKey, err = policy.getPaymentKey(req.ReturnAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "get payment key for return addr = %s", req.ReturnAddr)
		}
		secret.ReturnPubKey = returnKey.ReceiverKey
	}

	//获取auth addr对应的key
	var authKey *mixTy.PaymentKey
	secret.AuthorizePubKey = "0"
	if len(req.AuthorizeAddr) > 0 {
		authKey, err = policy.getPaymentKey(req.AuthorizeAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "get payment key for authorize addr = %s", req.AuthorizeAddr)
		}
		secret.AuthorizePubKey = authKey.ReceiverKey
	}

	//DH加密
	data := types.Encode(&secret)
	var group mixTy.DHSecretGroup

	group.Receiver = hex.EncodeToString(types.Encode(encryptData(toKey.SecretKey, data)))
	if returnKey != nil {
		group.Returner = hex.EncodeToString(types.Encode(encryptData(returnKey.SecretKey, data)))
	}
	if authKey != nil {
		group.Authorize = hex.EncodeToString(types.Encode(encryptData(authKey.SecretKey, data)))
	}

	var resp mixTy.DepositProofResp
	resp.Proof = &secret
	resp.Secrets = &group

	keys := []string{
		secret.ReceiverPubKey,
		secret.ReturnPubKey,
		secret.AuthorizePubKey,
		secret.Amount,
		secret.NoteRandom,
	}
	resp.NoteHash = getFrString(mimcHashString(keys))
	return &resp, nil

}

func (policy *mixPolicy) getPathProof(leaf string) (*mixTy.CommitTreeProve, error) {
	msg, err := policy.walletOperate.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "mix",
		FuncName: "GetTreePath",
		Param:    types.Encode(&mixTy.TreeInfoReq{LeafHash: leaf}),
	})
	if err != nil {
		return nil, err
	}
	return msg.(*mixTy.CommitTreeProve), nil
}

func (policy *mixPolicy) getNoteInfo(noteHash string, noteStatus mixTy.NoteStatus) (*mixTy.WalletIndexInfo, error) {
	if policy.walletOperate.IsWalletLocked() {
		return nil, types.ErrWalletIsLocked
	}

	var index mixTy.WalletMixIndexReq
	index.NoteHash = noteHash
	msg, err := policy.listMixInfos(&index)
	if err != nil {
		return nil, errors.Wrapf(err, "list table fail noteHash=%s", noteHash)
	}
	resp := msg.(*mixTy.WalletIndexResp)
	if len(resp.Notes) < 1 {
		return nil, errors.Wrapf(err, "list table lens=0 for noteHash=%s", noteHash)
	}

	note := msg.(*mixTy.WalletIndexResp).Notes[0]
	if note.Status != noteStatus {
		return nil, errors.Wrapf(types.ErrNotAllow, "note status=%s", note.Status.String())
	}
	return note, nil
}

func (policy *mixPolicy) withdrawProof(req *mixTy.WithdrawProofReq) (*mixTy.WithdrawProofResp, error) {
	note, err := policy.getNoteInfo(req.NoteHash, mixTy.NoteStatus_VALID)
	if err != nil {
		return nil, err
	}

	var resp mixTy.WithdrawProofResp
	resp.Secret = note.Secret
	resp.NullifierHash = note.Nullifier
	resp.AuthorizeSpendHash = note.AuthorizeSpendHash
	resp.NoteHash = note.NoteHash
	resp.SpendFlag = 1
	if note.IsReturner {
		resp.SpendFlag = 0
	}
	if len(resp.AuthorizeSpendHash) > LENNULLKEY {
		resp.AuthorizeFlag = 1
	}

	//get spend privacy key
	privacyKey, err := policy.getAccountPrivacyKey(note.Account)
	if err != nil {
		return nil, errors.Wrapf(err, "getAccountPrivacyKey addr=%s", note.Account)
	}
	resp.SpendPrivKey = privacyKey.Privacy.PaymentKey.SpendPriKey
	//get tree path
	treeProof, err := policy.getTreeProof(note.NoteHash)
	if err != nil {
		return nil, errors.Wrapf(err, "getTreeProof for hash=%s", note.NoteHash)
	}
	resp.TreeProof = treeProof

	return &resp, nil

}

func (policy *mixPolicy) authProof(req *mixTy.AuthProofReq) (*mixTy.AuthProofResp, error) {
	note, err := policy.getNoteInfo(req.NoteHash, mixTy.NoteStatus_FROZEN)
	if err != nil {
		return nil, err
	}

	//get spend privacy key
	privacyKey, err := policy.getAccountPrivacyKey(note.Account)
	if err != nil {
		return nil, errors.Wrapf(err, "getAccountPrivacyKey addr=%s", note.Account)
	}

	var resp mixTy.AuthProofResp
	resp.Proof = note.Secret
	resp.NoteHash = note.NoteHash
	resp.AuthPrivKey = privacyKey.Privacy.PaymentKey.SpendPriKey
	resp.AuthPubKey = privacyKey.Privacy.PaymentKey.ReceiverPubKey
	if privacyKey.Privacy.PaymentKey.ReceiverPubKey != note.Secret.AuthorizePubKey {
		return nil, errors.Wrapf(types.ErrInvalidParam, "auth pubkey from note=%s, from privacyKey=%s,for account =%s",
			note.Secret.AuthorizePubKey, privacyKey.Privacy.PaymentKey.ReceiverPubKey, note.Account)
	}

	resp.AuthHash = getFrString(mimcHashString([]string{resp.AuthPubKey, note.Secret.NoteRandom}))

	//default auto to paymenter
	resp.SpendFlag = 1
	resp.AuthorizeSpendHash = getFrString(mimcHashString([]string{note.Secret.ReceiverPubKey, note.Secret.Amount, note.Secret.NoteRandom}))
	if req.ToReturn != 0 {
		resp.SpendFlag = 0
		//auth to returner
		resp.AuthorizeSpendHash = getFrString(mimcHashString([]string{note.Secret.ReturnPubKey, note.Secret.Amount, note.Secret.NoteRandom}))
	}

	//get tree path
	treeProof, err := policy.getTreeProof(note.NoteHash)
	if err != nil {
		return nil, errors.Wrapf(err, "getTreeProof for hash=%s", note.NoteHash)
	}
	resp.TreeProof = treeProof

	return &resp, nil

}

func (policy *mixPolicy) getTreeProof(leaf string) (*mixTy.TreePathProof, error) {
	//get tree path
	path, err := policy.getPathProof(leaf)
	if err != nil {
		return nil, errors.Wrapf(err, "get tree proof for noteHash=%s", leaf)
	}
	var proof mixTy.TreePathProof
	proof.TreePath = path.ProofSet[1:]
	proof.Helpers = path.Helpers
	for i := 0; i < len(proof.TreePath); i++ {
		proof.ValidPath = append(proof.ValidPath, 1)
	}
	proof.TreeRootHash = path.RootHash
	return &proof, nil
}

func (policy *mixPolicy) getTransferInputPart(note *mixTy.WalletIndexInfo) (*mixTy.TransferInputProof, error) {
	//get spend privacy key
	privacyKey, err := policy.getAccountPrivacyKey(note.Account)
	if err != nil {
		return nil, errors.Wrapf(err, "getAccountPrivacyKey addr=%s", note.Account)
	}

	var input mixTy.TransferInputProof
	input.Proof = note.Secret
	input.NoteHash = note.NoteHash
	input.NullifierHash = note.Nullifier
	//自己是payment 还是returner已经在解析note时候算好了，authSpendHash也对应算好了，如果note valid,则就用本地即可
	input.AuthorizeSpendHash = note.AuthorizeSpendHash
	input.SpendPrivKey = privacyKey.Privacy.PaymentKey.SpendPriKey
	if privacyKey.Privacy.PaymentKey.ReceiverPubKey != note.Secret.ReceiverPubKey {
		return nil, errors.Wrapf(types.ErrInvalidParam, "payment pubkey from note=%s not match from privacyKey=%s,for account =%s",
			note.Secret.ReceiverPubKey, privacyKey.Privacy.PaymentKey.ReceiverPubKey, note.Account)
	}
	input.SpendFlag = 1
	if note.IsReturner {
		input.SpendFlag = 0
	}
	if len(input.AuthorizeSpendHash) > LENNULLKEY {
		input.AuthorizeFlag = 1
	}

	treeProof, err := policy.getTreeProof(note.NoteHash)
	if err != nil {
		return nil, errors.Wrapf(err, "getTreeProof for hash=%s", note.NoteHash)
	}
	input.TreeProof = treeProof

	return &input, nil
}

func getCommitValue(noteAmount, transferAmount, minTxFee uint64) (*mixTy.ShieldAmountRst, error) {
	if noteAmount < transferAmount+minTxFee {
		return nil, errors.Wrapf(types.ErrInvalidParam, "transfer amount=%d big than note=%d - fee=%d", transferAmount, noteAmount, minTxFee)
	}

	change := noteAmount - transferAmount - minTxFee
	//get amount*G point
	//note = transfer + change + minTxFee
	noteAmountG := mixTy.MulCurvePointG(noteAmount)
	transAmountG := mixTy.MulCurvePointG(transferAmount)
	changeAmountG := mixTy.MulCurvePointG(change)
	minTxFeeG := mixTy.MulCurvePointG(minTxFee)

	if !mixTy.CheckSumEqual(noteAmountG, transAmountG, changeAmountG, minTxFeeG) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount sum fail for mul G point")
	}

	//三个混淆随机值可以随机获取，这里noteRandom和为了Nullifier计算的NoteRandom不同。
	//获取随机值，截取一半给change和transfer,和值给Note,直接用完整的random值会溢出
	var rChange, rTrans, v fr_bn256.Element
	random := v.SetRandom().String()
	rChange.SetString(random[0 : len(random)/2])
	rTrans.SetString(random[len(random)/2:])

	var rNote fr_bn256.Element
	rNote.Add(&rChange, &rTrans)

	noteH := mixTy.MulCurvePointH(rNote.String())
	transferH := mixTy.MulCurvePointH(rTrans.String())
	changeH := mixTy.MulCurvePointH(rChange.String())
	//fmt.Println("change",changeRandom.String())
	//fmt.Println("transfer",transRandom.String())
	//fmt.Println("note",noteRandom.String())

	if !mixTy.CheckSumEqual(noteH, transferH, changeH) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "random sum error")
	}

	noteAmountG.Add(noteAmountG, noteH)
	transAmountG.Add(transAmountG, transferH)
	changeAmountG.Add(changeAmountG, changeH)

	if !mixTy.CheckSumEqual(noteAmountG, transAmountG, changeAmountG, minTxFeeG) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount sum fail for G+H point")
	}

	rst := &mixTy.ShieldAmountRst{
		NoteRandom:     rNote.String(),
		TransferRandom: rTrans.String(),
		ChangeRandom:   rChange.String(),
		Note:           &mixTy.ShieldAmount{X: noteAmountG.X.String(), Y: noteAmountG.Y.String()},
		Transfer:       &mixTy.ShieldAmount{X: transAmountG.X.String(), Y: transAmountG.Y.String()},
		Change:         &mixTy.ShieldAmount{X: changeAmountG.X.String(), Y: changeAmountG.Y.String()},
	}
	return rst, nil
}

func (policy *mixPolicy) transferProof(req *mixTy.TransferProofReq) (*mixTy.TransferProofResp, error) {
	note, err := policy.getNoteInfo(req.NoteHash, mixTy.NoteStatus_VALID)
	if err != nil {
		return nil, err
	}
	inputPart, err := policy.getTransferInputPart(note)
	if err != nil {
		return nil, errors.Wrapf(err, "getTransferInputPart note=%s", note.NoteHash)
	}
	bizlog.Info("transferProof get notes succ", "notehash", req.NoteHash)

	noteAmount, err := strconv.ParseUint(note.Secret.Amount, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "input part parseUint=%s", inputPart.Proof.Amount)
	}

	//output toAddr part
	reqTransfer := &mixTy.DepositProofReq{
		ReceiverAddr:  req.ToAddr,
		AuthorizeAddr: req.ToAuthAddr,
		ReturnAddr:    req.ReturnAddr,
		Amount:        req.Amount,
	}
	depositTransfer, err := policy.depositProof(reqTransfer)
	if err != nil {
		return nil, errors.Wrapf(err, "deposit toAddr")
	}
	bizlog.Info("transferProof deposit to receiver succ", "notehash", req.NoteHash)

	//还要扣除手续费
	//output 找零 part,如果找零为0也需要设置，否则只有一个输入一个输出，H部分的随机数要相等，就能推测出转账值来
	//在transfer output 部分特殊处理，如果amount是0的值则不加进tree
	reqChange := &mixTy.DepositProofReq{
		ReceiverAddr: note.Account,
		Amount:       noteAmount - req.Amount - uint64(mixTy.Privacy2PrivacyTxFee),
	}
	depositChange, err := policy.depositProof(reqChange)
	if err != nil {
		return nil, errors.Wrapf(err, "deposit toAddr")
	}
	bizlog.Info("transferProof deposit to change succ", "notehash", req.NoteHash)

	commitValue, err := getCommitValue(noteAmount, req.Amount, uint64(mixTy.Privacy2PrivacyTxFee))

	if err != nil {
		return nil, err
	}
	bizlog.Info("transferProof get commit value succ", "notehash", req.NoteHash)

	//noteCommitX, transferX, changeX
	inputPart.ShieldAmount = commitValue.Note
	inputPart.AmountRandom = commitValue.NoteRandom

	transferOutput := &mixTy.TransferOutputProof{
		Proof:        depositTransfer.Proof,
		NoteHash:     depositTransfer.NoteHash,
		Secrets:      depositTransfer.Secrets,
		ShieldAmount: commitValue.Transfer,
		AmountRandom: commitValue.TransferRandom,
	}
	changeOutput := &mixTy.TransferOutputProof{
		Proof:        depositChange.Proof,
		NoteHash:     depositChange.NoteHash,
		Secrets:      depositChange.Secrets,
		ShieldAmount: commitValue.Change,
		AmountRandom: commitValue.ChangeRandom,
	}
	return &mixTy.TransferProofResp{
		TransferInput: inputPart,
		TargetOutput:  transferOutput,
		ChangeOutput:  changeOutput}, nil
}
