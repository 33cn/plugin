// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/stretchr/testify/require"
	"github.com/33cn/chain33/types"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/account"
	_ "github.com/33cn/chain33/system"
	"github.com/stretchr/testify/mock"
	"github.com/33cn/chain33/common/address"
	drivers "github.com/33cn/chain33/system/dapp"
)


func TestRevokeProposalRule(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	// PropRule
	testPropRule(t, env, exec, stateDB, kvdb, true)
	//RevokeProposalRule
	revokeProposalRule(t, env, exec, stateDB, kvdb, false)
}

func TestVoteProposalRule(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	// PropRule
	testPropRule(t, env, exec, stateDB, kvdb, true)
	//voteProposalRule
	voteProposalRule(t, env, exec, stateDB, kvdb, true)
}

func TestTerminateProposalRule(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	// PropRule
	testPropRule(t, env, exec, stateDB, kvdb, true)
	//terminateProposalRule
	terminateProposalRule(t, env, exec, stateDB, kvdb, true)
}

func testPropRule(t *testing.T, env *execEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	opt1 :=  &auty.ProposalRule{
		Year: 2019,
		Month: 7,
		Day:     10,
		RuleCfg:    &auty.RuleConfig{BoardApproveRatio:60},
		StartBlockHeight:  env.blockHeight + 5,
		EndBlockHeight: env.blockHeight + 10,
	}
	pbtx, err := propRuleTx(opt1)
	require.NoError(t, err)
	pbtx, err = signTx(pbtx, PrivKeyA)
	require.NoError(t, err)

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(pbtx, int(1))
	require.NoError(t, err)
	require.NotNil(t, receipt)

	if save {
		for _, kv := range receipt.KV {
			stateDB.Set(kv.Key, kv.Value)
		}
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(pbtx, receiptData, int(1))
	require.NoError(t, err)
	require.NotNil(t, set)
	if save {
		for _, kv := range set.KV {
			kvdb.Set(kv.Key, kv.Value)
		}
	}
	// 更新tahash
	env.txHash = common.ToHex(pbtx.Hash())
	env.startHeight = opt1.StartBlockHeight
	env.endHeight = opt1.EndBlockHeight

	// check
	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, proposalAmount, account.Frozen)
}

func propRuleTx(parm *auty.ProposalRule) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPropRule,
		Value: &auty.AutonomyAction_PropRule{PropRule: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func revokeProposalRule(t *testing.T, env *execEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	proposalID := env.txHash
	opt2 :=  &auty.RevokeProposalRule{
		ProposalID:proposalID,
	}
	rtx, err := revokeProposalRuleTx(opt2)
	require.NoError(t, err)
	rtx, err = signTx(rtx, PrivKeyA)
	require.NoError(t, err)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(rtx, int(1))
	require.NoError(t, err)
	require.NotNil(t, receipt)
	if save {
		for _, kv := range receipt.KV {
			stateDB.Set(kv.Key, kv.Value)
		}
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(rtx, receiptData, int(1))
	require.NoError(t, err)
	require.NotNil(t, set)
	if save {
		for _, kv := range set.KV {
			kvdb.Set(kv.Key, kv.Value)
		}
	}
	// check
	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(0), account.Frozen)
}

func revokeProposalRuleTx(parm *auty.RevokeProposalRule) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionRvkPropRule,
		Value: &auty.AutonomyAction_RvkPropRule{RvkPropRule: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func voteProposalRule(t *testing.T, env *execEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	api := new(apimock.QueueProtocolAPI)
	api.On("StoreList", mock.Anything).Return(&types.StoreListReply{}, nil)
	api.On("GetLastHeader", mock.Anything).Return(&types.Header{StateHash: []byte("")}, nil)
	hear := &types.Header{StateHash: []byte("")}
	api.On("GetHeaders", mock.Anything).
		Return(&types.Headers{
		Items:[]*types.Header{hear}}, nil)
	acc := &types.Account{
		Currency: 0,
		Balance: total*4,
	}
	val := types.Encode(acc)
	values := [][]byte{val}
	api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values:values}, nil).Once()

	acc = &types.Account{
		Currency: 0,
		Balance: total,
	}
	val1 := types.Encode(acc)
	values1 := [][]byte{val1}
	api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values:values1}, nil).Once()
	exec.SetAPI(api)

	proposalID := env.txHash
	// 4人参与投票，3人赞成票，1人反对票
	type record struct {
		priv string
		appr bool
	}
	records := []record{
		{PrivKeyA, false},
		{PrivKeyB, false},
		{PrivKeyC, true},
		{PrivKeyD, true},
	}

	for _, record := range records {
		opt :=  &auty.VoteProposalRule{
			ProposalID:proposalID,
			Approve: record.appr,
		}
		tx, err := voteProposalRuleTx(opt)
		require.NoError(t, err)
		tx, err = signTx(tx, record.priv)
		require.NoError(t, err)
		// 设定当前高度为投票高度
		exec.SetEnv(env.startHeight, env.blockTime, env.difficulty)

		receipt, err := exec.Exec(tx, int(1))
		require.NoError(t, err)
		require.NotNil(t, receipt)
		if save {
			for _, kv := range receipt.KV {
				stateDB.Set(kv.Key, kv.Value)
			}
		}
		receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
		set, err := exec.ExecLocal(tx, receiptData, int(1))
		require.NoError(t, err)
		require.NotNil(t, set)
		if save {
			for _, kv := range set.KV {
				kvdb.Set(kv.Key, kv.Value)
			}
		}

		// 每次需要重新设置
		acc := &types.Account{
			Currency: 0,
			Balance: total,
		}
		val := types.Encode(acc)
		values := [][]byte{val}
		api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values:values}, nil).Once()
		exec.SetAPI(api)
	}
	// check
	// balance
	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(0), account.Frozen)
	account = accCoin.LoadExecAccount(autonomyAddr, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(proposalAmount), account.Balance)
	// status
	value, err := stateDB.Get(propRuleID(proposalID))
	require.NoError(t, err)
	cur := &auty.AutonomyProposalRule{}
	err = types.Decode(value, cur)
	require.NoError(t, err)
	require.Equal(t, int32(auty.AutonomyStatusTmintPropRule), cur.Status)
	require.Equal(t, AddrA, cur.Address)
	require.Equal(t, true, cur.VoteResult.Pass)
}

func voteProposalRuleTx(parm *auty.VoteProposalRule) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionVotePropRule,
		Value: &auty.AutonomyAction_VotePropRule{VotePropRule: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func terminateProposalRule(t *testing.T, env *execEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	api := new(apimock.QueueProtocolAPI)
	api.On("StoreList", mock.Anything).Return(&types.StoreListReply{}, nil)
	api.On("GetLastHeader", mock.Anything).Return(&types.Header{StateHash: []byte("")}, nil)
	hear := &types.Header{StateHash: []byte("")}
	api.On("GetHeaders", mock.Anything).
		Return(&types.Headers{
		Items:[]*types.Header{hear}}, nil)
	acc := &types.Account{
		Currency: 0,
		Balance: total*4,
	}
	val := types.Encode(acc)
	values := [][]byte{val}
	api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values:values}, nil).Once()
	exec.SetAPI(api)

	proposalID := env.txHash
	opt :=  &auty.TerminateProposalRule{
		ProposalID:proposalID,
	}
	tx, err := terminateProposalRuleTx(opt)
	require.NoError(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.NoError(t, err)
	exec.SetEnv(env.endHeight+1, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(tx, int(1))
	require.NoError(t, err)
	require.NotNil(t, receipt)
	if save {
		for _, kv := range receipt.KV {
			stateDB.Set(kv.Key, kv.Value)
		}
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(tx, receiptData, int(1))
	require.NoError(t, err)
	require.NotNil(t, set)
	if save {
		for _, kv := range set.KV {
			kvdb.Set(kv.Key, kv.Value)
		}
	}
	// check
	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(0), account.Frozen)
}

func terminateProposalRuleTx(parm *auty.TerminateProposalRule) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTmintPropRule,
		Value: &auty.AutonomyAction_TmintPropRule{TmintPropRule: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func TestGetRuleReceiptLog(t *testing.T) {
	pre := &auty.AutonomyProposalRule{
		PropRule: &auty.ProposalRule{Year: 1800, Month: 1},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status: 1,
		Address:"121",
	}
	cur := &auty.AutonomyProposalRule{
		PropRule: &auty.ProposalRule{Year: 1900, Month: 1},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status: 2,
		Address:"123",
	}
	log := getRuleReceiptLog(pre, cur, 2)
	require.Equal(t, int32(2), log.Ty)
	recpt := &auty.ReceiptProposalRule{}
	err := types.Decode(log.Log, recpt)
	require.NoError(t, err)
	require.Equal(t, int32(1800), recpt.Prev.PropRule.Year)
	require.Equal(t, int32(1900), recpt.Current.PropRule.Year)
}

func TestCopyAutonomyProposalRule(t *testing.T) {
	require.Nil(t, copyAutonomyProposalRule(nil))
	cur := &auty.AutonomyProposalRule{
		PropRule: &auty.ProposalRule{Year: 1900, Month: 1, RuleCfg:&auty.RuleConfig{BoardApproveRatio:80}},
		Rule: &auty.RuleConfig{BoardApproveRatio:100},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status: 2,
		Address:"123",
	}
	new := copyAutonomyProposalRule(cur)
	cur.PropRule.Year = 1800
	cur.PropRule.Month = 2
	cur.PropRule.RuleCfg.BoardApproveRatio = 90
	cur.Rule.BoardApproveRatio = 90
	cur.VoteResult.TotalVotes = 50
	cur.Address = "234"
	cur.Status = 1

	require.Equal(t, 1900, int(new.PropRule.Year))
	require.Equal(t, 1, int(new.PropRule.Month))
	require.Equal(t, 100, int(new.VoteResult.TotalVotes))
	require.Equal(t, "123", new.Address)
	require.Equal(t, 2, int(new.Status))
	require.Equal(t, 80, int(new.PropRule.RuleCfg.BoardApproveRatio))
	require.Equal(t, 100, int(new.Rule.BoardApproveRatio))
}