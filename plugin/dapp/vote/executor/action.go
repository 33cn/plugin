package executor

import (
	"encoding/hex"

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

	xhash := hex.EncodeToString(a.txHash)
	group := &vty.GroupInfo{}
	group.Name = create.Name
	group.ID = formatGroupID(xhash)
	//添加创建者作为默认管理员
	group.Admins = append(group.Admins, create.Admins...)
	if !checkSliceItemExist(a.fromAddr, create.Admins) {
		group.Admins = append(group.Admins, a.fromAddr)
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
	groupValue := types.Encode(group)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: formatStateIDKey(group.ID), Value: groupValue})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: vty.TyCreateGroupLog, Log: groupValue})

	return receipt, nil
}

func (a *action) updateMember(update *vty.UpdateMember) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}

	group, err := a.getGroupInfo(update.GroupID)
	if err != nil {
		elog.Error("vote exec updateMember", "txHash", a.txHash, "err", err)
		return nil, errStateDBGet
	}
	addrMap := make(map[string]int)
	for index, member := range group.Members {
		addrMap[member.Addr] = index
	}
	// remove members
	for _, addr := range update.GetRemoveMemberAddrs() {
		if index, ok := addrMap[addr]; ok {
			group.Members = append(group.Members[:index], group.Members[index+1:]...)
		}
	}

	// add members
	for _, member := range update.GetAddMembers() {
		group.Members = append(group.Members, member)
	}

	groupValue := types.Encode(group)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: formatStateIDKey(group.ID), Value: groupValue})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: vty.TyUpdateMemberLog, Log: groupValue})

	return receipt, nil
}

func (a *action) createVote(create *vty.CreateVote) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}

	vote := &vty.VoteInfo{}
	vote.ID = formatVoteID(hex.EncodeToString(a.txHash))
	vote.BeginTimestamp = create.BeginTimestamp
	vote.EndTimestamp = create.EndTimestamp
	vote.Name = create.Name
	vote.VoteGroups = create.VoteGroups

	vote.VoteOptions = make([]*vty.VoteOption, 0)
	for _, option := range create.VoteOptions {
		vote.VoteOptions = append(vote.VoteOptions, &vty.VoteOption{Option: option})
	}
	vote.Creator = a.fromAddr
	voteValue := types.Encode(vote)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: formatStateIDKey(vote.ID), Value: voteValue})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: vty.TyCreateVoteLog, Log: voteValue})

	return receipt, nil
}

func (a *action) commitVote(commit *vty.CommitVote) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}
	vote, err := a.getVoteInfo(commit.VoteID)
	group, err1 := a.getGroupInfo(commit.GroupID)
	if err != nil || err1 != nil {
		elog.Error("vote exec commitVote", "txHash", a.txHash, "err", err, "err1", err1)
		return nil, errStateDBGet
	}

	for _, member := range group.Members {
		if member.Addr == a.fromAddr {
			vote.VoteOptions[commit.OptionIndex].Score += member.VoteWeight
		}
	}
	vote.VotedMembers = append(vote.VotedMembers, a.fromAddr)
	voteValue := types.Encode(vote)
	receipt.KV = append(receipt.KV, &types.KeyValue{Key: formatStateIDKey(vote.ID), Value: voteValue})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: vty.TyCommitVoteLog, Log: voteValue})

	return receipt, nil
}
