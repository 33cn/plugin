package types

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

type TransferOutputCircuit struct {
	//public
	NoteHash      frontend.Variable `gnark:",public"`
	ShieldAmountX frontend.Variable `gnark:",public"`
	ShieldAmountY frontend.Variable `gnark:",public"`
	ShieldPointHX frontend.Variable `gnark:",public"`
	ShieldPointHY frontend.Variable `gnark:",public"`

	//secret
	ReceiverPubKey  frontend.Variable
	ReturnPubKey    frontend.Variable
	AuthorizePubKey frontend.Variable
	NoteRandom      frontend.Variable
	Amount          frontend.Variable
	AmountRandom    frontend.Variable
}

// Define declares the circuit's constraints
func (circuit *TransferOutputCircuit) Define(curveID ecc.ID, cs *frontend.ConstraintSystem) error {
	// hash function
	h, _ := mimc.NewMiMC(MimcHashSeed, curveID, cs)
	mimc := &h
	mimc.Write(circuit.ReceiverPubKey, circuit.ReturnPubKey, circuit.AuthorizePubKey, circuit.Amount, circuit.NoteRandom)
	cs.AssertIsEqual(circuit.NoteHash, mimc.Sum())

	CommitValueVerify(cs, circuit.Amount, circuit.AmountRandom, circuit.ShieldAmountX, circuit.ShieldAmountY, circuit.ShieldPointHX, circuit.ShieldPointHY)

	return nil
}
