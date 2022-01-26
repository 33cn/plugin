package types

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

//spend commit hash the circuit implementing
type DepositCircuit struct {
	NoteHash frontend.Variable `gnark:",public"`
	Amount   frontend.Variable `gnark:",public"`

	ReceiverPubKey  frontend.Variable
	ReturnPubKey    frontend.Variable
	AuthorizePubKey frontend.Variable
	NoteRandom      frontend.Variable
}

func (circuit *DepositCircuit) Define(curveID ecc.ID, api frontend.API) error {
	// hash function
	mimc, _ := mimc.NewMiMC(MimcHashSeed, curveID, api)

	mimc.Write(circuit.ReceiverPubKey, circuit.ReturnPubKey, circuit.AuthorizePubKey, circuit.Amount, circuit.NoteRandom)
	api.AssertIsEqual(circuit.NoteHash, mimc.Sum())

	return nil
}
