// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"time"

	"reflect"

	tmtypes "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
)

// RoundStepType enumerates the state of the consensus state machine
type RoundStepType uint8 // These must be numeric, ordered.

var (
	// MsgMap define
	MsgMap map[byte]reflect.Type
)

// step and message id define
const (
	RoundStepNewHeight        = RoundStepType(0x01) // Wait til CommitTime + timeoutCommit
	RoundStepNewRound         = RoundStepType(0x02) // Setup new round and go to RoundStepPropose
	RoundStepPropose          = RoundStepType(0x03) // Did propose, gossip proposal
	RoundStepPrevote          = RoundStepType(0x04) // Did prevote, gossip prevotes
	RoundStepAggPrevoteWait   = RoundStepType(0x05) // Did send prevote for aggregate, start timeout
	RoundStepPrevoteWait      = RoundStepType(0x06) // Did receive any +2/3 prevotes, start timeout
	RoundStepPrecommit        = RoundStepType(0x07) // Did precommit, gossip precommits
	RoundStepAggPrecommitWait = RoundStepType(0x08) // Did send precommit for aggregate, start timeout
	RoundStepPrecommitWait    = RoundStepType(0x09) // Did receive any +2/3 precommits, start timeout
	RoundStepCommit           = RoundStepType(0x10) // Entered commit state machine
	// NOTE: RoundStepNewHeight acts as RoundStepCommitWait.

	NewRoundStepID      = byte(0x01)
	ProposalID          = byte(0x02)
	ProposalPOLID       = byte(0x03)
	VoteID              = byte(0x04)
	HasVoteID           = byte(0x05)
	VoteSetMaj23ID      = byte(0x06)
	VoteSetBitsID       = byte(0x07)
	ProposalHeartbeatID = byte(0x08)
	ProposalBlockID     = byte(0x09)
	ValidBlockID        = byte(0x0a)
	AggVoteID           = byte(0x0b)
)

// InitMessageMap ...
func InitMessageMap() {
	MsgMap = map[byte]reflect.Type{
		NewRoundStepID:      reflect.TypeOf(tmtypes.QbftNewRoundStepMsg{}),
		ProposalID:          reflect.TypeOf(tmtypes.QbftProposal{}),
		ProposalPOLID:       reflect.TypeOf(tmtypes.QbftProposalPOLMsg{}),
		VoteID:              reflect.TypeOf(tmtypes.QbftVote{}),
		HasVoteID:           reflect.TypeOf(tmtypes.QbftHasVoteMsg{}),
		VoteSetMaj23ID:      reflect.TypeOf(tmtypes.QbftVoteSetMaj23Msg{}),
		VoteSetBitsID:       reflect.TypeOf(tmtypes.QbftVoteSetBitsMsg{}),
		ProposalHeartbeatID: reflect.TypeOf(tmtypes.QbftHeartbeat{}),
		ProposalBlockID:     reflect.TypeOf(tmtypes.QbftBlock{}),
		ValidBlockID:        reflect.TypeOf(tmtypes.QbftValidBlockMsg{}),
		AggVoteID:           reflect.TypeOf(tmtypes.QbftAggVote{}),
	}
}

// String returns a string
func (rs RoundStepType) String() string {
	switch rs {
	case RoundStepNewHeight:
		return "RoundStepNewHeight"
	case RoundStepNewRound:
		return "RoundStepNewRound"
	case RoundStepPropose:
		return "RoundStepPropose"
	case RoundStepPrevote:
		return "RoundStepPrevote"
	case RoundStepPrevoteWait:
		return "RoundStepPrevoteWait"
	case RoundStepPrecommit:
		return "RoundStepPrecommit"
	case RoundStepPrecommitWait:
		return "RoundStepPrecommitWait"
	case RoundStepCommit:
		return "RoundStepCommit"
	case RoundStepAggPrevoteWait:
		return "RoundStepAggPrevoteWait"
	case RoundStepAggPrecommitWait:
		return "RoundStepAggPrecommitWait"
	default:
		return "RoundStepUnknown" // Cannot panic.
	}
}

//-----------------------------------------------------------------------------

// RoundState defines the internal consensus state.
// It is Immutable when returned from ConsensusState.GetRoundState()
// TODO: Actually, only the top pointer is copied,
// so access to field pointers is still racey
// NOTE: Not thread safe. Should only be manipulated by functions downstream
// of the cs.receiveRoutine
type RoundState struct {
	Height         int64 // Height we are working on
	Round          int
	Step           RoundStepType
	StartTime      time.Time
	CommitTime     time.Time // Subjective time when +2/3 precommits for Block at Round were found
	Validators     *ValidatorSet
	Proposal       *tmtypes.QbftProposal
	ProposalBlock  *QbftBlock
	LockedRound    int
	LockedBlock    *QbftBlock
	ValidRound     int        // Last known round with POL for non-nil valid block.
	ValidBlock     *QbftBlock // Last known block of POL mentioned above.
	Votes          *HeightVoteSet
	CommitRound    int
	LastCommit     *VoteSet // Last precommits at Height-1
	LastValidators *ValidatorSet
}

// RoundStateMessage ...
func (rs *RoundState) RoundStateMessage() *tmtypes.QbftNewRoundStepMsg {
	return &tmtypes.QbftNewRoundStepMsg{
		Height:                rs.Height,
		Round:                 int32(rs.Round),
		Step:                  int32(rs.Step),
		SecondsSinceStartTime: int32(time.Since(rs.StartTime).Seconds()),
		LastCommitRound:       int32(rs.LastCommit.Round()),
	}
}

// String returns a string
func (rs *RoundState) String() string {
	return rs.StringIndented("")
}

// StringIndented returns a string
func (rs *RoundState) StringIndented(indent string) string {
	return Fmt(`RoundState{
%s  H:%v R:%v S:%v
%s  StartTime:     %v
%s  CommitTime:    %v
%s  Validators:    %v
%s  QbftProposal:      %v
%s  ProposalBlock: %v
%s  LockedRound:   %v
%s  LockedBlock:   %v
%s  ValidRound:    %v
%s  ValidBlock:    %v
%s  Votes:         %v
%s  LastCommit:    %v
%s  LastValidators:%v
%s}`,
		indent, rs.Height, rs.Round, rs.Step,
		indent, rs.StartTime,
		indent, rs.CommitTime,
		indent, rs.Validators.StringIndented(indent+"    "),
		indent, rs.Proposal,
		indent, rs.ProposalBlock.StringShort(),
		indent, rs.LockedRound,
		indent, rs.LockedBlock.StringShort(),
		indent, rs.ValidRound,
		indent, rs.ValidBlock.StringShort(),
		indent, rs.Votes.StringIndented(indent+"    "),
		indent, rs.LastCommit.StringShort(),
		indent, rs.LastValidators.StringIndented(indent+"    "),
		indent)
}

// StringShort returns a string
func (rs *RoundState) StringShort() string {
	return Fmt(`RoundState{H:%v R:%v S:%v ST:%v}`,
		rs.Height, rs.Round, rs.Step, rs.StartTime)
}

// PeerRoundState ...
type PeerRoundState struct {
	Height             int64         // Height peer is at
	Round              int           // Round peer is at, -1 if unknown.
	Step               RoundStepType // Step peer is at
	StartTime          time.Time     // Estimated start of round 0 at this height
	Proposal           bool          // True if peer has proposal for this round
	ProposalBlock      bool          // True if peer has proposal block for this round
	ProposalBlockHash  []byte
	ProposalPOLRound   int       // QbftProposal's POL round. -1 if none.
	ProposalPOL        *BitArray // nil until ProposalPOLMessage received.
	Prevotes           *BitArray // All votes peer has for this round
	Precommits         *BitArray // All precommits peer has for this round
	LastCommitRound    int       // Round of commit for last height. -1 if none.
	LastCommit         *BitArray // All commit precommits of commit for last height.
	CatchupCommitRound int       // Round that we have commit for. Not necessarily unique. -1 if none.
	CatchupCommit      *BitArray // All commit precommits peer has for this height & CatchupCommitRound
	AggPrevote         bool      // True if peer has aggregate prevote for this round
	AggPrecommit       bool      // True if peer has aggregate precommit for this round
}

// String returns a string representation of the PeerRoundState
func (prs PeerRoundState) String() string {
	return prs.StringIndented("")
}

// StringIndented returns a string representation of the PeerRoundState
func (prs PeerRoundState) StringIndented(indent string) string {
	return Fmt(`PeerRoundState{
%s  %v/%v/%v @%v
%s  QbftProposal %v
%s  ProposalBlock %v
%s  ProposalBlockHash %X
%s  POL      %v (round %v)
%s  Prevotes   %v
%s  Precommits %v
%s  LastCommit %v (round %v)
%s  CatchupCommit %v (round %v)
%s  AggPrevote %v
%s  AggPrecommit %v
%s}`,
		indent, prs.Height, prs.Round, prs.Step, prs.StartTime,
		indent, prs.Proposal,
		indent, prs.ProposalBlock,
		indent, prs.ProposalBlockHash,
		indent, prs.ProposalPOL, prs.ProposalPOLRound,
		indent, prs.Prevotes,
		indent, prs.Precommits,
		indent, prs.LastCommit, prs.LastCommitRound,
		indent, prs.CatchupCommit, prs.CatchupCommitRound,
		indent, prs.AggPrevote,
		indent, prs.AggPrecommit,
		indent)
}

//---------------------Canonical json-----------------------------------

// CanonicalJSONBlockID ...
type CanonicalJSONBlockID struct {
	Hash        []byte                     `json:"hash,omitempty"`
	PartsHeader CanonicalJSONPartSetHeader `json:"parts,omitempty"`
}

// CanonicalJSONPartSetHeader ...
type CanonicalJSONPartSetHeader struct {
	Hash  []byte `json:"hash"`
	Total int    `json:"total"`
}

// CanonicalJSONProposal ...
type CanonicalJSONProposal struct {
	BlockBytes []byte               `json:"block_parts_header"`
	Height     int64                `json:"height"`
	POLBlockID CanonicalJSONBlockID `json:"pol_block_id"`
	POLRound   int                  `json:"pol_round"`
	Round      int                  `json:"round"`
	Timestamp  string               `json:"timestamp"`
}

// CanonicalJSONVote ...
type CanonicalJSONVote struct {
	BlockID   CanonicalJSONBlockID `json:"block_id"`
	Height    int64                `json:"height"`
	Round     int                  `json:"round"`
	Timestamp string               `json:"timestamp"`
	Type      byte                 `json:"type"`
}

// CanonicalJSONHeartbeat ...
type CanonicalJSONHeartbeat struct {
	Height           int64  `json:"height"`
	Round            int    `json:"round"`
	Sequence         int    `json:"sequence"`
	ValidatorAddress []byte `json:"validator_address"`
	ValidatorIndex   int    `json:"validator_index"`
}

// Messages including a "chain id" can only be applied to one chain, hence "Once"

// CanonicalJSONOnceProposal ...
type CanonicalJSONOnceProposal struct {
	ChainID  string                `json:"chain_id"`
	Proposal CanonicalJSONProposal `json:"proposal"`
}

// CanonicalJSONOnceVote ...
type CanonicalJSONOnceVote struct {
	ChainID string            `json:"chain_id"`
	Vote    CanonicalJSONVote `json:"vote"`
}

// CanonicalJSONOnceAggVote ...
type CanonicalJSONOnceAggVote struct {
	ChainID string            `json:"chain_id"`
	AggVote CanonicalJSONVote `json:"agg_vote"`
}

// CanonicalJSONOnceHeartbeat ...
type CanonicalJSONOnceHeartbeat struct {
	ChainID   string                 `json:"chain_id"`
	Heartbeat CanonicalJSONHeartbeat `json:"heartbeat"`
}

// Canonicalize the structs

// CanonicalBlockID ...
func CanonicalBlockID(blockID BlockID) CanonicalJSONBlockID {
	return CanonicalJSONBlockID{
		Hash: blockID.Hash,
	}
}

// CanonicalProposal ...
func CanonicalProposal(proposal *Proposal) CanonicalJSONProposal {
	return CanonicalJSONProposal{
		//BlockBytes: proposal.BlockBytes,
		Height:    proposal.Height,
		Timestamp: CanonicalTime(time.Unix(0, proposal.Timestamp)),
		POLBlockID: CanonicalJSONBlockID{
			Hash: proposal.POLBlockID.Hash,
		},
		POLRound: int(proposal.POLRound),
		Round:    int(proposal.Round),
	}
}

// CanonicalVote ...
func CanonicalVote(vote *Vote) CanonicalJSONVote {
	timestamp := ""
	if !vote.UseAggSig {
		timestamp = CanonicalTime(time.Unix(0, vote.Timestamp))
	}
	return CanonicalJSONVote{
		BlockID:   CanonicalJSONBlockID{Hash: vote.BlockID.Hash},
		Height:    vote.Height,
		Round:     int(vote.Round),
		Timestamp: timestamp,
		Type:      byte(vote.Type),
	}
}

// CanonicalAggVote ...
func CanonicalAggVote(vote *AggVote) CanonicalJSONVote {
	return CanonicalJSONVote{
		BlockID:   CanonicalJSONBlockID{Hash: vote.BlockID.Hash},
		Height:    vote.Height,
		Round:     int(vote.Round),
		Timestamp: "",
		Type:      byte(vote.Type),
	}
}

// CanonicalHeartbeat ...
func CanonicalHeartbeat(heartbeat *Heartbeat) CanonicalJSONHeartbeat {
	return CanonicalJSONHeartbeat{
		heartbeat.Height,
		int(heartbeat.Round),
		int(heartbeat.Sequence),
		heartbeat.ValidatorAddress,
		int(heartbeat.ValidatorIndex),
	}
}

// CanonicalTime ...
func CanonicalTime(t time.Time) string {
	// note that sending time over go-wire resets it to
	// local time, we need to force UTC here, so the
	// signatures match
	return t.UTC().Format(timeFormat)
}
