package main

import (
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto"

	chain33Common "github.com/33cn/chain33/common"
	"github.com/33cn/chain33/system/crypto/sm2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	gmsm2 "github.com/tjfoc/gmsm/sm2"
	gmsm4 "github.com/tjfoc/gmsm/sm4"
)

func sm2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sm2",
		Short: "generate sm2 key and decrypt",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		createSm2Cmd(),
		decryptWithSm2Cmd(),
		encryptWithSm2Cmd(),
	)
	return cmd
}

func createSm2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create sm2 key",
		Run:   createSm2Key,
	}
	return cmd
}

func createSm2Key(cmd *cobra.Command, args []string) {
	privateKey, err := gmsm2.GenerateKey()
	if nil != err {
		fmt.Println("gmsm2.GenerateKeyfailed due to:" + err.Error())
		return
	}

	pub := sm2.SerializePublicKey(&privateKey.PublicKey, false)
	pri := sm2.SerializePrivateKey(privateKey)
	pub = pub[1:]
	fmt.Println("sm2 public  key = "+common.Bytes2Hex(pub), "len = ", len(pub))
	fmt.Println("sm2 private key = "+common.Bytes2Hex(pri), "len = ", len(pri))
}

func decryptWithSm2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decipher",
		Short: "decipher with sm2 to recover privake key",
		Run:   decryptWithSm2,
	}
	addDecryptWithSm2Flags(cmd)
	return cmd
}

func addDecryptWithSm2Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("sm2key", "k", "", "sm2 private key")
	_ = cmd.MarkFlagRequired("sm2key")

	cmd.Flags().StringP("symmKeyCipher", "s", "", "symmKey Ciphered by random sm4 key, will be prefixed by 0x04 automatically")
	_ = cmd.MarkFlagRequired("symmKeyCipher")

	cmd.Flags().StringP("cipher", "c", "", "ciphered text from private key")
	_ = cmd.MarkFlagRequired("cipher")
}

func decryptWithSm2(cmd *cobra.Command, args []string) {
	sm2keyStr, _ := cmd.Flags().GetString("sm2key")
	cipherStr, _ := cmd.Flags().GetString("cipher")
	symmKeyCipherStr, _ := cmd.Flags().GetString("symmKeyCipher")

	//第一步，解密数字信封
	sm2key, err := chain33Common.FromHex(sm2keyStr)
	if nil != err {
		fmt.Println("chain33Common.FromHex failed due to:" + err.Error())
		return
	}
	if 32 != len(sm2key) {
		fmt.Println("Wrong sm2key length", len(sm2key))
		return

	}

	curve := gmsm2.P256Sm2()
	x, y := curve.ScalarBaseMult(sm2key)
	sm2Priv := &gmsm2.PrivateKey{
		PublicKey: gmsm2.PublicKey{
			Curve: gmsm2.P256Sm2(),
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(sm2key),
	}

	symmKeyCipher, err := chain33Common.FromHex(symmKeyCipherStr)
	if nil != err {
		fmt.Println("chain33Common.FromHex failed for cipher due to:" + err.Error())
		return
	}
	symmKeyCipher = append([]byte{0x04}, symmKeyCipher...)
	sm4Key, err := sm2Priv.Decrypt(symmKeyCipher)
	if nil != err {
		fmt.Println("sm2 decrypt failed due to:" + err.Error())
		return
	}
	fmt.Println("The decrypted sm4 key is:"+chain33Common.ToHex(sm4Key), "len:", len(sm4Key))
	//第二步，通过数字信封中的对称密钥，进行sm4对称解密
	sm4Cihpher, err := gmsm4.NewCipher(sm4Key)
	if err != nil {
		fmt.Println("gmsm4.NewCipher failed due to:" + err.Error())
		return
	}

	cipher, err := chain33Common.FromHex(cipherStr)
	if nil != err {
		fmt.Println("chain33Common.FromHex failed for cipher due to:" + err.Error())
		return
	}
	dst := make([]byte, 32)
	sm4Cihpher.Decrypt(dst, cipher)
	sm4Cihpher.Decrypt(dst[16:], cipher[16:])
	fmt.Println(chain33Common.ToHex(dst))
}

func encryptWithSm2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encipher",
		Short: "encipher with sm2 to encipher privake key",
		Run:   encryptWithSm2,
	}
	addEncryptWithSm2Flags(cmd)
	return cmd
}

func addEncryptWithSm2Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("sm2key", "t", "", "sm2 private key to encrypt sm4 key")
	_ = cmd.MarkFlagRequired("sm2key")

	cmd.Flags().StringP("sm4key", "f", "", "sm4 symmKey, will be prefixed by 0x04 automatically")
	_ = cmd.MarkFlagRequired("sm4key")

	cmd.Flags().StringP("key", "k", "", "private key to be encrypted")
	_ = cmd.MarkFlagRequired("key")
}

func encryptWithSm2(cmd *cobra.Command, args []string) {
	sm2keyStr, _ := cmd.Flags().GetString("sm2key")
	privatKeyStr, _ := cmd.Flags().GetString("key")
	symmKeyCipher, _ := cmd.Flags().GetString("sm4key")
	sm4Key, err := chain33Common.FromHex(symmKeyCipher)
	if nil != err {
		fmt.Println("chain33Common.FromHex failed for cipher due to:" + err.Error())
		return
	}

	//第一步，通过数字信封中的对称密钥，进行sm4对称加密
	sm4Cihpher, err := gmsm4.NewCipher(sm4Key)
	if err != nil {
		fmt.Println("gmsm4.NewCipher failed due to:" + err.Error())
		return
	}

	privateKeySlice, err := chain33Common.FromHex(privatKeyStr)
	if nil != err {
		fmt.Println("chain33Common.FromHex failed for cipher due to:" + err.Error())
		return
	}
	if len(privateKeySlice) != 32 {
		fmt.Println("invalid priv key length", len(privateKeySlice))
		return
	}
	dst := make([]byte, 32)
	sm4Cihpher.Encrypt(dst, privateKeySlice)
	sm4Cihpher.Encrypt(dst[16:], privateKeySlice[16:])

	//第二步，加密数字信封
	sm2key, err := chain33Common.FromHex(sm2keyStr)
	if nil != err {
		fmt.Println("chain33Common.FromHex failed due to:" + err.Error())
		return
	}
	if 32 != len(sm2key) {
		fmt.Println("Wrong sm2key length", len(sm2key))
		return

	}

	curve := gmsm2.P256Sm2()
	x, y := curve.ScalarBaseMult(sm2key)
	sm2Priv := &gmsm2.PrivateKey{
		PublicKey: gmsm2.PublicKey{
			Curve: gmsm2.P256Sm2(),
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(sm2key),
	}

	sm4Key, err = sm2Priv.Encrypt(sm4Key)
	if nil != err {
		fmt.Println("sm2 Encrypt failed due to:" + err.Error())
		return
	}
	sm4Key = sm4Key[1:]

	//第三步，计算secp256k1对应的公钥，非压缩
	_, pubKey := btcec.PrivKeyFromBytes(crypto.S256(), privateKeySlice)
	uncompressedKey := pubKey.SerializeUncompressed()
	uncompressedKey = uncompressedKey[1:]

	fmt.Println("随机对称密钥:"+common.Bytes2Hex(sm4Key), "len:", len(sm4Key))
	fmt.Println("需要导入的公钥:"+common.Bytes2Hex(uncompressedKey), "len:", len(uncompressedKey))
	fmt.Println("需要导入的私钥:"+common.Bytes2Hex(dst), "len:", len(dst))
}
