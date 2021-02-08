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
	authorizePubKey
	authorizeHash(=hash(authpubkey+noterandom))
	authorizeSpendHash(=hash(spendpub+value+noterandom))

private:
	amount
	receiverPubKey
	returnPubKey
	authorizePriKey
	spendFlag
	noteRandom

	path...
	helper...
	valid...
*/
func TestAuthorizeSpend(t *testing.T) {

	assert := groth16.NewAssert(t)

	r1cs := NewAuth()
	r1csBN256 := backend_bn256.Cast(r1cs)
	{
		good := backend.NewAssignment()
		good.Assign(backend.Public, "TreeRootHash", "10531321614990797034921282585661869614556487056951485265320464926630499341310")
		good.Assign(backend.Public, "AuthorizePubKey", "13519883267141251871527102103999205179714486518503885909948192364772977661583")
		good.Assign(backend.Public, "AuthorizeHash", "1267825436937766239630340333349685320927256968591056373125946583184548355070")
		good.Assign(backend.Public, "AuthorizeSpendHash", "14468512365438613046028281588661351435476168610934165547900473609197783547663")

		good.Assign(backend.Secret, "Amount", "28242048")
		good.Assign(backend.Secret, "ReceiverPubKey", "13735985067536865723202617343666111332145536963656464451727087263423649028705")
		good.Assign(backend.Secret, "ReturnPubKey", "16067249407809359746114321133992130903102335882983385972747813693681808870497")
		good.Assign(backend.Secret, "AuthorizePriKey", "17822967620457187568904804290291537271142779717280482398091401115827760898835")
		good.Assign(backend.Secret, "SpendFlag", "1")
		good.Assign(backend.Secret, "NoteRandom", "2824204835")
		good.Assign(backend.Secret, "NoteHash", "16308793397024662832064523892418908145900866571524124093537199035808550255649")

		good.Assign(backend.Secret, "Path0", "19561523370160677851616596032513161448778901506614020103852017946679781620105")
		good.Assign(backend.Secret, "Path1", "13898857070666440684265042188056372750257678232709763835292910585848522658637")
		good.Assign(backend.Secret, "Path2", "15019169196974879571470243100379529757970866395477207575033769902587972032431")
		good.Assign(backend.Secret, "Path3", "0")
		good.Assign(backend.Secret, "Path4", "0")
		good.Assign(backend.Secret, "Path5", "0")
		good.Assign(backend.Secret, "Path6", "0")
		good.Assign(backend.Secret, "Path7", "0")
		good.Assign(backend.Secret, "Path8", "0")
		good.Assign(backend.Secret, "Path9", "0")

		good.Assign(backend.Secret, "Helper0", "1")
		good.Assign(backend.Secret, "Helper1", "1")
		good.Assign(backend.Secret, "Helper2", "1")
		good.Assign(backend.Secret, "Helper3", "0")
		good.Assign(backend.Secret, "Helper4", "0")
		good.Assign(backend.Secret, "Helper5", "0")
		good.Assign(backend.Secret, "Helper6", "0")
		good.Assign(backend.Secret, "Helper7", "0")
		good.Assign(backend.Secret, "Helper8", "0")
		good.Assign(backend.Secret, "Helper9", "0")

		good.Assign(backend.Secret, "Valid0", "1")
		good.Assign(backend.Secret, "Valid1", "1")
		good.Assign(backend.Secret, "Valid2", "1")
		good.Assign(backend.Secret, "Valid3", "0")
		good.Assign(backend.Secret, "Valid4", "0")
		good.Assign(backend.Secret, "Valid5", "0")
		good.Assign(backend.Secret, "Valid6", "0")
		good.Assign(backend.Secret, "Valid7", "0")
		good.Assign(backend.Secret, "Valid8", "0")
		good.Assign(backend.Secret, "Valid9", "0")

		assert.Solved(&r1csBN256, good, nil)
	}

}
