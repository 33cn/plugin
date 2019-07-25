// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/stretchr/testify/require"
)

func TestExecLocalBoard(t *testing.T) {
	au := &Autonomy{}
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
	set, err := au.execLocalBoard(receipt)
	require.NoError(t, err)
	require.NotNil(t, set)
	require.Equal(t, set.KV[0].Key, calcBoardKey4StatusHeight(cur.Status,
		dapp.HeightIndexStr(cur.Height, int64(cur.Index))))

	// TyLogRvkPropBoard
	pre1 := copyAutonomyProposalBoard(cur)
	cur.Status = auty.AutonomyStatusRvkPropBoard
	cur.Height = 2
	cur.Index = 3
	receiptBoard1 := &auty.ReceiptProposalBoard{
		Prev:    pre1,
		Current: cur,
	}
	set, err = au.execLocalBoard(&types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogRvkPropBoard, Log: types.Encode(receiptBoard1)},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, set)
	require.Equal(t, set.KV[0].Key, calcBoardKey4StatusHeight(pre1.Status,
		dapp.HeightIndexStr(pre1.Height, int64(pre1.Index))))
	require.Equal(t, set.KV[0].Value, []byte(nil))
	require.Equal(t, set.KV[1].Key, calcBoardKey4StatusHeight(cur.Status,
		dapp.HeightIndexStr(cur.Height, int64(cur.Index))))

	// TyLogVotePropBoard
	cur.Status = auty.AutonomyStatusProposalBoard
	cur.Height = 1
	cur.Index = 2
	pre2 := copyAutonomyProposalBoard(cur)
	cur.Status = auty.AutonomyStatusVotePropBoard
	cur.Height = 2
	cur.Index = 3
	receiptBoard2 := &auty.ReceiptProposalBoard{
		Prev:    pre2,
		Current: cur,
	}
	set, err = au.execLocalBoard(&types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogVotePropBoard, Log: types.Encode(receiptBoard2)},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, set)
	require.Equal(t, set.KV[0].Key, calcBoardKey4StatusHeight(pre2.Status,
		dapp.HeightIndexStr(pre1.Height, int64(pre2.Index))))
	require.Equal(t, set.KV[0].Value, []byte(nil))
	require.Equal(t, set.KV[1].Key, calcBoardKey4StatusHeight(cur.Status,
		dapp.HeightIndexStr(cur.Height, int64(cur.Index))))
}

func TestExecDelLocalBoard(t *testing.T) {
	au := &Autonomy{}
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
	set, err := au.execDelLocalBoard(receipt)
	require.NoError(t, err)
	require.NotNil(t, set)
	require.Equal(t, set.KV[0].Key, calcBoardKey4StatusHeight(cur.Status,
		dapp.HeightIndexStr(cur.Height, int64(cur.Index))))
	require.Equal(t, set.KV[0].Value, []byte(nil))

	// TyLogVotePropBoard
	pre1 := copyAutonomyProposalBoard(cur)
	cur.Status = auty.AutonomyStatusVotePropBoard
	cur.Height = 2
	cur.Index = 3
	receiptBoard2 := &auty.ReceiptProposalBoard{
		Prev:    pre1,
		Current: cur,
	}
	set, err = au.execDelLocalBoard(&types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogVotePropBoard, Log: types.Encode(receiptBoard2)},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, set)
	require.Equal(t, set.KV[0].Key, calcBoardKey4StatusHeight(cur.Status,
		dapp.HeightIndexStr(cur.Height, int64(cur.Index))))
	require.Equal(t, set.KV[0].Value, []byte(nil))
	require.Equal(t, set.KV[1].Key, calcBoardKey4StatusHeight(pre1.Status,
		dapp.HeightIndexStr(pre1.Height, int64(pre1.Index))))
	require.NotNil(t, set.KV[1].Value)
}

func TestGetProposalBoard(t *testing.T) {
	au := &Autonomy{
		dapp.DriverBase{},
	}
	_, storedb, _ := util.CreateTestDB()
	au.SetStateDB(storedb)
	tx := "1111111111111111111"
	storedb.Set(propBoardID(tx), types.Encode(&auty.AutonomyProposalBoard{}))
	rsp, err := au.getProposalBoard(&types.ReqString{Data: tx})
	require.NoError(t, err)
	require.NotNil(t, rsp)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalBoard).PropBoards), 1)
}

func TestListProposalBoard(t *testing.T) {
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
	for _, tcase := range testcase {
		key := calcBoardKey4StatusHeight(tcase.status,
			dapp.HeightIndexStr(tcase.height, int64(tcase.index)))
		cur.Status = tcase.status
		cur.Height = tcase.height
		cur.Index = int32(tcase.index)
		value := types.Encode(cur)
		kvdb.Set(key, value)
	}

	// 反向查找
	req := &auty.ReqQueryProposalBoard{
		Status:    auty.AutonomyStatusProposalBoard,
		Count:     10,
		Direction: 0,
		Index:     -1,
	}
	rsp, err := au.listProposalBoard(req)
	require.NoError(t, err)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalBoard).PropBoards), len(testcase2))
	k := 2
	for _, tcase := range testcase2 {
		require.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[k].Height, tcase.height)
		require.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[k].Index, int32(tcase.index))
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
	require.NoError(t, err)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalBoard).PropBoards), len(testcase2))
	for i, tcase := range testcase2 {
		require.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[i].Height, tcase.height)
		require.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[i].Index, int32(tcase.index))
	}

	// 翻页查找
	req = &auty.ReqQueryProposalBoard{
		Status:    auty.AutonomyStatusProposalBoard,
		Count:     1,
		Direction: 0,
		Index:     -1,
	}
	rsp, err = au.listProposalBoard(req)
	require.NoError(t, err)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalBoard).PropBoards), 1)
	height := rsp.(*auty.ReplyQueryProposalBoard).PropBoards[0].Height
	index := rsp.(*auty.ReplyQueryProposalBoard).PropBoards[0].Index
	require.Equal(t, height, testcase2[2].height)
	require.Equal(t, index, int32(testcase2[2].index))
	//
	Index := height*types.MaxTxsPerBlock + int64(index)
	req = &auty.ReqQueryProposalBoard{
		Status:    auty.AutonomyStatusProposalBoard,
		Count:     10,
		Direction: 0,
		Index:     Index,
	}
	rsp, err = au.listProposalBoard(req)
	require.NoError(t, err)
	require.Equal(t, len(rsp.(*auty.ReplyQueryProposalBoard).PropBoards), 2)
	require.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[0].Height, testcase2[1].height)
	require.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[0].Index, int32(testcase2[1].index))
	require.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[1].Height, testcase2[0].height)
	require.Equal(t, rsp.(*auty.ReplyQueryProposalBoard).PropBoards[1].Index, int32(testcase2[0].index))
}
