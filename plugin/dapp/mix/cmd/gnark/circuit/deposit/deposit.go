package main

import (
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/gadgets/hash/mimc"
	"github.com/consensys/gurvy"
)

func main() {
	circuit := NewDeposit()
	gob.Write("circuit_deposit.r1cs", circuit, gurvy.BN256)
}

//spend commit hash the circuit implementing
/*
public:
	NoteHash
	Amount

private:
	ReceiverPubKey
	ReturnPubKey
	AuthorizePubKey
	NoteRandom

*/
func NewDeposit() *frontend.R1CS {

	// create root constraint system
	circuit := frontend.New()

	//公共输入以验证
	amount := circuit.PUBLIC_INPUT("Amount")

	//spend pubkey
	receiverPubKey := circuit.SECRET_INPUT("ReceiverPubKey")
	returnPubkey := circuit.SECRET_INPUT("ReturnPubKey")
	authPubkey := circuit.SECRET_INPUT("AuthorizePubKey")

	// hash function
	mimc, _ := mimc.NewMiMCGadget("seed", gurvy.BN256)

	//note hash random
	noteRandom := circuit.SECRET_INPUT("NoteRandom")

	//通过merkle tree保证noteHash存在，即便return,auth都是null也是存在的，则可以不经过授权即可消费
	//preImage=hash(spendPubkey, returnPubkey,AuthPubkey,spendValue,noteRandom)
	noteHash := circuit.PUBLIC_INPUT("NoteHash")
	// specify note hash constraint
	preImage := mimc.Hash(&circuit, receiverPubKey, returnPubkey, authPubkey, amount, noteRandom)
	circuit.MUSTBE_EQ(noteHash, preImage)

	r1cs := circuit.ToR1CS()

	return r1cs
}
