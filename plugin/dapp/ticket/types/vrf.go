package types

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"math/big"

	"github.com/33cn/chain33/common/crypto"
	"github.com/google/keytransparency/core/crypto/vrf"
	"github.com/google/keytransparency/core/crypto/vrf/p256"
)

const publicKeyECDSALength = 65

// GenVrfKey returns vrf private and public key
func GenVrfKey(key crypto.PrivKey) (vrf.PrivateKey, vrf.PublicKey, []byte, error) {
	priv, pub := PrivKeyFromBytes(elliptic.P256(), key.Bytes())
	return &p256.PrivateKey{PrivateKey: priv}, &p256.PublicKey{PublicKey: pub}, SerializePublicKey(pub), nil
}

// PrivKeyFromBytes return ecdsa private and public key
func PrivKeyFromBytes(curve elliptic.Curve, pk []byte) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
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

// SerializePublicKey serialize public key
func SerializePublicKey(p *ecdsa.PublicKey) []byte {
	b := make([]byte, 0, publicKeyECDSALength)
	b = append(b, 0x4)
	b = paddedAppend(32, b, p.X.Bytes())
	return paddedAppend(32, b, p.Y.Bytes())
}

func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}

// ParseVrfPubKey parse public key
func ParseVrfPubKey(pubKeyStr []byte) (vrf.PublicKey, error) {
	pubkey := &ecdsa.PublicKey{}
	pubkey.Curve = elliptic.P256()

	if len(pubKeyStr) == 0 {
		return nil, errors.New("pubkey string is empty")
	}

	pubkey.X = new(big.Int).SetBytes(pubKeyStr[1:33])
	pubkey.Y = new(big.Int).SetBytes(pubKeyStr[33:])
	if pubkey.X.Cmp(pubkey.Curve.Params().P) >= 0 {
		return nil, errors.New("pubkey X parameter is >= to P")
	}
	if pubkey.Y.Cmp(pubkey.Curve.Params().P) >= 0 {
		return nil, errors.New("pubkey Y parameter is >= to P")
	}
	if !pubkey.Curve.IsOnCurve(pubkey.X, pubkey.Y) {
		return nil, errors.New("pubkey isn't on secp256k1 curve")
	}
	return &p256.PublicKey{PublicKey: pubkey}, nil
}
