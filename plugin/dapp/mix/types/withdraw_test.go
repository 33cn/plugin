package types

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
)

func TestWithdraw(t *testing.T) {
	assert := groth16.NewAssert(t)
	var withdrawCircuit WithdrawCircuit

	// compiles our circuit into a R1CS
	r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &withdrawCircuit)
	assert.NoError(err)
	{
		withdrawCircuit.TreeRootHash.Assign("950328190378475063285997697131233976268556642407874368795731039491163033995")
		withdrawCircuit.AuthorizeSpendHash.Assign("21866258877426223880121052705448065394371888667902748431050285218933372701264")
		withdrawCircuit.NullifierHash.Assign("18261754976334473090934939020486888794395514077667802499672726421629833403191")
		withdrawCircuit.Amount.Assign("28242048")

		withdrawCircuit.ReceiverPubKey.Assign("13496572805321444273664325641440458311310163934354047265362731297880627774936")
		withdrawCircuit.ReturnPubKey.Assign("10193030166569398670555398535278072963719579248877156082361830729347727033510")
		withdrawCircuit.AuthorizePubKey.Assign("2302306531516619173363925550130201424458047172090558749779153607734711372580")
		withdrawCircuit.NoteRandom.Assign("2824204835")
		withdrawCircuit.SpendPriKey.Assign("10190477835300927557649934238820360529458681672073866116232821892325659279502")
		withdrawCircuit.SpendFlag.Assign("1")
		withdrawCircuit.AuthorizeFlag.Assign("1")
		withdrawCircuit.NoteHash.Assign("11183619348394875496624033204802036013086293645689330234403504655205992608466")

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

		assert.ProverSucceeded(r1cs, &withdrawCircuit)
	}

}
