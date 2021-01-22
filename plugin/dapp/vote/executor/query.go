package executor

import (
	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
)

func (v *vote) getGroup(groupID string) (*vty.GroupInfo, error) {

	if len(groupID) != IDLen {
		return nil, errInvalidGroupID
	}
	table := newGroupTable(v.GetLocalDB())
	row, err := table.GetData([]byte(groupID))

	if err != nil {
		elog.Error("query getGroup", "groupID", groupID, "err", err)
		return nil, err
	}

	info, ok := row.Data.(*vty.GroupInfo)
	if !ok {
		return nil, types.ErrTypeAsset
	}

	return info, nil
}

// Query_GroupInfo query group info
func (v *vote) Query_GetGroups(in *vty.ReqStrings) (types.Message, error) {

	if in == nil {
		return nil, types.ErrInvalidParam
	}
	infos := &vty.GroupInfos{GroupList: make([]*vty.GroupInfo, 0, len(in.GetItems()))}
	for _, id := range in.GetItems() {
		info, err := v.getGroup(id)
		if err != nil {
			return nil, err
		}
		infos.GroupList = append(infos.GroupList, info)
	}
	return infos, nil
}

func (v *vote) getVote(voteID string) (*vty.VoteInfo, error) {
	if len(voteID) != IDLen {
		return nil, errInvalidVoteID
	}
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

func (v *vote) Query_GetVotes(in *vty.ReqStrings) (types.Message, error) {

	if in == nil {
		return nil, types.ErrInvalidParam
	}
	infos := &vty.VoteInfos{VoteList: make([]*vty.VoteInfo, 0, len(in.GetItems()))}
	for _, id := range in.GetItems() {
		info, err := v.getVote(id)
		if err != nil {
			return nil, err
		}
		infos.VoteList = append(infos.VoteList, info)
	}
	return classifyVoteList(infos), nil

}

func (v *vote) getMember(addr string) (*vty.MemberInfo, error) {

	if len(addr) != addrLen {
		return nil, types.ErrInvalidAddress
	}
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

func (v *vote) Query_GetMembers(in *vty.ReqStrings) (types.Message, error) {

	if in == nil {
		return nil, types.ErrInvalidParam
	}
	infos := &vty.MemberInfos{MemberList: make([]*vty.MemberInfo, 0, len(in.GetItems()))}
	for _, id := range in.GetItems() {
		info, err := v.getMember(id)
		if err != nil {
			return nil, err
		}
		infos.MemberList = append(infos.MemberList, info)
	}
	return infos, nil
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

	list := &vty.GroupInfos{GroupList: make([]*vty.GroupInfo, 0, len(rows))}
	for _, row := range rows {
		info, ok := row.Data.(*vty.GroupInfo)
		if !ok {
			return nil, types.ErrTypeAsset
		}
		list.GroupList = append(list.GroupList, info)
	}

	return list, nil
}

func (v *vote) Query_ListVote(in *vty.ReqListVote) (types.Message, error) {

	if in.GetListReq() == nil {
		return nil, types.ErrInvalidParam
	}
	table := newVoteTable(v.GetLocalDB())
	//指定了组ID，则查询对应组下的投票列表
	groupID := in.GetGroupID()
	indexName := voteTablePrimary
	var prefix, primaryKey []byte
	if len(groupID) > 0 {
		indexName = groupTablePrimary
		prefix = []byte(groupID)
	}
	primaryKey = append(primaryKey, []byte(in.GetListReq().GetStartItemID())...)
	rows, err := table.ListIndex(indexName, prefix, primaryKey, in.GetListReq().Count, in.GetListReq().Direction)
	if err != nil {
		elog.Error("query listVote", "err", err, "param", in)
		return nil, err
	}

	list := &vty.VoteInfos{VoteList: make([]*vty.VoteInfo, 0, len(rows))}
	for _, row := range rows {
		info, ok := row.Data.(*vty.VoteInfo)
		if !ok {
			return nil, types.ErrTypeAsset
		}
		list.VoteList = append(list.VoteList, info)
	}

	return classifyVoteList(list), nil
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

	list := &vty.MemberInfos{MemberList: make([]*vty.MemberInfo, 0, len(rows))}
	for _, row := range rows {
		info, ok := row.Data.(*vty.MemberInfo)
		if !ok {
			return nil, types.ErrTypeAsset
		}
		list.MemberList = append(list.MemberList, info)
	}

	return list, nil
}
