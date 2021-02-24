package output

import (
	"testing"

	backend_bn256 "github.com/consensys/gnark/backend/bn256"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/bn256/groth16"
)

/*
public:
	shieldAmountX
	shieldAmountY
	noteHash

private:
	amount
	amountRandom
	receiverPubKey
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
		good.Assign(backend.Public, "shieldAmountX", "14087975867275911077371231345227824611951436822132762463787130558957838320348")
		good.Assign(backend.Public, "shieldAmountY", "15113519960384204624879642069520481336224311978035289236693658603675385299879")
		good.Assign(backend.Public, "noteHash", "16308793397024662832064523892418908145900866571524124093537199035808550255649")

		good.Assign(backend.Secret, "amount", "28242048")
		good.Assign(backend.Secret, "amountRandom", "35")

		good.Assign(backend.Secret, "receiverPubKey", "13735985067536865723202617343666111332145536963656464451727087263423649028705")
		good.Assign(backend.Secret, "returnPubKey", "16067249407809359746114321133992130903102335882983385972747813693681808870497")
		good.Assign(backend.Secret, "authorizePubKey", "13519883267141251871527102103999205179714486518503885909948192364772977661583")

		good.Assign(backend.Secret, "noteRandom", "2824204835")

		assert.Solved(&r1csBN256, good, nil)
	}

}

func TestTransferOutputTemp(t *testing.T) {

	assert := groth16.NewAssert(t)

	r1cs := NewTransferOutput()
	r1csBN256 := backend_bn256.Cast(r1cs)
	{
		good := backend.NewAssignment()
		good.Assign(backend.Public, "shieldAmountX", "3403754862862244121869403834818720211897208891381676574399662060838495940078")
		good.Assign(backend.Public, "shieldAmountY", "21401902064308935591303802598664246616585474010691469717860664156067228175223")
		good.Assign(backend.Public, "noteHash", "13610259753835165822431524149670478281864477297016371975012414049080268826331")

		good.Assign(backend.Secret, "amount", "300000000")
		good.Assign(backend.Secret, "amountRandom", "17199160520698273243343882915453578587")

		good.Assign(backend.Secret, "receiverPubKey", "18829345085195922012068709111582461121107908772422825655963168999800303848486")
		good.Assign(backend.Secret, "returnPubKey", "0")
		good.Assign(backend.Secret, "authorizePubKey", "0")

		good.Assign(backend.Secret, "noteRandom", "5029847585956946251661044349066579681630691396824473307862642244158835326399")

		assert.Solved(&r1csBN256, good, nil)
	}

}

func TestTransferOutputChange(t *testing.T) {

	assert := groth16.NewAssert(t)

	r1cs := NewTransferOutput()
	r1csBN256 := backend_bn256.Cast(r1cs)
	{
		good := backend.NewAssignment()
		good.Assign(backend.Public, "shieldAmountX", "10703086269439751873106176219875739041896146845566831131812760688039385779519")
		good.Assign(backend.Public, "shieldAmountY", "19139103177181062461420753508628290808191900352948606822559796252948653071734")
		good.Assign(backend.Public, "noteHash", "13134546856103113099750783399130805737503059294172727906371169345876474249458")

		good.Assign(backend.Secret, "amount", "199900000")
		good.Assign(backend.Secret, "amountRandom", "86450085302571105354912213444290224646")

		good.Assign(backend.Secret, "receiverPubKey", "7244551457692363731356498279463138379576484998878425864678733206990733443457")
		good.Assign(backend.Secret, "returnPubKey", "0")
		good.Assign(backend.Secret, "authorizePubKey", "0")

		good.Assign(backend.Secret, "noteRandom", "7266395330102686861165120582739238575545854195882356283931287331463151808870")

		assert.Solved(&r1csBN256, good, nil)
	}

}
