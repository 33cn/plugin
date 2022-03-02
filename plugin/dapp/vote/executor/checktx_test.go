package executor

import (
	"encoding/hex"
	"testing"

	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"

	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
)

func TestVote_CheckTx_CreateGroup(t *testing.T) {

	tcArr := []*testcase{{
		index:          1,
		payload:        &vty.CreateGroup{},
		expectCheckErr: errEmptyName,
	}, {
		index:          2,
		payload:        &vty.CreateGroup{Name: "test", Members: []*vty.GroupMember{{}}},
		expectCheckErr: types.ErrInvalidAddress,
	}, {
		index: 3,
		payload: &vty.CreateGroup{
			Name: "test",
			Members: []*vty.GroupMember{{
				Addr: testAddrs[0],
			}, {
				Addr: testAddrs[0],
			}},
		},
		expectCheckErr: errDuplicateMember,
	}, {
		index:          4,
		payload:        &vty.CreateGroup{Name: "test", Admins: []string{testAddrs[0], testAddrs[0]}},
		expectCheckErr: errDuplicateAdmin,
	}, {
		index:   5,
		payload: &vty.CreateGroup{Name: "test"},
	},
	}

	testExec(t, nil, testTypeCheckTx, tcArr, privKeys[0])
}

func TestVote_CheckTx_UpdateGroup(t *testing.T) {

	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))

	tcArr := []*testcase{{
		index:    0,
		payload:  &vty.CreateGroup{Name: "test"},
		execType: testTypeExecLocal,
	}, {
		index:          1,
		payload:        &vty.UpdateGroup{},
		expectCheckErr: errGroupNotExist,
	}, {
		index:          2,
		payload:        &vty.UpdateGroup{GroupID: groupID},
		priv:           privKeys[1],
		expectCheckErr: errAddrPermissionDenied,
	}, {
		index:          3,
		payload:        &vty.UpdateGroup{GroupID: groupID, RemoveAdmins: testAddrs[:]},
		expectCheckErr: errAddrPermissionDenied,
	}, {
		index:          4,
		payload:        &vty.UpdateGroup{GroupID: groupID, AddMembers: []*vty.GroupMember{{Addr: "errAddr"}}},
		expectCheckErr: types.ErrInvalidAddress,
	}, {
		index:          5,
		payload:        &vty.UpdateGroup{GroupID: groupID, AddAdmins: []string{"errAddr"}},
		expectCheckErr: types.ErrInvalidAddress,
	}, {
		index:          6,
		payload:        &vty.UpdateGroup{GroupID: groupID, AddAdmins: []string{hex.EncodeToString(privKeys[0].PubKey().Bytes())}},
		expectCheckErr: types.ErrInvalidAddress,
	}, {
		index:   7,
		payload: &vty.UpdateGroup{GroupID: groupID, AddAdmins: []string{testAddrs[1]}},
	},
	}

	testExec(t, nil, testTypeCheckTx, tcArr, privKeys[0])
}

func TestVote_CheckTx_CreateVote(t *testing.T) {

	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))

	tcArr := []*testcase{{
		index:    0,
		payload:  &vty.CreateGroup{Name: "test"},
		execType: testTypeExecLocal,
	}, {
		index:          1,
		payload:        &vty.CreateVote{},
		expectCheckErr: errEmptyName,
	}, {
		index:          2,
		payload:        &vty.CreateVote{Name: "vote"},
		expectCheckErr: errGroupNotExist,
	}, {
		index:          3,
		payload:        &vty.CreateVote{Name: "vote", GroupID: groupID},
		priv:           privKeys[1],
		expectCheckErr: errAddrPermissionDenied,
	}, {
		index:          4,
		payload:        &vty.CreateVote{Name: "vote", GroupID: groupID},
		expectCheckErr: errInvalidVoteTime,
	}, {
		index: 5,
		payload: &vty.CreateVote{
			Name:           "vote",
			GroupID:        groupID,
			BeginTimestamp: testBlockTime + 1,
			EndTimestamp:   testBlockTime + 1,
		},
		expectCheckErr: errInvalidVoteTime,
	}, {
		index: 6,
		payload: &vty.CreateVote{
			Name:         "vote",
			GroupID:      groupID,
			EndTimestamp: testBlockTime + 1,
		},
		expectCheckErr: errInvalidVoteOption,
	}, {
		index: 7,
		payload: &vty.CreateVote{
			Name:         "vote",
			GroupID:      groupID,
			EndTimestamp: testBlockTime + 1,
			VoteOptions:  []string{"A", "B"},
		},
	},
	}

	testExec(t, nil, testTypeCheckTx, tcArr, privKeys[0])
}

func TestVote_CheckTx_CommitVote(t *testing.T) {

	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	voteID := formatVoteID(dapp.HeightIndexStr(testHeight, 1))
	tcArr := []*testcase{{
		index:    0,
		payload:  &vty.CreateGroup{Name: "test", Members: []*vty.GroupMember{{Addr: testAddrs[0]}}},
		execType: testTypeExecLocal,
	}, {
		index: 1,
		payload: &vty.CreateVote{
			Name:         "vote",
			GroupID:      groupID,
			EndTimestamp: testBlockTime + 1,
			VoteOptions:  []string{"A", "B"},
		},
		execType: testTypeExecLocal,
	}, {
		index:          2,
		payload:        &vty.CommitVote{},
		expectCheckErr: errVoteNotExist,
	}, {
		index:          3,
		payload:        &vty.CommitVote{VoteID: voteID, OptionIndex: 10},
		expectCheckErr: errInvalidOptionIndex,
	}, {
		index:          4,
		payload:        &vty.CommitVote{VoteID: voteID},
		priv:           privKeys[1],
		expectCheckErr: errAddrPermissionDenied,
	}, {
		index:    5,
		payload:  &vty.CommitVote{VoteID: voteID},
		execType: testTypeExecLocal,
	}, {
		index:          6,
		payload:        &vty.CommitVote{VoteID: voteID},
		expectCheckErr: errAddrAlreadyVoted,
	}, {
		index: 7,
		payload: &vty.CreateVote{
			Name:           "vote",
			GroupID:        groupID,
			BeginTimestamp: testBlockTime + 1,
			EndTimestamp:   testBlockTime + 2,
			VoteOptions:    []string{"A", "B"},
		},
		execType: testTypeExecLocal,
	}, {
		index:         8,
		payload:       &vty.CommitVote{VoteID: formatVoteID(dapp.HeightIndexStr(testHeight, 7))},
		execType:      testTypeExec,
		expectExecErr: errVoteNotStarted,
	}}

	testExec(t, nil, testTypeCheckTx, tcArr, privKeys[0])
}

func TestVote_CheckTx_CloseVote(t *testing.T) {

	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	voteID := formatVoteID(dapp.HeightIndexStr(testHeight, 1))
	tcArr := []*testcase{{
		index:    0,
		payload:  &vty.CreateGroup{Name: "test", Members: []*vty.GroupMember{{Addr: testAddrs[0]}}},
		execType: testTypeExecLocal,
	}, {
		index: 1,
		payload: &vty.CreateVote{
			Name:         "vote",
			GroupID:      groupID,
			EndTimestamp: testBlockTime + 1,
			VoteOptions:  []string{"A", "B"},
		},
		execType: testTypeExecLocal,
	}, {
		index:          2,
		payload:        &vty.CloseVote{},
		expectCheckErr: errVoteNotExist,
	}, {
		index:          3,
		payload:        &vty.CloseVote{VoteID: voteID},
		priv:           privKeys[1],
		expectCheckErr: errAddrPermissionDenied,
	}, {
		index:    4,
		payload:  &vty.CloseVote{VoteID: voteID},
		execType: testTypeExecLocal,
	}, {
		index:          5,
		payload:        &vty.CloseVote{VoteID: voteID},
		expectCheckErr: errVoteAlreadyClosed,
	},
	}

	testExec(t, nil, testTypeCheckTx, tcArr, privKeys[0])
}

func TestVote_CheckTx_UpdateMember(t *testing.T) {

	tcArr := []*testcase{{
		index:          0,
		payload:        &vty.UpdateMember{},
		expectCheckErr: errEmptyName,
	}, {
		index:   1,
		payload: &vty.UpdateMember{Name: "test"},
	},
	}

	testExec(t, nil, testTypeCheckTx, tcArr, privKeys[0])
}
