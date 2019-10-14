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
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/queue"
)

var chainTestCfg = types.NewChain33Config(types.GetDefaultCfgstring())

func newTestAutonomy() *Autonomy{
	au := &Autonomy{
		dapp.DriverBase{},
	}
	q := queue.New("channel")
	q.SetConfig(chainTestCfg)
	api, _ := client.New(q.Client(), nil)
	au.SetAPI(api)
	return au
}

func TestExecLocalBoard(t *testing.T) {
	testexecLocalBoard(t, false)
	testexecLocalBoard(t, true)
}

func testexecLocalBoard(t *testing.T, auto bool) {
	_, sdb, kvdb := util.CreateTestDB()
	au := &Autonomy{}
	au.SetLocalDB(kvdb)
	//TyLogPropBoard
	cur := &auty.AutonomyProposalBoard{
		PropBoard:  &auty.ProposalBoard{},
		CurRule:    &auty.RuleConfig{},
		VoteResult: &auty.VoteResult{},
		Status:     auty.AutonomyStatusProposalBoard,
		Address:    "11111111111111",
		Height:     1,
		Index:      2,
	}
	receiptBoard := &auty.ReceiptProposalBoard{
		Prev:    nil,
		Current: cur,
	}
	receipt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogPropBoard, Log: types.Encode(receiptBoard)},
		},
	}

	var set *types.LocalDBSet
	var err error
	if !auto {
		set, err = au.execLocalBoard(receipt)
		assert.NoError(t, err)
		assert.NotNil(t, set)
	} else {
		tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
		assert.NoError(t, err)
		set, err = au.execAutoLocalBoard(tx, receipt)
		assert.NoError(t, err)
		assert.NotNil(t, set)
	}

	//save to database
	saveKvs(sdb, set.KV)

	// check
	checkExecLocalBoard(t, kvdb, cur)

	// TyLogRvkPropBoard
	pre1 := copyAutonomyProposalBoard(cur)
	cur.Status = auty.AutonomyStatusRvkPropBoard
	receiptBoard1 := &auty.ReceiptProposalBoard{
		Prev:    pre1,
		Current: cur,
	}
	if !auto {
		set, err = au.execLocalBoard(&types.ReceiptData{
			Logs: []*types.ReceiptLog{
				{Ty: auty.TyLogRvkPropBoard, Log: types.Encode(receiptBoard1)},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	} else {
		tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
		assert.NoError(t, err)
		set, err = au.execAutoLocalBoard(tx,
			&types.ReceiptData{
				Logs: []*types.ReceiptLog{
					{Ty: auty.TyLogRvkPropBoard, Log: types.Encode(receiptBoard1)},
				},
			})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	}

	//save to database
	saveKvs(sdb, set.KV)

	// check
	checkExecLocalBoard(t, kvdb, cur)

	// TyLogVotePropBoard
	cur.Status = auty.AutonomyStatusProposalBoard
	pre2 := copyAutonomyProposalBoard(cur)
	cur.Status = auty.AutonomyStatusVotePropBoard
	cur.Address = "2222222222222"
	receiptBoard2 := &auty.ReceiptProposalBoard{
		Prev:    pre2,
		Current: cur,
	}
	if !auto {
		set, err = au.execLocalBoard(&types.ReceiptData{
			Logs: []*types.ReceiptLog{
				{Ty: auty.TyLogVotePropBoard, Log: types.Encode(receiptBoard2)},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	} else {
		tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
		assert.NoError(t, err)
		set, err = au.execAutoLocalBoard(tx,
			&types.ReceiptData{
				Logs: []*types.ReceiptLog{
					{Ty: auty.TyLogVotePropBoard, Log: types.Encode(receiptBoard2)},
				},
			})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	}

	//save to database
	saveKvs(sdb, set.KV)
	// check
	checkExecLocalBoard(t, kvdb, cur)
}

func TestExecDelLocalBoard(t *testing.T) {
	testexecDelLocalBoard(t)
}

func testexecDelLocalBoard(t *testing.T) {
	_, sdb, kvdb := util.CreateTestDB()
	au := &Autonomy{}
	au.SetLocalDB(kvdb)
	//TyLogPropBoard
	cur := &auty.AutonomyProposalBoard{
		PropBoard:  &auty.ProposalBoard{},
		CurRule:    &auty.RuleConfig{},
		VoteResult: &auty.VoteResult{},
		Status:     auty.AutonomyStatusProposalBoard,
		Address:    "11111111111111",
		Height:     1,
		Index:      2,
	}
	receiptBoard := &auty.ReceiptProposalBoard{
		Prev:    nil,
		Current: cur,
	}
	receipt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogPropBoard, Log: types.Encode(receiptBoard)},
		},
	}

	// 先执行local然后进行删除

	tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
	assert.NoError(t, err)
	set, err := au.execAutoLocalBoard(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	set, err = au.execAutoDelLocal(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
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

	// TyLogVotePropBoard
	pre1 := copyAutonomyProposalBoard(cur)
	cur.Status = auty.AutonomyStatusVotePropBoard
	receiptBoard2 := &auty.ReceiptProposalBoard{
		Prev:    pre1,
		Current: cur,
	}
	recpt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogVotePropBoard, Log: types.Encode(receiptBoard2)},
		}}
	// 先执行local然后进行删除

	// 自动回退测试时候，需要先设置一个前置状态
	tx, err = types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
	assert.NoError(t, err)
	set, err = au.execAutoLocalBoard(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	// 正常测试退回
	tx, err = types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
	assert.NoError(t, err)
	set, err = au.execAutoLocalBoard(tx, recpt)

	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)
	// check
	checkExecLocalBoard(t, kvdb, cur)

	set, err = au.execAutoDelLocal(tx, recpt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)
	// check
	checkExecLocalBoard(t, kvdb, pre1)
}

func TestGetProposalBoard(t *testing.T) {
	au := newTestAutonomy()
	_, storedb, _ := util.CreateTestDB()
	au.SetStateDB(storedb)
	tx := "1111111111111111111"
	storedb.Set(propBoardID(tx), types.Encode(&auty.AutonomyProposalBoard{}))
	rsp, err := au.getProposalBoard(&types.ReqString{Data: tx})
	assert.NoError(t, err)
	assert.NotNil(t, rsp)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalBoard).PropBoards), 1)
}

func TestListProposalBoard(t *testing.T) {
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
		{auty.AutonomyStatusRvkPropBoard, 10, 2},
		{auty.AutonomyStatusVotePropBoard, 15, 1},
		{auty.AutonomyStatusTmintPropBoard, 20, 1},
	}
	testcase2 := []statu{
		{auty.AutonomyStatusProposalBoard, 10, 1},
		{auty.AutonomyStatusProposalBoard, 20, 2},
		{auty.AutonomyStatusProposalBoard, 20, 5},
	}
	var testcase []statu
	testcase = append(testcase, testcase1...)
	testcase = append(testcase, testcase2...)
	cur := &auty.AutonomyProposalBoard{
		PropBoard:  &auty.ProposalBoard{},
		CurRule:    &auty.RuleConfig{},
		VoteResult: &auty.VoteResult{},
		Status:     auty.AutonomyStatusProposalBoard,
		Address:    "11111111111111",
		Height:     1,
		Index:      2,
	}

	//将数据保存下去
	var kvs []*types.KeyValue
	table := NewBoardTable(kvdb)
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
	req := &auty.ReqQueryProposalBoard{
		Status:    auty.AutonomyStatusProposalBoard,
		Count:     10,
		Direction: 0,
		Index:     -1,
	}
	rsp, err := au.listProposalBoard(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalBoard).PropBoards), len(testcase2))
	k := 2
	for _, tcase := range testcase2 {
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[k].Height, tcase.height)
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[k].Index, int32(tcase.index))
		k--
	}

	// 正向查找
	req = &auty.ReqQueryProposalBoard{
		Status:    auty.AutonomyStatusProposalBoard,
		Count:     10,
		Direction: 1,
		Index:     -1,
	}
	rsp, err = au.listProposalBoard(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalBoard).PropBoards), len(testcase2))
	for i, tcase := range testcase2 {
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[i].Height, tcase.height)
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[i].Index, int32(tcase.index))
	}

	// 翻页查找
	req = &auty.ReqQueryProposalBoard{
		Status:    auty.AutonomyStatusProposalBoard,
		Count:     1,
		Direction: 0,
		Index:     -1,
	}
	rsp, err = au.listProposalBoard(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalBoard).PropBoards), 1)
	height := rsp.(*auty.ReplyQueryProposalBoard).PropBoards[0].Height
	index := rsp.(*auty.ReplyQueryProposalBoard).PropBoards[0].Index
	assert.Equal(t, height, testcase2[2].height)
	assert.Equal(t, index, int32(testcase2[2].index))
	//
	req = &auty.ReqQueryProposalBoard{
		Status:    auty.AutonomyStatusProposalBoard,
		Count:     10,
		Direction: 0,
		Height:    height,
		Index:     index,
	}
	rsp, err = au.listProposalBoard(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalBoard).PropBoards), 2)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[0].Height, testcase2[1].height)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[0].Index, int32(testcase2[1].index))
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[1].Height, testcase2[0].height)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[1].Index, int32(testcase2[0].index))
}

func TestGetActiveBoard(t *testing.T) {
	au := newTestAutonomy()
	_, storedb, _ := util.CreateTestDB()
	au.SetStateDB(storedb)
	storedb.Set(activeBoardID(), types.Encode(&auty.ActiveBoard{Boards: []string{"111"}}))
	rsp, err := au.getActiveBoard()
	assert.NoError(t, err)
	assert.NotNil(t, rsp)
	assert.Equal(t, len(rsp.(*auty.ActiveBoard).Boards), 1)
}

func checkExecLocalBoard(t *testing.T, kvdb db.KVDB, cur *auty.AutonomyProposalBoard) {
	table := NewBoardTable(kvdb)
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

	prop, ok := rows[0].Data.(*auty.AutonomyProposalBoard)
	assert.Equal(t, true, ok)
	assert.Equal(t, prop.Status, cur.Status)
	assert.Equal(t, prop.Address, cur.Address)
	assert.Equal(t, prop.Height, cur.Height)
	assert.Equal(t, prop.Index, cur.Index)

}

func saveKvs(sdb db.DB, kvs []*types.KeyValue) {
	for _, kv := range kvs {
		if kv.Value == nil {
			sdb.Delete(kv.Key)
		} else {
			sdb.Set(kv.Key, kv.Value)
		}
	}
}
