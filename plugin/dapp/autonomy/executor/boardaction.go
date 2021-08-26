// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/common/address"
	ticket "github.com/33cn/plugin/plugin/dapp/ticket/executor"
	ticketTy "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

const (
	minBoards                 = 20
	maxBoards                 = 40
	publicPeriod        int32 = 17280 * 7   // 公示一周时间，以区块高度计算
	ticketPrice               = 3000        // 单张票价
	largeProjectAmount        = 100 * 10000 // 重大项目公示金额阈值
	proposalAmount            = 500         // 创建者消耗金额
	boardApproveRatio   int32 = 51          // 董事会成员赞成率，以%计，可修改
	pubAttendRatio      int32 = 75          // 全体持票人参与率，以%计
	pubApproveRatio     int32 = 66          // 全体持票人赞成率，以%计
	pubOpposeRatio      int32 = 33          // 全体持票人否决率，以%计
	startEndBlockPeriod       = 720         // 提案开始结束最小周期
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

func (a *action) getPropBoard(prob *auty.ProposalBoard) (*auty.ActiveBoard, error) {
	mpBd, err := filterPropBoard(prob.Boards)
	if err != nil {
		return nil, err
	}

	switch prob.BoardUpdate {
	case auty.BoardUpdate_WHOLE:
		return &auty.ActiveBoard{Boards: prob.Boards}, nil
	case auty.BoardUpdate_ADD:
		return a.addPropBoard(prob, mpBd)
	case auty.BoardUpdate_DEL:
		return a.delPropBoard(prob, mpBd)
	default:
		return nil, errors.Wrapf(types.ErrInvalidParam, "board update param=%d", prob.BoardUpdate)
	}

}

func (a *action) getOldPropBoard(prob *auty.ProposalBoard) (*auty.ActiveBoard, error) {
	mpBd, err := filterPropBoard(prob.Boards)
	if err != nil {
		return nil, err
	}

	//replace all
	if !prob.Update {
		return &auty.ActiveBoard{
			Boards: prob.Boards,
		}, nil
	}

	// only add member
	return a.addPropBoard(prob, mpBd)
}

func filterPropBoard(boards []string) (map[string]struct{}, error) {
	mpBd := make(map[string]struct{})
	for _, board := range boards {
		if err := address.CheckAddress(board); err != nil {
			return nil, errors.Wrapf(types.ErrInvalidAddress, "addr=%s", board)
		}
		// 提案board重复地址去重复
		if _, ok := mpBd[board]; ok {
			return nil, errors.Wrapf(auty.ErrRepeatAddr, "propBoard addr=%s repeated", board)
		}
		mpBd[board] = struct{}{}
	}
	return mpBd, nil
}

func (a *action) propBoard(prob *auty.ProposalBoard) (*types.Receipt, error) {
	if prob.StartBlockHeight < a.height || prob.EndBlockHeight < a.height ||
		prob.StartBlockHeight+startEndBlockPeriod > prob.EndBlockHeight {
		alog.Error("propBoard height invaild", "StartBlockHeight", prob.StartBlockHeight, "EndBlockHeight",
			prob.EndBlockHeight, "height", a.height)
		return nil, auty.ErrSetBlockHeight
	}
	if len(prob.Boards) == 0 {
		alog.Error("propBoard ", "proposal boards number is zero", len(prob.Boards))
		return nil, auty.ErrBoardNumber
	}

	var act *auty.ActiveBoard
	var err error
	cfg := a.api.GetConfig()
	if cfg.IsDappFork(a.height, auty.AutonomyX, auty.ForkAutonomyDelRule) {
		act, err = a.getPropBoard(prob)
	} else {
		act, err = a.getOldPropBoard(prob)
	}

	if len(act.Boards) > maxBoards || len(act.Boards) < minBoards {
		alog.Error("propBoard ", "proposal boards number is invaild", len(prob.Boards))
		return nil, auty.ErrBoardNumber
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
		PropBoard:  prob,
		CurRule:    rule,
		Board:      act,
		VoteResult: &auty.VoteResult{},
		Status:     auty.AutonomyStatusProposalBoard,
		Address:    a.fromaddr,
		Height:     a.height,
		Index:      a.index,
		ProposalID: common.ToHex(a.txhash),
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

	receiptLog := getReceiptLog(pre, cur, auty.TyLogRvkPropBoard)
	logs = append(logs, receiptLog)

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

	if len(voteProb.OriginAddr) > 0 {
		for _, board := range voteProb.OriginAddr {
			if err := address.CheckAddress(board); err != nil {
				alog.Error("votePropBoard ", "addr", board, "check toAddr error", err)
				return nil, types.ErrInvalidAddress
			}
		}
		// 挖矿地址验证
		addr, err := a.verifyMinerAddr(voteProb.OriginAddr, a.fromaddr)
		if err != nil {
			alog.Error("votePropBoard ", "from addr", a.fromaddr, "error addr", addr, "ProposalID",
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
		alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "checkVotesRecord failed",
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

	vtCouts, err := a.batchGetAddressVotes(addrs, start)
	if err != nil {
		alog.Error("votePropBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "batchGetAddressVotes failed",
			voteProb.ProposalID, "err", err)
		return nil, err
	}

	cfg := a.api.GetConfig()
	//fork之后增加了弃权选项
	if cfg.IsDappFork(a.height, auty.AutonomyX, auty.ForkAutonomyDelRule) {
		switch voteProb.VoteOption {
		case auty.AutonomyVoteOption_APPROVE:
			cur.VoteResult.ApproveVotes += vtCouts
		case auty.AutonomyVoteOption_OPPOSE:
			cur.VoteResult.OpposeVotes += vtCouts
		case auty.AutonomyVoteOption_QUIT:
			cur.VoteResult.QuitVotes += vtCouts
		default:
			return nil, errors.Wrapf(types.ErrInvalidParam, "vote option=%d", voteProb.VoteOption)

		}
	} else {
		if voteProb.Approve {
			cur.VoteResult.ApproveVotes += vtCouts
		} else {
			cur.VoteResult.OpposeVotes += vtCouts
		}
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 首次进入投票期,即将提案金转入自治系统地址
	if cur.Status == auty.AutonomyStatusProposalBoard {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, a.execaddr, a.execaddr, cur.CurRule.ProposalAmount)
		if err != nil {
			alog.Error("votePropBoard ", "addr", cur.Address, "execaddr", a.execaddr, "ExecTransferFrozen amount fail", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if cfg.IsDappFork(a.height, auty.AutonomyX, auty.ForkAutonomyDelRule) {
		if isApproved(cur.VoteResult.TotalVotes, cur.VoteResult.ApproveVotes, cur.VoteResult.OpposeVotes, cur.VoteResult.QuitVotes,
			cur.CurRule.PubAttendRatio, cur.CurRule.PubApproveRatio) {
			cur.VoteResult.Pass = true
			cur.PropBoard.RealEndBlockHeight = a.height
		}
	} else {
		if cur.VoteResult.TotalVotes != 0 &&
			cur.VoteResult.ApproveVotes+cur.VoteResult.OpposeVotes != 0 &&
			float32(cur.VoteResult.ApproveVotes+cur.VoteResult.OpposeVotes)/float32(cur.VoteResult.TotalVotes) > float32(pubAttendRatio)/100.0 &&
			float32(cur.VoteResult.ApproveVotes)/float32(cur.VoteResult.ApproveVotes+cur.VoteResult.OpposeVotes) > float32(pubApproveRatio)/100.0 {
			cur.VoteResult.Pass = true
			cur.PropBoard.RealEndBlockHeight = a.height
		}
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
		if a.api.GetConfig().IsDappFork(a.height, auty.AutonomyX, auty.ForkAutonomyDelRule) {
			if cur.PropBoard.BoardUpdate == auty.BoardUpdate_WHOLE {
				cur.Board.StartHeight = a.height
			}
		} else {
			if !cur.PropBoard.Update { // 非update才进行高度重写
				cur.Board.StartHeight = a.height
			}
		}
		kv = append(kv, &types.KeyValue{Key: activeBoardID(), Value: types.Encode(cur.Board)})
	}

	ty := auty.TyLogVotePropBoard
	if cur.VoteResult.Pass {
		ty = auty.TyLogTmintPropBoard
	}
	receiptLog := getReceiptLog(pre, cur, int32(ty))
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//统计参与率的时候，计算弃权票，但是统计赞成率的时候，忽略弃权票。比如10票，4票赞成，3票反对，2票弃权，那么参与率是 90%， 赞成 4/7 反对 3/7
func isApproved(totalVotes, approveVotes, opposeVotes, quitVotes, attendRation, approveRatio int32) bool {
	if attendRation <= 0 {
		attendRation = pubAttendRatio
	}
	if approveRatio <= 0 {
		approveRatio = pubApproveRatio
	}
	//参与率计算弃权票
	attendVotes := approveVotes + opposeVotes + quitVotes
	//赞成率，忽略弃权票
	validVotes := approveVotes + opposeVotes
	if totalVotes != 0 && attendVotes != 0 &&
		attendVotes*100 > attendRation*totalVotes &&
		approveVotes*100 > approveRatio*validVotes {
		return true
	}
	return false
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
		vtCouts, err := a.getTotalVotes(start)
		if err != nil {
			return nil, err
		}
		cur.VoteResult.TotalVotes = vtCouts
	}

	if a.api.GetConfig().IsDappFork(a.height, auty.AutonomyX, auty.ForkAutonomyDelRule) {
		cur.VoteResult.Pass = isApproved(cur.VoteResult.TotalVotes, cur.VoteResult.ApproveVotes, cur.VoteResult.OpposeVotes, cur.VoteResult.QuitVotes,
			cur.CurRule.PubAttendRatio, cur.CurRule.PubApproveRatio)
	} else {
		if float32(cur.VoteResult.ApproveVotes+cur.VoteResult.OpposeVotes)/float32(cur.VoteResult.TotalVotes) > float32(pubAttendRatio)/100.0 &&
			float32(cur.VoteResult.ApproveVotes)/float32(cur.VoteResult.ApproveVotes+cur.VoteResult.OpposeVotes) > float32(pubApproveRatio)/100.0 {
			cur.VoteResult.Pass = true
		} else {
			cur.VoteResult.Pass = false
		}
	}
	cur.PropBoard.RealEndBlockHeight = a.height

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 未进行投票情况下，符合提案关闭的也需要扣除提案费用
	if cur.Status == auty.AutonomyStatusProposalBoard {
		receipt, err := a.coinsAccount.ExecTransferFrozen(cur.Address, a.execaddr, a.execaddr, cur.CurRule.ProposalAmount)
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
		if a.api.GetConfig().IsDappFork(a.height, auty.AutonomyX, auty.ForkAutonomyDelRule) {
			if cur.PropBoard.BoardUpdate == auty.BoardUpdate_WHOLE {
				cur.Board.StartHeight = a.height
			}
		} else {
			if !cur.PropBoard.Update { // 非update才进行高度重写
				cur.Board.StartHeight = a.height
			}
		}
		kv = append(kv, &types.KeyValue{Key: activeBoardID(), Value: types.Encode(cur.Board)})
	}

	receiptLog := getReceiptLog(pre, cur, auty.TyLogTmintPropBoard)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (a *action) getTotalVotes(height int64) (int32, error) {
	addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
	if subcfg.Total != "" {
		addr = subcfg.Total
	}
	account, err := a.getStartHeightVoteAccount(addr, "", height)
	if err != nil {
		return 0, err
	}
	return int32(account.Balance / (ticketPrice * a.api.GetConfig().GetCoinPrecision())), nil
}

func (a *action) verifyMinerAddr(addrs []string, bindAddr string) (string, error) {
	// 验证绑定关系与重复地址
	mp := make(map[string]struct{})
	for _, addr := range addrs {
		value, err := a.db.Get(ticket.BindKey(addr))
		if err != nil {
			return addr, auty.ErrMinerAddr
		}
		tkBind := &ticketTy.TicketBind{}
		err = types.Decode(value, tkBind)
		if err != nil || tkBind.MinerAddress != bindAddr {
			return addr, auty.ErrBindAddr
		}
		if _, ok := mp[addr]; ok {
			return addr, auty.ErrRepeatAddr
		}
		mp[addr] = struct{}{}
	}
	return "", nil
}

func (a *action) batchGetAddressVotes(addrs []string, height int64) (int32, error) {
	total := int32(0)
	for _, addr := range addrs {
		count, err := a.getAddressVotes(addr, height)
		if err != nil {
			return 0, err
		}
		total += count
	}
	return total, nil
}

func (a *action) getAddressVotes(addr string, height int64) (int32, error) {
	account, err := a.getStartHeightVoteAccount(addr, ticketName, height)
	if err != nil {
		return 0, err
	}
	amount := account.Frozen
	if subcfg.UseBalance {
		amount = account.Balance
	}
	return int32(amount / (ticketPrice * a.api.GetConfig().GetCoinPrecision())), nil
}

func (a *action) getStartHeightVoteAccount(addr, execer string, height int64) (*types.Account, error) {
	param := &types.ReqBlocks{
		Start: height,
		End:   height,
	}
	head, err := a.api.GetHeaders(param)
	if err != nil || len(head.Items) == 0 {
		alog.Error("GetStartVoteAccount ", "addr", addr, "height", height, "get headers failed", err)
		return nil, err
	}

	stateHash := common.ToHex(head.Items[0].StateHash)

	account, err := a.coinsAccount.GetBalance(a.api, &types.ReqBalance{
		Addresses: []string{addr},
		Execer:    execer,
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
	cfg := a.api.GetConfig()
	// 获取当前生效提案规则,并且将不修改的规则补齐
	rule := &auty.RuleConfig{}
	value, err := a.db.Get(activeRuleID())
	if err == nil {
		err = types.Decode(value, rule)
		if err != nil {
			return nil, err
		}
		//在fork之前可能有修改了规则，但是这两个值没有修改到
		if cfg.IsDappFork(a.height, auty.AutonomyX, auty.ForkAutonomyDelRule) {
			if rule.PubApproveRatio <= 0 {
				rule.PubApproveRatio = pubApproveRatio
			}
			if rule.PubAttendRatio <= 0 {
				rule.PubAttendRatio = pubAttendRatio
			}
		}

	} else { // 载入系统默认值
		rule.BoardApproveRatio = boardApproveRatio
		rule.PubOpposeRatio = pubOpposeRatio
		rule.ProposalAmount = proposalAmount * cfg.GetCoinPrecision()
		rule.LargeProjectAmount = largeProjectAmount * cfg.GetCoinPrecision()
		rule.PublicPeriod = publicPeriod

		if cfg.IsDappFork(a.height, auty.AutonomyX, auty.ForkAutonomyDelRule) {
			rule.PubAttendRatio = pubAttendRatio
			rule.PubApproveRatio = pubApproveRatio
		}

	}
	return rule, nil
}

func (a *action) checkVotesRecord(addrs []string, key []byte) (*auty.VotesRecord, error) {
	var votes auty.VotesRecord
	value, err := a.db.Get(key)
	if err == nil {
		err = types.Decode(value, &votes)
		if err != nil {
			return nil, err
		}
	}
	mp := make(map[string]struct{}, len(addrs))
	for _, addr := range addrs {
		mp[addr] = struct{}{}
	}
	// 检查是否有重复
	for _, addr := range votes.Address {
		if _, ok := mp[addr]; ok {
			err := auty.ErrRepeatVoteAddr
			alog.Error("autonomy ", "addr", addr, "err", err)
			return nil, err
		}
	}
	return &votes, nil
}

//新增addr场景，任一probAddr在当前board里即返回true
func checkAddrInBoard(act *auty.ActiveBoard, probAddrs map[string]struct{}) bool {
	for _, board := range act.Boards {
		if _, ok := probAddrs[board]; ok {
			alog.Info("propBoard repeated addr in boards", "addr", board)
			return true
		}
	}
	for _, board := range act.Revboards {
		if _, ok := probAddrs[board]; ok {
			alog.Info("propBoard repeated addr in revboards", "addr", board)
			return true
		}
	}
	return false
}

func (a *action) addPropBoard(prob *auty.ProposalBoard, mpBd map[string]struct{}) (*auty.ActiveBoard, error) {
	// only add member
	act, err := a.getActiveBoard()
	if err != nil {
		alog.Error("propBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "getActiveBoard failed", err)
		return nil, err
	}
	if checkAddrInBoard(act, mpBd) {
		return nil, errors.Wrap(auty.ErrRepeatAddr, "repeated addr in current boards")
	}

	act.Boards = append(act.Boards, prob.Boards...)
	return act, nil
}

//删除addr场景，若任一proposal addr不存在 则返回true,
func checkAddrNotInBoard(act *auty.ActiveBoard, prob *auty.ProposalBoard) error {
	actBoards := make(map[string]bool)
	for _, board := range act.Boards {
		actBoards[board] = true
	}

	for _, addr := range prob.Boards {
		if !actBoards[addr] {
			return errors.Wrapf(types.ErrNotFound, "addr=%s not in boards", addr)
		}
	}

	return nil
}

//这里只考虑Board，不考虑revBoard
func (a *action) delPropBoard(prob *auty.ProposalBoard, mpBd map[string]struct{}) (*auty.ActiveBoard, error) {
	act, err := a.getActiveBoard()
	if err != nil {
		alog.Error("propBoard ", "addr", a.fromaddr, "execaddr", a.execaddr, "getActiveBoard failed", err)
		return nil, err
	}
	err = checkAddrNotInBoard(act, prob)
	if err != nil {
		return nil, err
	}

	var newBoard []string
	for _, board := range act.Boards {
		if _, ok := mpBd[board]; !ok {
			newBoard = append(newBoard, board)
		}
	}

	act.Boards = newBoard
	return act, nil
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
