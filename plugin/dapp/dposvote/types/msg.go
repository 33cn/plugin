package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/33cn/chain33/common/crypto"
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
		Cycle:      cb.Cycle,
		StopHeight: cb.StopHeight,
		StopHash:   cb.StopHash,
		Pubkey:     cb.Pubkey,
	}
}

// Verify ...
func (cb *DposCBInfo) Verify() error {
	buf := new(bytes.Buffer)

	canonical := CanonicalCBInfo(cb)

	byteCB, err := json.Marshal(&canonical)
	if err != nil {
		return fmt.Errorf("Error marshal CanonicalOnceCBInfo: %v", err)
	}

	_, err = buf.Write(byteCB)
	if err != nil {
		return fmt.Errorf("Error write buffer: %v", err)
	}

	bPubkey, err := hex.DecodeString(cb.Pubkey)
	if err != nil {
		return fmt.Errorf("Error Decode pubkey: %v", err)
	}
	pubkey, err := ttypes.ConsensusCrypto.PubKeyFromBytes(bPubkey)
	if err != nil {
		return fmt.Errorf("Error PubKeyFromBytes: %v", err)
	}

	signature, err := hex.DecodeString(cb.Signature)
	if err != nil {
		return fmt.Errorf("Error Decode Signature: %v", err)
	}

	sig, err := ttypes.ConsensusCrypto.SignatureFromBytes(signature)
	if err != nil {
		return fmt.Errorf("Error SignatureFromBytes: %v", err)
	}

	if !pubkey.VerifyBytes(buf.Bytes(), sig) {
		return fmt.Errorf("Error VerifyBytes: %v", err)
	}

	return nil
}

// OnceCandidator ...
type OnceCandidator struct {
	Pubkey  []byte `json:"pubkey,omitempty"`
	Address string `json:"address,omitempty"`
	IP      string `json:"ip,omitempty"`
}

// CanonicalOnceTopNCandidator ...
type CanonicalOnceTopNCandidator struct {
	Cands        []*OnceCandidator `json:"cands,omitempty"`
	Hash         []byte            `json:"hash,omitempty"`
	Height       int64             `json:"height,omitempty"`
	SignerPubkey []byte            `json:"signerPubkey,omitempty"`
	Signature    []byte            `json:"signature,omitempty"`
}

func (topN *CanonicalOnceTopNCandidator) onlyCopyCands() CanonicalOnceTopNCandidator {
	obj := CanonicalOnceTopNCandidator{}
	for i := 0; i < len(topN.Cands); i++ {
		cand := &OnceCandidator{
			Pubkey:  topN.Cands[i].Pubkey,
			Address: topN.Cands[i].Address,
			IP:      topN.Cands[i].IP,
		}
		obj.Cands = append(obj.Cands, cand)
	}

	return obj
}

// ID ...
func (topN *CanonicalOnceTopNCandidator) ID() []byte {
	obj := topN.onlyCopyCands()
	encode, err := json.Marshal(&obj)
	if err != nil {
		return nil
	}
	return crypto.Ripemd160(encode)
}

// CanonicalTopNCandidator ...
func CanonicalTopNCandidator(topN *TopNCandidator) CanonicalOnceTopNCandidator {
	onceTopNCandidator := CanonicalOnceTopNCandidator{
		Height:       topN.Height,
		SignerPubkey: topN.SignerPubkey,
	}

	for i := 0; i < len(topN.Cands); i++ {
		cand := &OnceCandidator{
			Pubkey:  topN.Cands[i].Pubkey,
			Address: topN.Cands[i].Address,
			IP:      topN.Cands[i].IP,
		}
		onceTopNCandidator.Cands = append(onceTopNCandidator.Cands, cand)
	}
	return onceTopNCandidator
}

func (topN *TopNCandidator) copyWithoutSig() *TopNCandidator {
	cpy := &TopNCandidator{
		Hash:         topN.Hash,
		Height:       topN.Height,
		SignerPubkey: topN.SignerPubkey,
	}

	cpy.Cands = make([]*Candidator, len(topN.Cands))
	for i := 0; i < len(topN.Cands); i++ {
		cpy.Cands[i] = topN.Cands[i]
	}
	return cpy
}

// Verify ...
func (topN *TopNCandidator) Verify() error {
	buf := new(bytes.Buffer)

	cpy := topN.copyWithoutSig()
	byteCB, err := json.Marshal(cpy)
	if err != nil {
		return fmt.Errorf("Error marshal TopNCandidator: %v", err)
	}

	_, err = buf.Write(byteCB)
	if err != nil {
		return fmt.Errorf("Error write buffer: %v", err)
	}

	pubkey, err := ttypes.ConsensusCrypto.PubKeyFromBytes(topN.SignerPubkey)
	if err != nil {
		return fmt.Errorf("Error PubKeyFromBytes: %v", err)
	}

	sig, err := ttypes.ConsensusCrypto.SignatureFromBytes(topN.Signature)
	if err != nil {
		return fmt.Errorf("Error SignatureFromBytes: %v", err)
	}

	if !pubkey.VerifyBytes(buf.Bytes(), sig) {
		return fmt.Errorf("Error VerifyBytes: %v", err)
	}

	return nil
}

// Copy ...
func (cand *Candidator) Copy() *Candidator {
	cpy := &Candidator{
		Address: cand.Address,
		IP:      cand.IP,
		Votes:   cand.Votes,
		Status:  cand.Status,
	}

	cpy.Pubkey = make([]byte, len(cand.Pubkey))
	copy(cpy.Pubkey, cand.Pubkey)
	return cpy
}

// CheckVoteStauts ...
func (topNs *TopNCandidators) CheckVoteStauts(delegateNum int64) {
	if topNs.Status == TopNCandidatorsVoteMajorOK || topNs.Status == TopNCandidatorsVoteMajorFail {
		return
	}

	voteMap := make(map[string]int64)

	for i := 0; i < len(topNs.CandsVotes); i++ {
		key := hex.EncodeToString(topNs.CandsVotes[i].Hash)
		if _, ok := voteMap[key]; ok {
			voteMap[key]++
			if voteMap[key] >= (delegateNum * 2 / 3) {
				topNs.Status = TopNCandidatorsVoteMajorOK
				for j := 0; j < len(topNs.CandsVotes[i].Cands); j++ {
					topNs.FinalCands = append(topNs.FinalCands, topNs.CandsVotes[i].Cands[j].Copy())
				}
				return
			}
		} else {
			voteMap[key] = 1
		}
	}

	var maxVotes int64
	var sumVotes int64
	for _, v := range voteMap {
		if v > maxVotes {
			maxVotes = v
		}
		sumVotes += v
	}

	if maxVotes+(delegateNum-sumVotes) < (delegateNum * 2 / 3) {
		topNs.Status = TopNCandidatorsVoteMajorFail
	}
}
