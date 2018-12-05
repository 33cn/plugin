// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package privacy

import (
	"bytes"
	"unsafe"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/crypto/sha3"
	"github.com/33cn/chain33/common/ed25519/edwards25519"
)

// PrivKeyPrivacy struct data type
type PrivKeyPrivacy [privateKeyLen]byte

// Bytes convert to bytes
func (privKey PrivKeyPrivacy) Bytes() []byte {
	return privKey[:]
}

// Sign signature trasaction
func (privKey PrivKeyPrivacy) Sign(msg []byte) crypto.Signature {

	temp := new([64]byte)
	randomScalar := new([32]byte)
	copy(temp[:], crypto.CRandBytes(64))
	edwards25519.ScReduce(randomScalar, temp)

	var sigcommdata sigcommArray
	sigcommPtr := (*sigcomm)(unsafe.Pointer(&sigcommdata))
	copy(sigcommPtr.pubkey[:], privKey.PubKey().Bytes())
	hash := sha3.Sum256(msg)
	copy(sigcommPtr.hash[:], hash[:])

	var K edwards25519.ExtendedGroupElement
	edwards25519.GeScalarMultBase(&K, randomScalar)
	K.ToBytes((*[KeyLen32]byte)(unsafe.Pointer(&sigcommPtr.comm[0])))

	var sigOnetime SignatureOnetime
	addr32 := (*[KeyLen32]byte)(unsafe.Pointer(&sigOnetime))
	hash2scalar(sigcommdata[:], addr32)

	addr32Latter := (*[KeyLen32]byte)(unsafe.Pointer(&sigOnetime[KeyLen32]))
	addr32Priv := (*[KeyLen32]byte)(unsafe.Pointer(&privKey))
	edwards25519.ScMulSub(addr32Latter, addr32, addr32Priv, randomScalar)

	return sigOnetime
}

// PubKey get public key
func (privKey PrivKeyPrivacy) PubKey() crypto.PubKey {

	var pubKeyPrivacy PubKeyPrivacy

	addr32 := (*[KeyLen32]byte)(unsafe.Pointer(&privKey.Bytes()[0]))
	addr64 := (*[privateKeyLen]byte)(unsafe.Pointer(&privKey.Bytes()[0]))

	var A edwards25519.ExtendedGroupElement
	pubKeyAddr32 := (*[KeyLen32]byte)(unsafe.Pointer(&pubKeyPrivacy))
	edwards25519.GeScalarMultBase(&A, addr32)
	A.ToBytes(pubKeyAddr32)
	copy(addr64[KeyLen32:], pubKeyAddr32[:])

	return pubKeyPrivacy
}

// Equals check equals
func (privKey PrivKeyPrivacy) Equals(other crypto.PrivKey) bool {
	if otherEd, ok := other.(PrivKeyPrivacy); ok {
		return bytes.Equal(privKey[:], otherEd[:])
	}
	return false
}
