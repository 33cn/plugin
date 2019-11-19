// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpos

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dpostype "github.com/33cn/plugin/plugin/consensus/dpos/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
)

var (
	// InitStateType 为状态机的初始状态
	InitStateType = 1
	// VotingStateType 为状态机的投票状态
	VotingStateType = 2
	// VotedStateType 为状态机的已投票状态
	VotedStateType = 3
	// WaitNotifyStateType 为状态机的等待通知状态
	WaitNotifyStateType = 4

	// StateTypeMapping 为状态的整型值和字符串值的对应关系
	StateTypeMapping = map[int]string{
		InitStateType:       "InitState",
		VotingStateType:     "VotingState",
		VotedStateType:      "VotedState",
		WaitNotifyStateType: "WaitNotifyState",
	}
)

// InitStateObj is the InitState obj
var InitStateObj = &InitState{}

// VotingStateObj is the VotingState obj
var VotingStateObj = &VotingState{}

// VotedStateObj is the VotedState obj
var VotedStateObj = &VotedState{}

// WaitNotifyStateObj is the WaitNotifyState obj
var WaitNotifyStateObj = &WaitNofifyState{}

// LastCheckVrfMTime is the Last Check Vrf M Time
var LastCheckVrfMTime = int64(0)

// LastCheckVrfRPTime is the Last Check Vrf RP Time
var LastCheckVrfRPTime = int64(0)

// LastCheckRegTopNTime is the Last Check Reg TopN Time
var LastCheckRegTopNTime = int64(0)

// LastCheckUpdateTopNTime is the Last Check Update TopN Time
var LastCheckUpdateTopNTime = int64(0)

// Task 为计算当前时间所属周期的数据结构
type Task struct {
	NodeID      int64
	Cycle       int64
	CycleStart  int64
	CycleStop   int64
	PeriodStart int64
	PeriodStop  int64
	BlockStart  int64
	BlockStop   int64
}

// TopNVersionInfo 为记录某一个区块高度对应的TopN更新的版本信息
type TopNVersionInfo struct {
	Version           int64
	HeightStart       int64
	HeightStop        int64
	HeightToStart     int64
	HeightRegLimit    int64
	HeightUpdateLimit int64
}

// CalcTopNVersion 根据某一个区块高度计算对应的TopN更新的版本信息
func CalcTopNVersion(height int64) (info TopNVersionInfo) {
	info = TopNVersionInfo{}
	info.Version = height / blockNumToUpdateDelegate
	info.HeightToStart = height % blockNumToUpdateDelegate
	info.HeightStart = info.Version * blockNumToUpdateDelegate
	info.HeightStop = (info.Version+1)*blockNumToUpdateDelegate - 1
	info.HeightRegLimit = info.HeightStart + registTopNHeightLimit
	info.HeightUpdateLimit = info.HeightStart + updateTopNHeightLimit
	return info
}

// DecideTaskByTime 根据时间戳计算所属的周期，包括cycle周期，负责出块周期，当前出块周期
func DecideTaskByTime(now int64) (task Task) {
	task.NodeID = now % dposCycle / dposPeriod
	task.Cycle = now / dposCycle
	task.CycleStart = now - now%dposCycle
	task.CycleStop = task.CycleStart + dposCycle - 1

	task.PeriodStart = task.CycleStart + task.NodeID*dposBlockInterval*dposContinueBlockNum
	task.PeriodStop = task.PeriodStart + dposPeriod - 1

	task.BlockStart = task.PeriodStart + now%dposCycle%dposPeriod/dposBlockInterval*dposBlockInterval
	task.BlockStop = task.BlockStart + dposBlockInterval - 1

	return task
}

func generateVote(cs *ConsensusState) *dpostype.Vote {
	//获得当前高度
	height := cs.client.GetCurrentHeight()
	now := time.Now().Unix()
	if cs.lastMyVote != nil && math.Abs(float64(now-cs.lastMyVote.VoteItem.PeriodStop)) <= 1 {
		now += 2
	}
	//计算当前时间，属于哪一个周期，应该哪一个节点出块，应该出块的高度
	task := DecideTaskByTime(now)

	cs.ShuffleValidators(task.Cycle)

	addr, validator := cs.validatorMgr.GetValidatorByIndex(int(task.NodeID))
	if addr == nil && validator == nil {
		dposlog.Error("Address and Validator is nil", "node index", task.NodeID, "now", now, "cycle", dposCycle, "period", dposPeriod)
		return nil
	}

	//生成vote， 对于vote进行签名
	voteItem := &dpostype.VoteItem{
		VotedNodeAddress: addr,
		VotedNodeIndex:   int32(task.NodeID),
		Cycle:            task.Cycle,
		CycleStart:       task.CycleStart,
		CycleStop:        task.CycleStop,
		PeriodStart:      task.PeriodStart,
		PeriodStop:       task.PeriodStop,
		Height:           height + 1,
	}

	encode, err := json.Marshal(voteItem)
	if err != nil {
		panic("Marshal vote failed.")
	}

	voteItem.VoteID = crypto.Ripemd160(encode)

	cs.validatorMgr.FillVoteItem(voteItem)

	index := cs.validatorMgr.GetIndexByPubKey(cs.privValidator.GetPubKey().Bytes())

	if index == -1 {
		panic("This node's address is not exist in Validators.")
	}

	vote := &dpostype.Vote{
		DPosVote: &dpostype.DPosVote{
			VoteItem:         voteItem,
			VoteTimestamp:    now,
			VoterNodeAddress: cs.privValidator.GetAddress(),
			VoterNodeIndex:   int32(index),
		},
	}

	return vote
}

func checkVrf(cs *ConsensusState) {
	if shuffleType != dposShuffleTypeOrderByVrfInfo {
		return
	}

	now := time.Now().Unix()
	task := DecideTaskByTime(now)
	middleTime := task.CycleStart + (task.CycleStop-task.CycleStart)/2
	if now < middleTime {
		if now-LastCheckVrfMTime < dposBlockInterval*2 {
			return
		}
		info := cs.GetVrfInfoByCircle(task.Cycle, VrfQueryTypeM)
		if info == nil {
			if cs.currentVote.LastCBInfo != nil {
				vrfM := &dty.DposVrfMRegist{
					Pubkey: strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes())),
					Cycle:  task.Cycle,
					//M: cs.currentVote.LastCBInfo.StopHash,
				}

				vrfM.M = cs.currentVote.LastCBInfo.StopHash
				dposlog.Info("SendRegistVrfMTx", "pubkey", vrfM.Pubkey, "cycle", vrfM.Cycle, "M", vrfM.M)
				cs.SendRegistVrfMTx(vrfM)
			} else {
				dposlog.Info("No available LastCBInfo, so don't SendRegistVrfMTx, just wait another cycle")
			}
		} else {
			dposlog.Info("VrfM is already registered", "now", now, "middle", middleTime, "cycle", task.Cycle, "pubkey", strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes())))
		}
		LastCheckVrfMTime = now
	} else {
		if now-LastCheckVrfRPTime < dposBlockInterval*2 {
			return
		}
		info := cs.GetVrfInfoByCircle(task.Cycle, VrfQueryTypeRP)
		if info != nil && len(info.M) > 0 && (len(info.R) == 0 || len(info.P) == 0) {
			hash, proof := cs.VrfEvaluate(info.M)

			vrfRP := &dty.DposVrfRPRegist{
				Pubkey: strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes())),
				Cycle:  task.Cycle,
				R:      hex.EncodeToString(hash[:]),
				P:      hex.EncodeToString(proof),
			}
			dposlog.Info("SendRegistVrfRPTx", "pubkey", vrfRP.Pubkey, "cycle", vrfRP.Cycle, "R", vrfRP.R, "P", vrfRP.P)

			cs.SendRegistVrfRPTx(vrfRP)
		} else if info != nil && len(info.M) > 0 && len(info.R) > 0 && len(info.P) > 0 {
			dposlog.Info("VrfRP is already registered", "now", now, "middle", middleTime, "cycle", task.Cycle, "pubkey", strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes())))
		} else {
			dposlog.Info("No available VrfM, so don't SendRegistVrfRPTx, just wait another cycle")
		}
		LastCheckVrfRPTime = now
	}

}

func checkTopNRegist(cs *ConsensusState) {
	if !whetherUpdateTopN {
		return
	}

	now := time.Now().Unix()
	if now-LastCheckRegTopNTime < dposBlockInterval*3 {
		//避免短时间频繁检查，5个区块以内不重复检查
		return
	}

	height := cs.client.GetCurrentHeight()
	info := CalcTopNVersion(height)
	if height <= info.HeightRegLimit {
		//在注册TOPN的区块区间内，则检查本节点是否注册成功，如果否则进行注册
		topN := cs.GetTopNCandidatorsByVersion(info.Version)
		if topN == nil || !cs.IsTopNRegisted(topN) {
			cands, err := cs.client.QueryCandidators()
			if err != nil || cands == nil || len(cands) != int(dposDelegateNum) {
				dposlog.Error("QueryCandidators failed", "now", now, "height", height, "HeightRegLimit", info.HeightRegLimit, "pubkey", strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes())))
				LastCheckRegTopNTime = now
				return

			}
			topNCand := &dty.TopNCandidator{
				Cands:        cands,
				Height:       height,
				SignerPubkey: cs.privValidator.GetPubKey().Bytes(),
			}
			obj := dty.CanonicalTopNCandidator(topNCand)
			topNCand.Hash = obj.ID()

			regist := &dty.TopNCandidatorRegist{
				Cand: topNCand,
			}

			cs.SendTopNRegistTx(regist)
			LastCheckRegTopNTime = now
		} else {
			dposlog.Info("TopN is already registered", "now", now, "height", height, "HeightRegLimit", info.HeightRegLimit, "pubkey", strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes())))
			LastCheckRegTopNTime = now + (info.HeightStop-height)*dposBlockInterval
		}
	} else {
		LastCheckRegTopNTime = now + (info.HeightStop-height)*dposBlockInterval
	}
}

func checkTopNUpdate(cs *ConsensusState) {
	if !whetherUpdateTopN {
		return
	}

	now := time.Now().Unix()
	if now-LastCheckUpdateTopNTime < dposBlockInterval*1 {
		//避免短时间频繁检查，1个区块以内不重复检查
		return
	}

	height := cs.client.GetCurrentHeight()
	info := CalcTopNVersion(height)
	if height >= info.HeightUpdateLimit {
		topN := cs.GetLastestTopNCandidators()
		if nil == topN {
			dposlog.Error("No valid topN, do nothing", "now", now, "height", height, "HeightUpdateLimit", info.HeightUpdateLimit, "pubkey", strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes())))
			LastCheckUpdateTopNTime = now + (info.HeightStop-height)*dposBlockInterval
			return
		}

		for i := 0; i < len(topN.FinalCands); i++ {
			if isPubkeyExist(topN.FinalCands[i].Pubkey, cs.validatorMgr.Validators.Validators) {
				continue
			} else {
				dposlog.Error("TopN changed, so restart to use latest topN", "now", now, "height", height, "HeightUpdateLimit", info.HeightUpdateLimit, "pubkey", strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes())))
				os.Exit(0)
			}
		}
		dposlog.Info("TopN not changed,so do nothing", "now", now, "height", height, "HeightUpdateLimit", info.HeightUpdateLimit, "pubkey", strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes())))
		LastCheckUpdateTopNTime = now + (info.HeightStop-height)*dposBlockInterval
	} else {
		LastCheckUpdateTopNTime = now + (info.HeightUpdateLimit-height-1)*dposBlockInterval
	}
}

func isPubkeyExist(pubkey []byte, validators []*dpostype.Validator) bool {
	for i := 0; i < len(validators); i++ {
		if bytes.Equal(pubkey, validators[i].PubKey) {
			return true
		}
	}

	return false
}

func recvCBInfo(cs *ConsensusState, info *dpostype.DPosCBInfo) {
	newInfo := &dty.DposCBInfo{
		Cycle:      info.Cycle,
		StopHeight: info.StopHeight,
		StopHash:   info.StopHash,
		Pubkey:     info.Pubkey,
		Signature:  info.Signature,
	}
	cs.UpdateCBInfo(newInfo)
}

func printNotify(notify *dpostype.DPosNotify) string {
	if notify.Vote.LastCBInfo != nil {
		return fmt.Sprintf("vote:[VotedNodeIndex:%d, VotedNodeAddr:%s,Cycle:%d,CycleStart:%d,CycleStop:%d,PeriodStart:%d,PeriodStop:%d,Height:%d,VoteID:%s,CBInfo[Cycle:%d,StopHeight:%d,StopHash:%s],ShuffleType:%d,ValidatorSize:%d,VrfValidatorSize:%d,NoVrfValidatorSize:%d];HeightStop:%d,HashStop:%s,NotifyTimestamp:%d,NotifyNodeIndex:%d,NotifyNodeAddrress:%s,Sig:%s",
			notify.Vote.VotedNodeIndex, hex.EncodeToString(notify.Vote.VotedNodeAddress), notify.Vote.Cycle, notify.Vote.CycleStart, notify.Vote.CycleStop, notify.Vote.PeriodStart, notify.Vote.PeriodStop, notify.Vote.Height, hex.EncodeToString(notify.Vote.VoteID),
			notify.Vote.LastCBInfo.Cycle, notify.Vote.LastCBInfo.StopHeight, notify.Vote.LastCBInfo.StopHash, notify.Vote.ShuffleType, len(notify.Vote.Validators), len(notify.Vote.VrfValidators), len(notify.Vote.NoVrfValidators),
			notify.HeightStop, hex.EncodeToString(notify.HashStop), notify.NotifyTimestamp, notify.NotifyNodeIndex, hex.EncodeToString(notify.NotifyNodeAddress), hex.EncodeToString(notify.Signature))
	}

	return fmt.Sprintf("vote:[VotedNodeIndex:%d, VotedNodeAddr:%s,Cycle:%d,CycleStart:%d,CycleStop:%d,PeriodStart:%d,PeriodStop:%d,Height:%d,VoteID:%s,CBInfo[],ShuffleType:%d,ValidatorSize:%d,VrfValidatorSize:%d,NoVrfValidatorSize:%d];HeightStop:%d,HashStop:%s,NotifyTimestamp:%d,NotifyNodeIndex:%d,NotifyNodeAddrress:%s,Sig:%s",
		notify.Vote.VotedNodeIndex, hex.EncodeToString(notify.Vote.VotedNodeAddress), notify.Vote.Cycle, notify.Vote.CycleStart, notify.Vote.CycleStop, notify.Vote.PeriodStart, notify.Vote.PeriodStop, notify.Vote.Height, hex.EncodeToString(notify.Vote.VoteID),
		notify.Vote.ShuffleType, len(notify.Vote.Validators), len(notify.Vote.VrfValidators), len(notify.Vote.NoVrfValidators),
		notify.HeightStop, hex.EncodeToString(notify.HashStop), notify.NotifyTimestamp, notify.NotifyNodeIndex, hex.EncodeToString(notify.NotifyNodeAddress), hex.EncodeToString(notify.Signature))
}

func printVote(vote *dpostype.DPosVote) string {
	return fmt.Sprintf("vote:[VotedNodeIndex:%d,VotedNodeAddress:%s,Cycle:%d,CycleStart:%d,CycleStop:%d,PeriodStart:%d,PeriodStop:%d,Height:%d,VoteID:%s,VoteTimestamp:%d,VoterNodeIndex:%d,VoterNodeAddress:%s,Sig:%s]",
		vote.VoteItem.VotedNodeIndex, common.ToHex(vote.VoteItem.VotedNodeAddress), vote.VoteItem.Cycle, vote.VoteItem.CycleStart, vote.VoteItem.CycleStop, vote.VoteItem.PeriodStart, vote.VoteItem.PeriodStop, vote.VoteItem.Height,
		hex.EncodeToString(vote.VoteItem.VoteID), vote.VoteTimestamp, vote.VoterNodeIndex, hex.EncodeToString(vote.VoterNodeAddress), hex.EncodeToString(vote.Signature))
}

func printVoteItem(voteItem *dpostype.VoteItem) string {
	if voteItem.LastCBInfo != nil {
		return fmt.Sprintf("vote:[VotedNodeIndex:%d, VotedNodeAddr:%s,Cycle:%d,CycleStart:%d,CycleStop:%d,PeriodStart:%d,PeriodStop:%d,Height:%d,VoteID:%s,CBInfo[Cycle:%d,StopHeight:%d,StopHash:%s],ShuffleType:%d,ValidatorSize:%d,VrfValidatorSize:%d,NoVrfValidatorSize:%d]",
			voteItem.VotedNodeIndex, hex.EncodeToString(voteItem.VotedNodeAddress), voteItem.Cycle, voteItem.CycleStart, voteItem.CycleStop, voteItem.PeriodStart, voteItem.PeriodStop, voteItem.Height, hex.EncodeToString(voteItem.VoteID),
			voteItem.LastCBInfo.Cycle, voteItem.LastCBInfo.StopHeight, voteItem.LastCBInfo.StopHash, voteItem.ShuffleType, len(voteItem.Validators), len(voteItem.VrfValidators), len(voteItem.NoVrfValidators))
	}

	return fmt.Sprintf("vote:[VotedNodeIndex:%d, VotedNodeAddr:%s,Cycle:%d,CycleStart:%d,CycleStop:%d,PeriodStart:%d,PeriodStop:%d,Height:%d,VoteID:%s,CBInfo[],ShuffleType:%d,ValidatorSize:%d,VrfValidatorSize:%d,NoVrfValidatorSize:%d]",
		voteItem.VotedNodeIndex, hex.EncodeToString(voteItem.VotedNodeAddress), voteItem.Cycle, voteItem.CycleStart, voteItem.CycleStop, voteItem.PeriodStart, voteItem.PeriodStop, voteItem.Height, hex.EncodeToString(voteItem.VoteID),
		voteItem.ShuffleType, len(voteItem.Validators), len(voteItem.VrfValidators), len(voteItem.NoVrfValidators))
}

// State is the base class of dpos state machine, it defines some interfaces.
type State interface {
	timeOut(cs *ConsensusState)
	sendVote(cs *ConsensusState, vote *dpostype.DPosVote)
	recvVote(cs *ConsensusState, vote *dpostype.DPosVote)
	sendVoteReply(cs *ConsensusState, reply *dpostype.DPosVoteReply)
	recvVoteReply(cs *ConsensusState, reply *dpostype.DPosVoteReply)
	sendNotify(cs *ConsensusState, notify *dpostype.DPosNotify)
	recvNotify(cs *ConsensusState, notify *dpostype.DPosNotify)
	recvCBInfo(cs *ConsensusState, info *dpostype.DPosCBInfo)
}

// InitState is the initial state of dpos state machine
type InitState struct {
}

func (init *InitState) timeOut(cs *ConsensusState) {
	//if available noes  < 2/3, don't change the state to voting.
	connections := cs.client.node.peerSet.Size()
	validators := cs.validatorMgr.Validators.Size()
	if dposDelegateNum > 1 && (connections == 0 || connections < (validators*2/3-1)) {
		dposlog.Error("InitState timeout but available nodes less than 2/3,waiting for more connections", "connections", connections, "validators", validators)
		cs.ClearVotes()

		//设定超时时间，超时后再检查链接数量
		cs.resetTimer(time.Duration(timeoutCheckConnections)*time.Millisecond, InitStateType)
	} else {
		vote := generateVote(cs)
		if nil == vote {
			cs.resetTimer(time.Duration(timeoutCheckConnections)*time.Millisecond, InitStateType)
			return
		}

		if err := cs.privValidator.SignVote(cs.validatorMgr.ChainID, vote); err != nil {
			dposlog.Error("SignVote failed", "vote", vote.String())
			cs.resetTimer(time.Duration(timeoutCheckConnections)*time.Millisecond, InitStateType)
			return
		}

		vote2 := *vote.DPosVote

		cs.AddVotes(&vote2)
		cs.SetMyVote(&vote2)
		dposlog.Info("Available nodes equal or more than 2/3,change state to VotingState", "connections", connections, "validators", validators)
		cs.SetState(VotingStateObj)
		dposlog.Info("Change state.", "from", "InitState", "to", "VotingState")
		//通过node发送p2p消息到其他节点
		dposlog.Info("VotingState send a vote", "vote info", printVote(vote.DPosVote), "localNodeIndex", cs.client.ValidatorIndex(), "now", time.Now().Unix())
		cs.dposState.sendVote(cs, vote.DPosVote)

		cs.resetTimer(time.Duration(timeoutVoting)*time.Millisecond, VotingStateType)
		//处理之前缓存的投票信息
		for i := 0; i < len(cs.cachedVotes); i++ {
			cs.dposState.recvVote(cs, cs.cachedVotes[i])
		}
		cs.ClearCachedVotes()
	}
}

func (init *InitState) sendVote(cs *ConsensusState, vote *dpostype.DPosVote) {
	dposlog.Info("InitState does not support sendVote,so do nothing")
}

func (init *InitState) recvVote(cs *ConsensusState, vote *dpostype.DPosVote) {
	dposlog.Info("InitState recvVote ,add it and will handle it later.")
	cs.CacheVotes(vote)
}

func (init *InitState) sendVoteReply(cs *ConsensusState, reply *dpostype.DPosVoteReply) {
	dposlog.Info("InitState don't support sendVoteReply,so do nothing")
}

func (init *InitState) recvVoteReply(cs *ConsensusState, reply *dpostype.DPosVoteReply) {
	dposlog.Info("InitState recv Vote reply,ignore it.")
}

func (init *InitState) sendNotify(cs *ConsensusState, notify *dpostype.DPosNotify) {
	dposlog.Info("InitState does not support sendNotify,so do nothing")
}

func (init *InitState) recvNotify(cs *ConsensusState, notify *dpostype.DPosNotify) {
	dposlog.Info("InitState recvNotify")

	//zzh:需要增加对Notify的处理，可以考虑记录已经确认过的出快记录
	cs.SetNotify(notify)
}

func (init *InitState) recvCBInfo(cs *ConsensusState, info *dpostype.DPosCBInfo) {
	dposlog.Info("InitState recvCBInfo")
	recvCBInfo(cs, info)
}

// VotingState is the voting state of dpos state machine until timeout or get an agreement by votes.
type VotingState struct {
}

func (voting *VotingState) timeOut(cs *ConsensusState) {
	//如果是测试场景，只有一个节点，也需要状态机能运转下去
	if dposDelegateNum == 1 {
		result, voteItem := cs.CheckVotes()

		if result == voteSuccess {
			dposlog.Info("VotingState get 2/3 result", "final vote:", printVoteItem(voteItem))
			dposlog.Info("VotingState change state to VotedState")
			//切换状态
			cs.SetState(VotedStateObj)
			dposlog.Info("Change state because of check votes successfully.", "from", "VotingState", "to", "VotedState")

			cs.SetCurrentVote(voteItem)

			//检查最终投票是否与自己的投票一致，如果不一致，需要更新本地的信息，保证各节点共识结果执行一致。
			if !bytes.Equal(cs.myVote.VoteItem.VoteID, voteItem.VoteID) {
				if !cs.validatorMgr.UpdateFromVoteItem(voteItem) {
					panic("This node's validators are not the same with final vote, please check")
				}
			}
			//1s后检查是否出块，是否需要重新投票
			cs.resetTimer(time.Millisecond*500, VotedStateType)
		}
		return
	}

	dposlog.Info("VotingState timeout but don't get an agreement. change state to InitState")

	//清理掉之前的选票记录，从初始状态重新开始
	cs.ClearVotes()
	cs.ClearCachedVotes()
	cs.SetState(InitStateObj)
	dposlog.Info("Change state because of timeOut.", "from", "VotingState", "to", "InitState")

	//由于连接多数情况下正常，快速触发InitState的超时处理
	cs.resetTimer(time.Duration(timeoutCheckConnections)*time.Millisecond, InitStateType)
}

func (voting *VotingState) sendVote(cs *ConsensusState, vote *dpostype.DPosVote) {
	cs.broadcastChannel <- MsgInfo{TypeID: dpostype.VoteID, Msg: vote, PeerID: cs.ourID, PeerIP: ""}
}

func (voting *VotingState) recvVote(cs *ConsensusState, vote *dpostype.DPosVote) {
	dposlog.Info("VotingState get a vote", "vote info", printVote(vote), "localNodeIndex", cs.client.ValidatorIndex(), "now", time.Now().Unix())

	if !cs.VerifyVote(vote) {
		dposlog.Info("VotingState verify vote failed")
		return
	}

	cs.AddVotes(vote)

	result, voteItem := cs.CheckVotes()

	if result == voteSuccess {
		dposlog.Info("VotingState get 2/3 result", "final vote:", printVoteItem(voteItem))
		dposlog.Info("VotingState change state to VotedState")
		//切换状态
		cs.SetState(VotedStateObj)
		dposlog.Info("Change state because of check votes successfully.", "from", "VotingState", "to", "VotedState")

		cs.SetCurrentVote(voteItem)

		//检查最终投票是否与自己的投票一致，如果不一致，需要更新本地的信息，保证各节点共识结果执行一致。
		if !bytes.Equal(cs.myVote.VoteItem.VoteID, voteItem.VoteID) {
			if !cs.validatorMgr.UpdateFromVoteItem(voteItem) {
				panic("This node's validators are not the same with final vote, please check")
			}
		}
		//1s后检查是否出块，是否需要重新投票
		cs.resetTimer(time.Millisecond*500, VotedStateType)
	} else if result == continueToVote {
		dposlog.Info("VotingState get a vote, but don't get an agreement,waiting for new votes...")
	} else {
		dposlog.Info("VotingState get a vote, but don't get an agreement,vote fail,abort voting")
		//清理掉之前的选票记录，从初始状态重新开始
		cs.ClearVotes()
		cs.SetState(InitStateObj)
		dposlog.Info("Change state because of vote failed.", "from", "VotingState", "to", "InitState")
		cs.resetTimer(time.Duration(timeoutCheckConnections)*time.Millisecond, InitStateType)
	}
}

func (voting *VotingState) sendVoteReply(cs *ConsensusState, reply *dpostype.DPosVoteReply) {
	dposlog.Info("VotingState don't support sendVoteReply,so do nothing")
}

func (voting *VotingState) recvVoteReply(cs *ConsensusState, reply *dpostype.DPosVoteReply) {
	dposlog.Info("VotingState recv Vote reply")
	voting.recvVote(cs, reply.Vote)
}

func (voting *VotingState) sendNotify(cs *ConsensusState, notify *dpostype.DPosNotify) {
	dposlog.Info("VotingState does not support sendNotify,so do nothing")
}

func (voting *VotingState) recvNotify(cs *ConsensusState, notify *dpostype.DPosNotify) {
	dposlog.Info("VotingState does not support recvNotify,so do nothing")
}

func (voting *VotingState) recvCBInfo(cs *ConsensusState, info *dpostype.DPosCBInfo) {
	dposlog.Info("VotingState recvCBInfo")
	recvCBInfo(cs, info)
}

// VotedState is the voted state of dpos state machine after getting an agreement for a period
type VotedState struct {
}

func (voted *VotedState) timeOut(cs *ConsensusState) {
	now := time.Now().Unix()
	block := cs.client.GetCurrentBlock()
	task := DecideTaskByTime(now)
	cfg := cs.client.GetAPI().GetConfig()

	dposlog.Info("address info", "privValidatorAddr", hex.EncodeToString(cs.privValidator.GetAddress()), "VotedNodeAddress", hex.EncodeToString(cs.currentVote.VotedNodeAddress))
	if bytes.Equal(cs.privValidator.GetAddress(), cs.currentVote.VotedNodeAddress) {
		//当前节点为出块节点

		//如果区块未同步，则等待；如果区块已同步，则进行后续正常出块的判断和处理。
		if block.Height+1 < cs.currentVote.Height {
			dposlog.Info("VotedState timeOut but block is not sync,wait...", "localHeight", block.Height, "vote height", cs.currentVote.Height)
			cs.resetTimer(time.Second*1, VotedStateType)
			return
		}

		//时间到了节点切换时刻
		if now >= cs.currentVote.PeriodStop {
			//当前时间超过了节点切换时间，需要进行重新投票
			dposlog.Info("VotedState timeOut over periodStop.", "periodStop", cs.currentVote.PeriodStop, "cycleStop", cs.currentVote.CycleStop)

			isCycleSwith := false
			//如果到了cycle结尾，需要构造一个交易，把最终的CycleBoundary信息发布出去
			if cs.currentVote.PeriodStop == cs.currentVote.CycleStop {
				dposlog.Info("Create new tx for cycle change to record cycle boundary info.", "height", block.Height)
				isCycleSwith = true

				info := &dty.DposCBInfo{
					Cycle:      cs.currentVote.Cycle,
					StopHeight: block.Height,
					StopHash:   hex.EncodeToString(block.Hash(cfg)),
					Pubkey:     strings.ToUpper(hex.EncodeToString(cs.privValidator.GetPubKey().Bytes())),
				}

				info2 := &dpostype.DPosCBInfo{
					Cycle:      info.Cycle,
					StopHeight: info.StopHeight,
					StopHash:   info.StopHash,
					Pubkey:     info.Pubkey,
					Signature:  info.Signature,
				}
				cs.SendCBTx(info)

				cs.UpdateCBInfo(info)

				dposlog.Info("Send CBInfo in consensus network", "cycle", info2.Cycle, "stopHeight", info2.StopHeight, "stopHash", info2.StopHash, "pubkey", info2.Pubkey)
				voted.sendCBInfo(cs, info2)
			}

			//当前时间超过了节点切换时间，需要进行重新投票
			notify := &dpostype.Notify{
				DPosNotify: &dpostype.DPosNotify{
					Vote:              cs.currentVote,
					HeightStop:        block.Height,
					HashStop:          block.Hash(cfg),
					NotifyTimestamp:   now,
					NotifyNodeAddress: cs.privValidator.GetAddress(),
					NotifyNodeIndex:   int32(cs.privValidatorIndex),
				},
			}

			dposlog.Info("Send notify.", "HeightStop", notify.HeightStop, "HashStop", common.ToHex(notify.HashStop))

			if err := cs.privValidator.SignNotify(cs.validatorMgr.ChainID, notify); err != nil {
				dposlog.Error("SignNotify failed", "notify", notify.String())
				cs.SaveVote()
				cs.SaveMyVote()
				cs.ClearVotes()

				cs.SetState(InitStateObj)
				dposlog.Info("Change state because of time.", "from", "VotedState", "to", "InitState")
				cs.resetTimer(time.Duration(timeoutCheckConnections)*time.Millisecond, InitStateType)
				return
			}

			cs.SaveVote()
			cs.SaveMyVote()
			cs.SaveNotify()

			notify2 := *notify
			cs.SetNotify(notify2.DPosNotify)
			cs.dposState.sendNotify(cs, notify.DPosNotify)
			cs.ClearVotes()

			//检查是否需要更新TopN，如果有更新，则更新TOPN节点后进入新的状态循环。
			if isCycleSwith {
				checkTopNUpdate(cs)
			}

			cs.SetState(InitStateObj)
			dposlog.Info("Change state because of time.", "from", "VotedState", "to", "InitState")
			cs.resetTimer(time.Duration(timeoutCheckConnections)*time.Millisecond, InitStateType)

			return
		}

		//根据时间进行vrf相关处理，如果在(cyclestart,middle)之间，发布M，如果在(middle,cyclestop)之间，发布R、P
		checkVrf(cs)

		//检查是否应该注册topN，是否已经注册topN
		checkTopNRegist(cs)

		//当前时间未到节点切换时间，则继续进行出块判断
		if block.BlockTime >= task.BlockStop {
			//已出块，或者时间落后了。
			dposlog.Info("VotedState timeOut but block already is generated.", "blocktime", block.BlockTime, "blockStop", task.BlockStop, "now", now)
			cs.resetTimer(time.Second*1, VotedStateType)

			return
		} else if block.BlockTime < task.BlockStart {
			//本出块周期尚未出块，则进行出块
			if task.BlockStop-now <= 1 {
				dposlog.Info("Create new block.", "height", block.Height+1)

				cs.client.SetBlockTime(task.BlockStop)
				cs.client.CreateBlock()
				cs.resetTimer(time.Millisecond*500, VotedStateType)
				return
			}

			dposlog.Info("Wait time to create block near blockStop.")
			cs.resetTimer(time.Millisecond*500, VotedStateType)
			return

		} else {
			//本周期已经出块
			dposlog.Info("Wait to next block cycle.", "waittime", task.BlockStop-now+1)

			//cs.scheduleDPosTimeout(time.Second * time.Duration(task.blockStop-now+1), VotedStateType)
			cs.resetTimer(time.Millisecond*500, VotedStateType)
			return
		}
	} else {
		dposlog.Info("This node is not current owner.", "current owner index", cs.currentVote.VotedNodeIndex, "this node index", cs.client.ValidatorIndex())

		//根据时间进行vrf相关处理，如果在(cyclestart,middle)之间，发布M，如果在(middle,cyclestop)之间，发布R、P
		checkVrf(cs)

		//检查是否应该注册topN，是否已经注册topN
		checkTopNRegist(cs)

		//非当前出块节点，如果到了切换出块节点的时间，则进行状态切换，进行投票
		if now >= cs.currentVote.PeriodStop {
			//当前时间超过了节点切换时间，需要进行重新投票
			cs.SaveVote()
			cs.SaveMyVote()
			cs.ClearVotes()
			cs.SetState(WaitNotifyStateObj)
			dposlog.Info("Change state because of time.", "from", "VotedState", "to", "WaitNotifyState")
			cs.resetTimer(time.Duration(timeoutWaitNotify)*time.Millisecond, WaitNotifyStateType)
			if cs.cachedNotify != nil {
				cs.dposState.recvNotify(cs, cs.cachedNotify)
			}
			return
		}

		//设置超时时间
		dposlog.Info("wait until change state.", "waittime", cs.currentVote.PeriodStop-now+1)
		cs.resetTimer(time.Second*time.Duration(cs.currentVote.PeriodStop-now+1), VotedStateType)
		return
	}
}

func (voted *VotedState) sendVoteReply(cs *ConsensusState, reply *dpostype.DPosVoteReply) {
	dposlog.Info("VotedState sendVoteReply")
	cs.broadcastChannel <- MsgInfo{TypeID: dpostype.VoteReplyID, Msg: reply, PeerID: cs.ourID, PeerIP: ""}
}

func (voted *VotedState) recvVoteReply(cs *ConsensusState, reply *dpostype.DPosVoteReply) {
	dposlog.Info("VotedState recv Vote reply", "from index", reply.Vote.VoterNodeIndex, "local index", cs.privValidatorIndex)
	cs.AddVotes(reply.Vote)
}

func (voted *VotedState) sendVote(cs *ConsensusState, vote *dpostype.DPosVote) {
	dposlog.Info("VotedState does not support sendVote,so do nothing")
}

func (voted *VotedState) recvVote(cs *ConsensusState, vote *dpostype.DPosVote) {
	dposlog.Info("VotedState recv Vote, will reply it", "from index", vote.VoterNodeIndex, "local index", cs.privValidatorIndex)
	if cs.currentVote.PeriodStart >= vote.VoteItem.PeriodStart {
		vote2 := *cs.myVote
		reply := &dpostype.DPosVoteReply{Vote: &vote2}
		cs.dposState.sendVoteReply(cs, reply)
	} else {
		dposlog.Info("VotedState recv future Vote, will cache it")

		cs.CacheVotes(vote)
	}
}

func (voted *VotedState) sendNotify(cs *ConsensusState, notify *dpostype.DPosNotify) {
	cs.broadcastChannel <- MsgInfo{TypeID: dpostype.NotifyID, Msg: notify, PeerID: cs.ourID, PeerIP: ""}
}

func (voted *VotedState) recvNotify(cs *ConsensusState, notify *dpostype.DPosNotify) {
	dposlog.Info("VotedState recvNotify")

	if bytes.Equal(cs.privValidator.GetAddress(), cs.currentVote.VotedNodeAddress) {
		dposlog.Info("ignore recvNotify because this node is the owner now.")
		return
	}

	cs.CacheNotify(notify)
	cs.SaveVote()
	cs.SaveMyVote()
	cs.ClearVotes()
	cs.SetState(WaitNotifyStateObj)
	dposlog.Info("Change state because of recv notify.", "from", "VotedState", "to", "WaitNotifyState")
	cs.resetTimer(time.Duration(timeoutWaitNotify)*time.Millisecond, WaitNotifyStateType)
	if cs.cachedNotify != nil {
		cs.dposState.recvNotify(cs, cs.cachedNotify)
	}
}

func (voted *VotedState) sendCBInfo(cs *ConsensusState, info *dpostype.DPosCBInfo) {
	cs.broadcastChannel <- MsgInfo{TypeID: dpostype.CBInfoID, Msg: info, PeerID: cs.ourID, PeerIP: ""}
}

func (voted *VotedState) recvCBInfo(cs *ConsensusState, info *dpostype.DPosCBInfo) {
	dposlog.Info("VotedState recvCBInfo")
	recvCBInfo(cs, info)
}

// WaitNofifyState is the state of dpos state machine to wait notify.
type WaitNofifyState struct {
}

func (wait *WaitNofifyState) timeOut(cs *ConsensusState) {
	//cs.clearVotes()

	//检查是否需要更新TopN，如果有更新，则更新TOPN节点后进入新的状态循环。
	now := time.Now().Unix()
	if now >= cs.lastVote.PeriodStop && cs.lastVote.PeriodStop == cs.lastVote.CycleStop {
		checkTopNUpdate(cs)
	}

	cs.SetState(InitStateObj)
	dposlog.Info("Change state because of time.", "from", "WaitNofifyState", "to", "InitState")
	cs.resetTimer(time.Duration(timeoutCheckConnections)*time.Millisecond, InitStateType)
}

func (wait *WaitNofifyState) sendVote(cs *ConsensusState, vote *dpostype.DPosVote) {
	dposlog.Info("WaitNofifyState does not support sendVote,so do nothing")
}

func (wait *WaitNofifyState) recvVote(cs *ConsensusState, vote *dpostype.DPosVote) {
	dposlog.Info("WaitNofifyState recvVote,store it.")
	//对于vote进行保存，在后续状态中进行处理。 保存的选票有先后，同一个节点发来的最新的选票被保存。
	cs.CacheVotes(vote)
}

func (wait *WaitNofifyState) sendVoteReply(cs *ConsensusState, reply *dpostype.DPosVoteReply) {
	dposlog.Info("WaitNofifyState does not support sendVoteReply,so do nothing")
}

func (wait *WaitNofifyState) recvVoteReply(cs *ConsensusState, reply *dpostype.DPosVoteReply) {
	dposlog.Info("WaitNofifyState recv Vote reply,ignore it.")
}

func (wait *WaitNofifyState) sendNotify(cs *ConsensusState, notify *dpostype.DPosNotify) {
	dposlog.Info("WaitNofifyState does not support sendNotify,so do nothing")
}

func (wait *WaitNofifyState) recvNotify(cs *ConsensusState, notify *dpostype.DPosNotify) {
	dposlog.Info("WaitNofifyState recvNotify")
	cfg := cs.client.GetAPI().GetConfig()
	//记录Notify，校验区块，标记不可逆区块高度
	if !cs.VerifyNotify(notify) {
		dposlog.Info("VotedState verify notify failed")
		return
	}

	block := cs.client.GetCurrentBlock()
	if block.Height > notify.HeightStop {
		dposlog.Info("Local block height is advanced than notify, discard it.", "localheight", block.Height, "notify", printNotify(notify))
		return
	} else if block.Height == notify.HeightStop && bytes.Equal(block.Hash(cfg), notify.HashStop) {
		dposlog.Info("Local block height is sync with notify", "notify", printNotify(notify))
	} else {
		dposlog.Info("Local block height is not sync with notify", "localheight", cs.client.GetCurrentHeight(), "notify", printNotify(notify))
		hint := time.NewTicker(3 * time.Second)
		beg := time.Now()
		catchupFlag := false
	OuterLoop:
		for !catchupFlag {
			select {
			case <-hint.C:
				dposlog.Info("Still catching up max height......", "Height", cs.client.GetCurrentHeight(), "notifyHeight", notify.HeightStop, "cost", time.Since(beg))
				if cs.client.IsCaughtUp() && cs.client.GetCurrentHeight() >= notify.HeightStop {
					dposlog.Info("This node has caught up max height", "Height", cs.client.GetCurrentHeight(), "isHashSame", bytes.Equal(block.Hash(cfg), notify.HashStop))
					break OuterLoop
				}

			default:
				if cs.client.IsCaughtUp() && cs.client.GetCurrentHeight() >= notify.HeightStop {
					dposlog.Info("This node has caught up max height", "Height", cs.client.GetCurrentHeight())
					break OuterLoop
				}
				time.Sleep(time.Second)
			}
		}
		hint.Stop()
	}

	cs.ClearCachedNotify()
	cs.SaveNotify()
	cs.SetNotify(notify)

	//检查是否需要更新TopN，如果有更新，则更新TOPN节点后进入新的状态循环。
	now := time.Now().Unix()
	if now >= cs.lastVote.PeriodStop && cs.lastVote.PeriodStop == cs.lastVote.CycleStop {
		checkTopNUpdate(cs)
	}

	cs.SetState(InitStateObj)
	dposlog.Info("Change state because recv notify.", "from", "WaitNofifyState", "to", "InitState")
	cs.dposState.timeOut(cs)
}

func (wait *WaitNofifyState) recvCBInfo(cs *ConsensusState, info *dpostype.DPosCBInfo) {
	dposlog.Info("WaitNofifyState recvCBInfo")
	recvCBInfo(cs, info)
}
