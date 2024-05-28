// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/33cn/chain33/common/address"

	"github.com/33cn/chain33/common/crypto"
	vrf "github.com/33cn/chain33/common/vrf/secp256k1"
	"github.com/33cn/chain33/types"
	secp256k1 "github.com/btcsuite/btcd/btcec"
)

// KeyText ...
type KeyText struct {
	Kind string `json:"type"`
	Data string `json:"data"`
}

// PrivValidator defines the functionality of a local Tendermint validator
// that signs votes, proposals, and heartbeats, and never double signs.
type PrivValidator interface {
	GetAddress() []byte // redundant since .PubKey().Address()
	GetPubKey() crypto.PubKey

	SignVote(chainID string, vote *Vote) error
	SignNotify(chainID string, notify *Notify) error
	SignMsg(msg []byte) (sig crypto.Signature, err error)
	SignTx(tx *types.Transaction)
	VrfEvaluate(input []byte) (hash [32]byte, proof []byte)
	VrfProof(pubkey []byte, input []byte, hash [32]byte, proof []byte) bool
}

// PrivValidatorFS implements PrivValidator using data persisted to disk
// to prevent double signing. The Signer itself can be mutated to use
// something besides the default, for instance a hardware signer.
type PrivValidatorFS struct {
	Address string  `json:"address"`
	PubKey  KeyText `json:"pub_key"`
	//LastSignature *KeyText `json:"last_signature,omitempty"` // so we dont lose signatures
	//LastSignBytes string   `json:"last_signbytes,omitempty"` // so we dont lose signatures

	// PrivKey should be empty if a Signer other than the default is being used.
	PrivKey KeyText `json:"priv_key"`
}

// PrivValidatorImp ...
type PrivValidatorImp struct {
	Address []byte
	PubKey  crypto.PubKey
	//LastSignature crypto.Signature
	//LastSignBytes []byte

	// PrivKey should be empty if a Signer other than the default is being used.
	PrivKey crypto.PrivKey
	Signer  `json:"-"`

	// For persistence.
	// Overloaded for testing.
	filePath string
	mtx      sync.Mutex
}

// Signer is an interface that defines how to sign messages.
// It is the caller's duty to verify the msg before calling Sign,
// eg. to avoid double signing.
// Currently, the only callers are SignVote, SignProposal, and SignHeartbeat.
type Signer interface {
	Sign(msg []byte) (crypto.Signature, error)
}

// DefaultSigner implements Signer.
// It uses a standard, unencrypted crypto.PrivKey.
type DefaultSigner struct {
	PrivKey crypto.PrivKey `json:"priv_key"`
}

// NewDefaultSigner returns an instance of DefaultSigner.
func NewDefaultSigner(priv crypto.PrivKey) *DefaultSigner {
	return &DefaultSigner{
		PrivKey: priv,
	}
}

// Sign implements Signer. It signs the byte slice with a private key.
func (ds *DefaultSigner) Sign(msg []byte) (crypto.Signature, error) {
	return ds.PrivKey.Sign(msg), nil
}

// GetAddress returns the address of the validator.
// Implements PrivValidator.
func (pv *PrivValidatorImp) GetAddress() []byte {
	return pv.Address
}

// GetPubKey returns the public key of the validator.
// Implements PrivValidator.
func (pv *PrivValidatorImp) GetPubKey() crypto.PubKey {
	return pv.PubKey
}

// GenAddressByPubKey ...
func GenAddressByPubKey(pubkey crypto.PubKey) []byte {
	//must add 3 bytes ahead to make compatibly
	typeAddr := append([]byte{byte(0x01), byte(0x01), byte(0x20)}, pubkey.Bytes()...)
	return crypto.Ripemd160(typeAddr)
}

// PubKeyFromString ...
func PubKeyFromString(pubkeystring string) (crypto.PubKey, error) {
	pub, err := hex.DecodeString(pubkeystring)
	if err != nil {
		return nil, errors.New(Fmt("PubKeyFromString:DecodeString:%v failed,err:%v", pubkeystring, err))
	}

	pubkey, err := ConsensusCrypto.PubKeyFromBytes(pub)
	if err != nil {
		return nil, errors.New(Fmt("PubKeyFromString:PubKeyFromBytes:%v failed,err:%v", pub, err))
	}
	return pubkey, nil
}

// GenPrivValidatorImp generates a new validator with randomly generated private key and sets the filePath, but does not call Save().
func GenPrivValidatorImp(filePath string) *PrivValidatorImp {
	privKey, err := ConsensusCrypto.GenKey()
	if err != nil {
		panic(Fmt("GenPrivValidatorImp: GenKey failed:%v", err))
	}
	return &PrivValidatorImp{
		//Address:  GenAddressByPubKey(privKey.PubKey()),
		Address:  address.BytesToBtcAddress(address.NormalVer, privKey.PubKey().Bytes()).Hash160[:],
		PubKey:   privKey.PubKey(),
		PrivKey:  privKey,
		Signer:   NewDefaultSigner(privKey),
		filePath: filePath,
	}
}

// LoadPrivValidatorFS loads a PrivValidatorImp from the filePath.
func LoadPrivValidatorFS(filePath string) *PrivValidatorImp {
	return LoadPrivValidatorFSWithSigner(filePath, func(privVal PrivValidator) Signer {
		return NewDefaultSigner(privVal.(*PrivValidatorImp).PrivKey)
	})
}

// LoadOrGenPrivValidatorFS loads a PrivValidatorFS from the given filePath
// or else generates a new one and saves it to the filePath.
func LoadOrGenPrivValidatorFS(filePath string) *PrivValidatorImp {
	var privVal *PrivValidatorImp
	if _, err := os.Stat(filePath); err == nil {
		privVal = LoadPrivValidatorFS(filePath)
	} else {
		privVal = GenPrivValidatorImp(filePath)
		privVal.Save()
	}
	return privVal
}

// LoadPrivValidatorFSWithSigner loads a PrivValidatorFS with a custom
// signer object. The PrivValidatorFS handles double signing prevention by persisting
// data to the filePath, while the Signer handles the signing.
// If the filePath does not exist, the PrivValidatorFS must be created manually and saved.
func LoadPrivValidatorFSWithSigner(filePath string, signerFunc func(PrivValidator) Signer) *PrivValidatorImp {
	privValJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		Exit(err.Error())
	}
	privVal := &PrivValidatorFS{}
	err = json.Unmarshal(privValJSONBytes, &privVal)
	if err != nil {
		Exit(Fmt("Error reading PrivValidator from %v: %v\n", filePath, err))
	}
	if len(privVal.PubKey.Data) == 0 {
		Exit("Error PrivValidator pubkey is empty\n")
	}
	if len(privVal.PrivKey.Data) == 0 {
		Exit("Error PrivValidator privkey is empty\n")
	}
	addr, err := hex.DecodeString(privVal.Address)
	if err != nil {
		Exit(Fmt("Error PrivValidator DecodeString failed:%v\n", err))
	}
	privValImp := &PrivValidatorImp{
		Address: addr,
	}
	tmp, err := hex.DecodeString(privVal.PrivKey.Data)
	if err != nil {
		Exit(Fmt("Error DecodeString PrivKey data failed: %v\n", err))
	}
	privKey, err := ConsensusCrypto.PrivKeyFromBytes(tmp)
	if err != nil {
		Exit(Fmt("Error PrivKeyFromBytes failed: %v\n", err))
	}
	privValImp.PrivKey = privKey

	pubKey, err := PubKeyFromString(privVal.PubKey.Data)
	if err != nil {
		Exit(Fmt("Error PubKeyFromBytes failed: %v\n", err))
	}
	privValImp.PubKey = pubKey

	/*
		if len(privVal.LastSignBytes) != 0 {
			tmp, err = hex.DecodeString(privVal.LastSignBytes)
			if err != nil {
				Exit(Fmt("Error DecodeString LastSignBytes data failed: %v\n", err))
			}
			privValImp.LastSignBytes = tmp
		}
		if privVal.LastSignature != nil {
			signature, err := SignatureFromString(privVal.LastSignature.Data)
			if err != nil {
				Exit(Fmt("Error SignatureFromBytes failed: %v\n", err))
			}
			privValImp.LastSignature = signature
		} else {
			privValImp.LastSignature = nil
		}
	*/
	privValImp.filePath = filePath
	privValImp.Signer = signerFunc(privValImp)
	return privValImp
}

// Save persists the PrivValidatorFS to disk.
func (pv *PrivValidatorImp) Save() {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()
	pv.save()
}

func (pv *PrivValidatorImp) save() {
	if pv.filePath == "" {
		PanicSanity("Cannot save PrivValidator: filePath not set")
	}
	addr := Fmt("%X", pv.Address[:])

	privValFS := &PrivValidatorFS{
		Address: addr,
		//LastSignature: nil,
	}
	privValFS.PrivKey = KeyText{Kind: "secp256k1", Data: Fmt("%X", pv.PrivKey.Bytes()[:])}
	privValFS.PubKey = KeyText{Kind: "secp256k1", Data: pv.PubKey.KeyString()}
	/*
		if len(pv.LastSignBytes) != 0 {
			tmp := Fmt("%X", pv.LastSignBytes[:])
			privValFS.LastSignBytes = tmp
		}
		if pv.LastSignature != nil {
			sig := Fmt("%X", pv.LastSignature.Bytes()[:])
			privValFS.LastSignature = &KeyText{Kind: "ed25519", Data: sig}
		}
	*/
	jsonBytes, err := json.Marshal(privValFS)
	if err != nil {
		// `@; BOOM!!!
		PanicCrisis(err)
	}
	err = WriteFileAtomic(pv.filePath, jsonBytes, 0600)
	if err != nil {
		// `@; BOOM!!!
		PanicCrisis(err)
	}
}

// Reset resets all fields in the PrivValidatorFS.
// NOTE: Unsafe!
func (pv *PrivValidatorImp) Reset() {
	//pv.LastSignature = nil
	//pv.LastSignBytes = nil
	pv.Save()
}

// SignVote signs a canonical representation of the vote, along with the
// chainID. Implements PrivValidator.
func (pv *PrivValidatorImp) SignVote(chainID string, vote *Vote) error {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()

	signBytes := SignBytes(chainID, vote)

	signature, err := pv.Sign(signBytes)
	if err != nil {
		return errors.New(Fmt("Error signing vote: %v", err))
	}
	vote.Signature = signature.Bytes()
	return nil
}

// SignNotify signs a canonical representation of the notify, along with the
// chainID. Implements PrivValidator.
func (pv *PrivValidatorImp) SignNotify(chainID string, notify *Notify) error {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()

	signBytes := SignBytes(chainID, notify)

	signature, err := pv.Sign(signBytes)
	if err != nil {
		return errors.New(Fmt("Error signing vote: %v", err))
	}
	notify.Signature = signature.Bytes()
	return nil
}

// SignMsg signs a msg.
func (pv *PrivValidatorImp) SignMsg(msg []byte) (sig crypto.Signature, err error) {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()

	signature := pv.PrivKey.Sign(msg)
	//sig = hex.EncodeToString(signature.Bytes())
	return signature, nil
}

// SignTx signs a tx, Implements PrivValidator.
func (pv *PrivValidatorImp) SignTx(tx *types.Transaction) {
	tx.Sign(types.SECP256K1, pv.PrivKey)
}

// VrfEvaluate use input to generate hash & proof.
func (pv *PrivValidatorImp) VrfEvaluate(input []byte) (hash [32]byte, proof []byte) {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()

	privKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), pv.PrivKey.Bytes())
	vrfPriv := &vrf.PrivateKey{PrivateKey: (*ecdsa.PrivateKey)(privKey)}
	hash, proof = vrfPriv.Evaluate(input)
	return hash, proof
}

// VrfProof check the vrf.
func (pv *PrivValidatorImp) VrfProof(pubkey []byte, input []byte, hash [32]byte, proof []byte) bool {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()

	pubKey, err := secp256k1.ParsePubKey(pubkey, secp256k1.S256())
	if err != nil {
		return false
	}
	vrfPub := &vrf.PublicKey{PublicKey: (*ecdsa.PublicKey)(pubKey)}
	vrfHash, err := vrfPub.ProofToHash(input, proof)
	if err != nil {
		return false
	}
	if bytes.Equal(hash[:], vrfHash[:]) {
		return true
	}

	return false
}

// Persist height/round/step and signature
/*
func (pv *PrivValidatorImp) saveSigned(signBytes []byte, sig crypto.Signature) {

	//pv.LastSignature = sig
	//pv.LastSignBytes = signBytes
	pv.save()
}
*/

// String returns a string representation of the PrivValidatorImp.
func (pv *PrivValidatorImp) String() string {
	return Fmt("PrivValidator{%v}", pv.GetAddress())
}

// PrivValidatorsByAddress ...
type PrivValidatorsByAddress []*PrivValidatorImp

func (pvs PrivValidatorsByAddress) Len() int {
	return len(pvs)
}

func (pvs PrivValidatorsByAddress) Less(i, j int) bool {
	return bytes.Compare(pvs[i].GetAddress(), pvs[j].GetAddress()) == -1
}

func (pvs PrivValidatorsByAddress) Swap(i, j int) {
	it := pvs[i]
	pvs[i] = pvs[j]
	pvs[j] = it
}
