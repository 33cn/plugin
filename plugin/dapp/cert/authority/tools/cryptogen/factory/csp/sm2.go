// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package csp

import (
	"encoding/asn1"
	"fmt"
	"math/big"

	"github.com/tjfoc/gmsm/sm2"
)

type sm2Signer struct{}

func (s *sm2Signer) Sign(k Key, digest []byte, opts SignerOpts) (signature []byte, err error) {
	return signSM2(k.(*SM2PrivateKey).PrivKey, digest, opts)
}

func signSM2(k *sm2.PrivateKey, digest []byte, opts SignerOpts) (signature []byte, err error) {
	r, s, err := sm2.Sm2Sign(k, digest, nil)
	if err != nil {
		return nil, err
	}

	return MarshalSM2Signature(r, s)
}

// SM2Signature SM2签名结构
type SM2Signature struct {
	R, S *big.Int
}

// MarshalSM2Signature 编码SM2起签名
func MarshalSM2Signature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(SM2Signature{r, s})
}

type sm2KeyGenerator struct {
}

func (kg *sm2KeyGenerator) KeyGen(opts int) (k Key, err error) {
	ln := big.NewInt(0).Rsh(sm2.P256Sm2().Params().N, 1)
	for {
		privKey, err := sm2.GenerateKey()
		if err != nil {
			return nil, fmt.Errorf("Failed generating SM2 key for: [%s]", err)
		}

		if ln.Cmp(privKey.D) == 1 {
			//fmt.Println("priv:"+common.ToHex(privKey.D.Bytes()))
			return &SM2PrivateKey{privKey}, nil
		}
	}

}
