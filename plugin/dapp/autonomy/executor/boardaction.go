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

	"github.com/33cn/chain33/common/address"
)

const (
	minBoards          = 3
	maxBoards          = 30
	publicPeriod      int32 = 120960             // 公示一周时间，以区块高度计算
	ticketPrice        = types.Coin * 3000       // 单张票价
	largeProjectAmount = types.Coin * 100 *10000 // 重大项目公示金额阈值
	proposalAmount     = types.Coin * 1000       // 创建者消耗金额
	boardAttendRatio  int32 = 66 // 董事会成员参与率，以%计，可修改
	boardApproveRatio int32 = 66 // 董事会成员赞成率，以%计，可修改
	pubAttendRatio    int32 = 50 // 全体持票人参与率，以%计
	pubApproveRatio   int32 = 50 // 全体持票人赞成率，以%计
	pubOpposeRatio    int32 = 33 // 全体持票人否决率，以%计
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
		alog.Error("propBoard ", "proposal boards number is invaild", len(prob.Boards))
		return  nil, types.ErrInvalidParam
	}

	if prob.StartBlockHeight < a.height || prob.EndBlockHeight < a.height {
		alog.Error("propBoard height invaild", "StartBlockHeight", prob.StartBlockHeight, "EndBlockHeight",
			prob.EndBlockHeight, "height", a.height)
		return  nil, types.ErrInvalidParam
	}

	for _, board := range prob.Boards {
		if err := address.CheckAddress(board); err != nil {
			alog.Error("propBoard ", "addr", board, "check toAddr error", err)
			return  nil, types.ErrInvalidAddress
		}
	}

	// 获取当前生效提案规则
	rule, err := a.getActiveRule()
	if err != nil {
		alog.Error("propBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "getActiveRule failed", err)
		return nil, err
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
	cur, err := a.getProposalBoard(rvkProb.ProposalID)
	if err != nil {
		alog.Error("rvkPropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalBoard failed",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalBoard(cur)

	// 检查当前状态
	if cur.Status != auty.AutonomyStatusProposalBoard {
		err := auty.ErrProposalStatus
		alog.Error("rvkPropBoard ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			rvkProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropBoard().StartBlockHeight
	if a.height >= start {
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

	kv = append(kv, &types.KeyValue{Key: propBoardID(rvkProb.ProposalID), Value: types.Encode(cur)})

	getReceiptLog(pre, cur, auty.TyLogRvkPropBoard)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) votePropBoard(voteProb *auty.VoteProposalBoard) (*types.Receipt, error) {
	cur, err := a.getProposalBoard(voteProb.ProposalID)
	if err != nil {
		alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalBoard failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalBoard(cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusRvkPropBoard ||
		cur.Status == auty.AutonomyStatusTmintPropBoard {
		err := auty.ErrProposalStatus
		alog.Error("votePropBoard ", "addr", a.fromaddr, "status", cur.Status, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropBoard().StartBlockHeight
	end := cur.GetPropBoard().EndBlockHeight
	real := cur.GetPropBoard().RealEndBlockHeight
	if a.height < start || a.height > end || real != 0 {
		err := auty.ErrVotePeriod
		alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "ProposalID",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	// 检查是否已经参与投票
	votes, err := a.checkVotesRecord(votesRecord(voteProb.ProposalID))
	if err != nil {
		alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "checkVotesRecord failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}
	// 更新投票记录
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

	// 首次进入投票期,即将提案金转入自治系统地址
	if cur.Status == auty.AutonomyStatusProposalBoard {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("votePropBoard ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if cur.VoteResult.TotalVotes != 0 &&
		cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes != 0 &&
	    float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) / float32(cur.VoteResult.TotalVotes) >=  float32(pubAttendRatio)/100.0 &&
		float32(cur.VoteResult.ApproveVotes) / float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) >= float32(pubApproveRatio)/100.0 {
		cur.VoteResult.Pass = true
		cur.PropBoard.RealEndBlockHeight = a.height
	}

	key := propBoardID(voteProb.ProposalID)
	cur.Status = auty.AutonomyStatusVotePropBoard
	if cur.VoteResult.Pass {
		cur.Status = auty.AutonomyStatusTmintPropBoard
	}
	kv = append(kv, &types.KeyValue{Key: key, Value: types.Encode(cur)})

	// 更新VotesRecord
	kv = append(kv, &types.KeyValue{Key: votesRecord(voteProb.ProposalID), Value: types.Encode(votes)})

	// 更新当前具有权利的董事会成员
	if cur.VoteResult.Pass {
		kv = append(kv, &types.KeyValue{Key: activeBoardID(), Value:types.Encode(cur.PropBoard)})
	}

	ty := auty.TyLogVotePropBoard
	if cur.VoteResult.Pass {
		ty = auty.TyLogTmintPropBoard
	}
	receiptLog := getReceiptLog(pre, cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) tmintPropBoard(tmintProb *auty.TerminateProposalBoard) (*types.Receipt, error) {
	cur, err := a.getProposalBoard(tmintProb.ProposalID)
	if err != nil {
		alog.Error("tmintPropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "getProposalBoard failed",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}
	pre := copyAutonomyProposalBoard(cur)

	// 检查当前状态
	if cur.Status == auty.AutonomyStatusTmintPropBoard ||
		cur.Status == auty.AutonomyStatusRvkPropBoard {
		err := auty.ErrProposalStatus
		alog.Error("tmintPropBoard ", "addr", a.fromaddr, "status", cur.Status, "status is not match",
			tmintProb.ProposalID, "err", err)
		return nil, err
	}

	start := cur.GetPropBoard().StartBlockHeight
	end := cur.GetPropBoard().EndBlockHeight
	if a.height <= end && !cur.VoteResult.Pass {
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

	if float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) / float32(cur.VoteResult.TotalVotes) >=  float32(pubAttendRatio)/100.0 &&
		float32(cur.VoteResult.ApproveVotes) / float32(cur.VoteResult.ApproveVotes + cur.VoteResult.OpposeVotes) >= float32(pubApproveRatio)/100.0 {
		cur.VoteResult.Pass = true
	} else {
		cur.VoteResult.Pass = false
	}
	cur.PropBoard.RealEndBlockHeight = a.height

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 未进行投票情况下，符合提案关闭的也需要扣除提案费用
	if cur.Status == auty.AutonomyStatusProposalBoard {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, autonomyAddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	cur.Status = auty.AutonomyStatusTmintPropBoard

	kv = append(kv, &types.KeyValue{Key: propBoardID(tmintProb.ProposalID), Value: types.Encode(cur)})

	// 更新当前具有权利的董事会成员
	if cur.VoteResult.Pass {
		kv = append(kv, &types.KeyValue{Key: activeBoardID(), Value:types.Encode(cur.PropBoard)})
	}

	getReceiptLog(pre, cur, auty.TyLogTmintPropBoard)

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

func (a *action) getProposalBoard(ID string) (*auty.AutonomyProposalBoard, error) {
	value, err := a.db.Get(propBoardID(ID))
	if err != nil {
		return nil, err
	}
	cur := &auty.AutonomyProposalBoard{}
	err = types.Decode(value, cur)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

func (a *action) getActiveRule() (*auty.RuleConfig, error) {
	// 获取当前生效提案规则,并且将不修改的规则补齐
	rule := &auty.RuleConfig{}
	value, err := a.db.Get(activeRuleID())
	if err == nil {
		err = types.Decode(value, rule)
		if err != nil {
			return nil, err
		}
	} else {// 载入系统默认值
		rule.BoardAttendRatio   = boardAttendRatio
		rule.BoardApproveRatio  = boardApproveRatio
		rule.PubOpposeRatio     = pubOpposeRatio
		rule.ProposalAmount     = proposalAmount
		rule.LargeProjectAmount = largeProjectAmount
		rule.PublicPeriod       = publicPeriod
	}
	return rule, nil
}

func (a *action) checkVotesRecord(key []byte) (*auty.VotesRecord, error) {
	var votes auty.VotesRecord
	value, err := a.db.Get(key)
	if err == nil {
		err = types.Decode(value, &votes)
		if err != nil {
			return nil, err
		}
	}
	// 检查是否有重复
	for _, addr := range votes.Address {
		if addr == a.fromaddr {
			err := auty.ErrRepeatVoteAddr
			return nil, err
		}
	}
	return &votes, nil
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
	if cur == nil {
		return nil
	}
	newAut := *cur
	if cur.PropBoard != nil {
		newBoard := *cur.GetPropBoard()
		newAut.PropBoard = &newBoard
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

