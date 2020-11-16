package main

import (
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/gadgets/hash/mimc"
	"github.com/consensys/gurvy"
	"strconv"
)

func main() {
	circuit := NewAuth()
	gob.Write("circuit_auth.r1cs", circuit, gurvy.BN256)
}

//spend commit hash the circuit implementing
/*
public:
	treeRootHash
	authorizePubKey
	authorizeHash(=hash(authpubkey+noterandom))
	authorizeSpendHash(=hash(spendpub+value+noterandom))

private:
	spendAmount
	spendPubKey
	returnPubKey
	authorizePubKey
	authorizePriKey
	spendFlag
	noteRandom

	path...
	helper...
	valid...
*/
func NewAuth() *frontend.R1CS {

	// create root constraint system
	circuit := frontend.New()

	spendAmount := circuit.SECRET_INPUT("spendAmount")

	//spend pubkey
	spendPubKey := circuit.SECRET_INPUT("spendPubKey")
	returnPubKey := circuit.SECRET_INPUT("returnPubKey")
	authorizePriKey := circuit.SECRET_INPUT("authorizePriKey")
	noteRandom := circuit.SECRET_INPUT("noteRandom")

	authPubKey := circuit.PUBLIC_INPUT("authorizePubKey")
	authorizeHash := circuit.PUBLIC_INPUT("authorizeHash")

	// hash function
	mimc, _ := mimc.NewMiMCGadget("seed", gurvy.BN256)
	calcAuthPubKey := mimc.Hash(&circuit, authorizePriKey)
	circuit.MUSTBE_EQ(authPubKey, calcAuthPubKey)

	circuit.MUSTBE_EQ(authorizeHash, mimc.Hash(&circuit, authPubKey, noteRandom))

	//note hash random
	authSpendHash := circuit.PUBLIC_INPUT("authorizeSpendHash")
	//spend_flag 0：return_pubkey, 1:  spend_pubkey
	spendFlag := circuit.SECRET_INPUT("spendFlag")
	circuit.MUSTBE_BOOLEAN(spendFlag)
	targetPubHash := circuit.SELECT(spendFlag, spendPubKey, returnPubKey)
	calcAuthSpendHash := mimc.Hash(&circuit, targetPubHash, spendAmount, noteRandom)
	circuit.MUSTBE_EQ(authSpendHash, calcAuthSpendHash)

	//通过merkle tree保证noteHash存在，即便return,auth都是null也是存在的，则可以不经过授权即可消费
	// specify note hash constraint
	preImage := mimc.Hash(&circuit, spendPubKey, returnPubKey, authPubKey, spendAmount, noteRandom)
	merkelPathPart(&circuit, mimc, preImage)

	r1cs := circuit.ToR1CS()

	return r1cs
}

func merkelPathPart(circuit *frontend.CS, mimc mimc.MiMCGadget, noteHash *frontend.Constraint) {
	var proofSet, helper, valid []*frontend.Constraint
	merkleRoot := circuit.PUBLIC_INPUT("treeRootHash")
	proofSet = append(proofSet, noteHash)
	//helper[0],valid[0]占位， 方便接口只设置有效值
	helper = append(helper, circuit.ALLOCATE("1"))
	valid = append(valid, circuit.ALLOCATE("1"))

	//depth:10, path num need be 9
	for i := 1; i < 10; i++ {
		proofSet = append(proofSet, circuit.SECRET_INPUT("path"+strconv.Itoa(i)))
		helper = append(helper, circuit.SECRET_INPUT("helper"+strconv.Itoa(i)))
		valid = append(valid, circuit.SECRET_INPUT("valid"+strconv.Itoa(i)))
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
