package executor

import (
	"testing"

	tab "github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
	"github.com/stretchr/testify/require"
)

type tableCase struct {
	index        int
	key          []byte
	expectGetErr error
	expectData   types.Message
}

func testTableData(t *testing.T, table *tab.Table, tcArr []*tableCase, msg string) {
	for _, tc := range tcArr {
		row, err := table.GetData(tc.key)
		require.Equalf(t, tc.expectGetErr, err, msg+",index=%d", tc.index)
		if err != nil {
			continue
		}
		require.Equalf(t, tc.expectData.String(), row.Data.String(),
			msg+",index=%d", tc.index)
	}
}

func TestVote_ExecLocal_CreateGroup(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	groupID2 := formatGroupID(dapp.HeightIndexStr(testHeight, 1))
	members1 := []*vty.GroupMember{{Addr: testAddrs[1], VoteWeight: 1}}
	members2 := []*vty.GroupMember{{Addr: testAddrs[2], VoteWeight: 1}}
	tcArr := []*testcase{{
		index: 0,
		payload: &vty.CreateGroup{
			Name:    "test",
			Members: members1},
	}, {
		index: 1,
		payload: &vty.CreateGroup{
			Name:    "test",
			Members: members2},
	},
	}
	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])
	table := newMemberTable(mock.exec.GetLocalDB())
	tcArr1 := []*tableCase{{
		index:      0,
		key:        []byte(testAddrs[0]),
		expectData: &vty.MemberInfo{Addr: testAddrs[0], GroupIDs: []string{groupID, groupID2}},
	}, {
		index:      1,
		key:        []byte(testAddrs[1]),
		expectData: &vty.MemberInfo{Addr: testAddrs[1], GroupIDs: []string{groupID}},
	}, {
		index:      2,
		key:        []byte(testAddrs[2]),
		expectData: &vty.MemberInfo{Addr: testAddrs[2], GroupIDs: []string{groupID2}},
	}, {
		index:        3,
		key:          []byte("addr"),
		expectGetErr: types.ErrNotFound,
	}}
	testTableData(t, table, tcArr1, "check member groupIDs")
	table = newGroupTable(mock.exec.GetLocalDB())
	tcArr1 = []*tableCase{{
		index: 0,
		key:   []byte(groupID),
		expectData: &vty.GroupInfo{ID: groupID, Name: "test", MemberNum: 1,
			Admins: []string{testAddrs[0]}, Creator: testAddrs[0], Members: members1},
	}, {
		index: 1,
		key:   []byte(groupID2),
		expectData: &vty.GroupInfo{ID: groupID2, Name: "test", MemberNum: 1,
			Admins: []string{testAddrs[0]}, Creator: testAddrs[0], Members: members2},
	}}
	testTableData(t, table, tcArr1, "check groupInfo")
}

func TestVote_ExecLocal_UpdateGroup(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	members := []*vty.GroupMember{{Addr: testAddrs[2], VoteWeight: 1}}
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "test"},
	}, {
		index: 1,
		payload: &vty.CreateVote{Name: "v1", GroupID: groupID, VoteOptions: []string{"A", "B"},
			BeginTimestamp: testBlockTime, EndTimestamp: testBlockTime + 1},
	}, {
		index:   2,
		payload: &vty.UpdateGroup{GroupID: groupID, RemoveAdmins: []string{testAddrs[0]}, AddAdmins: []string{testAddrs[1]}},
	}, {
		index:   3,
		priv:    privKeys[1],
		payload: &vty.UpdateGroup{GroupID: groupID, RemoveMembers: testAddrs, AddMembers: members},
	}}
	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])
	table := newMemberTable(mock.exec.GetLocalDB())
	tcArr1 := []*tableCase{{
		index:      0,
		key:        []byte(testAddrs[0]),
		expectData: &vty.MemberInfo{Addr: testAddrs[0]},
	}, {
		index:      1,
		key:        []byte(testAddrs[1]),
		expectData: &vty.MemberInfo{Addr: testAddrs[1], GroupIDs: []string{groupID}},
	}, {
		index:      2,
		key:        []byte(testAddrs[2]),
		expectData: &vty.MemberInfo{Addr: testAddrs[2], GroupIDs: []string{groupID}},
	}}
	testTableData(t, table, tcArr1, "check member groupIDs")
	table = newGroupTable(mock.exec.GetLocalDB())
	expectInfo := &vty.GroupInfo{ID: groupID, Name: "test", Admins: []string{testAddrs[1]},
		Members: members, MemberNum: 1, Creator: testAddrs[0], VoteNum: 1}
	testTableData(t, table, []*tableCase{{
		index:      0,
		key:        []byte(groupID),
		expectData: expectInfo,
	}, {
		index:        1,
		key:          []byte("testid"),
		expectGetErr: types.ErrNotFound,
	}}, "check group Info")

	tx := util.CreateNoneTx(mock.cfg, privKeys[0])
	group, err := newAction(mock.exec, tx, 0).getGroupInfo(groupID)
	require.Nil(t, err)
	group.VoteNum = 1
	require.Equal(t, expectInfo.String(), group.String())
}

func TestVote_ExecLocal_CreateVote(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	voteID := formatVoteID(dapp.HeightIndexStr(testHeight, 1))
	options := []*vty.VoteOption{{Option: "A"}, {Option: "B"}}
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "g1"},
	}, {
		index: 1,
		payload: &vty.CreateVote{Name: "v1", GroupID: groupID, VoteOptions: []string{"A", "B"},
			BeginTimestamp: testBlockTime, EndTimestamp: testBlockTime + 1},
	}, {
		index: 2,
		payload: &vty.CreateVote{Name: "v2", GroupID: groupID, VoteOptions: []string{"A", "B"},
			BeginTimestamp: testBlockTime, EndTimestamp: testBlockTime + 1},
	}, {
		index: 3,
		payload: &vty.CreateVote{Name: "v3", GroupID: groupID, VoteOptions: []string{"A", "B"},
			BeginTimestamp: testBlockTime, EndTimestamp: testBlockTime + 1},
	}}
	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])

	table := newVoteTable(mock.exec.GetLocalDB())
	expectVoteInfo := &vty.VoteInfo{
		Name: "v1", VoteOptions: options, BeginTimestamp: testBlockTime, EndTimestamp: testBlockTime + 1,
		GroupID: groupID, ID: voteID, Creator: testAddrs[0], GroupName: "g1",
	}
	testTableData(t, table, []*tableCase{{
		index:      0,
		key:        []byte(voteID),
		expectData: expectVoteInfo,
	}}, "check vote Info")

	table = newGroupTable(mock.exec.GetLocalDB())
	row, err := table.GetData([]byte(groupID))
	require.Nil(t, err)
	info, _ := row.Data.(*vty.GroupInfo)
	tx := util.CreateNoneTx(mock.cfg, privKeys[0])
	group, err := newAction(mock.exec, tx, 0).getGroupInfo(groupID)
	require.Nil(t, err)
	group.VoteNum = 3
	require.Equal(t, group.String(), info.String())
}

func TestVote_ExecLocal_CloseVote(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	voteID := formatVoteID(dapp.HeightIndexStr(testHeight, 1))
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "test"},
	}, {
		index: 1,
		payload: &vty.CreateVote{Name: "test", GroupID: groupID, VoteOptions: []string{"A", "B"},
			BeginTimestamp: testBlockTime, EndTimestamp: testBlockTime + 1},
	}, {
		index:   2,
		payload: &vty.CloseVote{VoteID: voteID},
	}}
	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])

	table := newVoteTable(mock.exec.GetLocalDB())
	row, err := table.GetData([]byte(voteID))
	require.Nil(t, err)
	info, _ := row.Data.(*vty.VoteInfo)
	require.Equal(t, uint32(voteStatusClosed), info.Status)
	tx := util.CreateNoneTx(mock.cfg, privKeys[0])
	vote, err := newAction(mock.exec, tx, 0).getVoteInfo(voteID)
	require.Nil(t, err)
	vote.GroupName = "test"
	require.Equal(t, vote.String(), info.String())
}

func TestVote_ExecLocal_CommitVote(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	voteID := formatVoteID(dapp.HeightIndexStr(testHeight, 1))
	members := []*vty.GroupMember{{Addr: testAddrs[0], VoteWeight: 1}}
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "test", Members: members},
	}, {
		index: 1,
		payload: &vty.CreateVote{Name: "test", GroupID: groupID, VoteOptions: []string{"A", "B"},
			BeginTimestamp: testBlockTime, EndTimestamp: testBlockTime + 1},
	}, {
		index:   2,
		payload: &vty.CommitVote{VoteID: voteID},
	}}
	testExec(t, mock, testTypeExecLocal, tcArr, privKeys[0])

	table := newVoteTable(mock.exec.GetLocalDB())
	row, err := table.GetData([]byte(voteID))
	require.Nil(t, err)
	info, _ := row.Data.(*vty.VoteInfo)
	require.Equal(t, testAddrs[0], info.CommitInfos[0].Addr)
	require.Equal(t, uint32(1), info.VoteOptions[0].Score)
	tx := util.CreateNoneTx(mock.cfg, privKeys[0])
	vote, err := newAction(mock.exec, tx, 0).getVoteInfo(voteID)
	require.Nil(t, err)
	vote.CommitInfos[0].TxHash = info.CommitInfos[0].TxHash
	vote.CommitInfos[0].VoteWeight = info.CommitInfos[0].VoteWeight
	vote.GroupName = "test"
	require.Equal(t, vote.String(), info.String())
}

func TestVote_ExecDelLocal(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	voteID := formatVoteID(dapp.HeightIndexStr(testHeight, 1))
	tcArr := []*testcase{{
		index:    0,
		payload:  &vty.CreateGroup{Name: "test"},
		execType: testTypeExecLocal,
	}, {
		index: 1,
		payload: &vty.CreateVote{Name: "test", GroupID: groupID, VoteOptions: []string{"A", "B"},
			BeginTimestamp: testBlockTime, EndTimestamp: testBlockTime + 1},
	}}
	testExec(t, mock, testTypeExecDelLocal, tcArr, privKeys[0])

	table := newVoteTable(mock.exec.GetLocalDB())
	_, err := table.GetData([]byte(voteID))
	require.Equal(t, types.ErrDecode, err)
}
