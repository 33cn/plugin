package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common/crypto"
)

// "input" is (hash, v, r, s), each 32 bytes
//cc2427f9c0de9457dbaeddc063888c47353ce8df2b8571944ebe389acf049dc0100000000000000000000000000000000000000000000000000000000000001b4cf8ae36067ccb3ee2f15b91e7f3e10fac9f38fc020ded6dbae65874c84a9c5d24c849055170d52dd5b19c5b4408e5f64e6f147cb456a01c4160b444fbb11a54
//0x883118bfb20db555e5fc151e81d54bf5718640d72ee28b88823c9686024d7b07000000000000000000000000000000000000000000000000000000000000001cb1badd81830bdc8de25b84333284c5e04241654b6f927cbf353bbba3df35da2a5254673fe57fcb047e334522a12ed49122895a050cfdeed35f98adb7c6cc0c21

func TestEcrecoverSignTypedMessage(t *testing.T) {
	hashStr := "0xcc2427f9c0de9457dbaeddc063888c47353ce8df2b8571944ebe389acf049dc0"
	sigStr := "0x4cf8ae36067ccb3ee2f15b91e7f3e10fac9f38fc020ded6dbae65874c84a9c5d24c849055170d52dd5b19c5b4408e5f64e6f147cb456a01c4160b444fbb11a541b"
	input, _ := common.HexToBytes(hashStr)
	sig, _ := common.HexToBytes(sigStr)
	sig[64] -= 27
	pubKey, err := crypto.Ecrecover(input[:32], sig)
	// make sure the public key is a valid one
	if err != nil {
		log15.Info("ecrecover", "failed due to", err.Error())
		panic("ecrecover")
	}
	pubKey, err = crypto.CompressPubKey(pubKey)
	assert.Nil(t, err)
	//fmt.Println("ecrecover", "pubkey", common.Bytes2Hex(pubKey))
	//fmt.Println("recoverd address", address.PubKeyToAddress(pubKey).String())
	addr := common.PubKey2Address(pubKey)
	hash160Str := common.Bytes2Hex(common.LeftPadBytes(addr.Bytes(), 32))
	assert.Equal(t, hash160Str, "0x000000000000000000000000245afbf176934ccdd7ca291a8dddaa13c8184822")
	assert.Equal(t, addr.String(), "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
	//fmt.Println("TestEcrecoverSignTypedMessage", "hash160Str", hash160Str)

	//privateKeyStr := "0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"
	//var driver secp256k1.Driver
	//privateKeySli, err := chain33Common.FromHex(privateKeyStr)
	//if nil != err {
	//	panic(err.Error())
	//
	//}
	//ownerPrivateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	//if nil != err {
	//	panic(err.Error())
	//}
	//fmt.Println("Origin user", "pubkey", common.Bytes2Hex(ownerPrivateKey.PubKey().Bytes()))
}
