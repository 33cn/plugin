// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"

)

const (
	ruleAttendRate  = 50 // 提案规则修改参与率
	ruleApproveRate = 50 // 提案规则修改赞成率
)


func (a *action) propRule(prob *auty.ProposalRule) (*types.Receipt, error) {
	//如果全小于等于0,则说明该提案规则参数不正确
	if prob.RuleCfg == nil || prob.RuleCfg.BoardAttendProb <= 0 && prob.RuleCfg.BoardPassProb <= 0  &&
	   prob.RuleCfg.OpposeProb <= 0 && prob.RuleCfg.ProposalAmount <= 0 && prob.RuleCfg.PubAmountThreshold <= 0 {
		return  nil, types.ErrInvalidParam
	}

	if prob.StartBlockHeight < a.height || prob.EndBlockHeight < a.height {
		return  nil, types.ErrInvalidParam
	}

	// 获取当前生效提案规则,并且将不修改的规则补齐
	rule := &auty.RuleConfig{}
	value, err := a.db.Get(activeRuleID())
	if err == nil {
		err = types.Decode(value, rule)
		if err != nil {
			alog.Error("propRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode ProposalRule failed", err)
			return nil, err
		}
	} else {// 载入系统默认值
	    rule.BoardAttendProb = participationRate
	    rule.BoardPassProb = approveRate
	    rule.OpposeProb = opposeRate
	    rule.ProposalAmount = lockAmount
	    rule.PubAmountThreshold = largeAmount
	}
	if prob.RuleCfg.BoardAttendProb > 0 {
		rule.BoardAttendProb = prob.RuleCfg.BoardAttendProb
	}
	if prob.RuleCfg.BoardPassProb > 0  {
		rule.BoardPassProb = prob.RuleCfg.BoardPassProb
	}
	if prob.RuleCfg.ProposalAmount > 0{
		rule.ProposalAmount = prob.RuleCfg.ProposalAmount
	}
	if prob.RuleCfg.PubAmountThreshold > 0 {
		rule.PubAmountThreshold = prob.RuleCfg.PubAmountThreshold
	}

	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, lockAmount)
	if err != nil {
		alog.Error("propRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen amount", lockAmount)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur := &auty.AutonomyProposalRule{
		PropRule:prob,
		Rule: rule,
		VoteResult: &auty.VoteResult{},
		Status: auty.AutonomyStatusProposalRule,
		Address: a.fromaddr,
		Height: a.height,
		Index: a.index,
	}

	key := propRuleID(common.ToHex(a.txhash))
	value = types.Encode(cur)
	kv = append(kv, &types.KeyValue{Key: key, Value: value})

	receiptLog := getRuleReceiptLog(nil, cur, auty.TyLogPropRule)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) rvkPropRule(rvkProb *auty.RevokeProposalRule) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(propRuleID(rvkProb.ProposalID))
	if err != nil {
		alog.Error("rvkPropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "get ProposalRule) failed",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}
	var cur auty.AutonomyProposalRule
	err = types.Decode(value, &cur)
	if err != nil {
		alog.Error("rvkPropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode ProposalRule failed",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalRule(&cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalRule {
		err := auty.ErrProposalStatus
		alog.Error("rvkPropRule ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropRule().StartBlockHeight
	if a.height > start {
		err := auty.ErrRevokeProposalPeriod
		alog.Error("rvkPropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	if a.fromaddr != cur.Address {
		err := auty.ErrRevokeProposalPower
		alog.Error("rvkPropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, lockAmount)
	if err != nil {
		alog.Error("rvkPropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive amount", lockAmount, "err", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur.Status = auty.AutonomyStatusRvkPropRule

	kv = append(kv, &types.KeyValue{Key: propRuleID(rvkProb.ProposalID), Value: types.Encode(&cur)})

	getRuleReceiptLog(pre, &cur, auty.TyLogRvkPropRule)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) votePropRule(voteProb *auty.VoteProposalRule) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(propRuleID(voteProb.ProposalID))
	if err != nil {
		alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "get propRuleID failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	var cur auty.AutonomyProposalRule
	err = types.Decode(value, &cur)
	if err != nil {
		alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode AutonomyProposalRule failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalRule(&cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalRule && cur.Status != auty.AutonomyStatusVotePropRule {
		err := auty.ErrProposalStatus
		alog.Error("votePropRule ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropRule().StartBlockHeight
	end := cur.GetPropRule().EndBlockHeight
	real := cur.GetPropRule().RealEndBlockHeight
	if start < a.height || end < a.height || (real != 0 && real < a.height) {
		err := auty.ErrVotePeriod
		alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 检查是否已经参与投票
	var votes auty.VotesRecord
	value, err = a.db.Get(VotesRecord(voteProb.ProposalID))
	if err == nil {
		err = types.Decode(value, &votes)
		if err != nil {
			alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode VotesRecord failed",
				voteProb.ProposalID, "err", err)
			return nil, err
		}
	}

	// 检查是否有重复
	for _, addr := range votes.Address {
		if addr == a.fromaddr {
			err := auty.ErrRepeatVoteAddr
			alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "repeat address GameID",
				voteProb.ProposalID, "err", err)
			return nil, err
		}
	}
	// 加入已经投票的
	votes.Address = append(votes.Address, a.fromaddr)

	if cur.GetVoteResult().TotalVotes == 0 { //需要统计票数
	    addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
		account, err := a.getStartHeightVoteAccount(addr, start)
		if err != nil {
			return nil, err
		}
		cur.VoteResult.TotalVotes = int32(account.Balance/ticketPrice)
	}

	account, err := a.getStartHeightVoteAccount(a.fromaddr, start)
	if err != nil {
		return nil, err
	}
	if voteProb.Approve {
		cur.VoteResult.ApproveVotes +=  int32(account.Balance/ticketPrice)
	} else {
		cur.VoteResult.OpposeVotes += int32(account.Balance/ticketPrice)
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if cur.VoteResult.TotalVotes != 0 &&
		cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes != 0 &&
	    float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) / float32(cur.VoteResult.TotalVotes) >= float32(ruleAttendRate)/100.0 &&
		float32(cur.VoteResult.ApproveVotes) / float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) >= float32(ruleApproveRate)/100.0 {
		cur.VoteResult.Pass = true
		cur.PropRule.RealEndBlockHeight = a.height

		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, cur.Rule.ProposalAmount)
		if err != nil {
			alog.Error("votePropRule ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	key := propRuleID(voteProb.ProposalID)
	cur.Status = auty.AutonomyStatusVotePropRule
	if cur.VoteResult.Pass {
		cur.Status = auty.AutonomyStatusTmintPropRule
	}
	value = types.Encode(&cur)
	kv = append(kv, &types.KeyValue{Key: key, Value: value})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: VotesRecord(voteProb.ProposalID), Value: types.Encode(&votes)})

	// 更新当前具有权利的董事会成员
	if cur.VoteResult.Pass {
		kv = append(kv, &types.KeyValue{Key: activeRuleID(), Value:types.Encode(cur.Rule)})
	}

	ty := auty.TyLogVotePropRule
	if cur.VoteResult.Pass {
		ty = auty.TyLogTmintPropRule
	}
	receiptLog := getRuleReceiptLog(pre, &cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) tmintPropRule(tmintProb *auty.TerminateProposalRule) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(propRuleID(tmintProb.ProposalID))
	if err != nil {
		alog.Error("tmintPropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "get propRuleID failed",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}
	var cur auty.AutonomyProposalRule
	err = types.Decode(value, &cur)
	if err != nil {
		alog.Error("tmintPropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode AutonomyProposalRule failed",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalRule(&cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusTmintPropRule {
		err := auty.ErrProposalStatus
		alog.Error("tmintPropRule ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropRule().StartBlockHeight
	end := cur.GetPropRule().EndBlockHeight
	if a.height < end && !cur.VoteResult.Pass {
		err := auty.ErrTerminatePeriod
		alog.Error("tmintPropRule ", "addr", a.fromaddr, "status", cur.Status, "height", a.height,
			"in vote period can not terminate", tmintProb.ProposalID, "err", err)
		return nil, err
	}

	if cur.GetVoteResult().TotalVotes == 0 { //需要统计票数
		addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
		account, err := a.getStartHeightVoteAccount(addr, start)
		if err != nil {
			return nil, err
		}
		cur.VoteResult.TotalVotes = int32(account.Balance/ticketPrice)
	}

	if float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) / float32(cur.VoteResult.TotalVotes) >=  float32(ruleAttendRate)/100.0 &&
		float32(cur.VoteResult.ApproveVotes) / float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) >= float32(ruleApproveRate)/100.0 {
		cur.VoteResult.Pass = true
	} else {
		cur.VoteResult.Pass = false
	}
	cur.PropRule.RealEndBlockHeight = a.height

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, cur.Rule.ProposalAmount)
	if err != nil {
		alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur.Status = auty.AutonomyStatusTmintPropRule

	kv = append(kv, &types.KeyValue{Key: propRuleID(tmintProb.ProposalID), Value: types.Encode(&cur)})

	// 更新当前具有权利的董事会成员
	if cur.VoteResult.Pass {
		kv = append(kv, &types.KeyValue{Key: activeRuleID(), Value:types.Encode(cur.Rule)})
	}

	getRuleReceiptLog(pre, &cur, auty.TyLogTmintPropRule)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

// getReceiptLog 根据提案信息获取log
// 状态变化：
func getRuleReceiptLog(pre, cur *auty.AutonomyProposalRule, ty int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = ty
	r := &auty.ReceiptProposalRule{Prev: pre, Current: cur}
	log.Log = types.Encode(r)
	return log
}

func copyAutonomyProposalRule(cur *auty.AutonomyProposalRule) *auty.AutonomyProposalRule {
	newAut := *cur
	newRule := *cur.GetPropRule()
	newCfg := *cur.GetPropRule().GetRuleCfg()
	newRes := *cur.GetVoteResult()
	newAut.PropRule = &newRule
	newAut.PropRule.RuleCfg = &newCfg
	newAut.VoteResult = &newRes
	return &newAut
}

