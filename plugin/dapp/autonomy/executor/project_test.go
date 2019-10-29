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
)

func TestExecLocalProject(t *testing.T) {
	testexecLocalProject(t, false)
	testexecLocalProject(t, true)
}

func testexecLocalProject(t *testing.T, auto bool) {
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

	var set *types.LocalDBSet
	var err error
	if !auto {
		set, err = au.execLocalProject(receipt)
		assert.NoError(t, err)
		assert.NotNil(t, set)
	} else {
		tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
		assert.NoError(t, err)
		set, err = au.execAutoLocalProject(tx, receipt)
		assert.NoError(t, err)
		assert.NotNil(t, set)
	}

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

	if !auto {
		set, err = au.execLocalProject(&types.ReceiptData{
			Logs: []*types.ReceiptLog{
				{Ty: auty.TyLogRvkPropProject, Log: types.Encode(receiptProject1)},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	} else {
		tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
		assert.NoError(t, err)
		set, err = au.execAutoLocalProject(tx,
			&types.ReceiptData{
				Logs: []*types.ReceiptLog{
					{Ty: auty.TyLogRvkPropProject, Log: types.Encode(receiptProject1)},
				},
			})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	}

	//save to database
	saveKvs(sdb, set.KV)

	// check
	checkExecLocalProject(t, kvdb, cur)

	// TyLogVotePropProject
	cur.Status = auty.AutonomyStatusProposalProject
	pre2 := copyAutonomyProposalProject(cur)
	cur.Status = auty.AutonomyStatusVotePropProject
	cur.Address = "2222222222222"
	receiptProject2 := &auty.ReceiptProposalProject{
		Prev:    pre2,
		Current: cur,
	}

	if !auto {
		set, err = au.execLocalProject(&types.ReceiptData{
			Logs: []*types.ReceiptLog{
				{Ty: auty.TyLogVotePropProject, Log: types.Encode(receiptProject2)},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	} else {
		tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
		assert.NoError(t, err)
		set, err = au.execAutoLocalProject(tx,
			&types.ReceiptData{
				Logs: []*types.ReceiptLog{
					{Ty: auty.TyLogVotePropProject, Log: types.Encode(receiptProject2)},
				},
			})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	}

	//save to database
	saveKvs(sdb, set.KV)

	// check
	checkExecLocalProject(t, kvdb, cur)
}

func TestExecDelLocalProject(t *testing.T) {
	testexecDelLocalProject(t)
}

func testexecDelLocalProject(t *testing.T) {
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
	tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
	assert.NoError(t, err)
	set, err := au.execAutoLocalProject(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	set, err = au.execAutoDelLocal(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	// check
	table := NewProjectTable(au.GetLocalDB())
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
	receiptProject2 := &auty.ReceiptProposalProject{
		Prev:    pre1,
		Current: cur,
	}
	recpt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogVotePropProject, Log: types.Encode(receiptProject2)},
		}}
	// 先执行local然后进行删除
	// 自动回退测试时候，需要先设置一个前置状态
	tx, err = types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
	assert.NoError(t, err)
	set, err = au.execAutoLocalProject(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	// 正常测试退回
	tx, err = types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
	assert.NoError(t, err)
	set, err = au.execAutoLocalProject(tx, recpt)

	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)
	// check
	checkExecLocalProject(t, kvdb, cur)

	set, err = au.execAutoDelLocal(tx, recpt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	checkExecLocalProject(t, kvdb, pre1)
}

func TestGetProposalProject(t *testing.T) {
	au := newTestAutonomy()
	_, storedb, _ := util.CreateTestDB()
	au.SetStateDB(storedb)
	tx := "1111111111111111111"
	storedb.Set(propProjectID(tx), types.Encode(&auty.AutonomyProposalProject{}))
	rsp, err := au.getProposalProject(&types.ReqString{Data: tx})
	assert.NoError(t, err)
	assert.NotNil(t, rsp)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), 1)
}

func TestListProposalProject(t *testing.T) {
	au := newTestAutonomy()
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
		assert.NoError(t, err)
		kv, err := table.Save()
		assert.NoError(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), len(testcase2))
	k := 2
	for _, tcase := range testcase2 {
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[k].Height, tcase.height)
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[k].Index, int32(tcase.index))
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
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), len(testcase2))
	for i, tcase := range testcase2 {
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[i].Height, tcase.height)
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[i].Index, int32(tcase.index))
	}

	// 翻页查找
	req = &auty.ReqQueryProposalProject{
		Status:    auty.AutonomyStatusProposalProject,
		Count:     1,
		Direction: 0,
		Index:     -1,
	}
	rsp, err = au.listProposalProject(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), 1)
	height := rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Height
	index := rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Index
	assert.Equal(t, height, testcase2[2].height)
	assert.Equal(t, index, int32(testcase2[2].index))
	//
	req = &auty.ReqQueryProposalProject{
		Status:    auty.AutonomyStatusProposalProject,
		Count:     10,
		Direction: 0,
		Height:    height,
		Index:     index,
	}
	rsp, err = au.listProposalProject(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalProject).PropProjects), 2)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Height, testcase2[1].height)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[0].Index, int32(testcase2[1].index))
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[1].Height, testcase2[0].height)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalProject).PropProjects[1].Index, int32(testcase2[0].index))
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
	assert.Equal(t, prop.Status, cur.Status)
	assert.Equal(t, prop.Address, cur.Address)
	assert.Equal(t, prop.Height, cur.Height)
	assert.Equal(t, prop.Index, cur.Index)

}
