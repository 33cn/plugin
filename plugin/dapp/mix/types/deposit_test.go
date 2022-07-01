package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/consensys/gnark/backend/witness"
	"github.com/stretchr/testify/assert"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
)

//receiver prikey="10190477835300927557649934238820360529458681672073866116232821892325659279502"
//receiver pubkey="13496572805321444273664325641440458311310163934354047265362731297880627774936"

//return prikey="7969140283216448215269095418467361784159407896899334866715345504515077887397"
//return pubkey="10193030166569398670555398535278072963719579248877156082361830729347727033510"

//authorize prikey="17822967620457187568904804290291537271142779717280482398091401115827760898835"
//authorize pubkey="2302306531516619173363925550130201424458047172090558749779153607734711372580"

//spend prikey="10407830929890509544473717262275616077696950294748419792758056545898949331744"
//spend pubkey="3656010751855274388516368747583374746848682779395325737100877017850943546836"

func TestDeposit(t *testing.T) {

	circuitAssert := test.NewAssert(t)

	var depositCircuit DepositCircuit

	// compiles our circuit into a R1CS
	r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &depositCircuit)
	assert.Nil(t, err)

	{
		//var witness Deposit
		depositCircuit.NoteHash.Assign("14803109164298493466684583242985432968056297173621710679077236816845588688436")
		depositCircuit.Amount.Assign(28242048)
		depositCircuit.ReceiverPubKey.Assign("13496572805321444273664325641440458311310163934354047265362731297880627774936")
		depositCircuit.ReturnPubKey.Assign("10193030166569398670555398535278072963719579248877156082361830729347727033510")
		depositCircuit.AuthorizePubKey.Assign("2302306531516619173363925550130201424458047172090558749779153607734711372580")
		depositCircuit.NoteRandom.Assign("2824204835")
		//assert.ProverSucceeded(r1cs, &depositCircuit)

		var circuit DepositCircuit
		circuitAssert.ProverSucceeded(&circuit, &depositCircuit,
			test.WithCurves(ecc.BN254), test.WithCompileOpts(), test.WithBackends(backend.GROTH16))

	}
	var pubBuf bytes.Buffer
	witness.WritePublicTo(&pubBuf, ecc.BN254, &depositCircuit)
	//fmt.Println("buf",hex.EncodeToString(pubBuf.Bytes()))
	pubStr := hex.EncodeToString(pubBuf.Bytes())

	var buf bytes.Buffer
	pk, vk, err := groth16.Setup(r1cs)
	assert.Nil(t, err)
	buf.Reset()
	vk.WriteTo(&buf)
	//fmt.Println("vk",hex.EncodeToString(buf.Bytes()))
	vkStr := hex.EncodeToString(buf.Bytes())

	proof, err := groth16.Prove(r1cs, pk, &depositCircuit)
	assert.Nil(t, err)
	buf.Reset()
	proof.WriteTo(&buf)
	//fmt.Println("proof",hex.EncodeToString(buf.Bytes()))
	pfStr := hex.EncodeToString(buf.Bytes())

	d, _ := hex.DecodeString(vkStr)
	vkt := groth16.NewVerifyingKey(ecc.BN254)
	buf.Reset()
	buf.Write(d)
	vkt.ReadFrom(&buf)

	d, _ = hex.DecodeString(pfStr)
	buf.Reset()
	buf.Write(d)
	prt := groth16.NewProof(ecc.BN254)
	prt.ReadFrom(&buf)

	d, _ = hex.DecodeString(pubStr)
	buf.Reset()
	buf.Write(d)

	err = groth16.ReadAndVerify(prt, vkt, &buf)
	assert.Nil(t, err)
}

func TestDepositSetVal(t *testing.T) {

	pubInput := "0000000218b9b448bd793dab9075f70ce404a87497b3a2d0d0b5d177441d6cada1e9b2d20000000000000000000000000000000000000000000000000000000001aef080"
	str, _ := hex.DecodeString(pubInput)
	var buf bytes.Buffer
	buf.Write(str)
	var val Witness
	n, err := val.LimitReadFrom(&buf)
	assert.Nil(t, err)
	fmt.Println("n=", n)

	//val=make([]fr.Element,2)
	//val[0].SetInterface(0x18b9b448bd793dab9075f70ce404a87497b3a2d0d0b5d177441d6cada1e9b2d2)
	//val[0].set
	//val[1].SetInterface(0x0000000000000000000000000000000000000000000000000000000001aef080)
	fmt.Println("0=", val[0].String(), "1=", val[1].String())

	var depositCircuit DepositCircuit
	getVal(&depositCircuit, val)
	fmt.Println("deposit", depositCircuit.NoteHash.GetWitnessValue(ecc.BN254))
	fmt.Println("amount", depositCircuit.Amount.GetWitnessValue(ecc.BN254))

}

func getVal(input frontend.Circuit, w Witness) {
	tValue := reflect.ValueOf(input)
	if tValue.Kind() == reflect.Ptr {
		tValue = tValue.Elem()
	}
	//reft := reflect.TypeOf(input)
	fmt.Println("fields", tValue.NumField())
	for i, v := range w {
		field := tValue.Type().Field((i))
		fmt.Println("i=", i, "name=", field.Name)
		f := tValue.FieldByName(field.Name)
		a := f.Addr().Interface().(*frontend.Variable)
		//a:=tValue.Field(i).Interface().(frontend.Variable)
		a.Assign(v.String())
	}
}

//UT 测试读取pk,vk文件验证
//
//func TestDepositFile(t *testing.T) {
//	var depositCircuit DepositCircuit
//
//	// compiles our circuit into a R1CS
//	r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &depositCircuit)
//	assert.Nil(t, err)
//
//	depositCircuit.NoteHash.Assign("14803109164298493466684583242985432968056297173621710679077236816845588688436")
//	depositCircuit.Amount.Assign(28242048)
//	depositCircuit.ReceiverPubKey.Assign("13496572805321444273664325641440458311310163934354047265362731297880627774936")
//	depositCircuit.ReturnPubKey.Assign("10193030166569398670555398535278072963719579248877156082361830729347727033510")
//	depositCircuit.AuthorizePubKey.Assign("2302306531516619173363925550130201424458047172090558749779153607734711372580")
//	depositCircuit.NoteRandom.Assign(2824204835)
//
//
//	pfStr,pubStr,vkStr := getZkProofKeys(t,r1cs,".","circuit_deposit",&depositCircuit)
//	assert.Nil(t, err)
//
//
//	bufVk, err := GetByteBuff(vkStr)
//	assert.Nil(t, err)
//	vkt := groth16.NewVerifyingKey(ecc.BN254)
//	_,err = vkt.ReadFrom(bufVk)
//	assert.Nil(t, err)
//
//	bufPk, err := GetByteBuff(pfStr)
//	assert.Nil(t, err)
//	prt := groth16.NewProof(ecc.BN254)
//	prt.ReadFrom(bufPk)
//
//	bufPub, err := GetByteBuff(pubStr)
//	assert.Nil(t, err)
//
//	err = groth16.ReadAndVerify(prt, vkt, bufPub)
//	assert.Nil(t,err)
//}
//
//
//func getZkProofKeys(t *testing.T,r1cs frontend.CompiledConstraintSystem, path, file string, inputs frontend.Circuit) (string,string,string) {
//	//从pv 文件读取Pk结构
//	pkFile := filepath.Join(path, file+".pk")
//	pkBuf, err := readZkKeyFile(pkFile)
//	assert.Nil(t, err)
//
//	pk := groth16.NewProvingKey(ecc.BN254)
//	_, err = pk.ReadFrom(pkBuf)
//	assert.Nil(t, err)
//
//	//vk
//	vkFile := filepath.Join(path, file+".vk")
//	vkBuf, err := readZkKeyFile(vkFile)
//	assert.Nil(t, err)
//	vkStr := hex.EncodeToString(vkBuf.Bytes())
//	vk := groth16.NewVerifyingKey(ecc.BN254)
//	_,err = vk.ReadFrom(vkBuf)
//	assert.Nil(t, err)
//
//	//产生zk 证明
//	proof, err := groth16.Prove(r1cs, pk, inputs)
//	assert.Nil(t, err)
//
//	err = groth16.Verify(proof,vk,inputs)
//	assert.Nil(t, err)
//
//
//	var proofKey bytes.Buffer
//	_, err = proof.WriteTo(&proofKey)
//	assert.Nil(t, err)
//	proofStr := hex.EncodeToString(proofKey.Bytes())
//	//公开输入序列化
//	var pubBuf bytes.Buffer
//	_, err = witness.WritePublicTo(&pubBuf, ecc.BN254, inputs)
//	assert.Nil(t, err)
//	pubStr := hex.EncodeToString(pubBuf.Bytes())
//
//	err = groth16.ReadAndVerify(proof,vk,&pubBuf)
//	assert.Nil(t, err)
//
//	return proofStr,pubStr,vkStr
//
//}
//
////文件内容存储的是hex string，读的时候直接转换为string即可
//func readZkKeyFile(path string) (*bytes.Buffer, error) {
//	f, err := os.Open(path)
//	if err != nil {
//		return nil, errors.Wrapf(err, "open file=%s", path)
//	}
//	var buf bytes.Buffer
//	_, err = buf.ReadFrom(f)
//	if err != nil {
//		return nil, errors.Wrapf(err, "read file=%s", path)
//	}
//
//	bytes, err := GetByteBuff(buf.String())
//	if err != nil {
//		return nil, err
//	}
//	return bytes,nil
//}
