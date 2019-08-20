// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	_ "github.com/33cn/chain33/system"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testBoardAttendRatio   int32 = 60
	testBoardApproveRatio  int32 = 60
	testPubOpposeRatio     int32 = 30
	testProposalAmount     int64 = 0
	testLargeProjectAmount int64 = 1
	testPublicPeriod       int32 = 100
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

func testPropRule(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	opt1 := &auty.ProposalRule{
		Year:  2019,
		Month: 7,
		Day:   10,
		RuleCfg: &auty.RuleConfig{
			BoardAttendRatio:   testBoardAttendRatio,
			BoardApproveRatio:  testBoardApproveRatio,
			PubOpposeRatio:     testPubOpposeRatio,
			ProposalAmount:     testProposalAmount,
			LargeProjectAmount: testLargeProjectAmount,
			PublicPeriod:       testPublicPeriod,
		},
		StartBlockHeight: env.blockHeight + 5,
		EndBlockHeight:   env.blockHeight + 10,
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

func revokeProposalRule(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	proposalID := env.txHash
	opt2 := &auty.RevokeProposalRule{
		ProposalID: proposalID,
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
	// del
	set, err = exec.ExecDelLocal(rtx, receiptData, int(1))
	require.NoError(t, err)
	require.NotNil(t, set)
	// check
	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(0), account.Frozen)
	// check rule
	au := &Autonomy{
		drivers.DriverBase{},
	}
	au.SetStateDB(stateDB)
	au.SetLocalDB(kvdb)
	action := newAction(au, &types.Transaction{}, 0)
	rule, err := action.getActiveRule()
	require.NoError(t, err)
	require.Equal(t, rule.BoardAttendRatio, boardAttendRatio)
	require.Equal(t, rule.BoardApproveRatio, boardApproveRatio)
	require.Equal(t, rule.PubOpposeRatio, pubOpposeRatio)
	require.Equal(t, rule.ProposalAmount, proposalAmount)
	require.Equal(t, rule.LargeProjectAmount, largeProjectAmount)
	require.Equal(t, rule.PublicPeriod, publicPeriod)
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

func voteProposalRule(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	api := new(apimock.QueueProtocolAPI)
	api.On("StoreList", mock.Anything).Return(&types.StoreListReply{}, nil)
	api.On("GetLastHeader", mock.Anything).Return(&types.Header{StateHash: []byte("")}, nil)
	hear := &types.Header{StateHash: []byte("")}
	api.On("GetHeaders", mock.Anything).
		Return(&types.Headers{
			Items: []*types.Header{hear}}, nil)
	acc := &types.Account{
		Currency: 0,
		Balance:  total * 4,
	}
	val := types.Encode(acc)
	values := [][]byte{val}
	api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values: values}, nil).Once()

	acc = &types.Account{
		Currency: 0,
		Frozen:   total,
	}
	val1 := types.Encode(acc)
	values1 := [][]byte{val1}
	api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values: values1}, nil).Once()
	exec.SetAPI(api)

	proposalID := env.txHash
	// 4人参与投票，3人赞成票，1人反对票
	type record struct {
		priv string
		appr bool
	}
	records := []record{
		{PrivKeyA, false},
		{PrivKeyB, true},
		{PrivKeyC, true},
		//{PrivKeyD, true},
	}

	for _, record := range records {
		opt := &auty.VoteProposalRule{
			ProposalID: proposalID,
			Approve:    record.appr,
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
		// del
		set, err = exec.ExecDelLocal(tx, receiptData, int(1))
		require.NoError(t, err)
		require.NotNil(t, set)

		// 每次需要重新设置
		acc := &types.Account{
			Currency: 0,
			Frozen:   total,
		}
		val := types.Encode(acc)
		values := [][]byte{val}
		api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values: values}, nil).Once()
		exec.SetAPI(api)
	}
	// check
	// balance
	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(0), account.Frozen)
	account = accCoin.LoadExecAccount(autonomyFundAddr, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, proposalAmount, account.Balance)
	// status
	value, err := stateDB.Get(propRuleID(proposalID))
	require.NoError(t, err)
	cur := &auty.AutonomyProposalRule{}
	err = types.Decode(value, cur)
	require.NoError(t, err)
	require.Equal(t, int32(auty.AutonomyStatusTmintPropRule), cur.Status)
	require.Equal(t, AddrA, cur.Address)
	require.Equal(t, true, cur.VoteResult.Pass)
	// check rule
	au := &Autonomy{
		drivers.DriverBase{},
	}
	au.SetStateDB(stateDB)
	au.SetLocalDB(kvdb)
	action := newAction(au, &types.Transaction{}, 0)
	rule, err := action.getActiveRule()
	require.NoError(t, err)
	require.Equal(t, rule.BoardAttendRatio, testBoardAttendRatio)
	require.Equal(t, rule.BoardApproveRatio, testBoardApproveRatio)
	require.Equal(t, rule.PubOpposeRatio, testPubOpposeRatio)
	require.Equal(t, rule.ProposalAmount, proposalAmount)
	require.Equal(t, rule.LargeProjectAmount, testLargeProjectAmount)
	require.Equal(t, rule.PublicPeriod, testPublicPeriod)
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

func terminateProposalRule(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	api := new(apimock.QueueProtocolAPI)
	api.On("StoreList", mock.Anything).Return(&types.StoreListReply{}, nil)
	api.On("GetLastHeader", mock.Anything).Return(&types.Header{StateHash: []byte("")}, nil)
	hear := &types.Header{StateHash: []byte("")}
	api.On("GetHeaders", mock.Anything).
		Return(&types.Headers{
			Items: []*types.Header{hear}}, nil)
	acc := &types.Account{
		Currency: 0,
		Balance:  total * 4,
	}
	val := types.Encode(acc)
	values := [][]byte{val}
	api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values: values}, nil).Once()
	exec.SetAPI(api)

	proposalID := env.txHash
	opt := &auty.TerminateProposalRule{
		ProposalID: proposalID,
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
	// del
	set, err = exec.ExecDelLocal(tx, receiptData, int(1))
	require.NoError(t, err)
	require.NotNil(t, set)
	// check
	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(0), account.Frozen)

	// check rule
	au := &Autonomy{
		drivers.DriverBase{},
	}
	au.SetStateDB(stateDB)
	au.SetLocalDB(kvdb)
	action := newAction(au, &types.Transaction{}, 0)
	rule, err := action.getActiveRule()
	require.NoError(t, err)
	require.Equal(t, rule.BoardAttendRatio, boardAttendRatio)
	require.Equal(t, rule.BoardApproveRatio, boardApproveRatio)
	require.Equal(t, rule.PubOpposeRatio, pubOpposeRatio)
	require.Equal(t, rule.ProposalAmount, proposalAmount)
	require.Equal(t, rule.LargeProjectAmount, largeProjectAmount)
	require.Equal(t, rule.PublicPeriod, publicPeriod)
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
		PropRule:   &auty.ProposalRule{Year: 1800, Month: 1},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status:     1,
		Address:    "121",
	}
	cur := &auty.AutonomyProposalRule{
		PropRule:   &auty.ProposalRule{Year: 1900, Month: 1},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status:     2,
		Address:    "123",
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
		PropRule:   &auty.ProposalRule{Year: 1900, Month: 1, RuleCfg: &auty.RuleConfig{BoardApproveRatio: 80}},
		CurRule:    &auty.RuleConfig{BoardApproveRatio: 100},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status:     2,
		Address:    "123",
	}
	pre := copyAutonomyProposalRule(cur)
	cur.PropRule.Year = 1800
	cur.PropRule.Month = 2
	cur.PropRule.RuleCfg.BoardApproveRatio = 90
	cur.CurRule.BoardApproveRatio = 90
	cur.VoteResult.TotalVotes = 50
	cur.Address = "234"
	cur.Status = 1

	require.Equal(t, 1900, int(pre.PropRule.Year))
	require.Equal(t, 1, int(pre.PropRule.Month))
	require.Equal(t, 100, int(pre.VoteResult.TotalVotes))
	require.Equal(t, "123", pre.Address)
	require.Equal(t, 2, int(pre.Status))
	require.Equal(t, 80, int(pre.PropRule.RuleCfg.BoardApproveRatio))
	require.Equal(t, 100, int(pre.CurRule.BoardApproveRatio))
}

func TestUpgradeRule(t *testing.T) {
	new := upgradeRule(nil, &auty.RuleConfig{})
	require.Nil(t, new)
	cur := &auty.RuleConfig{
		BoardAttendRatio:   1,
		BoardApproveRatio:  2,
		PubOpposeRatio:     3,
		ProposalAmount:     4,
		LargeProjectAmount: 5,
		PublicPeriod:       6,
	}
	modify := &auty.RuleConfig{
		BoardAttendRatio:   0,
		BoardApproveRatio:  -1,
		PubOpposeRatio:     0,
		ProposalAmount:     -1,
		LargeProjectAmount: 0,
		PublicPeriod:       0,
	}
	new = upgradeRule(cur, modify)
	require.NotNil(t, new)
	require.Equal(t, new.BoardAttendRatio, cur.BoardAttendRatio)
	require.Equal(t, new.BoardApproveRatio, cur.BoardApproveRatio)
	require.Equal(t, new.PubOpposeRatio, cur.PubOpposeRatio)
	require.Equal(t, new.ProposalAmount, cur.ProposalAmount)
	require.Equal(t, new.LargeProjectAmount, cur.LargeProjectAmount)
	require.Equal(t, new.PublicPeriod, cur.PublicPeriod)

	modify = &auty.RuleConfig{
		BoardAttendRatio:   10,
		BoardApproveRatio:  20,
		PubOpposeRatio:     30,
		ProposalAmount:     40,
		LargeProjectAmount: 50,
		PublicPeriod:       60,
	}
	new = upgradeRule(cur, modify)
	require.NotNil(t, new)
	require.Equal(t, new.BoardAttendRatio, modify.BoardAttendRatio)
	require.Equal(t, new.BoardApproveRatio, modify.BoardApproveRatio)
	require.Equal(t, new.PubOpposeRatio, modify.PubOpposeRatio)
	require.Equal(t, new.ProposalAmount, modify.ProposalAmount)
	require.Equal(t, new.LargeProjectAmount, modify.LargeProjectAmount)
	require.Equal(t, new.PublicPeriod, modify.PublicPeriod)
}

func TestTransfer(t *testing.T) {
	env, exec, stateDB, _ := InitEnv()

	opt1 := &auty.TransferFund{
		Amount: types.Coin * 190,
	}
	pbtx, err := transferFundTx(opt1)
	require.NoError(t, err)
	pbtx, err = signTx(pbtx, PrivKeyA)
	require.NoError(t, err)

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(pbtx, int(1))
	require.NoError(t, err)
	require.NotNil(t, receipt)

	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	// check
	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, total-types.Coin*190, account.Balance)
	account = accCoin.LoadExecAccount(autonomyFundAddr, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, types.Coin*190, account.Balance)
}

func transferFundTx(parm *auty.TransferFund) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTransfer,
		Value: &auty.AutonomyAction_Transfer{Transfer: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func TestComment(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()

	propID := "11111111111111"
	Repcmt := "2222222222"
	comment := "3333333333"
	opt1 := &auty.Comment{
		ProposalID: propID,
		RepHash:    Repcmt,
		Comment:    comment,
	}
	pbtx, err := commentPropTx(opt1)
	require.NoError(t, err)
	pbtx, err = signTx(pbtx, PrivKeyA)
	require.NoError(t, err)

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(pbtx, int(1))
	require.NoError(t, err)
	require.NotNil(t, receipt)

	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(pbtx, receiptData, int(1))
	require.NoError(t, err)
	require.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	// check
	value, err := kvdb.Get(calcCommentHeight(propID, drivers.HeightIndexStr(env.blockHeight, 1)))
	require.NoError(t, err)
	cmt := &auty.RelationCmt{}
	err = types.Decode(value, cmt)
	require.NoError(t, err)
	require.Equal(t, cmt.Comment, comment)
	require.Equal(t, cmt.RepHash, Repcmt)
}

func commentPropTx(parm *auty.Comment) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionCommentProp,
		Value: &auty.AutonomyAction_CommentProp{CommentProp: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}
