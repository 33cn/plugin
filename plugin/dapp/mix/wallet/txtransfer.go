// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"path/filepath"

	"github.com/consensys/gnark-crypto/ecc"

	"strconv"
	"strings"

	"github.com/33cn/chain33/common/address"

	"github.com/pkg/errors"

	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards"
)

func (p *mixPolicy) getTransferInputPart(note *mixTy.WalletNoteInfo) (*mixTy.TransferInputCircuit, error) {
	//get spend privacy key
	privacyKey, err := p.getAccountPrivacyKey(note.Account)
	if err != nil {
		return nil, errors.Wrapf(err, "getAccountPrivacyKey addr=%s", note.Account)
	}
	if privacyKey.Privacy.PaymentKey.ReceiveKey != note.Secret.ReceiverKey &&
		privacyKey.Privacy.PaymentKey.ReceiveKey != note.Secret.ReturnKey {
		return nil, errors.Wrapf(types.ErrInvalidParam, "receiver key from note=%s not match from key=%s,for account =%s",
			note.Secret.ReceiverKey, privacyKey.Privacy.PaymentKey.ReceiveKey, note.Account)
	}

	var input mixTy.TransferInputCircuit
	input.NoteHash.Assign(note.NoteHash)

	input.Amount.Assign(note.Secret.Amount)
	input.ReceiverPubKey.Assign(note.Secret.ReceiverKey)
	input.ReturnPubKey.Assign(note.Secret.ReturnKey)
	input.AuthorizePubKey.Assign(note.Secret.AuthorizeKey)
	input.NoteRandom.Assign(note.Secret.NoteRandom)

	//自己是payment 还是returner已经在解析note时候算好了，authSpendHash也对应算好了，如果note valid,则就用本地即可
	input.AuthorizeSpendHash.Assign(note.AuthorizeSpendHash)
	input.NullifierHash.Assign(note.Nullifier)

	input.SpendPriKey.Assign(privacyKey.Privacy.PaymentKey.SpendKey)

	//self is returner auth to returner
	if privacyKey.Privacy.PaymentKey.ReceiveKey == note.Secret.ReturnKey {
		input.SpendFlag.Assign("0")
	} else {
		input.SpendFlag.Assign("1")
	}
	if len(note.AuthorizeSpendHash) > LENNULLKEY {
		input.AuthorizeFlag.Assign("1")
	} else {
		input.AuthorizeFlag.Assign("0")
	}

	treeProof, err := p.getTreeProof(note.Secret.AssetExec, note.Secret.AssetSymbol, note.NoteHash)
	if err != nil {
		return nil, errors.Wrapf(err, "getTreeProof for hash=%s", note.NoteHash)
	}
	input.TreeRootHash.Assign(treeProof.TreeRootHash)
	updateTreePath(&input, treeProof)
	return &input, nil
}

func (p *mixPolicy) getTransferOutput(exec, symbol string, req *mixTy.DepositInfo) (*mixTy.TransferOutputCircuit, *mixTy.DHSecretGroup, error) {
	//目前只支持一个ReceiverAddr
	if strings.Contains(req.ReceiverAddrs, ",") || strings.Contains(req.Amounts, ",") {
		return nil, nil, errors.Wrapf(types.ErrInvalidParam, "only support one addr or amount,addrs=%s,amount=%s",
			req.ReceiverAddrs, req.Amounts)
	}
	resp, err := p.depositParams(exec, symbol, req.ReceiverAddrs, req.ReturnAddr, req.AuthorizeAddr, req.Amounts)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "deposit toAddr=%s", req.ReceiverAddrs)
	}

	var input mixTy.TransferOutputCircuit
	input.NoteHash.Assign(resp.NoteHash)
	input.Amount.Assign(resp.Proof.Amount)
	input.ReceiverPubKey.Assign(resp.Proof.ReceiverKey)
	input.AuthorizePubKey.Assign(resp.Proof.AuthorizeKey)
	input.ReturnPubKey.Assign(resp.Proof.ReturnKey)
	input.NoteRandom.Assign(resp.Proof.NoteRandom)

	return &input, resp.Secrets, nil

}

//input = output+找零+交易费
func getShieldValue(inputAmounts []uint64, outAmount, change, minTxFee uint64, pointHX, pointHY string) (*mixTy.ShieldAmountRst, error) {
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
	var inputGPoints []*twistededwards.PointAffine
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
	var rChange, rOut, v fr.Element
	_, err := v.SetRandom()
	if err != nil {
		return nil, errors.Wrapf(err, "getRandom")
	}

	random := v.String()
	rChange.SetString(random[0 : len(random)/2])
	rOut.SetString(random[len(random)/2:])

	var rSumIn, rSumOut fr.Element
	rSumIn.SetZero()
	rSumOut.Add(&rChange, &rOut)

	var rInputs []fr.Element
	rInputs = append(rInputs, rSumOut)

	//len(inputAmounts)>1场景,每个input的随机值设为随机值的1/3长度，这样加起来不会超过rOut+rChange
	for i := 1; i < len(inputAmounts); i++ {
		var a, v fr.Element
		_, err := v.SetRandom()
		if err != nil {
			return nil, errors.Wrapf(err, "getRandom")
		}
		rv := v.String()
		a.SetString(rv[0 : len(random)/3])
		rInputs = append(rInputs, a)
		rSumIn.Add(&rSumIn, &a)
	}
	//如果len(inputAmounts)>1，则把rInputs[0]替换为rrSumOut-rSumIn,rSumIn都是1/3的随机值长度，减法应该不会溢出
	if len(rInputs) > 1 {
		var sub fr.Element
		sub.Sub(&rSumOut, &rSumIn)
		rInputs[0] = sub
	}
	rSumIn.Add(&rSumIn, &rInputs[0])
	if !rSumIn.Equal(&rSumOut) {

		return nil, errors.Wrapf(types.ErrInvalidParam, "random sumIn=%s not equal sumOut=%s", rSumIn.String(), rSumOut.String())
	}

	var inputHPoints []*twistededwards.PointAffine
	for _, i := range rInputs {
		inputHPoints = append(inputHPoints, mixTy.MulCurvePointH(pointHX, pointHY, i.String()))
	}
	//noteH := mixTy.MulCurvePointH(rNote.String())
	outH := mixTy.MulCurvePointH(pointHX, pointHY, rOut.String())
	changeH := mixTy.MulCurvePointH(pointHX, pointHY, rChange.String())
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
	}
	for _, p := range inputGPoints {
		rst.Inputs = append(rst.Inputs, &mixTy.ShieldAmount{X: p.X.String(), Y: p.Y.String()})
	}
	return rst, nil
}

func (p *mixPolicy) createTransferTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
	var transfer mixTy.TransferTxReq
	err := types.Decode(req.Data, &transfer)
	if err != nil {
		return nil, errors.Wrap(err, "decode req fail")
	}

	if len(req.AssetExec) == 0 || len(req.AssetSymbol) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "asset exec=%s or symbol=%s not filled", req.AssetExec, req.AssetSymbol)
	}

	noteHashs := strings.Split(transfer.GetInput().NoteHashs, ",")
	var notes []*mixTy.WalletNoteInfo
	for _, h := range noteHashs {
		note, err := p.getNoteInfo(h)
		if err != nil {
			return nil, errors.Wrapf(err, "get note info for=%s", h)
		}
		if note.Status != mixTy.NoteStatus_VALID && note.Status != mixTy.NoteStatus_UNFROZEN {
			return nil, errors.Wrapf(types.ErrNotAllow, "wrong note status=%s", note.Status.String())
		}
		if note.Secret.AssetExec != req.AssetExec || note.Secret.AssetSymbol != req.AssetSymbol {
			return nil, errors.Wrapf(types.ErrInvalidParam, "note=%s,exec=%s,sym=%s not equal req's exec=%s,symbol=%s",
				h, note.Secret.AssetExec, note.Secret.AssetSymbol, req.AssetExec, req.AssetSymbol)
		}
		notes = append(notes, note)
	}

	//1.获取Input
	var inputParts []*mixTy.TransferInputCircuit
	for _, n := range notes {
		input, err := p.getTransferInputPart(n)
		if err != nil {
			return nil, errors.Wrapf(err, "getTransferInputPart note=%s", n.NoteHash)
		}
		inputParts = append(inputParts, input)
	}
	bizlog.Info("transferProof get input succ", "notehash", transfer.GetInput().NoteHashs)

	var inputAmounts []uint64
	var sumInput uint64
	for _, i := range inputParts {
		amount := i.Amount.GetWitnessValue(ecc.BN254)
		inputAmounts = append(inputAmounts, amount.Uint64())
		sumInput += amount.Uint64()
	}

	//2. 获取output

	//1.如果平行链，fee可以设为0
	//2. 如果token,且不收token费，则txfee=0, 单纯扣特殊地址mixtoken的bty
	//3. 如果token,且tokenFee=true，则按fee收token数量作为交易费
	txFee := mixTy.GetTransferTxFee(p.walletOperate.GetAPI().GetConfig(), req.AssetExec)

	outAmount, err := strconv.ParseUint(transfer.Output.Deposit.Amounts, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "output part parseUint=%s", transfer.Output.Deposit.Amounts)
	}
	if outAmount == 0 {
		return nil, errors.Wrapf(err, "output part amount=0, parseUint=%s", transfer.Output.Deposit.Amounts)
	}
	if sumInput < outAmount+uint64(txFee) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "out amount=%d big than input=%d - fee=%d", outAmount, sumInput, uint64(txFee))
	}
	outPart, outDHSecret, err := p.getTransferOutput(req.AssetExec, req.AssetSymbol, transfer.Output.Deposit)
	if err != nil {
		return nil, errors.Wrap(err, "getTransferOutput for deposit")
	}
	bizlog.Info("transferProof deposit to receiver succ")

	//3. 获取找零，并扣除手续费
	//如果找零为0也需要设置，否则只有一个输入一个输出，H部分的随机数要相等，就能推测出转账值来
	//在transfer output 部分特殊处理，如果amount是0的值则不加进tree
	changeAmount := sumInput - outAmount - uint64(txFee)
	change := &mixTy.DepositInfo{
		ReceiverAddrs: notes[0].Account,
		Amounts:       strconv.FormatUint(changeAmount, 10),
	}
	changePart, changeDHSecret, err := p.getTransferOutput(req.AssetExec, req.AssetSymbol, change)
	if err != nil {
		return nil, errors.Wrap(err, "getTransferOutput change part ")
	}
	bizlog.Info("transferProof deposit to change succ")

	conf := types.ConfSub(p.walletOperate.GetAPI().GetConfig(), mixTy.MixX)
	pointHX := conf.GStr("pointHX")
	pointHY := conf.GStr("pointHY")

	//获取shieldValue 输入输出对amount隐藏
	shieldValue, err := getShieldValue(inputAmounts, outAmount, changeAmount, uint64(txFee), pointHX, pointHY)
	if err != nil {
		return nil, err
	}
	bizlog.Info("transferProof get shield value succ")

	//noteCommitX, transferX, changeX
	for i, input := range inputParts {
		input.ShieldAmountX.Assign(shieldValue.Inputs[i].X)
		input.ShieldAmountY.Assign(shieldValue.Inputs[i].Y)
		input.AmountRandom.Assign(shieldValue.InputRandoms[i])
		input.ShieldPointHX.Assign(pointHX)
		input.ShieldPointHY.Assign(pointHY)
	}

	outPart.ShieldAmountX.Assign(shieldValue.Output.X)
	outPart.ShieldAmountY.Assign(shieldValue.Output.Y)
	outPart.AmountRandom.Assign(shieldValue.OutputRandom)
	outPart.ShieldPointHX.Assign(pointHX)
	outPart.ShieldPointHY.Assign(pointHY)

	changePart.ShieldAmountX.Assign(shieldValue.Change.X)
	changePart.ShieldAmountY.Assign(shieldValue.Change.Y)
	changePart.AmountRandom.Assign(shieldValue.ChangeRandom)
	changePart.ShieldPointHX.Assign(pointHX)
	changePart.ShieldPointHY.Assign(pointHY)

	//verify input
	var inputProofs []*mixTy.ZkProofInfo
	vkFile := filepath.Join(transfer.ZkPath, mixTy.TransInputVk)
	for i, input := range inputParts {
		inputProof, err := getZkProofKeys(mixTy.VerifyType_TRANSFERINPUT, transfer.ZkPath, mixTy.TransInputPk, input)
		if err != nil {
			return nil, errors.Wrapf(err, "verify.input getZkProofKeys,the i=%d", i)
		}
		if err := p.verifyProofOnChain(mixTy.VerifyType_TRANSFERINPUT, inputProof, vkFile, req.VerifyOnChain); err != nil {
			return nil, errors.Wrapf(err, "input verifyProof fail,the i=%d", i)
		}
		inputProofs = append(inputProofs, inputProof)
	}

	//verify output
	vkOutFile := filepath.Join(transfer.ZkPath, mixTy.TransOutputVk)
	outputProof, err := getZkProofKeys(mixTy.VerifyType_TRANSFEROUTPUT, transfer.ZkPath, mixTy.TransOutputPk, outPart)
	if err != nil {
		return nil, errors.Wrapf(err, "output getZkProofKeys")
	}
	if err := p.verifyProofOnChain(mixTy.VerifyType_TRANSFEROUTPUT, outputProof, vkOutFile, req.VerifyOnChain); err != nil {
		return nil, errors.Wrapf(err, "output verifyProof fail")
	}
	outputProof.Secrets = outDHSecret

	//verify change
	changeProof, err := getZkProofKeys(mixTy.VerifyType_TRANSFEROUTPUT, transfer.ZkPath, mixTy.TransOutputPk, changePart)
	if err != nil {
		return nil, errors.Wrapf(err, "change getZkProofKeys")
	}
	if err := p.verifyProofOnChain(mixTy.VerifyType_TRANSFEROUTPUT, changeProof, vkOutFile, req.VerifyOnChain); err != nil {
		return nil, errors.Wrapf(err, "change verifyProof fail")
	}
	changeProof.Secrets = changeDHSecret

	return p.getTransferTx(strings.TrimSpace(req.Title+mixTy.MixX), req.AssetExec, req.AssetSymbol, inputProofs, outputProof, changeProof)
}

func (p *mixPolicy) getTransferTx(execName, assetExec, assetSymbol string, inputProofs []*mixTy.ZkProofInfo, output, change *mixTy.ZkProofInfo) (*types.Transaction, error) {
	payload := &mixTy.MixTransferAction{}
	payload.AssetExec = assetExec
	payload.AssetSymbol = assetSymbol
	payload.Inputs = inputProofs
	payload.Output = output
	payload.Change = change

	cfg := p.getWalletOperate().GetAPI().GetConfig()
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
