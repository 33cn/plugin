package executor

import (
	"testing"

	"github.com/33cn/chain33/system/dapp"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/stretchr/testify/require"
)

const (
	testTypeCheckTx = iota + 1
	testTypeExec
	testTypeExecLocal
	testTypeExecDelLocal
)

func testExec(t *testing.T, mock *testExecMock, testExecType int, tcArr []*testcase, priv crypto.PrivKey) {

	if mock == nil {
		mock = &testExecMock{}
		mock.InitEnv()
		defer mock.FreeEnv()
	}
	exec := mock.exec
	for i, tc := range tcArr {

		signPriv := priv
		if tc.priv != nil {
			signPriv = tc.priv
		}
		tx, err := createTx(mock, tc.payload, signPriv)
		require.NoErrorf(t, err, "createTxErr, testIndex=%d", tc.index)
		if err != nil {
			continue
		}
		err = exec.CheckTx(tx, i)
		require.Equalf(t, tc.expectCheckErr, err, "checkTx err index %d", tc.index)
		execType := testExecType
		if tc.execType > 0 {
			execType = tc.execType
		}
		if execType == testTypeCheckTx {
			continue
		}

		recp, err := exec.Exec(tx, i)
		recpData := &types.ReceiptData{
			Ty:   recp.GetTy(),
			Logs: recp.GetLogs(),
		}
		if err == nil && len(recp.GetKV()) > 0 {
			util.SaveKVList(mock.stateDB, recp.KV)
		}
		require.Equalf(t, tc.expectExecErr, err, "execTx err index %d", tc.index)
		if execType == testTypeExec {
			continue
		}
		kvSet, err := exec.ExecLocal(tx, recpData, i)
		for _, kv := range kvSet.GetKV() {
			err := mock.localDB.Set(kv.Key, kv.Value)
			require.Nil(t, err)
		}
		require.Equalf(t, tc.expectExecLocalErr, err, "execLocalTx err index %d", tc.index)

		if execType == testTypeExecLocal {
			continue
		}

		kvSet, err = exec.ExecDelLocal(tx, recpData, i)
		for _, kv := range kvSet.GetKV() {
			err := mock.localDB.Set(kv.Key, kv.Value)
			require.Nil(t, err)
		}
		require.Equalf(t, tc.expectExecDelErr, err, "execDelLocalTx err index %d", tc.index)
	}
}

func TestVote_Exec(t *testing.T) {

	groupID := formatGroupID(dapp.HeightIndexStr(testHeight, 0))
	voteID := formatVoteID(dapp.HeightIndexStr(testHeight, 1))
	tcArr := []*testcase{{
		index:   0,
		payload: &vty.CreateGroup{Name: "test", Members: []*vty.GroupMember{{Addr: testAddrs[0]}}},
	}, {
		index: 1,
		payload: &vty.CreateVote{
			Name:         "vote",
			GroupID:      groupID,
			EndTimestamp: testBlockTime + 1,
			VoteOptions:  []string{"A", "B"},
		},
	}, {
		index:   2,
		payload: &vty.UpdateGroup{GroupID: groupID, RemoveAdmins: testAddrs, AddAdmins: []string{testAddrs[1]}},
	}, {
		index:   3,
		payload: &vty.CommitVote{VoteID: voteID},
	}, {
		index:          4,
		payload:        &vty.UpdateGroup{GroupID: groupID, AddAdmins: []string{testAddrs[0]}},
		expectCheckErr: errAddrPermissionDenied,
		execType:       testTypeCheckTx,
	}, {
		index:   5,
		payload: &vty.UpdateGroup{GroupID: groupID, RemoveAdmins: testAddrs, AddAdmins: []string{testAddrs[0]}},
		priv:    privKeys[1],
	}, {
		index:   6,
		payload: &vty.UpdateGroup{GroupID: groupID, AddMembers: []*vty.GroupMember{{Addr: testAddrs[1]}}, RemoveMembers: testAddrs},
	}, {
		index:          7,
		payload:        &vty.CommitVote{VoteID: voteID},
		expectCheckErr: errAddrPermissionDenied,
		execType:       testTypeCheckTx,
	}, {
		index:   8,
		payload: &vty.CommitVote{VoteID: voteID},
		priv:    privKeys[1],
	}, {
		index:   9,
		payload: &vty.CloseVote{VoteID: voteID},
	}, {
		index:          10,
		payload:        &vty.CloseVote{VoteID: voteID},
		execType:       testTypeCheckTx,
		expectCheckErr: errVoteAlreadyClosed,
	}, {
		index:   11,
		payload: &vty.UpdateMember{Name: "testName"},
	},
	}

	testExec(t, nil, testTypeExec, tcArr, privKeys[0])
}
