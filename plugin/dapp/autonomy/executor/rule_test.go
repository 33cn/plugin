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

func TestExecLocalRule(t *testing.T) {
	testexecLocalRule(t, false)
	testexecLocalRule(t, true)
}

func testexecLocalRule(t *testing.T, auto bool) {
	_, sdb, kvdb := util.CreateTestDB()
	au := &Autonomy{}
	au.SetLocalDB(kvdb)
	//TyLogPropRule
	cur := &auty.AutonomyProposalRule{
		PropRule:   &auty.ProposalRule{},
		CurRule:    &auty.RuleConfig{},
		VoteResult: &auty.VoteResult{},
		Status:     auty.AutonomyStatusProposalRule,
		Address:    "11111111111111",
		Height:     1,
		Index:      2,
	}
	receiptRule := &auty.ReceiptProposalRule{
		Prev:    nil,
		Current: cur,
	}
	receipt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogPropRule, Log: types.Encode(receiptRule)},
		},
	}
	var set *types.LocalDBSet
	var err error

	if !auto {
		set, err = au.execLocalRule(receipt)
		assert.NoError(t, err)
		assert.NotNil(t, set)
	} else {
		tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
		assert.NoError(t, err)
		set, err = au.execAutoLocalRule(tx, receipt)
		assert.NoError(t, err)
		assert.NotNil(t, set)
	}

	//save to database
	saveKvs(sdb, set.KV)

	// check
	checkExecLocalRule(t, kvdb, cur)

	// TyLogRvkPropRule
	pre1 := copyAutonomyProposalRule(cur)
	cur.Status = auty.AutonomyStatusRvkPropRule
	receiptRule1 := &auty.ReceiptProposalRule{
		Prev:    pre1,
		Current: cur,
	}
	if !auto {
		set, err = au.execLocalRule(&types.ReceiptData{
			Logs: []*types.ReceiptLog{
				{Ty: auty.TyLogRvkPropRule, Log: types.Encode(receiptRule1)},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	} else {
		tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
		assert.NoError(t, err)
		set, err = au.execAutoLocalRule(tx,
			&types.ReceiptData{
				Logs: []*types.ReceiptLog{
					{Ty: auty.TyLogRvkPropRule, Log: types.Encode(receiptRule1)},
				},
			})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	}

	//save to database
	saveKvs(sdb, set.KV)

	// check
	checkExecLocalRule(t, kvdb, cur)

	// TyLogVotePropRule
	cur.Status = auty.AutonomyStatusProposalRule
	pre2 := copyAutonomyProposalRule(cur)
	cur.Status = auty.AutonomyStatusVotePropRule
	cur.Address = "2222222222222"
	receiptRule2 := &auty.ReceiptProposalRule{
		Prev:    pre2,
		Current: cur,
	}
	if !auto {
		set, err = au.execLocalRule(&types.ReceiptData{
			Logs: []*types.ReceiptLog{
				{Ty: auty.TyLogVotePropRule, Log: types.Encode(receiptRule2)},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	} else {
		tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
		assert.NoError(t, err)
		set, err = au.execAutoLocalRule(tx,
			&types.ReceiptData{
				Logs: []*types.ReceiptLog{
					{Ty: auty.TyLogVotePropRule, Log: types.Encode(receiptRule2)},
				},
			})
		assert.NoError(t, err)
		assert.NotNil(t, set)
	}

	//save to database
	saveKvs(sdb, set.KV)
	// check
	checkExecLocalRule(t, kvdb, cur)
}

func TestExecDelLocalRule(t *testing.T) {
	testexecDelLocalRule(t)
}

func testexecDelLocalRule(t *testing.T) {
	_, sdb, kvdb := util.CreateTestDB()
	au := &Autonomy{}
	au.SetLocalDB(kvdb)
	//TyLogPropRule
	cur := &auty.AutonomyProposalRule{
		PropRule:   &auty.ProposalRule{},
		CurRule:    &auty.RuleConfig{},
		VoteResult: &auty.VoteResult{},
		Status:     auty.AutonomyStatusProposalRule,
		Address:    "11111111111111",
		Height:     1,
		Index:      2,
	}
	receiptRule := &auty.ReceiptProposalRule{
		Prev:    nil,
		Current: cur,
	}
	receipt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogPropRule, Log: types.Encode(receiptRule)},
		},
	}

	// 先执行local然后进行删除

	tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
	assert.NoError(t, err)
	set, err := au.execAutoLocalRule(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	set, err = au.execAutoDelLocal(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	// check
	table := NewRuleTable(au.GetLocalDB())
	query := table.GetQuery(kvdb)
	_, err = query.ListIndex("primary", nil, nil, 10, 0)
	assert.Equal(t, err, types.ErrNotFound)
	_, err = query.ListIndex("addr", nil, nil, 10, 0)
	assert.Equal(t, err, types.ErrNotFound)
	_, err = query.ListIndex("status", nil, nil, 10, 0)
	assert.Equal(t, err, types.ErrNotFound)
	_, err = query.ListIndex("addr_status", nil, nil, 10, 0)
	assert.Equal(t, err, types.ErrNotFound)

	// TyLogVotePropRule
	pre1 := copyAutonomyProposalRule(cur)
	cur.Status = auty.AutonomyStatusVotePropRule
	receiptRule2 := &auty.ReceiptProposalRule{
		Prev:    pre1,
		Current: cur,
	}

	recpt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogVotePropRule, Log: types.Encode(receiptRule2)},
		}}
	// 先执行local然后进行删除
	// 自动回退测试时候，需要先设置一个前置状态
	tx, err = types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
	assert.NoError(t, err)
	set, err = au.execAutoLocalRule(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	// 正常测试退回
	tx, err = types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
	assert.NoError(t, err)
	set, err = au.execAutoLocalRule(tx, recpt)

	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)
	// check
	checkExecLocalRule(t, kvdb, cur)

	set, err = au.execAutoDelLocal(tx, recpt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	// check
	checkExecLocalRule(t, kvdb, pre1)
}

func checkExecLocalRule(t *testing.T, kvdb db.KVDB, cur *auty.AutonomyProposalRule) {
	table := NewRuleTable(kvdb)
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

	prop, ok := rows[0].Data.(*auty.AutonomyProposalRule)
	assert.Equal(t, true, ok)
	assert.Equal(t, prop.Status, cur.Status)
	assert.Equal(t, prop.Address, cur.Address)
	assert.Equal(t, prop.Height, cur.Height)
	assert.Equal(t, prop.Index, cur.Index)

}

func TestGetProposalRule(t *testing.T) {
	au := newTestAutonomy()
	_, storedb, _ := util.CreateTestDB()
	au.SetStateDB(storedb)
	tx := "1111111111111111111"
	storedb.Set(propRuleID(tx), types.Encode(&auty.AutonomyProposalRule{}))
	rsp, err := au.getProposalRule(&types.ReqString{Data: tx})
	assert.NoError(t, err)
	assert.NotNil(t, rsp)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalRule).PropRules), 1)
}

func TestListProposalRule(t *testing.T) {
	au := newTestAutonomy()
	_, sdb, kvdb := util.CreateTestDB()
	au.SetLocalDB(kvdb)

	type statu struct {
		status int32
		height int64
		index  int64
	}

	testcase1 := []statu{
		{auty.AutonomyStatusRvkPropRule, 10, 2},
		{auty.AutonomyStatusVotePropRule, 15, 1},
		{auty.AutonomyStatusTmintPropRule, 20, 1},
	}
	testcase2 := []statu{
		{auty.AutonomyStatusProposalRule, 10, 1},
		{auty.AutonomyStatusProposalRule, 20, 2},
		{auty.AutonomyStatusProposalRule, 20, 5},
	}
	var testcase []statu
	testcase = append(testcase, testcase1...)
	testcase = append(testcase, testcase2...)
	cur := &auty.AutonomyProposalRule{
		PropRule:   &auty.ProposalRule{},
		CurRule:    &auty.RuleConfig{},
		VoteResult: &auty.VoteResult{},
		Status:     auty.AutonomyStatusProposalRule,
		Address:    "11111111111111",
		Height:     1,
		Index:      2,
	}

	//将数据保存下去
	var kvs []*types.KeyValue
	table := NewRuleTable(kvdb)
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
	req := &auty.ReqQueryProposalRule{
		Status:    auty.AutonomyStatusProposalRule,
		Count:     10,
		Direction: 0,
		Index:     -1,
	}
	rsp, err := au.listProposalRule(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalRule).PropRules), len(testcase2))
	k := 2
	for _, tcase := range testcase2 {
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalRule).PropRules[k].Height, tcase.height)
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalRule).PropRules[k].Index, int32(tcase.index))
		k--
	}

	// 正向查找
	req = &auty.ReqQueryProposalRule{
		Status:    auty.AutonomyStatusProposalRule,
		Count:     10,
		Direction: 1,
		Index:     -1,
	}
	rsp, err = au.listProposalRule(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalRule).PropRules), len(testcase2))
	for i, tcase := range testcase2 {
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalRule).PropRules[i].Height, tcase.height)
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalRule).PropRules[i].Index, int32(tcase.index))
	}

	// 翻页查找
	req = &auty.ReqQueryProposalRule{
		Status:    auty.AutonomyStatusProposalRule,
		Count:     1,
		Direction: 0,
		Index:     -1,
	}
	rsp, err = au.listProposalRule(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalRule).PropRules), 1)
	height := rsp.(*auty.ReplyQueryProposalRule).PropRules[0].Height
	index := rsp.(*auty.ReplyQueryProposalRule).PropRules[0].Index
	assert.Equal(t, height, testcase2[2].height)
	assert.Equal(t, index, int32(testcase2[2].index))
	//
	req = &auty.ReqQueryProposalRule{
		Status:    auty.AutonomyStatusProposalRule,
		Count:     10,
		Direction: 0,
		Height:    height,
		Index:     index,
	}
	rsp, err = au.listProposalRule(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalRule).PropRules), 2)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalRule).PropRules[0].Height, testcase2[1].height)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalRule).PropRules[0].Index, int32(testcase2[1].index))
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalRule).PropRules[1].Height, testcase2[0].height)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalRule).PropRules[1].Index, int32(testcase2[0].index))
}

func TestExecLocalCommentProp(t *testing.T) {
	testexecLocalCommentProp(t, false)
	testexecLocalCommentProp(t, true)
}

func testexecLocalCommentProp(t *testing.T, auto bool) {
	_, _, kvdb := util.CreateTestDB()
	au := newTestAutonomy()
	au.SetLocalDB(kvdb)
	propID := "11111111111111"
	Repcmt := "2222222222"
	comment := "3333333333"
	receiptCmt := &auty.ReceiptProposalComment{
		Cmt: &auty.Comment{
			ProposalID: propID,
			RepHash:    Repcmt,
			Comment:    comment,
		},
		Height: 11,
		Index:  1,
	}
	receipt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogCommentProp, Log: types.Encode(receiptCmt)},
		},
	}
	var set *types.LocalDBSet
	var err error
	if !auto {
		set, err = au.execLocalCommentProp(receipt)
		assert.NoError(t, err)
		assert.NotNil(t, set)
	} else {
		tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
		assert.NoError(t, err)
		set, err = au.execAutoLocalCommentProp(tx, receipt)
		assert.NoError(t, err)
		assert.NotNil(t, set)
	}

	assert.Equal(t, set.KV[0].Key, calcCommentHeight(propID,
		dapp.HeightIndexStr(receiptCmt.Height, int64(receiptCmt.Index))))
	assert.NotNil(t, set.KV[0].Value)
}

func TestExecDelLocalCommentProp(t *testing.T) {
	testexecDelLocalCommentProp(t)
}

func testexecDelLocalCommentProp(t *testing.T) {
	_, sdb, kvdb := util.CreateTestDB()
	au := newTestAutonomy()
	au.SetLocalDB(kvdb)
	propID := "11111111111111"
	Repcmt := "2222222222"
	comment := "3333333333"
	receiptCmt := &auty.ReceiptProposalComment{
		Cmt: &auty.Comment{
			ProposalID: propID,
			RepHash:    Repcmt,
			Comment:    comment,
		},
		Height: 11,
		Index:  1,
	}
	receipt := &types.ReceiptData{
		Logs: []*types.ReceiptLog{
			{Ty: auty.TyLogCommentProp, Log: types.Encode(receiptCmt)},
		},
	}
	var set *types.LocalDBSet
	// 先执行local然后进行删除

	tx, err := types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), nil)
	assert.NoError(t, err)

	set, err = au.execAutoLocalCommentProp(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	saveKvs(sdb, set.KV)

	set, err = au.execAutoDelLocal(tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, set)

	// check
	assert.Equal(t, set.KV[0].Key, calcCommentHeight(propID,
		dapp.HeightIndexStr(receiptCmt.Height, int64(receiptCmt.Index))))
	assert.Nil(t, set.KV[0].Value)
}

func TestListProposalComment(t *testing.T) {
	au := newTestAutonomy()
	_, _, kvdb := util.CreateTestDB()
	au.SetLocalDB(kvdb)

	type statu struct {
		propID string
		height int64
		index  int64
	}

	propID := "3333333333"
	propID1 := "2222222"
	propID2 := "111111111111"

	testcase1 := []statu{
		{propID, 10, 2},
		{propID1, 15, 1},
		{propID, 20, 1},
	}
	testcase2 := []statu{
		{propID2, 10, 1},
		{propID2, 20, 2},
		{propID2, 20, 5},
	}
	var testcase []statu
	testcase = append(testcase, testcase1...)
	testcase = append(testcase, testcase2...)
	cur := &auty.RelationCmt{
		RepHash: "aaaaaa",
		Comment: "bbbbbbbbbb",
	}
	for _, tcase := range testcase {
		key := calcCommentHeight(tcase.propID,
			dapp.HeightIndexStr(tcase.height, tcase.index))
		cur.Height = tcase.height
		cur.Index = int32(tcase.index)
		value := types.Encode(cur)
		kvdb.Set(key, value)
	}

	// 反向查找
	req := &auty.ReqQueryProposalComment{
		ProposalID: propID2,
		Count:      10,
		Direction:  0,
		Index:      -1,
	}
	rsp, err := au.listProposalComment(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalComment).RltCmt), len(testcase2))
	k := 2
	for _, tcase := range testcase2 {
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalComment).RltCmt[k].Height, tcase.height)
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalComment).RltCmt[k].Index, int32(tcase.index))
		k--
	}

	// 正向查找
	req = &auty.ReqQueryProposalComment{
		ProposalID: propID2,
		Count:      10,
		Direction:  1,
		Index:      -1,
	}
	rsp, err = au.listProposalComment(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalComment).RltCmt), len(testcase2))
	for i, tcase := range testcase2 {
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalComment).RltCmt[i].Height, tcase.height)
		assert.Equal(t, rsp.(*auty.ReplyQueryProposalComment).RltCmt[i].Index, int32(tcase.index))
	}

	// 翻页查找
	req = &auty.ReqQueryProposalComment{
		ProposalID: propID2,
		Count:      1,
		Direction:  0,
		Index:      -1,
	}
	rsp, err = au.listProposalComment(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalComment).RltCmt), 1)
	height := rsp.(*auty.ReplyQueryProposalComment).RltCmt[0].Height
	index := rsp.(*auty.ReplyQueryProposalComment).RltCmt[0].Index
	assert.Equal(t, height, testcase2[2].height)
	assert.Equal(t, index, int32(testcase2[2].index))
	//
	req = &auty.ReqQueryProposalComment{
		ProposalID: propID2,
		Count:      10,
		Direction:  0,
		Height:     height,
		Index:      index,
	}
	rsp, err = au.listProposalComment(req)
	assert.NoError(t, err)
	assert.Equal(t, len(rsp.(*auty.ReplyQueryProposalComment).RltCmt), 2)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalComment).RltCmt[0].Height, testcase2[1].height)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalComment).RltCmt[0].Index, int32(testcase2[1].index))
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalComment).RltCmt[1].Height, testcase2[0].height)
	assert.Equal(t, rsp.(*auty.ReplyQueryProposalComment).RltCmt[1].Index, int32(testcase2[0].index))
}
