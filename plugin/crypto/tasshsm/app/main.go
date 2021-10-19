package main

import (
	"fmt"
	"math/big"
	"time"

	"github.com/33cn/chain33/system/crypto/secp256k1"

	"github.com/33cn/chain33/common/crypto"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/33cn/chain33/system/crypto/sm2"
	gmsm_sm2 "github.com/tjfoc/gmsm/sm2"

	"github.com/33cn/chain33/common"
	"github.com/33cn/plugin/plugin/crypto/tasshsm/adapter"
)

func verifySM2Signature(rBytes, sBytes, msg []byte) bool {
	xBytes, _ := common.FromHex("0000000000000000000000000000000000000000000000000000000000000000FD4241057FEC6CBEEC501F7E1763751B8F6DFCFB910FB634FBB76A16639EF172")
	yBytes, _ := common.FromHex("00000000000000000000000000000000000000000000000000000000000000001C6DA89F9C1A5EE9B6108E5A2A5FE336962630A34DBA1AF428451E1CE63BB3CF")
	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)

	publicKey := &gmsm_sm2.PublicKey{
		X: x,
		Y: y,
	}
	var pubSM2 sm2.PubKeySM2
	copy(pubSM2[:], gmsm_sm2.Compress(publicKey))

	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)
	signature := sm2.SignatureSM2(sm2.Serialize(r, s))

	return pubSM2.VerifyBytes(msg, signature)
}

func main() {
	if err := adapter.OpenHSMSession(); nil != err {
		panic("Failed to OpenHSMSession")
	}
	fmt.Println("Succeed to OpenHSMSession")
	fmt.Println("   ")
	fmt.Println("   ")
	fmt.Println("   ")

	keyIndex := 1
	passwd := "a1234567"
	//passwd := []byte("a1234567")

	if err := adapter.GetPrivateKeyAccessRight(passwd, keyIndex); nil != err {
		panic("Failed to GetPrivateKeyAccessRight")
	}

	verifySecp256k1(keyIndex)

	if err := adapter.ReleaeAccessRight(keyIndex); nil != err {
		panic("Failed to GetPrivateKeyAccessRight")
	}
	adapter.CloseHSMSession()
}

func verifySecp256k1(keyIndex int) {
	msg, _ := common.FromHex("456789")
	r, s, err := adapter.SignSecp256k1(msg, keyIndex)
	if err != nil {
		panic("Failed to SignSecp256k1 due to:" + err.Error())
	}
	fmt.Println("signature R=", common.ToHex(r))
	fmt.Println("signature S=", common.ToHex(s))

	///////构建公钥////////
	pub, _ := common.FromHex("04C24FBA65F8CD81223D2935EDEA663048A1BEFB5A78BC67C80DCB5A1D601F898C35EA242D2E76CACE9EE5A61DBDA29A5076707325FE20B5A80DB0CA6D02C5D983")

	///////构建以太坊签名并验证////////
	ethSig := append(r, s...)
	hash := crypto.Sha256(msg)
	VerifyResult4Eth := ethCrypto.VerifySignature(pub, hash[:], ethSig[:64])
	if !VerifyResult4Eth {
		panic("Failed to do Signature verification for Ethereum")
	}
	fmt.Println(" ^-^ Succeed to do signature verification for Ethereum ^-^  ")
	fmt.Println("   ")
	fmt.Println("   ")
	fmt.Println("   ")

	///////构建chain33签名并验证////////
	derSig := adapter.MakeDERsignature(r, s)
	fmt.Println(" derSig ", common.ToHex(derSig))

	secpPubKey, err := ethCrypto.UnmarshalPubkey(pub)
	pub33Bytes := ethCrypto.CompressPubkey(secpPubKey)
	c := &secp256k1.Driver{}
	chain33PubKey, _ := c.PubKeyFromBytes(pub33Bytes)
	chain33Sig, err := c.SignatureFromBytes(derSig)
	VerifyResult4Chain33 := chain33PubKey.VerifyBytes(msg, chain33Sig)
	if !VerifyResult4Chain33 {
		panic("Failed to do Signature verification for Chain33")
	}
	fmt.Println(" ^-^ Succeed to do signature verification for Chain33 ^-^  ")
}

func verifySM2() {
	//msg := []byte("112233445566112233445566112233445566112233445566")
	msg, _ := common.FromHex("112233445566112233445566112233445566112233445566")
	r, s, err := adapter.SignSM2Internal(msg, 10)
	if err != nil {
		panic("Failed to SignSM2Internal due to:" + err.Error())
	}
	fmt.Println("signature R=", common.ToHex(r))
	fmt.Println("signature S=", common.ToHex(s))

	///////构建公钥////////
	xBytes, _ := common.FromHex("0000000000000000000000000000000000000000000000000000000000000000FD4241057FEC6CBEEC501F7E1763751B8F6DFCFB910FB634FBB76A16639EF172")
	yBytes, _ := common.FromHex("00000000000000000000000000000000000000000000000000000000000000001C6DA89F9C1A5EE9B6108E5A2A5FE336962630A34DBA1AF428451E1CE63BB3CF")
	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)

	publicKey := &gmsm_sm2.PublicKey{
		X: x,
		Y: y,
	}
	var pubSM2 sm2.PubKeySM2
	copy(pubSM2[:], gmsm_sm2.Compress(publicKey))

	///////开始循环验证签名////////
	now := time.Now()
	now.Nanosecond()

	msLater := now.Nanosecond()
	fmt.Printf("msLater = %d\n", msLater)
	time.Sleep(time.Millisecond)
	msLater = now.Nanosecond()
	fmt.Printf("msLater = %d\n", time.Now().Nanosecond())
	fmt.Printf("msLater Sec = %d\n", time.Now().Second())
	for i := 0; i < 10*1000; i++ {
		adapter.SignSM2Internal(msg, 10)
		//rBytes, sBytes, _ := adapter.SignSM2Internal(msg, 10)
		//fmt.Println("rBytes = ", common.ToHex(rBytes))
		//fmt.Println("sBytes = ", common.ToHex(sBytes))
		//r := new(big.Int).SetBytes(rBytes)
		//s := new(big.Int).SetBytes(sBytes)
		//signature := sm2.SignatureSM2(sm2.Serialize(r, s))
		//if !pubSM2.VerifyBytes(msg, signature) {
		//	panic("Failed to do VerifyBytes")
		//}
		//fmt.Println("Succeed to do VerifyBytes for times = ", i)
	}

	fmt.Println("      ")
	fmt.Printf("testLater = %d\n", time.Now().Nanosecond())
	fmt.Printf("testLater sec = %d\n", time.Now().Second())
	fmt.Println("      ")
	fmt.Println(" ^-^ Successful ^-^  ")
}
