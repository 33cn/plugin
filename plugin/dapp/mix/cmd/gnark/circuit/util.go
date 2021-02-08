package circuit

import (
	"strconv"

	"github.com/consensys/gnark/frontend"
	twistededwards_gadget "github.com/consensys/gnark/gadgets/algebra/twistededwards"
	"github.com/consensys/gnark/gadgets/hash/mimc"
	"github.com/consensys/gurvy"
	fr_bn256 "github.com/consensys/gurvy/bn256/fr"
)

func MerkelPathPart(circuit *frontend.CS, mimc mimc.MiMCGadget, noteHash *frontend.Constraint) {
	var proofSet, helper, valid []*frontend.Constraint
	merkleRoot := circuit.PUBLIC_INPUT("TreeRootHash")
	proofSet = append(proofSet, noteHash)
	//helper[0],valid[0]占位， 方便接口只设置有效值
	helper = append(helper, circuit.ALLOCATE("1"))
	valid = append(valid, circuit.ALLOCATE("1"))

	//depth:10, path num need be 9
	for i := 0; i < 10; i++ {
		proofSet = append(proofSet, circuit.SECRET_INPUT("Path"+strconv.Itoa(i)))
		helper = append(helper, circuit.SECRET_INPUT("Helper"+strconv.Itoa(i)))
		valid = append(valid, circuit.SECRET_INPUT("Valid"+strconv.Itoa(i)))
	}

	VerifyMerkleProof(circuit, mimc, merkleRoot, proofSet, helper, valid)
}

func VerifyMerkleProof(circuit *frontend.CS, h mimc.MiMCGadget, merkleRoot *frontend.Constraint, proofSet, helper, valid []*frontend.Constraint) {

	sum := leafSum(circuit, h, proofSet[0])

	for i := 1; i < len(proofSet); i++ {
		circuit.MUSTBE_BOOLEAN(helper[i])
		d1 := circuit.SELECT(helper[i], sum, proofSet[i])
		d2 := circuit.SELECT(helper[i], proofSet[i], sum)
		rst := nodeSum(circuit, h, d1, d2)
		sum = circuit.SELECT(valid[i], rst, sum)
	}

	// Compare our calculated Merkle root to the desired Merkle root.
	circuit.MUSTBE_EQ(sum, merkleRoot)

}

// nodeSum returns the hash created from data inserted to form a leaf.
// Without domain separation.
func nodeSum(circuit *frontend.CS, h mimc.MiMCGadget, a, b *frontend.Constraint) *frontend.Constraint {

	res := h.Hash(circuit, a, b)

	return res
}

// leafSum returns the hash created from data inserted to form a leaf.
// Without domain separation.
func leafSum(circuit *frontend.CS, h mimc.MiMCGadget, data *frontend.Constraint) *frontend.Constraint {

	res := h.Hash(circuit, data)

	return res
}

func CommitValuePart(circuit *frontend.CS, spendValue *frontend.Constraint) {
	//cmt=transfer_value*G + random_value*H
	cmtvalueX := circuit.PUBLIC_INPUT("ShieldAmountX")
	cmtvalueY := circuit.PUBLIC_INPUT("ShieldAmountY")

	// set curve parameters
	edgadget, _ := twistededwards_gadget.NewEdCurveGadget(gurvy.BN256)
	// set point G in the circuit
	pointGSnark := twistededwards_gadget.NewPointGadget(circuit, nil, nil)

	//to avoid <0 values input
	//negOne := circuit.ALLOCATE("-1")
	//negSpendVal := circuit.MUL(spendValue,negOne)
	//circuit.MUSTBE_LESS_OR_EQ(negSpendVal, 0, 256)
	circuit.MUSTBE_LESS_OR_EQ(spendValue, 1000000000000000000, 256)

	// set point G in the circuit
	pointGSnark.ScalarMulFixedBase(circuit, edgadget.BaseX, edgadget.BaseY, spendValue, edgadget)
	pointGSnark.X.Tag("xg")
	pointGSnark.Y.Tag("yg")

	transfer_random := circuit.SECRET_INPUT("AmountRandom")
	//circuit.MUSTBE_LESS_OR_EQ(random_value,10000000000,256)
	//H is not G, H should be a point that no one know the prikey
	var baseX_H, baseY_H fr_bn256.Element
	baseX_H.SetString("10190477835300927557649934238820360529458681672073866116232821892325659279502")
	baseY_H.SetString("7969140283216448215269095418467361784159407896899334866715345504515077887397")
	pointHSnark := twistededwards_gadget.NewPointGadget(circuit, nil, nil)
	// add points in circuit (the method updates the underlying plain points as well)
	pointHSnark.ScalarMulFixedBase(circuit, baseX_H, baseY_H, transfer_random, edgadget)
	pointHSnark.X.Tag("xh")
	pointHSnark.Y.Tag("yh")

	pointSumSnark := twistededwards_gadget.NewPointGadget(circuit, nil, nil)
	pointSumSnark.AddGeneric(circuit, &pointGSnark, &pointHSnark, edgadget)

	//cmtvalue=transfer_value*G + random_value*H
	circuit.MUSTBE_EQ(cmtvalueX, pointSumSnark.X)
	circuit.MUSTBE_EQ(cmtvalueY, pointSumSnark.Y)
}
