package main

import (
	"testing"

	backend_bn256 "github.com/consensys/gnark/backend/bn256"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/bn256/groth16"
)

/*
public:
	nodeHash
	amount

private:
	receiverPubKey
	returnPubkey
	authorizePubkey
	noteRandom

*/

func TestDeposit(t *testing.T) {

	assert := groth16.NewAssert(t)

	//spend prikey="10190477835300927557649934238820360529458681672073866116232821892325659279502"
	//spend pubkey="13735985067536865723202617343666111332145536963656464451727087263423649028705"

	//return prikey="7969140283216448215269095418467361784159407896899334866715345504515077887397"
	//return pubkey="16067249407809359746114321133992130903102335882983385972747813693681808870497"

	//authorize prikey="17822967620457187568904804290291537271142779717280482398091401115827760898835"
	//authorize pubkey="13519883267141251871527102103999205179714486518503885909948192364772977661583"

	//spend prikey="10407830929890509544473717262275616077696950294748419792758056545898949331744"
	//spend pubkey="12419942056983622012214804185935674735538011812395392042541464417352183370586"

	r1cs := NewDeposit()
	r1csBN256 := backend_bn256.Cast(r1cs)
	{
		good := backend.NewAssignment()
		good.Assign(backend.Public, "noteHash", "16308793397024662832064523892418908145900866571524124093537199035808550255649")
		good.Assign(backend.Public, "amount", "28242048")

		good.Assign(backend.Secret, "receiverPubKey", "13735985067536865723202617343666111332145536963656464451727087263423649028705")
		good.Assign(backend.Secret, "returnPubKey", "16067249407809359746114321133992130903102335882983385972747813693681808870497")
		good.Assign(backend.Secret, "authorizePubKey", "13519883267141251871527102103999205179714486518503885909948192364772977661583")
		good.Assign(backend.Secret, "noteRandom", "2824204835")

		assert.Solved(&r1csBN256, good, nil)
	}

}
