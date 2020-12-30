package executor

import (
	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
)

// Query_GroupInfo query group info
func (v *vote) Query_GetGroup(in *types.ReqString) (types.Message, error) {

	if len(in.GetData()) != IDLen {
		return nil, errInvalidGroupID
	}
	groupID := in.Data
	table := newGroupTable(v.GetLocalDB())
	row, err := table.GetData([]byte(groupID))

	if err != nil {
		elog.Error("query getGroup", "id", groupID, "err", err)
		return nil, err
	}

	info, ok := row.Data.(*vty.GroupVoteInfo)
	if !ok {
		return nil, types.ErrTypeAsset
	}

	return info, nil

}

func (v *vote) Query_GetVote(in *types.ReqString) (types.Message, error) {

	if len(in.GetData()) != IDLen {
		return nil, errInvalidVoteID
	}
	voteID := in.Data
	table := newVoteTable(v.GetLocalDB())
	row, err := table.GetData([]byte(voteID))

	if err != nil {
		elog.Error("query getVote", "id", voteID, "err", err)
		return nil, err
	}

	info, ok := row.Data.(*vty.VoteInfo)
	if !ok {
		return nil, types.ErrTypeAsset
	}

	return info, nil

}

func (v *vote) Query_GetMember(in *types.ReqString) (types.Message, error) {

	if len(in.GetData()) != addrLen {
		return nil, types.ErrInvalidAddress
	}
	addr := in.Data
	table := newMemberTable(v.GetLocalDB())
	row, err := table.GetData([]byte(addr))

	if err != nil {
		elog.Error("query getMember", "addr", addr, "err", err)
		return nil, err
	}

	info, ok := row.Data.(*vty.MemberInfo)
	if !ok {
		return nil, types.ErrTypeAsset
	}
	return info, nil
}

func (v *vote) Query_ListGroup(in *vty.ReqListItem) (types.Message, error) {

	if in == nil {
		return nil, types.ErrInvalidParam
	}
	table := newGroupTable(v.GetLocalDB())
	var primaryKey []byte
	primaryKey = append(primaryKey, []byte(in.StartItemID)...)
	rows, err := table.ListIndex(groupTablePrimary, nil, primaryKey, in.Count, in.Direction)
	if err != nil {
		elog.Error("query listGroup", "err", err, "param", in)
		return nil, err
	}

	list := &vty.GroupVoteInfos{GroupList: make([]*vty.GroupVoteInfo, 0)}
	for _, row := range rows {
		info, ok := row.Data.(*vty.GroupVoteInfo)
		if !ok {
			return nil, types.ErrTypeAsset
		}
		list.GroupList = append(list.GroupList, info)
	}

	return list, nil
}

func (v *vote) Query_ListVote(in *vty.ReqListItem) (types.Message, error) {

	if in == nil {
		return nil, types.ErrInvalidParam
	}
	table := newVoteTable(v.GetLocalDB())
	var primaryKey []byte
	primaryKey = append(primaryKey, []byte(in.StartItemID)...)
	rows, err := table.ListIndex(voteTablePrimary, nil, primaryKey, in.Count, in.Direction)
	if err != nil {
		elog.Error("query listVote", "err", err, "param", in)
		return nil, err
	}

	list := &vty.VoteInfos{VoteList: make([]*vty.VoteInfo, 0)}
	for _, row := range rows {
		info, ok := row.Data.(*vty.VoteInfo)
		if !ok {
			return nil, types.ErrTypeAsset
		}
		list.VoteList = append(list.VoteList, info)
	}

	return list, nil
}

func (v *vote) Query_ListMember(in *vty.ReqListItem) (types.Message, error) {

	if in == nil {
		return nil, types.ErrInvalidParam
	}
	table := newMemberTable(v.GetLocalDB())
	var primaryKey []byte
	primaryKey = append(primaryKey, []byte(in.StartItemID)...)
	rows, err := table.ListIndex(memberTablePrimary, nil, primaryKey, in.Count, in.Direction)
	if err != nil {
		elog.Error("query listMember", "err", err, "param", in)
		return nil, err
	}

	list := &vty.MemberInfos{MemberList: make([]*vty.MemberInfo, 0)}
	for _, row := range rows {
		info, ok := row.Data.(*vty.MemberInfo)
		if !ok {
			return nil, types.ErrTypeAsset
		}
		list.MemberList = append(list.MemberList, info)
	}

	return list, nil
}
