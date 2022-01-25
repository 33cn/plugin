package types

import (
	"github.com/consensys/gnark/test"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
)

func TestTransferOutput(t *testing.T) {

	circuitAssert := test.NewAssert(t)
	var outCircuit TransferOutputCircuit

	// compiles our circuit into a R1CS
	//r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &outCircuit)
	//assert.NoError(err)

	{

		outCircuit.NoteHash.Assign("14803109164298493466684583242985432968056297173621710679077236816845588688436")
		outCircuit.ShieldAmountX.Assign("12598656472198560295956115825363858683566688303969048230275808317634686855820")
		outCircuit.ShieldAmountY.Assign("5287524325952639485224317845546845679649328720392059741208352845659048630229")
		outCircuit.ShieldPointHX.Assign("19172955941344617222923168298456110557655645809646772800021167670156933290312")
		outCircuit.ShieldPointHY.Assign("21116962883761739586121793871108889864627195706475546685847911817475098399811")

		outCircuit.ReceiverPubKey.Assign("13496572805321444273664325641440458311310163934354047265362731297880627774936")
		outCircuit.ReturnPubKey.Assign("10193030166569398670555398535278072963719579248877156082361830729347727033510")
		outCircuit.AuthorizePubKey.Assign("2302306531516619173363925550130201424458047172090558749779153607734711372580")
		outCircuit.NoteRandom.Assign("2824204835")
		outCircuit.Amount.Assign("28242048")
		outCircuit.AmountRandom.Assign("282420481")
		//assert.ProverSucceeded(r1cs, &outCircuit)

		var circuit TransferOutputCircuit
		circuitAssert.ProverSucceeded(&circuit, &outCircuit,
			test.WithCurves(ecc.BN254), test.WithCompileOpts(), test.WithBackends(backend.GROTH16))

	}

}
