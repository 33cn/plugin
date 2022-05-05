// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/pkg/errors"
)

func (a *action) propItem(prob *auty.ProposalItem) (*types.Receipt, error) {
	autoCfg := GetAutonomyParam(a.api.GetConfig(), a.height)
	//start和end之间不能小于720高度，end不能超过当前高度+100w
	if prob.StartBlockHeight < a.height || prob.StartBlockHeight >= prob.EndBlockHeight ||
		prob.StartBlockHeight+autoCfg.StartEndBlockPeriod > prob.EndBlockHeight ||
		prob.EndBlockHeight > a.height+autoCfg.PropEndBlockPeriod ||
		prob.RealEndBlockHeight != 0 {
		return nil, errors.Wrapf(auty.ErrSetBlockHeight, "propItem exe height=%d,start=%d,end=%d,realEnd=%d",
			a.height, prob.StartBlockHeight, prob.EndBlockHeight, prob.RealEndBlockHeight)

	}

	if len(prob.ItemTxHash) <= 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "propItem tx hash nil")
	}

	// 获取董事会成员
	pboard, err := a.getActiveBoard()
	if err != nil {
		return nil, errors.Wrapf(err, "propItem.getActiveBoard")
	}

	// 获取当前生效提案规则
	rule, err := a.getActiveRule()
	if err != nil {
		alog.Error("propItem ", "addr", a.fromaddr, "execaddr", a.execaddr, "getActiveRule failed", err)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 冻结提案金
	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, rule.ProposalAmount)
	if err != nil {
		return nil, errors.Wrapf(err, "propItem.accountFrozen,proposalAmount=%d,addr=%s", rule.ProposalAmount, a.fromaddr)
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur := &auty.AutonomyProposalItem{
		PropItem:     prob,
		CurRule:      rule,
		Boards:       pboard.Boards,
		BoardVoteRes: &auty.VoteResult{TotalVotes: int32(len(pboard.Boards))},
		Status:       auty.AutonomyStatusProposalItem,
		Address:      a.fromaddr,
		Height:       a.height,
		Index:        a.index,
		ProposalID:   common.ToHex(a.txhash),
	}
	kv = append(kv, &types.KeyValue{Key: propItemID(common.ToHex(a.txhash)), Value: types.Encode(cur)})
	receiptLog := getItemReceiptLog(nil, cur, auty.TyLogPropItem)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) rvkPropItem(rvkProb *auty.RevokeProposalItem) (*types.Receipt, error) {
	cur, err := a.getProposalItem(rvkProb.ProposalID)
	if err != nil {
		alog.Error("rvkPropItem ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalItem failed",
			rvkProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "rvkPropItem.getItem,id=%s", rvkProb.ProposalID)
	}
	pre := copyAutonomyProposalItem(cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalItem {
		err := auty.ErrProposalStatus
		alog.Error("rvkPropItem ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			rvkProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "rvkPropItem wrong status =%d", cur.Status)
	}

	start := cur.GetPropItem().StartBlockHeight
	if a.height >= start {
		err := auty.ErrRevokeProposalPeriod
		alog.Error("rvkPropItem ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			rvkProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "rvkPropItem item started, startheight=%d,cur=%d", start, a.height)
	}

	if a.fromaddr != cur.Address {
		err := auty.ErrRevokeProposalPower
		alog.Error("rvkPropItem ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			rvkProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "rvkPropItem wrong from addr, from=%s,,cur=%s", a.fromaddr, cur.Address)
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 解冻提案金
	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, cur.CurRule.ProposalAmount)
	if err != nil {
		alog.Error("rvkPropItem ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive amount", cur.CurRule.ProposalAmount, "err", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur.Status = auty.AutonomyStatusRvkPropItem

	kv = append(kv, &types.KeyValue{Key: propItemID(rvkProb.ProposalID), Value: types.Encode(cur)})

	receiptLog := getItemReceiptLog(pre, cur, auty.TyLogRvkPropItem)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) votePropItem(voteProb *auty.VoteProposalItem) (*types.Receipt, error) {
	cur, err := a.getProposalItem(voteProb.ProposalID)
	if err != nil {
		alog.Error("votePropItem ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalItem failed",
			voteProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "votePropItem.getItem id=%s", voteProb.ProposalID)
	}
	pre := copyAutonomyProposalItem(cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusRvkPropItem ||
		cur.Status == auty.AutonomyStatusTmintPropItem {
		err := auty.ErrProposalStatus
		alog.Error("votePropItem ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "votePropItem cur status=%d", cur.Status)
	}

	start := cur.GetPropItem().StartBlockHeight
	end := cur.GetPropItem().EndBlockHeight
	realHeight := cur.GetPropItem().RealEndBlockHeight
	if a.height < start || a.height > end || realHeight != 0 {
		err := auty.ErrVotePeriod
		alog.Error("votePropItem ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "votePropItem current height=%d,start=%d,end=%d,real=%d",
			a.height, start, end, realHeight)
	}

	// 董事会成员验证
	var isBoard bool
	for _, addr := range cur.Boards {
		if addr == a.fromaddr {
			isBoard = true
			break
		}
	}
	if !isBoard {
		err = auty.ErrNoActiveBoard
		alog.Error("votePropItem ", "addr", a.fromaddr, "this addr is not active board member",
			voteProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "fromAddr notActiveBoardMember proposalid=%s", voteProb.ProposalID)
	}

	// 检查是否已经参与投票
	votes, err := a.checkVotesRecord([]string{a.fromaddr}, boardVotesRecord(voteProb.ProposalID))
	if err != nil {
		alog.Error("votePropItem ", "addr", a.fromaddr, "execaddr", a.execaddr, "checkVotesRecord boardVotesRecord failed",
			voteProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "check votes record from addr=%s", a.fromaddr)
	}

	// 更新已经投票地址
	votes.Address = append(votes.Address, a.fromaddr)

	// 更新投票结果
	switch voteProb.Vote {
	case auty.AutonomyVoteOption_APPROVE:
		cur.BoardVoteRes.ApproveVotes++
	case auty.AutonomyVoteOption_OPPOSE:
		cur.BoardVoteRes.OpposeVotes++
	case auty.AutonomyVoteOption_QUIT:
		cur.BoardVoteRes.QuitVotes++
	default:
		return nil, errors.Wrapf(types.ErrInvalidParam, "vote option=%d", voteProb.Vote)
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 首次进入投票期,即将提案金转入自治系统地址
	if cur.Status == auty.AutonomyStatusProposalItem {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, a.execaddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("votePropItem ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if cur.BoardVoteRes.TotalVotes != 0 && cur.BoardVoteRes.TotalVotes > cur.BoardVoteRes.QuitVotes &&
		cur.BoardVoteRes.ApproveVotes*100 >= (cur.BoardVoteRes.TotalVotes-cur.BoardVoteRes.QuitVotes)*cur.CurRule.BoardApproveRatio {
		cur.BoardVoteRes.Pass = true
		cur.PropItem.RealEndBlockHeight = a.height
	}

	key := propItemID(voteProb.ProposalID)
	cur.Status = auty.AutonomyStatusVotePropItem
	if cur.BoardVoteRes.Pass {
		cur.Status = auty.AutonomyStatusTmintPropItem
	}
	kv = append(kv, &types.KeyValue{Key: key, Value: types.Encode(cur)})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: boardVotesRecord(voteProb.ProposalID), Value: types.Encode(votes)})

	ty := auty.TyLogVotePropItem
	if cur.BoardVoteRes.Pass {
		ty = auty.TyLogTmintPropItem
	}
	receiptLog := getItemReceiptLog(pre, cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) tmintPropItem(tmintProb *auty.TerminateProposalItem) (*types.Receipt, error) {
	cur, err := a.getProposalItem(tmintProb.ProposalID)
	if err != nil {
		alog.Error("tmintPropItem ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalItem failed",
			tmintProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "get item id=%s", tmintProb.ProposalID)
	}
	pre := copyAutonomyProposalItem(cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusTmintPropItem ||
		cur.Status == auty.AutonomyStatusRvkPropItem {
		err := auty.ErrProposalStatus
		alog.Error("tmintPropItem ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			tmintProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "cur status=%d", cur.Status)
	}

	// 董事会投票期间不能终止
	end := cur.GetPropItem().EndBlockHeight
	if !cur.BoardVoteRes.Pass && a.height <= end {
		err := auty.ErrTerminatePeriod
		alog.Error("tmintPropItem ", "addr", a.fromaddr, "status", cur.Status, "height", a.height,
			"in board vote period can not terminate", tmintProb.ProposalID, "err", err)
		return nil, errors.Wrapf(err, "vote period not should be terminated")
	}

	if cur.BoardVoteRes.TotalVotes != 0 && cur.BoardVoteRes.TotalVotes > cur.BoardVoteRes.QuitVotes &&
		cur.BoardVoteRes.ApproveVotes*100 >= (cur.BoardVoteRes.TotalVotes-cur.BoardVoteRes.QuitVotes)*cur.CurRule.BoardApproveRatio {
		cur.BoardVoteRes.Pass = true
	} else {
		cur.BoardVoteRes.Pass = false
	}

	cur.PropItem.RealEndBlockHeight = a.height

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 如果为提案状态，则判断是否需要扣除提案费
	if cur.Status == auty.AutonomyStatusProposalItem && a.height > end {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, a.execaddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("tmintPropItem ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	cur.Status = auty.AutonomyStatusTmintPropItem

	kv = append(kv, &types.KeyValue{Key: propItemID(tmintProb.ProposalID), Value: types.Encode(cur)})

	receiptLog := getItemReceiptLog(pre, cur, auty.TyLogTmintPropItem)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) getProposalItem(ID string) (*auty.AutonomyProposalItem, error) {
	value, err := a.db.Get(propItemID(ID))
	if err != nil {
		return nil, err
	}
	cur := &auty.AutonomyProposalItem{}
	err = types.Decode(value, cur)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

// getItemReceiptLog 根据提案信息获取log
// 状态变化：
func getItemReceiptLog(pre, cur *auty.AutonomyProposalItem, ty int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = ty
	r := &auty.ReceiptProposalItem{Prev: pre, Current: cur}
	log.Log = types.Encode(r)
	return log
}

func copyAutonomyProposalItem(cur *auty.AutonomyProposalItem) *auty.AutonomyProposalItem {
	if cur == nil {
		return nil
	}
	newAut := *cur
	if cur.PropItem != nil {
		newItem := *cur.GetPropItem()
		newAut.PropItem = &newItem
	}
	if cur.CurRule != nil {
		newRule := *cur.GetCurRule()
		newAut.CurRule = &newRule
	}
	if len(cur.Boards) > 0 {
		newAut.Boards = make([]string, len(cur.Boards))
		copy(newAut.Boards, cur.Boards)
	}
	if cur.BoardVoteRes != nil {
		newRes := *cur.GetBoardVoteRes()
		newAut.BoardVoteRes = &newRes
	}

	return &newAut
}
