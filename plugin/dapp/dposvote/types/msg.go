package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
)

// CanonicalOnceCBInfo ...
type CanonicalOnceCBInfo struct {
	Cycle      int64  `json:"cycle,omitempty"`
	StopHeight int64  `json:"stopHeight,omitempty"`
	StopHash   string `json:"stopHash,omitempty"`
	Pubkey     string `json:"pubkey,omitempty"`
}

// CanonicalCBInfo ...
func CanonicalCBInfo(cb *DposCBInfo) CanonicalOnceCBInfo {
	return CanonicalOnceCBInfo{
		Cycle: cb.Cycle,
		StopHeight: cb.StopHeight,
		StopHash: cb.StopHash,
		Pubkey: cb.Pubkey,
	}
}


// CanonicalCBInfo ...
func (cb *DposCBInfo)Verify() error {
	buf := new(bytes.Buffer)

	canonical := CanonicalOnceCBInfo{
		Cycle: cb.Cycle,
		StopHeight: cb.StopHeight,
		StopHash: cb.StopHash,
		Pubkey: cb.Pubkey,
	}

	byteCB, err := json.Marshal(&canonical)
	if err != nil {
		return errors.New(fmt.Sprintf("Error marshal CanonicalOnceCBInfo: %v", err))
	}

	_, err = buf.Write(byteCB)
	if err != nil {
		return errors.New(fmt.Sprintf("Error write buffer: %v", err))
	}

	bPubkey, err := hex.DecodeString(cb.Pubkey)
	if err != nil {
		return errors.New(fmt.Sprintf("Error Decode pubkey: %v", err))
	}
	pubkey, err := ttypes.ConsensusCrypto.PubKeyFromBytes(bPubkey)
	if err != nil {
		return errors.New(fmt.Sprintf("Error PubKeyFromBytes: %v", err))
	}

	signature, err := hex.DecodeString(cb.Signature)
	if err != nil {
		return errors.New(fmt.Sprintf("Error Decode Signature: %v", err))
	}

	sig, err := ttypes.ConsensusCrypto.SignatureFromBytes(signature)
	if err != nil {
		return errors.New(fmt.Sprintf("Error SignatureFromBytes: %v", err))
	}

	if !pubkey.VerifyBytes(buf.Bytes(), sig) {
		return errors.New(fmt.Sprintf("Error VerifyBytes: %v", err))
	}

	return nil
}