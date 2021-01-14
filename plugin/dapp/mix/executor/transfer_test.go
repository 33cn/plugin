// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gurvy/bn256/fr"
	bn256 "github.com/consensys/gurvy/bn256/twistededwards"
	"github.com/stretchr/testify/assert"
)

const (
	baseHX = "10190477835300927557649934238820360529458681672073866116232821892325659279502"
	baseHY = "7969140283216448215269095418467361784159407896899334866715345504515077887397"
	fee    = 100000
)

func TestVerifyCommitValuesBasePoint(t *testing.T) {
	var in44, out10, out34, rIn100, rOut40, rOut60 fr.Element

	//calc p1=44*G+100*G, p2=10*G+30*G,p3=34*G+70*G, p1=p2+p3+fee
	in44.SetUint64(4400000).FromMont()
	out10.SetUint64(1000000).FromMont()
	out34.SetUint64(3400000 - fee).FromMont()

	rIn100.SetUint64(10000000).FromMont()
	rOut40.SetUint64(3000000).FromMont()
	rOut60.SetUint64(7000000).FromMont()

	ed := bn256.GetEdwardsCurve()
	var p1, p2, p3, r1, r2, r3 bn256.Point
	p1.ScalarMul(&ed.Base, in44)
	p2.ScalarMul(&ed.Base, out10)
	p3.ScalarMul(&ed.Base, out34)
	r1.ScalarMul(&ed.Base, rIn100)
	r2.ScalarMul(&ed.Base, rOut40)
	r3.ScalarMul(&ed.Base, rOut60)

	//p1.Add(&p1,&r1)
	//p2.Add(&p2,&r2)
	//p3.Add(&p3,&r3)

	t.Log("p1.x", p1.X.String())
	t.Log("p1.y", p1.Y.String())
	t.Log("p2.x", p2.X.String())
	t.Log("p2.y", p2.Y.String())
	t.Log("p3.x", p3.X.String())
	t.Log("p3.y", p3.Y.String())
	input1 := &mixTy.TransferInputPublicInput{
		AmountX: p1.X.String(),
		AmountY: p1.Y.String(),
	}

	var inputs []*mixTy.TransferInputPublicInput
	inputs = append(inputs, input1)

	output1 := &mixTy.TransferOutputPublicInput{
		AmountX: p2.X.String(),
		AmountY: p2.Y.String(),
	}

	output2 := &mixTy.TransferOutputPublicInput{
		AmountX: p3.X.String(),
		AmountY: p3.Y.String(),
	}

	var outputs []*mixTy.TransferOutputPublicInput
	outputs = append(outputs, output1)
	outputs = append(outputs, output2)

	ret := VerifyCommitValues(inputs, outputs)
	assert.Equal(t, true, ret)

}

func TestVerifyCommitValuesBaseAddHPoint(t *testing.T) {
	var in44, out10, out34 fr.Element
	in44.SetUint64(4400000).FromMont()
	out10.SetUint64(1000000).FromMont()
	out34.SetUint64(3400000 - fee).FromMont()

	//random value
	var rIn100, rOut40, rOut60 fr.Element
	rIn100.SetUint64(10000000).FromMont()
	rOut40.SetUint64(3000000).FromMont()
	rOut60.SetUint64(7000000).FromMont()

	var baseH bn256.Point
	baseH.X.SetString(baseHX)
	baseH.Y.SetString(baseHY)

	ed := bn256.GetEdwardsCurve()
	var p1, p2, p3, r1, r2, r3 bn256.Point
	p1.ScalarMul(&ed.Base, in44)
	p2.ScalarMul(&ed.Base, out10)
	p3.ScalarMul(&ed.Base, out34)
	r1.ScalarMul(&baseH, rIn100)
	r2.ScalarMul(&baseH, rOut40)
	r3.ScalarMul(&baseH, rOut60)

	p1.Add(&p1, &r1)
	p2.Add(&p2, &r2)
	p3.Add(&p3, &r3)

	input1 := &mixTy.TransferInputPublicInput{
		AmountX: p1.X.String(),
		AmountY: p1.Y.String(),
	}

	var inputs []*mixTy.TransferInputPublicInput
	inputs = append(inputs, input1)

	output1 := &mixTy.TransferOutputPublicInput{
		AmountX: p2.X.String(),
		AmountY: p2.Y.String(),
	}

	output2 := &mixTy.TransferOutputPublicInput{
		AmountX: p3.X.String(),
		AmountY: p3.Y.String(),
	}

	var outputs []*mixTy.TransferOutputPublicInput
	outputs = append(outputs, output1)
	outputs = append(outputs, output2)

	ret := VerifyCommitValues(inputs, outputs)
	assert.Equal(t, true, ret)

}
