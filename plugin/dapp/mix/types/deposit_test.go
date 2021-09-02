package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/consensys/gnark/backend/witness"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
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
	assert := groth16.NewAssert(t)

	var depositCircuit DepositCircuit

	// compiles our circuit into a R1CS
	r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &depositCircuit)
	assert.NoError(err)

	{
		//var witness Deposit
		depositCircuit.NoteHash.Assign("11183619348394875496624033204802036013086293645689330234403504655205992608466")
		depositCircuit.Amount.Assign(28242048)
		depositCircuit.ReceiverPubKey.Assign("13496572805321444273664325641440458311310163934354047265362731297880627774936")
		depositCircuit.ReturnPubKey.Assign("10193030166569398670555398535278072963719579248877156082361830729347727033510")
		depositCircuit.AuthorizePubKey.Assign("2302306531516619173363925550130201424458047172090558749779153607734711372580")
		depositCircuit.NoteRandom.Assign(2824204835)
		assert.ProverSucceeded(r1cs, &depositCircuit)
	}
	var pubBuf bytes.Buffer
	witness.WritePublicTo(&pubBuf, ecc.BN254, &depositCircuit)
	//fmt.Println("buf",hex.EncodeToString(pubBuf.Bytes()))
	pubStr := hex.EncodeToString(pubBuf.Bytes())

	var buf bytes.Buffer
	pk, vk, err := groth16.Setup(r1cs)
	assert.Nil(err)
	buf.Reset()
	vk.WriteTo(&buf)
	//fmt.Println("vk",hex.EncodeToString(buf.Bytes()))
	vkStr := hex.EncodeToString(buf.Bytes())

	proof, err := groth16.Prove(r1cs, pk, &depositCircuit)
	assert.Nil(err)
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
	assert.Nil(err)
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
	fmt.Println("deposit", frontend.GetAssignedValue(depositCircuit.NoteHash))
	fmt.Println("amount", frontend.GetAssignedValue(depositCircuit.Amount))

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
