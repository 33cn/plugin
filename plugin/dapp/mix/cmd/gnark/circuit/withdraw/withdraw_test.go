package main

import (
	"testing"

	backend_bn256 "github.com/consensys/gnark/backend/bn256"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/bn256/groth16"
)

/*
public:
	treeRootHash
	authorizeHash
	nullifierHash
	amount

private:
	receiverPubKey
	returnPubKey
	authorizePubKey
	spendPriKey
	spendFlag
	authorizeFlag
	noteRandom

	path...
	helper...
	valid...
*/
func TestWithdraw(t *testing.T) {

	assert := groth16.NewAssert(t)

	r1cs := NewWithdraw()
	r1csBN256 := backend_bn256.Cast(r1cs)
	{
		good := backend.NewAssignment()
		good.Assign(backend.Public, "treeRootHash", "10531321614990797034921282585661869614556487056951485265320464926630499341310")
		good.Assign(backend.Public, "authorizeSpendHash", "14468512365438613046028281588661351435476168610934165547900473609197783547663")
		good.Assign(backend.Public, "nullifierHash", "6747518781649068310795677405858353007442326529625450860668944156162052335195")
		good.Assign(backend.Public, "amount", "28242048")

		good.Assign(backend.Secret, "receiverPubKey", "13735985067536865723202617343666111332145536963656464451727087263423649028705")
		good.Assign(backend.Secret, "returnPubKey", "16067249407809359746114321133992130903102335882983385972747813693681808870497")
		good.Assign(backend.Secret, "authorizePubKey", "13519883267141251871527102103999205179714486518503885909948192364772977661583")

		good.Assign(backend.Secret, "spendPriKey", "10190477835300927557649934238820360529458681672073866116232821892325659279502")
		good.Assign(backend.Secret, "spendFlag", "1")
		good.Assign(backend.Secret, "authorizeFlag", "1")

		good.Assign(backend.Secret, "noteRandom", "2824204835")

		good.Assign(backend.Secret, "noteHash", "16308793397024662832064523892418908145900866571524124093537199035808550255649")

		//nodehash="16308793397024662832064523892418908145900866571524124093537199035808550255649"
		good.Assign(backend.Secret, "path1", "19561523370160677851616596032513161448778901506614020103852017946679781620105")
		good.Assign(backend.Secret, "path2", "13898857070666440684265042188056372750257678232709763835292910585848522658637")
		good.Assign(backend.Secret, "path3", "15019169196974879571470243100379529757970866395477207575033769902587972032431")
		good.Assign(backend.Secret, "path4", "0")
		good.Assign(backend.Secret, "path5", "0")
		good.Assign(backend.Secret, "path6", "0")
		good.Assign(backend.Secret, "path7", "0")
		good.Assign(backend.Secret, "path8", "0")
		good.Assign(backend.Secret, "path9", "0")

		good.Assign(backend.Secret, "helper1", "1")
		good.Assign(backend.Secret, "helper2", "1")
		good.Assign(backend.Secret, "helper3", "1")
		good.Assign(backend.Secret, "helper4", "0")
		good.Assign(backend.Secret, "helper5", "0")
		good.Assign(backend.Secret, "helper6", "0")
		good.Assign(backend.Secret, "helper7", "0")
		good.Assign(backend.Secret, "helper8", "0")
		good.Assign(backend.Secret, "helper9", "0")

		good.Assign(backend.Secret, "valid1", "1")
		good.Assign(backend.Secret, "valid2", "1")
		good.Assign(backend.Secret, "valid3", "1")
		good.Assign(backend.Secret, "valid4", "0")
		good.Assign(backend.Secret, "valid5", "0")
		good.Assign(backend.Secret, "valid6", "0")
		good.Assign(backend.Secret, "valid7", "0")
		good.Assign(backend.Secret, "valid8", "0")
		good.Assign(backend.Secret, "valid9", "0")

		assert.Solved(&r1csBN256, good, nil)
	}

}
