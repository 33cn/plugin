package executor

import (
	"testing"

	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
	"github.com/stretchr/testify/require"
)

func TestVote_Query_GetGroups(t *testing.T) {
	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	groupID2 := formatGroupID(dapp.HeightIndexStr(testHeight, 1))
	groupID3 := formatGroupID(dapp.HeightIndexStr(testHeight, 2))
	groupIDs := []string{groupID, groupID2, groupID3, "testid"}
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "test"},
	}, {
		index:   1,
		payload: &vty.CreateGroup{Name: "test"},
	},
	}

	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])

	exec := mock.exec
	funcName := "GetGroups"
	_, err := exec.Query(funcName, nil)
	require.Equal(t, types.ErrInvalidParam, err)
	_, err = exec.Query(funcName, types.Encode(&vty.ReqStrings{Items: groupIDs[3:]}))
	require.Equal(t, errInvalidGroupID, err)
	_, err = exec.Query(funcName, types.Encode(&vty.ReqStrings{Items: groupIDs[:3]}))
	require.Equal(t, types.ErrNotFound, err)
	data, err := exec.Query(funcName, types.Encode(&vty.ReqStrings{Items: groupIDs[:2]}))
	require.Equal(t, nil, err)
	groups := data.(*vty.GroupInfos)
	require.Equal(t, 2, len(groups.GroupList))
	require.Equal(t, groupID, groups.GroupList[0].ID)
	require.Equal(t, groupID2, groups.GroupList[1].ID)
}

func TestVote_Query_GetVotes(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	voteID := formatVoteID(dapp.HeightIndexStr(testHeight, 1))
	voteID2 := formatVoteID(dapp.HeightIndexStr(testHeight, 2))
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "test"},
	}, {
		index: 1,
		payload: &vty.CreateVote{Name: "test", GroupID: groupID, VoteOptions: []string{"A", "B"},
			BeginTimestamp: testBlockTime, EndTimestamp: testBlockTime + 1},
	}}
	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])

	exec := mock.exec
	funcName := "GetVotes"
	_, err := exec.Query(funcName, nil)
	require.Equal(t, types.ErrInvalidParam, err)
	_, err = exec.Query(funcName, types.Encode(&vty.ReqStrings{Items: []string{voteID2}}))
	require.Equal(t, types.ErrNotFound, err)
	_, err = exec.Query(funcName, types.Encode(&vty.ReqStrings{Items: []string{"voteid"}}))
	require.Equal(t, errInvalidVoteID, err)
	data, err := exec.Query(funcName, types.Encode(&vty.ReqStrings{Items: []string{voteID}}))
	require.Equal(t, nil, err)
	vote := data.(*vty.ReplyVoteList)
	require.Equal(t, voteID, vote.VoteList[0].ID)
}

func TestVote_Query_GetMembers(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "test", Members: []*vty.GroupMember{{Addr: testAddrs[0]}}},
	}}
	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])

	exec := mock.exec
	funcName := "GetMembers"
	_, err := exec.Query(funcName, nil)
	require.Equal(t, types.ErrInvalidParam, err)
	_, err = exec.Query(funcName, types.Encode(&vty.ReqStrings{Items: []string{testAddrs[1]}}))
	require.Equal(t, types.ErrNotFound, err)
	_, err = exec.Query(funcName, types.Encode(&vty.ReqStrings{Items: []string{"addr"}}))
	require.Equal(t, types.ErrInvalidAddress, err)
	data, err := exec.Query(funcName, types.Encode(&vty.ReqStrings{Items: []string{testAddrs[0]}}))
	require.Equal(t, nil, err)
	members := data.(*vty.MemberInfos)
	require.Equal(t, testAddrs[0], members.MemberList[0].Addr)
	require.Equal(t, []string{groupID}, members.MemberList[0].GroupIDs)
}

func TestVote_Query_ListGroup(t *testing.T) {
	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	groupID2 := formatGroupID(dapp.HeightIndexStr(testHeight, 1))
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "test"},
	}, {
		index:   1,
		payload: &vty.CreateGroup{Name: "test"},
	},
	}

	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])

	exec := mock.exec
	funcName := "ListGroup"
	data, err := exec.Query(funcName, nil)
	require.Equal(t, nil, err)
	list := data.(*vty.GroupInfos)
	require.Equal(t, 2, len(list.GroupList))
	data, err = exec.Query(funcName, types.Encode(&vty.ReqListItem{Count: 1, Direction: 1}))
	require.Equal(t, nil, err)
	list = data.(*vty.GroupInfos)
	require.Equal(t, 1, len(list.GroupList))
	require.Equal(t, groupID, list.GroupList[0].ID)
	data, err = exec.Query(funcName, types.Encode(&vty.ReqListItem{StartItemID: groupID2}))
	require.Equal(t, nil, err)
	list = data.(*vty.GroupInfos)
	require.Equal(t, 1, len(list.GroupList))
	require.Equal(t, groupID, list.GroupList[0].ID)
}

func TestVote_Query_ListVote(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	now := types.Now().Unix() + 1000
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "test"},
	}, {
		index: 1,
		payload: &vty.CreateVote{Name: "test", GroupID: groupID, VoteOptions: []string{"A", "B"},
			EndTimestamp: testBlockTime + 1},
	}, {
		index: 2,
		payload: &vty.CreateVote{Name: "test", GroupID: groupID, VoteOptions: []string{"A", "B"},
			BeginTimestamp: now, EndTimestamp: now + 1},
	}}
	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])

	exec := mock.exec
	funcName := "ListVote"
	_, err := exec.Query(funcName, nil)
	require.Equal(t, types.ErrInvalidParam, err)
	data, err := exec.Query(funcName, types.Encode(&vty.ReqListVote{GroupID: groupID, ListReq: &vty.ReqListItem{}}))
	require.Nil(t, err)
	list := data.(*vty.ReplyVoteList)
	require.Equal(t, 2, len(list.VoteList))

	data, err = exec.Query(funcName, types.Encode(&vty.ReqListVote{GroupID: groupID, Status: voteStatusPending, ListReq: &vty.ReqListItem{}}))
	require.Nil(t, err)
	list = data.(*vty.ReplyVoteList)
	require.Equal(t, 1, len(list.VoteList))
	require.Equal(t, uint32(voteStatusPending), list.VoteList[0].Status)
}

func TestVote_Query_ListMember(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "test", Members: []*vty.GroupMember{{Addr: testAddrs[0]}}},
	}}
	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])

	exec := mock.exec
	funcName := "ListMember"
	data, err := exec.Query(funcName, nil)
	require.Equal(t, nil, err)
	list := data.(*vty.MemberInfos)
	require.Equal(t, 1, len(list.MemberList))
	require.Equal(t, testAddrs[0], list.MemberList[0].Addr)
	require.Equal(t, []string{groupID}, list.MemberList[0].GroupIDs)
	data, err = exec.Query(funcName, types.Encode(&vty.ReqListItem{StartItemID: "addr", Direction: 1}))
	require.Equal(t, nil, err)
	list = data.(*vty.MemberInfos)
	require.Equal(t, &vty.MemberInfos{}, list)
}
