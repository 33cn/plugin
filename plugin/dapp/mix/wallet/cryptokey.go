// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"

	"github.com/pkg/errors"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"

	wcom "github.com/33cn/chain33/wallet/common"
	"github.com/33cn/plugin/plugin/crypto/legacymimc"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

const CECBLOCKSIZE = 32

/*
从secp256k1根私钥创建支票需要的私钥和公钥
payPrivKey = rootPrivKey *G_X25519 这样很难泄露rootPrivKey

支票花费key:  payPrivKey
支票收款key： ReceiveKey= hash(payPrivKey)  --或者*G的X坐标值, 看哪个电路少？
DH加解密key: encryptPubKey= payPrivKey *G_X25519, 也是很安全的，只是电路里面目前不支持x25519
*/
func newPrivacyKey(rootPrivKey []byte) *mixTy.AccountPrivacyKey {
	ecdh := X25519()
	key := ecdh.PublicKey(rootPrivKey)
	payPrivKey := key.([32]byte)

	//payPrivKey := mimcHashByte([][]byte{rootPrivKey})
	//payPrivKey 可能超出fr的模，spendKey是payPrivKey对fr取的模，有可能和payPrivKey不相等，这里用spendKey取hash
	//mimcHashByte 会对输入参数对fr取模，在电路上不会影响ReceiveKey
	paymentKey := &mixTy.NoteKeyPair{}
	paymentKey.SpendKey = mixTy.Byte2Str(payPrivKey[:])
	paymentKey.ReceiveKey = mixTy.Byte2Str(mimcHashByte([][]byte{mixTy.Str2Byte(paymentKey.SpendKey)}))

	encryptKeyPair := &mixTy.EncryptSecretKeyPair{}
	pubkey := ecdh.PublicKey(payPrivKey)
	//加解密是在x25519域，需要Hex编码，不要使用fr.string, 模范围不同
	encryptKeyPair.SecretPrivKey = hex.EncodeToString(payPrivKey[:])
	pubData := pubkey.([32]byte)
	encryptKeyPair.SecretPubKey = hex.EncodeToString(pubData[:])

	privacy := &mixTy.AccountPrivacyKey{}
	privacy.PaymentKey = paymentKey
	privacy.SecretKey = encryptKeyPair

	return privacy
}

// CEC加密需要保证明文是秘钥的倍数，如果不是，则需要填充明文，在解密时候把填充物去掉
// 填充算法有pkcs5,pkcs7, 比如Pkcs5的思想填充的值为填充的长度，比如加密he,不足8
// 则填充为he666666, 解密后直接算最后一个值为6，把解密值的后6个Byte去掉即可
func pKCS5Padding(plainText []byte, blockSize int) []byte {
	if blockSize < CECBLOCKSIZE {
		blockSize = CECBLOCKSIZE
	}
	padding := blockSize - (len(plainText) % blockSize)
	//fmt.Println("pading", "passsize", blockSize, "plaintext", len(plainText), "pad", padding)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	//fmt.Println("padding", padding, "text", common.ToHex(padText[:]))
	newText := append(plainText, padText...)
	return newText
}

func pKCS5UnPadding(plainText []byte) ([]byte, error) {
	length := len(plainText)
	number := int(plainText[length-1])
	if number > length {
		return nil, types.ErrInvalidParam
	}
	return plainText[:length-number], nil
}

func encryptDataWithPadding(password, data []byte) []byte {
	paddingText := pKCS5Padding(data, len(password))
	return wcom.CBCEncrypterPrivkey(password, paddingText)
}

func encryptData(peerPubKey string, data []byte) (*mixTy.DHSecret, error) {
	ecdh := X25519()
	oncePriv, oncePub, err := ecdh.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "x25519 generate key")
	}

	peerPubByte, err := hex.DecodeString(peerPubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "encrypt Decode peer pubkey=%s", peerPubKey)
	}
	password := ecdh.ComputeSecret(oncePriv, peerPubByte)
	encrypt := encryptDataWithPadding(password, data)

	pubData := oncePub.([32]byte)
	return &mixTy.DHSecret{OneTimePubKey: hex.EncodeToString(pubData[:]), Secret: common.ToHex(encrypt)}, nil

}

func hasPrivKeyMagic(data []byte) bool {
	if len(data) < len(wcom.MagicPrivKey)+1 {
		return false
	}
	return bytes.Equal(data[:len(wcom.MagicPrivKey)], wcom.MagicPrivKey) &&
		data[len(wcom.MagicPrivKey)] == wcom.KdfVersion
}

func unpadCBCPlain(plainData []byte, format string, cipherLen int) ([]byte, error) {
	if len(plainData) == 0 {
		bizlog.Error("decryptDataWithPading empty plaintext", "format", format, "cipherLen", cipherLen)
		return nil, types.ErrInvalidParam
	}
	plain, err := pKCS5UnPadding(plainData)
	if err != nil {
		bizlog.Error("decryptDataWithPading unpadding", "format", format, "cipherLen", cipherLen,
			"plainLen", len(plainData), "err", err)
		return nil, errors.Wrapf(err, "decryptDataWithPading %s unpadding", format)
	}
	return plain, nil
}

func decryptRandomIV(password, data []byte) ([]byte, error) {
	key := make([]byte, 32)
	copy(key, password)
	block, err := aes.NewCipher(key)
	if err != nil {
		bizlog.Error("decryptDataWithPading aes.NewCipher", "cipherLen", len(data), "err", err)
		return nil, errors.Wrapf(err, "decryptDataWithPading aes.NewCipher")
	}
	iv := data[:block.BlockSize()]
	ciphertext := data[block.BlockSize():]
	if len(ciphertext) == 0 || len(ciphertext)%block.BlockSize() != 0 {
		bizlog.Error("decryptDataWithPading invalid random-IV ciphertext",
			"cipherLen", len(data), "ctLen", len(ciphertext))
		return nil, types.ErrInvalidParam
	}
	decrypted := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(decrypted, ciphertext)
	return unpadCBCPlain(decrypted, "random-IV", len(data))
}

func decryptDataWithPading(password, data []byte) ([]byte, error) {
	// 按内容分流，兼容 chain33 三代密文：
	//  1. v1.71.0+：Magic(C33K)+version+salt+IV+ciphertext，PBKDF2 派生密钥
	//  2. v0.69.1~v1.70：IV(16)+ciphertext，口令零填充为密钥（mix 明文可为 96/160，
	//     chain33 CBCDecrypterPrivkey 第 2 代只认 32/64，必须由 mix 自解）
	//  3. 最初：ciphertext-only，IV=key[:16]
	if len(data) == 0 {
		bizlog.Error("decryptDataWithPading empty ciphertext")
		return nil, types.ErrInvalidParam
	}

	if hasPrivKeyMagic(data) {
		return unpadCBCPlain(wcom.CBCDecrypterPrivkey(password, data), "magic-kdf", len(data))
	}
	if len(data)%32 == 16 {
		return decryptRandomIV(password, data)
	}
	if len(data)%16 != 0 {
		bizlog.Error("decryptDataWithPading invalid cipher length",
			"cipherLen", len(data), "mod16", len(data)%16, "mod32", len(data)%32)
		return nil, types.ErrInvalidParam
	}
	return unpadCBCPlain(wcom.CBCDecrypterPrivkey(password, data), "legacy", len(data))
}

func decryptData(selfPrivKey string, peerPubKey string, cryptData []byte) ([]byte, error) {
	ecdh := X25519()
	self, err := hex.DecodeString(selfPrivKey)
	if err != nil {
		return nil, errors.Wrapf(err, "decrypt Decode self prikey=%s", selfPrivKey)
	}
	peer, err := hex.DecodeString(peerPubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "decrypt Decode peer pubkey=%s", peerPubKey)
	}
	password := ecdh.ComputeSecret(self, peer)
	return decryptDataWithPading(password, cryptData)
}

func mimcHashString(params []string) []byte {
	var sum []byte
	for _, k := range params {
		sum = append(sum, mixTy.Str2Byte(k)...)
	}
	hash := mimcHashCalc(sum)
	return hash

}

func mimcHashByte(params [][]byte) []byte {
	var sum []byte
	for _, k := range params {
		sum = append(sum, k...)
	}
	hash := mimcHashCalc(sum)
	return hash

}

func mimcHashCalc(sum []byte) []byte {
	h := legacymimc.NewMiMC(mixTy.MimcHashSeed)
	h.Write(sum)
	return h.Sum(nil)
}
