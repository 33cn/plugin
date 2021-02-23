// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/33cn/chain33/common/address"

	"github.com/pkg/errors"

	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"

	fr_bn256 "github.com/consensys/gurvy/bn256/fr"
	"github.com/consensys/gurvy/bn256/twistededwards"
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
	//目前只支持一个ReceiverAddr
	if strings.Contains(req.ReceiverAddrs, ",") || strings.Contains(req.Amounts, ",") {
		return nil, nil, errors.Wrapf(types.ErrInvalidParam, "only support one addr or amount,addrs=%s,amount=%s",
			req.ReceiverAddrs, req.Amounts)
	}
	resp, err := policy.depositParams(req.ReceiverAddrs, req.ReturnAddr, req.AuthorizeAddr, req.Amounts)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "deposit toAddr=%s", req.ReceiverAddrs)
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

//input = output+找零+交易费
func getShieldValue(inputAmounts []uint64, outAmount, change, minTxFee uint64) (*mixTy.ShieldAmountRst, error) {
	var sum uint64
	for _, i := range inputAmounts {
		sum += i
	}
	if sum != outAmount+change+minTxFee {
		return nil, errors.Wrapf(types.ErrInvalidParam, "getShieldValue.sum error,sum=%d,out=%d,change=%d,fee=%d",
			sum, outAmount, change, minTxFee)
	}
	//get amount*G point
	//note = transfer + change + minTxFee
	var inputGPoints []*twistededwards.Point
	for _, i := range inputAmounts {
		inputGPoints = append(inputGPoints, mixTy.MulCurvePointG(i))
	}
	//noteAmountG := mixTy.MulCurvePointG(inputAmount)
	outAmountG := mixTy.MulCurvePointG(outAmount)
	changeAmountG := mixTy.MulCurvePointG(change)
	minTxFeeG := mixTy.MulCurvePointG(minTxFee)

	sumPointG := mixTy.GetCurveSum(inputGPoints...)
	if !mixTy.CheckSumEqual(sumPointG, outAmountG, changeAmountG, minTxFeeG) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount sum fail for mul G point")
	}

	//三个混淆随机值可以随机获取，这里noteRandom和为了Nullifier计算的NoteRandom不同。
	//获取随机值，截取一半给change和transfer,和值给Note,直接用完整的random值会溢出
	var rChange, rOut, v fr_bn256.Element
	random := v.SetRandom().String()
	rChange.SetString(random[0 : len(random)/2])
	rOut.SetString(random[len(random)/2:])
	fmt.Println("rOut", rOut.String())
	fmt.Println("rChange", rChange.String())

	var rSumIn, rSumOut fr_bn256.Element
	rSumIn.SetZero()
	rSumOut.Add(&rChange, &rOut)

	var rInputs []fr_bn256.Element
	rInputs = append(rInputs, rSumOut)

	//len(inputAmounts)>1场景,每个input的随机值设为随机值的1/3长度，这样加起来不会超过rOut+rChange
	for i := 1; i < len(inputAmounts); i++ {
		var a, v fr_bn256.Element
		rv := v.SetRandom().String()
		a.SetString(rv[0 : len(random)/3])
		rInputs = append(rInputs, a)
		rSumIn.Add(&rSumIn, &a)
	}
	//如果len(inputAmounts)>1，则把rInputs[0]替换为rrSumOut-rSumIn,rSumIn都是1/3的随机值长度，减法应该不会溢出
	if len(rInputs) > 1 {
		var sub fr_bn256.Element
		sub.Sub(&rSumOut, &rSumIn)
		rInputs[0] = sub
	}
	rSumIn.Add(&rSumIn, &rInputs[0])
	if !rSumIn.Equal(&rSumOut) {

		return nil, errors.Wrapf(types.ErrInvalidParam, "random sumIn=%s not equal sumOut=%s", rSumIn.String(), rSumOut.String())
	}

	var inputHPoints []*twistededwards.Point
	for _, i := range rInputs {
		inputHPoints = append(inputHPoints, mixTy.MulCurvePointH(i.String()))
	}
	//noteH := mixTy.MulCurvePointH(rNote.String())
	outH := mixTy.MulCurvePointH(rOut.String())
	changeH := mixTy.MulCurvePointH(rChange.String())
	//fmt.Println("change",changeRandom.String())
	//fmt.Println("transfer",transRandom.String())
	//fmt.Println("note",noteRandom.String())
	sumPointH := mixTy.GetCurveSum(inputHPoints...)
	if !mixTy.CheckSumEqual(sumPointH, outH, changeH) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "random sum error")
	}

	for i, p := range inputGPoints {
		p.Add(p, inputHPoints[i])
	}
	outAmountG.Add(outAmountG, outH)
	changeAmountG.Add(changeAmountG, changeH)
	sumPointG = mixTy.GetCurveSum(inputGPoints...)
	if !mixTy.CheckSumEqual(sumPointG, outAmountG, changeAmountG, minTxFeeG) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount sum fail for G+H point")
	}

	rst := &mixTy.ShieldAmountRst{
		OutputRandom: rOut.String(),
		ChangeRandom: rChange.String(),
		Output:       &mixTy.ShieldAmount{X: outAmountG.X.String(), Y: outAmountG.Y.String()},
		Change:       &mixTy.ShieldAmount{X: changeAmountG.X.String(), Y: changeAmountG.Y.String()},
	}
	for _, r := range rInputs {
		rst.InputRandoms = append(rst.InputRandoms, r.String())
		fmt.Println("inputRandom", r.String())
	}
	for _, p := range inputGPoints {
		rst.Inputs = append(rst.Inputs, &mixTy.ShieldAmount{X: p.X.String(), Y: p.Y.String()})
	}
	return rst, nil
}

func (policy *mixPolicy) createTransferTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
	var transfer mixTy.TransferTxReq
	err := types.Decode(req.Data, &transfer)
	if err != nil {
		return nil, errors.Wrap(err, "decode req fail")
	}

	noteHashs := strings.Split(transfer.GetInput().NoteHashs, ",")
	var notes []*mixTy.WalletIndexInfo
	for _, h := range noteHashs {
		note, err := policy.getNoteInfo(h, mixTy.NoteStatus_VALID)
		if err != nil {
			return nil, errors.Wrapf(err, "get note info for=%s", h)
		}
		notes = append(notes, note)
	}

	//1.获取Input
	var inputParts []*TransferInput
	for _, n := range notes {
		input, err := policy.getTransferInputPart(n)
		if err != nil {
			return nil, errors.Wrapf(err, "getTransferInputPart note=%s", n.NoteHash)
		}
		inputParts = append(inputParts, input)
	}
	bizlog.Info("transferProof get input succ", "notehash", transfer.GetInput().NoteHashs)

	var inputAmounts []uint64
	var sumInput uint64
	for _, i := range inputParts {
		amount, err := strconv.ParseUint(i.Amount, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "input part parseUint=%s", i.Amount)
		}
		inputAmounts = append(inputAmounts, amount)
		sumInput += amount
	}

	//2. 获取output
	outAmount, err := strconv.ParseUint(transfer.Output.Deposit.Amounts, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "output part parseUint=%s", transfer.Output.Deposit.Amounts)
	}
	if outAmount == 0 {
		return nil, errors.Wrapf(err, "output part amount=0, parseUint=%s", transfer.Output.Deposit.Amounts)
	}
	if sumInput < outAmount+uint64(mixTy.Privacy2PrivacyTxFee) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "out amount=%d big than input=%d - fee=%d", outAmount, sumInput, uint64(mixTy.Privacy2PrivacyTxFee))
	}
	outPart, outDHSecret, err := policy.getTransferOutput(transfer.Output.Deposit)
	if err != nil {
		return nil, errors.Wrap(err, "getTransferOutput for deposit")
	}
	bizlog.Info("transferProof deposit to receiver succ")

	//3. 获取找零，并扣除手续费
	//如果找零为0也需要设置，否则只有一个输入一个输出，H部分的随机数要相等，就能推测出转账值来
	//在transfer output 部分特殊处理，如果amount是0的值则不加进tree
	changeAmount := sumInput - outAmount - uint64(mixTy.Privacy2PrivacyTxFee)
	change := &mixTy.DepositInfo{
		ReceiverAddrs: notes[0].Account,
		Amounts:       strconv.FormatUint(changeAmount, 10),
	}
	changePart, changeDHSecret, err := policy.getTransferOutput(change)
	if err != nil {
		return nil, errors.Wrap(err, "getTransferOutput change part ")
	}
	bizlog.Info("transferProof deposit to change succ")

	//获取shieldValue 输入输出对amount隐藏
	shieldValue, err := getShieldValue(inputAmounts, outAmount, changeAmount, uint64(mixTy.Privacy2PrivacyTxFee))
	if err != nil {
		return nil, err
	}
	bizlog.Info("transferProof get shield value succ")

	//noteCommitX, transferX, changeX
	for i, input := range inputParts {
		input.ShieldAmountX = shieldValue.Inputs[i].X
		input.ShieldAmountY = shieldValue.Inputs[i].Y
		input.AmountRandom = shieldValue.InputRandoms[i]
	}

	outPart.ShieldAmountX = shieldValue.Output.X
	outPart.ShieldAmountY = shieldValue.Output.Y
	outPart.AmountRandom = shieldValue.OutputRandom

	changePart.ShieldAmountX = shieldValue.Change.X
	changePart.ShieldAmountY = shieldValue.Change.Y
	changePart.AmountRandom = shieldValue.ChangeRandom

	//verify input
	var inputProofs []*mixTy.ZkProofInfo
	for i, input := range inputParts {
		inputProof, err := getZkProofKeys(transfer.Input.ZkPath+mixTy.TransInputCircuit, transfer.Input.ZkPath+mixTy.TransInputPk, *input)
		if err != nil {
			return nil, errors.Wrapf(err, "verify.input getZkProofKeys,the i=%d", i)
		}
		if err := policy.verifyProofOnChain(mixTy.VerifyType_TRANSFERINPUT, inputProof, transfer.Input.ZkPath+mixTy.TransInputVk); err != nil {
			return nil, errors.Wrapf(err, "input verifyProof fail,the i=%d", i)
		}
		inputProofs = append(inputProofs, inputProof)
	}

	//verify output
	outputProof, err := getZkProofKeys(transfer.Output.ZkPath+mixTy.TransOutputCircuit, transfer.Output.ZkPath+mixTy.TransOutputPk, *outPart)
	if err != nil {
		return nil, errors.Wrapf(err, "output getZkProofKeys")
	}
	if err := policy.verifyProofOnChain(mixTy.VerifyType_TRANSFEROUTPUT, outputProof, transfer.Output.ZkPath+mixTy.TransOutputVk); err != nil {
		return nil, errors.Wrapf(err, "output verifyProof fail")
	}
	outputProof.Secrets = outDHSecret

	//verify change
	changeProof, err := getZkProofKeys(transfer.Output.ZkPath+mixTy.TransOutputCircuit, transfer.Output.ZkPath+mixTy.TransOutputPk, *changePart)
	if err != nil {
		return nil, errors.Wrapf(err, "change getZkProofKeys")
	}
	if err := policy.verifyProofOnChain(mixTy.VerifyType_TRANSFEROUTPUT, changeProof, transfer.Output.ZkPath+mixTy.TransOutputVk); err != nil {
		return nil, errors.Wrapf(err, "change verifyProof fail")
	}
	changeProof.Secrets = changeDHSecret

	return policy.getTransferTx(strings.TrimSpace(req.Title+mixTy.MixX), inputProofs, outputProof, changeProof)
}

func (policy *mixPolicy) getTransferTx(execName string, inputProofs []*mixTy.ZkProofInfo, proofs ...*mixTy.ZkProofInfo) (*types.Transaction, error) {
	payload := &mixTy.MixTransferAction{}
	payload.Inputs = inputProofs
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
