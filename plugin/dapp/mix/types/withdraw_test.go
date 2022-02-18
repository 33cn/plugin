package types

import (
	"testing"

	"github.com/consensys/gnark/test"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
)

func TestWithdraw(t *testing.T) {
	circuitAssert := test.NewAssert(t)
	var withdrawCircuit WithdrawCircuit

	// compiles our circuit into a R1CS
	//r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &withdrawCircuit)
	//assert.NoError(err)
	{
		withdrawCircuit.TreeRootHash.Assign("457812157273975068180144939194931372467682914013265626991402231230450012330")
		withdrawCircuit.AuthorizeSpendHash.Assign("14463129595522277797353018005538222902035087589748809554960616199173731919802")
		withdrawCircuit.NullifierHash.Assign("12376093571606701949533526735186436482268907783512509935977783346861805262929")
		withdrawCircuit.Amount.Assign("28242048")

		withdrawCircuit.ReceiverPubKey.Assign("20094753906906836700810108535649927887994772258248603565615394844515069419451")
		withdrawCircuit.ReturnPubKey.Assign("10193030166569398670555398535278072963719579248877156082361830729347727033510")
		withdrawCircuit.AuthorizePubKey.Assign("2302306531516619173363925550130201424458047172090558749779153607734711372580")
		withdrawCircuit.NoteRandom.Assign("2824204835")
		withdrawCircuit.SpendPriKey.Assign("10190477835300927557649934238820360529458681672073866116232821892325659279502")
		withdrawCircuit.SpendFlag.Assign("1")
		withdrawCircuit.AuthorizeFlag.Assign("1")
		withdrawCircuit.NoteHash.Assign("1933334234871933218683301093524793045543211425994253628606123874146452475778")

		withdrawCircuit.Path0.Assign("19561523370160677851616596032513161448778901506614020103852017946679781620105")
		withdrawCircuit.Path1.Assign("13898857070666440684265042188056372750257678232709763835292910585848522658637")
		withdrawCircuit.Path2.Assign("15019169196974879571470243100379529757970866395477207575033769902587972032431")
		withdrawCircuit.Path3.Assign("0")
		withdrawCircuit.Path4.Assign("0")
		withdrawCircuit.Path5.Assign("0")
		withdrawCircuit.Path6.Assign("0")
		withdrawCircuit.Path7.Assign("0")
		withdrawCircuit.Path8.Assign("0")
		withdrawCircuit.Path9.Assign("0")

		withdrawCircuit.Helper0.Assign("1")
		withdrawCircuit.Helper1.Assign("1")
		withdrawCircuit.Helper2.Assign("1")
		withdrawCircuit.Helper3.Assign("0")
		withdrawCircuit.Helper4.Assign("0")
		withdrawCircuit.Helper5.Assign("0")
		withdrawCircuit.Helper6.Assign("0")
		withdrawCircuit.Helper7.Assign("0")
		withdrawCircuit.Helper8.Assign("0")
		withdrawCircuit.Helper9.Assign("0")

		withdrawCircuit.Valid0.Assign("1")
		withdrawCircuit.Valid1.Assign("1")
		withdrawCircuit.Valid2.Assign("1")
		withdrawCircuit.Valid3.Assign("0")
		withdrawCircuit.Valid4.Assign("0")
		withdrawCircuit.Valid5.Assign("0")
		withdrawCircuit.Valid6.Assign("0")
		withdrawCircuit.Valid7.Assign("0")
		withdrawCircuit.Valid8.Assign("0")
		withdrawCircuit.Valid9.Assign("0")

		var circuit WithdrawCircuit
		circuitAssert.ProverSucceeded(&circuit, &withdrawCircuit,
			test.WithCurves(ecc.BN254), test.WithCompileOpts(), test.WithBackends(backend.GROTH16))

	}

}
