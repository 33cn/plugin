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
	kvs, err := v.updateAndSaveTable(table.Add, table.Save, groupInfo, tx, vty.NameCreateGroupAction, "group")
	if err != nil {
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

func (v *vote) ExecLocal_UpdateGroup(update *vty.UpdateGroup, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	groupInfo := decodeGroupInfo(receiptData.Logs[0].Log)
	table := newGroupTable(v.GetLocalDB())
	kvs, err := v.updateAndSaveTable(table.Replace, table.Save, groupInfo, tx, vty.NameUpdateGroupAction, "group")
	if err != nil {
		return nil, err
	}
	dbSet.KV = kvs
	kvs, err = v.removeGroupMember(groupInfo.GetID(), update.RemoveMembers)
	if err != nil {
		elog.Error("execLocal UpdateGroup", "txHash", hex.EncodeToString(tx.Hash()), "removeMemberErr", err)
		return nil, err
	}
	dbSet.KV = append(dbSet.KV, kvs...)
	kvs, err = v.addGroupMember(groupInfo.GetID(), update.AddMembers)
	if err != nil {
		elog.Error("execLocal UpdateGroup", "txHash", hex.EncodeToString(tx.Hash()), "addMemberErr", err)
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
	kvs, err := v.updateAndSaveTable(table.Add, table.Save, voteInfo, tx, vty.NameCreateVoteAction, "vote")
	if err != nil {
		return nil, err
	}
	dbSet.KV = kvs
	table = newGroupTable(v.GetLocalDB())
	row, err := table.GetData([]byte(voteInfo.GroupID))
	if err != nil {
		elog.Error("execLocal createVote", "txHash", hex.EncodeToString(tx.Hash()), "voteTable get", err)
		return nil, err
	}
	groupInfo, _ := row.Data.(*vty.GroupInfo)
	groupInfo.VoteNum++
	kvs, err = v.updateAndSaveTable(table.Replace, table.Save, groupInfo, tx, vty.NameCreateVoteAction, "group")
	if err != nil {
		return nil, err
	}
	dbSet.KV = append(dbSet.KV, kvs...)
	//auto gen for localdb auto rollback
	return v.addAutoRollBack(tx, dbSet.KV), nil
}

func (v *vote) ExecLocal_CommitVote(payload *vty.CommitVote, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code, add customize kv to dbSet...
	commitInfo := decodeCommitInfo(receiptData.Logs[0].Log)
	table := newVoteTable(v.GetLocalDB())
	row, err := table.GetData([]byte(payload.GetVoteID()))
	if err != nil {
		elog.Error("execLocal commitVote", "txHash", hex.EncodeToString(tx.Hash()), "table get", err)
		return nil, err
	}
	voteInfo, _ := row.Data.(*vty.VoteInfo)
	voteInfo.CommitInfos = append(voteInfo.CommitInfos, commitInfo)
	dbSet.KV, err = v.updateAndSaveTable(table.Replace, table.Save, voteInfo, tx, vty.NameCommitVoteAction, "vote")
	if err != nil {
		return nil, err
	}
	//auto gen for localdb auto rollback
	return v.addAutoRollBack(tx, dbSet.KV), nil
}

func (v *vote) ExecLocal_CloseVote(payload *vty.CloseVote, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	table := newVoteTable(v.GetLocalDB())
	row, err := table.GetData([]byte(payload.GetVoteID()))
	if err != nil {
		elog.Error("execLocal closeVote", "txHash", hex.EncodeToString(tx.Hash()), "table get", err)
		return nil, err
	}
	voteInfo, ok := row.Data.(*vty.VoteInfo)

	if !ok {
		elog.Error("execLocal closeVote", "txHash", hex.EncodeToString(tx.Hash()), "voteInfo type asset", err)
		return nil, types.ErrTypeAsset
	}

	voteInfo.Status = voteStatusClosed
	dbSet.KV, err = v.updateAndSaveTable(table.Replace, table.Save, voteInfo, tx, vty.NameCloseVoteAction, "vote")
	if err != nil {
		return nil, err
	}
	//auto gen for localdb auto rollback
	return v.addAutoRollBack(tx, dbSet.KV), nil
}

func (v *vote) ExecLocal_UpdateMember(payload *vty.UpdateMember, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	table := newMemberTable(v.GetLocalDB())
	row, err := table.GetData([]byte(tx.From()))
	if err != nil {
		elog.Error("execLocal updateMember", "txHash", hex.EncodeToString(tx.Hash()), "table get", err)
		return nil, err
	}
	memberInfo, _ := row.Data.(*vty.MemberInfo)
	memberInfo.Name = payload.GetName()
	dbSet.KV, err = v.updateAndSaveTable(table.Replace, table.Save, memberInfo, tx, vty.NameUpdateMemberAction, "member")
	if err != nil {
		return nil, err
	}
	//auto gen for localdb auto rollback
	return v.addAutoRollBack(tx, dbSet.KV), nil
}

//当区块回滚时，框架支持自动回滚localdb kv，需要对exec-local返回的kv进行封装
func (v *vote) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = v.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

type updateFunc func(message types.Message) error
type saveFunc func() ([]*types.KeyValue, error)

func (v *vote) updateAndSaveTable(update updateFunc, save saveFunc, data types.Message, tx *types.Transaction, actionName, tableName string) ([]*types.KeyValue, error) {

	err := update(data)
	if err != nil {
		elog.Error("execLocal "+actionName, "txHash", hex.EncodeToString(tx.Hash()), tableName+" table update", err)
		return nil, err
	}
	kvs, err := save()
	if err != nil {
		elog.Error("execLocal "+actionName, "txHash", hex.EncodeToString(tx.Hash()), tableName+" table save", err)
		return nil, err
	}
	return kvs, nil
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
