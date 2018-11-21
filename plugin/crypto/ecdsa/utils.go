// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
)

// MarshalECDSASignature marshal ECDSA signature
func MarshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(signatureECDSA{r, s})
}

// UnmarshalECDSASignature unmarshal ECDSA signature
func UnmarshalECDSASignature(raw []byte) (*big.Int, *big.Int, error) {
	sig := new(signatureECDSA)
	_, err := asn1.Unmarshal(raw, sig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed unmashalling signature [%s]", err)
	}

	if sig.R == nil {
		return nil, nil, errors.New("invalid signature, R must be different from nil")
	}
	if sig.S == nil {
		return nil, nil, errors.New("invalid signature, S must be different from nil")
	}

	if sig.R.Sign() != 1 {
		return nil, nil, errors.New("invalid signature, R must be larger than zero")
	}
	if sig.S.Sign() != 1 {
		return nil, nil, errors.New("invalid signature, S must be larger than zero")
	}

	return sig.R, sig.S, nil
}

// ToLowS convert to low int
func ToLowS(k *ecdsa.PublicKey, s *big.Int) *big.Int {
	lowS := IsLowS(s)
	if !lowS {
		s.Sub(k.Params().N, s)

		return s
	}

	return s
}

// IsLowS check is low int
func IsLowS(s *big.Int) bool {
	return s.Cmp(new(big.Int).Rsh(elliptic.P256().Params().N, 1)) != 1
}

func parsePubKey(pubKeyStr []byte, curve elliptic.Curve) (key *ecdsa.PublicKey, err error) {
	pubkey := ecdsa.PublicKey{}
	pubkey.Curve = curve

	if len(pubKeyStr) == 0 {
		return nil, errors.New("pubkey string is empty")
	}

	pubkey.X = new(big.Int).SetBytes(pubKeyStr[1:33])
	pubkey.Y = new(big.Int).SetBytes(pubKeyStr[33:])
	if pubkey.X.Cmp(pubkey.Curve.Params().P) >= 0 {
		return nil, fmt.Errorf("pubkey X parameter is >= to P")
	}
	if pubkey.Y.Cmp(pubkey.Curve.Params().P) >= 0 {
		return nil, fmt.Errorf("pubkey Y parameter is >= to P")
	}
	if !pubkey.Curve.IsOnCurve(pubkey.X, pubkey.Y) {
		return nil, fmt.Errorf("pubkey isn't on secp256k1 curve")
	}
	return &pubkey, nil
}

// SerializePublicKey serialize public key
func SerializePublicKey(p *ecdsa.PublicKey) []byte {
	b := make([]byte, 0, publicKeyECDSALength)
	b = append(b, 0x4)
	b = paddedAppend(32, b, p.X.Bytes())
	return paddedAppend(32, b, p.Y.Bytes())
}

// SerializePrivateKey serialize private key
func SerializePrivateKey(p *ecdsa.PrivateKey) []byte {
	b := make([]byte, 0, privateKeyECDSALength)
	return paddedAppend(privateKeyECDSALength, b, p.D.Bytes())
}

func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}
