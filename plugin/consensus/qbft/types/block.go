// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/types"
	tmtypes "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
)

// QbftBlockID struct
type BlockID struct {
	*tmtypes.QbftBlockID
}

// IsZero returns true if this is the QbftBlockID for a nil-block
func (blockID BlockID) IsZero() bool {
	if blockID.QbftBlockID == nil {
		return true
	}
	return len(blockID.Hash) == 0
}

// Equals returns true if the QbftBlockID matches the given QbftBlockID
func (blockID BlockID) Equals(other BlockID) bool {
	return blockID.Key() == other.Key()
}

// Key returns a machine-readable string representation of the QbftBlockID
func (blockID BlockID) Key() string {
	if blockID.QbftBlockID == nil {
		return "nil"
	}
	return string(blockID.Hash)
}

// String returns a human readable string representation of the QbftBlockID
func (blockID BlockID) String() string {
	if blockID.QbftBlockID == nil {
		return "nil"
	}
	return Fmt("%X", blockID.Hash)
}

//QbftBlock struct
type QbftBlock struct {
	*tmtypes.QbftBlock
}

// MakeBlock returns a new block with an empty header, except what can be computed from itself.
// It populates the same set of fields validated by ValidateBasic
func MakeBlock(height int64, round int64, pblock *types.Block, commit *tmtypes.QbftCommit) *QbftBlock {
	block := &QbftBlock{
		&tmtypes.QbftBlock{
			Header: &tmtypes.QbftBlockHeader{
				Height: height,
				Round:  round,
				Time:   pblock.BlockTime,
				NumTxs: int64(len(pblock.Txs)),
			},
			Data:       pblock,
			LastCommit: commit,
		},
	}
	block.FillHeader()
	return block
}

// ValidateBasic performs basic validation that doesn't involve state data.
// It checks the internal consistency of the block.
// Further validation is done using state#ValidateBlock.
func (b *QbftBlock) ValidateBasic() error {
	if b == nil {
		return errors.New("nil block")
	}

	if b.Header.Height < 0 {
		return errors.New("Negative Header.Height")
	} else if b.Header.Height == 0 {
		return errors.New("Zero Header.Height")
	}

	if b.Header.TotalTxs < 0 {
		return errors.New("Negative Header.TotalTxs")
	}

	if b.Header.LastSequence == 0 {
		lastCommit := Commit{
			QbftCommit: b.LastCommit,
		}
		if b.Header.Height > 1 {
			if b.LastCommit == nil {
				return errors.New("nil LastCommit")
			}
			if err := lastCommit.ValidateBasic(); err != nil {
				return err
			}
		}
		if !bytes.Equal(b.Header.LastCommitHash, lastCommit.Hash()) {
			return fmt.Errorf("Wrong Header.LastCommitHash.  Expected %v, got %v", b.Header.LastCommitHash, lastCommit.Hash())
		}
	}

	return nil
}

// FillHeader fills in any remaining header fields that are a function of the block data
func (b *QbftBlock) FillHeader() {
	if b.Header.LastCommitHash == nil {
		lastCommit := &Commit{
			QbftCommit: b.LastCommit,
		}
		b.Header.LastCommitHash = lastCommit.Hash()
	}
}

// Hash computes and returns the block hash.
// If the block is incomplete, block hash is nil for safety.
func (b *QbftBlock) Hash() []byte {
	if b == nil || b.Header == nil || b.LastCommit == nil {
		return nil
	}
	b.FillHeader()
	header := &Header{QbftBlockHeader: b.Header}
	return header.Hash()
}

// HashesTo is a convenience function that checks if a block hashes to the given argument.
// A nil block never hashes to anything, and nothing hashes to a nil hash.
func (b *QbftBlock) HashesTo(hash []byte) bool {
	if len(hash) == 0 {
		return false
	}
	if b == nil {
		return false
	}
	return bytes.Equal(b.Hash(), hash)
}

// String returns a string representation of the block
func (b *QbftBlock) String() string {
	return b.StringIndented("")
}

// StringIndented returns a string representation of the block
func (b *QbftBlock) StringIndented(indent string) string {
	if b == nil {
		return "nil-Block"
	}
	header := &Header{QbftBlockHeader: b.Header}
	lastCommit := &Commit{QbftCommit: b.LastCommit}
	return Fmt(`Block{
%s  %v
%s  %v
%s}#%v`,
		indent, header.StringIndented(indent+"  "),
		indent, lastCommit.StringIndented(indent+"  "),
		indent, b.Hash())
}

// StringShort returns a shortened string representation of the block
func (b *QbftBlock) StringShort() string {
	if b == nil {
		return "nil-Block"
	}
	return Fmt("Block#%v", b.Hash())
}

// Header defines the structure of a Qbft block header
// TODO: limit header size
// NOTE: changes to the Header should be duplicated in the abci Header
type Header struct {
	*tmtypes.QbftBlockHeader
}

// Hash returns the hash of the header.
// Returns nil if ValidatorHash is missing.
func (h *Header) Hash() []byte {
	if len(h.ValidatorsHash) == 0 {
		return nil
	}
	bytes, err := json.Marshal(h)
	if err != nil {
		ttlog.Error("block header Hash() marshal failed", "error", err)
		return nil
	}
	return crypto.Ripemd160(bytes)
}

// StringIndented returns a string representation of the header
func (h *Header) StringIndented(indent string) string {
	if h == nil {
		return "nil-Header"
	}
	return Fmt(`Header{
%s  ChainID:        %v
%s  Height:         %v
%s  Time:           %v
%s  NumTxs:         %v
%s  TotalTxs:       %v
%s  LastBlockID:    %v
%s  LastCommit:     %v
%s  Validators:     %v
%s  App:            %v
%s  Consensus:      %v
%s  Results:        %v
%s  ProposerAddr:   %v
%s  Sequence:       %v
%s  LastSequence:   %v
%s}#%v`,
		indent, h.ChainID,
		indent, h.Height,
		indent, time.Unix(0, h.Time),
		indent, h.NumTxs,
		indent, h.TotalTxs,
		indent, h.LastBlockID,
		indent, h.LastCommitHash,
		indent, h.ValidatorsHash,
		indent, h.AppHash,
		indent, h.ConsensusHash,
		indent, h.LastResultsHash,
		indent, h.ProposerAddr,
		indent, h.Sequence,
		indent, h.LastSequence,
		indent, h.Hash())
}

// Commit struct
type Commit struct {
	*tmtypes.QbftCommit
	hash      []byte
	bitArray  *BitArray
	firstVote *tmtypes.QbftVote
}

// FirstVote returns the first non-nil prevote/precommit in the commit
func (commit *Commit) FirstVote() *tmtypes.QbftVote {
	if commit.firstVote != nil {
		return commit.firstVote
	}
	if commit.VoteType == uint32(VoteTypePrecommit) {
		for _, precommit := range commit.Precommits {
			if precommit != nil && len(precommit.Signature) > 0 {
				commit.firstVote = precommit
				return precommit
			}
		}
	} else {
		for _, prevote := range commit.Prevotes {
			if prevote != nil && len(prevote.Signature) > 0 {
				commit.firstVote = prevote
				return prevote
			}
		}
	}
	return nil
}

// Height returns the height of the commit
func (commit *Commit) Height() int64 {
	if commit.AggVote != nil {
		return commit.AggVote.Height
	}
	if commit.VoteType == uint32(VoteTypePrecommit) && len(commit.Precommits) == 0 {
		return -1
	} else if commit.VoteType == uint32(VoteTypePrevote) && len(commit.Prevotes) == 0 {
		return -1
	}
	return commit.FirstVote().Height
}

// Round returns the round of the commit
func (commit *Commit) Round() int {
	if commit.AggVote != nil {
		return int(commit.AggVote.Round)
	}
	if commit.VoteType == uint32(VoteTypePrecommit) && len(commit.Precommits) == 0 {
		return -1
	} else if commit.VoteType == uint32(VoteTypePrevote) && len(commit.Prevotes) == 0 {
		return -1
	}
	return int(commit.FirstVote().Round)
}

// Type returns the vote type of the commit, which is always VoteTypePrecommit
func (commit *Commit) Type() byte {
	return byte(commit.VoteType)
}

// Size returns the number of votes in the commit
func (commit *Commit) Size() int {
	if commit == nil {
		return 0
	}
	if commit.VoteType == uint32(VoteTypePrecommit) {
		return len(commit.Precommits)
	}
	return len(commit.Prevotes)
}

// BitArray returns a BitArray of which validators voted in this commit
func (commit *Commit) BitArray() *BitArray {
	if commit.AggVote != nil {
		bitArray := &BitArray{QbftBitArray: commit.AggVote.ValidatorArray}
		return bitArray.copy()
	}
	if commit.bitArray == nil {
		if commit.VoteType == uint32(VoteTypePrecommit) {
			commit.bitArray = NewBitArray(len(commit.Precommits))
			for i, precommit := range commit.Precommits {
				// TODO: need to check the QbftBlockID otherwise we could be counting conflicts,
				// not just the one with +2/3 !
				commit.bitArray.SetIndex(i, precommit.ValidatorAddress != nil)
			}
		} else {
			commit.bitArray = NewBitArray(len(commit.Prevotes))
			for i, prevote := range commit.Prevotes {
				commit.bitArray.SetIndex(i, prevote.ValidatorAddress != nil)
			}
		}
	}
	return commit.bitArray
}

// GetByIndex returns the vote corresponding to a given validator index
func (commit *Commit) GetByIndex(index int) *Vote {
	if commit.VoteType == uint32(VoteTypePrecommit) {
		return &Vote{QbftVote: commit.Precommits[index]}
	} else {
		return &Vote{QbftVote: commit.Prevotes[index]}
	}
}

// IsCommit returns true if there is at least one vote
func (commit *Commit) IsCommit() bool {
	return len(commit.Precommits) != 0 || len(commit.Precommits) != 0 || commit.AggVote != nil
}

// GetAggVote ...
func (commit *Commit) GetAggVote() *AggVote {
	if commit == nil || commit.AggVote == nil {
		return nil
	}
	aggVote := &AggVote{commit.AggVote}
	return aggVote.Copy()
}

// ValidateBasic performs basic validation that doesn't involve state data.
func (commit *Commit) ValidateBasic() error {
	blockID := &BlockID{QbftBlockID: commit.BlockID}
	if blockID.IsZero() {
		return errors.New("Commit cannot be for nil block")
	}
	if len(commit.Prevotes) == 0 && len(commit.Precommits) == 0 {
		return errors.New("No prevotes and precommits in commit")
	}
	height, round := commit.Height(), commit.Round()

	// validate the prevotes
	for _, item := range commit.Prevotes {
		// may be nil if validator skipped.
		if item == nil || len(item.Signature) == 0 {
			continue
		}
		prevote := &Vote{QbftVote: item}
		// Ensure that all votes are prevotes
		if prevote.Type != uint32(VoteTypePrevote) {
			return fmt.Errorf("Invalid commit vote. Expected prevote, got %v",
				prevote.Type)
		}
		// Ensure that all heights are the same
		if prevote.Height != height {
			return fmt.Errorf("Invalid commit prevote height. Expected %v, got %v",
				height, prevote.Height)
		}
		// Ensure that all rounds are the same
		if int(prevote.Round) != round {
			return fmt.Errorf("Invalid commit prevote round. Expected %v, got %v",
				round, prevote.Round)
		}
	}
	// validate the precommits
	for _, item := range commit.Precommits {
		// may be nil if validator skipped.
		if item == nil || len(item.Signature) == 0 {
			continue
		}
		precommit := &Vote{QbftVote: item}
		// Ensure that all votes are precommits
		if precommit.Type != uint32(VoteTypePrecommit) {
			return fmt.Errorf("Invalid commit vote. Expected precommit, got %v",
				precommit.Type)
		}
		// Ensure that all heights are the same
		if precommit.Height != height {
			return fmt.Errorf("Invalid commit precommit height. Expected %v, got %v",
				height, precommit.Height)
		}
		// Ensure that all rounds are the same
		if int(precommit.Round) != round {
			return fmt.Errorf("Invalid commit precommit round. Expected %v, got %v",
				round, precommit.Round)
		}
	}
	// validate the aggVote
	if commit.AggVote != nil {
		if commit.AggVote.Type != commit.VoteType {
			return fmt.Errorf("Invalid aggVote type. Expected %v, got %v", commit.VoteType, commit.AggVote.Type)
		}
		if commit.AggVote.Height != height {
			return fmt.Errorf("Invalid aggVote height. Expected %v, got %v", height, commit.AggVote.Height)
		}
		if int(commit.AggVote.Round) != round {
			return fmt.Errorf("Invalid aggVote round. Expected %v, got %v", round, commit.AggVote.Round)
		}
	}
	return nil
}

// Hash returns the hash of the commit
func (commit *Commit) Hash() []byte {
	if commit.hash == nil {
		if commit.AggVote != nil {
			aggVote := &AggVote{QbftAggVote: commit.AggVote}
			commit.hash = aggVote.Hash()
		} else if commit.VoteType == uint32(VoteTypePrecommit) {
			pc := make([][]byte, len(commit.Precommits))
			for i, item := range commit.Precommits {
				precommit := Vote{QbftVote: item}
				pc[i] = precommit.Hash()
			}
			commit.hash = merkle.GetMerkleRoot(pc)
		} else {
			pv := make([][]byte, len(commit.Prevotes))
			for i, item := range commit.Prevotes {
				prevote := Vote{QbftVote: item}
				pv[i] = prevote.Hash()
			}
			commit.hash = merkle.GetMerkleRoot(pv)
		}
	}
	return commit.hash
}

// StringIndented returns a string representation of the commit
func (commit *Commit) StringIndented(indent string) string {
	if commit == nil {
		return "nil-Commit"
	}
	prevoteStrings := make([]string, len(commit.Prevotes))
	for i, prevote := range commit.Prevotes {
		prevoteStrings[i] = prevote.String()
	}
	precommitStrings := make([]string, len(commit.Precommits))
	for i, precommit := range commit.Precommits {
		precommitStrings[i] = precommit.String()
	}
	return Fmt(`Commit{
%s  QbftBlockID:    %v
%s  Prevotes:   %v
%s  Precommits: %v
%s  QbftAggVote:    %v
%s  VoteType:   %v
%s}#%v`,
		indent, commit.BlockID,
		indent, strings.Join(prevoteStrings, "\n"+indent+"  "),
		indent, strings.Join(precommitStrings, "\n"+indent+"  "),
		indent, commit.AggVote.String(),
		indent, commit.VoteType,
		indent, commit.hash)
}

// SignedHeader is a header along with the commits that prove it
type SignedHeader struct {
	Header *Header `json:"header"`
	Commit *Commit `json:"commit"`
}
