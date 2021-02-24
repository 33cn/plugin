package output

import (
	util "github.com/33cn/plugin/plugin/dapp/mix/cmd/gnark/circuit"
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/gadgets/hash/mimc"
	"github.com/consensys/gurvy"
)

func main() {
	circuit := NewTransferOutput()
	gob.Write("circuit_transfer_output.r1cs", circuit, gurvy.BN256)
}

//spend commit hash the circuit implementing
/*
public:
	shieldAmountX
	shieldAmountY
	noteHash

private:
	amount
	amountRandom
	receiverPubKey
	returnPubKey
	authorizePubKey
	noteRandom

*/
func NewTransferOutput() *frontend.R1CS {

	// create root constraint system
	circuit := frontend.New()

	spendValue := circuit.SECRET_INPUT("Amount")

	//spend pubkey
	spendPubkey := circuit.SECRET_INPUT("ReceiverPubKey")
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
	preImage := mimc.Hash(&circuit, spendPubkey, returnPubkey, authPubkey, spendValue, noteRandom)
	circuit.MUSTBE_EQ(noteHash, preImage)

	util.CommitValuePart(&circuit, spendValue)

	r1cs := circuit.ToR1CS()

	return r1cs
}
