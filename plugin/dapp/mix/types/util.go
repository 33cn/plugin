package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254/twistededwards"
	twistededwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/pkg/errors"

	"github.com/33cn/plugin/plugin/crypto/legacymimc"
	stdtwistededwards "github.com/consensys/gnark/std/algebra/native/twistededwards"

	ecc_bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

type Witness []fr.Element

// VariableToElement 将 circuit 的 public input 变量值转换回 fr.Element。
// gnark v0.9.0 后 frontend.Variable 为 interface{}，保存的是字符串值，
// 替代旧的 GetWitnessValue(ecc.BN254) 调用。
func VariableToElement(v frontend.Variable) fr.Element {
	var el fr.Element
	el.SetString(fmt.Sprint(v))
	return el
}

func VerifyMerkleProof(cs frontend.API, mimc *legacymimc.CircuitMiMC, treeRootHash frontend.Variable, proofSet, helper, valid []frontend.Variable) {
	sum := leafSum(mimc, proofSet[0])

	for i := 1; i < len(proofSet); i++ {
		cs.AssertIsBoolean(helper[i])
		d1 := cs.Select(helper[i], sum, proofSet[i])
		d2 := cs.Select(helper[i], proofSet[i], sum)
		rst := nodeSum(mimc, d1, d2)
		sum = cs.Select(valid[i], rst, sum)
	}

	// Compare our calculated Merkle root to the desired Merkle root.
	cs.AssertIsEqual(sum, treeRootHash)

}

// nodeSum returns the hash created from data inserted to form a leaf.
// Without domain separation.
func nodeSum(mimc *legacymimc.CircuitMiMC, a, b frontend.Variable) frontend.Variable {
	mimc.Reset()
	mimc.Write(a, b)
	return mimc.Sum()

}

// leafSum returns the hash created from data inserted to form a leaf.
// Without domain separation.
func leafSum(mimc *legacymimc.CircuitMiMC, data frontend.Variable) frontend.Variable {
	mimc.Reset()
	mimc.Write(data)
	return mimc.Sum()
}

func CommitValueVerify(cs frontend.API, amount, amountRandom,
	shieldAmountX, shieldAmountY, shieldPointHX, shieldPointHY frontend.Variable) {
	cs.AssertIsLessOrEqual(amount, "9000000000000000000")

	curve, _ := stdtwistededwards.NewEdCurve(cs, twistededwards.BN254)
	params := curve.Params()
	pointAmount := curve.ScalarMul(stdtwistededwards.Point{X: params.Base[0], Y: params.Base[1]}, amount)

	pointH := stdtwistededwards.Point{X: shieldPointHX, Y: shieldPointHY}

	pointRandom := curve.ScalarMul(pointH, amountRandom)

	pointSum := curve.Add(pointAmount, pointRandom)
	cs.AssertIsEqual(pointSum.X, shieldAmountX)
	cs.AssertIsEqual(pointSum.Y, shieldAmountY)
}

// ReadWitnessCompatible 兼容读取 gnark v0.5.2（旧）与 v0.9.0（新）两种 witness 字节格式。
//
// 旧格式（v0.5.2）: [uint32(n)][n×32B]，总长 %32==4
// 新格式（v0.9.0）: [uint32(nbPub)][uint32(nbSec)][uint32(len)][len×32B]，总长 %32==12
// mix/zksync 链上存量交易的 pubInput 由旧格式产生，升级后必须能继续解析以保持共识兼容。
func ReadWitnessCompatible(buf []byte) (witness.Witness, error) {
	if len(buf)%32 == 4 && len(buf) >= 4 {
		// 旧格式（v0.5.2）
		n := binary.BigEndian.Uint32(buf[:4])
		if int(n)*32+4 != len(buf) {
			return nil, errors.Wrapf(fmt.Errorf("old witness length mismatch"), "n=%d bufLen=%d", n, len(buf))
		}
		vec := make([]fr.Element, n)
		dec := ecc_bn254.NewDecoder(bytes.NewReader(buf[4:]))
		for i := range vec {
			if err := dec.Decode(&vec[i]); err != nil {
				return nil, errors.Wrapf(err, "decode old witness element %d", i)
			}
		}
		w, err := witness.New(ecc.BN254.ScalarField())
		if err != nil {
			return nil, errors.Wrapf(err, "new witness")
		}
		ch := make(chan any, len(vec))
		for _, v := range vec {
			ch <- v
		}
		close(ch)
		if err := w.Fill(int(n), 0, ch); err != nil {
			return nil, errors.Wrapf(err, "fill witness")
		}
		return w, nil
	}
	// 新格式（v0.9.0）：总长必须为 %32==12，否则视为非法输入而非静默交给 ReadFrom
	if len(buf)%32 != 12 || len(buf) < 12 {
		return nil, errors.Wrapf(fmt.Errorf("invalid witness format"), "len=%d (expect %%32==4 old or %%32==12 new)", len(buf))
	}
	w, err := witness.New(ecc.BN254.ScalarField())
	if err != nil {
		return nil, errors.Wrapf(err, "new witness")
	}
	if _, err := w.ReadFrom(bytes.NewReader(buf)); err != nil {
		return nil, errors.Wrapf(err, "read witness")
	}
	return w, nil
}

func ConstructCircuitPubInput(pubInput string, circuit frontend.Circuit) error {
	buf, err := GetByteBuff(pubInput)
	if err != nil {
		return errors.Wrapf(err, "decode string=%s", pubInput)
	}

	// 兼容读取新旧 witness 格式，与 wallet 端 w.Public().WriteTo()（新）及历史（旧）格式匹配
	pubW, err := ReadWitnessCompatible(buf.Bytes())
	if err != nil {
		return errors.Wrapf(err, "ReadWitnessCompatible pub input=%s", pubInput)
	}

	// 从 witness 提取 Vector 并赋值到 circuit 字段
	vec := pubW.Vector().(fr.Vector)
	tValue := reflect.ValueOf(circuit)
	if tValue.Kind() == reflect.Ptr {
		tValue = tValue.Elem()
	}
	for i := 0; i < len(vec); i++ {
		field := tValue.Type().Field(i)
		*(tValue.FieldByName(field.Name).Addr().Interface().(*frontend.Variable)) = vec[i].String()
	}
	return nil
}

func MulCurvePointG(val interface{}) *bn254.PointAffine {
	var v fr.Element
	v.SetInterface(val)

	var scale big.Int
	v.ToBigIntRegular(&scale)

	var point bn254.PointAffine
	ed := bn254.GetEdwardsCurve()

	point.ScalarMultiplication(&ed.Base, &scale)
	return &point
}

func MulCurvePointH(pointHX, pointHY, val string) *bn254.PointAffine {
	var v fr.Element
	v.SetInterface(val)
	var scale big.Int
	v.ToBigIntRegular(&scale)

	var pointV, pointH bn254.PointAffine
	pointH.X.SetString(pointHX)
	pointH.Y.SetString(pointHY)

	pointV.ScalarMultiplication(&pointH, &scale)
	return &pointV
}

func GetCurveSum(points ...*bn254.PointAffine) *bn254.PointAffine {

	//Add之前需初始化pointSum,不能空值，不然会等于0
	pointSum := bn254.NewPointAffine(points[0].X, points[0].Y)
	for _, a := range points[1:] {
		pointSum.Add(&pointSum, a)
	}

	return &pointSum
}

// A=B+C
func CheckSumEqual(points ...*bn254.PointAffine) bool {
	if len(points) < 2 {
		return false
	}
	//Add之前需初始化pointSum,不能空值，不然会等于0
	pointSum := bn254.NewPointAffine(points[1].X, points[1].Y)
	for _, a := range points[2:] {
		pointSum.Add(&pointSum, a)
	}

	if pointSum.X.Equal(&points[0].X) && pointSum.Y.Equal(&points[0].Y) {
		return true
	}
	return false

}

func GetByteBuff(input string) (*bytes.Buffer, error) {
	var buffInput bytes.Buffer
	res, err := hex.DecodeString(input)
	if err != nil {
		return nil, errors.Wrapf(err, "getByteBuff to %s", input)
	}
	_, err = buffInput.Write(res)
	if err != nil {
		return nil, errors.Wrapf(err, "write buff %s", input)
	}
	return &buffInput, nil

}

func Str2Byte(v string) []byte {
	var fr fr.Element
	fr.SetString(v)
	b := fr.Bytes()
	return b[:]
}
func Byte2Str(v []byte) string {
	var f fr.Element
	f.SetBytes(v)
	return f.String()
}
