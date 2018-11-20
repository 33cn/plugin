// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package csp

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"fmt"
	"math/big"

	auth "github.com/33cn/plugin/plugin/crypto/ecdsa"
)

type ecdsaSigner struct{}

func (s *ecdsaSigner) Sign(k Key, digest []byte, opts SignerOpts) (signature []byte, err error) {
	return signECDSA(k.(*ecdsaPrivateKey).privKey, digest, opts)
}

func signECDSA(k *ecdsa.PrivateKey, digest []byte, opts SignerOpts) (signature []byte, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, k, digest)
	if err != nil {
		return nil, err
	}

	s = auth.ToLowS(&k.PublicKey, s)

	return MarshalECDSASignature(r, s)
}

// ECDSASignature ECDSA签名结构
type ECDSASignature struct {
	R, S *big.Int
}

// MarshalECDSASignature 编码ECDSA类型签名
func MarshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(ECDSASignature{r, s})
}

type ecdsaKeyGenerator struct {
	curve elliptic.Curve
}

func (kg *ecdsaKeyGenerator) KeyGen(opts int) (k Key, err error) {
	privKey, err := ecdsa.GenerateKey(kg.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Failed generating ECDSA key for [%v]: [%s]", kg.curve, err)
	}

	return &ecdsaPrivateKey{privKey}, nil
}
