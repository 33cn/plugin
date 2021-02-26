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
	shieldAmountX
	shieldAmountY
	authorizeHash
	nullifierHash

private:
	amount
	amountRandom
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
func TestTransferInputAuth(t *testing.T) {

	assert := groth16.NewAssert(t)

	r1cs := NewTransferInput()
	r1csBN256 := backend_bn256.Cast(r1cs)
	{
		good := backend.NewAssignment()
		good.Assign(backend.Public, "treeRootHash", "10531321614990797034921282585661869614556487056951485265320464926630499341310")
		good.Assign(backend.Public, "shieldAmountX", "14087975867275911077371231345227824611951436822132762463787130558957838320348")
		good.Assign(backend.Public, "shieldAmountY", "15113519960384204624879642069520481336224311978035289236693658603675385299879")
		good.Assign(backend.Public, "authorizeSpendHash", "14468512365438613046028281588661351435476168610934165547900473609197783547663")
		good.Assign(backend.Public, "nullifierHash", "6747518781649068310795677405858353007442326529625450860668944156162052335195")

		good.Assign(backend.Secret, "amount", "28242048")
		good.Assign(backend.Secret, "amountRandom", "35")

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

func TestTransferInputReturnKey(t *testing.T) {

	assert := groth16.NewAssert(t)

	r1cs := NewTransferInput()
	r1csBN256 := backend_bn256.Cast(r1cs)
	{
		good := backend.NewAssignment()
		good.Assign(backend.Public, "treeRootHash", "10531321614990797034921282585661869614556487056951485265320464926630499341310")
		good.Assign(backend.Public, "shieldAmountX", "14087975867275911077371231345227824611951436822132762463787130558957838320348")
		good.Assign(backend.Public, "shieldAmountY", "15113519960384204624879642069520481336224311978035289236693658603675385299879")
		good.Assign(backend.Public, "authorizeSpendHash", "6026163592877030954825395224309219861774131411806846860652261047183070579370")
		good.Assign(backend.Public, "nullifierHash", "6747518781649068310795677405858353007442326529625450860668944156162052335195")

		good.Assign(backend.Secret, "amount", "28242048")
		good.Assign(backend.Secret, "amountRandom", "35")

		good.Assign(backend.Secret, "receiverPubKey", "13735985067536865723202617343666111332145536963656464451727087263423649028705")
		good.Assign(backend.Secret, "returnPubKey", "16067249407809359746114321133992130903102335882983385972747813693681808870497")
		good.Assign(backend.Secret, "authorizePubKey", "13519883267141251871527102103999205179714486518503885909948192364772977661583")

		good.Assign(backend.Secret, "spendPriKey", "7969140283216448215269095418467361784159407896899334866715345504515077887397")
		//returnkey spend notehash
		good.Assign(backend.Secret, "spendFlag", "0")

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

func TestTransferInputNoAuthorize(t *testing.T) {

	assert := groth16.NewAssert(t)

	r1cs := NewTransferInput()
	r1csBN256 := backend_bn256.Cast(r1cs)
	{
		good := backend.NewAssignment()
		good.Assign(backend.Public, "treeRootHash", "8924377726623516198388981994706612588174229761660626844219523809311621081152")
		good.Assign(backend.Public, "shieldAmountX", "20026900249169569699397829614948056401416692452575929785554743563301443795984")
		good.Assign(backend.Public, "shieldAmountY", "11443294504840468048882645872852838384649876010412151915870299030068051779303")
		good.Assign(backend.Public, "authorizeSpendHash", "0")
		good.Assign(backend.Public, "nullifierHash", "4493238794492517147695618716694376637191823831910850819304582851540887491471")

		good.Assign(backend.Secret, "amount", "500000000")
		good.Assign(backend.Secret, "amountRandom", "103649245823269378598256096359743803233")

		good.Assign(backend.Secret, "receiverPubKey", "7244551457692363731356498279463138379576484998878425864678733206990733443457")
		good.Assign(backend.Secret, "returnPubKey", "0")
		good.Assign(backend.Secret, "authorizePubKey", "0")

		good.Assign(backend.Secret, "spendPriKey", "19115616183616714814727844928908633989028519974595353009754871398745087846141")
		good.Assign(backend.Secret, "spendFlag", "1")
		//not need authorize
		good.Assign(backend.Secret, "authorizeFlag", "0")

		good.Assign(backend.Secret, "noteRandom", "16855817802811010832998322637530013398737002960466904173163094025121554818471")

		good.Assign(backend.Secret, "noteHash", "4757455985754753449547885621755931629265767091930770913671501411452663313694")

		good.Assign(backend.Secret, "path1", "21609869341494920403470153054548069228540665950349313465330160010270609674984")
		good.Assign(backend.Secret, "path2", "0")
		good.Assign(backend.Secret, "path3", "0")
		good.Assign(backend.Secret, "path4", "0")
		good.Assign(backend.Secret, "path5", "0")
		good.Assign(backend.Secret, "path6", "0")
		good.Assign(backend.Secret, "path7", "0")
		good.Assign(backend.Secret, "path8", "0")
		good.Assign(backend.Secret, "path9", "0")

		good.Assign(backend.Secret, "helper1", "0")
		good.Assign(backend.Secret, "helper2", "1")
		good.Assign(backend.Secret, "helper3", "1")
		good.Assign(backend.Secret, "helper4", "0")
		good.Assign(backend.Secret, "helper5", "0")
		good.Assign(backend.Secret, "helper6", "0")
		good.Assign(backend.Secret, "helper7", "0")
		good.Assign(backend.Secret, "helper8", "0")
		good.Assign(backend.Secret, "helper9", "0")

		good.Assign(backend.Secret, "valid1", "1")
		good.Assign(backend.Secret, "valid2", "0")
		good.Assign(backend.Secret, "valid3", "0")
		good.Assign(backend.Secret, "valid4", "0")
		good.Assign(backend.Secret, "valid5", "0")
		good.Assign(backend.Secret, "valid6", "0")
		good.Assign(backend.Secret, "valid7", "0")
		good.Assign(backend.Secret, "valid8", "0")
		good.Assign(backend.Secret, "valid9", "0")

		assert.Solved(&r1csBN256, good, nil)
	}

}
