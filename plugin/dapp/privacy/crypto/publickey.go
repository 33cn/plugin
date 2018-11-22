// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package privacy

import (
	"bytes"
	"fmt"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	privacytypes "github.com/33cn/plugin/plugin/dapp/privacy/types"
)

// PubKeyPrivacy key struct types
type PubKeyPrivacy [publicKeyLen]byte

// Bytes convert to bytes
func (pubKey PubKeyPrivacy) Bytes() []byte {
	return pubKey[:]
}

// Bytes2PubKeyPrivacy convert bytes to PubKeyPrivacy
func Bytes2PubKeyPrivacy(in []byte) PubKeyPrivacy {
	var temp PubKeyPrivacy
	copy(temp[:], in)
	return temp
}

// VerifyBytes verify bytes
func (pubKey PubKeyPrivacy) VerifyBytes(msg []byte, sig crypto.Signature) bool {
	var tx types.Transaction
	if err := types.Decode(msg, &tx); err != nil {
		return false
	}
	if privacytypes.PrivacyX != string(tx.Execer) {
		return false
	}
	var action privacytypes.PrivacyAction
	if err := types.Decode(tx.Payload, &action); err != nil {
		return false
	}
	var privacyInput *privacytypes.PrivacyInput
	if action.Ty == privacytypes.ActionPrivacy2Privacy && action.GetPrivacy2Privacy() != nil {
		privacyInput = action.GetPrivacy2Privacy().Input
	} else if action.Ty == privacytypes.ActionPrivacy2Public && action.GetPrivacy2Public() != nil {
		privacyInput = action.GetPrivacy2Public().Input
	} else {
		return false
	}
	var ringSign types.RingSignature
	if err := types.Decode(sig.Bytes(), &ringSign); err != nil {
		return false
	}

	h := common.BytesToHash(msg)
	for i, ringSignItem := range ringSign.GetItems() {
		if !CheckRingSignature(h.Bytes(), ringSignItem, ringSignItem.Pubkey, privacyInput.Keyinput[i].KeyImage) {
			return false
		}
	}
	return true
}

//func (pubKey PubKeyPrivacy) VerifyBytes(msg []byte, sig_ Signature) bool {
//	// unwrap if needed
//	if wrap, ok := sig_.(SignatureS); ok {
//		sig_ = wrap.Signature
//	}
//	// make sure we use the same algorithm to sign
//	sig, ok := sig_.(SignatureOnetime)
//	if !ok {
//		return false
//	}
//	pubKeyBytes := [32]byte(pubKey)
//	sigBytes := [64]byte(sig)
//
//	var ege edwards25519.ExtendedGroupElement
//	if !ege.FromBytes(&pubKeyBytes) {
//		return false
//	}
//
//	sigAddr32a := (*[KeyLen32]byte)(unsafe.Pointer(&sigBytes[0]))
//	sigAddr32b := (*[KeyLen32]byte)(unsafe.Pointer(&sigBytes[32]))
//	if !edwards25519.ScCheck(sigAddr32a) || !edwards25519.ScCheck(sigAddr32b) {
//		return false
//	}
//
//	var sigcommdata sigcommArray
//	sigcommPtr := (*sigcomm)(unsafe.Pointer(&sigcommdata))
//	copy(sigcommPtr.pubkey[:], pubKey.Bytes())
//	hash := sha3.Sum256(msg)
//	copy(sigcommPtr.hash[:], hash[:])
//
//	var rge edwards25519.ProjectiveGroupElement
//	edwards25519.GeDoubleScalarMultVartime(&rge, sigAddr32a, &ege, sigAddr32b)
//	rge.ToBytes((*[KeyLen32]byte)(unsafe.Pointer(&sigcommPtr.comm[0])))
//
//	out32 := new([32]byte)
//	hash2scalar(sigcommdata[:], out32)
//	return subtle.ConstantTimeCompare(sigAddr32a[:], out32[:]) == 1
//}

// KeyString convert to string
func (pubKey PubKeyPrivacy) KeyString() string {
	return fmt.Sprintf("%X", pubKey[:])
}

// Equals check equals
func (pubKey PubKeyPrivacy) Equals(other crypto.PubKey) bool {
	if otherEd, ok := other.(PubKeyPrivacy); ok {
		return bytes.Equal(pubKey[:], otherEd[:])
	}
	return false
}
