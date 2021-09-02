package types

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
)

func TestTransferOutput(t *testing.T) {

	assert := groth16.NewAssert(t)
	var outCircuit TransferOutputCircuit

	// compiles our circuit into a R1CS
	r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &outCircuit)
	assert.NoError(err)

	{

		outCircuit.NoteHash.Assign("11183619348394875496624033204802036013086293645689330234403504655205992608466")
		outCircuit.ShieldAmountX.Assign("2999198834503527181782558341022909853195739283744640133924786234819945005771")
		outCircuit.ShieldAmountY.Assign("19443413539487113257436159186910517766382570615508121086985490610335878889881")

		outCircuit.ReceiverPubKey.Assign("13496572805321444273664325641440458311310163934354047265362731297880627774936")
		outCircuit.ReturnPubKey.Assign("10193030166569398670555398535278072963719579248877156082361830729347727033510")
		outCircuit.AuthorizePubKey.Assign("2302306531516619173363925550130201424458047172090558749779153607734711372580")
		outCircuit.NoteRandom.Assign("2824204835")
		outCircuit.Amount.Assign("28242048")
		outCircuit.AmountRandom.Assign("282420481")
		assert.ProverSucceeded(r1cs, &outCircuit)

	}

}
