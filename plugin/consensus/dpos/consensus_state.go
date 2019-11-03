// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpos

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/types"

	"github.com/33cn/chain33/common/crypto"
	dpostype "github.com/33cn/plugin/plugin/consensus/dpos/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
	"github.com/golang/protobuf/proto"
)

//-----------------------------------------------------------------------------
// Config

const (
	continueToVote = 0
	voteSuccess    = 1
	voteFail       = 2

	//VrfQueryTypeM vrf query type 为查询M信息
	VrfQueryTypeM = 0

	//VrfQueryTypeRP vrf query type 为查询RP信息
	VrfQueryTypeRP = 1
)

// Errors define
var (
	ErrInvalidVoteSignature      = errors.New("Error invalid vote signature")
	ErrInvalidVoteReplySignature = errors.New("Error invalid vote reply signature")
	ErrInvalidNotifySignature    = errors.New("Error invalid notify signature")
)

//-----------------------------------------------------------------------------

var (
	msgQueueSize = 1000
)

// ConsensusState handles execution of the consensus algorithm.
type ConsensusState struct {
	// config details
	client             *Client
	privValidator      ttypes.PrivValidator // for signing votes
	privValidatorIndex int

	// internal state
	mtx          sync.Mutex
	validatorMgr ValidatorMgr // State until height-1.

	// state changes may be triggered by msgs from peers,
	// msgs from ourself, or by timeouts
	peerMsgQueue     chan MsgInfo
	internalMsgQueue chan MsgInfo
	timer            *time.Timer

	broadcastChannel chan<- MsgInfo
	ourID            ID
	started          uint32 // atomic
	stopped          uint32 // atomic
	Quit             chan struct{}

	//当前状态
	dposState State

	//所有选票，包括自己的和从网络中接收到的
	dposVotes []*dpostype.DPosVote

	//当前达成共识的选票
	currentVote *dpostype.VoteItem
	lastVote    *dpostype.VoteItem

	myVote     *dpostype.DPosVote
	lastMyVote *dpostype.DPosVote

	notify     *dpostype.DPosNotify
	lastNotify *dpostype.DPosNotify

	//所有选票，包括自己的和从网络中接收到的
	cachedVotes []*dpostype.DPosVote

	cachedNotify *dpostype.DPosNotify

	cycleBoundaryMap map[int64]*dty.DposCBInfo
	vrfInfoMap       map[int64]*dty.VrfInfo
	vrfInfosMap      map[int64][]*dty.VrfInfo

	cachedTopNCands []*dty.TopNCandidators
}

// NewConsensusState returns a new ConsensusState.
func NewConsensusState(client *Client, valMgr ValidatorMgr) *ConsensusState {
	cs := &ConsensusState{
		client:           client,
		peerMsgQueue:     make(chan MsgInfo, msgQueueSize),
		internalMsgQueue: make(chan MsgInfo, msgQueueSize),

		Quit:             make(chan struct{}),
		dposState:        InitStateObj,
		dposVotes:        nil,
		cycleBoundaryMap: make(map[int64]*dty.DposCBInfo),
		vrfInfoMap:       make(map[int64]*dty.VrfInfo),
		vrfInfosMap:      make(map[int64][]*dty.VrfInfo),
	}

	cs.updateToValMgr(valMgr)

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

// IsRunning method
func (cs *ConsensusState) IsRunning() bool {
	return atomic.LoadUint32(&cs.started) == 1 && atomic.LoadUint32(&cs.stopped) == 0
}

//----------------------------------------
// String returns a string.
func (cs *ConsensusState) String() string {
	// better not to access shared variables
	return fmt.Sprintf("ConsensusState") //(H:%v R:%v S:%v", cs.Height, cs.Round, cs.Step)
}

// GetValidatorMgr returns a copy of the ValidatorMgr.
func (cs *ConsensusState) GetValidatorMgr() ValidatorMgr {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	return cs.validatorMgr.Copy()
}

// GetPrivValidator returns the pointer of PrivValidator
func (cs *ConsensusState) GetPrivValidator() ttypes.PrivValidator {
	return cs.privValidator
}

// GetValidators returns a copy of the current validators.
func (cs *ConsensusState) GetValidators() []*ttypes.Validator {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	return cs.validatorMgr.Validators.Copy().Validators
}

// SetPrivValidator sets the private validator account for signing votes.
func (cs *ConsensusState) SetPrivValidator(priv ttypes.PrivValidator, index int) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	cs.privValidator = priv
	cs.privValidatorIndex = index
}

// Start It start first time starts the timeout receive routines.
func (cs *ConsensusState) Start() {
	if atomic.CompareAndSwapUint32(&cs.started, 0, 1) {
		if atomic.LoadUint32(&cs.stopped) == 1 {
			dposlog.Error("ConsensusState already stoped")
		}

		// now start the receiveRoutine
		go cs.receiveRoutine()
	}
}

// Stop timer and receive routine
func (cs *ConsensusState) Stop() {
	cs.Quit <- struct{}{}
}

// Attempt to reset the timer
func (cs *ConsensusState) resetTimer(duration time.Duration, stateType int) {
	dposlog.Info("set timer", "duration", duration, "state", StateTypeMapping[stateType])
	if !cs.timer.Stop() {
		select {
		case <-cs.timer.C:
		default:
		}
	}
	cs.timer.Reset(duration)
}

// Updates ConsensusState and increments height to match that of state.
// The round becomes 0 and cs.Step becomes ttypes.RoundStepNewHeight.
func (cs *ConsensusState) updateToValMgr(valMgr ValidatorMgr) {
	cs.validatorMgr = valMgr
}

//-----------------------------------------
// the main go routines

// receiveRoutine handles messages which may cause state transitions.
// it's argument (n) is the number of messages to process before exiting - use 0 to run forever
// It keeps the RoundState and is the only thing that updates it.
// Updates (state transitions) happen on timeouts, complete proposals, and 2/3 majorities.
// ConsensusState must be locked before any internal state is updated.
func (cs *ConsensusState) receiveRoutine() {
	defer func() {
		if r := recover(); r != nil {
			dposlog.Error("CONSENSUS FAILURE!!!", "err", r, "stack", string(debug.Stack()))
		}
	}()

	cs.timer = time.NewTimer(time.Second * 3)

	for {
		var mi MsgInfo

		select {
		case mi = <-cs.peerMsgQueue:
			// handles proposals, block parts, votes
			// may generate internal events (votes, complete proposals, 2/3 majorities)
			cs.handleMsg(mi)
		case mi = <-cs.internalMsgQueue:
			// handles proposals, block parts, votes
			cs.handleMsg(mi)
		case <-cs.timer.C:
			cs.handleTimeout()
		case <-cs.Quit:
			dposlog.Info("ConsensusState recv quit signal.")
			cs.timer.Stop()
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
	dposlog.Info("Recv consensus msg", "msg type", fmt.Sprintf("%T", msg), "peerid", peerID, "peerip", peerIP)

	switch msg := msg.(type) {
	case *dpostype.DPosVote:
		cs.dposState.recvVote(cs, msg)
	case *dpostype.DPosNotify:
		cs.dposState.recvNotify(cs, msg)
	case *dpostype.DPosVoteReply:
		cs.dposState.recvVoteReply(cs, msg)
	case *dpostype.DPosCBInfo:
		cs.dposState.recvCBInfo(cs, msg)
	default:
		dposlog.Error("Unknown msg type", "msg", msg.String(), "peerid", peerID, "peerip", peerIP)
	}
	if err != nil {
		dposlog.Error("Error with msg", "type", reflect.TypeOf(msg), "peerid", peerID, "peerip", peerIP, "err", err, "msg", msg)
	}
}

func (cs *ConsensusState) handleTimeout() {
	// the timeout will now cause a state transition
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	//由具体的状态来处理超时消息
	cs.dposState.timeOut(cs)
}

// IsProposer method
func (cs *ConsensusState) IsProposer() bool {
	if cs.currentVote != nil {
		return bytes.Equal(cs.currentVote.VotedNodeAddress, cs.privValidator.GetAddress())
	}

	return false
}

// SetState method
func (cs *ConsensusState) SetState(state State) {
	cs.dposState = state
}

// SaveVote method
func (cs *ConsensusState) SaveVote() {
	if cs.lastVote == nil {
		cs.lastVote = cs.currentVote
	} else if cs.currentVote != nil && !bytes.Equal(cs.currentVote.VoteID, cs.lastVote.VoteID) {
		cs.lastVote = cs.currentVote
	}
}

// SetCurrentVote method
func (cs *ConsensusState) SetCurrentVote(vote *dpostype.VoteItem) {
	cs.currentVote = vote
}

// SaveMyVote method
func (cs *ConsensusState) SaveMyVote() {
	if cs.lastMyVote == nil {
		cs.lastMyVote = cs.myVote
	} else if cs.myVote != nil && !bytes.Equal(cs.myVote.Signature, cs.lastMyVote.Signature) {
		cs.lastMyVote = cs.myVote
	}
}

// SetMyVote method
func (cs *ConsensusState) SetMyVote(vote *dpostype.DPosVote) {
	cs.myVote = vote
}

// SaveNotify method
func (cs *ConsensusState) SaveNotify() {
	if cs.lastNotify == nil {
		cs.lastNotify = cs.notify
	} else if cs.notify != nil && !bytes.Equal(cs.notify.Signature, cs.lastNotify.Signature) {
		cs.lastNotify = cs.notify
	}
}

// SetNotify method
func (cs *ConsensusState) SetNotify(notify *dpostype.DPosNotify) {
	if cs.notify != nil && !bytes.Equal(cs.notify.Signature, notify.Signature) {
		cs.lastNotify = cs.notify
	}

	cs.notify = notify
}

// CacheNotify method
func (cs *ConsensusState) CacheNotify(notify *dpostype.DPosNotify) {
	cs.cachedNotify = notify
}

// ClearCachedNotify method
func (cs *ConsensusState) ClearCachedNotify() {
	cs.cachedNotify = nil
}

// AddVotes method
func (cs *ConsensusState) AddVotes(vote *dpostype.DPosVote) {
	repeatFlag := false
	addrExistFlag := false
	index := -1

	if cs.lastVote != nil && vote.VoteItem.PeriodStart < cs.lastVote.PeriodStop {
		dposlog.Info("Old vote, discard it", "vote.PeriodStart", vote.VoteItem.PeriodStart, "last vote.PeriodStop", cs.lastVote.PeriodStop)

		return
	}

	for i := 0; i < len(cs.dposVotes); i++ {
		if bytes.Equal(cs.dposVotes[i].Signature, vote.Signature) {
			repeatFlag = true
			break
		} else if bytes.Equal(cs.dposVotes[i].VoterNodeAddress, vote.VoterNodeAddress) {
			addrExistFlag = true
			index = i
			break
		}
	}

	//有重复投票，则不需要处理
	if repeatFlag {
		return
	}

	//投票不重复，如果地址也不重复，则直接加入;如果地址重复了，则替换老的投票
	if !addrExistFlag {
		cs.dposVotes = append(cs.dposVotes, vote)
	} else if vote.VoteTimestamp > cs.dposVotes[index].VoteTimestamp {
		cs.dposVotes[index] = vote
	}
}

// CacheVotes method
func (cs *ConsensusState) CacheVotes(vote *dpostype.DPosVote) {
	repeatFlag := false
	addrExistFlag := false
	index := -1

	for i := 0; i < len(cs.cachedVotes); i++ {
		if bytes.Equal(cs.cachedVotes[i].Signature, vote.Signature) {
			repeatFlag = true
			break
		} else if bytes.Equal(cs.cachedVotes[i].VoterNodeAddress, vote.VoterNodeAddress) {
			addrExistFlag = true
			index = i
			break
		}
	}

	//有重复投票，则不需要处理
	if repeatFlag {
		return
	}

	//投票不重复，如果地址也不重复，则直接加入;如果地址重复了，则替换老的投票
	if !addrExistFlag {
		cs.cachedVotes = append(cs.cachedVotes, vote)
	} else if vote.VoteTimestamp > cs.cachedVotes[index].VoteTimestamp {
		cs.cachedVotes[index] = vote
	}
}

// CheckVotes method
func (cs *ConsensusState) CheckVotes() (ty int, vote *dpostype.VoteItem) {
	major32 := int(dposDelegateNum * 2 / 3)

	//总的票数还不够2/3，先不做决定
	if len(cs.dposVotes) < major32 {
		return continueToVote, nil
	}

	voteStat := map[string]int{}
	for i := 0; i < len(cs.dposVotes); i++ {
		key := string(cs.dposVotes[i].VoteItem.VoteID)
		if _, ok := voteStat[key]; ok {
			voteStat[key]++
		} else {
			voteStat[key] = 1
		}
	}

	key := ""
	value := 0

	for k, v := range voteStat {
		if v > value {
			value = v
			key = k
		}
	}

	//如果一个节点的投票数已经过2/3，则返回最终票数超过2/3的选票
	if value >= major32 {
		for i := 0; i < len(cs.dposVotes); i++ {
			if key == string(cs.dposVotes[i].VoteItem.VoteID) {
				return voteSuccess, cs.dposVotes[i].VoteItem
			}
		}
	} else if (value + (int(dposDelegateNum) - len(cs.dposVotes))) < major32 {
		//得票最多的节点，即使后续所有票都选它，也不满足2/3多数，不能达成共识。
		return voteFail, nil
	}

	return continueToVote, nil
}

// ClearVotes method
func (cs *ConsensusState) ClearVotes() {
	cs.dposVotes = nil
	cs.currentVote = nil
	cs.myVote = nil
}

// ClearCachedVotes method
func (cs *ConsensusState) ClearCachedVotes() {
	cs.cachedVotes = nil
}

// VerifyVote method
func (cs *ConsensusState) VerifyVote(vote *dpostype.DPosVote) bool {
	// Check validator
	index, val := cs.validatorMgr.Validators.GetByAddress(vote.VoterNodeAddress)
	if index == -1 && val == nil {
		dposlog.Info("The voter is not a legal validator, so discard this vote", "vote", vote.String())
		return false
	}
	// Verify signature
	pubkey, err := dpostype.ConsensusCrypto.PubKeyFromBytes(val.PubKey)
	if err != nil {
		dposlog.Error("Error pubkey from bytes", "err", err)
		return false
	}

	voteTmp := &dpostype.Vote{DPosVote: vote}
	if err := voteTmp.Verify(cs.validatorMgr.ChainID, pubkey); err != nil {
		dposlog.Error("Verify vote signature failed", "err", err)
		return false
	}

	return true
}

// VerifyNotify method
func (cs *ConsensusState) VerifyNotify(notify *dpostype.DPosNotify) bool {
	// Check validator
	index, val := cs.validatorMgr.Validators.GetByAddress(notify.NotifyNodeAddress)
	if index == -1 && val == nil {
		dposlog.Info("The notifier is not a legal validator, so discard this notify", "notify", notify.String())
		return false
	}
	// Verify signature
	pubkey, err := dpostype.ConsensusCrypto.PubKeyFromBytes(val.PubKey)
	if err != nil {
		dposlog.Error("Error pubkey from bytes", "err", err)
		return false
	}

	notifyTmp := &dpostype.Notify{DPosNotify: notify}
	if err := notifyTmp.Verify(cs.validatorMgr.ChainID, pubkey); err != nil {
		dposlog.Error("Verify vote signature failed", "err", err)
		return false
	}

	return true
}

// QueryCycleBoundaryInfo method
func (cs *ConsensusState) QueryCycleBoundaryInfo(cycle int64) (*dty.DposCBInfo, error) {
	req := &dty.DposCBQuery{Cycle: cycle, Ty: dty.QueryCBInfoByCycle}
	param, err := proto.Marshal(req)
	if err != nil {
		dposlog.Error("Marshal DposCBQuery failed", "cycle", cycle, "err", err)
		return nil, err
	}
	msg := cs.client.GetQueueClient().NewMessage("execs", types.EventBlockChainQuery,
		&types.ChainExecutor{
			Driver:    dty.DPosX,
			FuncName:  dty.FuncNameQueryCBInfoByCycle,
			StateHash: zeroHash[:],
			Param:     param,
		})

	err = cs.client.GetQueueClient().Send(msg, true)
	if err != nil {
		dposlog.Error("send DposCBQuery to dpos exec failed", "cycle", cycle, "err", err)
		return nil, err
	}

	msg, err = cs.client.GetQueueClient().Wait(msg)
	if err != nil {
		dposlog.Error("send DposCBQuery wait failed", "cycle", cycle, "err", err)
		return nil, err
	}

	res := msg.GetData().(types.Message).(*dty.DposCBReply)
	info := res.CbInfo
	dposlog.Info("DposCBQuery get reply", "cycle", cycle, "stopHeight", info.StopHeight, "stopHash", info.StopHash, "pubkey", info.Pubkey)

	return info, nil
}

// Init method
func (cs *ConsensusState) Init() {
	now := time.Now().Unix()
	task := DecideTaskByTime(now)
	cs.InitCycleBoundaryInfo(task)
	if shuffleType == dposShuffleTypeOrderByVrfInfo {
		cs.InitCycleVrfInfo(task)
		cs.InitCycleVrfInfos(task)
	}

	info := CalcTopNVersion(cs.client.GetCurrentHeight())
	cs.InitTopNCandidators(info.Version)
}

// InitTopNCandidators method
func (cs *ConsensusState) InitTopNCandidators(version int64) {
	for version > 0 && whetherUpdateTopN {
		info, err := cs.client.QueryTopNCandidators(version)
		if err == nil && info != nil && info.Status == dty.TopNCandidatorsVoteMajorOK {
			cs.UpdateTopNCandidators(info)
			return
		}

		version--
	}
}

// UpdateTopNCandidators method
func (cs *ConsensusState) UpdateTopNCandidators(info *dty.TopNCandidators) {
	if len(cs.cachedTopNCands) == 0 {
		cs.cachedTopNCands = append(cs.cachedTopNCands, info)
		return
	}

	if cs.cachedTopNCands[len(cs.cachedTopNCands)-1].Version < info.Version {
		cs.cachedTopNCands = append(cs.cachedTopNCands, info)
	}
}

// GetTopNCandidatorsByVersion method
func (cs *ConsensusState) GetTopNCandidatorsByVersion(version int64) (info *dty.TopNCandidators) {
	if len(cs.cachedTopNCands) == 0 || cs.cachedTopNCands[len(cs.cachedTopNCands)-1].Version < version {
		info, err := cs.client.QueryTopNCandidators(version)
		if err == nil && info != nil {
			if info.Status == dty.TopNCandidatorsVoteMajorOK {
				cs.UpdateTopNCandidators(info)
			}
			return info
		}
		return nil
	}

	for i := len(cs.cachedTopNCands) - 1; i >= 0; i-- {
		if cs.cachedTopNCands[i].Version == version {
			return cs.cachedTopNCands[i]
		} else if cs.cachedTopNCands[i].Version < version {
			return nil
		}
	}

	return nil
}

// GetLastestTopNCandidators method
func (cs *ConsensusState) GetLastestTopNCandidators() (info *dty.TopNCandidators) {
	length := len(cs.cachedTopNCands)
	if length > 0 {
		return cs.cachedTopNCands[length-1]
	}

	return nil
}

// IsTopNRegisted method
func (cs *ConsensusState) IsTopNRegisted(info *dty.TopNCandidators) bool {
	if nil == info {
		return false
	}

	for i := 0; i < len(info.CandsVotes); i++ {
		if bytes.Equal(info.CandsVotes[i].SignerPubkey, cs.privValidator.GetPubKey().Bytes()) {
			return true
		}
	}

	return false
}

// IsInTopN method
func (cs *ConsensusState) IsInTopN(info *dty.TopNCandidators) bool {
	if nil == info || info.Status != dty.TopNCandidatorsVoteMajorOK || len(info.FinalCands) == 0 {
		return false
	}

	for i := 0; i < len(info.FinalCands); i++ {
		if bytes.Equal(info.FinalCands[i].Pubkey, cs.privValidator.GetPubKey().Bytes()) {
			return true
		}
	}

	return false
}

// InitCycleBoundaryInfo method
func (cs *ConsensusState) InitCycleBoundaryInfo(task Task) {
	info, err := cs.QueryCycleBoundaryInfo(task.Cycle)
	if err == nil && info != nil {
		//cs.cycleBoundaryMap[task.cycle] = info
		cs.UpdateCBInfo(info)
		return
	}

	info, err = cs.QueryCycleBoundaryInfo(task.Cycle - 1)
	if err == nil && info != nil {
		//cs.cycleBoundaryMap[task.cycle] = info
		cs.UpdateCBInfo(info)
	}
}

// UpdateCBInfo method
func (cs *ConsensusState) UpdateCBInfo(info *dty.DposCBInfo) {
	valueNumber := len(cs.cycleBoundaryMap)
	if valueNumber == 0 {
		cs.cycleBoundaryMap[info.Cycle] = info
		return
	}

	oldestCycle := int64(0)
	for k := range cs.cycleBoundaryMap {
		if k == info.Cycle {
			cs.cycleBoundaryMap[info.Cycle] = info
			return
		}

		if oldestCycle == 0 {
			oldestCycle = k
		} else if oldestCycle > k {
			oldestCycle = k
		}
	}

	if valueNumber >= 5 {
		delete(cs.cycleBoundaryMap, oldestCycle)
	}

	cs.cycleBoundaryMap[info.Cycle] = info
}

// GetCBInfoByCircle method
func (cs *ConsensusState) GetCBInfoByCircle(cycle int64) (info *dty.DposCBInfo) {
	if v, ok := cs.cycleBoundaryMap[cycle]; ok {
		info = v
		return info
	}

	info, err := cs.QueryCycleBoundaryInfo(cycle)
	if err == nil && info != nil {
		cs.UpdateCBInfo(info)
		return info
	}

	return nil
}

// VerifyCBInfo method
func (cs *ConsensusState) VerifyCBInfo(info *dty.DposCBInfo) bool {
	// Verify signature
	bPubkey, err := hex.DecodeString(info.Pubkey)
	if err != nil {
		return false
	}
	pubkey, err := dpostype.ConsensusCrypto.PubKeyFromBytes(bPubkey)
	if err != nil {
		dposlog.Error("Error pubkey from bytes", "err", err)
		return false
	}

	bSig, err := hex.DecodeString(info.Signature)
	if err != nil {
		dposlog.Error("Error signature from bytes", "err", err)
		return false
	}

	sig, err := ttypes.ConsensusCrypto.SignatureFromBytes(bSig)
	if err != nil {
		dposlog.Error("CBInfo Verify failed", "err", err)
		return false
	}

	buf := new(bytes.Buffer)

	canonical := dty.CanonicalOnceCBInfo{
		Cycle:      info.Cycle,
		StopHeight: info.StopHeight,
		StopHash:   info.StopHash,
		Pubkey:     info.Pubkey,
	}

	byteCB, err := json.Marshal(&canonical)
	if err != nil {
		dposlog.Error("Error Marshal failed: ", "err", err)
		return false
	}

	_, err = buf.Write(byteCB)
	if err != nil {
		dposlog.Error("Error buf.Write failed: ", "err", err)
		return false
	}

	if !pubkey.VerifyBytes(buf.Bytes(), sig) {
		dposlog.Error("Error Verify Bytes failed: ", "err", err)
		return false
	}

	return true
}

// SendCBTx method
func (cs *ConsensusState) SendCBTx(info *dty.DposCBInfo) bool {
	//info.Pubkey = strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes()))
	canonical := dty.CanonicalOnceCBInfo{
		Cycle:      info.Cycle,
		StopHeight: info.StopHeight,
		StopHash:   info.StopHash,
		Pubkey:     info.Pubkey,
	}

	byteCB, err := json.Marshal(&canonical)
	if err != nil {
		dposlog.Error("marshal CanonicalOnceCBInfo failed", "err", err)
	}

	sig, err := cs.privValidator.SignMsg(byteCB)
	if err != nil {
		dposlog.Error("SignCBInfo failed.", "err", err)
		return false
	}

	info.Signature = hex.EncodeToString(sig.Bytes())
	tx, err := cs.client.CreateRecordCBTx(info)
	if err != nil {
		dposlog.Error("CreateRecordCBTx failed.", "err", err)
		return false
	}

	cs.privValidator.SignTx(tx)
	dposlog.Info("Sign RecordCBTx ok.")
	//将交易发往交易池中，方便后续重启或者新加入的超级节点查询
	msg := cs.client.GetQueueClient().NewMessage("mempool", types.EventTx, tx)
	err = cs.client.GetQueueClient().Send(msg, false)
	if err != nil {
		dposlog.Error("Send RecordCBTx to mempool failed.", "err", err)
		return false
	}

	dposlog.Info("Send RecordCBTx to mempool ok.")

	return true
}

// SendRegistVrfMTx method
func (cs *ConsensusState) SendRegistVrfMTx(info *dty.DposVrfMRegist) bool {
	tx, err := cs.client.CreateRegVrfMTx(info)
	if err != nil {
		dposlog.Error("CreateRegVrfMTx failed.", "err", err)
		return false
	}
	cs.privValidator.SignTx(tx)
	dposlog.Info("Sign RegistVrfMTx ok.")
	//将交易发往交易池中，方便后续重启或者新加入的超级节点查询
	msg := cs.client.GetQueueClient().NewMessage("mempool", types.EventTx, tx)
	err = cs.client.GetQueueClient().Send(msg, false)
	if err != nil {
		dposlog.Error("Send RegistVrfMTx to mempool failed.", "err", err)
		return false
	}

	dposlog.Info("Send RegistVrfMTx to mempool ok.")

	return true
}

// SendRegistVrfRPTx method
func (cs *ConsensusState) SendRegistVrfRPTx(info *dty.DposVrfRPRegist) bool {
	tx, err := cs.client.CreateRegVrfRPTx(info)
	if err != nil {
		dposlog.Error("CreateRegVrfRPTx failed.", "err", err)
		return false
	}

	cs.privValidator.SignTx(tx)
	dposlog.Info("Sign RegVrfRPTx ok.")
	//将交易发往交易池中，方便后续重启或者新加入的超级节点查询
	msg := cs.client.GetQueueClient().NewMessage("mempool", types.EventTx, tx)
	err = cs.client.GetQueueClient().Send(msg, false)
	if err != nil {
		dposlog.Error("Send RegVrfRPTx to mempool failed.", "err", err)
		return false
	}

	dposlog.Info("Send RegVrfRPTx to mempool ok.", "err", err)

	return true
}

// QueryVrf method
func (cs *ConsensusState) QueryVrf(pubkey []byte, cycle int64) (info *dty.VrfInfo, err error) {
	var pubkeys [][]byte
	pubkeys = append(pubkeys, pubkey)
	infos, err := cs.client.QueryVrfInfos(pubkeys, cycle)
	if err != nil {
		return nil, err
	}

	info = nil
	if len(infos) > 0 {
		info = infos[0]
	}

	return info, nil
}

// InitCycleVrfInfo method
func (cs *ConsensusState) InitCycleVrfInfo(task Task) {
	info, err := cs.QueryVrf(cs.privValidator.GetPubKey().Bytes(), task.Cycle)
	if err == nil && info != nil {
		//cs.cycleBoundaryMap[task.cycle] = info
		cs.UpdateVrfInfo(info)
		return
	}

	info, err = cs.QueryVrf(cs.privValidator.GetPubKey().Bytes(), task.Cycle-1)
	if err == nil && info != nil {
		//cs.cycleBoundaryMap[task.cycle] = info
		cs.UpdateVrfInfo(info)
	}
}

// UpdateVrfInfo method
func (cs *ConsensusState) UpdateVrfInfo(info *dty.VrfInfo) {
	valueNumber := len(cs.vrfInfoMap)
	if valueNumber == 0 {
		cs.vrfInfoMap[info.Cycle] = info
		return
	}

	oldestCycle := int64(0)
	for k := range cs.vrfInfoMap {
		if k == info.Cycle {
			cs.vrfInfoMap[info.Cycle] = info
			return
		}

		if oldestCycle == 0 {
			oldestCycle = k
		} else if oldestCycle > k {
			oldestCycle = k
		}
	}

	if valueNumber >= 5 {
		delete(cs.vrfInfoMap, oldestCycle)
	}

	cs.vrfInfoMap[info.Cycle] = info
}

// GetVrfInfoByCircle method
func (cs *ConsensusState) GetVrfInfoByCircle(cycle int64, ty int) (info *dty.VrfInfo) {
	if v, ok := cs.vrfInfoMap[cycle]; ok {
		info = v
		if VrfQueryTypeM == ty && len(info.M) > 0 {
			return info
		} else if VrfQueryTypeRP == ty && len(info.M) > 0 && len(info.R) > 0 && len(info.P) > 0 {
			return info
		}
	}

	info, err := cs.QueryVrf(cs.privValidator.GetPubKey().Bytes(), cycle)
	if err == nil && info != nil {
		cs.UpdateVrfInfo(info)
		return info
	}

	return nil
}

// QueryVrfs method
func (cs *ConsensusState) QueryVrfs(set *ttypes.ValidatorSet, cycle int64) (infos []*dty.VrfInfo, err error) {
	var pubkeys [][]byte

	for i := 0; i < set.Size(); i++ {
		pubkeys = append(pubkeys, set.Validators[i].PubKey)
	}
	infos, err = cs.client.QueryVrfInfos(pubkeys, cycle)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

// InitCycleVrfInfos method
func (cs *ConsensusState) InitCycleVrfInfos(task Task) {
	infos, err := cs.QueryVrfs(cs.validatorMgr.Validators, task.Cycle-1)
	if err == nil && infos != nil {
		//cs.cycleBoundaryMap[task.cycle] = info
		cs.UpdateVrfInfos(task.Cycle, infos)
	}
}

// UpdateVrfInfos method
func (cs *ConsensusState) UpdateVrfInfos(cycle int64, infos []*dty.VrfInfo) {
	if len(cs.validatorMgr.Validators.Validators) != len(infos) {
		return
	}

	for i := 0; i < len(infos); i++ {
		if len(infos[i].M) == 0 || len(infos[i].R) == 0 || len(infos[i].P) == 0 {
			return
		}
	}

	valueNumber := len(cs.vrfInfosMap)
	if valueNumber == 0 {
		cs.vrfInfosMap[cycle] = infos
		return
	}

	oldestCycle := int64(0)
	for k := range cs.vrfInfosMap {
		if k == cycle {
			cs.vrfInfosMap[cycle] = infos
			return
		}

		if oldestCycle == 0 {
			oldestCycle = k
		} else if oldestCycle > k {
			oldestCycle = k
		}
	}

	if valueNumber >= 5 {
		delete(cs.vrfInfosMap, oldestCycle)
	}

	cs.vrfInfosMap[cycle] = infos
}

// GetVrfInfosByCircle method
func (cs *ConsensusState) GetVrfInfosByCircle(cycle int64) (infos []*dty.VrfInfo) {
	if v, ok := cs.vrfInfosMap[cycle]; ok {
		infos = v
		return infos
	}

	infos, err := cs.QueryVrfs(cs.validatorMgr.Validators, cycle)
	if err == nil && len(infos) > 0 {
		cs.UpdateVrfInfos(cycle, infos)
		return infos
	}

	return nil
}

// ShuffleValidators method
func (cs *ConsensusState) ShuffleValidators(cycle int64) {
	if shuffleType == dposShuffleTypeFixOrderByAddr {
		dposlog.Info("ShuffleType FixOrderByAddr,so do nothing", "cycle", cycle)

		cs.validatorMgr.VrfValidators = nil
		cs.validatorMgr.NoVrfValidators = nil
		cs.validatorMgr.ShuffleCycle = cycle
		cs.validatorMgr.ShuffleType = ShuffleTypeNoVrf
		return
	}

	if cycle == cs.validatorMgr.ShuffleCycle {
		//如果已经洗过牌，则直接返回，不重复洗牌
		dposlog.Info("Shuffle for this cycle is done already.", "cycle", cycle)
		return
	}

	cbInfo := cs.GetCBInfoByCircle(cycle - 1)
	if cbInfo == nil {
		dposlog.Info("GetCBInfoByCircle failed", "cycle", cycle)
	} else {
		cs.validatorMgr.LastCycleBoundaryInfo = cbInfo
		dposlog.Info("GetCBInfoByCircle ok", "cycle", cycle, "stopHeight", cbInfo.StopHeight, "stopHash", cbInfo.StopHash)
	}

	infos := cs.GetVrfInfosByCircle(cycle - 1)
	if infos == nil {
		dposlog.Info("GetVrfInfosByCircle for Shuffle failed, don't use vrf to shuffle.", "cycle", cycle)

		cs.validatorMgr.VrfValidators = nil
		cs.validatorMgr.NoVrfValidators = nil
		cs.validatorMgr.ShuffleCycle = cycle
		cs.validatorMgr.ShuffleType = ShuffleTypeNoVrf
		return
	}

	var vrfValidators []*ttypes.Validator
	var noVrfValidators []*ttypes.Validator

	for i := 0; i < len(infos); i++ {
		if isValidVrfInfo(infos[i]) {
			var vrfBytes []byte
			//vrfBytes = append(vrfBytes, []byte(cbInfo.StopHash)...)
			vrfBytes = append(vrfBytes, infos[i].R...)

			item := &ttypes.Validator{
				PubKey: infos[i].Pubkey,
			}
			item.Address = crypto.Ripemd160(vrfBytes)
			vrfValidators = append(vrfValidators, item)
		}
	}

	set := cs.validatorMgr.Validators.Validators

	if len(vrfValidators) == 0 {
		dposlog.Info("Vrf validators is zero, don't use vrf to shuffle.", "cycle", cycle)

		cs.validatorMgr.ShuffleCycle = cycle
		cs.validatorMgr.ShuffleType = ShuffleTypeNoVrf
		return
	} else if len(vrfValidators) == len(set) {
		dposlog.Info("Vrf validators is full,use pure vrf to shuffle.", "cycle", cycle)

		cs.validatorMgr.ShuffleCycle = cycle
		cs.validatorMgr.ShuffleType = ShuffleTypeVrf
		cs.validatorMgr.VrfValidators = ttypes.NewValidatorSet(vrfValidators)
		return
	}

	cs.validatorMgr.ShuffleCycle = cycle
	cs.validatorMgr.ShuffleType = ShuffleTypePartVrf

	for i := 0; i < len(set); i++ {
		//如果节点信息不在VrfValidators，则说明没有完整的VRF信息，将被放入NoVrfValidators中
		if !isValidatorExist(set[i].PubKey, vrfValidators) {
			item := &ttypes.Validator{
				PubKey:  set[i].PubKey,
				Address: set[i].Address,
			}

			noVrfValidators = append(noVrfValidators, item)
		}
	}

	cs.validatorMgr.VrfValidators = ttypes.NewValidatorSet(vrfValidators)
	cs.validatorMgr.NoVrfValidators = ttypes.NewValidatorSet(noVrfValidators)
	dposlog.Info("Vrf validators is part,use part vrf to shuffle.", "cycle", cycle, "vrf validators size", cs.validatorMgr.VrfValidators.Size(), "non vrf validators size", cs.validatorMgr.NoVrfValidators.Size())
}

func isValidVrfInfo(info *dty.VrfInfo) bool {
	if info != nil && len(info.M) > 0 && len(info.R) > 0 && len(info.P) > 0 {
		return true
	}

	return false
}

func isValidatorExist(pubkey []byte, set []*ttypes.Validator) bool {
	for i := 0; i < len(set); i++ {
		if bytes.Equal(pubkey, set[i].PubKey) {
			return true
		}
	}

	return false
}

// VrfEvaluate method
func (cs *ConsensusState) VrfEvaluate(input []byte) (hash [32]byte, proof []byte) {
	return cs.privValidator.VrfEvaluate(input)
}

// VrfProof method
func (cs *ConsensusState) VrfProof(pubkey []byte, input []byte, hash [32]byte, proof []byte) bool {
	return cs.privValidator.VrfProof(pubkey, input, hash, proof)
}

// SendTopNRegistTx method
func (cs *ConsensusState) SendTopNRegistTx(reg *dty.TopNCandidatorRegist) bool {
	//info.Pubkey = strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes()))
	obj := dty.CanonicalTopNCandidator(reg.Cand)
	reg.Cand.Hash = obj.ID()
	reg.Cand.SignerPubkey = cs.privValidator.GetPubKey().Bytes()

	byteCB, err := json.Marshal(reg.Cand)
	if err != nil {
		dposlog.Error("marshal TopNCandidator failed", "err", err)
	}

	sig, err := cs.privValidator.SignMsg(byteCB)
	if err != nil {
		dposlog.Error("TopNCandidator failed.", "err", err)
		return false
	}

	reg.Cand.Signature = sig.Bytes()
	tx, err := cs.client.CreateTopNRegistTx(reg)
	if err != nil {
		dposlog.Error("CreateTopNRegistTx failed.", "err", err)
		return false
	}

	cs.privValidator.SignTx(tx)
	dposlog.Info("Sign TopNRegistTx ok.")
	//将交易发往交易池中，方便后续重启或者新加入的超级节点查询
	msg := cs.client.GetQueueClient().NewMessage("mempool", types.EventTx, tx)
	err = cs.client.GetQueueClient().Send(msg, false)
	if err != nil {
		dposlog.Error("Send TopNRegistTx to mempool failed.", "err", err)
		return false
	}
	dposlog.Info("Send TopNRegistTx to mempool ok.")

	return true
}
