package types

import (
	"testing"

	"github.com/consensys/gnark/test"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
)

func TestTransferInput(t *testing.T) {
	circuitAssert := test.NewAssert(t)

	var inputCircuit TransferInputCircuit
	// compiles our circuit into a R1CS
	//r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &inputCircuit)
	//assert.Nil(t, err)

	{
		inputCircuit.TreeRootHash.Assign("457812157273975068180144939194931372467682914013265626991402231230450012330")
		inputCircuit.AuthorizeSpendHash.Assign("14463129595522277797353018005538222902035087589748809554960616199173731919802")
		inputCircuit.NullifierHash.Assign("12376093571606701949533526735186436482268907783512509935977783346861805262929")
		inputCircuit.ShieldAmountX.Assign("12598656472198560295956115825363858683566688303969048230275808317634686855820")
		inputCircuit.ShieldAmountY.Assign("5287524325952639485224317845546845679649328720392059741208352845659048630229")
		inputCircuit.ShieldPointHX.Assign("19172955941344617222923168298456110557655645809646772800021167670156933290312")
		inputCircuit.ShieldPointHY.Assign("21116962883761739586121793871108889864627195706475546685847911817475098399811")

		inputCircuit.ReceiverPubKey.Assign("20094753906906836700810108535649927887994772258248603565615394844515069419451")
		inputCircuit.ReturnPubKey.Assign("10193030166569398670555398535278072963719579248877156082361830729347727033510")
		inputCircuit.AuthorizePubKey.Assign("2302306531516619173363925550130201424458047172090558749779153607734711372580")
		inputCircuit.NoteRandom.Assign("2824204835")
		inputCircuit.Amount.Assign("28242048")
		inputCircuit.AmountRandom.Assign("282420481")
		inputCircuit.SpendPriKey.Assign("10190477835300927557649934238820360529458681672073866116232821892325659279502")
		inputCircuit.SpendFlag.Assign("1")
		inputCircuit.AuthorizeFlag.Assign("1")
		inputCircuit.NoteHash.Assign("1933334234871933218683301093524793045543211425994253628606123874146452475778")

		inputCircuit.Path0.Assign("19561523370160677851616596032513161448778901506614020103852017946679781620105")
		inputCircuit.Path1.Assign("13898857070666440684265042188056372750257678232709763835292910585848522658637")
		inputCircuit.Path2.Assign("15019169196974879571470243100379529757970866395477207575033769902587972032431")
		inputCircuit.Path3.Assign("0")
		inputCircuit.Path4.Assign("0")
		inputCircuit.Path5.Assign("0")
		inputCircuit.Path6.Assign("0")
		inputCircuit.Path7.Assign("0")
		inputCircuit.Path8.Assign("0")
		inputCircuit.Path9.Assign("0")

		inputCircuit.Helper0.Assign("1")
		inputCircuit.Helper1.Assign("1")
		inputCircuit.Helper2.Assign("1")
		inputCircuit.Helper3.Assign("0")
		inputCircuit.Helper4.Assign("0")
		inputCircuit.Helper5.Assign("0")
		inputCircuit.Helper6.Assign("0")
		inputCircuit.Helper7.Assign("0")
		inputCircuit.Helper8.Assign("0")
		inputCircuit.Helper9.Assign("0")

		inputCircuit.Valid0.Assign("1")
		inputCircuit.Valid1.Assign("1")
		inputCircuit.Valid2.Assign("1")
		inputCircuit.Valid3.Assign("0")
		inputCircuit.Valid4.Assign("0")
		inputCircuit.Valid5.Assign("0")
		inputCircuit.Valid6.Assign("0")
		inputCircuit.Valid7.Assign("0")
		inputCircuit.Valid8.Assign("0")
		inputCircuit.Valid9.Assign("0")

		var circuit TransferInputCircuit
		circuitAssert.ProverSucceeded(&circuit, &inputCircuit,
			test.WithCurves(ecc.BN254), test.WithCompileOpts(), test.WithBackends(backend.GROTH16))

	}

}

//
//func TestTransferInputReturnKey(t *testing.T) {
//
//	assert := groth16.NewAssert(t)
//
//	r1cs := NewTransferInput()
//	r1csBN256 := backend_bn256.Cast(r1cs)
//	{
//		good := backend.NewAssignment()
//		good.Assign(backend.Public, "treeRootHash", "10531321614990797034921282585661869614556487056951485265320464926630499341310")
//		good.Assign(backend.Public, "shieldAmountX", "14087975867275911077371231345227824611951436822132762463787130558957838320348")
//		good.Assign(backend.Public, "shieldAmountY", "15113519960384204624879642069520481336224311978035289236693658603675385299879")
//		good.Assign(backend.Public, "authorizeSpendHash", "6026163592877030954825395224309219861774131411806846860652261047183070579370")
//		good.Assign(backend.Public, "nullifierHash", "6747518781649068310795677405858353007442326529625450860668944156162052335195")
//
//		good.Assign(backend.Secret, "amount", "28242048")
//		good.Assign(backend.Secret, "amountRandom", "35")
//
//		good.Assign(backend.Secret, "receiverPubKey", "13735985067536865723202617343666111332145536963656464451727087263423649028705")
//		good.Assign(backend.Secret, "returnPubKey", "16067249407809359746114321133992130903102335882983385972747813693681808870497")
//		good.Assign(backend.Secret, "authorizePubKey", "13519883267141251871527102103999205179714486518503885909948192364772977661583")
//
//		good.Assign(backend.Secret, "spendPriKey", "7969140283216448215269095418467361784159407896899334866715345504515077887397")
//		//returnkey spend notehash
//		good.Assign(backend.Secret, "spendFlag", "0")
//
//		good.Assign(backend.Secret, "authorizeFlag", "1")
//
//		good.Assign(backend.Secret, "noteRandom", "2824204835")
//
//		good.Assign(backend.Secret, "noteHash", "16308793397024662832064523892418908145900866571524124093537199035808550255649")
//
//		//nodehash="16308793397024662832064523892418908145900866571524124093537199035808550255649"
//		good.Assign(backend.Secret, "path1", "19561523370160677851616596032513161448778901506614020103852017946679781620105")
//		good.Assign(backend.Secret, "path2", "13898857070666440684265042188056372750257678232709763835292910585848522658637")
//		good.Assign(backend.Secret, "path3", "15019169196974879571470243100379529757970866395477207575033769902587972032431")
//		good.Assign(backend.Secret, "path4", "0")
//		good.Assign(backend.Secret, "path5", "0")
//		good.Assign(backend.Secret, "path6", "0")
//		good.Assign(backend.Secret, "path7", "0")
//		good.Assign(backend.Secret, "path8", "0")
//		good.Assign(backend.Secret, "path9", "0")
//
//		good.Assign(backend.Secret, "helper1", "1")
//		good.Assign(backend.Secret, "helper2", "1")
//		good.Assign(backend.Secret, "helper3", "1")
//		good.Assign(backend.Secret, "helper4", "0")
//		good.Assign(backend.Secret, "helper5", "0")
//		good.Assign(backend.Secret, "helper6", "0")
//		good.Assign(backend.Secret, "helper7", "0")
//		good.Assign(backend.Secret, "helper8", "0")
//		good.Assign(backend.Secret, "helper9", "0")
//
//		good.Assign(backend.Secret, "valid1", "1")
//		good.Assign(backend.Secret, "valid2", "1")
//		good.Assign(backend.Secret, "valid3", "1")
//		good.Assign(backend.Secret, "valid4", "0")
//		good.Assign(backend.Secret, "valid5", "0")
//		good.Assign(backend.Secret, "valid6", "0")
//		good.Assign(backend.Secret, "valid7", "0")
//		good.Assign(backend.Secret, "valid8", "0")
//		good.Assign(backend.Secret, "valid9", "0")
//
//		assert.Solved(&r1csBN256, good, nil)
//	}
//
//}
//
//func TestTransferInputNoAuthorize(t *testing.T) {
//
//	assert := groth16.NewAssert(t)
//
//	r1cs := NewTransferInput()
//	r1csBN256 := backend_bn256.Cast(r1cs)
//	{
//		good := backend.NewAssignment()
//		good.Assign(backend.Public, "treeRootHash", "8924377726623516198388981994706612588174229761660626844219523809311621081152")
//		good.Assign(backend.Public, "shieldAmountX", "20026900249169569699397829614948056401416692452575929785554743563301443795984")
//		good.Assign(backend.Public, "shieldAmountY", "11443294504840468048882645872852838384649876010412151915870299030068051779303")
//		good.Assign(backend.Public, "authorizeSpendHash", "0")
//		good.Assign(backend.Public, "nullifierHash", "4493238794492517147695618716694376637191823831910850819304582851540887491471")
//
//		good.Assign(backend.Secret, "amount", "500000000")
//		good.Assign(backend.Secret, "amountRandom", "103649245823269378598256096359743803233")
//
//		good.Assign(backend.Secret, "receiverPubKey", "7244551457692363731356498279463138379576484998878425864678733206990733443457")
//		good.Assign(backend.Secret, "returnPubKey", "0")
//		good.Assign(backend.Secret, "authorizePubKey", "0")
//
//		good.Assign(backend.Secret, "spendPriKey", "19115616183616714814727844928908633989028519974595353009754871398745087846141")
//		good.Assign(backend.Secret, "spendFlag", "1")
//		//not need authorize
//		good.Assign(backend.Secret, "authorizeFlag", "0")
//
//		good.Assign(backend.Secret, "noteRandom", "16855817802811010832998322637530013398737002960466904173163094025121554818471")
//
//		good.Assign(backend.Secret, "noteHash", "4757455985754753449547885621755931629265767091930770913671501411452663313694")
//
//		good.Assign(backend.Secret, "path1", "21609869341494920403470153054548069228540665950349313465330160010270609674984")
//		good.Assign(backend.Secret, "path2", "0")
//		good.Assign(backend.Secret, "path3", "0")
//		good.Assign(backend.Secret, "path4", "0")
//		good.Assign(backend.Secret, "path5", "0")
//		good.Assign(backend.Secret, "path6", "0")
//		good.Assign(backend.Secret, "path7", "0")
//		good.Assign(backend.Secret, "path8", "0")
//		good.Assign(backend.Secret, "path9", "0")
//
//		good.Assign(backend.Secret, "helper1", "0")
//		good.Assign(backend.Secret, "helper2", "1")
//		good.Assign(backend.Secret, "helper3", "1")
//		good.Assign(backend.Secret, "helper4", "0")
//		good.Assign(backend.Secret, "helper5", "0")
//		good.Assign(backend.Secret, "helper6", "0")
//		good.Assign(backend.Secret, "helper7", "0")
//		good.Assign(backend.Secret, "helper8", "0")
//		good.Assign(backend.Secret, "helper9", "0")
//
//		good.Assign(backend.Secret, "valid1", "1")
//		good.Assign(backend.Secret, "valid2", "0")
//		good.Assign(backend.Secret, "valid3", "0")
//		good.Assign(backend.Secret, "valid4", "0")
//		good.Assign(backend.Secret, "valid5", "0")
//		good.Assign(backend.Secret, "valid6", "0")
//		good.Assign(backend.Secret, "valid7", "0")
//		good.Assign(backend.Secret, "valid8", "0")
//		good.Assign(backend.Secret, "valid9", "0")
//
//		assert.Solved(&r1csBN256, good, nil)
//	}
//
//}
