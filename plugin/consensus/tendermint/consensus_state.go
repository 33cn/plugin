// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tendermint

import (
	"bytes"
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/tendermint/types"
	tmtypes "github.com/33cn/plugin/plugin/dapp/valnode/types"
	siasync "github.com/NebulousLabs/Sia/sync"
)

//-----------------------------------------------------------------------------
// Config

const (
	proposalHeartbeatIntervalSeconds = 1
)

// Errors define
var (
	ErrInvalidProposalSignature = errors.New("Error invalid proposal signature")
	ErrInvalidProposalPOLRound  = errors.New("Error invalid proposal POL round")
	ErrAddingVote               = errors.New("Error adding vote")
	ErrAddingAggVote            = errors.New("Error adding aggregate vote")
	ErrVoteHeightMismatch       = errors.New("Error vote height mismatch")
)

//-----------------------------------------------------------------------------

var (
	msgQueueSize = 1000
)

// internally generated messages which may update the state
type timeoutInfo struct {
	Duration time.Duration        `json:"duration"`
	Height   int64                `json:"height"`
	Round    int                  `json:"round"`
	Step     ttypes.RoundStepType `json:"step"`
}

func (ti *timeoutInfo) String() string {
	return fmt.Sprintf("%v ; %d/%d %v", ti.Duration, ti.Height, ti.Round, ti.Step)
}

// ConsensusState handles execution of the consensus algorithm.
// It processes votes and proposals, and upon reaching agreement,
// commits blocks to the chain and executes them against the application.
// The internal state machine receives input from peers, the internal validator, and from a timer.
type ConsensusState struct {
	// config details
	client        *Client
	privValidator ttypes.PrivValidator // for signing votes

	// services for creating and executing blocks
	// TODO: encapsulate all of this in one "BlockManager"
	blockExec *BlockExecutor

	// internal state
	mtx siasync.TryMutex
	ttypes.RoundState
	state State // State until height-1.

	// state changes may be triggered by msgs from peers,
	// msgs from ourself, or by timeouts
	peerMsgQueue     chan MsgInfo
	internalMsgQueue chan MsgInfo
	timeoutTicker    TimeoutTicker

	// for tests where we want to limit the number of transitions the state makes
	nSteps int

	// some functions can be overwritten for testing
	decideProposal func(height int64, round int)
	doPrevote      func(height int64, round int)
	setProposal    func(proposal *tmtypes.Proposal) error

	broadcastChannel chan<- MsgInfo
	unicastChannel   chan<- MsgInfo
	ourID            ID
	status           uint32 // 0-stop, 1-start
	quit             chan struct{}

	txsAvailable      chan int64
	begCons           time.Time
	ProposalBlockHash []byte
}

// NewConsensusState returns a new ConsensusState.
func NewConsensusState(client *Client, state State, blockExec *BlockExecutor) *ConsensusState {
	cs := &ConsensusState{
		client:           client,
		blockExec:        blockExec,
		peerMsgQueue:     make(chan MsgInfo, msgQueueSize),
		internalMsgQueue: make(chan MsgInfo, msgQueueSize),
		timeoutTicker:    NewTimeoutTicker(),

		quit:         make(chan struct{}),
		txsAvailable: make(chan int64, 1),
		begCons:      time.Time{},
	}
	atomic.CompareAndSwapUint32(&cs.status, 0, 0)
	// set function defaults (may be overwritten before calling Start)
	cs.decideProposal = cs.defaultDecideProposal
	cs.doPrevote = cs.defaultDoPrevote
	cs.setProposal = cs.defaultSetProposal

	cs.updateToState(state)
	// Don't call scheduleRound0 yet.
	// We do that upon Start().
	cs.reconstructLastCommit(state)

	return cs
}

// SetOurID method
func (cs *ConsensusState) SetOurID(id ID) {
	cs.ourID = id
}

// SetBroadcastChannel method
func (cs *ConsensusState) SetBroadcastChannel(broadcastChannel chan<- MsgInfo) {
	cs.broadcastChannel = broadcastChannel
}

// SetUnicastChannel method
func (cs *ConsensusState) SetUnicastChannel(unicastChannel chan<- MsgInfo) {
	cs.unicastChannel = unicastChannel
}

// IsRunning method
func (cs *ConsensusState) IsRunning() bool {
	return atomic.LoadUint32(&cs.status) == 1
}

//----------------------------------------

// QueryRoundState returns a copy of the internal consensus state.
func (cs *ConsensusState) QueryRoundState() *ttypes.RoundState {
	if cs == nil {
		return nil
	}
	if !cs.mtx.TryLockTimed(time.Millisecond * 500) {
		return nil
	}
	defer cs.mtx.Unlock()

	rs := cs.RoundState
	return &rs
}

// GetState returns a copy of the chain state.
func (cs *ConsensusState) GetState() State {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	return cs.state.Copy()
}

// GetRoundState returns a copy of the internal consensus state.
func (cs *ConsensusState) GetRoundState() *ttypes.RoundState {
	// need avoid deadlock in gossipVotesRoutine
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	rs := cs.RoundState // copy
	return &rs
}

// GetValidators returns a copy of the current validators.
func (cs *ConsensusState) GetValidators() (int64, []*ttypes.Validator) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	return cs.state.LastBlockHeight, cs.state.Validators.Copy().Validators
}

// GetPrivValidator returns the private validator account for signing votes.
func (cs *ConsensusState) GetPrivValidator() ttypes.PrivValidator {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	pv := cs.privValidator
	return pv
}

// SetPrivValidator sets the private validator account for signing votes.
func (cs *ConsensusState) SetPrivValidator(priv ttypes.PrivValidator) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	cs.privValidator = priv
}

// SetTimeoutTicker sets the local timer. It may be useful to overwrite for testing.
func (cs *ConsensusState) SetTimeoutTicker(timeoutTicker TimeoutTicker) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	cs.timeoutTicker = timeoutTicker
}

// LoadCommit loads the commit for a given height.
func (cs *ConsensusState) LoadCommit(height int64) *tmtypes.TendermintCommit {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	if height == cs.client.csStore.LoadStateHeight() {
		return cs.client.csStore.LoadSeenCommit(height)
	}
	return cs.client.LoadBlockCommit(height + 1)
}

// Start It start first time starts the timeout checkTxsAvailable routine and receive routines.
func (cs *ConsensusState) Start() {
	if atomic.CompareAndSwapUint32(&cs.status, 0, 1) {
		cs.timeoutTicker.Start()

		go cs.checkTxsAvailable()
		// now start the receiveRoutine
		go cs.receiveRoutine(0)

		// schedule the first round!
		// use GetRoundState so we don't race the receiveRoutine for access
		cs.scheduleRound0(cs.GetRoundState())
	}
}

// Stop timer and receive routine
func (cs *ConsensusState) Stop() {
	atomic.CompareAndSwapUint32(&cs.status, 1, 0)
	cs.timeoutTicker.Stop()
	cs.quit <- struct{}{}
}

//------------------------------------------------------------
// internal functions for managing the state

func (cs *ConsensusState) updateHeight(height int64) {
	cs.Height = height
}

func (cs *ConsensusState) updateRoundStep(round int, step ttypes.RoundStepType) {
	cs.Round = round
	cs.Step = step
}

// enterNewRound(height, 0) at cs.StartTime.
func (cs *ConsensusState) scheduleRound0(rs *ttypes.RoundState) {
	sleepDuration := rs.StartTime.Sub(time.Now()) // nolint: gotype, gosimple
	cs.scheduleTimeout(sleepDuration, rs.Height, 0, ttypes.RoundStepNewHeight)
}

// Attempt to schedule a timeout (by sending timeoutInfo on the tickChan)
func (cs *ConsensusState) scheduleTimeout(duration time.Duration, height int64, round int, step ttypes.RoundStepType) {
	cs.timeoutTicker.ScheduleTimeout(timeoutInfo{duration, height, round, step})
}

// send a msg into the receiveRoutine regarding our own proposal, block part, or vote
func (cs *ConsensusState) sendInternalMessage(mi MsgInfo) {
	select {
	case cs.internalMsgQueue <- mi:
	default:
		// NOTE: using the go-routine means our votes can
		// be processed out of order.
		// TODO: use CList here for strict determinism and
		// attempt push to internalMsgQueue in receiveRoutine
		tendermintlog.Info("Internal msg queue is full. Using a go-routine")
		go func() { cs.internalMsgQueue <- mi }()
	}
}

// Reconstruct LastCommit from SeenCommit, which we saved along with the block,
// (which happens even before saving the state)
func (cs *ConsensusState) reconstructLastCommit(state State) {
	if state.LastBlockHeight == 0 {
		return
	}
	seenCommit := cs.client.csStore.LoadSeenCommit(state.LastBlockHeight)
	seenCommitC := ttypes.Commit{TendermintCommit: seenCommit}
	lastPrecommits := ttypes.NewVoteSet(state.ChainID, state.LastBlockHeight, seenCommitC.Round(), ttypes.VoteTypePrecommit, state.LastValidators)
	for _, item := range seenCommit.Precommits {
		if item == nil || len(item.Signature) == 0 {
			continue
		}
		precommit := &ttypes.Vote{Vote: item}
		added, err := lastPrecommits.AddVote(precommit)
		if !added || err != nil {
			panic(fmt.Sprintf("Failed to reconstruct LastCommit: %v", err))
		}
	}
	if seenCommit.AggVote != nil {
		aggVote := &ttypes.AggVote{AggVote: seenCommit.AggVote}
		added, err := lastPrecommits.AddAggVote(aggVote)
		if !added || err != nil {
			panic(fmt.Sprintf("Failed to reconstruct LastCommit: %v", err))
		}
	}
	if !lastPrecommits.HasTwoThirdsMajority() {
		panic("Failed to reconstruct LastCommit: Does not have +2/3 maj")
	}
	cs.LastCommit = lastPrecommits
}

// Updates ConsensusState and increments height to match that of state.
// The round becomes 0 and cs.Step becomes ttypes.RoundStepNewHeight.
func (cs *ConsensusState) updateToState(state State) {
	if cs.CommitRound > -1 && 0 < cs.Height && cs.Height != state.LastBlockHeight {
		panic(fmt.Sprintf("updateToState expected state height of %v but found %v", cs.Height, state.LastBlockHeight))
	}
	if !cs.state.IsEmpty() && cs.state.LastBlockHeight+1 != cs.Height {
		// This might happen when someone else is mutating cs.state.
		// Someone forgot to pass in state.Copy() somewhere?!
		panic(fmt.Sprintf("Inconsistent cs.state.LastBlockHeight+1 %v vs cs.Height %v", cs.state.LastBlockHeight+1, cs.Height))
	}

	// If state isn't further out than cs.state, just ignore.
	// This happens when SwitchToConsensus() is called in the reactor.
	// We don't want to reset e.g. the Votes.
	if !cs.state.IsEmpty() && (state.LastBlockHeight <= cs.state.LastBlockHeight) {
		tendermintlog.Info("Ignoring updateToState()", "newHeight", state.LastBlockHeight+1, "oldHeight", cs.state.LastBlockHeight+1)
		return
	}
	// disable gossip votes
	if useAggSig && gossipVotes.Load().(bool) {
		gossipVotes.Store(false)
	}

	// Reset fields based on state.
	validators := state.Validators
	lastPrecommits := (*ttypes.VoteSet)(nil)
	if cs.CommitRound > -1 && cs.Votes != nil {
		if !cs.Votes.Precommits(cs.CommitRound).HasTwoThirdsMajority() {
			panic("updateToState called but last Precommit round didn't have +2/3")
		}
		lastPrecommits = cs.Votes.Precommits(cs.CommitRound)
	}

	// Next desired block height
	height := state.LastBlockHeight + 1

	// RoundState fields
	cs.updateHeight(height)
	cs.updateRoundStep(0, ttypes.RoundStepNewHeight)
	if cs.CommitTime.IsZero() {
		// "Now" makes it easier to sync up dev nodes.
		// We add timeoutCommit to allow transactions
		// to be gathered for the first block.
		// And alternative solution that relies on clocks:
		//  cs.StartTime = state.LastBlockTime.Add(timeoutCommit)
		cs.StartTime = cs.Commit(time.Now())
	} else {
		cs.StartTime = cs.Commit(cs.CommitTime)
	}
	cs.Validators = validators
	cs.Proposal = nil
	cs.ProposalBlock = nil
	cs.ProposalBlockHash = nil
	cs.LockedRound = -1
	cs.LockedBlock = nil
	cs.ValidRound = -1
	cs.ValidBlock = nil
	cs.Votes = ttypes.NewHeightVoteSet(state.ChainID, height, validators)
	cs.CommitRound = -1
	cs.LastCommit = lastPrecommits
	cs.LastValidators = state.LastValidators
	cs.begCons = time.Time{}

	cs.state = state

	// Finally, broadcast RoundState
	cs.newStep()
}

func (cs *ConsensusState) newStep() {
	if cs.broadcastChannel != nil {
		cs.broadcastChannel <- MsgInfo{TypeID: ttypes.NewRoundStepID, Msg: cs.RoundStateMessage(), PeerID: "", PeerIP: ""}
	}
	cs.nSteps++
}

//-----------------------------------------
// the main go routines

// receiveRoutine handles messages which may cause state transitions.
// it's argument (n) is the number of messages to process before exiting - use 0 to run forever
// It keeps the RoundState and is the only thing that updates it.
// Updates (state transitions) happen on timeouts, complete proposals, and 2/3 majorities.
// ConsensusState must be locked before any internal state is updated.
func (cs *ConsensusState) receiveRoutine(maxSteps int) {
	defer func() {
		if r := recover(); r != nil {
			tendermintlog.Error("CONSENSUS FAILURE!!!", "err", r, "stack", string(debug.Stack()))
		}
	}()

	for {
		if maxSteps > 0 {
			if cs.nSteps >= maxSteps {
				tendermintlog.Info("reached max steps. exiting receive routine")
				cs.nSteps = 0
				return
			}
		}
		rs := cs.RoundState
		var mi MsgInfo

		select {
		case height := <-cs.txsAvailable:
			cs.handleTxsAvailable(height)
		case mi = <-cs.peerMsgQueue:
			// handles proposals, block parts, votes
			// may generate internal events (votes, complete proposals, 2/3 majorities)
			cs.handleMsg(mi)
		case mi = <-cs.internalMsgQueue:
			// handles proposals, block parts, votes
			cs.handleMsg(mi)
		case ti := <-cs.timeoutTicker.Chan(): // tockChan:
			// if the timeout is relevant to the rs
			// go to the next step
			cs.handleTimeout(ti, rs)
		case <-cs.quit:
			// NOTE: the internalMsgQueue may have signed messages from our
			// priv_val that haven't hit the WAL, but its ok because
			// priv_val tracks LastSig
			return
		}
	}
}

// state transitions on complete-proposal, 2/3-any, 2/3-one
func (cs *ConsensusState) handleMsg(mi MsgInfo) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	var err error
	msg, peerID, peerIP := mi.Msg, string(mi.PeerID), mi.PeerIP
	switch msg := msg.(type) {
	case *tmtypes.Proposal:
		// will not cause transition.
		// once proposal is set, we can receive block parts
		err = cs.setProposal(msg)
		if err != nil {
			tendermintlog.Error("handleMsg proposalMsg failed", "msg", msg, "error", err)
		}
	case *tmtypes.TendermintBlock:
		// if the block is complete, we'll enterPrevote or tryFinalizeCommit
		err = cs.addProposalBlock(msg)
		if err != nil && cs.Round != int(msg.Header.Round) {
			tendermintlog.Debug("Received block from wrong round", "height", cs.Height, "csRound", cs.Round, "blockRound", msg.Header.Round)
			err = nil
		}
	case *tmtypes.Vote:
		// attempt to add the vote and dupeout the validator if its a duplicate signature
		// if the vote gives us a 2/3-any or 2/3-one, we transition
		err = cs.tryAddVote(msg, peerID, peerIP)

		//if err == ErrAddingVote {
		//	TODO: punish peer
		//}

		// NOTE: the vote is broadcast to peers by the reactor listening
		// for vote events

		// TODO: If rs.Height == vote.Height && rs.Round < vote.Round,
		// the peer is sending us CatchupCommit precommits.
		// We could make note of this and help filter in broadcastHasVoteMessage().
	case *tmtypes.AggVote:
		err = cs.tryAddAggVote(msg, peerID)
	default:
		tendermintlog.Error("Unknown msg type", msg.String(), "peerid", peerID, "peerip", peerIP)
	}
	if err != nil {
		//tendermintlog.Error("Error with msg", "type", reflect.TypeOf(msg), "peerid", peerID, "peerip", peerIP, "err", err, "msg", msg)
	}
}

func (cs *ConsensusState) handleTimeout(ti timeoutInfo, rs ttypes.RoundState) {
	tendermintlog.Debug("Received tock", "timeout", ti.String())

	// timeouts must be for current height, round, step
	if ti.Height != rs.Height || ti.Round < rs.Round || (ti.Round == rs.Round && ti.Step < rs.Step) {
		tendermintlog.Debug("Ignoring tock because we're ahead", "height", rs.Height, "round", rs.Round, "step", rs.Step)
		return
	}

	// the timeout will now cause a state transition
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	switch ti.Step {
	case ttypes.RoundStepNewHeight:
		// NewRound event fired from enterNewRound.
		// XXX: should we fire timeout here (for timeout commit)?
		cs.enterNewRound(ti.Height, 0)
	case ttypes.RoundStepNewRound:
		cs.enterPropose(ti.Height, 0)
	case ttypes.RoundStepPropose:
		cs.enterPrevote(ti.Height, ti.Round)
	case ttypes.RoundStepPrevoteWait:
		cs.enterPrecommit(ti.Height, ti.Round)
	case ttypes.RoundStepPrecommitWait:
		cs.enterPrecommit(ti.Height, ti.Round)
		cs.enterNewRound(ti.Height, ti.Round+1)
	case ttypes.RoundStepAggPrevoteWait:
		cs.enterAggPrevoteWait(ti.Height, ti.Round)
	case ttypes.RoundStepAggPrecommitWait:
		cs.enterAggPrecommitWait(ti.Height, ti.Round)
	default:
		panic(fmt.Sprintf("Invalid timeout step: %v", ti.Step))
	}

}

func (cs *ConsensusState) checkTxsAvailable() {
	for {
		select {
		case height := <-cs.client.TxsAvailable():
			rs := cs.GetRoundState()
			tendermintlog.Debug(fmt.Sprintf("checkTxsAvailable. Current: %v/%v/%v", rs.Height, rs.Round, rs.Step), "height", height)
			if rs.Height != height {
				tendermintlog.Info(fmt.Sprintf("blockchain(H: %v) and consensus(H: %v) are not sync", height, rs.Height))
				break
			}
			cs.txsAvailable <- height
		case <-cs.client.StopC():
			tendermintlog.Info("checkTxsAvailable quit")
			return
		}
	}
}

func (cs *ConsensusState) checkProposalComplete() bool {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	return cs.isProposalComplete()
}

func (cs *ConsensusState) handleTxsAvailable(height int64) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	// we only need to do this for round 0
	if cs.Round != 0 {
		return
	}

	switch cs.Step {
	case ttypes.RoundStepNewHeight: // timeoutCommit phase
		// +1ms to ensure RoundStepNewRound timeout always happens after RoundStepNewHeight
		timeoutCommit := cs.StartTime.Sub(time.Now()) + 1*time.Millisecond
		cs.scheduleTimeout(timeoutCommit, height, 0, ttypes.RoundStepNewRound)
	case ttypes.RoundStepNewRound: // after timeoutCommit
		cs.enterPropose(height, 0)
	}
}

//-----------------------------------------------------------------------------
// State functions
// Used internally by handleTimeout and handleMsg to make state transitions

// Enter: `timeoutNewHeight` by startTime (commitTime+timeoutCommit),
// 	or, if SkipTimeout==true, after receiving all precommits from (height,round-1)
// Enter: `timeoutPrecommits` after any +2/3 precommits from (height,round-1)
// Enter: +2/3 precommits for nil at (height,round-1)
// Enter: +2/3 prevotes any or +2/3 precommits for block or any from (height, round)
// NOTE: cs.StartTime was already set for height.
func (cs *ConsensusState) enterNewRound(height int64, round int) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && cs.Step != ttypes.RoundStepNewHeight) {
		tendermintlog.Debug(fmt.Sprintf("enterNewRound(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
		return
	}

	if now := time.Now(); cs.StartTime.After(now) {
		tendermintlog.Info("Need to set a buffer and log message here for sanity.", "startTime", cs.StartTime, "now", now)
	}

	tendermintlog.Info(fmt.Sprintf("enterNewRound(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	// disable gossip votes
	if useAggSig && gossipVotes.Load().(bool) {
		gossipVotes.Store(false)
	}
	// Increment validators if necessary
	validators := cs.Validators
	if cs.Round < round {
		validators = validators.Copy()
		validators.IncrementAccum(round - cs.Round)
		tendermintlog.Debug("enterNewRound validator changed", "cs.Round", cs.Round, "round", round)
	}
	tendermintlog.Debug("enterNewRound proposer", "proposer", validators.Proposer, "validators", validators)
	// Setup new round
	// we don't fire newStep for this step,
	// but we fire an event, so update the round step first
	cs.updateRoundStep(round, ttypes.RoundStepNewRound)
	cs.Validators = validators
	if round == 0 {
		// We've already reset these upon new height,
		// and meanwhile we might have received a proposal
		// for round 0.
	} else {
		tendermintlog.Info("Resetting Proposal info")
		cs.Proposal = nil
		cs.ProposalBlock = nil
		cs.ProposalBlockHash = nil
	}
	cs.Votes.SetRound(round + 1) // also track next round (round+1) to allow round-skipping

	//cs.eventBus.PublishEventNewRound(cs.RoundStateEvent())
	cs.broadcastChannel <- MsgInfo{TypeID: ttypes.NewRoundStepID, Msg: cs.RoundStateMessage(), PeerID: "", PeerIP: ""}

	// Wait for txs to be available in the mempool
	// before we enterPropose in round 0.
	waitForTxs := cs.WaitForTxs() && round == 0
	if waitForTxs {
		if createEmptyBlocksInterval > 0 {
			cs.scheduleTimeout(cs.EmptyBlocksInterval(), height, round, ttypes.RoundStepNewRound)
		}
		go cs.proposalHeartbeat(height, round)
	} else {
		cs.enterPropose(height, round)
	}
}

func (cs *ConsensusState) proposalHeartbeat(height int64, round int) {
	counter := 0
	addr := cs.privValidator.GetAddress()
	valIndex, _ := cs.Validators.GetByAddress(addr)
	chainID := cs.state.ChainID
	for {
		rs := cs.GetRoundState()
		// if we've already moved on, no need to send more heartbeats
		if rs.Step > ttypes.RoundStepNewRound || rs.Round > round || rs.Height > height {
			return
		}
		heartbeat := &tmtypes.Heartbeat{
			Height:           rs.Height,
			Round:            int32(rs.Round),
			Sequence:         int32(counter),
			ValidatorAddress: addr,
			ValidatorIndex:   int32(valIndex),
		}
		heartbeatMsg := &ttypes.Heartbeat{Heartbeat: heartbeat}
		err := cs.GetPrivValidator().SignHeartbeat(chainID, heartbeatMsg)
		if err != nil {
			tendermintlog.Error("SignHeartbeat failed", "err", err)
			continue
		}
		cs.broadcastChannel <- MsgInfo{TypeID: ttypes.ProposalHeartbeatID, Msg: heartbeat, PeerID: "", PeerIP: ""}
		cs.broadcastChannel <- MsgInfo{TypeID: ttypes.NewRoundStepID, Msg: rs.RoundStateMessage(), PeerID: "", PeerIP: ""}
		counter++
		time.Sleep(proposalHeartbeatIntervalSeconds * time.Second)
	}
}

// Enter (CreateEmptyBlocks): from enterNewRound(height,round)
// Enter (CreateEmptyBlocks, CreateEmptyBlocksInterval > 0 ): after enterNewRound(height,round), after timeout of CreateEmptyBlocksInterval
// Enter (!CreateEmptyBlocks) : after enterNewRound(height,round), once txs are in the mempool
func (cs *ConsensusState) enterPropose(height int64, round int) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && ttypes.RoundStepPropose <= cs.Step) {
		tendermintlog.Debug(fmt.Sprintf("enterPropose(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
		return
	}
	tendermintlog.Info(fmt.Sprintf("enterPropose(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
	if cs.Round == 0 && cs.begCons.IsZero() {
		cs.begCons = time.Now()
	}

	defer func() {
		// Done enterPropose:
		cs.updateRoundStep(round, ttypes.RoundStepPropose)
		cs.newStep()

		// If we have the whole proposal + POL, then goto Prevote now.
		// else, we'll enterPrevote when the rest of the proposal is received (in AddProposalBlockPart),
		// or else after timeoutPropose
		if cs.isProposalComplete() {
			cs.enterPrevote(height, cs.Round)
		}
	}()

	// If we don't get the proposal and all block parts quick enough, enterPrevote
	cs.scheduleTimeout(cs.Propose(round), height, round, ttypes.RoundStepPropose)

	// Nothing more to do if we're not a validator
	if cs.privValidator == nil {
		tendermintlog.Debug("This node is not a validator")
		return
	}

	// if not a validator, we're done
	if !cs.Validators.HasAddress(cs.privValidator.GetAddress()) {
		tendermintlog.Debug("This node is not a validator",
			"privValidator", fmt.Sprintf("%X", ttypes.Fingerprint(cs.privValidator.GetAddress())),
			"Validators", cs.Validators.String())
		return
	}
	tendermintlog.Debug("This node is a validator")

	if cs.isProposer() {
		tendermintlog.Info("enterPropose: Our turn to propose",
			"proposer", fmt.Sprintf("%X", ttypes.Fingerprint(cs.Validators.GetProposer().Address)),
			"privValidator", fmt.Sprintf("%X", ttypes.Fingerprint(cs.privValidator.GetAddress())))
		cs.decideProposal(height, round)
	} else {
		tendermintlog.Info("enterPropose: Not our turn to propose",
			"proposer", fmt.Sprintf("%X", ttypes.Fingerprint(cs.Validators.GetProposer().Address)),
			"privValidator", fmt.Sprintf("%X", ttypes.Fingerprint(cs.privValidator.GetAddress())))
	}
}

func (cs *ConsensusState) isProposer() bool {
	return bytes.Equal(cs.Validators.GetProposer().Address, cs.privValidator.GetAddress())
}

func (cs *ConsensusState) getProposerID() ID {
	pub := cs.Validators.GetProposer().PubKey
	pubkey, err := ttypes.ConsensusCrypto.PubKeyFromBytes(pub)
	if err != nil {
		tendermintlog.Error("getProposerID fail", "err", err)
	}
	return GenIDByPubKey(pubkey)
}

func (cs *ConsensusState) defaultDecideProposal(height int64, round int) {

	var block *ttypes.TendermintBlock

	// Decide on block
	if cs.ValidBlock != nil {
		if preExec {
			// If there is valid block, PreExec that.
			pblockNew := cs.client.PreExecBlock(cs.ValidBlock.Data, false)
			if pblockNew == nil {
				tendermintlog.Error("defaultDecideProposal PreExecBlock fail")
				return
			}
			cs.ValidBlock.Data = pblockNew
		}
		block = cs.ValidBlock
	} else {
		// Create a new proposal block from state/txs from the mempool.
		block = cs.createProposalBlock()
		if block == nil { // on error
			return
		}
	}

	// Make proposal
	propBlockID := tmtypes.BlockID{Hash: block.Hash()}
	proposal := ttypes.NewProposal(height, round, block.Hash(), cs.ValidRound, propBlockID)
	if err := cs.privValidator.SignProposal(cs.state.ChainID, proposal); err == nil {
		// send proposal and block on internal msg queue
		cs.sendInternalMessage(MsgInfo{ttypes.ProposalID, &proposal.Proposal, cs.ourID, ""})
		cs.sendInternalMessage(MsgInfo{ttypes.ProposalBlockID, block.TendermintBlock, cs.ourID, ""})
		tendermintlog.Info("Signed proposal", "height", height, "round", round, "proposal", proposal)
	} else {
		tendermintlog.Error("enterPropose: Error signing proposal", "height", height, "round", round, "err", err)
	}
}

// Returns true if the proposal block is complete &&
// (if POLRound was proposed, we have +2/3 prevotes from there).
func (cs *ConsensusState) isProposalComplete() bool {
	if cs.Proposal == nil || cs.ProposalBlock == nil {
		return false
	}
	// we have the proposal. if there's a POLRound,
	// make sure we have the prevotes from it too
	if cs.Proposal.POLRound < 0 {
		return true
	}
	// if this is false the proposer is lying or we haven't received the POL yet
	return cs.Votes.Prevotes(int(cs.Proposal.POLRound)).HasTwoThirdsMajority()
}

// Create the next block to propose and return it.
// Returns nil block upon error.
// NOTE: keep it side-effect free for clarity.
func (cs *ConsensusState) createProposalBlock() (block *ttypes.TendermintBlock) {
	var commit *tmtypes.TendermintCommit
	if cs.Height == 1 {
		// We're creating a proposal for the first block.
		// The commit is empty, but not nil.
		commit = &tmtypes.TendermintCommit{}
	} else if cs.LastCommit.HasTwoThirdsMajority() {
		// Make the commit from LastCommit
		commit = cs.LastCommit.MakeCommit()
	} else {
		// This shouldn't happen.
		tendermintlog.Error("enterPropose: Cannot propose anything: No commit for the previous block.")
		return nil
	}

	// Mempool validated transactions
	beg := time.Now()
	pblock := cs.client.BuildBlock()
	if pblock == nil {
		tendermintlog.Error("createProposalBlock BuildBlock fail")
		return nil
	}
	tendermintlog.Info(fmt.Sprintf("createProposalBlock BuildBlock. Current: %v/%v/%v", cs.Height, cs.Round, cs.Step),
		"height", pblock.Height, "txs-len", len(pblock.Txs), "cost", types.Since(beg))

	if pblock.Height != cs.Height {
		tendermintlog.Error("pblock.Height is not equal to cs.Height")
		return nil
	}

	proposerAddr := cs.privValidator.GetAddress()
	block = cs.state.MakeBlock(cs.Height, int64(cs.Round), pblock, commit, proposerAddr)
	baseTx := cs.createBaseTx(block.TendermintBlock)
	if baseTx == nil {
		tendermintlog.Error("createProposalBlock createBaseTx fail")
		return nil
	}
	cfg := cs.client.GetQueueClient().GetConfig()

	block.Data.Txs[0] = baseTx
	//需要首先对交易进行排序然后再计算TxHash
	if cfg.IsFork(block.Data.Height, "ForkRootHash") {
		block.Data.Txs = types.TransactionSort(block.Data.Txs)
	}

	block.Data.TxHash = merkle.CalcMerkleRoot(cfg, block.Data.Height, block.Data.Txs)
	if preExec {
		pblockNew := cs.client.PreExecBlock(block.Data, false)
		if pblockNew == nil {
			tendermintlog.Error("createProposalBlock PreExecBlock fail")
			return nil
		}
		block.Data = pblockNew
	}
	return block
}

func (cs *ConsensusState) createBaseTx(block *tmtypes.TendermintBlock) *types.Transaction {
	var state *tmtypes.State
	if cs.Height == 1 {
		genState := cs.client.GenesisState()
		if genState == nil {
			return nil
		}
		state = SaveState(*genState)
	} else {
		state = cs.client.csStore.LoadStateFromStore()
		if state == nil {
			return nil
		}
	}
	tx := CreateBlockInfoTx(cs.client.pubKey, state, block)
	return tx
}

// Enter: `timeoutPropose` after entering Propose.
// Enter: proposal block and POL is ready.
// Enter: any +2/3 prevotes for future round.
// Prevote for LockedBlock if we're locked, or ProposalBlock if valid.
// Otherwise vote nil.
func (cs *ConsensusState) enterPrevote(height int64, round int) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && ttypes.RoundStepPrevote <= cs.Step) {
		tendermintlog.Debug(fmt.Sprintf("enterPrevote(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
		return
	}

	defer func() {
		// Done enterPrevote:
		cs.updateRoundStep(round, ttypes.RoundStepPrevote)
		cs.newStep()
	}()

	tendermintlog.Info(fmt.Sprintf("enterPrevote(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step), "cost", types.Since(cs.begCons))

	// Sign and broadcast vote as necessary
	cs.doPrevote(height, round)

	// Once `addVote` hits any +2/3 prevotes, we will go to PrevoteWait
	// (so we have more time to try and collect +2/3 prevotes for a single block)
}

func (cs *ConsensusState) defaultDoPrevote(height int64, round int) {
	// If a block is locked, prevote that.
	if cs.LockedBlock != nil {
		tendermintlog.Info("enterPrevote: Block was locked", "hash", fmt.Sprintf("%X", cs.LockedBlock.Hash()))
		cs.signAddVote(ttypes.VoteTypePrevote, cs.LockedBlock.Hash())
		return
	}

	// If ProposalBlock is nil, prevote nil.
	if cs.ProposalBlock == nil {
		tendermintlog.Info("enterPrevote: ProposalBlock is nil")
		cs.signAddVote(ttypes.VoteTypePrevote, nil)
		return
	}

	// Validate proposal block
	err := cs.blockExec.ValidateBlock(cs.state, cs.ProposalBlock)
	if err != nil {
		// ProposalBlock is invalid, prevote nil.
		tendermintlog.Error("enterPrevote: ProposalBlock is invalid", "err", err)
		cs.signAddVote(ttypes.VoteTypePrevote, nil)
		return
	}

	if preExec && !cs.isProposer() {
		// PreExec proposal block
		blockCopy := *cs.ProposalBlock.Data
		blockNew := cs.client.PreExecBlock(&blockCopy, true)
		if blockNew == nil {
			tendermintlog.Error("enterPrevote: PreExec ProposalBlock fail")
			cs.signAddVote(ttypes.VoteTypePrevote, nil)
			return
		}
		cfg := cs.client.GetQueueClient().GetConfig()
		if !bytes.Equal(blockNew.Hash(cfg), cs.ProposalBlock.Data.Hash(cfg)) {
			tendermintlog.Error("enterPrevote: PreExec ProposalBlock has change")
			cs.signAddVote(ttypes.VoteTypePrevote, nil)
			return
		}
	}

	// Prevote cs.ProposalBlock
	// NOTE: the proposal signature is validated when it is received,
	// and the proposal block parts are validated as they are received (against the merkle hash in the proposal)
	tendermintlog.Info("enterPrevote: ProposalBlock is valid")
	cs.signAddVote(ttypes.VoteTypePrevote, cs.ProposalBlock.Hash())
}

// Enter: send prevote for aggregate at next round.
func (cs *ConsensusState) enterAggPrevoteWait(height int64, round int) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && ttypes.RoundStepAggPrevoteWait <= cs.Step) {
		tendermintlog.Debug(fmt.Sprintf("enterAggPrevoteWait(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
		return
	}
	tendermintlog.Info(fmt.Sprintf("enterAggPrevoteWait(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	// enable gossip votes when timeout waiting for aggregate prevote
	gossipVotes.Store(true)
}

// Enter: any +2/3 prevotes at next round.
func (cs *ConsensusState) enterPrevoteWait(height int64, round int) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && ttypes.RoundStepPrevoteWait <= cs.Step) {
		tendermintlog.Debug(fmt.Sprintf("enterPrevoteWait(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
		return
	}
	if !cs.Votes.Prevotes(round).HasTwoThirdsAny() {
		panic(fmt.Sprintf("enterPrevoteWait(%v/%v), but Prevotes does not have any +2/3 votes", height, round))
	}
	tendermintlog.Info(fmt.Sprintf("enterPrevoteWait(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	defer func() {
		// Done enterPrevoteWait:
		cs.updateRoundStep(round, ttypes.RoundStepPrevoteWait)
		cs.newStep()
	}()

	// enable gossip votes in case other validators enterAggPrevoteWait
	if useAggSig && !cs.isProposer() && cs.Votes.Prevotes(round).GetAggVote() == nil {
		gossipVotes.Store(true)
	}

	// Wait for some more prevotes; enterPrecommit
	cs.scheduleTimeout(cs.Prevote(round), height, round, ttypes.RoundStepPrevoteWait)
}

// Enter: `timeoutPrevote` after any +2/3 prevotes.
// Enter: +2/3 precomits for block or nil.
// Enter: any +2/3 precommits for next round.
// Lock & precommit the ProposalBlock if we have enough prevotes for it (a POL in this round)
// else, unlock an existing lock and precommit nil if +2/3 of prevotes were nil,
// else, precommit nil otherwise.
func (cs *ConsensusState) enterPrecommit(height int64, round int) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && ttypes.RoundStepPrecommit <= cs.Step) {
		tendermintlog.Debug(fmt.Sprintf("enterPrecommit(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
		return
	}

	tendermintlog.Info(fmt.Sprintf("enterPrecommit(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step), "cost", types.Since(cs.begCons))

	defer func() {
		// Done enterPrecommit:
		cs.updateRoundStep(round, ttypes.RoundStepPrecommit)
		cs.newStep()
	}()

	blockID, ok := cs.Votes.Prevotes(round).TwoThirdsMajority()

	// If we don't have a polka, we must precommit nil
	if !ok {
		if cs.LockedBlock != nil {
			tendermintlog.Info("enterPrecommit: No +2/3 prevotes during enterPrecommit while we're locked. Precommitting nil")
		} else {
			tendermintlog.Info("enterPrecommit: No +2/3 prevotes during enterPrecommit. Precommitting nil")
		}
		cs.signAddVote(ttypes.VoteTypePrecommit, nil)
		return
	}

	// the latest POLRound should be this round
	polRound, _ := cs.Votes.POLInfo()
	if polRound < round {
		panic(fmt.Sprintf("This POLRound should be %v but got %v", round, polRound))
	}

	// +2/3 prevoted nil. Unlock and precommit nil.
	if len(blockID.Hash) == 0 {
		if cs.LockedBlock == nil {
			tendermintlog.Info("enterPrecommit: +2/3 prevoted for nil.")
		} else {
			tendermintlog.Info("enterPrecommit: +2/3 prevoted for nil. Unlocking")
			cs.LockedRound = -1
			cs.LockedBlock = nil
		}
		cs.signAddVote(ttypes.VoteTypePrecommit, nil)
		return
	}

	// At this point, +2/3 prevoted for a particular block.

	// If we're already locked on that block, precommit it, and update the LockedRound
	if cs.LockedBlock.HashesTo(blockID.Hash) {
		tendermintlog.Info("enterPrecommit: +2/3 prevoted locked block. Relocking", "hash", fmt.Sprintf("%X", blockID.Hash))
		cs.LockedRound = round
		cs.signAddVote(ttypes.VoteTypePrecommit, blockID.Hash)
		return
	}

	// If +2/3 prevoted for proposal block, stage and precommit it
	if cs.ProposalBlock.HashesTo(blockID.Hash) {
		tendermintlog.Info("enterPrecommit: +2/3 prevoted proposal block. Locking", "hash", fmt.Sprintf("%X", blockID.Hash))
		// Validate the block.
		if err := cs.blockExec.ValidateBlock(cs.state, cs.ProposalBlock); err != nil {
			panic(fmt.Sprintf("enterPrecommit: +2/3 prevoted for an invalid block: %v", err))
		}
		cs.LockedRound = round
		cs.LockedBlock = cs.ProposalBlock
		cs.signAddVote(ttypes.VoteTypePrecommit, blockID.Hash)
		return
	}

	// There was a polka in this round for a block we don't have.
	// Fetch that block, unlock, and precommit nil.
	// The +2/3 prevotes for this round is the POL for our unlock.
	// TODO: In the future save the POL prevotes for justification.
	tendermintlog.Info("enterPrecommit: +2/3 prevotes for a block we do not have. Precommitting nil", "hash", fmt.Sprintf("%X", blockID.Hash))
	cs.LockedRound = -1
	cs.LockedBlock = nil
	if !bytes.Equal(cs.ProposalBlockHash, blockID.Hash) {
		cs.ProposalBlock = nil
		cs.ProposalBlockHash = blockID.Hash
	}
	cs.signAddVote(ttypes.VoteTypePrecommit, nil)
}

// Enter: send precommit for aggregate at next round.
func (cs *ConsensusState) enterAggPrecommitWait(height int64, round int) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && ttypes.RoundStepAggPrecommitWait < cs.Step) {
		tendermintlog.Debug(fmt.Sprintf("enterAggPrecommitWait(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
		return
	}
	tendermintlog.Info(fmt.Sprintf("enterAggPrecommitWait(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	// enable gossip votes when timeout waiting for aggregate precommit
	gossipVotes.Store(true)

	//addr := cs.privValidator.GetAddress()
	//vote := cs.Votes.Precommits(round).GetByAddress(addr)
	//cs.broadcastChannel <- MsgInfo{TypeID: ttypes.VoteID, Msg: vote.Vote, PeerID: cs.ourID, PeerIP: ""}
}

// Enter: any +2/3 precommits for next round.
func (cs *ConsensusState) enterPrecommitWait(height int64, round int) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && ttypes.RoundStepPrecommitWait <= cs.Step) {
		tendermintlog.Debug(fmt.Sprintf("enterPrecommitWait(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
		return
	}
	if !cs.Votes.Precommits(round).HasTwoThirdsAny() {
		panic(fmt.Sprintf("enterPrecommitWait(%v/%v), but Precommits does not have any +2/3 votes", height, round))
	}
	tendermintlog.Info(fmt.Sprintf("enterPrecommitWait(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	defer func() {
		// Done enterPrecommitWait:
		cs.updateRoundStep(round, ttypes.RoundStepPrecommitWait)
		cs.newStep()
	}()

	// enable gossip votes in case other validators enterAggPrecommitWait
	if useAggSig && !cs.isProposer() && cs.Votes.Precommits(round).GetAggVote() == nil {
		gossipVotes.Store(true)
	}

	// Wait for some more precommits; enterNewRound
	cs.scheduleTimeout(cs.Precommit(round), height, round, ttypes.RoundStepPrecommitWait)
}

// Enter: +2/3 precommits for block
func (cs *ConsensusState) enterCommit(height int64, commitRound int) {
	if cs.Height != height || ttypes.RoundStepCommit <= cs.Step {
		tendermintlog.Debug(fmt.Sprintf("enterCommit(%v/%v): Invalid args. Current step: %v/%v/%v", height, commitRound, cs.Height, cs.Round, cs.Step))
		return
	}
	tendermintlog.Info(fmt.Sprintf("enterCommit(%v/%v). Current: %v/%v/%v", height, commitRound, cs.Height, cs.Round, cs.Step), "cost", types.Since(cs.begCons))

	defer func() {
		// Done enterCommit:
		// keep cs.Round the same, commitRound points to the right Precommits set.
		cs.updateRoundStep(cs.Round, ttypes.RoundStepCommit)
		cs.CommitRound = commitRound
		cs.CommitTime = time.Now()
		cs.newStep()

		// Maybe finalize immediately.
		cs.tryFinalizeCommit(height)
	}()

	blockID, ok := cs.Votes.Precommits(commitRound).TwoThirdsMajority()
	if !ok {
		panic("enterCommit expects +2/3 precommits")
	}

	// The Locked* fields no longer matter.
	// Move them over to ProposalBlock if they match the commit hash,
	// otherwise they'll be cleared in updateToState.
	if cs.LockedBlock.HashesTo(blockID.Hash) {
		tendermintlog.Info("Commit is for locked block. Set ProposalBlock=LockedBlock", "LockedBlock-hash", fmt.Sprintf("%X", blockID.Hash))
		cs.ProposalBlock = cs.LockedBlock
		cs.ProposalBlockHash = blockID.Hash
	}

	// If we don't have the block being committed, set up to get it.
	if !cs.ProposalBlock.HashesTo(blockID.Hash) {
		tendermintlog.Info("Commit is for a block we don't know about. Set ProposalBlock=nil",
			"proposal", fmt.Sprintf("%X", cs.ProposalBlock.Hash()),
			"ProposalBlockHash", fmt.Sprintf("%X", cs.ProposalBlockHash),
			"commit", fmt.Sprintf("%X", blockID.Hash))
		// We're getting the wrong block.
		// Set up ProposalBlockHash and keep waiting.
		cs.ProposalBlock = nil
		cs.ProposalBlockHash = blockID.Hash

		validBlockMsg := &tmtypes.ValidBlockMsg{
			Height:    cs.Height,
			Round:     int32(cs.Round),
			Blockhash: cs.ProposalBlockHash,
			IsCommit:  true,
		}
		cs.broadcastChannel <- MsgInfo{TypeID: ttypes.ValidBlockID, Msg: validBlockMsg, PeerID: "", PeerIP: ""}
		//else {
		// We just need to keep waiting.
		//}
	}
}

// If we have the block AND +2/3 commits for it, finalize.
func (cs *ConsensusState) tryFinalizeCommit(height int64) {
	if cs.Height != height {
		panic(fmt.Sprintf("tryFinalizeCommit() cs.Height: %v vs height: %v", cs.Height, height))
	}

	blockID, ok := cs.Votes.Precommits(cs.CommitRound).TwoThirdsMajority()
	if !ok || len(blockID.Hash) == 0 {
		tendermintlog.Error("Attempt to finalize failed. There was no +2/3 majority, or +2/3 was for <nil>.")
		tendermintlog.Info(fmt.Sprintf("Continue consensus. Current: %v/%v/%v", cs.Height, cs.CommitRound, cs.Step), "cost", types.Since(cs.begCons))
		return
	}
	if !cs.ProposalBlock.HashesTo(blockID.Hash) {
		// TODO: this happens every time if we're not a validator (ugly logs)
		// TODO: ^^ wait, why does it matter that we're a validator?
		tendermintlog.Error("Attempt to finalize failed. We don't have the commit block.", "ProposalBlock-hash", fmt.Sprintf("%X", cs.ProposalBlock.Hash()),
			"CommitBlock-hash", fmt.Sprintf("%X", blockID.Hash))
		tendermintlog.Info(fmt.Sprintf("Continue consensus. Current: %v/%v/%v", cs.Height, cs.CommitRound, cs.Step), "cost", types.Since(cs.begCons))
		return
	}

	// go
	cs.finalizeCommit(height)
}

// Increment height and goto ttypes.RoundStepNewHeight
func (cs *ConsensusState) finalizeCommit(height int64) {
	if cs.Height != height || cs.Step != ttypes.RoundStepCommit {
		tendermintlog.Debug(fmt.Sprintf("finalizeCommit(%v): Invalid args. Current step: %v/%v/%v", height, cs.Height, cs.Round, cs.Step))
		return
	}

	blockID, ok := cs.Votes.Precommits(cs.CommitRound).TwoThirdsMajority()
	block := cs.ProposalBlock

	if !ok {
		panic(fmt.Sprintf("Cannot finalizeCommit, commit does not have two thirds majority"))
	}

	if !block.HashesTo(blockID.Hash) {
		panic(fmt.Sprintf("Cannot finalizeCommit, ProposalBlock does not hash to commit hash"))
	}
	if err := cs.blockExec.ValidateBlock(cs.state, block); err != nil {
		panic(fmt.Sprintf("+2/3 committed an invalid block: %v", err))
	}

	stateCopy := cs.state.Copy()
	tendermintlog.Debug("finalizeCommit validators of statecopy", "validators", stateCopy.Validators)
	// NOTE: the block.AppHash wont reflect these txs until the next block
	var err error
	stateCopy, err = cs.blockExec.ApplyBlock(stateCopy, ttypes.BlockID{BlockID: tmtypes.BlockID{Hash: block.Hash()}}, block)
	if err != nil {
		panic(fmt.Sprintf("finalizeCommit ApplyBlock fail: %v", err))
	}

	// commit block
	commitBlock := cs.ProposalBlock.Data
	err = cs.client.CommitBlock(commitBlock)
	if err != nil {
		panic(fmt.Sprintf("finalizeCommit CommitBlock fail: %v", err))
	}
	if bytes.Equal(cs.privValidator.GetAddress(), block.TendermintBlock.Header.ProposerAddr) {
		tendermintlog.Info(fmt.Sprintf("Proposer reach consensus. Current: %v/%v/%v", cs.Height, cs.Round, cs.Step), "CommitRound", cs.CommitRound,
			"tx-len", len(commitBlock.Txs), "cost", types.Since(cs.begCons), "proposer-addr", fmt.Sprintf("%X", ttypes.Fingerprint(block.TendermintBlock.Header.ProposerAddr)))
	} else {
		tendermintlog.Info(fmt.Sprintf("Not-Proposer reach consensus. Current: %v/%v/%v", cs.Height, cs.Round, cs.Step), "CommitRound", cs.CommitRound,
			"tx-len", len(commitBlock.Txs), "cost", types.Since(cs.begCons), "proposer-addr", fmt.Sprintf("%X", ttypes.Fingerprint(block.TendermintBlock.Header.ProposerAddr)))
	}
	reqblock, err := cs.client.RequestBlock(height)
	if err == nil {
		stateCopy.LastResultsHash = reqblock.Hash(cs.client.GetAPI().GetConfig())
	}

	//check whether need update validator nodes
	valNodes, err := cs.client.QueryValidatorsByHeight(block.Header.Height)
	if err == nil && valNodes != nil {
		if len(valNodes.Nodes) > 0 {
			tendermintlog.Info("finalizeCommit validators of statecopy update", "update-valnodes", valNodes)
			prevValSet := stateCopy.LastValidators.Copy()
			nextValSet := prevValSet.Copy()
			err := updateValidators(nextValSet, valNodes.Nodes)
			if err != nil {
				tendermintlog.Error("Error changing validator set", "error", err)
			}
			// change results from this height but only applies to the next height
			stateCopy.LastHeightValidatorsChanged = block.Header.Height + 1
			nextValSet.IncrementAccum(1)
			stateCopy.Validators = nextValSet
		}
	}
	tendermintlog.Debug("finalizeCommit validators of statecopy", "validators", stateCopy.Validators.String())

	// save local state and seen commit
	precommits := cs.Votes.Precommits(cs.CommitRound)
	seenCommit := precommits.MakeCommit()
	newState := SaveState(stateCopy)
	err = cs.client.csStore.SaveConsensusState(height, newState, seenCommit)
	if err != nil {
		panic(fmt.Sprintf("finalizeCommit SaveSeenCommit fail: %v", err))
	}
	tendermintlog.Info(fmt.Sprintf("Save consensus state. Current: %v/%v/%v", cs.Height, cs.CommitRound, cs.Step), "cost", types.Since(cs.begCons))

	// NewHeightStep!
	cs.updateToState(stateCopy)

	// cs.StartTime is already set.
	// Schedule Round0 to start soon.
	cs.scheduleRound0(&cs.RoundState)

	// By here,
	// * cs.Height has been increment to height+1
	// * cs.Step is now ttypes.RoundStepNewHeight
	// * cs.StartTime is set to when we will start round0.
	// Execute and commit the block, update and save the state, and update the mempool.

}

//-----------------------------------------------------------------------------

func (cs *ConsensusState) defaultSetProposal(proposal *tmtypes.Proposal) error {
	tendermintlog.Debug(fmt.Sprintf("Consensus receive proposal. Current: %v/%v/%v", cs.Height, cs.Round, cs.Step),
		"proposal", fmt.Sprintf("%v/%v", proposal.Height, proposal.Round))
	// Already have one
	// TODO: possibly catch double proposals
	if cs.Proposal != nil {
		tendermintlog.Debug("defaultSetProposal: already has proposal")
		return nil
	}

	// Does not apply
	if proposal.Height != cs.Height || int(proposal.Round) != cs.Round {
		tendermintlog.Debug("defaultSetProposal: height is not equal or round is not equal", "proposal-height", proposal.Height, "cs-height", cs.Height,
			"proposal-round", proposal.Round, "cs-round", cs.Round)
		return nil
	}

	if cs.begCons.IsZero() {
		cs.begCons = time.Now()
	}

	// Verify POLRound, which must be -1 or in range [0, proposal.Round).
	if proposal.POLRound < -1 ||
		(proposal.POLRound >= 0 && proposal.POLRound >= proposal.Round) {
		tendermintlog.Error("Invalid POLRound", "POLRound", proposal.POLRound, "Round", proposal.Round)
		return ErrInvalidProposalPOLRound
	}

	// Verify signature
	pubkey, err := ttypes.ConsensusCrypto.PubKeyFromBytes(cs.Validators.GetProposer().PubKey)
	if err != nil {
		return fmt.Errorf("Error pubkey from bytes:%v", err)
	}
	proposalTmp := &ttypes.Proposal{Proposal: *proposal}
	signature, err := ttypes.ConsensusCrypto.SignatureFromBytes(proposal.Signature)
	if err != nil {
		return fmt.Errorf("defaultSetProposal Error: SIGA[%v] to signature failed:%v", proposal.Signature, err)
	}
	if !pubkey.VerifyBytes(ttypes.SignBytes(cs.state.ChainID, proposalTmp), signature) {
		return ErrInvalidProposalSignature
	}
	cs.Proposal = proposal
	cs.ProposalBlockHash = proposal.Blockhash
	tendermintlog.Info(fmt.Sprintf("Consensus set proposal. Current: %v/%v/%v", cs.Height, cs.Round, cs.Step),
		"blockhash", fmt.Sprintf("%X", cs.Proposal.Blockhash), "cost", types.Since(cs.begCons))
	return nil
}

// NOTE: block is not necessarily valid.
// Asynchronously triggers either enterPrevote (before we timeout of propose) or tryFinalizeCommit, once we have the full block.
func (cs *ConsensusState) addProposalBlock(proposalBlock *tmtypes.TendermintBlock) (err error) {
	block := &ttypes.TendermintBlock{TendermintBlock: proposalBlock}
	height, round := block.Header.Height, block.Header.Round
	tendermintlog.Debug(fmt.Sprintf("Consensus receive proposal block. Current: %v/%v/%v", cs.Height, cs.Round, cs.Step),
		"block(H/R/hash)", fmt.Sprintf("%v/%v/%X", height, round, block.Hash()))

	// Blocks might be reused, so round mismatch is OK
	if cs.Height != height {
		tendermintlog.Debug("Received block from wrong height", "height", height, "round", round)
		return nil
	}

	if cs.begCons.IsZero() {
		cs.begCons = time.Now()
	}

	// We're not expecting a block
	if !block.HashesTo(cs.ProposalBlockHash) {
		// NOTE: this can happen when we've gone to a higher round and
		// then receive parts from the previous round - not necessarily a bad peer.
		tendermintlog.Debug("Received block when we're not expecting any", "ProposalBlockHash", fmt.Sprintf("%X", cs.ProposalBlockHash),
			"height", height, "round", round, "hash", fmt.Sprintf("%X", block.Hash()))
		return nil
	}

	// Already have expected proposal block
	if block.HashesTo(cs.ProposalBlock.Hash()) {
		tendermintlog.Debug("addProposalBlock: already has proposal block")
		return nil
	}

	cs.ProposalBlock = block

	// NOTE: it's possible to receive proposal block for future rounds without having the proposal
	tendermintlog.Info(fmt.Sprintf("Consensus set proposal block. Current: %v/%v/%v", cs.Height, cs.Round, cs.Step),
		"ProposalBlockHash", fmt.Sprintf("%X", cs.ProposalBlockHash), "cost", types.Since(cs.begCons))

	// Update Valid* if we can.
	prevotes := cs.Votes.Prevotes(cs.Round)
	blockID, hasTwoThirds := prevotes.TwoThirdsMajority()
	if hasTwoThirds && len(blockID.Hash) == 0 && (cs.ValidRound < cs.Round) {
		if cs.ProposalBlock.HashesTo(blockID.Hash) {
			tendermintlog.Info("Updating valid block to new proposal block",
				"valid-round", cs.Round, "valid-block-hash", cs.ProposalBlock.Hash())
			cs.ValidRound = cs.Round
			cs.ValidBlock = cs.ProposalBlock
		}
		// TODO: In case there is +2/3 majority in Prevotes set for some
		// block and cs.ProposalBlock contains different block, either
		// proposer is faulty or voting power of faulty processes is more
		// than 1/3. We should trigger in the future accountability
		// procedure at this point.
	}

	if cs.Step <= ttypes.RoundStepPropose && cs.isProposalComplete() {
		// Move onto the next step
		cs.enterPrevote(cs.Height, cs.Round)
		if hasTwoThirds { // this is optimisation as this will be triggered when prevote is added
			cs.enterPrecommit(cs.Height, cs.Round)
		}
	} else if cs.Step == ttypes.RoundStepCommit {
		// If we're waiting on the proposal block...
		cs.tryFinalizeCommit(cs.Height)
	}
	return nil
}

// Attempt to add aggregate vote.
func (cs *ConsensusState) tryAddAggVote(aggVoteRaw *tmtypes.AggVote, peerID string) error {
	aggVote := &ttypes.AggVote{AggVote: aggVoteRaw}
	if aggVote.Height != cs.Height {
		tendermintlog.Debug("AggVote ignored and not added", "voteHeight", aggVote.Height, "csHeight", cs.Height, "peerID", peerID)
		return ErrVoteHeightMismatch
	}
	if cs.begCons.IsZero() {
		cs.begCons = time.Now()
	}
	height := cs.Height
	_, err := cs.Votes.AddAggVote(aggVote, peerID)
	if err != nil {
		tendermintlog.Error("Error attempting to add aggregate vote", "err", err)
		return ErrAddingAggVote
	}

	switch aggVote.Type {
	case uint32(ttypes.VoteTypePrevote):
		prevotes := cs.Votes.Prevotes(int(aggVote.Round))
		tendermintlog.Info("Added aggregate prevote", "aggVote", aggVote, "prevotes", prevotes.StringShort())

		// If +2/3 prevotes for a block or nil for *any* round:
		if blockID, ok := prevotes.TwoThirdsMajority(); ok {

			// There was a polka!
			// If we're locked but this is a recent polka, unlock.
			// If it matches our ProposalBlock, update the ValidBlock

			// Unlock if `cs.LockedRound < vote.Round <= cs.Round`
			// NOTE: If vote.Round > cs.Round, we'll deal with it when we get to vote.Round
			if (cs.LockedBlock != nil) &&
				(cs.LockedRound < int(aggVote.Round)) &&
				(int(aggVote.Round) <= cs.Round) &&
				!cs.LockedBlock.HashesTo(blockID.Hash) {

				tendermintlog.Info("Unlocking because of POL.", "lockedRound", cs.LockedRound, "POLRound", aggVote.Round)
				cs.LockedRound = -1
				cs.LockedBlock = nil
			}

			// Update Valid* if we can.
			// NOTE: our proposal block may be nil or not what received a polka..
			if len(blockID.Hash) != 0 && (cs.ValidRound < int(aggVote.Round)) && (int(aggVote.Round) == cs.Round) {

				if cs.ProposalBlock.HashesTo(blockID.Hash) {
					tendermintlog.Info("Updating ValidBlock because of POL.", "validRound", cs.ValidRound, "POLRound", aggVote.Round)
					cs.ValidRound = int(aggVote.Round)
					cs.ValidBlock = cs.ProposalBlock
				} else {
					tendermintlog.Info(
						"Valid block we don't know about. Set ProposalBlock=nil",
						"proposal", fmt.Sprintf("%X", cs.ProposalBlock.Hash()), "blockId", fmt.Sprintf("%X", blockID.Hash))
					// We're getting the wrong block.
					cs.ProposalBlock = nil
				}
				cs.ProposalBlockHash = blockID.Hash

				validBlockMsg := &tmtypes.ValidBlockMsg{
					Height:    aggVote.Height,
					Round:     aggVote.Round,
					Blockhash: cs.ProposalBlockHash,
					IsCommit:  false,
				}
				cs.broadcastChannel <- MsgInfo{TypeID: ttypes.ValidBlockID, Msg: validBlockMsg, PeerID: "", PeerIP: ""}
			}
		}

		// If +2/3 prevotes for *anything* for this or future round:
		switch {
		case cs.Round < int(aggVote.Round) && prevotes.HasTwoThirdsAny():
			// Round-skip if there is any 2/3+ of votes ahead of us
			cs.enterNewRound(height, int(aggVote.Round))
		case cs.Round == int(aggVote.Round) && ttypes.RoundStepPrevote <= cs.Step: // current round
			blockID, ok := prevotes.TwoThirdsMajority()
			if ok && (cs.isProposalComplete() || len(blockID.Hash) == 0) {
				cs.enterPrecommit(height, int(aggVote.Round))
			} else if prevotes.HasTwoThirdsAny() {
				cs.enterPrevoteWait(height, int(aggVote.Round))
			}
		case cs.Proposal != nil && 0 <= cs.Proposal.POLRound && cs.Proposal.POLRound == aggVote.Round:
			// If the proposal is now complete, enter prevote of cs.Round.
			if cs.isProposalComplete() {
				cs.enterPrevote(height, cs.Round)
			}
		}

	case uint32(ttypes.VoteTypePrecommit):
		precommits := cs.Votes.Precommits(int(aggVote.Round))
		tendermintlog.Info("Added aggregate precommit", "aggVote", aggVote, "precommits", precommits.StringShort())

		blockID, ok := precommits.TwoThirdsMajority()
		if ok {
			// Executed as TwoThirdsMajority could be from a higher round
			cs.enterNewRound(height, int(aggVote.Round))
			cs.enterPrecommit(height, int(aggVote.Round))
			if len(blockID.Hash) != 0 {
				cs.enterCommit(height, int(aggVote.Round))
				if skipTimeoutCommit && precommits.HasAll() {
					cs.enterNewRound(cs.Height, 0)
				}
			} else {
				cs.enterPrecommitWait(height, int(aggVote.Round))
			}
		} else if cs.Round <= int(aggVote.Round) && precommits.HasTwoThirdsAny() {
			cs.enterNewRound(height, int(aggVote.Round))
			cs.enterPrecommitWait(height, int(aggVote.Round))
		}

	default:
		panic(fmt.Sprintf("Unexpected aggVote type %X", aggVote.Type)) // Should not happen.
	}
	return nil
}

// Attempt to add the vote. if its a duplicate signature, dupeout the validator
func (cs *ConsensusState) tryAddVote(voteRaw *tmtypes.Vote, peerID string, peerIP string) error {
	vote := &ttypes.Vote{Vote: voteRaw}
	_, err := cs.addVote(vote, peerID, peerIP)
	if err != nil {
		// If the vote height is off, we'll just ignore it,
		// But if it's a conflicting sig, add it to the cs.evpool.
		// If it's otherwise invalid, punish peer.
		if err == ErrVoteHeightMismatch {
			return err
		} else if err == ttypes.ErrVoteConflict {
			if bytes.Equal(vote.ValidatorAddress, cs.privValidator.GetAddress()) {
				tendermintlog.Error("Found conflicting vote from ourselves. Did you unsafe_reset a validator?", "height", vote.Height, "round", vote.Round, "type", vote.Type)
				return err
			}
		} else {
			// Probably an invalid signature / Bad peer.
			// Seems this can also err sometimes with "Unexpected step" - perhaps not from a bad peer ?
			tendermintlog.Error("Error attempting to add vote", "err", err)
			return ErrAddingVote
		}
	}
	return nil
}

//-----------------------------------------------------------------------------

func (cs *ConsensusState) addVote(vote *ttypes.Vote, peerID string, peerIP string) (added bool, err error) {
	tendermintlog.Debug(fmt.Sprintf("Consensus receive vote. Current: %v/%v/%v", cs.Height, cs.Round, cs.Step),
		"vote", fmt.Sprintf("{%v:%X %v/%02d/%v}", vote.ValidatorIndex, ttypes.Fingerprint(vote.ValidatorAddress), vote.Height, vote.Round, vote.Type), "peerip", peerIP)

	// A precommit for the previous height
	// These come in while we wait timeoutCommit
	if vote.Height+1 == cs.Height {
		if !(cs.Step == ttypes.RoundStepNewHeight && vote.Type == uint32(ttypes.VoteTypePrecommit)) {
			// TODO: give the reason ..
			// fmt.Errorf("tryAddVote: Wrong height, not a LastCommit straggler commit.")
			return added, ErrVoteHeightMismatch
		}
		added, err = cs.LastCommit.AddVote(vote)
		if !added {
			return added, err
		}
		tendermintlog.Info(fmt.Sprintf("Added to lastPrecommits: %v", cs.LastCommit.StringShort()))

		hasVoteMsg := &tmtypes.HasVoteMsg{
			Height: vote.Height,
			Round:  vote.Round,
			Type:   int32(vote.Type),
			Index:  vote.ValidatorIndex,
		}
		cs.broadcastChannel <- MsgInfo{TypeID: ttypes.HasVoteID, Msg: hasVoteMsg, PeerID: "", PeerIP: ""}

		// if we can skip timeoutCommit and have all the votes now,
		if skipTimeoutCommit && cs.LastCommit.HasAll() {
			// go straight to new round (skip timeout commit)
			// cs.scheduleTimeout(time.Duration(0), cs.Height, 0, cstypes.RoundStepNewHeight)
			cs.enterNewRound(cs.Height, 0)
		}
		return
	}

	// Height mismatch is ignored.
	// Not necessarily a bad peer, but not favourable behaviour.
	if vote.Height != cs.Height {
		err = ErrVoteHeightMismatch
		tendermintlog.Debug("Vote ignored and not added", "voteHeight", vote.Height, "csHeight", cs.Height, "peerID", peerID)
		return
	}

	if cs.begCons.IsZero() {
		cs.begCons = time.Now()
	}

	height := cs.Height
	added, err = cs.Votes.AddVote(vote, peerID)
	if !added {
		// Either duplicate, or error upon cs.Votes.AddByIndex()
		return
	}
	//cs.broadcastChannel <- MsgInfo{TypeID: ttypes.VoteID, Msg: vote.Vote, PeerID: cs.ourID, PeerIP: ""}
	hasVoteMsg := &tmtypes.HasVoteMsg{
		Height: vote.Height,
		Round:  vote.Round,
		Type:   int32(vote.Type),
		Index:  vote.ValidatorIndex,
	}
	cs.broadcastChannel <- MsgInfo{TypeID: ttypes.HasVoteID, Msg: hasVoteMsg, PeerID: "", PeerIP: ""}

	switch vote.Type {
	case uint32(ttypes.VoteTypePrevote):
		prevotes := cs.Votes.Prevotes(int(vote.Round))
		tendermintlog.Info("Added to prevote", "vote", vote, "prevotes", prevotes.StringShort())

		// If +2/3 prevotes for a block or nil for *any* round:
		if blockID, ok := prevotes.TwoThirdsMajority(); ok {

			// There was a polka!
			// If we're locked but this is a recent polka, unlock.
			// If it matches our ProposalBlock, update the ValidBlock

			// Unlock if `cs.LockedRound < vote.Round <= cs.Round`
			// NOTE: If vote.Round > cs.Round, we'll deal with it when we get to vote.Round
			if (cs.LockedBlock != nil) &&
				(cs.LockedRound < int(vote.Round)) &&
				(int(vote.Round) <= cs.Round) &&
				!cs.LockedBlock.HashesTo(blockID.Hash) {

				tendermintlog.Info("Unlocking because of POL.", "lockedRound", cs.LockedRound, "POLRound", vote.Round)
				cs.LockedRound = -1
				cs.LockedBlock = nil
			}

			// Update Valid* if we can.
			// NOTE: our proposal block may be nil or not what received a polka..
			if len(blockID.Hash) != 0 && (cs.ValidRound < int(vote.Round)) && (int(vote.Round) == cs.Round) {

				if cs.ProposalBlock.HashesTo(blockID.Hash) {
					tendermintlog.Info("Updating ValidBlock because of POL.", "validRound", cs.ValidRound, "POLRound", vote.Round)
					cs.ValidRound = int(vote.Round)
					cs.ValidBlock = cs.ProposalBlock
				} else {
					tendermintlog.Info(
						"Valid block we don't know about. Set ProposalBlock=nil",
						"proposal", fmt.Sprintf("%X", cs.ProposalBlock.Hash()), "blockId", fmt.Sprintf("%X", blockID.Hash))
					// We're getting the wrong block.
					cs.ProposalBlock = nil
				}
				cs.ProposalBlockHash = blockID.Hash

				validBlockMsg := &tmtypes.ValidBlockMsg{
					Height:    vote.Height,
					Round:     vote.Round,
					Blockhash: cs.ProposalBlockHash,
					IsCommit:  false,
				}
				cs.broadcastChannel <- MsgInfo{TypeID: ttypes.ValidBlockID, Msg: validBlockMsg, PeerID: "", PeerIP: ""}
			}
		}

		// If +2/3 prevotes for *anything* for this or future round:
		switch {
		case cs.Round < int(vote.Round) && prevotes.HasTwoThirdsAny():
			// Round-skip if there is any 2/3+ of votes ahead of us
			cs.enterNewRound(height, int(vote.Round))
		case cs.Round == int(vote.Round) && ttypes.RoundStepPrevote <= cs.Step: // current round
			blockID, ok := prevotes.TwoThirdsMajority()
			if ok && (cs.isProposalComplete() || len(blockID.Hash) == 0) {
				if useAggSig && cs.isProposer() && prevotes.GetAggVote() == nil {
					err := prevotes.SetAggVote()
					if err != nil {
						tendermintlog.Error("prevotes SetAggVote fail", "err", err)
						break
					}
					aggVoteMsg := prevotes.GetAggVote().AggVote
					cs.broadcastChannel <- MsgInfo{TypeID: ttypes.AggVoteID, Msg: aggVoteMsg, PeerID: "", PeerIP: ""}
					tendermintlog.Info("Send aggregate prevote", "aggVote", prevotes.GetAggVote())
				}
				cs.enterPrecommit(height, int(vote.Round))
			} else if prevotes.HasTwoThirdsAny() {
				cs.enterPrevoteWait(height, int(vote.Round))
			}
		case cs.Proposal != nil && 0 <= cs.Proposal.POLRound && cs.Proposal.POLRound == vote.Round:
			// If the proposal is now complete, enter prevote of cs.Round.
			if cs.isProposalComplete() {
				cs.enterPrevote(height, cs.Round)
			}
		}

	case uint32(ttypes.VoteTypePrecommit):
		precommits := cs.Votes.Precommits(int(vote.Round))
		tendermintlog.Info("Added to precommit", "vote", vote, "precommits", precommits.StringShort())

		blockID, ok := precommits.TwoThirdsMajority()
		if ok {
			// Executed as TwoThirdsMajority could be from a higher round
			cs.enterNewRound(height, int(vote.Round))
			cs.enterPrecommit(height, int(vote.Round))
			if len(blockID.Hash) != 0 {
				if useAggSig && cs.isProposer() && precommits.GetAggVote() == nil {
					err := precommits.SetAggVote()
					if err != nil {
						tendermintlog.Error("precommits SetAggVote fail", "err", err)
						break
					}
					aggVoteMsg := precommits.GetAggVote().AggVote
					cs.broadcastChannel <- MsgInfo{TypeID: ttypes.AggVoteID, Msg: aggVoteMsg, PeerID: "", PeerIP: ""}
					tendermintlog.Info("Send aggregate precommit", "aggVote", precommits.GetAggVote())
				}
				cs.enterCommit(height, int(vote.Round))
				if skipTimeoutCommit && precommits.HasAll() {
					cs.enterNewRound(cs.Height, 0)
				}
			} else {
				cs.enterPrecommitWait(height, int(vote.Round))
			}
		} else if cs.Round <= int(vote.Round) && precommits.HasTwoThirdsAny() {
			cs.enterNewRound(height, int(vote.Round))
			cs.enterPrecommitWait(height, int(vote.Round))
		}

	default:
		panic(fmt.Sprintf("Unexpected vote type %X", vote.Type)) // Should not happen.
	}

	return added, err
}

func (cs *ConsensusState) signVote(voteType byte, hash []byte) (*ttypes.Vote, error) {
	addr := cs.privValidator.GetAddress()
	valIndex, _ := cs.Validators.GetByAddress(addr)
	tVote := &tmtypes.Vote{}
	tVote.BlockID = &tmtypes.BlockID{Hash: hash}
	vote := &ttypes.Vote{
		Vote: &tmtypes.Vote{
			ValidatorAddress: addr,
			ValidatorIndex:   int32(valIndex),
			Height:           cs.Height,
			Round:            int32(cs.Round),
			Timestamp:        time.Now().UnixNano(),
			Type:             uint32(voteType),
			BlockID:          &tmtypes.BlockID{Hash: hash},
			Signature:        nil,
			UseAggSig:        useAggSig,
		},
	}
	err := cs.privValidator.SignVote(cs.state.ChainID, vote)
	return vote, err
}

// sign the vote and publish on internalMsgQueue
func (cs *ConsensusState) signAddVote(voteType byte, hash []byte) *ttypes.Vote {
	// if we don't have a key or we're not in the validator set, do nothing
	if cs.privValidator == nil || !cs.Validators.HasAddress(cs.privValidator.GetAddress()) {
		return nil
	}
	vote, err := cs.signVote(voteType, hash)
	if err == nil {
		// send to self
		cs.sendInternalMessage(MsgInfo{TypeID: ttypes.VoteID, Msg: vote.Vote, PeerID: cs.ourID, PeerIP: ""})
		if useAggSig {
			// send to proposer
			cs.unicastChannel <- MsgInfo{TypeID: ttypes.VoteID, Msg: vote.Vote, PeerID: cs.getProposerID(), PeerIP: ""}
			// wait for aggregate vote
			if voteType == ttypes.VoteTypePrevote {
				cs.scheduleTimeout(cs.Prevote(cs.Round), cs.Height, cs.Round, ttypes.RoundStepAggPrevoteWait)
			} else if voteType == ttypes.VoteTypePrecommit {
				cs.scheduleTimeout(cs.Precommit(cs.Round), cs.Height, cs.Round, ttypes.RoundStepAggPrecommitWait)
			}
		}
		tendermintlog.Info("Sign and send vote", "height", cs.Height, "round", cs.Round, "vote", vote)
		return vote
	}

	tendermintlog.Error("Error signing vote", "height", cs.Height, "round", cs.Round, "vote", vote, "err", err)
	return nil
}

// CompareHRS method
func CompareHRS(h1 int64, r1 int, s1 ttypes.RoundStepType, h2 int64, r2 int, s2 ttypes.RoundStepType) int {
	if h1 < h2 {
		return -1
	} else if h1 > h2 {
		return 1
	}
	if r1 < r2 {
		return -1
	} else if r1 > r2 {
		return 1
	}
	if s1 < s2 {
		return -1
	} else if s1 > s2 {
		return 1
	}
	return 0
}

// Commit returns the amount of time to wait for straggler votes after receiving +2/3 precommits for a single block (ie. a commit).
func (cs *ConsensusState) Commit(t time.Time) time.Time {
	return t.Add(time.Duration(timeoutCommit) * time.Millisecond)
}

// Propose returns the amount of time to wait for a proposal
func (cs *ConsensusState) Propose(round int) time.Duration {
	return time.Duration(timeoutPropose+timeoutProposeDelta*int32(round)) * time.Millisecond
}

// Prevote returns the amount of time to wait for straggler votes after receiving any +2/3 prevotes
func (cs *ConsensusState) Prevote(round int) time.Duration {
	return time.Duration(timeoutPrevote+timeoutPrevoteDelta*int32(round)) * time.Millisecond
}

// Precommit returns the amount of time to wait for straggler votes after receiving any +2/3 precommits
func (cs *ConsensusState) Precommit(round int) time.Duration {
	return time.Duration(timeoutPrecommit+timeoutPrecommitDelta*int32(round)) * time.Millisecond
}

// WaitForTxs returns true if the consensus should wait for transactions before entering the propose step
func (cs *ConsensusState) WaitForTxs() bool {
	return !createEmptyBlocks || createEmptyBlocksInterval > 0
}

// EmptyBlocksInterval returns the amount of time to wait before proposing an empty block or starting the propose timer if there are no txs available
func (cs *ConsensusState) EmptyBlocksInterval() time.Duration {
	return time.Duration(createEmptyBlocksInterval) * time.Second
}

// PeerGossipSleep returns the amount of time to sleep if there is nothing to send from the ConsensusReactor
func (cs *ConsensusState) PeerGossipSleep() time.Duration {
	return time.Duration(peerGossipSleepDuration) * time.Millisecond
}

// PeerQueryMaj23Sleep returns the amount of time to sleep after each VoteSetMaj23Message is sent in the ConsensusReactor
func (cs *ConsensusState) PeerQueryMaj23Sleep() time.Duration {
	return time.Duration(peerQueryMaj23SleepDuration) * time.Millisecond
}

// IsProposer method
func (cs *ConsensusState) IsProposer() bool {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	return cs.isProposer()
}

// GetPrevotesState method
func (cs *ConsensusState) GetPrevotesState(height int64, round int, blockID *tmtypes.BlockID) *ttypes.BitArray {
	if height != cs.Height {
		return nil
	}
	return cs.Votes.Prevotes(round).BitArrayByBlockID(blockID)
}

// GetPrecommitsState method
func (cs *ConsensusState) GetPrecommitsState(height int64, round int, blockID *tmtypes.BlockID) *ttypes.BitArray {
	if height != cs.Height {
		return nil
	}
	return cs.Votes.Precommits(round).BitArrayByBlockID(blockID)
}

// SetPeerMaj23 when reach maj 2/3
func (cs *ConsensusState) SetPeerMaj23(height int64, round int, voteType byte, peerID ID, blockID *tmtypes.BlockID) {
	if height == cs.Height {
		cs.Votes.SetPeerMaj23(round, voteType, string(peerID), blockID)
	}
}
