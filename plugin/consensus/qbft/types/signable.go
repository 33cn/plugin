// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/33cn/chain33/common/crypto"
	tmtypes "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
)

// error defines
var (
	ErrVoteUnexpectedStep            = errors.New("Unexpected step")
	ErrVoteInvalidValidatorIndex     = errors.New("Invalid validator index")
	ErrVoteInvalidValidatorAddress   = errors.New("Invalid validator address")
	ErrVoteInvalidSignature          = errors.New("Invalid signature")
	ErrVoteInvalidBlockHash          = errors.New("Invalid block hash")
	ErrVoteNonDeterministicSignature = errors.New("Non-deterministic signature")
	ErrVoteConflict                  = errors.New("Conflicting vote")
	ErrVoteNil                       = errors.New("Nil vote")
	ErrAggVoteNil                    = errors.New("Nil aggregate vote")
)

// Signable is an interface for all signable things.
// It typically removes signatures before serializing.
type Signable interface {
	WriteSignBytes(chainID string, w io.Writer, n *int, err *error)
}

// SignBytes is a convenience method for getting the bytes to sign of a Signable.
func SignBytes(chainID string, o Signable) []byte {
	buf, n, err := new(bytes.Buffer), new(int), new(error)
	o.WriteSignBytes(chainID, buf, n, err)
	if *err != nil {
		PanicCrisis(err)
	}
	return buf.Bytes()
}

// Proposal defines a block proposal for the consensus.
// It refers to the block only by its PartSetHeader.
// It must be signed by the correct proposer for the given Height/Round
// to be considered valid. It may depend on votes from a previous round,
// a so-called Proof-of-Lock (POL) round, as noted in the POLRound and POLBlockID.
type Proposal struct {
	tmtypes.QbftProposal
}

// NewProposal returns a new Proposal.
// If there is no POLRound, polRound should be -1.
func NewProposal(height int64, round int, blockhash []byte, polRound int, polBlockID tmtypes.QbftBlockID, seq int64) *Proposal {
	return &Proposal{tmtypes.QbftProposal{
		Height:     height,
		Round:      int32(round),
		Timestamp:  time.Now().UnixNano(),
		POLRound:   int32(polRound),
		POLBlockID: &polBlockID,
		Blockhash:  blockhash,
		Sequence:   seq,
	},
	}
}

// String returns a string representation of the Proposal.
func (p *Proposal) String() string {
	return fmt.Sprintf("Proposal{%v/%v (%v, %X) %X %X %v @ %s}",
		p.Height, p.Round, p.POLRound, p.POLBlockID.Hash,
		p.Blockhash, p.Signature, p.Sequence, CanonicalTime(time.Unix(0, p.Timestamp)))
}

// WriteSignBytes writes the Proposal bytes for signing
func (p *Proposal) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	if *err != nil {
		return
	}
	canonical := CanonicalJSONOnceProposal{
		ChainID:  chainID,
		Proposal: CanonicalProposal(p),
	}
	byteOnceProposal, e := json.Marshal(&canonical)
	if e != nil {
		*err = e
		return
	}
	number, writeErr := w.Write(byteOnceProposal)
	*n = number
	*err = writeErr
}

// Heartbeat ...
type Heartbeat struct {
	*tmtypes.QbftHeartbeat
}

// WriteSignBytes writes the Heartbeat for signing.
// It panics if the Heartbeat is nil.
func (heartbeat *Heartbeat) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	if *err != nil {
		return
	}
	canonical := CanonicalJSONOnceHeartbeat{
		chainID,
		CanonicalHeartbeat(heartbeat),
	}
	byteHeartbeat, e := json.Marshal(&canonical)
	if e != nil {
		*err = e
		return
	}
	number, writeErr := w.Write(byteHeartbeat)
	*n = number
	*err = writeErr
}

// Types of votes
// TODO Make a new type "VoteType"
const (
	VoteTypeNone      = byte(0x0)
	VoteTypePrevote   = byte(0x01)
	VoteTypePrecommit = byte(0x02)
)

// IsVoteTypeValid ...
func IsVoteTypeValid(voteType byte) bool {
	switch voteType {
	case VoteTypePrevote:
		return true
	case VoteTypePrecommit:
		return true
	default:
		return false
	}
}

// Vote Represents a prevote, precommit, or commit vote from validators for consensus.
type Vote struct {
	*tmtypes.QbftVote
}

// WriteSignBytes ...
func (vote *Vote) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	if *err != nil {
		return
	}
	canonical := CanonicalJSONOnceVote{
		chainID,
		CanonicalVote(vote),
	}
	byteVote, e := json.Marshal(&canonical)
	if e != nil {
		*err = e
		ttlog.Error("vote WriteSignBytes marshal failed", "err", e)
		return
	}
	number, writeErr := w.Write(byteVote)
	*n = number
	*err = writeErr
}

// Copy ...
func (vote *Vote) Copy() *Vote {
	voteCopy := *vote
	return &voteCopy
}

func (vote *Vote) String() string {
	if vote == nil {
		return "nil-Vote"
	}
	var typeString string
	switch byte(vote.Type) {
	case VoteTypePrevote:
		typeString = "Prevote"
	case VoteTypePrecommit:
		typeString = "Precommit"
	default:
		PanicSanity("Unknown vote type")
	}

	return fmt.Sprintf("Vote{%v:%X %v/%02d/%v(%v) %X %X @ %s}",
		vote.ValidatorIndex, Fingerprint(vote.ValidatorAddress),
		vote.Height, vote.Round, vote.Type, typeString,
		Fingerprint(vote.BlockID.Hash), vote.Signature,
		CanonicalTime(time.Unix(0, vote.Timestamp)))
}

// Verify ...
func (vote *Vote) Verify(chainID string, pubKey crypto.PubKey) error {
	addr := GenAddressByPubKey(pubKey)
	if !bytes.Equal(addr, vote.ValidatorAddress) {
		return ErrVoteInvalidValidatorAddress
	}

	sig, err := ConsensusCrypto.SignatureFromBytes(vote.Signature)
	if err != nil {
		ttlog.Error("vote Verify fail", "err", err)
		return err
	}

	if !pubKey.VerifyBytes(SignBytes(chainID, vote), sig) {
		return ErrVoteInvalidSignature
	}
	return nil
}

// Hash ...
func (vote *Vote) Hash() []byte {
	if vote == nil {
		return nil
	}
	bytes, err := json.Marshal(vote)
	if err != nil {
		ttlog.Error("vote hash marshal failed", "err", err)
		return nil
	}

	return crypto.Ripemd160(bytes)
}

// AggVote Represents a prevote, precommit, or commit vote from validators for consensus.
type AggVote struct {
	*tmtypes.QbftAggVote
}

// WriteSignBytes ...
func (aggVote *AggVote) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	if *err != nil {
		return
	}
	canonical := CanonicalJSONOnceAggVote{
		chainID,
		CanonicalAggVote(aggVote),
	}
	byteVote, e := json.Marshal(&canonical)
	if e != nil {
		*err = e
		ttlog.Error("aggVote WriteSignBytes marshal failed", "err", e)
		return
	}
	number, writeErr := w.Write(byteVote)
	*n = number
	*err = writeErr
}

// Verify ...
func (aggVote *AggVote) Verify(chainID string, valSet *ValidatorSet) error {
	aggSig, err := ConsensusCrypto.SignatureFromBytes(aggVote.Signature)
	if err != nil {
		return errors.New("invalid aggregate signature")
	}
	pubs := make([]crypto.PubKey, 0)
	arr := &BitArray{QbftBitArray: aggVote.ValidatorArray}
	for i, val := range valSet.Validators {
		if arr.GetIndex(i) {
			pub, _ := ConsensusCrypto.PubKeyFromBytes(val.PubKey)
			pubs = append(pubs, pub)
		}
	}
	origVote := &Vote{&tmtypes.QbftVote{
		BlockID:   aggVote.BlockID,
		Height:    aggVote.Height,
		Round:     aggVote.Round,
		Timestamp: aggVote.Timestamp,
		Type:      aggVote.Type,
		UseAggSig: true,
	}}
	aggr, err := crypto.ToAggregate(ConsensusCrypto)
	if err != nil {
		return err
	}
	err = aggr.VerifyAggregatedOne(pubs, SignBytes(chainID, origVote), aggSig)
	if err != nil {
		ttlog.Error("aggVote Verify fail", "err", err, "aggVote", aggVote, "aggSig", aggSig)
		return err
	}
	return nil
}

// Copy ...
func (aggVote *AggVote) Copy() *AggVote {
	copy := *aggVote
	return &copy
}

func (aggVote *AggVote) String() string {
	if aggVote == nil {
		return "nil-AggVote"
	}
	var typeString string
	switch byte(aggVote.Type) {
	case VoteTypePrevote:
		typeString = "Prevote"
	case VoteTypePrecommit:
		typeString = "Precommit"
	default:
		PanicSanity("Unknown vote type")
	}
	bitArray := &BitArray{QbftBitArray: aggVote.ValidatorArray}

	return fmt.Sprintf("AggVote{%X %v/%02d/%v(%v) %X %X @ %s %v}",
		Fingerprint(aggVote.ValidatorAddress),
		aggVote.Height, aggVote.Round, aggVote.Type, typeString,
		Fingerprint(aggVote.BlockID.Hash), aggVote.Signature,
		CanonicalTime(time.Unix(0, aggVote.Timestamp)),
		bitArray)
}

// Hash ...
func (aggVote *AggVote) Hash() []byte {
	if aggVote == nil {
		return nil
	}
	bytes, err := json.Marshal(aggVote)
	if err != nil {
		ttlog.Error("aggVote hash marshal failed", "err", err)
		return nil
	}

	return crypto.Ripemd160(bytes)
}
