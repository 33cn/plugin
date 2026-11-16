//go:build !386

// gnark v0.9.0 constraint/bn254/solver.go does unaligned 64-bit atomic on 386,
// causing "panic: unaligned 64-bit atomic operation". Skip this test on 386.

package types

import (
	"testing"

	"github.com/consensys/gnark/test"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
)

func TestAuthorize(t *testing.T) {
	assert := test.NewAssert(t)

	var authCircuit AuthorizeCircuit
	// compiles our circuit into a R1CS
	//r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &authCircuit)
	//assert.NoError(err)
	{
		authCircuit.TreeRootHash = "18953560960857123326054550555759265877143310030168748002053709716397549796490"
		authCircuit.AuthorizeHash = "4895770928816523282558547614022568289586238930922185617307655942541278140196"
		authCircuit.AuthorizeSpendHash = "17847836824302447823607018011193117302314262324241905063439417486141908449945"

		authCircuit.ReceiverPubKey = "13496572805321444273664325641440458311310163934354047265362731297880627774936"
		authCircuit.ReturnPubKey = "10193030166569398670555398535278072963719579248877156082361830729347727033510"
		authCircuit.AuthorizePubKey = "21375443884718346645287794853944958188610587325651351394209548420684867245331"
		authCircuit.AuthorizePriKey = "17822967620457187568904804290291537271142779717280482398091401115827760898835"
		authCircuit.NoteRandom = "2824204835"
		authCircuit.Amount = "28242048"
		authCircuit.SpendFlag = "1"
		authCircuit.NoteHash = "4641322019922509455032097629889269851124503217947103069347447050214760728147"

		authCircuit.Path0 = "19561523370160677851616596032513161448778901506614020103852017946679781620105"
		authCircuit.Path1 = "13898857070666440684265042188056372750257678232709763835292910585848522658637"
		authCircuit.Path2 = "15019169196974879571470243100379529757970866395477207575033769902587972032431"
		authCircuit.Path3 = "0"
		authCircuit.Path4 = "0"
		authCircuit.Path5 = "0"
		authCircuit.Path6 = "0"
		authCircuit.Path7 = "0"
		authCircuit.Path8 = "0"
		authCircuit.Path9 = "0"

		authCircuit.Helper0 = "1"
		authCircuit.Helper1 = "1"
		authCircuit.Helper2 = "1"
		authCircuit.Helper3 = "0"
		authCircuit.Helper4 = "0"
		authCircuit.Helper5 = "0"
		authCircuit.Helper6 = "0"
		authCircuit.Helper7 = "0"
		authCircuit.Helper8 = "0"
		authCircuit.Helper9 = "0"

		authCircuit.Valid0 = "1"
		authCircuit.Valid1 = "1"
		authCircuit.Valid2 = "1"
		authCircuit.Valid3 = "0"
		authCircuit.Valid4 = "0"
		authCircuit.Valid5 = "0"
		authCircuit.Valid6 = "0"
		authCircuit.Valid7 = "0"
		authCircuit.Valid8 = "0"
		authCircuit.Valid9 = "0"

		var circuit AuthorizeCircuit
		assert.ProverSucceeded(&circuit, &authCircuit,
			test.WithCurves(ecc.BN254), test.WithCompileOpts(), test.WithBackends(backend.GROTH16))

	}

}
