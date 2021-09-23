// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"math/big"
	"testing"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards"
	"github.com/stretchr/testify/assert"
)

const (
	baseHX = "19172955941344617222923168298456110557655645809646772800021167670156933290312"
	baseHY = "21116962883761739586121793871108889864627195706475546685847911817475098399811"
	fee    = 100000
)

func TestVerifyCommitValuesBasePoint(t *testing.T) {
	var in44, out10, out34 big.Int

	//calc p1=44*G+100*G, p2=10*G+30*G,p3=34*G+70*G, p1=p2+p3+fee
	in44.SetUint64(4400000)
	out10.SetUint64(1000000)
	out34.SetUint64(3400000 - fee)

	ed := twistededwards.GetEdwardsCurve()
	var p1, p2, p3 twistededwards.PointAffine
	p1.ScalarMul(&ed.Base, &in44)
	p2.ScalarMul(&ed.Base, &out10)
	p3.ScalarMul(&ed.Base, &out34)

	//t.Log("p1.x", p1.X.String())
	//t.Log("p1.y", p1.Y.String())
	//t.Log("p2.x", p2.X.String())
	//t.Log("p2.y", p2.Y.String())
	//t.Log("p3.x", p3.X.String())
	//t.Log("p3.y", p3.Y.String())

	var input1 mixTy.TransferInputCircuit
	input1.ShieldAmountX.Assign(p1.X.String())
	input1.ShieldAmountY.Assign(p1.Y.String())

	var inputs []*mixTy.TransferInputCircuit
	inputs = append(inputs, &input1)

	var output1, output2 mixTy.TransferOutputCircuit
	output1.ShieldAmountX.Assign(p2.X.String())
	output1.ShieldAmountY.Assign(p2.Y.String())

	output2.ShieldAmountX.Assign(p3.X.String())
	output2.ShieldAmountY.Assign(p3.Y.String())

	var outputs []*mixTy.TransferOutputCircuit
	outputs = append(outputs, &output1)
	outputs = append(outputs, &output2)

	ret := VerifyCommitValues(inputs, outputs, fee)
	assert.Equal(t, true, ret)

}

func TestVerifyCommitValuesBaseAddHPoint(t *testing.T) {
	var in44, out10, out34 big.Int
	in44.SetUint64(4400000)
	out10.SetUint64(1000000)
	out34.SetUint64(3400000 - fee)

	//random value
	var rIn100, rOut40, rOut60 big.Int
	rIn100.SetUint64(10000000)
	rOut40.SetUint64(3000000)
	rOut60.SetUint64(7000000)

	var baseH twistededwards.PointAffine
	baseH.X.SetString(baseHX)
	baseH.Y.SetString(baseHY)

	ed := twistededwards.GetEdwardsCurve()
	var p1, p2, p3, r1, r2, r3 twistededwards.PointAffine
	p1.ScalarMul(&ed.Base, &in44)
	p2.ScalarMul(&ed.Base, &out10)
	p3.ScalarMul(&ed.Base, &out34)
	r1.ScalarMul(&baseH, &rIn100)
	r2.ScalarMul(&baseH, &rOut40)
	r3.ScalarMul(&baseH, &rOut60)

	p1.Add(&p1, &r1)
	p2.Add(&p2, &r2)
	p3.Add(&p3, &r3)

	var input1 mixTy.TransferInputCircuit
	input1.ShieldAmountX.Assign(p1.X.String())
	input1.ShieldAmountY.Assign(p1.Y.String())

	var inputs []*mixTy.TransferInputCircuit
	inputs = append(inputs, &input1)

	var output1, output2 mixTy.TransferOutputCircuit
	output1.ShieldAmountX.Assign(p2.X.String())
	output1.ShieldAmountY.Assign(p2.Y.String())

	output2.ShieldAmountX.Assign(p3.X.String())
	output2.ShieldAmountY.Assign(p3.Y.String())

	var outputs []*mixTy.TransferOutputCircuit
	outputs = append(outputs, &output1)
	outputs = append(outputs, &output2)

	ret := VerifyCommitValues(inputs, outputs, fee)
	assert.Equal(t, true, ret)

}
