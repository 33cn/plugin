package main

import (
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/gadgets/hash/mimc"
	"github.com/consensys/gurvy"
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

	spendValue := circuit.SECRET_INPUT("spendAmount")

	//spend pubkey
	spendPubkey := circuit.SECRET_INPUT("spendPubKey")
	returnPubkey := circuit.SECRET_INPUT("returnPubKey")
	authPubkey := circuit.SECRET_INPUT("authorizePubKey")
	authorizePrikey := circuit.SECRET_INPUT("authorizePriKey")

	authPubHashInput := circuit.PUBLIC_INPUT("authorizePubKey")
	// hash function
	mimc, _ := mimc.NewMiMCGadget("seed", gurvy.BN256)
	calcAuthPubHash := mimc.Hash(&circuit, authorizePrikey)
	circuit.MUSTBE_EQ(authPubkey, calcAuthPubHash)
	circuit.MUSTBE_EQ(authPubHashInput, mimc.Hash(&circuit, authPubkey))

	//note hash random
	noteRandom := circuit.SECRET_INPUT("noteRandom")
	authSpendHash := circuit.PUBLIC_INPUT("authorizeSpendHash")
	//spend_flag 0：return_pubkey, 1:  spend_pubkey
	spendFlag := circuit.SECRET_INPUT("spendFlag")
	circuit.MUSTBE_BOOLEAN(spendFlag)
	targetPubHash := circuit.SELECT(spendFlag, spendPubkey, returnPubkey)
	calcAuthSpendHash := mimc.Hash(&circuit, targetPubHash, spendValue, noteRandom)
	circuit.MUSTBE_EQ(authSpendHash, calcAuthSpendHash)

	//通过merkle tree保证noteHash存在，即便return,auth都是null也是存在的，则可以不经过授权即可消费
	//preImage=hash(spendPubkey, returnPubkey,AuthPubkey,spendValue,noteRandom)
	noteHash := circuit.SECRET_INPUT("noteHash")
	// specify note hash constraint
	preImage := mimc.Hash(&circuit, spendPubkey, returnPubkey, authPubkey, spendValue, noteRandom)
	circuit.MUSTBE_EQ(noteHash, mimc.Hash(&circuit, preImage))

	merkelPathPart(&circuit, mimc, noteHash)

	r1cs := circuit.ToR1CS()

	return r1cs
}

func VerifyMerkleProof(circuit *frontend.CS, h mimc.MiMCGadget, merkleRoot *frontend.Constraint, proofSet, helper, valid []*frontend.Constraint) {

	sum := leafSum(circuit, h, proofSet[0])

	for i := 1; i < len(proofSet); i++ {
		circuit.MUSTBE_BOOLEAN(helper[i-1])
		d1 := circuit.SELECT(helper[i-1], sum, proofSet[i])
		d2 := circuit.SELECT(helper[i-1], proofSet[i], sum)
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
