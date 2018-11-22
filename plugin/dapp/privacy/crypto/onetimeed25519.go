// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package privacy

import (
	"errors"
	"unsafe"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/ed25519/edwards25519"
	privacytypes "github.com/33cn/plugin/plugin/dapp/privacy/types"
)

type oneTimeEd25519 struct{}

func init() {
	crypto.Register(privacytypes.SignNameOnetimeED25519, &oneTimeEd25519{})
}

func (onetime *oneTimeEd25519) GenKey() (crypto.PrivKey, error) {
	privKeyPrivacyPtr := &PrivKeyPrivacy{}
	pubKeyPrivacyPtr := &PubKeyPrivacy{}
	copy(privKeyPrivacyPtr[:privateKeyLen], crypto.CRandBytes(privateKeyLen))

	addr32 := (*[KeyLen32]byte)(unsafe.Pointer(privKeyPrivacyPtr))
	addr64 := (*[privateKeyLen]byte)(unsafe.Pointer(privKeyPrivacyPtr))
	edwards25519.ScReduce(addr32, addr64)

	//to generate the publickey
	var A edwards25519.ExtendedGroupElement
	pubKeyAddr32 := (*[KeyLen32]byte)(unsafe.Pointer(pubKeyPrivacyPtr))
	edwards25519.GeScalarMultBase(&A, addr32)
	A.ToBytes(pubKeyAddr32)
	copy(addr64[KeyLen32:], pubKeyAddr32[:])

	return *privKeyPrivacyPtr, nil
}

func (onetime *oneTimeEd25519) PrivKeyFromBytes(b []byte) (privKey crypto.PrivKey, err error) {
	if len(b) != 64 {
		return nil, errors.New("invalid priv key byte")
	}
	privKeyBytes := new([privateKeyLen]byte)
	pubKeyBytes := new([publicKeyLen]byte)
	copy(privKeyBytes[:KeyLen32], b[:KeyLen32])

	addr32 := (*[KeyLen32]byte)(unsafe.Pointer(privKeyBytes))
	addr64 := (*[privateKeyLen]byte)(unsafe.Pointer(privKeyBytes))

	//to generate the publickey
	var A edwards25519.ExtendedGroupElement
	pubKeyAddr32 := (*[KeyLen32]byte)(unsafe.Pointer(pubKeyBytes))
	edwards25519.GeScalarMultBase(&A, addr32)
	A.ToBytes(pubKeyAddr32)
	copy(addr64[KeyLen32:], pubKeyAddr32[:])

	return PrivKeyPrivacy(*privKeyBytes), nil
}

func (onetime *oneTimeEd25519) PubKeyFromBytes(b []byte) (pubKey crypto.PubKey, err error) {
	if len(b) != 32 {
		return nil, errors.New("invalid pub key byte")
	}
	pubKeyBytes := new([32]byte)
	copy(pubKeyBytes[:], b[:])
	return PubKeyPrivacy(*pubKeyBytes), nil
}

func (onetime *oneTimeEd25519) SignatureFromBytes(b []byte) (sig crypto.Signature, err error) {
	sigBytes := new([64]byte)
	copy(sigBytes[:], b[:])
	return SignatureOnetime(*sigBytes), nil
}
