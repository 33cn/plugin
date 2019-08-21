// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"

	"github.com/33cn/chain33/system/dapp"
)

const (
	// 最小董事会赞成率
	minBoardApproveRatio = 50
	// 最大董事会赞成率
	maxBoardApproveRatio = 66
	// 最小全体持票人否决率
	minPubOpposeRatio    = 33
	// 最大全体持票人否决率
	maxPubOpposeRatio    = 50
	// 最小公示周期
	minPublicPeriod    int32 = 17280 * 7
	// 最大公示周期
	maxPublicPeriod    int32 = 17280 * 14
)

func (a *action) propRule(prob *auty.ProposalRule) (*types.Receipt, error) {
	//如果全小于等于0,则说明该提案规则参数不正确
	if prob.RuleCfg == nil || prob.RuleCfg.BoardAttendRatio < minBoardAttendRatio && prob.RuleCfg.BoardApproveRatio < minBoardApproveRatio &&
		prob.RuleCfg.PubOpposeRatio <= minPubOpposeRatio && prob.RuleCfg.ProposalAmount <= 0 && prob.RuleCfg.LargeProjectAmount <= 0 &&
		prob.RuleCfg.PublicPeriod <= 0 {
		alog.Error("propRule ", "ProposalRule RuleCfg invaild or have no modify param", prob.RuleCfg)
		return nil, types.ErrInvalidParam
	}
	if prob.RuleCfg.BoardAttendRatio > MaxBoardAttendRatio || prob.RuleCfg.BoardApproveRatio > maxBoardApproveRatio || prob.RuleCfg.PubOpposeRatio > maxPubOpposeRatio {
		alog.Error("propRule RuleCfg invaild", "BoardAttendRatio", prob.RuleCfg.BoardAttendRatio, "BoardApproveRatio",
			prob.RuleCfg.BoardApproveRatio, "PubOpposeRatio", prob.RuleCfg.PubOpposeRatio)
		return nil, types.ErrInvalidParam
	}
	if prob.StartBlockHeight < a.height || prob.EndBlockHeight < a.height {
		alog.Error("propRule height invaild", "StartBlockHeight", prob.StartBlockHeight, "EndBlockHeight",
			prob.EndBlockHeight, "height", a.height)
		return nil, types.ErrInvalidParam
	}

	// 获取当前生效提案规则,并且将不修改的规则补齐
	rule, err := a.getActiveRule()
	if err != nil {
		alog.Error("propRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "getActiveRule failed", err)
		return nil, err
	}

	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, rule.ProposalAmount)
	if err != nil {
		alog.Error("propRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen amount", rule.ProposalAmount)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur := &auty.AutonomyProposalRule{
		PropRule:   prob,
		CurRule:    rule,
		VoteResult: &auty.VoteResult{},
		Status:     auty.AutonomyStatusProposalRule,
		Address:    a.fromaddr,
		Height:     a.height,
		Index:      a.index,
		ProposalID: common.ToHex(a.txhash),
	}

	key := propRuleID(common.ToHex(a.txhash))
	value := types.Encode(cur)
	kv = append(kv, &types.KeyValue{Key: key, Value: value})

	receiptLog := getRuleReceiptLog(nil, cur, auty.TyLogPropRule)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) rvkPropRule(rvkProb *auty.RevokeProposalRule) (*types.Receipt, error) {
	cur, err := a.getProposalRule(rvkProb.ProposalID)
	if err != nil {
		alog.Error("rvkPropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalRule failed",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalRule(cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalRule {
		err := auty.ErrProposalStatus
		alog.Error("rvkPropRule ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropRule().StartBlockHeight
	if a.height >= start {
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

	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, cur.CurRule.ProposalAmount)
	if err != nil {
		alog.Error("rvkPropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive amount", cur.CurRule.ProposalAmount, "err", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur.Status = auty.AutonomyStatusRvkPropRule

	kv = append(kv, &types.KeyValue{Key: propRuleID(rvkProb.ProposalID), Value: types.Encode(cur)})

	receiptLog := getRuleReceiptLog(pre, cur, auty.TyLogRvkPropRule)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) votePropRule(voteProb *auty.VoteProposalRule) (*types.Receipt, error) {
	cur, err := a.getProposalRule(voteProb.ProposalID)
	if err != nil {
		alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalRule failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalRule(cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusRvkPropRule ||
		cur.Status == auty.AutonomyStatusTmintPropRule {
		err := auty.ErrProposalStatus
		alog.Error("votePropRule ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropRule().StartBlockHeight
	end := cur.GetPropRule().EndBlockHeight
	real := cur.GetPropRule().RealEndBlockHeight
	if a.height < start || a.height > end || real != 0 {
		err := auty.ErrVotePeriod
		alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 挖矿地址验证
	if len(voteProb.OriginAddr) > 0 {
		addr, err := a.verifyMinerAddr(voteProb.OriginAddr, a.fromaddr)
		if err != nil {
			alog.Error("votePropRule ", "from addr", a.fromaddr, "error addr", addr, "ProposalID",
				voteProb.ProposalID, "err", err)
			return nil, err
		}
	}

	// 本次参与投票地址
	var addrs []string
	if len(voteProb.OriginAddr) == 0 {
		addrs = append(addrs, a.fromaddr)
	} else {
		addrs = append(addrs, voteProb.OriginAddr...)
	}

	// 检查是否已经参与投票
	votes, err := a.checkVotesRecord(addrs, votesRecord(voteProb.ProposalID))
	if err != nil {
		alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "checkVotesRecord failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	// 更新投票记录
	votes.Address = append(votes.Address, addrs...)

	if cur.GetVoteResult().TotalVotes == 0 { //需要统计票数
		vtCouts, err := a.getTotalVotes(start)
		if err != nil {
			return nil, err
		}
		cur.VoteResult.TotalVotes = vtCouts
	}

	// 获取可投票数
	vtCouts, err := a.batchGetAddressVotes(addrs, start)
	if err != nil {
		alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "batchGetAddressVotes failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	if voteProb.Approve {
		cur.VoteResult.ApproveVotes += vtCouts
	} else {
		cur.VoteResult.OpposeVotes += vtCouts
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 首次进入投票期,即将提案金转入自治系统地址
	if cur.Status == auty.AutonomyStatusProposalRule {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyFundAddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("votePropRule ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if cur.VoteResult.TotalVotes != 0 &&
		cur.VoteResult.ApproveVotes+cur.VoteResult.OpposeVotes != 0 &&
		float32(cur.VoteResult.ApproveVotes+cur.VoteResult.OpposeVotes)/float32(cur.VoteResult.TotalVotes) > float32(pubAttendRatio)/100.0 &&
		float32(cur.VoteResult.ApproveVotes)/float32(cur.VoteResult.ApproveVotes+cur.VoteResult.OpposeVotes) > float32(pubApproveRatio)/100.0 {
		cur.VoteResult.Pass = true
		cur.PropRule.RealEndBlockHeight = a.height
	}

	key := propRuleID(voteProb.ProposalID)
	cur.Status = auty.AutonomyStatusVotePropRule
	if cur.VoteResult.Pass {
		cur.Status = auty.AutonomyStatusTmintPropRule
	}
	kv = append(kv, &types.KeyValue{Key: key, Value: types.Encode(cur)})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: votesRecord(voteProb.ProposalID), Value: types.Encode(votes)})

	// 更新系统规则
	if cur.VoteResult.Pass {
		upRule := upgradeRule(cur.CurRule, cur.PropRule.RuleCfg)
		kv = append(kv, &types.KeyValue{Key: activeRuleID(), Value: types.Encode(upRule)})
	}

	ty := auty.TyLogVotePropRule
	if cur.VoteResult.Pass {
		ty = auty.TyLogTmintPropRule
	}
	receiptLog := getRuleReceiptLog(pre, cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) tmintPropRule(tmintProb *auty.TerminateProposalRule) (*types.Receipt, error) {
	cur, err := a.getProposalRule(tmintProb.ProposalID)
	if err != nil {
		alog.Error("tmintPropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalRule failed",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}

	pre := copyAutonomyProposalRule(cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusTmintPropRule ||
		cur.Status == auty.AutonomyStatusRvkPropRule {
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
		vtCouts, err := a.getTotalVotes(start)
		if err != nil {
			return nil, err
		}
		cur.VoteResult.TotalVotes = vtCouts
	}

	if float32(cur.VoteResult.ApproveVotes+cur.VoteResult.OpposeVotes)/float32(cur.VoteResult.TotalVotes) > float32(pubAttendRatio)/100.0 &&
		float32(cur.VoteResult.ApproveVotes)/float32(cur.VoteResult.ApproveVotes+cur.VoteResult.OpposeVotes) > float32(pubApproveRatio)/100.0 {
		cur.VoteResult.Pass = true
	} else {
		cur.VoteResult.Pass = false
	}
	cur.PropRule.RealEndBlockHeight = a.height

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 未进行投票情况下，符合提案关闭的也需要扣除提案费用
	if cur.Status == auty.AutonomyStatusProposalRule {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyFundAddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("votePropRule ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)

	}

	cur.Status = auty.AutonomyStatusTmintPropRule

	kv = append(kv, &types.KeyValue{Key: propRuleID(tmintProb.ProposalID), Value: types.Encode(cur)})

	// 更新系统规则
	if cur.VoteResult.Pass {
		upRule := upgradeRule(cur.CurRule, cur.PropRule.RuleCfg)
		kv = append(kv, &types.KeyValue{Key: activeRuleID(), Value: types.Encode(upRule)})
	}
	receiptLog := getRuleReceiptLog(pre, cur, auty.TyLogTmintPropRule)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) transfer(tf *auty.TransferFund) (*types.Receipt, error) {
	if a.execaddr != dapp.ExecAddress(auty.AutonomyX) {
		err := auty.ErrNoAutonomyExec
		alog.Error("autonomy transfer ", "addr", a.fromaddr, "execaddr", a.execaddr, "this exec is not autonomy", err)
		return nil, err
	}
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	receipt, err := a.coinsAccount.ExecTransfer(a.fromaddr, autonomyFundAddr, a.execaddr, tf.Amount)
	if err != nil {
		alog.Error("autonomy transfer ", "addr", a.fromaddr, "amount", tf.Amount, "ExecTransfer fail", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) commentProp(cm *auty.Comment) (*types.Receipt, error) {
	if cm.Comment == "" || cm.ProposalID == "" {
		err := types.ErrInvalidParam
		alog.Error("autonomy commentProp ", "addr", a.fromaddr, "execaddr", a.execaddr, "Comment or proposalID empty", err)
		return nil, err
	}
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	receiptLog := getCommentReceiptLog(cm, a.height, a.index, common.ToHex(a.txhash), auty.TyLogCommentProp)
	logs = append(logs, receiptLog)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func getCommentReceiptLog(cur *auty.Comment, height int64, index int32, hash string, ty int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = ty
	r := &auty.ReceiptProposalComment{Cmt: cur, Height: height, Index: index, Hash: hash}
	log.Log = types.Encode(r)
	return log
}

func (a *action) getProposalRule(ID string) (*auty.AutonomyProposalRule, error) {
	value, err := a.db.Get(propRuleID(ID))
	if err != nil {
		return nil, err
	}
	cur := &auty.AutonomyProposalRule{}
	err = types.Decode(value, cur)
	if err != nil {
		return nil, err
	}
	return cur, nil
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
	if cur == nil {
		return nil
	}
	newAut := *cur
	if cur.PropRule != nil {
		newPropRule := *cur.GetPropRule()
		newAut.PropRule = &newPropRule
		if cur.PropRule.RuleCfg != nil {
			cfg := *cur.GetPropRule().GetRuleCfg()
			newAut.PropRule.RuleCfg = &cfg
		}
	}
	if cur.CurRule != nil {
		newRule := *cur.GetCurRule()
		newAut.CurRule = &newRule
	}
	if cur.VoteResult != nil {
		newRes := *cur.GetVoteResult()
		newAut.VoteResult = &newRes
	}
	return &newAut
}

func upgradeRule(cur, modify *auty.RuleConfig) *auty.RuleConfig {
	if cur == nil || modify == nil {
		return nil
	}
	new := *cur
	if modify.BoardAttendRatio > 0 {
		new.BoardAttendRatio = modify.BoardAttendRatio
	}
	if modify.BoardApproveRatio > 0 {
		new.BoardApproveRatio = modify.BoardApproveRatio
	}
	if modify.PubOpposeRatio > 0 {
		new.PubOpposeRatio = modify.PubOpposeRatio
	}
	if modify.ProposalAmount > 0 {
		new.ProposalAmount = modify.ProposalAmount
	}
	if modify.LargeProjectAmount > 0 {
		new.LargeProjectAmount = modify.LargeProjectAmount
	}
	if modify.PublicPeriod > 0 {
		new.PublicPeriod = modify.PublicPeriod
	}
	return &new
}
