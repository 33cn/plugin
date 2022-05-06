// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	_ "github.com/33cn/chain33/system"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testBoardApproveRatio  int32 = 60
	testPubOpposeRatio     int32 = 35
	testProposalAmount           = 20 * types.DefaultCoinPrecision * 2
	testLargeProjectAmount       = 1000000 * 2 * types.DefaultCoinPrecision
	testPublicPeriod       int32 = 17280 * 7
	testPubAttendRatio     int32 = 80
	testPubApproveRatio    int32 = 50
)

func TestPropRule(t *testing.T) {
	env, exec, _, _ := InitEnv()
	opts := []*auty.ProposalRule{
		{ // 全0测试
			RuleCfg:          &auty.RuleConfig{},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // 边界测试
			RuleCfg: &auty.RuleConfig{
				BoardApproveRatio:  int32(autoCfg.MaxBoardApproveRatio),
				PubOpposeRatio:     int32(autoCfg.MaxPubOpposeRatio),
				ProposalAmount:     autoCfg.MaxProposalAmount * types.DefaultCoinPrecision,
				LargeProjectAmount: autoCfg.MaxLargeProjectAmount * types.DefaultCoinPrecision,
				PublicPeriod:       int32(autoCfg.MaxPublicPeriod),
				PubAttendRatio:     int32(autoCfg.MaxPubAttendRatio),
				PubApproveRatio:    int32(autoCfg.MaxPubApproveRatio),
			},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{
			RuleCfg: &auty.RuleConfig{
				BoardApproveRatio:  int32(autoCfg.MinBoardApproveRatio),
				PubOpposeRatio:     int32(autoCfg.MinPubOpposeRatio),
				ProposalAmount:     autoCfg.MinProposalAmount * types.DefaultCoinPrecision,
				LargeProjectAmount: autoCfg.MinLargeProjectAmount * types.DefaultCoinPrecision,
				PublicPeriod:       int32(autoCfg.MinPublicPeriod),
				PubAttendRatio:     int32(autoCfg.MinPubAttendRatio),
				PubApproveRatio:    int32(autoCfg.MinPubApproveRatio),
			},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{
			RuleCfg: &auty.RuleConfig{
				BoardApproveRatio:  int32(autoCfg.MinBoardApproveRatio) - 1,
				PubOpposeRatio:     int32(autoCfg.MinPubOpposeRatio) - 1,
				ProposalAmount:     autoCfg.MinProposalAmount*types.DefaultCoinPrecision - 1,
				LargeProjectAmount: autoCfg.MinLargeProjectAmount*types.DefaultCoinPrecision - 1,
				PublicPeriod:       int32(autoCfg.MinPublicPeriod) - 1,
				PubAttendRatio:     int32(autoCfg.MinPubAttendRatio) - 1,
				PubApproveRatio:    int32(autoCfg.MinPubApproveRatio) - 1,
			},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // 边界测试
			RuleCfg: &auty.RuleConfig{
				BoardApproveRatio:  int32(autoCfg.MaxBoardApproveRatio) + 1,
				PubOpposeRatio:     int32(autoCfg.MaxPubOpposeRatio) + 1,
				ProposalAmount:     autoCfg.MaxProposalAmount*types.DefaultCoinPrecision + 1,
				LargeProjectAmount: autoCfg.MaxLargeProjectAmount*types.DefaultCoinPrecision + 1,
				PublicPeriod:       int32(autoCfg.MaxPublicPeriod) + 1,
				PubAttendRatio:     int32(autoCfg.MaxPubAttendRatio) + 1,
				PubApproveRatio:    int32(autoCfg.MaxPubApproveRatio) + 1,
			},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // 配置参数其中之一不合法
			RuleCfg: &auty.RuleConfig{
				BoardApproveRatio:  1,
				PubOpposeRatio:     int32(autoCfg.MinPubOpposeRatio) + 1,
				ProposalAmount:     autoCfg.MinProposalAmount*types.DefaultCoinPrecision + 1,
				LargeProjectAmount: autoCfg.MinLargeProjectAmount*types.DefaultCoinPrecision + 1,
				PublicPeriod:       int32(autoCfg.MinPublicPeriod) + 1,
				PubAttendRatio:     int32(autoCfg.MinPubAttendRatio) + 1,
				PubApproveRatio:    int32(autoCfg.MinPubApproveRatio) + 1,
			},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
	}

	result := []error{
		types.ErrInvalidParam,
		nil,
		nil,
		types.ErrInvalidParam,
		types.ErrInvalidParam,
		types.ErrInvalidParam,
	}

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	for i, tcase := range opts {
		pbtx, err := propRuleTx(tcase)
		assert.NoError(t, err)
		pbtx, err = signTx(pbtx, PrivKeyA)
		assert.NoError(t, err)
		_, err = exec.Exec(pbtx, i)
		assert.Equal(t, err, result[i])
	}
}

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
			BoardApproveRatio:  testBoardApproveRatio,
			PubOpposeRatio:     testPubOpposeRatio,
			ProposalAmount:     testProposalAmount,
			LargeProjectAmount: testLargeProjectAmount,
			PublicPeriod:       testPublicPeriod,
			PubAttendRatio:     testPubAttendRatio,
			PubApproveRatio:    testPubApproveRatio,
		},
		StartBlockHeight: env.blockHeight + 5,
		EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
	}
	pbtx, err := propRuleTx(opt1)
	assert.NoError(t, err)
	pbtx, err = signTx(pbtx, PrivKeyA)
	assert.NoError(t, err)

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(pbtx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	if save {
		for _, kv := range receipt.KV {
			_ = stateDB.Set(kv.Key, kv.Value)
		}
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(pbtx, receiptData, 1)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	if save {
		for _, kv := range set.KV {
			_ = kvdb.Set(kv.Key, kv.Value)
		}
	}

	// 更新tahash
	env.txHash = common.ToHex(pbtx.Hash())
	env.startHeight = opt1.StartBlockHeight
	env.endHeight = opt1.EndBlockHeight

	// check
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	accountAddr := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, autoCfg.ProposalAmount*types.DefaultCoinPrecision, accountAddr.Frozen)
}

func propRuleTx(parm *auty.ProposalRule) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPropRule,
		Value: &auty.AutonomyAction_PropRule{PropRule: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func revokeProposalRule(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	proposalID := env.txHash
	opt2 := &auty.RevokeProposalRule{
		ProposalID: proposalID,
	}
	rtx, err := revokeProposalRuleTx(opt2)
	assert.NoError(t, err)
	rtx, err = signTx(rtx, PrivKeyA)
	assert.NoError(t, err)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(rtx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)
	if save {
		for _, kv := range receipt.KV {
			_ = stateDB.Set(kv.Key, kv.Value)
		}
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(rtx, receiptData, 1)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	if save {
		for _, kv := range set.KV {
			_ = kvdb.Set(kv.Key, kv.Value)
		}
	}
	// del
	set, err = exec.ExecDelLocal(rtx, receiptData, 1)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	// check
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	accountAddr := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, int64(0), accountAddr.Frozen)
	// check rule
	au := newTestAutonomy()
	au.SetStateDB(stateDB)
	au.SetLocalDB(kvdb)
	action := newAction(au, &types.Transaction{}, 0)
	rule, err := action.getActiveRule()
	assert.NoError(t, err)
	assert.Equal(t, rule.BoardApproveRatio, autoCfg.BoardApproveRatio)
	assert.Equal(t, rule.PubOpposeRatio, autoCfg.PubOpposeRatio)
	assert.Equal(t, rule.ProposalAmount, autoCfg.ProposalAmount*types.DefaultCoinPrecision)
	assert.Equal(t, rule.LargeProjectAmount, autoCfg.LargeProjectAmount*types.DefaultCoinPrecision)
	assert.Equal(t, rule.PublicPeriod, autoCfg.PublicPeriod)
}

func revokeProposalRuleTx(parm *auty.RevokeProposalRule) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionRvkPropRule,
		Value: &auty.AutonomyAction_RvkPropRule{RvkPropRule: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func voteProposalRule(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)
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
		priv   string
		appr   bool
		origin []string
	}
	records := []record{
		{priv: PrivKeyA, appr: false},
		{priv: PrivKey1, appr: true, origin: []string{AddrB, AddrC, AddrD}},
	}
	InitMinerAddr(stateDB, []string{AddrB, AddrC, AddrD}, Addr1)

	for i, record := range records {
		opt := &auty.VoteProposalRule{
			ProposalID: proposalID,
			OriginAddr: record.origin,
		}
		if record.appr {
			opt.Vote = auty.AutonomyVoteOption_APPROVE
		} else {
			opt.Vote = auty.AutonomyVoteOption_OPPOSE
		}
		tx, err := voteProposalRuleTx(opt)
		assert.NoError(t, err)
		tx, err = signTx(tx, record.priv)
		assert.NoError(t, err)
		// 设定当前高度为投票高度
		exec.SetEnv(env.startHeight, env.blockTime, env.difficulty)

		receipt, err := exec.Exec(tx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, receipt)
		if save {
			for _, kv := range receipt.KV {
				_ = stateDB.Set(kv.Key, kv.Value)
			}
		}
		receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
		set, err := exec.ExecLocal(tx, receiptData, 1)
		assert.NoError(t, err)
		assert.NotNil(t, set)
		if save {
			for _, kv := range set.KV {
				_ = kvdb.Set(kv.Key, kv.Value)
			}
		}
		// del
		set, err = exec.ExecDelLocal(tx, receiptData, 1)
		assert.NoError(t, err)
		assert.NotNil(t, set)

		// 每次需要重新设置,对于下一个是多个授权地址的需要设置多次
		if i+1 < len(records) {
			for j := 0; j < len(records[i+1].origin); j++ {
				acc := &types.Account{
					Currency: 0,
					Frozen:   total,
				}
				val := types.Encode(acc)
				values := [][]byte{val}
				api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values: values}, nil).Once()
				exec.SetAPI(api)
			}
		}
	}
	// check
	// balance
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	accountAddr := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, int64(0), accountAddr.Frozen)
	accountAddr = accCoin.LoadExecAccount(autonomyAddr, autonomyAddr)
	assert.Equal(t, autoCfg.ProposalAmount*types.DefaultCoinPrecision, accountAddr.Balance)
	// status
	value, err := stateDB.Get(propRuleID(proposalID))
	assert.NoError(t, err)
	cur := &auty.AutonomyProposalRule{}
	err = types.Decode(value, cur)
	assert.NoError(t, err)
	assert.Equal(t, int32(auty.AutonomyStatusTmintPropRule), cur.Status)
	assert.Equal(t, AddrA, cur.Address)
	assert.Equal(t, true, cur.VoteResult.Pass)
	// check rule
	au := newTestAutonomy()
	au.SetStateDB(stateDB)
	au.SetLocalDB(kvdb)
	action := newAction(au, &types.Transaction{}, 0)
	rule, err := action.getActiveRule()
	assert.NoError(t, err)
	assert.Equal(t, rule.BoardApproveRatio, testBoardApproveRatio)
	assert.Equal(t, rule.PubOpposeRatio, testPubOpposeRatio)
	assert.Equal(t, rule.ProposalAmount, testProposalAmount)
	assert.Equal(t, rule.LargeProjectAmount, testLargeProjectAmount)
	assert.Equal(t, rule.PublicPeriod, testPublicPeriod)
	assert.Equal(t, rule.PubAttendRatio, testPubAttendRatio)
	assert.Equal(t, rule.PubApproveRatio, testPubApproveRatio)
}

func voteProposalRuleTx(parm *auty.VoteProposalRule) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionVotePropRule,
		Value: &auty.AutonomyAction_VotePropRule{VotePropRule: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func terminateProposalRule(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)
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
	assert.NoError(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.NoError(t, err)
	exec.SetEnv(env.endHeight+1, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(tx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)
	if save {
		for _, kv := range receipt.KV {
			_ = stateDB.Set(kv.Key, kv.Value)
		}
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(tx, receiptData, 1)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	if save {
		for _, kv := range set.KV {
			_ = kvdb.Set(kv.Key, kv.Value)
		}
	}
	// del
	set, err = exec.ExecDelLocal(tx, receiptData, 1)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	// check
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	accountAddr := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, int64(0), accountAddr.Frozen)
	accountAddr = accCoin.LoadExecAccount(autonomyAddr, autonomyAddr)
	assert.Equal(t, autoCfg.ProposalAmount*types.DefaultCoinPrecision, accountAddr.Balance)

	// check rule
	au := newTestAutonomy()
	au.SetStateDB(stateDB)
	au.SetLocalDB(kvdb)
	action := newAction(au, &types.Transaction{}, 0)
	rule, err := action.getActiveRule()
	assert.NoError(t, err)
	assert.Equal(t, rule.BoardApproveRatio, autoCfg.BoardApproveRatio)
	assert.Equal(t, rule.PubOpposeRatio, autoCfg.PubOpposeRatio)
	assert.Equal(t, rule.ProposalAmount, autoCfg.ProposalAmount*types.DefaultCoinPrecision)
	assert.Equal(t, rule.LargeProjectAmount, autoCfg.LargeProjectAmount*types.DefaultCoinPrecision)
	assert.Equal(t, rule.PublicPeriod, autoCfg.PublicPeriod)
}

func terminateProposalRuleTx(parm *auty.TerminateProposalRule) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTmintPropRule,
		Value: &auty.AutonomyAction_TmintPropRule{TmintPropRule: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
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
	assert.Equal(t, int32(2), log.Ty)
	recpt := &auty.ReceiptProposalRule{}
	err := types.Decode(log.Log, recpt)
	assert.NoError(t, err)
	assert.Equal(t, int32(1800), recpt.Prev.PropRule.Year)
	assert.Equal(t, int32(1900), recpt.Current.PropRule.Year)
}

func TestCopyAutonomyProposalRule(t *testing.T) {
	assert.Nil(t, copyAutonomyProposalRule(nil))
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

	assert.Equal(t, 1900, int(pre.PropRule.Year))
	assert.Equal(t, 1, int(pre.PropRule.Month))
	assert.Equal(t, 100, int(pre.VoteResult.TotalVotes))
	assert.Equal(t, "123", pre.Address)
	assert.Equal(t, 2, int(pre.Status))
	assert.Equal(t, 80, int(pre.PropRule.RuleCfg.BoardApproveRatio))
	assert.Equal(t, 100, int(pre.CurRule.BoardApproveRatio))
}

func TestUpgradeRule(t *testing.T) {
	newRule := upgradeRule(nil, &auty.RuleConfig{})
	assert.Nil(t, newRule)
	cur := &auty.RuleConfig{
		BoardApproveRatio:  2,
		PubOpposeRatio:     3,
		ProposalAmount:     4,
		LargeProjectAmount: 5,
		PublicPeriod:       6,
		PubAttendRatio:     7,
		PubApproveRatio:    8,
	}
	modify := &auty.RuleConfig{
		BoardApproveRatio:  -1,
		PubOpposeRatio:     0,
		ProposalAmount:     -1,
		LargeProjectAmount: 0,
		PublicPeriod:       0,
		PubAttendRatio:     -1,
		PubApproveRatio:    0,
	}
	newRule = upgradeRule(cur, modify)
	assert.NotNil(t, newRule)
	assert.Equal(t, newRule.BoardApproveRatio, cur.BoardApproveRatio)
	assert.Equal(t, newRule.PubOpposeRatio, cur.PubOpposeRatio)
	assert.Equal(t, newRule.ProposalAmount, cur.ProposalAmount)
	assert.Equal(t, newRule.LargeProjectAmount, cur.LargeProjectAmount)
	assert.Equal(t, newRule.PublicPeriod, cur.PublicPeriod)
	assert.Equal(t, newRule.PubAttendRatio, cur.PubAttendRatio)
	assert.Equal(t, newRule.PubApproveRatio, cur.PubApproveRatio)

	modify = &auty.RuleConfig{
		BoardApproveRatio:  20,
		PubOpposeRatio:     30,
		ProposalAmount:     40,
		LargeProjectAmount: 50,
		PublicPeriod:       60,
		PubAttendRatio:     70,
		PubApproveRatio:    80,
	}
	newRule = upgradeRule(cur, modify)
	assert.NotNil(t, newRule)
	assert.Equal(t, newRule.BoardApproveRatio, modify.BoardApproveRatio)
	assert.Equal(t, newRule.PubOpposeRatio, modify.PubOpposeRatio)
	assert.Equal(t, newRule.ProposalAmount, modify.ProposalAmount)
	assert.Equal(t, newRule.LargeProjectAmount, modify.LargeProjectAmount)
	assert.Equal(t, newRule.PublicPeriod, modify.PublicPeriod)
	assert.Equal(t, newRule.PubAttendRatio, modify.PubAttendRatio)
	assert.Equal(t, newRule.PubApproveRatio, modify.PubApproveRatio)
}

func TestTransfer(t *testing.T) {
	env, exec, stateDB, _ := InitEnv()

	opt1 := &auty.TransferFund{
		Amount: types.DefaultCoinPrecision * 190,
	}
	pbtx, err := transferFundTx(opt1)
	assert.NoError(t, err)
	pbtx, err = signTx(pbtx, PrivKeyA)
	assert.NoError(t, err)

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(pbtx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	for _, kv := range receipt.KV {
		_ = stateDB.Set(kv.Key, kv.Value)
	}
	// check
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	accountAddr := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, total-types.DefaultCoinPrecision*190, accountAddr.Balance)
	accountAddr = accCoin.LoadExecAccount(autonomyAddr, autonomyAddr)
	assert.Equal(t, types.DefaultCoinPrecision*190, accountAddr.Balance)
}

func transferFundTx(parm *auty.TransferFund) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTransfer,
		Value: &auty.AutonomyAction_Transfer{Transfer: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
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
	assert.NoError(t, err)
	pbtx, err = signTx(pbtx, PrivKeyA)
	assert.NoError(t, err)

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(pbtx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	for _, kv := range receipt.KV {
		_ = stateDB.Set(kv.Key, kv.Value)
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(pbtx, receiptData, 1)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		_ = kvdb.Set(kv.Key, kv.Value)
	}

	// check
	value, err := kvdb.Get(calcCommentHeight(propID, drivers.HeightIndexStr(env.blockHeight, 1)))
	assert.NoError(t, err)
	cmt := &auty.RelationCmt{}
	err = types.Decode(value, cmt)
	assert.NoError(t, err)
	assert.Equal(t, cmt.Comment, comment)
	assert.Equal(t, cmt.RepHash, Repcmt)
}

func commentPropTx(parm *auty.Comment) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionCommentProp,
		Value: &auty.AutonomyAction_CommentProp{CommentProp: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}
