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
		good.Assign(backend.Public, "TreeRootHash", "10531321614990797034921282585661869614556487056951485265320464926630499341310")
		good.Assign(backend.Public, "AuthorizeSpendHash", "14468512365438613046028281588661351435476168610934165547900473609197783547663")
		good.Assign(backend.Public, "NullifierHash", "6747518781649068310795677405858353007442326529625450860668944156162052335195")
		good.Assign(backend.Public, "Amount", "28242048")

		good.Assign(backend.Secret, "ReceiverPubKey", "13735985067536865723202617343666111332145536963656464451727087263423649028705")
		good.Assign(backend.Secret, "ReturnPubKey", "16067249407809359746114321133992130903102335882983385972747813693681808870497")
		good.Assign(backend.Secret, "AuthorizePubKey", "13519883267141251871527102103999205179714486518503885909948192364772977661583")

		good.Assign(backend.Secret, "SpendPriKey", "10190477835300927557649934238820360529458681672073866116232821892325659279502")
		good.Assign(backend.Secret, "SpendFlag", "1")
		good.Assign(backend.Secret, "AuthorizeFlag", "1")

		good.Assign(backend.Secret, "NoteRandom", "2824204835")

		good.Assign(backend.Secret, "NoteHash", "16308793397024662832064523892418908145900866571524124093537199035808550255649")

		//nodehash="16308793397024662832064523892418908145900866571524124093537199035808550255649"
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

func TestWithdraw2(t *testing.T) {

	assert := groth16.NewAssert(t)

	r1cs := NewWithdraw()
	r1csBN256 := backend_bn256.Cast(r1cs)
	{
		good := backend.NewAssignment()
		good.Assign(backend.Public, "TreeRootHash", "7407373673604276152801354004851383461539317977307945198624806503955302548700")
		good.Assign(backend.Public, "AuthorizeSpendHash", "0")
		good.Assign(backend.Public, "NullifierHash", "3911774040567972872956008387141175001419649692949203140089059098956773329188")
		good.Assign(backend.Public, "Amount", "500000000")

		good.Assign(backend.Secret, "ReceiverPubKey", "7244551457692363731356498279463138379576484998878425864678733206990733443457")
		good.Assign(backend.Secret, "ReturnPubKey", "0")
		good.Assign(backend.Secret, "AuthorizePubKey", "0")

		good.Assign(backend.Secret, "SpendPriKey", "19115616183616714814727844928908633989028519974595353009754871398745087846141")
		good.Assign(backend.Secret, "SpendFlag", "1")
		good.Assign(backend.Secret, "AuthorizeFlag", "0")

		good.Assign(backend.Secret, "NoteRandom", "13093524699504167542220418896875211339267114119238016501132859435646426190390")

		good.Assign(backend.Secret, "NoteHash", "6441981280327245191543878922302235060888201984279829017462503030039593719666")

		//nodehash="16308793397024662832064523892418908145900866571524124093537199035808550255649"
		good.Assign(backend.Secret, "Path0", "16592081253758169453601069427813166612800474003570537799829885686701266956141")
		good.Assign(backend.Secret, "Path1", "0")
		good.Assign(backend.Secret, "Path2", "0")
		good.Assign(backend.Secret, "Path3", "0")
		good.Assign(backend.Secret, "Path4", "0")
		good.Assign(backend.Secret, "Path5", "0")
		good.Assign(backend.Secret, "Path6", "0")
		good.Assign(backend.Secret, "Path7", "0")
		good.Assign(backend.Secret, "Path8", "0")
		good.Assign(backend.Secret, "Path9", "0")

		good.Assign(backend.Secret, "Helper0", "0")
		good.Assign(backend.Secret, "Helper1", "0")
		good.Assign(backend.Secret, "Helper2", "0")
		good.Assign(backend.Secret, "Helper3", "0")
		good.Assign(backend.Secret, "Helper4", "0")
		good.Assign(backend.Secret, "Helper5", "0")
		good.Assign(backend.Secret, "Helper6", "0")
		good.Assign(backend.Secret, "Helper7", "0")
		good.Assign(backend.Secret, "Helper8", "0")
		good.Assign(backend.Secret, "Helper9", "0")

		good.Assign(backend.Secret, "Valid0", "1")
		good.Assign(backend.Secret, "Valid1", "0")
		good.Assign(backend.Secret, "Valid2", "0")
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
