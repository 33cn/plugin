package wallet

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"

	"github.com/33cn/chain33/common"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/stretchr/testify/assert"
)

//func TestGetCommitValue(t *testing.T) {
//	var out, change, minFee, sum uint64
//	var inputs []uint64
//	inputs = []uint64{100, 80}
//	out = 60
//	minFee = 1
//	for _, i := range inputs {
//		sum += i
//	}
//	change = sum - out - minFee
//	_, err := getShieldValue(inputs, out, change, minFee)
//	assert.Nil(t, err)
//
//	a := "0a9c010a4d3136323433323838333039363632323833373538343930323239313730303834393836343035373630373234353332323934333436353837323033353436363930353333373131303333323139124b3238383637383239373931373237373235343930333236303134303538313534363138303135353433383231393339363836333632313634323236303434353739313434393237383237331a82033078656663333331616261616139653039353966636536356163343364626534306364646139356534356261636163613161326166626265366637323533633132326233346264323337353932343066306237623836653363343635666131343065666332636665623861653035366234323163303665353062396532646564636236383963336536656435363636373731343235663736313931653831356665666633646432393965633535386261323731343238333131623130353364376265633864646163313733393632326238666138326438373336666531623332633835376438343330643634646637336530643265326238373932396335633762366437336534383365363130303561313361376531643730636637653834656132613235343166373235363834656266613737653235313232326466313039336230313964646165623963376134393763316538653737386462313730323636323536666666363332643437363738626633366634383361373334346666326330"
//	da, err := hex.DecodeString(a)
//	assert.Nil(t, err)
//
//	var data mixTy.DHSecret
//
//	err = types.Decode(da, &data)
//	assert.Nil(t, err)
//	//fmt.Println("data", data)
//
//	var deposit mixTy.DepositProofResp
//	deposit.NoteHash = "notehashstr"
//	deposit.Proof = &mixTy.SecretData{
//		ReceiverKey: "receiverstr",
//		ReturnKey:   "returnval",
//	}
//	deposit.Secrets = &mixTy.DHSecretGroup{
//		Receiver:  "receiverstr",
//		Authorize: "authval",
//	}
//	ty := reflect.TypeOf(deposit)
//	val := reflect.ValueOf(deposit)
//	n := ty.NumField()
//	for i := 0; i < n; i++ {
//
//		fmt.Println("i=", i, "name=", ty.Field(i).Name, "valid", val.Field(i).IsZero(), "name", val.Field(i), "ty", val.FieldByName(ty.Field(i).Name))
//	}
//
//	//type strA struct{
//	//	a backend.Assignment `secret:"public"`
//	//	b backend.Assignment `secret:"private"`
//	//}
//	//tt := strA{}
//	//
//	//tt.a.Value.SetString("123",10)
//	//tt.a.IsPublic
//	//tt.a.Value.SetString("567",10)
//	//tp := reflect.TypeOf(tt)
//	//fmt.Println("tt",tp.Field(0).Tag.Get("secret"))
//}

//
//func TestGetAssignments(t *testing.T) {
//	deposit := DepositInput{
//		NoteHash:        "111",
//		Amount:          "222",
//		ReceiverPubKey:  "333",
//		ReturnPubKey:    "444",
//		AuthorizePubKey: "555",
//		NoteRandom:      "666",
//	}
//	assigns, err := getAssignments(deposit)
//	assert.Nil(t, err)
//	val := assigns["NoteHash"].Value
//
//	assert.Equal(t, val.String(), deposit.NoteHash)
//	assert.Equal(t, assigns["NoteHash"].IsPublic, true)
//	assert.Equal(t, assigns["ReceiverPubKey"].IsPublic, false)
//
//	reduceAssign := assigns.DiscardSecrets()
//	_, ok := reduceAssign["ReceiverPubKey"]
//	assert.Equal(t, ok, false)
//
//	tv := reflect.ValueOf(&deposit)
//	tv.Elem().FieldByName("NoteHash").SetString("999")
//	//tv.FieldByName("NoteHash").Elem().SetString("999")
//	assert.Equal(t, "999", deposit.NoteHash)
//	//var in WithdrawInput
//	//initTreePath(&in)
//	//assert.Equal(t, "99", in.Path1)
//	printObj(deposit)
//	rst, err := json.MarshalIndent(deposit, "", "    ")
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Println(string(rst))
//}
//
//func TestVerifyProof(t *testing.T) {
//	deposit := DepositInput{
//		NoteHash:        "319044369386253980478484545601022272388174242630360020319556034291986094405",
//		Amount:          "500000000",
//		ReceiverPubKey:  "7244551457692363731356498279463138379576484998878425864678733206990733443457",
//		ReturnPubKey:    "0",
//		AuthorizePubKey: "0",
//		NoteRandom:      "21887946084880143097415438560893808581456164284155969619878297484093938793578",
//	}
//	assigns, err := getAssignments(deposit)
//
//	//从电路文件获取电路约束
//	circuit, err := getCircuit("../cmd/gnark/circuit/deposit/circuit_deposit.r1cs")
//	assert.Nil(t, err)
//
//	//从pv 文件读取Pk结构
//	pk, err := getProveKey("../cmd/gnark/circuit/deposit/circuit_deposit.pk")
//	assert.Nil(t, err)
//
//	proof, err := createProof(circuit, pk, assigns)
//	assert.Nil(t, err)
//
//	vk, err := getVerifyKey("../cmd/gnark/circuit/deposit/circuit_deposit.vk")
//	assert.Nil(t, err)
//	rst := verifyProof(proof, vk, assigns.DiscardSecrets())
//	assert.Equal(t, true, rst)
//
//	proofKey, err := serializeObj(proof)
//	assert.Nil(t, err)
//
//	verifyKey, err := serializeObj(vk)
//	assert.Nil(t, err)
//
//	proofInput, err := serialInputs(assigns)
//	//fmt.Println("proofinput",proofInput)
//	assert.Nil(t, err)
//	rt, err := zksnark.Verify(verifyKey, proofKey, proofInput)
//	assert.Nil(t, err)
//	assert.Equal(t, true, rt)
//
//}

//func TestGetZkProofKeys(t *testing.T) {
//	var depositCircuit mixTy.DepositCircuit
//	depositCircuit.NoteHash.Assign("11183619348394875496624033204802036013086293645689330234403504655205992608466")
//	depositCircuit.Amount.Assign(28242048)
//	depositCircuit.ReceiverPubKey.Assign("13496572805321444273664325641440458311310163934354047265362731297880627774936")
//	depositCircuit.ReturnPubKey.Assign("10193030166569398670555398535278072963719579248877156082361830729347727033510")
//	depositCircuit.AuthorizePubKey.Assign("2302306531516619173363925550130201424458047172090558749779153607734711372580")
//	depositCircuit.NoteRandom.Assign(2824204835)
//
//	pkFile := "../cmd/gnark/circuit_deposit.pk"
//	ret, err := getZkProofKeys(mixTy.VerifyType_DEPOSIT, pkFile, &depositCircuit, 0)
//
//	assert.Nil(t, err)
//	fmt.Println("rst", ret.PublicInput)
//}

func TestUpdateTreePath(t *testing.T) {
	var proof mixTy.TreePathProof
	proof.TreeRootHash = common.ToHex(common.Sha2Sum([]byte{1}))
	path0 := "11183619348394875496624033204802036013086293645689330234403504655205992608466"
	path1 := "13496572805321444273664325641440458311310163934354047265362731297880627774936"
	proof.TreePath = append(proof.TreePath, path0)
	proof.TreePath = append(proof.TreePath, path1)
	proof.Helpers = append(proof.Helpers, []uint32{1, 2}...)

	var input mixTy.AuthorizeCircuit
	updateTreePath(&input, &proof)
	ret0 := input.Path0.GetWitnessValue(ecc.BN254)
	ret1 := input.Path1.GetWitnessValue(ecc.BN254)

	assert.Equal(t, path0, ret0.String())
	assert.Equal(t, path1, ret1.String())

	path2 := "0"
	ret2 := input.Path2.GetWitnessValue(ecc.BN254)
	assert.Equal(t, path2, ret2.String())

}
