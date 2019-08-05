// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecLocalProject(t *testing.T) {
	_, sdb, kvdb := util.CreateTestDB()
	au := &Autonomy{}
	au.SetLocalDB(kvdb)
	//TyLogPropProject
	cur := &auty.AutonomyProposalProject{
		PropProject:  &auty.ProposalProject{},
		CurRule:      &auty.RuleConfig{},
		Boards:       []string{"111", "222", "333"},
		BoardVoteRes: &auty.VoteResult{},
		PubVote:      &auty.PublicVote{},
		Status:       auty.AutonomyStatusProposalProject,
		Address:      "11111111111111",
		Height:       1,
		Index:        2,
	}
	receiptProject := &auty.ReceiptProposalProject{
		Prev:    nil,
		Current: cur,
	}
	receipt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogPropProject, Log: types.Encode(receiptProject)},
		},
	}
	set, err := au.execLocalProject(receipt)
	require.NoError(t, err)
	require.NotNil(t, set)
	//save to database
	saveKvs(sdb, set.KV)

	// check
	checkExecLocalProject(t, kvdb, cur)

	// TyLogRvkPropProject
	pre1 := copyAutonomyProposalProject(cur)
	cur.Status = auty.AutonomyStatusRvkPropProject
	receiptProject1 := &auty.ReceiptProposalProject{
		Prev:    pre1,
		Current: cur,
	}
	set, err = au.execLocalProject(&types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogRvkPropProject, Log: types.Encode(receiptProject1)},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, set)

	//save to database
	saveKvs(sdb, set.KV)

	// check
	checkExecLocalProject(t, kvdb, cur)

	// TyLogVotePropProject
	cur.Status = auty.AutonomyStatusProposalProject
	cur.Height = 1
	cur.Index = 2
	pre2 := copyAutonomyProposalProject(cur)
	cur.Status = auty.AutonomyStatusVotePropProject
	cur.Height = 1
	cur.Index = 2
	cur.Address = "2222222222222"
	receiptProject2 := &auty.ReceiptProposalProject{
		Prev:    pre2,
		Current: cur,
	}
	set, err = au.execLocalProject(&types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogVotePropProject, Log: types.Encode(receiptProject2)},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, set)

	//save to database
	saveKvs(sdb, set.KV)

	// check
	checkExecLocalProject(t, kvdb, cur)
}

func TestExecDelLocalProject(t *testing.T) {
	_, sdb, kvdb := util.CreateTestDB()
	au := &Autonomy{}
	au.SetLocalDB(kvdb)
	//TyLogPropProject
	cur := &auty.AutonomyProposalProject{
		PropProject:  &auty.ProposalProject{},
		CurRule:      &auty.RuleConfig{},
		Boards:       []string{"111", "222", "333"},
		BoardVoteRes: &auty.VoteResult{},
		PubVote:      &auty.PublicVote{},
		Status:       auty.AutonomyStatusProposalProject,
		Address:      "11111111111111",
		Height:       1,
		Index:        2,
	}
	receiptProject := &auty.ReceiptProposalProject{
		Prev:    nil,
		Current: cur,
	}
	receipt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogPropProject, Log: types.Encode(receiptProject)},
		},
	}
	// 先执行local然后进行删除
	set, err := au.execLocalProject(receipt)
	require.NoError(t, err)
	require.NotNil(t, set)
	saveKvs(sdb, set.KV)

	set, err = au.execDelLocalProject(receipt)
	require.NoError(t, err)
	require.NotNil(t, set)
	saveKvs(sdb, set.KV)

	// check
	table := NewBoardTable(au.GetLocalDB())
	query := table.GetQuery(kvdb)
	_, err = query.ListIndex("primary", nil, nil, 10, 0)
	assert.Equal(t, err, types.ErrNotFound)
	_, err = query.ListIndex("addr", nil, nil, 10, 0)
	assert.Equal(t, err, types.ErrNotFound)
	_, err = query.ListIndex("status", nil, nil, 10, 0)
	assert.Equal(t, err, types.ErrNotFound)
	_, err = query.ListIndex("addr_status", nil, nil, 10, 0)
	assert.Equal(t, err, types.ErrNotFound)

	// TyLogVotePropProject
	pre1 := copyAutonomyProposalProject(cur)
	cur.Status = auty.AutonomyStatusVotePropProject
	cur.Height = 1
	cur.Index = 2
	receiptProject2 := &auty.ReceiptProposalProject{
		Prev:    pre1,
		Current: cur,
	}
	// 先执行local然后进行删除
	set, err = au.execLocalProject(&types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogVotePropProject, Log: types.Encode(receiptProject2)},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, set)
	saveKvs(sdb, set.KV)
	// check
	checkExecLocalProject(t, kvdb, cur)
	set, err = au.execDelLocalProject(&types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogVotePropProject, Log: types.Encode(receiptProject2)},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, set)
	saveKvs(sdb, set.KV)
	// check
	checkExecLocalProject(t, kvdb, pre1)
}

func TestGetProposalProject(t *testing.T) {
	au := &Autonomy{
		dapp.DriverBase{},
	}
	_, storedb, _ := util.CreateTestDB()
	au.SetStateDB(storedb)
	tx := "1111111111111111111"
	storedb.Set(propProjectID(tx), types.Encode(&auty.AutonomyProposalProject{}))
	rsp, err := au.getProposalProject(&types.ReqString{Data: tx})
	require.NoError(t, err)
	require.NotNil(t, rsp)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), 1)
}

func TestListProposalProject(t *testing.T) {
	au := &Autonomy{
		dapp.DriverBase{},
	}
	_, sdb, kvdb := util.CreateTestDB()
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
		PropProject:  &auty.ProposalProject{},
		CurRule:      &auty.RuleConfig{},
		Boards:       []string{"111", "222", "333"},
		BoardVoteRes: &auty.VoteResult{},
		PubVote:      &auty.PublicVote{},
		Status:       auty.AutonomyStatusProposalProject,
		Address:      "11111111111111",
		Height:       1,
		Index:        2,
	}

	//将数据保存下去
	var kvs []*types.KeyValue
	table := NewProjectTable(kvdb)
	for _, tcase := range testcase {
		cur.Status = tcase.status
		cur.Height = tcase.height
		cur.Index = int32(tcase.index)

		err := table.Replace(cur)
		require.NoError(t, err)
		kv, err := table.Save()
		require.NoError(t, err)
		kvs = append(kvs, kv...)
	}
	saveKvs(sdb, kvs)

	// 反向查找
	req := &auty.ReqQueryProposalProject{
		Status:    auty.AutonomyStatusProposalProject,
		Count:     10,
		Direction: 0,
		Index:     -1,
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
		Status:    auty.AutonomyStatusProposalProject,
		Count:     10,
		Direction: 1,
		Index:     -1,
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
		Status:    auty.AutonomyStatusProposalProject,
		Count:     1,
		Direction: 0,
		Index:     -1,
	}
	rsp, err = au.listProposalProject(req)
	require.NoError(t, err)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), 1)
	height := rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Height
	index := rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Index
	require.Equal(t, height, testcase2[2].height)
	require.Equal(t, index, int32(testcase2[2].index))
	//
	req = &auty.ReqQueryProposalProject{
		Status:    auty.AutonomyStatusProposalProject,
		Count:     10,
		Direction: 0,
		Height:    height,
		Index:     index,
	}
	rsp, err = au.listProposalProject(req)
	require.NoError(t, err)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), 2)
	require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Height, testcase2[1].height)
	require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Index, int32(testcase2[1].index))
	require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[1].Height, testcase2[0].height)
	require.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[1].Index, int32(testcase2[0].index))
}

func checkExecLocalProject(t *testing.T, kvdb db.KVDB, cur *auty.AutonomyProposalProject) {
	table := NewProjectTable(kvdb)
	query := table.GetQuery(kvdb)

	rows, err := query.ListIndex("primary", nil, nil, 10, 0)
	assert.Equal(t, err, nil)
	assert.Equal(t, string(rows[0].Primary), dapp.HeightIndexStr(1, 2))

	rows, err = query.ListIndex("addr", nil, nil, 10, 0)
	assert.Equal(t, err, nil)
	assert.Equal(t, 1, len(rows))

	rows, err = query.ListIndex("status", nil, nil, 10, 0)
	assert.Equal(t, err, nil)
	assert.Equal(t, 1, len(rows))

	rows, err = query.ListIndex("addr_status", nil, nil, 10, 0)
	assert.Equal(t, err, nil)
	assert.Equal(t, 1, len(rows))

	prop, ok := rows[0].Data.(*auty.AutonomyProposalProject)
	assert.Equal(t, true, ok)
	assert.Equal(t, prop, cur)

}
