package main

import (
	util "github.com/33cn/plugin/plugin/dapp/mix/cmd/gnark/circuit"
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/gadgets/hash/mimc"
	"github.com/consensys/gurvy"
)

func main() {
	circuit := NewWithdraw()
	gob.Write("circuit_withdraw.r1cs", circuit, gurvy.BN256)
}

//withdraw commit hash the circuit implementing
/*
public:
	treeRootHash
	authorizeSpendHash
	nullifierHash
	amount

private:
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
func NewWithdraw() *frontend.R1CS {

	// create root constraint system
	circuit := frontend.New()

	spendValue := circuit.PUBLIC_INPUT("amount")

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
	// specify auth hash constraint
	calcAuthHash := mimc.Hash(&circuit, targetPubHash, spendValue, noteRandom)
	targetAuthHash := circuit.SELECT(authFlag, calcAuthHash, nullValue)
	circuit.MUSTBE_EQ(authHash, targetAuthHash)

	//通过merkle tree保证noteHash存在，即便return,auth都是null也是存在的，则可以不经过授权即可消费
	//preImage=hash(spendPubkey, returnPubkey,AuthPubkey,spendValue,noteRandom)
	noteHash := circuit.SECRET_INPUT("noteHash")
	calcReturnPubkey := circuit.SELECT(authFlag, returnPubkey, nullValue)
	calcAuthPubkey := circuit.SELECT(authFlag, authPubkey, nullValue)
	// specify note hash constraint
	preImage := mimc.Hash(&circuit, spendPubkey, calcReturnPubkey, calcAuthPubkey, spendValue, noteRandom)
	circuit.MUSTBE_EQ(noteHash, preImage)

	util.MerkelPathPart(&circuit, mimc, preImage)

	r1cs := circuit.ToR1CS()

	return r1cs
}
