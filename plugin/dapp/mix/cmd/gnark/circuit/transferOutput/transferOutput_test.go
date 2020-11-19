package main

import (
	"testing"

	backend_bn256 "github.com/consensys/gnark/backend/bn256"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/bn256/groth16"
)

/*
public:
	commitValueX
	commitValueY
	nodeHash

private:
	spendAmount
	spendRandom
	spendPubKey
	returnPubKey
	authorizePubKey
	noteRandom

*/
func TestTransferOutput(t *testing.T) {

	assert := groth16.NewAssert(t)

	r1cs := NewTransferOutput()
	r1csBN256 := backend_bn256.Cast(r1cs)
	{
		good := backend.NewAssignment()
		good.Assign(backend.Public, "commitValueX", "14087975867275911077371231345227824611951436822132762463787130558957838320348")
		good.Assign(backend.Public, "commitValueY", "15113519960384204624879642069520481336224311978035289236693658603675385299879")
		good.Assign(backend.Public, "nodeHash", "16308793397024662832064523892418908145900866571524124093537199035808550255649")

		good.Assign(backend.Secret, "spendAmount", "28242048")
		good.Assign(backend.Secret, "spendRandom", "35")

		good.Assign(backend.Secret, "spendPubKey", "13735985067536865723202617343666111332145536963656464451727087263423649028705")
		good.Assign(backend.Secret, "returnPubKey", "16067249407809359746114321133992130903102335882983385972747813693681808870497")
		good.Assign(backend.Secret, "authorizePubKey", "13519883267141251871527102103999205179714486518503885909948192364772977661583")

		good.Assign(backend.Secret, "noteRandom", "2824204835")

		assert.Solved(&r1csBN256, good, nil)
	}

}
