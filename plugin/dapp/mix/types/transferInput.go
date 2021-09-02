package types

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

type TransferInputCircuit struct {
	TreeRootHash       frontend.Variable `gnark:",public"`
	AuthorizeSpendHash frontend.Variable `gnark:",public"`
	NullifierHash      frontend.Variable `gnark:",public"`
	ShieldAmountX      frontend.Variable `gnark:",public"`
	ShieldAmountY      frontend.Variable `gnark:",public"`

	//secret
	ReceiverPubKey  frontend.Variable
	ReturnPubKey    frontend.Variable
	AuthorizePubKey frontend.Variable
	NoteRandom      frontend.Variable

	Amount        frontend.Variable
	AmountRandom  frontend.Variable
	SpendPriKey   frontend.Variable
	SpendFlag     frontend.Variable
	AuthorizeFlag frontend.Variable
	NoteHash      frontend.Variable

	//tree path info
	Path0 frontend.Variable
	Path1 frontend.Variable
	Path2 frontend.Variable
	Path3 frontend.Variable
	Path4 frontend.Variable
	Path5 frontend.Variable
	Path6 frontend.Variable
	Path7 frontend.Variable
	Path8 frontend.Variable
	Path9 frontend.Variable

	Helper0 frontend.Variable
	Helper1 frontend.Variable
	Helper2 frontend.Variable
	Helper3 frontend.Variable
	Helper4 frontend.Variable
	Helper5 frontend.Variable
	Helper6 frontend.Variable
	Helper7 frontend.Variable
	Helper8 frontend.Variable
	Helper9 frontend.Variable

	Valid0 frontend.Variable
	Valid1 frontend.Variable
	Valid2 frontend.Variable
	Valid3 frontend.Variable
	Valid4 frontend.Variable
	Valid5 frontend.Variable
	Valid6 frontend.Variable
	Valid7 frontend.Variable
	Valid8 frontend.Variable
	Valid9 frontend.Variable
}

// Define declares the circuit's constraints
func (circuit *TransferInputCircuit) Define(curveID ecc.ID, cs *frontend.ConstraintSystem) error {
	cs.AssertIsBoolean(circuit.SpendFlag)
	cs.AssertIsBoolean(circuit.AuthorizeFlag)

	// hash function
	h, _ := mimc.NewMiMC("seed", curveID, cs)
	mimc := &h

	//verify spend private key
	targetSpendKey := cs.Select(circuit.SpendFlag, circuit.ReceiverPubKey, circuit.ReturnPubKey)
	mimc.Write(circuit.SpendPriKey)
	cs.AssertIsEqual(targetSpendKey, mimc.Sum())

	nullValue := cs.Constant(0)
	mimc.Reset()
	mimc.Write(targetSpendKey, circuit.Amount, circuit.NoteRandom)
	calcAuthSpendHash := mimc.Sum()
	targetAuthSpendHash := cs.Select(circuit.AuthorizeFlag, calcAuthSpendHash, nullValue)
	cs.AssertIsEqual(circuit.AuthorizeSpendHash, targetAuthSpendHash)

	mimc.Reset()
	mimc.Write(circuit.NoteRandom)
	cs.AssertIsEqual(circuit.NullifierHash, mimc.Sum())
	//
	calcReturnPubkey := cs.Select(circuit.AuthorizeFlag, circuit.ReturnPubKey, nullValue)
	calcAuthPubkey := cs.Select(circuit.AuthorizeFlag, circuit.AuthorizePubKey, nullValue)
	mimc.Reset()
	mimc.Write(circuit.ReceiverPubKey, calcReturnPubkey, calcAuthPubkey, circuit.Amount, circuit.NoteRandom)
	cs.AssertIsEqual(circuit.NoteHash, mimc.Sum())

	var proofSet, helper, valid []frontend.Variable
	proofSet = append(proofSet, circuit.NoteHash)
	proofSet = append(proofSet, circuit.Path0)
	proofSet = append(proofSet, circuit.Path1)
	proofSet = append(proofSet, circuit.Path2)
	proofSet = append(proofSet, circuit.Path3)
	proofSet = append(proofSet, circuit.Path4)
	proofSet = append(proofSet, circuit.Path5)
	proofSet = append(proofSet, circuit.Path6)
	proofSet = append(proofSet, circuit.Path7)
	proofSet = append(proofSet, circuit.Path8)
	proofSet = append(proofSet, circuit.Path9)

	//helper[0],valid[0]占位， 方便接口只设置有效值
	helper = append(helper, cs.Constant("1"))
	helper = append(helper, circuit.Helper0)
	helper = append(helper, circuit.Helper1)
	helper = append(helper, circuit.Helper2)
	helper = append(helper, circuit.Helper3)
	helper = append(helper, circuit.Helper4)
	helper = append(helper, circuit.Helper5)
	helper = append(helper, circuit.Helper6)
	helper = append(helper, circuit.Helper7)
	helper = append(helper, circuit.Helper8)
	helper = append(helper, circuit.Helper9)

	valid = append(valid, cs.Constant("1"))
	valid = append(valid, circuit.Valid0)
	valid = append(valid, circuit.Valid1)
	valid = append(valid, circuit.Valid2)
	valid = append(valid, circuit.Valid3)
	valid = append(valid, circuit.Valid4)
	valid = append(valid, circuit.Valid5)
	valid = append(valid, circuit.Valid6)
	valid = append(valid, circuit.Valid7)
	valid = append(valid, circuit.Valid8)
	valid = append(valid, circuit.Valid9)

	CommitValueVerify(cs, circuit.Amount, circuit.AmountRandom, circuit.ShieldAmountX, circuit.ShieldAmountY)
	VerifyMerkleProof(cs, mimc, circuit.TreeRootHash, proofSet, helper, valid)

	return nil
}
