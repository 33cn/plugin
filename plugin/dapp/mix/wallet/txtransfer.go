// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"strconv"
	"strings"

	"github.com/33cn/chain33/common/address"

	"github.com/pkg/errors"

	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"

	fr_bn256 "github.com/consensys/gurvy/bn256/fr"
)

type TransferInput struct {
	//public
	TreeRootHash       string `tag:"public"`
	AuthorizeSpendHash string `tag:"public"`
	NullifierHash      string `tag:"public"`
	ShieldAmountX      string `tag:"public"`
	ShieldAmountY      string `tag:"public"`

	//secret
	ReceiverPubKey  string `tag:"secret"`
	ReturnPubKey    string `tag:"secret"`
	AuthorizePubKey string `tag:"secret"`
	NoteRandom      string `tag:"secret"`

	Amount        string `tag:"secret"`
	AmountRandom  string `tag:"secret"`
	SpendPriKey   string `tag:"secret"`
	SpendFlag     string `tag:"secret"`
	AuthorizeFlag string `tag:"secret"`
	NoteHash      string `tag:"secret"`

	//tree path info
	Path0 string `tag:"secret"`
	Path1 string `tag:"secret"`
	Path2 string `tag:"secret"`
	Path3 string `tag:"secret"`
	Path4 string `tag:"secret"`
	Path5 string `tag:"secret"`
	Path6 string `tag:"secret"`
	Path7 string `tag:"secret"`
	Path8 string `tag:"secret"`
	Path9 string `tag:"secret"`

	Helper0 string `tag:"secret"`
	Helper1 string `tag:"secret"`
	Helper2 string `tag:"secret"`
	Helper3 string `tag:"secret"`
	Helper4 string `tag:"secret"`
	Helper5 string `tag:"secret"`
	Helper6 string `tag:"secret"`
	Helper7 string `tag:"secret"`
	Helper8 string `tag:"secret"`
	Helper9 string `tag:"secret"`

	Valid0 string `tag:"secret"`
	Valid1 string `tag:"secret"`
	Valid2 string `tag:"secret"`
	Valid3 string `tag:"secret"`
	Valid4 string `tag:"secret"`
	Valid5 string `tag:"secret"`
	Valid6 string `tag:"secret"`
	Valid7 string `tag:"secret"`
	Valid8 string `tag:"secret"`
	Valid9 string `tag:"secret"`
}

type TransferOutput struct {
	//public
	NoteHash      string `tag:"public"`
	ShieldAmountX string `tag:"public"`
	ShieldAmountY string `tag:"public"`

	//secret
	ReceiverPubKey  string `tag:"secret"`
	ReturnPubKey    string `tag:"secret"`
	AuthorizePubKey string `tag:"secret"`
	NoteRandom      string `tag:"secret"`
	Amount          string `tag:"secret"`
	AmountRandom    string `tag:"secret"`
}

func (policy *mixPolicy) getTransferInputPart(note *mixTy.WalletIndexInfo) (*TransferInput, error) {

	//get spend privacy key
	privacyKey, err := policy.getAccountPrivacyKey(note.Account)
	if err != nil {
		return nil, errors.Wrapf(err, "getAccountPrivacyKey addr=%s", note.Account)
	}
	if privacyKey.Privacy.PaymentKey.ReceiveKey != note.Secret.ReceiverKey &&
		privacyKey.Privacy.PaymentKey.ReceiveKey != note.Secret.ReturnKey {
		return nil, errors.Wrapf(types.ErrInvalidParam, "payment pubkey from note=%s not match from privacyKey=%s,for account =%s",
			note.Secret.ReceiverKey, privacyKey.Privacy.PaymentKey.ReceiveKey, note.Account)
	}

	var input TransferInput
	initTreePath(&input)

	input.NoteHash = note.NoteHash

	input.Amount = note.Secret.Amount
	input.ReceiverPubKey = note.Secret.ReceiverKey
	input.ReturnPubKey = note.Secret.ReturnKey
	input.AuthorizePubKey = note.Secret.AuthorizeKey
	input.NoteRandom = note.Secret.NoteRandom

	//自己是payment 还是returner已经在解析note时候算好了，authSpendHash也对应算好了，如果note valid,则就用本地即可
	input.AuthorizeSpendHash = note.AuthorizeSpendHash
	input.NullifierHash = note.Nullifier

	input.SpendPriKey = privacyKey.Privacy.PaymentKey.SpendKey

	//default auto to receiver
	input.SpendFlag = "1"
	//self is returner auth to returner
	if privacyKey.Privacy.PaymentKey.ReceiveKey == note.Secret.ReturnKey {
		input.SpendFlag = "0"
	}
	input.AuthorizeFlag = "0"
	if len(input.AuthorizeSpendHash) > LENNULLKEY {
		input.AuthorizeFlag = "1"
	}

	treeProof, err := policy.getTreeProof(note.NoteHash)
	if err != nil {
		return nil, errors.Wrapf(err, "getTreeProof for hash=%s", note.NoteHash)
	}
	input.TreeRootHash = treeProof.TreeRootHash
	updateTreePath(&input, treeProof)

	return &input, nil
}

func (policy *mixPolicy) getTransferOutput(req *mixTy.DepositInfo) (*TransferOutput, *mixTy.DHSecretGroup, error) {
	resp, err := policy.depositParams(req)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "deposit toAddr")
	}

	var input TransferOutput
	input.NoteHash = resp.NoteHash
	input.Amount = resp.Proof.Amount
	input.ReceiverPubKey = resp.Proof.ReceiverKey
	input.AuthorizePubKey = resp.Proof.AuthorizeKey
	input.ReturnPubKey = resp.Proof.ReturnKey
	input.NoteRandom = resp.Proof.NoteRandom

	return &input, resp.Secrets, nil

}

func getShieldValue(noteAmount, transferAmount, minTxFee uint64) (*mixTy.ShieldAmountRst, error) {
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
		InputRandom:  rNote.String(),
		OutputRandom: rTrans.String(),
		ChangeRandom: rChange.String(),
		Input:        &mixTy.ShieldAmount{X: noteAmountG.X.String(), Y: noteAmountG.Y.String()},
		Output:       &mixTy.ShieldAmount{X: transAmountG.X.String(), Y: transAmountG.Y.String()},
		Change:       &mixTy.ShieldAmount{X: changeAmountG.X.String(), Y: changeAmountG.Y.String()},
	}
	return rst, nil
}

func (policy *mixPolicy) createTransferTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
	var transfer mixTy.TransferTxReq
	err := types.Decode(req.Data, &transfer)
	if err != nil {
		return nil, errors.Wrap(err, "decode req fail")
	}
	note, err := policy.getNoteInfo(transfer.GetInput().NoteHash, mixTy.NoteStatus_VALID)
	if err != nil {
		return nil, err
	}

	//1.获取Input
	inputPart, err := policy.getTransferInputPart(note)
	if err != nil {
		return nil, errors.Wrapf(err, "getTransferInputPart note=%s", inputPart.NoteHash)
	}
	bizlog.Info("transferProof get input succ", "notehash", inputPart.NoteHash)

	noteAmount, err := strconv.ParseUint(inputPart.Amount, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "input part parseUint=%s", inputPart.Amount)
	}

	//2. 获取output
	outPart, outDHSecret, err := policy.getTransferOutput(transfer.Output.Deposit)
	if err != nil {
		return nil, errors.Wrapf(err, "getTransferOutput note=%s", inputPart.NoteHash)
	}
	bizlog.Info("transferProof deposit to receiver succ")

	//3. 获取找零，并扣除手续费
	//如果找零为0也需要设置，否则只有一个输入一个输出，H部分的随机数要相等，就能推测出转账值来
	//在transfer output 部分特殊处理，如果amount是0的值则不加进tree
	change := &mixTy.DepositInfo{
		Addr:   note.Account,
		Amount: noteAmount - transfer.Output.Deposit.Amount - uint64(mixTy.Privacy2PrivacyTxFee),
	}
	changePart, changeDHSecret, err := policy.getTransferOutput(change)
	if err != nil {
		return nil, errors.Wrapf(err, "change part note=%s", inputPart.NoteHash)
	}
	bizlog.Info("transferProof deposit to change succ", "notehash", inputPart.NoteHash)

	//获取shieldValue 输入输出对amount隐藏
	shieldValue, err := getShieldValue(noteAmount, transfer.Output.Deposit.Amount, uint64(mixTy.Privacy2PrivacyTxFee))
	if err != nil {
		return nil, err
	}
	bizlog.Info("transferProof get commit value succ", "notehash", inputPart.NoteHash)

	//noteCommitX, transferX, changeX
	inputPart.ShieldAmountX = shieldValue.Input.X
	inputPart.ShieldAmountY = shieldValue.Input.Y
	inputPart.AmountRandom = shieldValue.InputRandom

	outPart.ShieldAmountX = shieldValue.Output.X
	outPart.ShieldAmountY = shieldValue.Output.Y
	outPart.AmountRandom = shieldValue.OutputRandom

	changePart.ShieldAmountX = shieldValue.Change.X
	changePart.ShieldAmountY = shieldValue.Change.Y
	changePart.AmountRandom = shieldValue.ChangeRandom

	//verify input
	inputProof, err := getZkProofKeys(transfer.Input.ZkPath.Path+mixTy.TransInputCircuit, transfer.Input.ZkPath.Path+mixTy.TransInputPk, *inputPart)
	if err != nil {
		return nil, errors.Wrapf(err, "input getZkProofKeys note=%s", note)
	}
	if err := policy.verifyProofOnChain(mixTy.VerifyType_TRANSFERINPUT, inputProof, transfer.Input.ZkPath.Path+mixTy.TransInputVk); err != nil {
		return nil, errors.Wrapf(err, "input verifyProof fail for note=%s", note)
	}

	//verify output
	outputProof, err := getZkProofKeys(transfer.Output.ZkPath.Path+mixTy.TransOutputCircuit, transfer.Output.ZkPath.Path+mixTy.TransOutputPk, *outPart)
	if err != nil {
		return nil, errors.Wrapf(err, "output getZkProofKeys note=%s", note)
	}
	if err := policy.verifyProofOnChain(mixTy.VerifyType_TRANSFEROUTPUT, outputProof, transfer.Output.ZkPath.Path+mixTy.TransOutputVk); err != nil {
		return nil, errors.Wrapf(err, "output verifyProof fail for note=%s", note)
	}
	outputProof.Secrets = outDHSecret

	//verify change
	changeProof, err := getZkProofKeys(transfer.Output.ZkPath.Path+mixTy.TransOutputCircuit, transfer.Output.ZkPath.Path+mixTy.TransOutputPk, *changePart)
	if err != nil {
		return nil, errors.Wrapf(err, "change getZkProofKeys note=%s", note)
	}
	if err := policy.verifyProofOnChain(mixTy.VerifyType_TRANSFEROUTPUT, changeProof, transfer.Output.ZkPath.Path+mixTy.TransOutputVk); err != nil {
		return nil, errors.Wrapf(err, "change verifyProof fail for note=%s", note)
	}
	changeProof.Secrets = changeDHSecret

	return policy.getTransferTx(strings.TrimSpace(req.Title+mixTy.MixX), inputProof, outputProof, changeProof)
}

func (policy *mixPolicy) getTransferTx(execName string, proofs ...*mixTy.ZkProofInfo) (*types.Transaction, error) {
	payload := &mixTy.MixTransferAction{}
	payload.Input = proofs[0]
	payload.Output = proofs[1]
	payload.Change = proofs[2]

	cfg := policy.getWalletOperate().GetAPI().GetConfig()
	action := &mixTy.MixAction{
		Ty:    mixTy.MixActionTransfer,
		Value: &mixTy.MixAction_Transfer{Transfer: payload},
	}

	tx := &types.Transaction{
		Execer:  []byte(execName),
		Payload: types.Encode(action),
		To:      address.ExecAddress(execName),
		Expire:  types.Now().Unix() + int64(300), //5 min
	}
	return types.FormatTx(cfg, execName, tx)
}
