package main

import (
	util "github.com/33cn/plugin/plugin/dapp/mix/cmd/gnark/circuit"
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
	amount
	receiverPubKey
	returnPubKey
	authorizePriKey
	spendFlag
	noteRandom
	noteHash

	path...
	helper...
	valid...
*/
func NewAuth() *frontend.R1CS {

	// create root constraint system
	circuit := frontend.New()

	amount := circuit.SECRET_INPUT("amount")

	//spend pubkey
	receiverPubKey := circuit.SECRET_INPUT("receiverPubKey")
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
	targetPubHash := circuit.SELECT(spendFlag, receiverPubKey, returnPubKey)
	calcAuthSpendHash := mimc.Hash(&circuit, targetPubHash, amount, noteRandom)
	circuit.MUSTBE_EQ(authSpendHash, calcAuthSpendHash)

	//通过merkle tree保证noteHash存在，即便return,auth都是null也是存在的，则可以不经过授权即可消费
	// specify note hash constraint
	preImage := mimc.Hash(&circuit, receiverPubKey, returnPubKey, authPubKey, amount, noteRandom)
	noteHash := circuit.SECRET_INPUT("noteHash")
	circuit.MUSTBE_EQ(noteHash, preImage)

	util.MerkelPathPart(&circuit, mimc, preImage)

	r1cs := circuit.ToR1CS()

	return r1cs
}
