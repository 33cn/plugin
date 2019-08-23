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
	maxBoardPeriodAmount = types.Coin * 10000 * 300 // 每个时期董事会审批最大额度300万
	boardPeriod          = 17280 * 30 * 1           // 时期为一个月
)

func (a *action) propProject(prob *auty.ProposalProject) (*types.Receipt, error) {
	if err := address.CheckAddress(prob.ToAddr); err != nil {
		alog.Error("propProject ", "addr", prob.ToAddr, "check toAddr error", err)
		return nil, types.ErrInvalidAddress
	}

	if prob.StartBlockHeight < a.height || prob.EndBlockHeight < a.height || prob.Amount <= 0 ||
		prob.StartBlockHeight+startEndBlockPeriod > prob.EndBlockHeight {
		alog.Error("propProject height or amount invaild", "StartBlockHeight", prob.StartBlockHeight, "EndBlockHeight",
			prob.EndBlockHeight, "height", a.height, "amount", prob.Amount)
		return nil, types.ErrInvalidParam
	}

	// 获取董事会成员
	pboard, err := a.getActiveBoard()
	if err != nil {
		alog.Error("propProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "get getActiveBoard failed", err)
		return nil, err
	}
	// 检查是否可以对已审批额度归0,如果可以则设置kv
	var kva *types.KeyValue
	if a.height > pboard.StartHeight+boardPeriod {
		pboard.StartHeight = a.height
		pboard.Amount = 0
		kva = &types.KeyValue{Key: activeBoardID(), Value: types.Encode(pboard)}
	}
	// 检查额度
	pass := a.checkPeriodAmount(pboard, prob.Amount)
	if !pass {
		err = auty.ErrNoPeriodAmount
		alog.Error("propProject ", "addr", a.fromaddr, "cumsum amount", pboard.Amount, "this period board have enough amount", err)
		return nil, err
	}
	// 获取当前生效提案规则
	rule, err := a.getActiveRule()
	if err != nil {
		alog.Error("propProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "getActiveRule failed", err)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 冻结提案金
	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, rule.ProposalAmount)
	if err != nil {
		alog.Error("propProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen proposal amount", rule.ProposalAmount, "error", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 冻结项目金
	receiptPrj, err := a.coinsAccount.ExecFrozen(autonomyFundAddr, a.execaddr, prob.Amount)
	if err != nil {
		alog.Error("propProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen project amount", prob.Amount, "error", err)
		return nil, err
	}
	logs = append(logs, receiptPrj.Logs...)
	kv = append(kv, receiptPrj.KV...)

	var isPubVote bool
	if prob.Amount >= rule.LargeProjectAmount {
		isPubVote = true
	}
	cur := &auty.AutonomyProposalProject{
		PropProject:  prob,
		CurRule:      rule,
		Boards:       pboard.Boards,
		BoardVoteRes: &auty.VoteResult{TotalVotes: int32(len(pboard.Boards))},
		PubVote:      &auty.PublicVote{Publicity: isPubVote},
		Status:       auty.AutonomyStatusProposalProject,
		Address:      a.fromaddr,
		Height:       a.height,
		Index:        a.index,
		ProposalID:   common.ToHex(a.txhash),
	}
	kv = append(kv, &types.KeyValue{Key: propProjectID(common.ToHex(a.txhash)), Value: types.Encode(cur)})
	if kva != nil {
		kv = append(kv, kva)
	}
	receiptLog := getProjectReceiptLog(nil, cur, auty.TyLogPropProject)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) rvkPropProject(rvkProb *auty.RevokeProposalProject) (*types.Receipt, error) {
	cur, err := a.getProposalProject(rvkProb.ProposalID)
	if err != nil {
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalProject failed",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalProject(cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalProject {
		err := auty.ErrProposalStatus
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropProject().StartBlockHeight
	if a.height >= start {
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

	// 解冻提案金
	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, cur.CurRule.ProposalAmount)
	if err != nil {
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive amount", cur.CurRule.ProposalAmount, "err", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 解冻项目金
	receiptPrj, err := a.coinsAccount.ExecActive(autonomyFundAddr, a.execaddr, cur.PropProject.Amount)
	if err != nil {
		alog.Error("rvkPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive project amount", cur.PropProject.Amount, "error", err)
		return nil, err
	}
	logs = append(logs, receiptPrj.Logs...)
	kv = append(kv, receiptPrj.KV...)

	cur.Status = auty.AutonomyStatusRvkPropProject

	kv = append(kv, &types.KeyValue{Key: propProjectID(rvkProb.ProposalID), Value: types.Encode(cur)})

	receiptLog := getProjectReceiptLog(pre, cur, auty.TyLogRvkPropProject)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) votePropProject(voteProb *auty.VoteProposalProject) (*types.Receipt, error) {
	cur, err := a.getProposalProject(voteProb.ProposalID)
	if err != nil {
		alog.Error("votePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalProject failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalProject(cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusRvkPropProject ||
		cur.Status == auty.AutonomyStatusPubVotePropProject ||
		cur.Status == auty.AutonomyStatusTmintPropProject {
		err := auty.ErrProposalStatus
		alog.Error("votePropProject ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropProject().StartBlockHeight
	end := cur.GetPropProject().EndBlockHeight
	real := cur.GetPropProject().RealEndBlockHeight
	if a.height < start || a.height > end || real != 0 {
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
			break
		}
	}
	if !isBoard {
		err = auty.ErrNoActiveBoard
		alog.Error("votePropProject ", "addr", a.fromaddr, "this addr is not active board member",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 检查是否已经参与投票
	votes, err := a.checkVotesRecord([]string{a.fromaddr}, boardVotesRecord(voteProb.ProposalID))
	if err != nil {
		alog.Error("votePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "checkVotesRecord boardVotesRecord failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 更新已经投票地址
	votes.Address = append(votes.Address, a.fromaddr)
	// 更新投票结果
	if voteProb.Approve {
		cur.BoardVoteRes.ApproveVotes++
	} else {
		cur.BoardVoteRes.OpposeVotes++
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 首次进入投票期,即将提案金转入自治系统地址
	if cur.Status == auty.AutonomyStatusProposalProject {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyFundAddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("votePropProject ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if cur.BoardVoteRes.TotalVotes != 0 &&
		float32(cur.BoardVoteRes.ApproveVotes)/float32(cur.BoardVoteRes.TotalVotes) >= float32(cur.CurRule.BoardApproveRatio)/100.0 {
		cur.BoardVoteRes.Pass = true
		cur.PropProject.RealEndBlockHeight = a.height
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
			receipt, err := a.coinsAccount.ExecTransferFrozen(autonomyFundAddr, cur.PropProject.ToAddr, a.execaddr, cur.PropProject.Amount)
			if err != nil {
				alog.Error("votePropProject ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen to contractor project amount fail", err)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
			// 需要更新该董事会的累计审批金
			pakv, err := a.updatePeriodAmount(cur.PropProject.Amount)
			if err != nil {
				alog.Error("votePropProject ", "addr", cur.Address, "execaddr", a.execaddr, "updatePeriodAmount fail", err)
				return nil, err
			}
			kv = append(kv, pakv)
		}
	}
	kv = append(kv, &types.KeyValue{Key: key, Value: types.Encode(cur)})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: boardVotesRecord(voteProb.ProposalID), Value: types.Encode(votes)})

	ty := auty.TyLogVotePropProject
	if cur.BoardVoteRes.Pass {
		if cur.PubVote.Publicity {
			ty = auty.TyLogPubVotePropProject
		} else {
			ty = auty.TyLogTmintPropProject
		}
	}
	receiptLog := getProjectReceiptLog(pre, cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) pubVotePropProject(voteProb *auty.PubVoteProposalProject) (*types.Receipt, error) {
	cur, err := a.getProposalProject(voteProb.ProposalID)
	if err != nil {
		alog.Error("pubVotePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalProject failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalProject(cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusPubVotePropProject {
		err := auty.ErrProposalStatus
		alog.Error("pubVotePropProject ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropProject().StartBlockHeight
	if a.height < start {
		err := auty.ErrVotePeriod
		alog.Error("pubVotePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	if len(voteProb.OriginAddr) > 0 {
		for _, board := range voteProb.OriginAddr {
			if err := address.CheckAddress(board); err != nil {
				alog.Error("pubVotePropProject ", "addr", board, "check toAddr error", err)
				return nil, types.ErrInvalidAddress
			}
		}
		// 挖矿地址验证
		addr, err := a.verifyMinerAddr(voteProb.OriginAddr, a.fromaddr)
		if err != nil {
			alog.Error("pubVotePropProject ", "from addr", a.fromaddr, "error addr", addr, "ProposalID",
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
		alog.Error("pubVotePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "checkVotesRecord failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	// 更新投票记录
	votes.Address = append(votes.Address, addrs...)

	if cur.GetPubVote().TotalVotes == 0 { //需要统计总票数
		vtCouts, err := a.getTotalVotes(start)
		if err != nil {
			return nil, err
		}
		cur.PubVote.TotalVotes = vtCouts
	}

	// 获取该地址票数
	vtCouts, err := a.batchGetAddressVotes(addrs, start)
	if err != nil {
		alog.Error("pubVotePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "batchGetAddressVotes failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	if voteProb.Oppose { //投反对票
		cur.PubVote.OpposeVotes += vtCouts
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if cur.PubVote.TotalVotes != 0 &&
		float32(cur.PubVote.OpposeVotes)/float32(cur.PubVote.TotalVotes) >= float32(cur.CurRule.PubOpposeRatio)/100.0 {

		cur.PubVote.PubPass = false
		cur.PropProject.RealEndBlockHeight = a.height
		// 解冻项目金
		receiptPrj, err := a.coinsAccount.ExecActive(autonomyFundAddr, a.execaddr, cur.PropProject.Amount)
		if err != nil {
			alog.Error("pubVotePropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive project amount", cur.PropProject.Amount, "error", err)
			return nil, err
		}
		logs = append(logs, receiptPrj.Logs...)
		kv = append(kv, receiptPrj.KV...)
		// 需要更新该董事会的累计审批金
		pakv, err := a.updatePeriodAmount(cur.PropProject.Amount)
		if err != nil {
			alog.Error("pubVotePropProject ", "addr", cur.Address, "execaddr", a.execaddr, "updatePeriodAmount fail", err)
			return nil, err
		}
		kv = append(kv, pakv)
	}

	key := propProjectID(voteProb.ProposalID)
	cur.Status = auty.AutonomyStatusPubVotePropProject
	ty := auty.TyLogPubVotePropProject
	if !cur.PubVote.PubPass {
		cur.Status = auty.AutonomyStatusTmintPropProject
		ty = auty.TyLogTmintPropProject
	}
	kv = append(kv, &types.KeyValue{Key: key, Value: types.Encode(cur)})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: votesRecord(voteProb.ProposalID), Value: types.Encode(votes)})

	receiptLog := getProjectReceiptLog(pre, cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) tmintPropProject(tmintProb *auty.TerminateProposalProject) (*types.Receipt, error) {
	cur, err := a.getProposalProject(tmintProb.ProposalID)
	if err != nil {
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalProject failed",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalProject(cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusTmintPropProject ||
		cur.Status == auty.AutonomyStatusRvkPropProject {
		err := auty.ErrProposalStatus
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}

	// 公示期间不能终止
	if cur.PubVote.Publicity && cur.PubVote.PubPass &&
		a.height <= cur.PropProject.RealEndBlockHeight+int64(cur.CurRule.PublicPeriod) {
		err := auty.ErrTerminatePeriod
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "status", cur.Status,
			"in publicity vote period can not terminate", tmintProb.ProposalID, "err", err)
		return nil, err
	}

	// 董事会投票期间不能终止
	start := cur.GetPropProject().StartBlockHeight
	end := cur.GetPropProject().EndBlockHeight
	if !cur.BoardVoteRes.Pass && a.height <= end {
		err := auty.ErrTerminatePeriod
		alog.Error("tmintPropProject ", "addr", a.fromaddr, "status", cur.Status, "height", a.height,
			"in board vote period can not terminate", tmintProb.ProposalID, "err", err)
		return nil, err
	}

	if cur.BoardVoteRes.TotalVotes != 0 &&
		float32(cur.BoardVoteRes.ApproveVotes)/float32(cur.BoardVoteRes.TotalVotes) >= float32(cur.CurRule.BoardApproveRatio)/100.0 {
		cur.BoardVoteRes.Pass = true
	} else {
		cur.BoardVoteRes.Pass = false
	}

	if cur.PubVote.Publicity {
		if cur.PubVote.TotalVotes == 0 { //需要统计总票数
			vtCouts, err := a.getTotalVotes(start)
			if err != nil {
				return nil, err
			}
			cur.PubVote.TotalVotes = vtCouts
		}
		if cur.PubVote.TotalVotes != 0 &&
			float32(cur.PubVote.OpposeVotes)/float32(cur.PubVote.TotalVotes) >= float32(cur.CurRule.PubOpposeRatio)/100.0 {
			cur.PubVote.PubPass = false
		}
	}

	cur.PropProject.RealEndBlockHeight = a.height

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 如果为提案状态，则判断是否需要扣除提案费
	if cur.Status == auty.AutonomyStatusProposalProject && a.height > end {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyFundAddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("tmintPropProject ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if (cur.PubVote.Publicity && cur.PubVote.PubPass) || // 需要公示且公示通过
		(!cur.PubVote.Publicity && cur.BoardVoteRes.Pass) { // 不需要公示且董事会通过
		// 提案通过，将工程金额从基金付款给承包商
		receipt, err := a.coinsAccount.ExecTransferFrozen(autonomyFundAddr, cur.PropProject.ToAddr, a.execaddr, cur.PropProject.Amount)
		if err != nil {
			alog.Error("tmintPropProject ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen to contractor project amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
		// 需要更新该董事会的累计审批金
		pakv, err := a.updatePeriodAmount(cur.PropProject.Amount)
		if err != nil {
			alog.Error("tmintPropProject ", "addr", cur.Address, "execaddr", a.execaddr, "updatePeriodAmount fail", err)
			return nil, err
		}
		kv = append(kv, pakv)
	} else {
		// 解冻项目金
		receiptPrj, err := a.coinsAccount.ExecActive(autonomyFundAddr, a.execaddr, cur.PropProject.Amount)
		if err != nil {
			alog.Error("tmintPropProject ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive project amount", cur.PropProject.Amount, "error", err)
			return nil, err
		}
		logs = append(logs, receiptPrj.Logs...)
		kv = append(kv, receiptPrj.KV...)
	}

	cur.Status = auty.AutonomyStatusTmintPropProject

	kv = append(kv, &types.KeyValue{Key: propProjectID(tmintProb.ProposalID), Value: types.Encode(cur)})

	receiptLog := getProjectReceiptLog(pre, cur, auty.TyLogTmintPropProject)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) getProposalProject(ID string) (*auty.AutonomyProposalProject, error) {
	value, err := a.db.Get(propProjectID(ID))
	if err != nil {
		return nil, err
	}
	cur := &auty.AutonomyProposalProject{}
	err = types.Decode(value, cur)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

func (a *action) getActiveBoard() (*auty.ActiveBoard, error) {
	value, err := a.db.Get(activeBoardID())
	if err != nil {
		return nil, err
	}
	pboard := &auty.ActiveBoard{}
	err = types.Decode(value, pboard)
	if err != nil {
		return nil, err
	}
	if len(pboard.Boards) > maxBoards || len(pboard.Boards) < minBoards {
		err = auty.ErrNoActiveBoard
		return nil, err
	}
	return pboard, nil
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
	if cur == nil {
		return nil
	}
	newAut := *cur
	if cur.PropProject != nil {
		newProject := *cur.GetPropProject()
		newAut.PropProject = &newProject
	}
	if cur.CurRule != nil {
		newRule := *cur.GetCurRule()
		newAut.CurRule = &newRule
	}
	if cur.BoardVoteRes != nil {
		newRes := *cur.GetBoardVoteRes()
		newAut.BoardVoteRes = &newRes
	}
	if cur.PubVote != nil {
		newPub := *cur.GetPubVote()
		newAut.PubVote = &newPub
	}
	return &newAut
}

func (a *action) checkPeriodAmount(act *auty.ActiveBoard, amount int64) bool {
	if act == nil {
		return false
	}
	if act.Amount+amount > maxBoardPeriodAmount {
		return false
	}
	return true
}

func (a *action) updatePeriodAmount(amount int64) (*types.KeyValue, error) {
	act, err := a.getActiveBoard()
	if err != nil {
		return nil, err
	}
	if a.height > act.StartHeight+boardPeriod {
		act.StartHeight = a.height
		act.Amount = 0
	}
	act.Amount += amount
	return &types.KeyValue{Key: activeBoardID(), Value: types.Encode(act)}, nil
}
