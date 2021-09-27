// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
基于框架中Crypto接口，实现签名、验证的处理
*/

package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"

	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

const (
	MixSignName   = "mixZkSnark"
	MixSignID     = 6
	publicKeyLen  = 32
	privateKeyLen = 32
)

func init() {
	crypto.Register(MixSignName, &MixSignZkSnark{}, crypto.WithRegOptionTypeID(MixSignID))
}

// MixSignature mix签名中对于crypto.Signature接口实现
type MixSignature struct {
	sign []byte
}

// Bytes convert to bytest
func (r *MixSignature) Bytes() []byte {
	return r.sign[:]
}

// IsZero check is zero
func (r *MixSignature) IsZero() bool {
	return len(r.sign) == 0
}

// String convert to string
func (r *MixSignature) String() string {
	return hex.EncodeToString(r.sign)
}

// Equals check equals
func (r *MixSignature) Equals(other crypto.Signature) bool {
	if _, ok := other.(*MixSignature); ok {
		return bytes.Equal(r.Bytes(), other.Bytes())
	}
	return false
}

// RingSignPrivateKey 环签名中对于crypto.PrivKey接口实现
type MixSignPrivateKey struct {
	key [publicKeyLen]byte
}

// Bytes convert key to bytest
func (privkey *MixSignPrivateKey) Bytes() []byte {
	return privkey.key[:]
}

// Sign signature trasaction
func (privkey *MixSignPrivateKey) Sign(msg []byte) crypto.Signature {
	return &MixSignature{}
}

// PubKey convert to public key
func (privkey *MixSignPrivateKey) PubKey() crypto.PubKey {
	publicKey := new(MixSignPublicKey)
	return publicKey
}

// Equals check key equal
func (privkey *MixSignPrivateKey) Equals(other crypto.PrivKey) bool {
	if otherPrivKey, ok := other.(*MixSignPrivateKey); ok {
		return bytes.Equal(privkey.key[:], otherPrivKey.key[:])
	}
	return false
}

// RingSignPublicKey 环签名中对于crypto.PubKey接口实现
type MixSignPublicKey struct {
	key [publicKeyLen]byte
}

// Bytes convert key to bytes
func (pubkey *MixSignPublicKey) Bytes() []byte {
	return pubkey.key[:]
}

//
//func verifyCommitAmount(transfer *mixTy.MixTransferAction) error {
//	var inputs []*mixTy.TransferInputPublicInput
//	var outputs []*mixTy.TransferOutputPublicInput
//
//	for _, k := range transfer.Input {
//		v, err := mixTy.DecodePubInput(mixTy.VerifyType_TRANSFERINPUT, k.PublicInput)
//		if err != nil {
//			return errors.Wrap(types.ErrInvalidParam, "decode transfer Input")
//		}
//		inputs = append(inputs, v.(*mixTy.TransferInputPublicInput))
//	}
//
//	for _, k := range transfer.Output {
//		v, err := mixTy.DecodePubInput(mixTy.VerifyType_TRANSFEROUTPUT, k.PublicInput)
//		if err != nil {
//			return errors.Wrap(types.ErrInvalidParam, "decode transfer output")
//		}
//		outputs = append(outputs, v.(*mixTy.TransferOutputPublicInput))
//	}
//
//	if !mixExec.VerifyCommitValues(inputs, outputs) {
//		return errors.Wrap(types.ErrInvalidParam, "verify commit amount")
//	}
//	return nil
//}

// VerifyBytes verify bytes
func (pubkey *MixSignPublicKey) VerifyBytes(msg []byte, sign crypto.Signature) bool {
	if len(msg) <= 0 {
		return false
	}

	tx := new(types.Transaction)
	if err := types.Decode(msg, tx); err != nil || !bytes.Equal([]byte(mixTy.MixX), types.GetRealExecName(tx.Execer)) {
		// mix特定执行器的签名
		bizlog.Error("pubkey.VerifyBytes", "err", err, "exec", string(types.GetRealExecName(tx.Execer)))
		return false
	}
	action := new(mixTy.MixAction)
	if err := types.Decode(tx.Payload, action); err != nil {
		bizlog.Error("pubkey.VerifyBytes decode tx")
		return false
	}
	if action.Ty != mixTy.MixActionTransfer {
		// mix隐私交易，只私对私需要特殊签名验证
		bizlog.Error("pubkey.VerifyBytes", "ty", action.Ty)
		return false
	}

	//确保签名数据和tx 一致
	if !bytes.Equal(sign.Bytes(), common.BytesToHash(types.Encode(action.GetTransfer())).Bytes()) {
		bizlog.Error("pubkey.VerifyBytes tx and sign not match", "sign", common.ToHex(sign.Bytes()), "tx", common.ToHex(common.BytesToHash(types.Encode(action.GetTransfer())).Bytes()))
		return false
	}

	return true
}

// KeyString convert  key to string
func (pubkey *MixSignPublicKey) KeyString() string {
	return fmt.Sprintf("%X", pubkey.key[:])
}

// Equals check key is equal
func (pubkey *MixSignPublicKey) Equals(other crypto.PubKey) bool {
	if otherPubKey, ok := other.(*MixSignPublicKey); ok {
		return bytes.Equal(pubkey.key[:], otherPubKey.key[:])
	}
	return false
}

// MixSignZkSnark 对应crypto.Crypto的接口实现
type MixSignZkSnark struct {
}

// GenKey create privacy key
func (r *MixSignZkSnark) GenKey() (crypto.PrivKey, error) {
	priKey := new(MixSignPrivateKey)
	return priKey, nil

}

// PrivKeyFromBytes create private key from bytes
func (r *MixSignZkSnark) PrivKeyFromBytes(b []byte) (crypto.PrivKey, error) {
	if len(b) <= 0 {
		return nil, types.ErrInvalidParam
	}
	if len(b) != privateKeyLen {
		return nil, types.ErrPrivateKeyLen
	}
	privateKey := new(MixSignPrivateKey)
	copy(privateKey.key[:], b)
	return privateKey, nil
}

// PubKeyFromBytes create publick key from bytes
func (r *MixSignZkSnark) PubKeyFromBytes(b []byte) (crypto.PubKey, error) {
	if len(b) <= 0 {
		return nil, types.ErrInvalidParam
	}
	if len(b) != publicKeyLen {
		return nil, types.ErrPubKeyLen
	}
	publicKey := new(MixSignPublicKey)
	copy(publicKey.key[:], b)
	return publicKey, nil
}

// SignatureFromBytes create signature from bytes
func (r *MixSignZkSnark) SignatureFromBytes(b []byte) (crypto.Signature, error) {
	if len(b) <= 0 {
		return nil, types.ErrInvalidParam
	}

	var mixSig MixSignature
	mixSig.sign = append(mixSig.sign, b...)

	return &mixSig, nil
}

// Validate validate msg and signature
func (r *MixSignZkSnark) Validate(msg, pub, sig []byte) error {
	return crypto.BasicValidation(r, msg, pub, sig)
}
