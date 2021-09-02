package types

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
)

func TestAuthorize(t *testing.T) {
	assert := groth16.NewAssert(t)

	var authCircuit AuthorizeCircuit
	// compiles our circuit into a R1CS
	r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &authCircuit)
	assert.NoError(err)
	{
		authCircuit.TreeRootHash.Assign("950328190378475063285997697131233976268556642407874368795731039491163033995")
		authCircuit.AuthorizeHash.Assign("12118078255438699776281693699296572905522325280239057706039401491884511470990")
		authCircuit.AuthorizeSpendHash.Assign("21866258877426223880121052705448065394371888667902748431050285218933372701264")

		authCircuit.ReceiverPubKey.Assign("13496572805321444273664325641440458311310163934354047265362731297880627774936")
		authCircuit.ReturnPubKey.Assign("10193030166569398670555398535278072963719579248877156082361830729347727033510")
		authCircuit.AuthorizePubKey.Assign("2302306531516619173363925550130201424458047172090558749779153607734711372580")
		authCircuit.AuthorizePriKey.Assign("17822967620457187568904804290291537271142779717280482398091401115827760898835")
		authCircuit.NoteRandom.Assign("2824204835")
		authCircuit.Amount.Assign("28242048")
		authCircuit.SpendFlag.Assign("1")
		authCircuit.NoteHash.Assign("11183619348394875496624033204802036013086293645689330234403504655205992608466")

		authCircuit.Path0.Assign("19561523370160677851616596032513161448778901506614020103852017946679781620105")
		authCircuit.Path1.Assign("13898857070666440684265042188056372750257678232709763835292910585848522658637")
		authCircuit.Path2.Assign("15019169196974879571470243100379529757970866395477207575033769902587972032431")
		authCircuit.Path3.Assign("0")
		authCircuit.Path4.Assign("0")
		authCircuit.Path5.Assign("0")
		authCircuit.Path6.Assign("0")
		authCircuit.Path7.Assign("0")
		authCircuit.Path8.Assign("0")
		authCircuit.Path9.Assign("0")

		authCircuit.Helper0.Assign("1")
		authCircuit.Helper1.Assign("1")
		authCircuit.Helper2.Assign("1")
		authCircuit.Helper3.Assign("0")
		authCircuit.Helper4.Assign("0")
		authCircuit.Helper5.Assign("0")
		authCircuit.Helper6.Assign("0")
		authCircuit.Helper7.Assign("0")
		authCircuit.Helper8.Assign("0")
		authCircuit.Helper9.Assign("0")

		authCircuit.Valid0.Assign("1")
		authCircuit.Valid1.Assign("1")
		authCircuit.Valid2.Assign("1")
		authCircuit.Valid3.Assign("0")
		authCircuit.Valid4.Assign("0")
		authCircuit.Valid5.Assign("0")
		authCircuit.Valid6.Assign("0")
		authCircuit.Valid7.Assign("0")
		authCircuit.Valid8.Assign("0")
		authCircuit.Valid9.Assign("0")

		assert.ProverSucceeded(r1cs, &authCircuit)

	}

}
