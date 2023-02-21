// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bls

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/33cn/chain33/common"

	"github.com/33cn/chain33/common/crypto"
	"github.com/phoreproject/bls/g1pubs"
)

//setting
const (
	BLSPrivateKeyLength = 32
	BLSPublicKeyLength  = 48
	BLSSignatureLength  = 96
)

// Driver driver
type Driver struct{}

// GenKey create private key
func (d Driver) GenKey() (crypto.PrivKey, error) {
	privKeyBytes := new([BLSPrivateKeyLength]byte)
	priv, err := g1pubs.RandKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	privBytes := priv.Serialize()
	copy(privKeyBytes[:], privBytes[:])
	return PrivKeyBLS(*privKeyBytes), nil
}

// MustPrivKeyFromBytes must get bls private key from bytes
func MustPrivKeyFromBytes(b []byte) (crypto.Crypto, crypto.PrivKey) {

	d := Driver{}
	key, err := d.PrivKeyFromBytes(b)

	for err != nil {
		copy(b[:], common.Sha256(b[:]))
		key, err = d.PrivKeyFromBytes(b)
	}
	return d, key
}

// PrivKeyFromBytes create private key from bytes
func (d Driver) PrivKeyFromBytes(b []byte) (privKey crypto.PrivKey, err error) {
	if len(b) != BLSPrivateKeyLength {
		return nil, errors.New("invalid bls priv key byte")
	}

	privKeyBytes := new([BLSPrivateKeyLength]byte)
	copy(privKeyBytes[:], b[:BLSPrivateKeyLength])
	priv := g1pubs.DeserializeSecretKey(*privKeyBytes)
	if priv.GetFRElement() == nil {
		return nil, errors.New("invalid bls privkey")
	}
	privBytes := priv.Serialize()
	copy(privKeyBytes[:], privBytes[:])
	return PrivKeyBLS(*privKeyBytes), nil
}

// PubKeyFromBytes create public key from bytes
func (d Driver) PubKeyFromBytes(b []byte) (pubKey crypto.PubKey, err error) {
	if len(b) != BLSPublicKeyLength {
		return nil, errors.New("invalid bls pub key byte")
	}
	pubKeyBytes := new([BLSPublicKeyLength]byte)
	copy(pubKeyBytes[:], b[:])
	return PubKeyBLS(*pubKeyBytes), nil
}

// SignatureFromBytes create signature from bytes
func (d Driver) SignatureFromBytes(b []byte) (sig crypto.Signature, err error) {
	sigBytes := new([BLSSignatureLength]byte)
	copy(sigBytes[:], b[:])
	return SignatureBLS(*sigBytes), nil
}

// Validate validate msg and signature
func (d Driver) Validate(msg, pub, sig []byte) error {
	return crypto.BasicValidation(d, msg, pub, sig)
}

//Aggregate aggregates signatures together into a new signature.
func (d Driver) Aggregate(sigs []crypto.Signature) (crypto.Signature, error) {
	if len(sigs) == 0 {
		return nil, errors.New("no signatures to aggregate")
	}
	g1sigs := make([]*g1pubs.Signature, 0, len(sigs))
	for i, sig := range sigs {
		g1sig, err := ConvertToSignature(sig)
		if err != nil {
			return nil, fmt.Errorf("%v(index: %d)", err, i)
		}
		g1sigs = append(g1sigs, g1sig)
	}
	agsig := g1pubs.AggregateSignatures(g1sigs)
	return SignatureBLS(agsig.Serialize()), nil
}

//AggregatePublic aggregates public keys together into a new PublicKey.
func (d Driver) AggregatePublic(pubs []crypto.PubKey) (crypto.PubKey, error) {
	if len(pubs) == 0 {
		return nil, errors.New("no public keys to aggregate")
	}
	//blank public key
	g1pubs := g1pubs.NewAggregatePubkey()
	for i, pub := range pubs {
		g1pub, err := ConvertToPublicKey(pub)
		if err != nil {
			return nil, fmt.Errorf("%v(index: %d)", err, i)
		}
		g1pubs.Aggregate(g1pub)
	}
	return PubKeyBLS(g1pubs.Serialize()), nil
}

// VerifyAggregatedOne verifies each public key against a message.
func (d Driver) VerifyAggregatedOne(pubs []crypto.PubKey, m []byte, sig crypto.Signature) error {
	g1pubs := make([]*g1pubs.PublicKey, 0, len(pubs))
	for i, pub := range pubs {
		g1pub, err := ConvertToPublicKey(pub)
		if err != nil {
			return fmt.Errorf("%v(index: %d)", err, i)
		}
		g1pubs = append(g1pubs, g1pub)
	}

	g1sig, err := ConvertToSignature(sig)
	if err != nil {
		return err
	}

	if g1sig.VerifyAggregateCommon(g1pubs, m) {
		return nil
	}
	return errors.New("bls signature mismatch")
}

// VerifyAggregatedN verifies each public key against each message.
func (d Driver) VerifyAggregatedN(pubs []crypto.PubKey, ms [][]byte, sig crypto.Signature) error {
	g1pubs := make([]*g1pubs.PublicKey, 0, len(pubs))
	for i, pub := range pubs {
		g1pub, err := ConvertToPublicKey(pub)
		if err != nil {
			return fmt.Errorf("%v(index: %d)", err, i)
		}
		g1pubs = append(g1pubs, g1pub)
	}

	g1sig, err := ConvertToSignature(sig)
	if err != nil {
		return err
	}

	if len(g1pubs) != len(ms) {
		return fmt.Errorf("different length of pubs and messages, %d vs %d", len(g1pubs), len(ms))
	}
	if g1sig.VerifyAggregate(g1pubs, ms) {
		return nil
	}
	return errors.New("bls signature mismatch")
}

// ConvertToSignature convert to BLS Signature
func ConvertToSignature(sig crypto.Signature) (*g1pubs.Signature, error) {
	// unwrap if needed
	if wrap, ok := sig.(SignatureS); ok {
		sig = wrap.Signature
	}
	sigBLS, ok := sig.(SignatureBLS)
	if !ok {
		return nil, errors.New("invalid bls signature")
	}
	g1sig, err := g1pubs.DeserializeSignature(sigBLS)
	if err != nil {
		return nil, err
	}
	return g1sig, nil
}

// ConvertToPublicKey convert to BLS PublicKey
func ConvertToPublicKey(pub crypto.PubKey) (*g1pubs.PublicKey, error) {
	pubBLS, ok := pub.(PubKeyBLS)
	if !ok {
		return nil, errors.New("invalid bls public key")
	}
	g1pub, err := g1pubs.DeserializePublicKey(pubBLS)
	if err != nil {
		return nil, err
	}
	return g1pub, nil
}

// PrivKeyBLS PrivKey
type PrivKeyBLS [BLSPrivateKeyLength]byte

// Bytes convert to bytes
func (privKey PrivKeyBLS) Bytes() []byte {
	s := make([]byte, BLSPrivateKeyLength)
	copy(s, privKey[:])
	return s
}

// Sign create signature
func (privKey PrivKeyBLS) Sign(msg []byte) crypto.Signature {
	priv := g1pubs.DeserializeSecretKey(privKey)
	sig := g1pubs.Sign(msg, priv)
	return SignatureBLS(sig.Serialize())
}

// PubKey convert to public key
func (privKey PrivKeyBLS) PubKey() crypto.PubKey {
	priv := g1pubs.DeserializeSecretKey(privKey)
	return PubKeyBLS(g1pubs.PrivToPub(priv).Serialize())
}

// Equals check privkey is equal
func (privKey PrivKeyBLS) Equals(other crypto.PrivKey) bool {
	if otherSecp, ok := other.(PrivKeyBLS); ok {
		return bytes.Equal(privKey[:], otherSecp[:])
	}
	return false
}

// String convert to string
func (privKey PrivKeyBLS) String() string {
	return fmt.Sprintf("PrivKeyBLS{*****}")
}

// PubKeyBLS PubKey
type PubKeyBLS [BLSPublicKeyLength]byte

// Bytes convert to bytes
func (pubKey PubKeyBLS) Bytes() []byte {
	s := make([]byte, BLSPublicKeyLength)
	copy(s, pubKey[:])
	return s
}

// VerifyBytes verify signature
func (pubKey PubKeyBLS) VerifyBytes(msg []byte, sig crypto.Signature) bool {
	pub, err := g1pubs.DeserializePublicKey(pubKey)
	if err != nil {
		fmt.Println("invalid bls pubkey")
		return false
	}

	g1sig, err := ConvertToSignature(sig)
	if err != nil {
		fmt.Println("ConvertToSignature fail:", err)
		return false
	}

	return g1pubs.Verify(msg, pub, g1sig)
}

// String convert to string
func (pubKey PubKeyBLS) String() string {
	return fmt.Sprintf("PubKeyBLS{%X}", pubKey[:])
}

// KeyString Must return the full bytes in hex.
// Used for map keying, etc.
func (pubKey PubKeyBLS) KeyString() string {
	return fmt.Sprintf("%X", pubKey[:])
}

// Equals check public key is equal
func (pubKey PubKeyBLS) Equals(other crypto.PubKey) bool {
	if otherSecp, ok := other.(PubKeyBLS); ok {
		return bytes.Equal(pubKey[:], otherSecp[:])
	}
	return false
}

// SignatureBLS Signature
type SignatureBLS [BLSSignatureLength]byte

// SignatureS signature struct
type SignatureS struct {
	crypto.Signature
}

// Bytes convert signature to bytes
func (sig SignatureBLS) Bytes() []byte {
	s := make([]byte, len(sig))
	copy(s, sig[:])
	return s
}

// IsZero check signature is zero
func (sig SignatureBLS) IsZero() bool { return len(sig) == 0 }

// String convert signature to string
func (sig SignatureBLS) String() string {
	fingerprint := make([]byte, len(sig[:]))
	copy(fingerprint, sig[:])
	return fmt.Sprintf("/%X.../", fingerprint)

}

// Equals check signature equals
func (sig SignatureBLS) Equals(other crypto.Signature) bool {
	if otherEd, ok := other.(SignatureBLS); ok {
		return bytes.Equal(sig[:], otherEd[:])
	}
	return false
}

// Name name
const Name = "bls"

// ID id
const ID = 259

func init() {
	crypto.Register(Name, &Driver{}, crypto.WithRegOptionTypeID(ID))
}
