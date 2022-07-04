package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254/twistededwards"
	"github.com/consensys/gnark/frontend"
	"github.com/pkg/errors"

	"github.com/consensys/gnark/std/algebra/twistededwards"
	"github.com/consensys/gnark/std/hash/mimc"

	ecc_bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

type Witness []fr.Element

func (witness *Witness) LimitReadFrom(r io.Reader) (int64, error) {

	var buf [4]byte
	if read, err := io.ReadFull(r, buf[:4]); err != nil {
		return int64(read), err
	}
	sliceLen := binary.BigEndian.Uint32(buf[:4])

	if len(*witness) != int(sliceLen) {
		*witness = make([]fr.Element, sliceLen)
	}

	lr := io.LimitReader(r, int64(sliceLen*fr.Limbs*8))
	dec := ecc_bn254.NewDecoder(lr)

	for i := 0; i < int(sliceLen); i++ {
		if err := dec.Decode(&(*witness)[i]); err != nil {
			return dec.BytesRead() + 4, err
		}
	}

	return dec.BytesRead() + 4, nil
}

func VerifyMerkleProof(cs frontend.API, mimc *mimc.MiMC, treeRootHash frontend.Variable, proofSet, helper, valid []frontend.Variable) {
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
func nodeSum(mimc *mimc.MiMC, a, b frontend.Variable) frontend.Variable {
	mimc.Reset()
	mimc.Write(a, b)
	return mimc.Sum()

}

// leafSum returns the hash created from data inserted to form a leaf.
// Without domain separation.
func leafSum(mimc *mimc.MiMC, data frontend.Variable) frontend.Variable {
	mimc.Reset()
	mimc.Write(data)
	return mimc.Sum()
}

func CommitValueVerify(cs frontend.API, amount, amountRandom,
	shieldAmountX, shieldAmountY, shieldPointHX, shieldPointHY frontend.Variable) {
	cs.AssertIsLessOrEqual(amount, "9000000000000000000")

	curve, _ := twistededwards.NewEdCurve(ecc.BN254)
	var pointAmount twistededwards.Point
	pointAmount.ScalarMulFixedBase(cs, curve.BaseX, curve.BaseY, amount, curve)

	var pointH twistededwards.Point
	pointH.X = shieldPointHX
	pointH.Y = shieldPointHY

	var pointRandom twistededwards.Point
	pointRandom.ScalarMulNonFixedBase(cs, &pointH, amountRandom, curve)

	var pointSum twistededwards.Point
	pointSum.AddGeneric(cs, &pointAmount, &pointRandom, curve)
	cs.AssertIsEqual(pointSum.X, shieldAmountX)
	cs.AssertIsEqual(pointSum.Y, shieldAmountY)
}

func ConstructCircuitPubInput(pubInput string, circuit frontend.Circuit) error {
	buf, err := GetByteBuff(pubInput)
	if err != nil {
		return errors.Wrapf(err, "decode string=%s", pubInput)
	}

	var witness Witness
	_, err = witness.LimitReadFrom(buf)
	if err != nil {
		return errors.Wrapf(err, "LimitReadFrom pub input=%s", pubInput)
	}

	tValue := reflect.ValueOf(circuit)
	if tValue.Kind() == reflect.Ptr {
		tValue = tValue.Elem()
	}
	for i, v := range witness {
		field := tValue.Type().Field(i)
		tValue.FieldByName(field.Name).Addr().Interface().(*frontend.Variable).Assign(v.String())
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

	point.ScalarMul(&ed.Base, &scale)
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

	pointV.ScalarMul(&pointH, &scale)
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

//A=B+C
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
