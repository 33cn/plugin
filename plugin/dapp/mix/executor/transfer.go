// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"encoding/json"

	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gurvy/bn256/twistededwards"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/consensys/gurvy/bn256/fr"
	"github.com/pkg/errors"
)

/*
1. verify(zk-proof)
2. check if exist in authorize pool and nullifier pool

*/
func transferInputVerify(db dbm.KV, proof *mixTy.ZkProofInfo) (*mixTy.TransferInputPublicInput, error) {
	var input mixTy.TransferInputPublicInput
	data, err := hex.DecodeString(proof.PublicInput)
	if err != nil {
		return nil, errors.Wrapf(err, "transferInput verify decode string=%s", proof.PublicInput)
	}
	err = json.Unmarshal(data, &input)
	if err != nil {
		return nil, errors.Wrapf(err, "transferInput verify unmarshal string=%s", proof.PublicInput)
	}

	err = spendVerify(db, input.TreeRootHash, input.NullifierHash, input.AuthorizeSpendHash)
	if err != nil {
		return nil, errors.Wrap(err, "transferInput verify spendVerify")
	}

	err = zkProofVerify(db, proof, mixTy.VerifyType_TRANSFERINPUT)
	if err != nil {
		return nil, errors.Wrap(err, "transferInput verify proof verify")
	}

	return &input, nil

}

/*
1. verify(zk-proof)
2. check if exist in authorize pool and nullifier pool

*/
func transferOutputVerify(db dbm.KV, proof *mixTy.ZkProofInfo) (*mixTy.TransferOutputPublicInput, error) {
	var input mixTy.TransferOutputPublicInput
	data, err := hex.DecodeString(proof.PublicInput)
	if err != nil {
		return nil, errors.Wrapf(err, "Output verify decode string=%s", proof.PublicInput)
	}
	err = json.Unmarshal(data, &input)
	if err != nil {
		return nil, errors.Wrapf(err, "Output verify  unmarshal string=%s", proof.PublicInput)
	}

	err = zkProofVerify(db, proof, mixTy.VerifyType_TRANSFEROUTPUT)
	if err != nil {
		return nil, errors.Wrap(err, "Output verify proof verify")
	}

	return &input, nil

}

func getFee() *twistededwards.Point {
	//手续费 可配， 缺省100000， 即0.001， point=fee*G + 0*H
	var fee fr.Element
	fee.SetUint64(100000).FromMont()

	var pointFee twistededwards.Point
	ed := twistededwards.GetEdwardsCurve()
	pointFee.ScalarMul(&ed.Base, fee)
	return &pointFee
}

func VerifyCommitValues(inputs []*mixTy.TransferInputPublicInput, outputs []*mixTy.TransferOutputPublicInput) bool {
	var inputPoints, outputPoints []*twistededwards.Point
	for _, in := range inputs {
		var p twistededwards.Point
		p.X.SetString(in.AmountX)
		p.Y.SetString(in.AmountY)
		inputPoints = append(inputPoints, &p)
	}

	for _, out := range outputs {
		var p twistededwards.Point
		p.X.SetString(out.AmountX)
		p.Y.SetString(out.AmountY)
		outputPoints = append(outputPoints, &p)
	}
	//out value add fee
	outputPoints = append(outputPoints, getFee())

	//sum input and output
	sumInput := inputPoints[0]
	sumOutput := outputPoints[0]
	for _, p := range inputPoints[1:] {
		sumInput.Add(sumInput, p)
	}
	for _, p := range outputPoints[1:] {
		sumOutput.Add(sumOutput, p)
	}

	if sumInput.X.Equal(&sumOutput.X) && sumInput.Y.Equal(&sumOutput.Y) {
		return true
	}
	return false
}

func MixTransferInfoVerify(db dbm.KV, transfer *mixTy.MixTransferAction) ([]*mixTy.TransferInputPublicInput, []*mixTy.TransferOutputPublicInput, error) {
	var inputs []*mixTy.TransferInputPublicInput
	var outputs []*mixTy.TransferOutputPublicInput

	for _, k := range transfer.Input {
		in, err := transferInputVerify(db, k)
		if err != nil {
			return nil, nil, err
		}
		inputs = append(inputs, in)
	}

	for _, k := range transfer.Output {
		out, err := transferOutputVerify(db, k)
		if err != nil {
			return nil, nil, err
		}
		outputs = append(outputs, out)
	}

	if !VerifyCommitValues(inputs, outputs) {
		return nil, nil, errors.Wrap(mixTy.ErrSpendInOutValueNotMatch, "verifyValue")
	}

	return inputs, outputs, nil
}

/*
1. verify(zk-proof, sum value of spend and new commits)
2. check if exist in authorize pool and nullifier pool
3. add nullifier to pool
*/
func (a *action) Transfer(transfer *mixTy.MixTransferAction) (*types.Receipt, error) {
	inputs, outputs, err := MixTransferInfoVerify(a.db, transfer)
	if err != nil {
		return nil, errors.Wrap(err, "Transfer.MixTransferInfoVerify")
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	for _, k := range inputs {
		r := makeNullifierSetReceipt(k.NullifierHash, &mixTy.ExistValue{Data: true})
		mergeReceipt(receipt, r)
	}

	//push new commit to merkle tree
	var leaves [][]byte
	for _, h := range outputs {
		leaves = append(leaves, transferFr2Bytes(h.NoteHash))
	}
	rpt, err := pushTree(a.db, leaves)
	if err != nil {
		return nil, errors.Wrap(err, "transfer.pushTree")
	}
	mergeReceipt(receipt, rpt)
	return receipt, nil

}
