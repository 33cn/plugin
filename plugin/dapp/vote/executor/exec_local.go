package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (v *vote) ExecLocal_CreateGroup(payload *vty.CreateGroup, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}

	groupInfo := decodeGroupInfo(receiptData.Logs[0].Log)
	table := newGroupTable(v.GetLocalDB())
	err := table.Add(&vty.GroupVoteInfo{GroupInfo: groupInfo})
	if err != nil {
		elog.Error("execLocal createGroup", "txHash", hex.EncodeToString(tx.Hash()), "table add", err)
		return nil, err
	}
	kvs, err := table.Save()
	if err != nil {
		elog.Error("execLocal createGroup", "txHash", hex.EncodeToString(tx.Hash()), "table save", err)
		return nil, err
	}
	dbSet.KV = kvs

	kvs, err = v.addGroupMember(groupInfo.GetID(), groupInfo.Members)
	if err != nil {
		elog.Error("execLocal createGroup", "txHash", hex.EncodeToString(tx.Hash()), "addMemberErr", err)
		return nil, err
	}
	dbSet.KV = append(dbSet.KV, kvs...)

	//auto gen for localdb auto rollback
	return v.addAutoRollBack(tx, dbSet.KV), nil
}

func (v *vote) ExecLocal_UpdateMember(update *vty.UpdateMember, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	groupInfo := decodeGroupInfo(receiptData.Logs[0].Log)
	table := newGroupTable(v.GetLocalDB())
	err := table.Replace(&vty.GroupVoteInfo{GroupInfo: groupInfo})
	if err != nil {
		elog.Error("execLocal updateMember", "txHash", hex.EncodeToString(tx.Hash()), "groupID", groupInfo.ID, "table replace", err)
		return nil, err
	}
	kvs, err := table.Save()
	if err != nil {
		elog.Error("execLocal updateMember", "txHash", hex.EncodeToString(tx.Hash()), "groupID", groupInfo.ID, "table save", err)
		return nil, err
	}
	dbSet.KV = kvs

	kvs, err = v.addGroupMember(groupInfo.GetID(), update.AddMembers)
	if err != nil {
		elog.Error("execLocal updateMember", "txHash", hex.EncodeToString(tx.Hash()), "addMemberErr", err)
		return nil, err
	}
	dbSet.KV = append(dbSet.KV, kvs...)
	kvs, err = v.removeGroupMember(groupInfo.GetID(), update.RemoveMemberAddrs)
	if err != nil {
		elog.Error("execLocal updateMember", "txHash", hex.EncodeToString(tx.Hash()), "removeMemberErr", err)
		return nil, err
	}
	dbSet.KV = append(dbSet.KV, kvs...)
	//auto gen for localdb auto rollback
	return v.addAutoRollBack(tx, dbSet.KV), nil
}

func (v *vote) ExecLocal_CreateVote(payload *vty.CreateVote, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	voteInfo := decodeVoteInfo(receiptData.Logs[0].Log)
	table := newVoteTable(v.GetLocalDB())
	err := table.Add(voteInfo)
	if err != nil {
		elog.Error("execLocal createVote", "txHash", hex.EncodeToString(tx.Hash()), "voteTable add", err)
		return nil, err
	}
	kvs, err := table.Save()
	if err != nil {
		elog.Error("execLocal createVote", "txHash", hex.EncodeToString(tx.Hash()), "voteTable save", err)
		return nil, err
	}
	dbSet.KV = kvs

	// 在关联的投票组表中记录voteID信息
	table = newGroupTable(v.GetLocalDB())
	for _, groupID := range voteInfo.VoteGroups {
		row, err := table.GetData([]byte(groupID))
		if err != nil {
			continue
		}
		if info, ok := row.Data.(*vty.GroupVoteInfo); ok {
			info.VoteIDs = append(info.VoteIDs, voteInfo.ID)
			err = table.Replace(info)
			if err != nil {
				elog.Error("execLocal createVote", "txHash", hex.EncodeToString(tx.Hash()),
					"groupID", groupID, "groupTable replace", err)
				return nil, err
			}
		}
	}
	kvs, err = table.Save()
	if err != nil {
		elog.Error("execLocal createVote", "txHash", hex.EncodeToString(tx.Hash()), "groupTable save", err)
		return nil, err
	}
	dbSet.KV = append(dbSet.KV, kvs...)
	//auto gen for localdb auto rollback
	return v.addAutoRollBack(tx, dbSet.KV), nil
}

func (v *vote) ExecLocal_CommitVote(payload *vty.CommitVote, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code, add customize kv to dbSet...
	voteInfo := decodeVoteInfo(receiptData.Logs[0].Log)
	table := newVoteTable(v.GetLocalDB())
	err := table.Replace(voteInfo)
	if err != nil {
		elog.Error("execLocal commitVote", "txHash", hex.EncodeToString(tx.Hash()), "voteTable add", err)
		return nil, err
	}
	kvs, err := table.Save()
	if err != nil {
		elog.Error("execLocal commitVote", "txHash", hex.EncodeToString(tx.Hash()), "voteTable save", err)
		return nil, err
	}
	dbSet.KV = kvs
	//auto gen for localdb auto rollback
	return v.addAutoRollBack(tx, dbSet.KV), nil
}

//当区块回滚时，框架支持自动回滚localdb kv，需要对exec-local返回的kv进行封装
func (v *vote) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = v.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

func (v *vote) addGroupMember(groupID string, members []*vty.GroupMember) ([]*types.KeyValue, error) {

	table := newMemberTable(v.GetLocalDB())
	for _, member := range members {
		addrKey := []byte(member.Addr)
		row, err := table.GetData(addrKey)
		if err == nil {
			info, ok := row.Data.(*vty.MemberInfo)
			if ok && !checkSliceItemExist(groupID, info.GroupIDs) {
				info.GroupIDs = append(info.GroupIDs, groupID)
				err = table.Replace(info)
			}
		} else if err == types.ErrNotFound {
			err = table.Add(&vty.MemberInfo{Addr: member.Addr, GroupIDs: []string{groupID}})
		}

		// 这个错可能由GetData，Replace，Add返回
		if err != nil {
			elog.Error("execLocal addMember", "member table Add/Replace", err)
			return nil, err
		}
	}
	kvs, err := table.Save()
	if err != nil {
		elog.Error("execLocal addMember", "member table save", err)
		return nil, err
	}
	return kvs, nil
}

func (v *vote) removeGroupMember(groupID string, addrs []string) ([]*types.KeyValue, error) {

	table := newMemberTable(v.GetLocalDB())
	for _, addr := range addrs {
		addrKey := []byte(addr)
		row, err := table.GetData(addrKey)
		if err == types.ErrNotFound {
			return nil, nil
		} else if err != nil {
			elog.Error("execLocal removeMember", "member table getData", err)
			return nil, err
		}

		info, ok := row.Data.(*vty.MemberInfo)
		if !ok {
			return nil, types.ErrTypeAsset
		}
		for index, id := range info.GroupIDs {
			if id == groupID {
				info.GroupIDs = append(info.GroupIDs[:index], info.GroupIDs[index+1:]...)
				err = table.Replace(info)
				if err != nil {
					elog.Error("execLocal removeMember", "member table replace", err)
					return nil, err
				}
				break
			}
		}
	}
	kvs, err := table.Save()
	if err != nil {
		elog.Error("execLocal addMember", "member table save", err)
		return nil, err
	}
	return kvs, nil
}
