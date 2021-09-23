package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/system/dapp"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
)

type action struct {
	db        dbm.KV
	txHash    []byte
	fromAddr  string
	blockTime int64
	height    int64
	index     int
}

func newAction(v *vote, tx *types.Transaction, index int) *action {

	return &action{db: v.GetStateDB(),
		txHash:    tx.Hash(),
		fromAddr:  tx.From(),
		blockTime: v.GetBlockTime(),
		height:    v.GetHeight(),
		index:     index}
}

func (a *action) getGroupInfo(groupID string) (*vty.GroupInfo, error) {

	info := &vty.GroupInfo{}
	err := readStateDB(a.db, formatStateIDKey(groupID), info)
	if err == types.ErrNotFound {
		err = errGroupNotExist
	}
	return info, err
}

func (a *action) getVoteInfo(voteID string) (*vty.VoteInfo, error) {

	info := &vty.VoteInfo{}
	err := readStateDB(a.db, formatStateIDKey(voteID), info)
	if err == types.ErrNotFound {
		err = errVoteNotExist
	}
	return info, err
}

func (a *action) createGroup(create *vty.CreateGroup) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}

	group := &vty.GroupInfo{}
	group.Name = create.Name
	group.ID = formatGroupID(dapp.HeightIndexStr(a.height, int64(a.index)))
	//添加创建者作为默认管理员
	group.Admins = append(group.Admins, a.fromAddr)
	for _, addr := range create.GetAdmins() {
		if addr != a.fromAddr {
			group.Admins = append(group.Admins, addr)
		}
	}

	group.Members = create.Members
	// set default vote weight
	for _, member := range group.Members {
		if member.VoteWeight < 1 {
			member.VoteWeight = 1
		}
	}
	group.MemberNum = uint32(len(group.Members))
	group.Creator = a.fromAddr
	group.Description = create.GetDescription()
	groupValue := types.Encode(group)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: formatStateIDKey(group.ID), Value: groupValue})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: vty.TyCreateGroupLog, Log: groupValue})

	return receipt, nil
}

func (a *action) updateGroup(update *vty.UpdateGroup) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}

	group, err := a.getGroupInfo(update.GroupID)
	if err != nil {
		elog.Error("vote exec updateGroup", "txHash", a.txHash, "err", err)
		return nil, errStateDBGet
	}
	addrMap := make(map[string]int)
	for index, member := range group.Members {
		addrMap[member.Addr] = index
	}
	// remove members
	for _, addr := range update.GetRemoveMembers() {
		if index, ok := addrMap[addr]; ok {
			group.Members = append(group.Members[:index], group.Members[index+1:]...)
			delete(addrMap, addr)
		}
	}

	// add members
	for _, member := range update.GetAddMembers() {
		if _, ok := addrMap[member.Addr]; !ok {
			group.Members = append(group.Members, member)
		}
	}
	group.MemberNum = uint32(len(group.Members))
	adminMap := make(map[string]int)
	for index, addr := range group.Admins {
		adminMap[addr] = index
	}

	// remove admins
	for _, addr := range update.GetRemoveAdmins() {
		if index, ok := adminMap[addr]; ok {
			group.Admins = append(group.Admins[:index], group.Admins[index+1:]...)
			delete(adminMap, addr)
		}
	}

	// add admins
	for _, addr := range update.GetAddAdmins() {
		if _, ok := adminMap[addr]; !ok {
			group.Admins = append(group.Admins, addr)
		}
	}

	groupValue := types.Encode(group)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: formatStateIDKey(group.ID), Value: groupValue})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: vty.TyUpdateGroupLog, Log: groupValue})

	return receipt, nil
}

func (a *action) createVote(create *vty.CreateVote) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}

	vote := &vty.VoteInfo{}
	vote.ID = formatVoteID(dapp.HeightIndexStr(a.height, int64(a.index)))
	vote.BeginTimestamp = create.BeginTimestamp
	vote.EndTimestamp = create.EndTimestamp
	vote.Name = create.Name
	vote.GroupID = create.GroupID
	vote.Description = create.Description
	vote.Creator = a.fromAddr
	vote.VoteOptions = make([]*vty.VoteOption, 0)
	for _, option := range create.VoteOptions {
		vote.VoteOptions = append(vote.VoteOptions, &vty.VoteOption{Option: option})
	}

	voteValue := types.Encode(vote)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: formatStateIDKey(vote.ID), Value: voteValue})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: vty.TyCreateVoteLog, Log: voteValue})

	return receipt, nil
}

func (a *action) commitVote(commit *vty.CommitVote) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}
	vote, err := a.getVoteInfo(commit.VoteID)
	if err != nil {
		elog.Error("vote exec commitVote", "txHash", a.txHash, "vid", commit.VoteID, "get vote err", err)
		return nil, errStateDBGet
	}
	// 由于目前底层检测交易阶段提供区块时间可能不是最新区块的，依赖区块时间的比较需要放在执行阶段处理
	if vote.BeginTimestamp > a.blockTime {
		elog.Error("vote exec commitVote", "txHash", a.txHash, "vid", commit.VoteID,
			"beginTime", vote.BeginTimestamp, "blockTime", a.blockTime, "err", errVoteNotStarted)
		return nil, errVoteNotStarted
	}
	if vote.EndTimestamp <= a.blockTime {
		elog.Error("vote exec commitVote", "txHash", a.txHash, "vid", commit.VoteID,
			"endTime", vote.EndTimestamp, "blockTime", a.blockTime, "err", errVoteAlreadyFinished)
		return nil, errVoteAlreadyFinished
	}

	group, err := a.getGroupInfo(vote.GroupID)
	if err != nil {
		elog.Error("vote exec commitVote", "txHash", a.txHash, "get group err", err)
		return nil, errStateDBGet
	}
	var voteWeight uint32
	for _, member := range group.Members {
		if member.Addr == a.fromAddr {
			voteWeight = member.VoteWeight
		}
	}
	vote.VoteOptions[commit.OptionIndex].Score += voteWeight
	info := &vty.CommitInfo{Addr: a.fromAddr}
	vote.CommitInfos = append(vote.CommitInfos, info)
	voteValue := types.Encode(vote)
	//提交的哈希和权重等信息不记录到statedb中
	info.VoteWeight = voteWeight
	info.TxHash = hex.EncodeToString(a.txHash)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: formatStateIDKey(vote.ID), Value: voteValue})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: vty.TyCommitVoteLog, Log: types.Encode(info)})

	return receipt, nil
}

func (a *action) closeVote(close *vty.CloseVote) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}
	vote, err := a.getVoteInfo(close.VoteID)
	if err != nil {
		elog.Error("vote exec commitVote", "txHash", a.txHash, "get vote err", err)
		return nil, errStateDBGet
	}
	vote.Status = voteStatusClosed
	voteValue := types.Encode(vote)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: formatStateIDKey(vote.ID), Value: voteValue})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: vty.TyCloseVoteLog, Log: voteValue})

	return receipt, nil
}

func (a *action) updateMember(update *vty.UpdateMember) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}
	return receipt, nil
}
