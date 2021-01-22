// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
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
	if len(secret.PaymentPubKey) <= 0 {
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

	return encryptData(req.ReceivingPk, secret), nil
}

func decryptSecretData(req *mixTy.DecryptSecretData) (*mixTy.SecretData, error) {
	secret, err := common.FromHex(req.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "decode req.secret")
	}
	decrypt, err := decryptData(req.ReceivingPriKey, req.Epk, secret)
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
	if req == nil || len(req.PaymentAddr) <= 0 {
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

	// 获取addr对应的paymentKey
	toKey, errr := policy.getPaymentKey(req.PaymentAddr)
	if errr != nil {
		return nil, errors.Wrapf(errr, "get payment key for addr = %s", req.PaymentAddr)
	}
	secret.PaymentPubKey = toKey.PayingKey

	var returnKey *mixTy.PaymentKey
	var err error
	if len(req.ReturnAddr) > 0 {
		returnKey, err = policy.getPaymentKey(req.PaymentAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "get payment key for addr = %s", req.PaymentAddr)
		}
		secret.ReturnPubKey = returnKey.PayingKey
	}

	var authKey *mixTy.PaymentKey
	if len(req.ReturnAddr) > 0 {
		authKey, err = policy.getPaymentKey(req.PaymentAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "get payment key for addr = %s", req.PaymentAddr)
		}
		secret.AuthorizePubKey = authKey.PayingKey
	}

	//DH加密
	data := types.Encode(&secret)
	var group mixTy.DHSecretGroup
	group.Payment = encryptData(toKey.ReceivingKey, data)
	if returnKey != nil {
		group.Returner = encryptData(returnKey.ReceivingKey, data)
	}
	if authKey != nil {
		group.Authorize = encryptData(authKey.ReceivingKey, data)
	}

	var resp mixTy.DepositProofResp
	resp.Proof = &secret
	resp.Secrets = &group

	keys := []string{
		secret.PaymentPubKey,
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

	note := msg.(*mixTy.WalletIndexResp).Datas[0]
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
	resp.Proof = note.Secret
	resp.NullifierHash = note.Nullifier
	resp.AuthSpendHash = note.AuthSpendHash
	resp.NoteHash = note.NoteHash
	resp.SpendFlag = 1
	if note.IsReturner {
		resp.SpendFlag = 0
	}
	if len(resp.AuthSpendHash) > 0 {
		resp.AuthFlag = 1
	}

	//get spend privacy key
	privacyKey, err := policy.getAccountPrivacyKey(note.Account)
	if err != nil {
		return nil, errors.Wrapf(err, "getAccountPrivacyKey addr=%s", note.Account)
	}
	resp.SpendPrivKey = privacyKey.Privacy.PaymentKey.SpendKey
	//get tree path
	path, err := policy.getPathProof(note.NoteHash)
	if err != nil {
		return nil, errors.Wrapf(err, "get tree proof for noteHash=%s", note.NoteHash)
	}
	resp.TreeProof.TreePath = path.ProofSet[1:]
	resp.TreeProof.Helpers = path.Helpers
	for i := 0; i < len(resp.TreeProof.TreePath); i++ {
		resp.TreeProof.ValidPath = append(resp.TreeProof.ValidPath, 1)
	}
	resp.TreeProof.TreeRootHash = path.RootHash

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
	resp.AuthPrivKey = privacyKey.Privacy.PaymentKey.SpendKey
	resp.AuthPubKey = privacyKey.Privacy.PaymentKey.PayKey
	if privacyKey.Privacy.PaymentKey.PayKey != note.Secret.AuthorizePubKey {
		return nil, errors.Wrapf(types.ErrInvalidParam, "auth pubkey from note=%s, from privacyKey=%s,for account =%s",
			note.Secret.AuthorizePubKey, privacyKey.Privacy.PaymentKey.PayKey, note.Account)
	}

	resp.AuthHash = getFrString(mimcHashString([]string{resp.AuthPubKey, note.Secret.NoteRandom}))

	//default auto to paymenter
	resp.SpendFlag = 1
	resp.AuthSpendHash = getFrString(mimcHashString([]string{note.Secret.PaymentPubKey, note.Secret.Amount, note.Secret.NoteRandom}))
	if req.ToReturn != 0 {
		resp.SpendFlag = 0
		//auth to returner
		resp.AuthSpendHash = getFrString(mimcHashString([]string{note.Secret.ReturnPubKey, note.Secret.Amount, note.Secret.NoteRandom}))
	}

	//get tree path
	path, err := policy.getPathProof(note.NoteHash)
	if err != nil {
		return nil, errors.Wrapf(err, "get tree proof for noteHash=%s", note.NoteHash)
	}
	resp.TreeProof.TreePath = path.ProofSet[1:]
	resp.TreeProof.Helpers = path.Helpers
	for i := 0; i < len(resp.TreeProof.TreePath); i++ {
		resp.TreeProof.ValidPath = append(resp.TreeProof.ValidPath, 1)
	}
	resp.TreeProof.TreeRootHash = path.RootHash

	return &resp, nil

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
	input.AuthSpendHash = note.AuthSpendHash
	input.SpendPrivKey = privacyKey.Privacy.PaymentKey.SpendKey
	if privacyKey.Privacy.PaymentKey.PayKey != note.Secret.AuthorizePubKey {
		return nil, errors.Wrapf(types.ErrInvalidParam, "payment pubkey from note=%s not match from privacyKey=%s,for account =%s",
			note.Secret.AuthorizePubKey, privacyKey.Privacy.PaymentKey.PayKey, note.Account)
	}
	input.SpendFlag = 1
	if note.IsReturner {
		input.SpendFlag = 0
	}
	if len(input.AuthSpendHash) > 0 {
		input.AuthFlag = 1
	}
	return &input, nil
}

func getCommitValue(noteAmount, transferAmount, minTxFee uint64) (*mixTy.CommitValueRst, error) {
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

	//获取随机值，截取一半给change和transfer,和值给Note,直接用完整的random值会溢出
	var changeRandom, transRandom, v fr_bn256.Element
	random := v.SetRandom().String()
	changeRandom.SetString(random[0 : len(random)/2])
	transRandom.SetString(random[len(random)/2:])

	var noteRandom fr_bn256.Element
	noteRandom.Add(&changeRandom, &transRandom)

	noteH := mixTy.MulCurvePointH(noteRandom.String())
	transferH := mixTy.MulCurvePointH(transRandom.String())
	changeH := mixTy.MulCurvePointH(changeRandom.String())
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

	rst := &mixTy.CommitValueRst{
		NoteRandom:     noteRandom.String(),
		TransferRandom: transRandom.String(),
		ChangeRandom:   changeRandom.String(),
		Note:           &mixTy.CommitValue{X: noteAmountG.X.String(), Y: noteAmountG.Y.String()},
		Transfer:       &mixTy.CommitValue{X: transAmountG.X.String(), Y: transAmountG.Y.String()},
		Change:         &mixTy.CommitValue{X: changeAmountG.X.String(), Y: changeAmountG.Y.String()},
	}
	return rst, nil
}

func (policy *mixPolicy) transferProof(req *mixTy.TransferProofReq) (*mixTy.TransferProofResp, error) {
	note, err := policy.getNoteInfo(req.NoteHash, mixTy.NoteStatus_VALID)
	if err != nil {
		return nil, err
	}
	inputPart, err := policy.getTransferInputPart(note)

	noteAmount, err := strconv.ParseUint(note.Secret.Amount, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "input part parseUint=%s", inputPart.Proof.Amount)
	}
	//还要扣除手续费
	minTxFee := uint64(policy.walletOperate.GetConfig().MinFee)

	//output toAddr part
	reqTransfer := &mixTy.DepositProofReq{
		PaymentAddr:   req.ToAddr,
		AuthorizeAddr: req.ToAuthAddr,
		ReturnAddr:    req.ReturnAddr,
		Amount:        req.Amount,
	}
	depositTransfer, err := policy.depositProof(reqTransfer)
	if err != nil {
		return nil, errors.Wrapf(err, "deposit toAddr")
	}

	//output 找零 part,如果找零为0也需要设置，否则只有一个输入一个输出，H部分的随机数要相等，就能推测出转账值来
	//在transfer output 部分特殊处理，如果amount是0的值则不加进tree
	reqChange := &mixTy.DepositProofReq{
		PaymentAddr: note.Account,
		Amount:      noteAmount - req.Amount - minTxFee,
	}
	depositChange, err := policy.depositProof(reqChange)
	if err != nil {
		return nil, errors.Wrapf(err, "deposit toAddr")
	}

	commitValue, err := getCommitValue(noteAmount, req.Amount, minTxFee)
	if err != nil {
		return nil, err
	}

	//noteCommitX, transferX, changeX
	inputPart.CommitValue = commitValue.Note
	inputPart.SpendRandom = commitValue.NoteRandom
	transferOutput := &mixTy.TransferOutputProof{
		Proof:       depositTransfer.Proof,
		NoteHash:    depositTransfer.NoteHash,
		Secrets:     depositTransfer.Secrets,
		CommitValue: commitValue.Transfer,
		SpendRandom: commitValue.TransferRandom,
	}
	changeOutput := &mixTy.TransferOutputProof{
		Proof:       depositChange.Proof,
		NoteHash:    depositChange.NoteHash,
		Secrets:     depositChange.Secrets,
		CommitValue: commitValue.Change,
		SpendRandom: commitValue.ChangeRandom,
	}

	return &mixTy.TransferProofResp{
		TransferInput: inputPart,
		TargetOutput:  transferOutput,
		ChangeOutput:  changeOutput}, nil
}
