// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/client"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"

)

const (
	minBoards          = 3
	maxBoards          = 30
	lockAmount   int64 = types.Coin * 1000  // 创建者消耗金额
	ticketPrice        = types.Coin * 3000  // 单张票价
	participationRate int32 = 66 // 参与率以%计
	approveRate       int32 = 66 // 赞成率以%计
)

type action struct {
	api          client.QueueProtocolAPI
	coinsAccount *account.DB
	db           dbm.KV
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	index        int32
	execaddr     string
}

func newAction(a *Autonomy, tx *types.Transaction, index int32) *action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &action{a.GetAPI(), a.GetCoinsAccount(), a.GetStateDB(), hash, fromaddr,
		a.GetBlockTime(), a.GetHeight(), index, dapp.ExecAddress(string(tx.Execer))}
}

func (a *action) propBoard(prob *auty.ProposalBoard) (*types.Receipt, error) {
	if len(prob.Boards) > maxBoards || len(prob.Boards) < minBoards {
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
			alog.Error("propBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode ProposalRule failed", err)
			return nil, err
		}
	} else {// 载入系统默认值
		rule.BoardAttendProb = participationRate
		rule.BoardPassProb = approveRate
		rule.OpposeProb = opposeRate
		rule.ProposalAmount = lockAmount
		rule.PubAmountThreshold = largeAmount
	}

	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, rule.ProposalAmount)
	if err != nil {
		alog.Error("propBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen amount", rule.ProposalAmount)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur := &auty.AutonomyProposalBoard{
		PropBoard:prob,
		CurRule:rule,
		VoteResult: &auty.VoteResult{},
		Status: auty.AutonomyStatusProposalBoard,
		Address: a.fromaddr,
		Height: a.height,
		Index: a.index,
	}

	kv = append(kv, &types.KeyValue{Key: propBoardID(common.ToHex(a.txhash)), Value: types.Encode(cur)})

	receiptLog := getReceiptLog(nil, cur, auty.TyLogPropBoard)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) rvkPropBoard(rvkProb *auty.RevokeProposalBoard) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(propBoardID(rvkProb.ProposalID))
	if err != nil {
		alog.Error("rvkPropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "get ProposalBoard) failed",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}
	var cur auty.AutonomyProposalBoard
	err = types.Decode(value, &cur)
	if err != nil {
		alog.Error("rvkPropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode ProposalBoard failed",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalBoard(&cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalBoard {
		err := auty.ErrProposalStatus
		alog.Error("rvkPropBoard ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropBoard().StartBlockHeight
	if a.height > start {
		err := auty.ErrRevokeProposalPeriod
		alog.Error("rvkPropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	if a.fromaddr != cur.Address {
		err := auty.ErrRevokeProposalPower
		alog.Error("rvkPropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, cur.CurRule.ProposalAmount)
	if err != nil {
		alog.Error("rvkPropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecActive amount", cur.CurRule.ProposalAmount, "err", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur.Status = auty.AutonomyStatusRvkPropBoard

	kv = append(kv, &types.KeyValue{Key: propBoardID(rvkProb.ProposalID), Value: types.Encode(&cur)})

	getReceiptLog(pre, &cur, auty.TyLogRvkPropBoard)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) votePropBoard(voteProb *auty.VoteProposalBoard) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(propBoardID(voteProb.ProposalID))
	if err != nil {
		alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "get propBoardID failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	var cur auty.AutonomyProposalBoard
	err = types.Decode(value, &cur)
	if err != nil {
		alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode AutonomyProposalBoard failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalBoard(&cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalBoard && cur.Status != auty.AutonomyStatusVotePropBoard {
		err := auty.ErrProposalStatus
		alog.Error("votePropBoard ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropBoard().StartBlockHeight
	end := cur.GetPropBoard().EndBlockHeight
	real := cur.GetPropBoard().RealEndBlockHeight
	if start < a.height || end < a.height || (real != 0 && real < a.height) {
		err := auty.ErrVotePeriod
		alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 检查是否已经参与投票
	var votes auty.VotesRecord
	value, err = a.db.Get(VotesRecord(voteProb.ProposalID))
	if err == nil {
		err = types.Decode(value, &votes)
		if err != nil {
			alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode VotesRecord failed",
				voteProb.ProposalID, "err", err)
			return nil, err
		}
	}

	// 检查是否有重复
	for _, addr := range votes.Address {
		if addr == a.fromaddr {
			err := auty.ErrRepeatVoteAddr
			alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "repeat address GameID",
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
	    float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) / float32(cur.VoteResult.TotalVotes) >=  float32(cur.CurRule.BoardAttendProb)/100.0 &&
		float32(cur.VoteResult.ApproveVotes) / float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) >= float32(cur.CurRule.OpposeProb)/100.0 {
		cur.VoteResult.Pass = true
		cur.PropBoard.RealEndBlockHeight = a.height

		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("votePropBoard ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	key := propBoardID(voteProb.ProposalID)
	cur.Status = auty.AutonomyStatusVotePropBoard
	if cur.VoteResult.Pass {
		cur.Status = auty.AutonomyStatusTmintPropBoard
	}
	value = types.Encode(&cur)
	kv = append(kv, &types.KeyValue{Key: key, Value: value})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: VotesRecord(voteProb.ProposalID), Value: types.Encode(&votes)})

	// 更新当前具有权利的董事会成员
	if cur.VoteResult.Pass {
		kv = append(kv, &types.KeyValue{Key: activeBoardID(), Value:types.Encode(cur.PropBoard)})
	}

	ty := auty.TyLogVotePropBoard
	if cur.VoteResult.Pass {
		ty = auty.TyLogTmintPropBoard
	}
	receiptLog := getReceiptLog(pre, &cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) tmintPropBoard(tmintProb *auty.TerminateProposalBoard) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(propBoardID(tmintProb.ProposalID))
	if err != nil {
		alog.Error("tmintPropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "get propBoardID failed",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}
	var cur auty.AutonomyProposalBoard
	err = types.Decode(value, &cur)
	if err != nil {
		alog.Error("tmintPropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode AutonomyProposalBoard failed",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalBoard(&cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusTmintPropBoard {
		err := auty.ErrProposalStatus
		alog.Error("tmintPropBoard ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropBoard().StartBlockHeight
	end := cur.GetPropBoard().EndBlockHeight
	if a.height < end && !cur.VoteResult.Pass {
		err := auty.ErrTerminatePeriod
		alog.Error("tmintPropBoard ", "addr", a.fromaddr, "status", cur.Status, "height", a.height,
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

	if float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) / float32(cur.VoteResult.TotalVotes) >=  float32(cur.CurRule.BoardAttendProb)/100.0 &&
		float32(cur.VoteResult.ApproveVotes) / float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) >= float32(cur.CurRule.OpposeProb)/100.0 {
		cur.VoteResult.Pass = true
	} else {
		cur.VoteResult.Pass = false
	}
	cur.PropBoard.RealEndBlockHeight = a.height

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, cur.CurRule.ProposalAmount)
	if err != nil {
		alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	cur.Status = auty.AutonomyStatusTmintPropBoard

	kv = append(kv, &types.KeyValue{Key: propBoardID(tmintProb.ProposalID), Value: types.Encode(&cur)})

	// 更新当前具有权利的董事会成员
	if cur.VoteResult.Pass {
		kv = append(kv, &types.KeyValue{Key: activeBoardID(), Value:types.Encode(cur.PropBoard)})
	}

	getReceiptLog(pre, &cur, auty.TyLogTmintPropBoard)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) getStartHeightVoteAccount(addr string, height int64) (*types.Account, error) {
	param := &types.ReqBlocks{
		Start: height,
		End:height,
	}
	head, err := a.api.GetHeaders(param)
	if err != nil || len(head.Items) == 0 {
		alog.Error("GetStartVoteAccount ", "addr", addr, "height", height, "get headers failed", err)
		return nil, err
	}

	stateHash := common.ToHex(head.Items[0].StateHash)

	account, err := a.coinsAccount.GetBalance(a.api, &types.ReqBalance{
		Addresses: []string{addr},
		StateHash: stateHash,
	})
	if err != nil || len(account) == 0 {
		alog.Error("GetStartVoteAccount ", "addr", addr, "height", height, "GetBalance failed", err)
		return nil, err
	}
	return account[0], nil
}

// getReceiptLog 根据提案信息获取log
// 状态变化：
func getReceiptLog(pre, cur *auty.AutonomyProposalBoard, ty int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = ty
	r := &auty.ReceiptProposalBoard{Prev: pre, Current: cur}
	log.Log = types.Encode(r)
	return log
}

func copyAutonomyProposalBoard(cur *auty.AutonomyProposalBoard) *auty.AutonomyProposalBoard {
	newAut := *cur
	newBoard := *cur.GetPropBoard()
	newRes := *cur.GetVoteResult()
	newAut.PropBoard = &newBoard
	newAut.VoteResult = &newRes
	return &newAut
}

