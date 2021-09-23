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

	if len(in.GetItems()) == 0 {
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

	if len(in.GetItems()) == 0 {
		return nil, types.ErrInvalidParam
	}
	voteList := make([]*vty.VoteInfo, 0, len(in.GetItems()))
	for _, id := range in.GetItems() {
		info, err := v.getVote(id)
		if err != nil {
			return nil, err
		}
		voteList = append(voteList, info)
	}
	reply := &vty.ReplyVoteList{CurrentTimestamp: types.Now().Unix()}
	reply.VoteList = filterVoteWithStatus(voteList, 0, reply.CurrentTimestamp)
	return reply, nil

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

	if len(in.GetItems()) == 0 {
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
	list := &vty.GroupInfos{}
	rows, err := table.ListIndex(groupTablePrimary, nil, primaryKey, in.Count, in.Direction)
	// 已经没有数据，直接返回
	if err == types.ErrNotFound {
		return list, nil
	}
	if err != nil {
		elog.Error("query listGroup", "err", err, "param", in)
		return nil, err
	}

	list.GroupList = make([]*vty.GroupInfo, 0, len(rows))
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
	reply := &vty.ReplyVoteList{CurrentTimestamp: types.Now().Unix()}
	listCount := in.ListReq.GetCount()

listMore:
	rows, err := table.ListIndex(indexName, prefix, primaryKey, listCount, in.GetListReq().Direction)
	// 已经没有数据，直接返回
	if err == types.ErrNotFound {
		return reply, nil
	}

	if err != nil {
		elog.Error("query listVote", "err", err, "param", in)
		return nil, err
	}

	list := make([]*vty.VoteInfo, 0, len(rows))
	for _, row := range rows {
		info, ok := row.Data.(*vty.VoteInfo)
		if !ok {
			return nil, types.ErrTypeAsset
		}
		list = append(list, info)
	}
	primaryKey = append(primaryKey[:0], []byte(list[len(list)-1].ID)...)
	list = filterVoteWithStatus(list, in.Status, reply.CurrentTimestamp)
	reply.VoteList = append(reply.VoteList, list...)
	//经过筛选后，数量小于请求数量，则需要再次list, 需要满足len(rows)==listCount, 否则表示已经没有数据
	if len(rows) == int(listCount) && int(listCount) > len(list) {
		listCount -= int32(len(list))
		goto listMore
	}

	return reply, nil
}

func (v *vote) Query_ListMember(in *vty.ReqListItem) (types.Message, error) {

	if in == nil {
		return nil, types.ErrInvalidParam
	}
	table := newMemberTable(v.GetLocalDB())
	var primaryKey []byte
	primaryKey = append(primaryKey, []byte(in.StartItemID)...)
	list := &vty.MemberInfos{}
	rows, err := table.ListIndex(memberTablePrimary, nil, primaryKey, in.Count, in.Direction)
	// 已经没有数据，直接返回
	if err == types.ErrNotFound {
		return list, nil
	}
	if err != nil {
		elog.Error("query listMember", "err", err, "param", in)
		return nil, err
	}

	list.MemberList = make([]*vty.MemberInfo, 0, len(rows))
	for _, row := range rows {
		info, ok := row.Data.(*vty.MemberInfo)
		if !ok {
			return nil, types.ErrTypeAsset
		}
		list.MemberList = append(list.MemberList, info)
	}

	return list, nil
}
