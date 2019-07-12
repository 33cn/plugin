// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"

	"github.com/33cn/chain33/common/address"
)

const (
	// 重大项目金额阈值
	largeAmount = types.Coin * 100 *10000
)


func (a *action) propProject(prob *auty.ProposalProject) (*types.Receipt, error) {
	if err := address.CheckAddress(prob.ToAddr); err != nil {
		alog.Error("propProject ", "check toAddr error", err)
		return  nil, types.ErrInvalidParam
	}

	if prob.StartBlockHeight < a.height || prob.EndBlockHeight < a.height || prob.Amount <= 0 {
		return  nil, types.ErrInvalidParam
	}

	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, lockAmount)
	if err != nil {
		alog.Error("propProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen amount", lockAmount)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	var isPubVote bool
	if prob.Amount >= largeAmount {
		isPubVote = true
	}
	cur := &auty.AutonomyProposalProject{
		PropProject:prob,
		BoardVoteRes: &auty.VoteResult{},
		PubVote: &auty.PublicVote{Publicity:isPubVote},
		Status: auty.AutonomyStatusProposalProject,
		Address: a.fromaddr,
		Height: a.height,
		Index: a.index,
	}

	key := propProjectID(common.ToHex(a.txhash))
	value := types.Encode(cur)
	kv = append(kv, &types.KeyValue{Key: key, Value: value})

	receiptLog := getProjectReceiptLog(nil, cur, auty.TyLogPropProject)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) rvkPropProject(rvkProb *auty.RevokeProposalProject) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(propProjectID(rvkProb.ProposalID))
	if err != nil {
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "get ProposalProject) failed",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}
	var cur auty.AutonomyProposalProject
	err = types.Decode(value, &cur)
	if err != nil {
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode ProposalProject failed",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalProject(&cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalProject {
		err := auty.ErrProposalStatus
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropProject().StartBlockHeight
	if a.height > start {
		err := auty.ErrRevokeProposalPeriod
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	if a.fromaddr != cur.Address {
		err := auty.ErrRevokeProposalPower
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, lockAmount)
	if err != nil {
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive amount", lockAmount, "err", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur.Status = auty.AutonomyStatusRvkPropProject

	kv = append(kv, &types.KeyValue{Key: propProjectID(rvkProb.ProposalID), Value: types.Encode(&cur)})

	getProjectReceiptLog(pre, &cur, auty.TyLogRvkPropProject)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) votePropProject(voteProb *auty.VoteProposalProject) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(propProjectID(voteProb.ProposalID))
	if err != nil {
		alog.Error("votePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "get propProjectID failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	var cur auty.AutonomyProposalProject
	err = types.Decode(value, &cur)
	if err != nil {
		alog.Error("votePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode AutonomyProposalProject failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalProject(&cur)

	start := cur.GetPropProject().StartBlockHeight
	end := cur.GetPropProject().EndBlockHeight
	real := cur.GetPropProject().RealEndBlockHeight
	if start < a.height || end < a.height || (real != 0 && real < a.height) {
		err := auty.ErrVotePeriod
		alog.Error("votePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalProject && cur.Status != auty.AutonomyStatusVotePropProject {
		err := auty.ErrProposalStatus
		alog.Error("votePropProject ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 检查是否已经参与投票
	var votes auty.VotesRecord
	value, err = a.db.Get(VotesRecord(voteProb.ProposalID))
	if err == nil {
		err = types.Decode(value, &votes)
		if err != nil {
			alog.Error("votePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode VotesRecord failed",
				voteProb.ProposalID, "err", err)
			return nil, err
		}
	}

	// 检查是否有重复
	for _, addr := range votes.Address {
		if addr == a.fromaddr {
			err := auty.ErrRepeatVoteAddr
			alog.Error("votePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "repeat address GameID",
				voteProb.ProposalID, "err", err)
			return nil, err
		}
	}
	// 加入已经投票的
	votes.Address = append(votes.Address, a.fromaddr)

	if cur.GetBoardVoteRes().TotalVotes == 0 { //需要统计票数
	    addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
		account, err := a.getStartHeightVoteAccount(addr, start)
		if err != nil {
			return nil, err
		}
		cur.BoardVoteRes.TotalVotes = int32(account.Balance/ticketPrice)
	}

	account, err := a.getStartHeightVoteAccount(a.fromaddr, start)
	if err != nil {
		return nil, err
	}
	if voteProb.Approve {
		cur.BoardVoteRes.ApproveVotes +=  int32(account.Balance/ticketPrice)
	} else {
		cur.BoardVoteRes.OpposeVotes += int32(account.Balance/ticketPrice)
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if float32(cur.BoardVoteRes.ApproveVotes + cur.BoardVoteRes.OpposeVotes) / float32(cur.BoardVoteRes.TotalVotes) >=  participationRate &&
		float32(cur.BoardVoteRes.ApproveVotes) / float32(cur.BoardVoteRes.ApproveVotes + cur.BoardVoteRes.OpposeVotes) >= approveRate {
		cur.BoardVoteRes.Pass = true
		cur.PropProject.RealEndBlockHeight = a.height

		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, lockAmount)
		if err != nil {
			alog.Error("votePropProject ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	key := propProjectID(voteProb.ProposalID)
	cur.Status = auty.AutonomyStatusVotePropProject
	if cur.BoardVoteRes.Pass {
		cur.Status = auty.AutonomyStatusTmintPropProject
	}
	value = types.Encode(&cur)
	kv = append(kv, &types.KeyValue{Key: key, Value: value})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: VotesRecord(voteProb.ProposalID), Value: types.Encode(&votes)})

	ty := auty.TyLogVotePropProject
	if cur.BoardVoteRes.Pass {
		ty = auty.TyLogTmintPropProject
	}
	receiptLog := getProjectReceiptLog(pre, &cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) pubVotePropProject(voteProb *auty.PubVoteProposalProject) (*types.Receipt, error) {
	return nil, nil
}


func (a *action) tmintPropProject(tmintProb *auty.TerminateProposalProject) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(propProjectID(tmintProb.ProposalID))
	if err != nil {
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "get propProjectID failed",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}
	var cur auty.AutonomyProposalProject
	err = types.Decode(value, &cur)
	if err != nil {
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode AutonomyProposalProject failed",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalProject(&cur)

	start := cur.GetPropProject().StartBlockHeight
	end := cur.GetPropProject().EndBlockHeight
	if a.height < end && cur.Status != auty.AutonomyStatusVotePropProject {
		err := auty.ErrTerminatePeriod
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "status", cur.Status, "height", a.height, "ProposalID",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusTmintPropProject {
		err := auty.ErrProposalStatus
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}

	if cur.GetBoardVoteRes().TotalVotes == 0 { //需要统计票数
		addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
		account, err := a.getStartHeightVoteAccount(addr, start)
		if err != nil {
			return nil, err
		}
		cur.BoardVoteRes.TotalVotes = int32(account.Balance/ticketPrice)
	}

	if float32(cur.BoardVoteRes.ApproveVotes + cur.BoardVoteRes.OpposeVotes) / float32(cur.BoardVoteRes.TotalVotes) >=  participationRate &&
		float32(cur.BoardVoteRes.ApproveVotes) / float32(cur.BoardVoteRes.ApproveVotes + cur.BoardVoteRes.OpposeVotes) >= approveRate {
		cur.BoardVoteRes.Pass = true
	} else {
		cur.BoardVoteRes.Pass = false
	}
	cur.PropProject.RealEndBlockHeight = a.height

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, lockAmount)
	if err != nil {
		alog.Error("votePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur.Status = auty.AutonomyStatusTmintPropProject

	kv = append(kv, &types.KeyValue{Key: propProjectID(tmintProb.ProposalID), Value: types.Encode(&cur)})

	getProjectReceiptLog(pre, &cur, auty.TyLogTmintPropProject)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

// getProjectReceiptLog 根据提案信息获取log
// 状态变化：
func getProjectReceiptLog(pre, cur *auty.AutonomyProposalProject, ty int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = ty
	r := &auty.ReceiptProposalProject{Prev: pre, Current: cur}
	log.Log = types.Encode(r)
	return log
}

func copyAutonomyProposalProject(cur *auty.AutonomyProposalProject) *auty.AutonomyProposalProject {
	newAut := *cur
	newProject := *cur.GetPropProject()
	newRes := *cur.GetBoardVoteRes()
	newPub := *cur.GetPubVote()
	newAut.PropProject = &newProject
	newAut.BoardVoteRes = &newRes
	newAut.PubVote = &newPub
	return &newAut
}

