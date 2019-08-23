// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

func (a *action) propChange(prob *auty.ProposalChange) (*types.Receipt, error) {
	//如果全小于等于0,则说明该提案规则参数不正确
	if prob == nil || len(prob.Changes) == 0 {
		alog.Error("propChange ", "ProposalChange ChangeCfg invaild or have no modify param", prob)
		return nil, types.ErrInvalidParam
	}
	if prob.StartBlockHeight < a.height || prob.EndBlockHeight < a.height {
		alog.Error("propChange height invaild", "StartBlockHeight", prob.StartBlockHeight, "EndBlockHeight",
			prob.EndBlockHeight, "height", a.height)
		return nil, types.ErrInvalidParam
	}

	act, err := a.getActiveBoard()
	if err != nil {
		alog.Error("propChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "getActiveBoard failed", err)
		return nil, err
	}
	// 检查是否符合提案修改
	new, err := a.checkChangeable(act, prob.Changes)
	if err != nil {
		alog.Error("propChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "checkChangeable failed", err)
		return nil, err
	}

	// 获取当前生效提案规则,并且将不修改的规则补齐
	rule, err := a.getActiveRule()
	if err != nil {
		alog.Error("propChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "getActiveRule failed", err)
		return nil, err
	}

	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, rule.ProposalAmount)
	if err != nil {
		alog.Error("propChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen amount", rule.ProposalAmount)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur := &auty.AutonomyProposalChange{
		PropChange: prob,
		CurRule:    rule,
		Board:      new,
		VoteResult: &auty.VoteResult{TotalVotes: int32(len(act.Boards))},
		Status:     auty.AutonomyStatusProposalChange,
		Address:    a.fromaddr,
		Height:     a.height,
		Index:      a.index,
		ProposalID: common.ToHex(a.txhash),
	}

	key := propChangeID(common.ToHex(a.txhash))
	value := types.Encode(cur)
	kv = append(kv, &types.KeyValue{Key: key, Value: value})

	receiptLog := getChangeReceiptLog(nil, cur, auty.TyLogPropChange)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) rvkPropChange(rvkProb *auty.RevokeProposalChange) (*types.Receipt, error) {
	cur, err := a.getProposalChange(rvkProb.ProposalID)
	if err != nil {
		alog.Error("rvkPropChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalChange failed",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalChange(cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalChange {
		err := auty.ErrProposalStatus
		alog.Error("rvkPropChange ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropChange().StartBlockHeight
	if a.height >= start {
		err := auty.ErrRevokeProposalPeriod
		alog.Error("rvkPropChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	if a.fromaddr != cur.Address {
		err := auty.ErrRevokeProposalPower
		alog.Error("rvkPropChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, cur.CurRule.ProposalAmount)
	if err != nil {
		alog.Error("rvkPropChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive amount", cur.CurRule.ProposalAmount, "err", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur.Status = auty.AutonomyStatusRvkPropChange

	kv = append(kv, &types.KeyValue{Key: propChangeID(rvkProb.ProposalID), Value: types.Encode(cur)})

	receiptLog := getChangeReceiptLog(pre, cur, auty.TyLogRvkPropChange)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) votePropChange(voteProb *auty.VoteProposalChange) (*types.Receipt, error) {
	cur, err := a.getProposalChange(voteProb.ProposalID)
	if err != nil {
		alog.Error("votePropChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalChange failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalChange(cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusRvkPropChange ||
		cur.Status == auty.AutonomyStatusTmintPropChange {
		err := auty.ErrProposalStatus
		alog.Error("votePropChange ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropChange().StartBlockHeight
	end := cur.GetPropChange().EndBlockHeight
	real := cur.GetPropChange().RealEndBlockHeight
	if a.height < start || a.height > end || real != 0 {
		err := auty.ErrVotePeriod
		alog.Error("votePropChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 检查是否已经参与投票
	votes, err := a.checkVotesRecord([]string{a.fromaddr}, votesRecord(voteProb.ProposalID))
	if err != nil {
		alog.Error("votePropChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "checkVotesRecord failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	// 更新投票记录
	votes.Address = append(votes.Address, a.fromaddr)

	if voteProb.Approve {
		cur.VoteResult.ApproveVotes++
	} else {
		cur.VoteResult.OpposeVotes++
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 首次进入投票期,即将提案金转入自治系统地址
	if cur.Status == auty.AutonomyStatusProposalChange {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyFundAddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("votePropChange ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if cur.VoteResult.TotalVotes != 0 &&
		float32(cur.VoteResult.ApproveVotes)/float32(cur.VoteResult.TotalVotes) > float32(cur.CurRule.BoardApproveRatio)/100.0 {
		cur.VoteResult.Pass = true
		cur.PropChange.RealEndBlockHeight = a.height
	}

	key := propChangeID(voteProb.ProposalID)
	cur.Status = auty.AutonomyStatusVotePropChange
	if cur.VoteResult.Pass {
		cur.Status = auty.AutonomyStatusTmintPropChange
	}
	kv = append(kv, &types.KeyValue{Key: key, Value: types.Encode(cur)})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: votesRecord(voteProb.ProposalID), Value: types.Encode(votes)})

	// 更新activeBoard
	if cur.VoteResult.Pass {
		kv = append(kv, &types.KeyValue{Key: activeBoardID(), Value: types.Encode(cur.Board)})
	}

	ty := auty.TyLogVotePropChange
	if cur.VoteResult.Pass {
		ty = auty.TyLogTmintPropChange
	}
	receiptLog := getChangeReceiptLog(pre, cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) tmintPropChange(tmintProb *auty.TerminateProposalChange) (*types.Receipt, error) {
	cur, err := a.getProposalChange(tmintProb.ProposalID)
	if err != nil {
		alog.Error("tmintPropChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalChange failed",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}

	pre := copyAutonomyProposalChange(cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusTmintPropChange ||
		cur.Status == auty.AutonomyStatusRvkPropChange {
		err := auty.ErrProposalStatus
		alog.Error("tmintPropChange ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}

	end := cur.GetPropChange().EndBlockHeight
	if a.height < end && !cur.VoteResult.Pass {
		err := auty.ErrTerminatePeriod
		alog.Error("tmintPropChange ", "addr", a.fromaddr, "status", cur.Status, "height", a.height,
			"in vote period can not terminate", tmintProb.ProposalID, "err", err)
		return nil, err
	}

	if cur.VoteResult.TotalVotes != 0 &&
		float32(cur.VoteResult.ApproveVotes)/float32(cur.VoteResult.TotalVotes) > float32(cur.CurRule.BoardApproveRatio)/100.0 {
		cur.VoteResult.Pass = true
	} else {
		cur.VoteResult.Pass = false
	}
	cur.PropChange.RealEndBlockHeight = a.height

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 未进行投票情况下，符合提案关闭的也需要扣除提案费用
	if cur.Status == auty.AutonomyStatusProposalChange {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyFundAddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("votePropChange ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)

	}

	cur.Status = auty.AutonomyStatusTmintPropChange

	kv = append(kv, &types.KeyValue{Key: propChangeID(tmintProb.ProposalID), Value: types.Encode(cur)})

	// 更新系统规则
	if cur.VoteResult.Pass {
		kv = append(kv, &types.KeyValue{Key: activeBoardID(), Value: types.Encode(cur.Board)})
	}
	receiptLog := getChangeReceiptLog(pre, cur, auty.TyLogTmintPropChange)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) getProposalChange(ID string) (*auty.AutonomyProposalChange, error) {
	value, err := a.db.Get(propChangeID(ID))
	if err != nil {
		return nil, err
	}
	cur := &auty.AutonomyProposalChange{}
	err = types.Decode(value, cur)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

func (a *action) checkChangeable(act *auty.ActiveBoard, change []*auty.Change) (*auty.ActiveBoard, error) {
	mpBd := make(map[string]struct{})
	mpRbd := make(map[string]struct{})
	for _, b := range act.Boards {
		mpBd[b] = struct{}{}
	}
	for _, b := range act.Revboards {
		mpRbd[b] = struct{}{}
	}
	for _, ch := range change {
		if ch.Cancel {
			if _, ok := mpBd[ch.Addr]; !ok {
				return nil, auty.ErrChangeBoardAddr
			}
			// 将删除的加入对端
			delete(mpBd, ch.Addr)
			mpRbd[ch.Addr] = struct{}{}
		} else {
			if _, ok := mpRbd[ch.Addr]; !ok {
				return nil, auty.ErrChangeBoardAddr
			}
			// 将删除的加入对端
			delete(mpRbd, ch.Addr)
			mpBd[ch.Addr] = struct{}{}
		}
	}
	if len(mpBd) > maxBoards || len(mpBd) < minBoards {
		return nil, auty.ErrBoardNumber
	}
	new := &auty.ActiveBoard{
		Amount:      act.Amount,
		StartHeight: act.StartHeight,
	}
	for k := range mpBd {
		new.Boards = append(new.Boards, k)
	}
	for k := range mpRbd {
		new.Revboards = append(new.Revboards, k)
	}
	return new, nil
}

// getReceiptLog 根据提案信息获取log
// 状态变化：
func getChangeReceiptLog(pre, cur *auty.AutonomyProposalChange, ty int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = ty
	r := &auty.ReceiptProposalChange{Prev: pre, Current: cur}
	log.Log = types.Encode(r)
	return log
}

func copyAutonomyProposalChange(cur *auty.AutonomyProposalChange) *auty.AutonomyProposalChange {
	if cur == nil {
		return nil
	}
	newAut := *cur
	if cur.PropChange != nil {
		newPropChange := *cur.PropChange
		newAut.PropChange = &newPropChange
		if cur.PropChange.Changes != nil {
			newAut.PropChange.Changes = make([]*auty.Change, len(cur.PropChange.Changes))
			chs := cur.PropChange.Changes
			for i, ch := range chs {
				newch := *ch
				newAut.PropChange.Changes[i] = &newch
			}
		}
	}
	if cur.CurRule != nil {
		newChange := *cur.GetCurRule()
		newAut.CurRule = &newChange
	}
	if cur.Board != nil {
		newBoard := *cur.GetBoard()
		newBoard.Boards = make([]string, len(cur.Board.Boards))
		copy(newBoard.Boards, cur.Board.Boards)
		newBoard.Revboards = make([]string, len(cur.Board.Revboards))
		copy(newBoard.Revboards, cur.Board.Revboards)
		newAut.Board = &newBoard
	}
	if cur.VoteResult != nil {
		newRes := *cur.GetVoteResult()
		newAut.VoteResult = &newRes
	}
	return &newAut
}
