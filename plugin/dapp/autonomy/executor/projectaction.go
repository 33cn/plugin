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


func (a *action) propProject(prob *auty.ProposalProject) (*types.Receipt, error) {
	if err := address.CheckAddress(prob.ToAddr); err != nil {
		alog.Error("propProject ", "check toAddr error", err)
		return  nil, types.ErrInvalidParam
	}

	if prob.StartBlockHeight < a.height || prob.EndBlockHeight < a.height || prob.Amount <= 0 {
		return  nil, types.ErrInvalidParam
	}

	// 获取董事会成员
	value, err := a.db.Get(activeBoardID())
	if err != nil {
		err = auty.ErrNoActiveBoard
		alog.Error("propProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "get activeBoardID failed", err)
		return nil, err
	}
	pboard :=  &auty.ProposalBoard{}
	err = types.Decode(value, pboard)
	if err != nil {
		alog.Error("propProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode ProposalBoard failed", err)
		return nil, err
	}
	if len(pboard.Boards) > maxBoards || len(pboard.Boards) < minBoards  {
		err = auty.ErrNoActiveBoard
		alog.Error("propProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "illegality  boards number", err)
		return nil, err
	}

	// 获取当前生效提案规则,并且将不修改的规则补齐
	rule := &auty.RuleConfig{}
	value, err = a.db.Get(activeRuleID())
	if err == nil {
		err = types.Decode(value, rule)
		if err != nil {
			alog.Error("propProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode ProposalRule failed", err)
			return nil, err
		}
	} else {// 载入系统默认值
		rule.BoardAttendRatio   = boardAttendRatio
		rule.BoardApproveRatio  = boardApproveRatio
		rule.PubOpposeRatio     = pubOpposeRatio
		rule.ProposalAmount     = proposalAmount
		rule.LargeProjectAmount = largeProjectAmount
	}

	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, rule.ProposalAmount)
	if err != nil {
		alog.Error("propProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen amount", rule.ProposalAmount)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	var isPubVote bool
	if prob.Amount >= rule.LargeProjectAmount {
		isPubVote = true
	}
	cur := &auty.AutonomyProposalProject{
		PropProject:prob,
		CurRule:rule,
		Boards: pboard.Boards,
		BoardVoteRes: &auty.VoteResult{},
		PubVote: &auty.PublicVote{Publicity:isPubVote},
		Status: auty.AutonomyStatusProposalProject,
		Address: a.fromaddr,
		Height: a.height,
		Index: a.index,
	}
	kv = append(kv, &types.KeyValue{Key: propProjectID(common.ToHex(a.txhash)), Value: types.Encode(cur)})
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

	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, cur.CurRule.ProposalAmount)
	if err != nil {
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive amount", cur.CurRule.ProposalAmount, "err", err)
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

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalProject && cur.Status != auty.AutonomyStatusVotePropProject {
		err := auty.ErrProposalStatus
		alog.Error("votePropProject ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropProject().StartBlockHeight
	end := cur.GetPropProject().EndBlockHeight
	real := cur.GetPropProject().RealEndBlockHeight
	if start < a.height || end < a.height || (real != 0 && real < a.height) {
		err := auty.ErrVotePeriod
		alog.Error("votePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 董事会成员验证
	var isBoard bool
	for _, addr := range cur.Boards {
		if addr == a.fromaddr {
			isBoard = true
		}
	}
	if !isBoard {
		err = auty.ErrNoActiveBoard
		alog.Error("votePropProject ", "addr", a.fromaddr, "this addr is not active board member",
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
	for _, addr := range votes.Address {
		if addr == a.fromaddr {
			err := auty.ErrRepeatVoteAddr
			alog.Error("votePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "repeat address ProposalID",
				voteProb.ProposalID, "err", err)
			return nil, err
		}
	}
	// 更新已经投票地址
	votes.Address = append(votes.Address, a.fromaddr)
	// 更新投票结果
	cur.BoardVoteRes.TotalVotes = int32(len(cur.Boards))
	if voteProb.Approve {
		cur.BoardVoteRes.ApproveVotes += 1
	} else {
		cur.BoardVoteRes.OpposeVotes += 1
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if cur.BoardVoteRes.TotalVotes != 0 &&
		cur.BoardVoteRes.ApproveVotes + cur.BoardVoteRes.OpposeVotes != 0 &&
	    float32(cur.BoardVoteRes.ApproveVotes + cur.BoardVoteRes.OpposeVotes) / float32(cur.BoardVoteRes.TotalVotes) >=  float32(cur.CurRule.BoardAttendRatio)/100.0 &&
		float32(cur.BoardVoteRes.ApproveVotes) / float32(cur.BoardVoteRes.ApproveVotes + cur.BoardVoteRes.OpposeVotes) >= float32(cur.CurRule.BoardApproveRatio)/100.0 {
		cur.BoardVoteRes.Pass = true
		cur.PropProject.RealEndBlockHeight = a.height
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, cur.CurRule.ProposalAmount)
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
		if cur.PubVote.Publicity { // 进入公示
			cur.Status = auty.AutonomyStatusPubVotePropProject
			// 进入公示期默认为该提案通过，只有反对票达到三分之一才不会通过该提案
			cur.PubVote.PubPass = true
		} else {
			cur.Status = auty.AutonomyStatusTmintPropProject
			// 提案通过，将工程金额从基金付款给承包商
			receipt, err := a.coinsAccount.ExecTransferFrozen(autonomyAddr, cur.PropProject.ToAddr, a.execaddr, cur.PropProject.Amount)
			if err != nil {
				alog.Error("votePropProject ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen to contractor fail", err)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}
	}
	value = types.Encode(&cur)
	kv = append(kv, &types.KeyValue{Key: key, Value: value})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: VotesRecord(voteProb.ProposalID), Value: types.Encode(&votes)})

	ty := auty.TyLogVotePropProject
	if cur.BoardVoteRes.Pass {
		if cur.PubVote.Publicity {
			ty = auty.TyLogPubVotePropProject
		} else {
			ty = auty.TyLogTmintPropProject
		}
	}
	receiptLog := getProjectReceiptLog(pre, &cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) pubVotePropProject(voteProb *auty.PubVoteProposalProject) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(propProjectID(voteProb.ProposalID))
	if err != nil {
		alog.Error("pubVotePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "get propProjectID failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	var cur auty.AutonomyProposalProject
	err = types.Decode(value, &cur)
	if err != nil {
		alog.Error("pubVotePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode AutonomyProposalProject failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalProject(&cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusPubVotePropProject {
		err := auty.ErrProposalStatus
		alog.Error("pubVotePropProject ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropProject().StartBlockHeight
	end := cur.GetPropProject().EndBlockHeight
	real := cur.GetPropProject().RealEndBlockHeight
	if start < a.height || end < a.height || (real != 0 && real < a.height) {
		err := auty.ErrVotePeriod
		alog.Error("pubVotePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 检查是否已经参与投票
	var votes auty.VotesRecord
	value, err = a.db.Get(VotesRecord(voteProb.ProposalID))
	if err == nil {
		err = types.Decode(value, &votes)
		if err != nil {
			alog.Error("pubVotePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode VotesRecord failed",
				voteProb.ProposalID, "err", err)
			return nil, err
		}
	}
	for _, addr := range votes.Address {
		if addr == a.fromaddr {
			err := auty.ErrRepeatVoteAddr
			alog.Error("pubVotePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "repeat address GameID",
				voteProb.ProposalID, "err", err)
			return nil, err
		}
	}
	// 加入已经投票的
	votes.Address = append(votes.Address, a.fromaddr)

	if cur.GetBoardVoteRes().TotalVotes == 0 { //需要统计总票数
		addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
		account, err := a.getStartHeightVoteAccount(addr, start)
		if err != nil {
			return nil, err
		}
		cur.PubVote.TotalVotes = int32(account.Balance/ticketPrice)
	}

	// 获取该地址票数
	account, err := a.getStartHeightVoteAccount(a.fromaddr, start)
	if err != nil {
		return nil, err
	}
	if voteProb.Oppose { //投反对票
		cur.PubVote.OpposeVotes +=  int32(account.Balance/ticketPrice)
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if cur.PubVote.TotalVotes != 0 &&
	   float32(cur.PubVote.OpposeVotes) / float32(cur.PubVote.TotalVotes) >=  float32(cur.CurRule.PubOpposeRatio) {
		cur.PubVote.PubPass = false
		cur.PropProject.RealEndBlockHeight = a.height

		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("pubVotePropProject ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	key := propProjectID(voteProb.ProposalID)
	cur.Status = auty.AutonomyStatusPubVotePropProject
	ty := auty.TyLogPubVotePropProject
	if !cur.PubVote.PubPass {
		cur.Status = auty.AutonomyStatusTmintPropProject
		ty = auty.TyLogTmintPropProject
	}
	value = types.Encode(&cur)
	kv = append(kv, &types.KeyValue{Key: key, Value: value})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: VotesRecord(voteProb.ProposalID), Value: types.Encode(&votes)})

	receiptLog := getProjectReceiptLog(pre, &cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
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

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusTmintPropProject {
		err := auty.ErrProposalStatus
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}

	// 公示期间不能终止
	if cur.PubVote.Publicity && cur.PubVote.PubPass &&
		a.height <= cur.PropProject.EndBlockHeight + publicPeriod {
		err := auty.ErrTerminatePeriod
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "status", cur.Status,
			"in publicity vote period can not terminate", tmintProb.ProposalID, "err", err)
		return nil, err
	}

	// 董事会投票期间不能终止
	start := cur.GetPropProject().StartBlockHeight
	end := cur.GetPropProject().EndBlockHeight
	if !cur.PubVote.Publicity && a.height < end && !cur.BoardVoteRes.Pass {
		err := auty.ErrTerminatePeriod
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "status", cur.Status, "height", a.height,
			"in board vote period can not terminate", tmintProb.ProposalID, "err", err)
		return nil, err
	}

	if cur.GetBoardVoteRes().TotalVotes == 0 { //需要统计票数
		// 董事会成员验证
		value, err = a.db.Get(activeBoardID())
		if err != nil {
			err = auty.ErrNoActiveBoard
			alog.Error("tmintPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "get activeBoardID failed",
				tmintProb.ProposalID, "err", err)
			return nil, err
		}
		prob :=  &auty.ProposalBoard{}
		err = types.Decode(value, prob)
		if err != nil {
			alog.Error("tmintPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode ProposalBoard failed",
				tmintProb.ProposalID, "err", err)
			return nil, err
		}
		if len(prob.Boards) > maxBoards || len(prob.Boards) < minBoards  {
			err = auty.ErrNoActiveBoard
			alog.Error("tmintPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "illegality  boards number",
				tmintProb.ProposalID, "err", err)
			return nil, err
		}
		cur.BoardVoteRes.TotalVotes = int32(len(prob.Boards))
	}

	if cur.BoardVoteRes.TotalVotes != 0 &&
		cur.BoardVoteRes.ApproveVotes + cur.BoardVoteRes.OpposeVotes != 0 &&
		float32(cur.BoardVoteRes.ApproveVotes + cur.BoardVoteRes.OpposeVotes) / float32(cur.BoardVoteRes.TotalVotes) >=  float32(cur.CurRule.BoardAttendRatio)/100.0 &&
		float32(cur.BoardVoteRes.ApproveVotes) / float32(cur.BoardVoteRes.ApproveVotes + cur.BoardVoteRes.OpposeVotes) >= float32(cur.CurRule.BoardApproveRatio)/100.0 {
		cur.BoardVoteRes.Pass = true
	} else {
		cur.BoardVoteRes.Pass = false
	}

	if cur.PubVote.Publicity {
		if cur.GetBoardVoteRes().TotalVotes == 0 { //需要统计总票数
			addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
			account, err := a.getStartHeightVoteAccount(addr, start)
			if err != nil {
				return nil, err
			}
			cur.PubVote.TotalVotes = int32(account.Balance/ticketPrice)
		}
		if cur.PubVote.TotalVotes != 0 &&
		   float32(cur.PubVote.OpposeVotes) / float32(cur.PubVote.TotalVotes) >=  float32(cur.CurRule.PubOpposeRatio) {
		   	cur.PubVote.PubPass = false
		}
	}

	cur.PropProject.RealEndBlockHeight = a.height

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, cur.CurRule.ProposalAmount)
	if err != nil {
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	if (cur.PubVote.Publicity && cur.PubVote.PubPass) ||    // 需要公示且公示通过
		(!cur.PubVote.Publicity && cur.BoardVoteRes.Pass){  // 不需要公示且董事会通过
		// 提案通过，将工程金额从基金付款给承包商
		receipt, err := a.coinsAccount.ExecTransferFrozen(autonomyAddr, cur.PropProject.ToAddr, a.execaddr, cur.PropProject.Amount)
		if err != nil {
			alog.Error("tmintPropProject ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen to contractor fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

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

