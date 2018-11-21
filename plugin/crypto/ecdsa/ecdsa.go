// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ecdsa

import (
	"bytes"
	"errors"
	"fmt"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"math/big"

	"github.com/33cn/chain33/common/crypto"
)

const (
	privateKeyECDSALength = 32
	publicKeyECDSALength  = 65
)

// Driver driver
type Driver struct{}

// GenKey create private key
func (d Driver) GenKey() (crypto.PrivKey, error) {
	privKeyBytes := [privateKeyECDSALength]byte{}
	copy(privKeyBytes[:], crypto.CRandBytes(privateKeyECDSALength))
	priv, _ := privKeyFromBytes(elliptic.P256(), privKeyBytes[:])
	copy(privKeyBytes[:], SerializePrivateKey(priv))
	return PrivKeyECDSA(privKeyBytes), nil
}

// PrivKeyFromBytes create private key from bytes
func (d Driver) PrivKeyFromBytes(b []byte) (privKey crypto.PrivKey, err error) {
	if len(b) != privateKeyECDSALength {
		return nil, errors.New("invalid priv key byte")
	}

	privKeyBytes := new([privateKeyECDSALength]byte)
	copy(privKeyBytes[:], b[:privateKeyECDSALength])
	priv, _ := privKeyFromBytes(elliptic.P256(), privKeyBytes[:])

	copy(privKeyBytes[:], SerializePrivateKey(priv))
	return PrivKeyECDSA(*privKeyBytes), nil
}

// PubKeyFromBytes create public key from bytes
func (d Driver) PubKeyFromBytes(b []byte) (pubKey crypto.PubKey, err error) {
	if len(b) != publicKeyECDSALength {
		return nil, errors.New("invalid pub key byte")
	}
	pubKeyBytes := new([publicKeyECDSALength]byte)
	copy(pubKeyBytes[:], b[:])
	return PubKeyECDSA(*pubKeyBytes), nil
}

// SignatureFromBytes create signature from bytes
func (d Driver) SignatureFromBytes(b []byte) (sig crypto.Signature, err error) {
	var certSignature crypto.CertSignature
	_, err = asn1.Unmarshal(b, &certSignature)
	if err != nil {
		return SignatureECDSA(b), nil
	}

	if len(certSignature.Cert) == 0 {
		return SignatureECDSA(b), nil
	}

	return SignatureECDSA(certSignature.Signature), nil
}

// PrivKeyECDSA PrivKey
type PrivKeyECDSA [privateKeyECDSALength]byte

// Bytes convert to bytes
func (privKey PrivKeyECDSA) Bytes() []byte {
	s := make([]byte, privateKeyECDSALength)
	copy(s, privKey[:])
	return s
}

// Sign create signature
func (privKey PrivKeyECDSA) Sign(msg []byte) crypto.Signature {
	priv, pub := privKeyFromBytes(elliptic.P256(), privKey[:])
	r, s, err := ecdsa.Sign(rand.Reader, priv, crypto.Sha256(msg))
	if err != nil {
		return nil
	}

	s = ToLowS(pub, s)
	ecdsaSigByte, _ := MarshalECDSASignature(r, s)
	return SignatureECDSA(ecdsaSigByte)
}

// PubKey convert to public key
func (privKey PrivKeyECDSA) PubKey() crypto.PubKey {
	_, pub := privKeyFromBytes(elliptic.P256(), privKey[:])
	var pubECDSA PubKeyECDSA
	copy(pubECDSA[:], SerializePublicKey(pub))
	return pubECDSA
}

// Equals check privkey is equal
func (privKey PrivKeyECDSA) Equals(other crypto.PrivKey) bool {
	if otherSecp, ok := other.(PrivKeyECDSA); ok {
		return bytes.Equal(privKey[:], otherSecp[:])
	}

	return false
}

// String convert to string
func (privKey PrivKeyECDSA) String() string {
	return fmt.Sprintf("PrivKeyECDSA{*****}")
}

// PubKeyECDSA PubKey
// prefixed with 0x02 or 0x03, depending on the y-cord.
type PubKeyECDSA [publicKeyECDSALength]byte

// Bytes convert to bytes
func (pubKey PubKeyECDSA) Bytes() []byte {
	s := make([]byte, publicKeyECDSALength)
	copy(s, pubKey[:])
	return s
}

// VerifyBytes verify signature
func (pubKey PubKeyECDSA) VerifyBytes(msg []byte, sig crypto.Signature) bool {
	// unwrap if needed
	if wrap, ok := sig.(SignatureS); ok {
		sig = wrap.Signature
	}
	// and assert same algorithm to sign and verify
	sigECDSA, ok := sig.(SignatureECDSA)
	if !ok {
		return false
	}

	pub, err := parsePubKey(pubKey[:], elliptic.P256())
	if err != nil {
		return false
	}

	r, s, err := UnmarshalECDSASignature(sigECDSA)
	if err != nil {
		return false
	}

	lowS := IsLowS(s)
	if !lowS {
		return false
	}
	return ecdsa.Verify(pub, crypto.Sha256(msg), r, s)
}

// String convert to string
func (pubKey PubKeyECDSA) String() string {
	return fmt.Sprintf("PubKeyECDSA{%X}", pubKey[:])
}

// KeyString Must return the full bytes in hex.
// Used for map keying, etc.
func (pubKey PubKeyECDSA) KeyString() string {
	return fmt.Sprintf("%X", pubKey[:])
}

// Equals check public key is equal
func (pubKey PubKeyECDSA) Equals(other crypto.PubKey) bool {
	if otherSecp, ok := other.(PubKeyECDSA); ok {
		return bytes.Equal(pubKey[:], otherSecp[:])
	}
	return false
}

type signatureECDSA struct {
	R, S *big.Int
}

// SignatureECDSA Signature
type SignatureECDSA []byte

// SignatureS signature struct
type SignatureS struct {
	crypto.Signature
}

// Bytes convert signature to bytes
func (sig SignatureECDSA) Bytes() []byte {
	s := make([]byte, len(sig))
	copy(s, sig[:])
	return s
}

// IsZero check signature is zero
func (sig SignatureECDSA) IsZero() bool { return len(sig) == 0 }

// String convert signature to string
func (sig SignatureECDSA) String() string {
	fingerprint := make([]byte, len(sig[:]))
	copy(fingerprint, sig[:])
	return fmt.Sprintf("/%X.../", fingerprint)

}

// Equals check signature equals
func (sig SignatureECDSA) Equals(other crypto.Signature) bool {
	if otherEd, ok := other.(SignatureECDSA); ok {
		return bytes.Equal(sig[:], otherEd[:])
	}
	return false
}

// Name name
const Name = "auth_ecdsa"

// ID id
const ID = 257

func init() {
	crypto.Register(Name, &Driver{})
	crypto.RegisterType(Name, ID)
}

func privKeyFromBytes(curve elliptic.Curve, pk []byte) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	x, y := curve.ScalarBaseMult(pk)

	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(pk),
	}

	return priv, &priv.PublicKey
}
