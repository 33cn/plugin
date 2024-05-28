// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crypto

import (
	"crypto/ecdsa"
	"math/big"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/btcsuite/btcd/btcec"
	"golang.org/x/crypto/sha3"
)

// ValidateSignatureValues 校验签名信息是否正确
func ValidateSignatureValues(r, s *big.Int) bool {
	if r.Cmp(common.Big1) < 0 || s.Cmp(common.Big1) < 0 {
		return false
	}
	return true
}

// Ecrecover 根据压缩消息和签名，返回非压缩的公钥信息
func Ecrecover(hash, sig []byte) ([]byte, error) {
	return ethCrypto.Ecrecover(hash, sig)
}

// CompressPubKey compress pub key
func CompressPubKey(pubKey []byte) ([]byte, error) {
	pub, err := ethCrypto.UnmarshalPubkey(pubKey)
	if err != nil {
		return nil, err
	}
	return ethCrypto.CompressPubkey(pub), nil
}

// SigToPub 根据签名返回公钥信息
func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	btcsig := make([]byte, 65)
	btcsig[0] = sig[64] + 27
	copy(btcsig[1:], sig)

	pub, _, err := btcec.RecoverCompact(btcec.S256(), btcsig, hash)
	return (*ecdsa.PublicKey)(pub), err
}

// Keccak256 计算并返回 Keccak256 哈希
func Keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h common.Hash) {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// CreateAddress2 creates an ethereum address given the address bytes, initial
// contract code hash and a salt.
func CreateAddress2(b common.Address, salt [32]byte, inithash []byte) common.Address {
	return common.BytesToAddress(Keccak256([]byte{0xff}, b.Bytes(), salt[:], inithash)[12:])
}
