package main

import (
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gnark/frontend"
	twistededwards_gadget "github.com/consensys/gnark/gadgets/algebra/twistededwards"
	"github.com/consensys/gnark/gadgets/hash/mimc"
	"github.com/consensys/gurvy"
	fr_bn256 "github.com/consensys/gurvy/bn256/fr"
)

func main() {
	circuit := NewTransferInput()
	gob.Write("circuit_transfer_input.r1cs", circuit, gurvy.BN256)
}

//spend commit hash the circuit implementing
/*
public:
	treeRootHash
	commitValueX
	commitValueY
	authorizeHash
	nullifierHash

private:
	spendAmount
	spendRandom
	spendPubKey
	returnPubKey
	authorizePubKey
	spendPriKey
	spendFlag
	authorizeFlag
	noteRandom

	path...
	helper...
	valid...
*/
func NewTransferInput() *frontend.R1CS {

	// create root constraint system
	circuit := frontend.New()

	spendValue := circuit.SECRET_INPUT("spendAmount")

	//spend pubkey
	spendPubkey := circuit.SECRET_INPUT("spendPubKey")
	returnPubkey := circuit.SECRET_INPUT("returnPubKey")
	authPubkey := circuit.SECRET_INPUT("authorizePubKey")
	spendPrikey := circuit.SECRET_INPUT("spendPriKey")
	//spend_flag 0：return_pubkey, 1:  spend_pubkey
	spendFlag := circuit.SECRET_INPUT("spendFlag")
	circuit.MUSTBE_BOOLEAN(spendFlag)
	//auth_check 0: not need auth check, 1:need check
	authFlag := circuit.SECRET_INPUT("authorizeFlag")
	circuit.MUSTBE_BOOLEAN(authFlag)

	// hash function
	mimc, _ := mimc.NewMiMCGadget("seed", gurvy.BN256)
	calcPubHash := mimc.Hash(&circuit, spendPrikey)
	targetPubHash := circuit.SELECT(spendFlag, spendPubkey, returnPubkey)
	circuit.MUSTBE_EQ(targetPubHash, calcPubHash)

	//note hash random
	noteRandom := circuit.SECRET_INPUT("noteRandom")

	//need check in database if not null
	authHash := circuit.PUBLIC_INPUT("authorizeSpendHash")

	nullValue := circuit.ALLOCATE(0)
	//// specify auth hash constraint
	calcAuthSpendHash := mimc.Hash(&circuit, targetPubHash, spendValue, noteRandom)
	targetAuthHash := circuit.SELECT(authFlag, calcAuthSpendHash, nullValue)
	circuit.MUSTBE_EQ(authHash, targetAuthHash)

	//need check in database if not null
	nullifierHash := circuit.PUBLIC_INPUT("nullifierHash")
	calcNullifierHash := mimc.Hash(&circuit, noteRandom)
	circuit.MUSTBE_EQ(nullifierHash, calcNullifierHash)

	//通过merkle tree保证noteHash存在，即便return,auth都是null也是存在的，则可以不经过授权即可消费
	noteHash := circuit.SECRET_INPUT("noteHash")
	calcReturnPubkey := circuit.SELECT(authFlag, returnPubkey, nullValue)
	calcAuthPubkey := circuit.SELECT(authFlag, authPubkey, nullValue)
	// specify note hash constraint
	preImage := mimc.Hash(&circuit, spendPubkey, calcReturnPubkey, calcAuthPubkey, spendValue, noteRandom)
	circuit.MUSTBE_EQ(noteHash, preImage)

	commitValuePart(&circuit, spendValue)
	merkelPathPart(&circuit, mimc, preImage)

	r1cs := circuit.ToR1CS()

	return r1cs
}

func commitValuePart(circuit *frontend.CS, spendValue *frontend.Constraint) {
	//cmt=transfer_value*G + random_value*H
	cmtvalueX := circuit.PUBLIC_INPUT("commitValueX")
	cmtvalueY := circuit.PUBLIC_INPUT("commitValueY")

	// set curve parameters
	edgadget, _ := twistededwards_gadget.NewEdCurveGadget(gurvy.BN256)
	// set point G in the circuit
	pointGSnark := twistededwards_gadget.NewPointGadget(circuit, nil, nil)

	//scalar := circuit.ALLOCATE("-1")
	circuit.MUSTBE_LESS_OR_EQ(spendValue, 10000000000, 256)

	// set point G in the circuit
	pointGSnark.ScalarMulFixedBase(circuit, edgadget.BaseX, edgadget.BaseY, spendValue, edgadget)
	pointGSnark.X.Tag("xg")
	pointGSnark.Y.Tag("yg")

	transfer_random := circuit.SECRET_INPUT("spendRandom")
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
