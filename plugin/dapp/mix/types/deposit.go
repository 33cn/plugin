package types

import (
	"github.com/consensys/gnark/frontend"
	"github.com/33cn/plugin/plugin/crypto/legacymimc"
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

func (circuit *DepositCircuit) Define(api frontend.API) error {
	// hash function
	mimc, _ := legacymimc.NewCircuitMiMC(api, MimcHashSeed)

	mimc.Write(circuit.ReceiverPubKey, circuit.ReturnPubKey, circuit.AuthorizePubKey, circuit.Amount, circuit.NoteRandom)
	api.AssertIsEqual(circuit.NoteHash, mimc.Sum())

	return nil
}
