// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/stretchr/testify/require"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/util"
)

func TestExecLocalProject(t *testing.T) {
	au := &Autonomy{}
	//TyLogPropProject
	cur := &auty.AutonomyProposalProject{
		PropProject: &auty.ProposalProject{},
		CurRule: &auty.RuleConfig{},
		Boards: []string{"111", "222", "333"},
		BoardVoteRes: &auty.VoteResult{},
		PubVote: &auty.PublicVote{},
		Status: auty.AutonomyStatusProposalProject,
		Address: "11111111111111",
		Height: 1,
		Index: 2,
	}
	receiptProject := &auty.ReceiptProposalProject{
		Prev: nil,
		Current: cur,
	}
	receipt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogPropProject, Log:types.Encode(receiptProject)},
		},
	}
	set, err := au.execLocalProject(receipt)
	require.NoError(t, err)
	require.NotNil(t, set)
	require.Equal(t, set.KV[0].Key, calcProjectKey4StatusHeight(cur.Status,
		dapp.HeightIndexStr(cur.Height, int64(cur.Index))))

	// TyLogRvkPropProject
	pre1 := copyAutonomyProposalProject(cur)
	cur.Status = auty.AutonomyStatusRvkPropProject
	cur.Height = 2
	cur.Index = 3
	receiptProject1 := &auty.ReceiptProposalProject{
		Prev: pre1,
		Current: cur,
	}
	set, err = au.execLocalProject(&types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogRvkPropProject, Log:types.Encode(receiptProject1)},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, set)
	require.Equal(t, set.KV[0].Key, calcProjectKey4StatusHeight(pre1.Status,
		dapp.HeightIndexStr(pre1.Height, int64(pre1.Index))))
	require.Equal(t, set.KV[0].Value, []byte(nil))
	require.Equal(t, set.KV[1].Key, calcProjectKey4StatusHeight(cur.Status,
		dapp.HeightIndexStr(cur.Height, int64(cur.Index))))

	// TyLogVotePropProject
	cur.Status = auty.AutonomyStatusProposalProject
	cur.Height = 1
	cur.Index = 2
	pre2 := copyAutonomyProposalProject(cur)
	cur.Status = auty.AutonomyStatusVotePropProject
	cur.Height = 2
	cur.Index = 3
	receiptProject2 := &auty.ReceiptProposalProject{
		Prev: pre2,
		Current: cur,
	}
	set, err = au.execLocalProject(&types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogVotePropProject, Log:types.Encode(receiptProject2)},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, set)
	require.Equal(t, set.KV[0].Key, calcProjectKey4StatusHeight(pre2.Status,
		dapp.HeightIndexStr(pre1.Height, int64(pre2.Index))))
	require.Equal(t, set.KV[0].Value, []byte(nil))
	require.Equal(t, set.KV[1].Key, calcProjectKey4StatusHeight(cur.Status,
		dapp.HeightIndexStr(cur.Height, int64(cur.Index))))
}

func TestExecDelLocalProject(t *testing.T) {
	au := &Autonomy{}
	//TyLogPropProject
	cur := &auty.AutonomyProposalProject{
		PropProject: &auty.ProposalProject{},
		CurRule: &auty.RuleConfig{},
		Boards: []string{"111", "222", "333"},
		BoardVoteRes: &auty.VoteResult{},
		PubVote: &auty.PublicVote{},
		Status: auty.AutonomyStatusProposalProject,
		Address: "11111111111111",
		Height: 1,
		Index: 2,
	}
	receiptProject := &auty.ReceiptProposalProject{
		Prev: nil,
		Current: cur,
	}
	receipt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogPropProject, Log:types.Encode(receiptProject)},
		},
	}
	set, err := au.execDelLocalProject(receipt)
	require.NoError(t, err)
	require.NotNil(t, set)
	require.Equal(t, set.KV[0].Key, calcProjectKey4StatusHeight(cur.Status,
		dapp.HeightIndexStr(cur.Height, int64(cur.Index))))
	require.Equal(t, set.KV[0].Value, []byte(nil))

	// TyLogVotePropProject
	pre1 := copyAutonomyProposalProject(cur)
	cur.Status = auty.AutonomyStatusVotePropProject
	cur.Height = 2
	cur.Index = 3
	receiptProject2 := &auty.ReceiptProposalProject{
		Prev: pre1,
		Current: cur,
	}
	set, err = au.execDelLocalProject(&types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogVotePropProject, Log:types.Encode(receiptProject2)},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, set)
	require.Equal(t, set.KV[0].Key, calcProjectKey4StatusHeight(cur.Status,
		dapp.HeightIndexStr(cur.Height, int64(cur.Index))))
	require.Equal(t, set.KV[0].Value, []byte(nil))
	require.Equal(t, set.KV[1].Key, calcProjectKey4StatusHeight(pre1.Status,
		dapp.HeightIndexStr(pre1.Height, int64(pre1.Index))))
	require.NotNil(t, set.KV[1].Value)
}

func TestGetProposalProject(t *testing.T) {
	au := &Autonomy{
		dapp.DriverBase{},
	}
	_, storedb, _ := util.CreateTestDB()
	au.SetStateDB(storedb)
	tx := "1111111111111111111"
	storedb.Set(propProjectID(tx), types.Encode(&auty.AutonomyProposalProject{}))
	rsp, err := au.getProposalProject(&types.ReqString{Data:tx})
	require.NoError(t, err)
	require.NotNil(t, rsp)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), 1)
}

func TestListProposalProject(t *testing.T) {
	au := &Autonomy{
		dapp.DriverBase{},
	}
	_, _, kvdb := util.CreateTestDB()
	au.SetLocalDB(kvdb)

	type statu struct {
		status int32
		height int64
		index  int64
	}

	testcase1 := []statu{
		{auty.AutonomyStatusRvkPropProject, 10, 2},
		{auty.AutonomyStatusVotePropProject, 15, 1},
		{auty.AutonomyStatusTmintPropProject, 20, 1},
	}
	testcase2 := []statu{
		{auty.AutonomyStatusProposalProject, 10, 1},
		{auty.AutonomyStatusProposalProject, 20, 2},
		{auty.AutonomyStatusProposalProject, 20, 5},
	}
	var testcase []statu
	testcase = append(testcase, testcase1...)
	testcase = append(testcase, testcase2...)
	cur := &auty.AutonomyProposalProject{
		PropProject: &auty.ProposalProject{},
		CurRule: &auty.RuleConfig{},
		Boards: []string{"111", "222", "333"},
		BoardVoteRes: &auty.VoteResult{},
		PubVote: &auty.PublicVote{},
		Status: auty.AutonomyStatusProposalProject,
		Address: "11111111111111",
		Height: 1,
		Index: 2,
	}
	for _, tcase := range testcase {
		key := calcProjectKey4StatusHeight(tcase.status,
			dapp.HeightIndexStr(tcase.height, int64(tcase.index)))
		cur.Status = tcase.status
		cur.Height = tcase.height
		cur.Index = int32(tcase.index)
		value := types.Encode(cur)
		kvdb.Set(key, value)
	}

	// 反向查找
	req := &auty.ReqQueryProposalProject{
		Status:auty.AutonomyStatusProposalProject,
		Count:10,
		Direction:0,
		Index: -1,
	}
	rsp, err := au.listProposalProject(req)
	require.NoError(t, err)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), len(testcase2))
	k := 2
	for _, tcase := range testcase2 {
		require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[k].Height, tcase.height)
		require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[k].Index, int32(tcase.index))
		k--
	}

	// 正向查找
	req = &auty.ReqQueryProposalProject{
		Status:auty.AutonomyStatusProposalProject,
		Count:10,
		Direction:1,
		Index: -1,
	}
	rsp, err = au.listProposalProject(req)
	require.NoError(t, err)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), len(testcase2))
	for i, tcase := range testcase2 {
		require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[i].Height, tcase.height)
		require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[i].Index, int32(tcase.index))
	}

	// 翻页查找
	req = &auty.ReqQueryProposalProject{
		Status:auty.AutonomyStatusProposalProject,
		Count:1,
		Direction:0,
		Index: -1,
	}
	rsp, err = au.listProposalProject(req)
	require.NoError(t, err)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), 1)
	height := rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Height
	index := rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Index
	require.Equal(t, height, testcase2[2].height)
	require.Equal(t, index, int32(testcase2[2].index))
	//
	Index := height*types.MaxTxsPerBlock + int64(index)
	req = &auty.ReqQueryProposalProject{
		Status:auty.AutonomyStatusProposalProject,
		Count:10,
		Direction:0,
		Index: Index,
	}
	rsp, err = au.listProposalProject(req)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), 2)
	require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Height, testcase2[1].height)
	require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Index, int32(testcase2[1].index))
	require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[1].Height, testcase2[0].height)
	require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[1].Index, int32(testcase2[0].index))
}