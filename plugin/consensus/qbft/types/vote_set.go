// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"bytes"
	"strings"
	"sync"
	"time"

	"github.com/33cn/chain33/common/crypto"
	tmtypes "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
	"github.com/pkg/errors"
)

/*
	VoteSet helps collect signatures from validators at each height+round for a
	predefined vote type.

	We need VoteSet to be able to keep track of conflicting votes when validators
	double-sign.  Yet, we can't keep track of *all* the votes seen, as that could
	be a DoS attack vector.

	There are two storage areas for votes.
	1. voteSet.votes
	2. voteSet.votesByBlock

	`.votes` is the "canonical" list of votes.  It always has at least one vote,
	if a vote from a validator had been seen at all.  Usually it keeps track of
	the first vote seen, but when a 2/3 majority is found, votes for that get
	priority and are copied over from `.votesByBlock`.

	`.votesByBlock` keeps track of a list of votes for a particular block.  There
	are two ways a &blockVotes{} gets created in `.votesByBlock`.
	1. the first vote seen by a validator was for the particular block.
	2. a peer claims to have seen 2/3 majority for the particular block.

	Since the first vote from a validator will always get added in `.votesByBlock`
	, all votes in `.votes` will have a corresponding entry in `.votesByBlock`.

	When a &blockVotes{} in `.votesByBlock` reaches a 2/3 majority quorum, its
	votes are copied into `.votes`.

	All this is memory bounded because conflicting votes only get added if a peer
	told us to track that block, each peer only gets to tell us 1 such block, and,
	there's only a limited number of peers.

	NOTE: Assumes that the sum total of voting power does not exceed MaxUInt64.
*/

// VoteSet ...
type VoteSet struct {
	chainID  string
	height   int64
	round    int
	voteType byte

	mtx           sync.Mutex
	valSet        *ValidatorSet
	votesBitArray *BitArray
	votes         []*Vote                // Primary votes to share
	sum           int64                  // Sum of voting power for seen votes, discounting conflicts
	maj23         *tmtypes.QbftBlockID   // First 2/3 majority seen
	votesByBlock  map[string]*blockVotes // string(blockHash|blockParts) -> blockVotes
	peerMaj23s    map[string]BlockID     // Maj23 for each peer
	aggVote       *AggVote               // aggregate vote
}

// NewVoteSet Constructs a new VoteSet struct used to accumulate votes for given height/round.
func NewVoteSet(chainID string, height int64, round int, voteType byte, valSet *ValidatorSet) *VoteSet {
	if height == 0 {
		PanicSanity("Cannot make VoteSet for height == 0, doesn't make sense.")
	}
	return &VoteSet{
		chainID:       chainID,
		height:        height,
		round:         round,
		voteType:      voteType,
		valSet:        valSet,
		votesBitArray: NewBitArray(valSet.Size()),
		votes:         make([]*Vote, valSet.Size()),
		sum:           0,
		maj23:         nil,
		votesByBlock:  make(map[string]*blockVotes, valSet.Size()),
		peerMaj23s:    make(map[string]BlockID),
		aggVote:       nil,
	}
}

// ChainID ...
func (voteSet *VoteSet) ChainID() string {
	return voteSet.chainID
}

// Height ...
func (voteSet *VoteSet) Height() int64 {
	if voteSet == nil {
		return 0
	}
	return voteSet.height
}

// Round ...
func (voteSet *VoteSet) Round() int {
	if voteSet == nil {
		return -1
	}
	return voteSet.round
}

// Type ...
func (voteSet *VoteSet) Type() byte {
	if voteSet == nil {
		return 0x00
	}
	return voteSet.voteType
}

// Size ...
func (voteSet *VoteSet) Size() int {
	if voteSet == nil {
		return 0
	}
	return voteSet.valSet.Size()
}

// AddVote Returns added=true if vote is valid and new.
// Otherwise returns err=ErrVote[
//    UnexpectedStep | InvalidIndex | InvalidAddress |
//    InvalidSignature | InvalidBlockHash | ConflictingVotes ]
// Duplicate votes return added=false, err=nil.
// Conflicting votes return added=*, err=ErrVoteConflictingVotes.
// NOTE: vote should not be mutated after adding.
// NOTE: VoteSet must not be nil
// NOTE: QbftVote must not be nil
func (voteSet *VoteSet) AddVote(vote *Vote) (added bool, err error) {
	if voteSet == nil {
		return false, errors.New("nil vote set")
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()

	return voteSet.addVote(vote)
}

// NOTE: Validates as much as possible before attempting to verify the signature.
func (voteSet *VoteSet) addVote(vote *Vote) (added bool, err error) {
	if vote == nil {
		return false, ErrVoteNil
	}
	valIndex := int(vote.ValidatorIndex)
	valAddr := vote.ValidatorAddress
	blockKey := BlockID{QbftBlockID: vote.BlockID}.Key()

	// Ensure that validator index was set
	if valIndex < 0 {
		return false, errors.Wrap(ErrVoteInvalidValidatorIndex, "Index < 0")
	} else if len(valAddr) == 0 {
		return false, errors.Wrap(ErrVoteInvalidValidatorAddress, "Empty address")
	}

	// Make sure the step matches.
	if (vote.Height != voteSet.height) ||
		(int(vote.Round) != voteSet.round) ||
		(vote.Type != uint32(voteSet.voteType)) {
		return false, errors.Wrapf(ErrVoteUnexpectedStep, "Got %d/%d/%d, expected %d/%d/%d",
			voteSet.height, voteSet.round, voteSet.voteType,
			vote.Height, vote.Round, vote.Type)
	}

	// Ensure that signer is a validator.
	lookupAddr, val := voteSet.valSet.GetByIndex(valIndex)
	if val == nil {
		return false, errors.Wrapf(ErrVoteInvalidValidatorIndex,
			"Cannot find validator %d in valSet of size %d", valIndex, voteSet.valSet.Size())
	}

	// Ensure that the signer has the right address
	if !bytes.Equal(valAddr, lookupAddr) {
		return false, errors.Wrapf(ErrVoteInvalidValidatorAddress,
			"vote.ValidatorAddress (%X) does not match address (%X) for vote.ValidatorIndex (%d)",
			valAddr, lookupAddr, valIndex)
	}

	// If we already know of this vote, return false.
	if existing, ok := voteSet.getVote(valIndex, blockKey); ok {
		if bytes.Equal(existing.Signature[:], vote.Signature[:]) {
			return false, nil // duplicate
		}
		return false, errors.Wrapf(ErrVoteNonDeterministicSignature, "Existing vote: %v; New vote: %v", existing, vote)
	}

	// Check signature.
	pubkey, err := ConsensusCrypto.PubKeyFromBytes(val.PubKey)
	if err != nil {
		return false, errors.Wrapf(err, "PubKeyFromBytes: %X failed", val.PubKey)
	}
	if err := vote.Verify(voteSet.chainID, pubkey); err != nil {
		return false, errors.Wrapf(err, "Failed to verify vote with ChainID %s and PubKey %X", voteSet.chainID, val.PubKey)
	}

	// Add vote and get conflicting vote if any
	added, conflicting := voteSet.addVerifiedVote(vote, blockKey, val.VotingPower)
	if conflicting != nil {
		return added, errors.Wrapf(ErrVoteConflict, "Conflicting vote: %v; New vote: %v", conflicting, vote)
	}
	if !added {
		PanicSanity("Expected to add non-conflicting vote")
	}
	return added, nil
}

// Returns (vote, true) if vote exists for valIndex and blockKey
func (voteSet *VoteSet) getVote(valIndex int, blockKey string) (vote *Vote, ok bool) {

	if existing := voteSet.votes[valIndex]; existing != nil && string(existing.BlockID.Hash) == blockKey {
		return existing, true
	}
	if existing := voteSet.votesByBlock[blockKey].getByIndex(valIndex); existing != nil {
		return existing, true
	}
	return nil, false
}

// Assumes signature is valid.
// If conflicting vote exists, returns it.
func (voteSet *VoteSet) addVerifiedVote(vote *Vote, blockKey string, votingPower int64) (added bool, conflicting *Vote) {
	valIndex := int(vote.ValidatorIndex)

	// Already exists in voteSet.votes?
	if existing := voteSet.votes[valIndex]; existing != nil {
		if bytes.Equal(existing.BlockID.Hash, vote.BlockID.Hash) {
			PanicSanity("addVerifiedVote does not expect duplicate votes")
		} else {
			conflicting = existing
		}
		// Replace vote if blockKey matches voteSet.maj23.
		if voteSet.maj23 != nil && string(voteSet.maj23.Hash) == blockKey {
			voteSet.votes[valIndex] = vote
			voteSet.votesBitArray.SetIndex(valIndex, true)
		}
		// Otherwise don't add it to voteSet.votes
	} else {
		// Add to voteSet.votes and incr .sum
		voteSet.votes[valIndex] = vote
		voteSet.votesBitArray.SetIndex(valIndex, true)
		voteSet.sum += votingPower
	}

	votesByBlock, ok := voteSet.votesByBlock[blockKey]
	if ok {
		if conflicting != nil && !votesByBlock.peerMaj23 {
			// There's a conflict and no peer claims that this block is special.
			return false, conflicting
		}
		// We'll add the vote in a bit.
	} else {
		// .votesByBlock doesn't exist...
		if conflicting != nil {
			// ... and there's a conflicting vote.
			// We're not even tracking this blockKey, so just forget it.
			return false, conflicting
		}
		// ... and there's no conflicting vote.
		// Start tracking this blockKey
		votesByBlock = newBlockVotes(false, voteSet.valSet.Size())
		voteSet.votesByBlock[blockKey] = votesByBlock
		// We'll add the vote in a bit.
	}

	// Before adding to votesByBlock, see if we'll exceed quorum
	origSum := votesByBlock.sum
	quorum := voteSet.valSet.TotalVotingPower()*2/3 + 1

	// Add vote to votesByBlock
	votesByBlock.addVerifiedVote(vote, votingPower)

	// If we just crossed the quorum threshold and have 2/3 majority...
	if origSum < quorum && quorum <= votesByBlock.sum {
		// Only consider the first quorum reached
		if voteSet.maj23 == nil {
			voteSet.maj23 = &tmtypes.QbftBlockID{Hash: make([]byte, len(vote.BlockID.Hash))}
			copy(voteSet.maj23.Hash, vote.BlockID.Hash)
			// And also copy votes over to voteSet.votes
			for i, vote := range votesByBlock.votes {
				if vote != nil {
					voteSet.votes[i] = vote
				}
			}
		}
	}

	return true, conflicting
}

// AddAggVote Returns added=true if aggVote is valid and new
func (voteSet *VoteSet) AddAggVote(vote *AggVote) (bool, error) {
	if voteSet == nil {
		return false, errors.New("nil vote set")
	}
	if vote == nil {
		return false, ErrAggVoteNil
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()

	valAddr := vote.ValidatorAddress
	valset := voteSet.valSet
	if len(valAddr) == 0 {
		return false, errors.Wrap(ErrVoteInvalidValidatorAddress, "Empty address")
	}

	// Make sure the step matches
	if (vote.Height != voteSet.height) ||
		(int(vote.Round) != voteSet.round) ||
		(vote.Type != uint32(voteSet.voteType)) {
		return false, errors.Wrapf(ErrVoteUnexpectedStep, "Got %d/%d/%d, expected %d/%d/%d",
			voteSet.height, voteSet.round, voteSet.voteType,
			vote.Height, vote.Round, vote.Type)
	}
	// Ensure that signer is proposer
	propAddr := valset.Proposer.Address
	if !bytes.Equal(valAddr, propAddr) {
		return false, errors.Wrapf(ErrVoteInvalidValidatorAddress,
			"aggVote.ValidatorAddress (%X) does not match proposer address (%X)",
			valAddr, propAddr)
	}
	// If we already know of this vote, return false
	if voteSet.aggVote != nil {
		if bytes.Equal(voteSet.aggVote.Signature, vote.Signature) {
			return false, nil // duplicate
		}
		return false, errors.Wrapf(ErrVoteNonDeterministicSignature, "Existing vote: %v; New vote: %v", voteSet.aggVote, vote)
	}

	// Check signature
	err := vote.Verify(voteSet.chainID, voteSet.valSet)
	if err != nil {
		return false, err
	}

	// Check maj32
	sum := int64(0)
	arr := &BitArray{QbftBitArray: vote.ValidatorArray}
	for i, val := range valset.Validators {
		if arr.GetIndex(i) {
			sum += val.VotingPower
		}
	}
	quorum := voteSet.valSet.TotalVotingPower()*2/3 + 1
	if sum < quorum {
		return false, errors.New("less than 2/3 total power")
	}

	voteSet.votesBitArray = arr.copy()
	voteSet.aggVote = vote
	voteSet.maj23 = &tmtypes.QbftBlockID{Hash: make([]byte, len(vote.BlockID.Hash))}
	copy(voteSet.maj23.Hash, vote.BlockID.Hash)
	voteSet.sum = sum
	votesByBlock := newBlockVotes(false, voteSet.valSet.Size())
	votesByBlock.bitArray = arr.copy()
	votesByBlock.sum = sum
	voteSet.votesByBlock[string(voteSet.maj23.Hash)] = votesByBlock

	return true, nil
}

// SetAggVote generate aggregate vote when voteSet have 2/3 majority
func (voteSet *VoteSet) SetAggVote() error {
	if voteSet == nil {
		return errors.New("nil vote set")
	}
	if voteSet.maj23 == nil {
		return errors.New("no 2/3 majority in voteSet")
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()

	blockKey := string(voteSet.maj23.Hash)
	votesByBlock, ok := voteSet.votesByBlock[blockKey]
	if !ok {
		return errors.New("no 2/3 majority blockKey")
	}
	bitArray := votesByBlock.bitArray.copy()

	sigs := make([]crypto.Signature, 0)
	for _, vote := range votesByBlock.votes {
		if vote != nil {
			sig, err := ConsensusCrypto.SignatureFromBytes(vote.Signature)
			if err != nil {
				return errors.New("invalid aggregate signature")
			}
			sigs = append(sigs, sig)
		}
	}
	aggr, err := crypto.ToAggregate(ConsensusCrypto)
	if err != nil {
		return err
	}
	aggSig, err := aggr.Aggregate(sigs)
	if err != nil {
		return err
	}

	aggVote := &AggVote{&tmtypes.QbftAggVote{
		ValidatorAddress: voteSet.valSet.Proposer.Address,
		ValidatorArray:   bitArray.QbftBitArray,
		Height:           voteSet.height,
		Round:            int32(voteSet.round),
		Timestamp:        time.Now().UnixNano(),
		Type:             uint32(voteSet.voteType),
		BlockID:          voteSet.maj23,
		Signature:        aggSig.Bytes(),
	}}
	// Verify aggVote
	err = aggVote.Verify(voteSet.chainID, voteSet.valSet)
	if err != nil {
		return err
	}
	voteSet.aggVote = aggVote
	return nil
}

// GetAggVote ...
func (voteSet *VoteSet) GetAggVote() *AggVote {
	if voteSet == nil {
		return nil
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	if voteSet.aggVote == nil {
		return nil
	}
	return voteSet.aggVote.Copy()
}

// SetPeerMaj23 If a peer claims that it has 2/3 majority for given blockKey, call this.
// NOTE: if there are too many peers, or too much peer churn,
// this can cause memory issues.
// TODO: implement ability to remove peers too
// NOTE: VoteSet must not be nil
func (voteSet *VoteSet) SetPeerMaj23(peerID string, blockID *tmtypes.QbftBlockID) {
	if voteSet == nil {
		PanicSanity("SetPeerMaj23() on nil VoteSet")
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()

	blockKey := string(blockID.Hash)

	// Make sure peer hasn't already told us something.
	if existing, ok := voteSet.peerMaj23s[peerID]; ok {
		if bytes.Equal(existing.Hash, blockID.Hash) {
			return // Nothing to do
		}
		return // TODO bad peer!
	}
	voteSet.peerMaj23s[peerID] = BlockID{blockID}

	// Create .votesByBlock entry if needed.
	votesByBlock, ok := voteSet.votesByBlock[blockKey]
	if ok {
		if votesByBlock.peerMaj23 {
			return // Nothing to do
		}
		votesByBlock.peerMaj23 = true // No need to copy votes, already there.
	} else {
		votesByBlock = newBlockVotes(true, voteSet.valSet.Size())
		voteSet.votesByBlock[blockKey] = votesByBlock
		// No need to copy votes, no votes to copy over.
	}
}

// BitArray ...
func (voteSet *VoteSet) BitArray() *BitArray {
	if voteSet == nil {
		return nil
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	return voteSet.votesBitArray.Copy()
}

// BitArrayByBlockID ...
func (voteSet *VoteSet) BitArrayByBlockID(blockID *tmtypes.QbftBlockID) *BitArray {
	if voteSet == nil {
		return nil
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	votesByBlock, ok := voteSet.votesByBlock[string(blockID.Hash)]
	if ok {
		return votesByBlock.bitArray.Copy()
	}
	return nil
}

// GetByIndex if validator has conflicting votes, returns "canonical" vote
func (voteSet *VoteSet) GetByIndex(valIndex int) *Vote {
	if voteSet == nil {
		return nil
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	return voteSet.votes[valIndex]
}

// GetByAddress ...
func (voteSet *VoteSet) GetByAddress(address []byte) *Vote {
	if voteSet == nil {
		return nil
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	valIndex, val := voteSet.valSet.GetByAddress(address)
	if val == nil {
		PanicSanity("GetByAddress(address) returned nil")
	}
	return voteSet.votes[valIndex]
}

// HasTwoThirdsMajority ...
func (voteSet *VoteSet) HasTwoThirdsMajority() bool {
	if voteSet == nil {
		return false
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	return voteSet.maj23 != nil
}

// IsCommit ...
func (voteSet *VoteSet) IsCommit() bool {
	if voteSet == nil {
		return false
	}
	if voteSet.voteType != VoteTypePrecommit {
		return false
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	return voteSet.maj23 != nil
}

// HasTwoThirdsAny ...
func (voteSet *VoteSet) HasTwoThirdsAny() bool {
	if voteSet == nil {
		return false
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	return voteSet.sum > voteSet.valSet.TotalVotingPower()*2/3
}

// HasAll ...
func (voteSet *VoteSet) HasAll() bool {
	return voteSet.sum == voteSet.valSet.TotalVotingPower()
}

// TwoThirdsMajority Returns either a blockhash (or nil) that received +2/3 majority.
// If there exists no such majority, returns (nil, PartSetHeader{}, false).
func (voteSet *VoteSet) TwoThirdsMajority() (tmtypes.QbftBlockID, bool) {
	if voteSet == nil {
		return tmtypes.QbftBlockID{}, false
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	if voteSet.maj23 != nil {
		blockID := tmtypes.QbftBlockID{Hash: make([]byte, len(voteSet.maj23.Hash))}
		copy(blockID.Hash, voteSet.maj23.Hash)
		return blockID, true
	}
	return tmtypes.QbftBlockID{}, false
}

func (voteSet *VoteSet) String() string {
	if voteSet == nil {
		return "nil-VoteSet"
	}
	return voteSet.StringIndented("")
}

// StringIndented ...
func (voteSet *VoteSet) StringIndented(indent string) string {
	voteStrings := make([]string, len(voteSet.votes))
	for i, vote := range voteSet.votes {
		if vote == nil {
			voteStrings[i] = "nil-QbftVote"
		} else {
			voteStrings[i] = vote.String()
		}
	}
	return Fmt(`VoteSet{
%s  H:%v R:%v T:%v +2/3:%X
%s  %v
%s  %v
%s  %v
%s}`,
		indent, voteSet.height, voteSet.round, voteSet.voteType, BlockID{voteSet.maj23}.String(),
		indent, strings.Join(voteStrings, "\n"+indent+"  "),
		indent, voteSet.votesBitArray,
		indent, voteSet.peerMaj23s,
		indent)
}

// StringShort ...
func (voteSet *VoteSet) StringShort() string {
	if voteSet == nil {
		return "nil-VoteSet"
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	return Fmt(`VoteSet{H:%v R:%v T:%v +2/3:%v %v %v}`,
		voteSet.height, voteSet.round, voteSet.voteType, BlockID{voteSet.maj23}.String(), voteSet.votesBitArray, voteSet.peerMaj23s)
}

// MakeCommit ...
func (voteSet *VoteSet) MakeCommit() *tmtypes.QbftCommit {
	if voteSet.voteType != VoteTypePrecommit {
		PanicSanity("Cannot MakeCommit unless VoteSet.Type is VoteTypePrecommit")
	}
	return voteSet.MakeCommonCommit()
}

// MakeCommit ...
func (voteSet *VoteSet) MakeCommonCommit() *tmtypes.QbftCommit {
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()

	if voteSet.voteType != VoteTypePrecommit && voteSet.voteType != VoteTypePrevote {
		PanicSanity("Cannot MakeCommonCommit unless VoteSet.Type is VoteTypePrecommit or VoteTypePrevote")
	}

	// Make sure we have a 2/3 majority
	if voteSet.maj23 == nil {
		PanicSanity("Cannot MakeCommonCommit unless a blockhash has +2/3")
	}

	// For every validator, get the precommit
	votesCopy := make([]*tmtypes.QbftVote, len(voteSet.votes))
	for i, item := range voteSet.votes {
		if item != nil {
			votesCopy[i] = item.QbftVote
		} else {
			votesCopy[i] = &tmtypes.QbftVote{}
		}
	}
	var aggVote *tmtypes.QbftAggVote
	if voteSet.aggVote != nil {
		copy := voteSet.aggVote.Copy()
		aggVote = copy.QbftAggVote
	}

	if voteSet.voteType == VoteTypePrecommit {
		return &tmtypes.QbftCommit{
			BlockID:    voteSet.maj23,
			Precommits: votesCopy,
			AggVote:    aggVote,
			VoteType:   uint32(VoteTypePrecommit),
		}
	} else {
		return &tmtypes.QbftCommit{
			BlockID:  voteSet.maj23,
			Prevotes: votesCopy,
			AggVote:  aggVote,
			VoteType: uint32(VoteTypePrevote),
		}
	}
}

func (voteSet *VoteSet) AddCommitVote(votes []*tmtypes.QbftVote, aggVote *tmtypes.QbftAggVote) error {
	for _, item := range votes {
		if item == nil || len(item.Signature) == 0 {
			continue
		}
		vote := &Vote{item}
		added, err := voteSet.AddVote(vote)
		if !added || err != nil {
			return err
		}
	}
	if aggVote != nil {
		aggVote := &AggVote{aggVote}
		added, err := voteSet.AddAggVote(aggVote)
		if !added || err != nil {
			return err
		}
	}
	if !voteSet.HasTwoThirdsMajority() {
		return errors.New("not have +2/3 votes")
	}
	return nil
}

//--------------------------------------------------------------------------------

/*
	Votes for a particular block
	There are two ways a *blockVotes gets created for a blockKey.
	1. first (non-conflicting) vote of a validator w/ blockKey (peerMaj23=false)
	2. A peer claims to have a 2/3 majority w/ blockKey (peerMaj23=true)
*/
type blockVotes struct {
	peerMaj23 bool      // peer claims to have maj23
	bitArray  *BitArray // valIndex -> hasVote?
	votes     []*Vote   // valIndex -> *QbftVote
	sum       int64     // vote sum
}

func newBlockVotes(peerMaj23 bool, numValidators int) *blockVotes {
	return &blockVotes{
		peerMaj23: peerMaj23,
		bitArray:  NewBitArray(numValidators),
		votes:     make([]*Vote, numValidators),
		sum:       0,
	}
}

func (vs *blockVotes) addVerifiedVote(vote *Vote, votingPower int64) {
	valIndex := int(vote.ValidatorIndex)
	if existing := vs.votes[valIndex]; existing == nil {
		vs.bitArray.SetIndex(valIndex, true)
		vs.votes[valIndex] = vote
		vs.sum += votingPower
	}
}

func (vs *blockVotes) getByIndex(index int) *Vote {
	if vs == nil {
		return nil
	}
	return vs.votes[index]
}

//--------------------------------------------------------------------------------

// VoteSetReader Common interface between *consensus.VoteSet and types.Commit
type VoteSetReader interface {
	Height() int64
	Round() int
	Type() byte
	Size() int
	BitArray() *BitArray
	GetByIndex(int) *Vote
	IsCommit() bool
	GetAggVote() *AggVote
}
