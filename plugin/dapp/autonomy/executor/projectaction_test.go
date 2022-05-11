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
	testProjectAmount = types.DefaultCoinPrecision * 100  // 工程需要资金
	testFundAmount    = types.DefaultCoinPrecision * 1000 // 工程需要资金
)

func InitBoard(stateDB dbm.KV) {
	// add active board
	act := &auty.ActiveBoard{
		Boards: boards,
	}
	_ = stateDB.Set(activeBoardID(), types.Encode(act))
}

func InitRule(stateDB dbm.KV) {
	// add active rule
	rule := &auty.RuleConfig{
		BoardApproveRatio:  autoCfg.BoardApproveRatio,
		PubOpposeRatio:     autoCfg.PubOpposeRatio,
		ProposalAmount:     autoCfg.ProposalAmount * types.DefaultCoinPrecision,
		LargeProjectAmount: 100 * types.DefaultCoinPrecision,
		PublicPeriod:       autoCfg.PublicPeriod,
	}
	_ = stateDB.Set(activeRuleID(), types.Encode(rule))
}

func InitFund(stateDB dbm.KV, amount int64) {
	accountA := types.Account{
		Balance: amount,
		Frozen:  0,
		Addr:    autonomyAddr,
	}
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	//accCoin.ExecIssueCoins(autonomyAddr, amount)
	accCoin.SaveAccount(&accountA)
}

func TestPropProject(t *testing.T) {
	env, exec, stateDB, _ := InitEnv()

	opts := []*auty.ProposalProject{
		{ // check toaddr
			ToAddr:           "1111111111",
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // check amount
			Amount:           0,
			ToAddr:           AddrA,
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // check StartBlockHeight EndBlockHeight
			Amount:           10,
			ToAddr:           AddrA,
			StartBlockHeight: env.blockHeight - 1,
			EndBlockHeight:   env.blockHeight - 1,
		},
		{ // check activeboard
			Amount:           100,
			ToAddr:           AddrA,
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // checkPeriodAmount
			Amount:           100,
			ToAddr:           AddrA,
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // ErrSetBlockHeight
			Amount:           100,
			ToAddr:           AddrA,
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.PropEndBlockPeriod + 10,
		},
	}

	result := []error{
		types.ErrInvalidAddress,
		types.ErrInvalidParam,
		auty.ErrSetBlockHeight,
		types.ErrNotFound,
		auty.ErrNoPeriodAmount,
		auty.ErrSetBlockHeight,
	}

	exec.SetStateDB(stateDB)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	for i, tcase := range opts {
		pbtx, err := propProjectTx(tcase)
		assert.NoError(t, err)
		pbtx, err = signTx(pbtx, PrivKeyA)
		assert.NoError(t, err)
		if i == 4 {
			act := &auty.ActiveBoard{
				Boards: boards,
				Amount: autoCfg.MaxBoardPeriodAmount * types.DefaultCoinPrecision,
			}
			err := stateDB.Set(activeBoardID(), types.Encode(act))
			assert.NoError(t, err)
		}
		_, err = exec.Exec(pbtx, i)
		assert.Equal(t, err, result[i])
	}
}

func TestRevokeProposalProject(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	InitBoard(stateDB)
	InitFund(stateDB, testFundAmount)
	// PropProject
	testPropProject(t, env, exec, stateDB, kvdb, true)
	//RevokeProposalProject
	revokeProposalProject(t, env, exec, stateDB, kvdb, false)
}

func TestVoteProposalProject(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	InitBoard(stateDB)
	InitFund(stateDB, testFundAmount)
	// PropProject
	testPropProject(t, env, exec, stateDB, kvdb, true)
	//voteProposalProject
	voteProposalProject(t, env, exec, stateDB, kvdb, true)
	// check
	checkVoteProposalProjectResult(t, stateDB, env.txHash)
}

func TestPubVoteProposalProject(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	InitBoard(stateDB)
	InitRule(stateDB)
	InitFund(stateDB, testFundAmount)
	// PropProject
	testPropProject(t, env, exec, stateDB, kvdb, true)
	// voteProposalProject
	voteProposalProject(t, env, exec, stateDB, kvdb, true)
	// pubVoteProposalProject
	pubVoteProposalProject(t, env, exec, stateDB, kvdb, true) // 未通过全体持票人投票
	// terminate
	// check
	checkPubVoteProposalProjectResult(t, stateDB, env.txHash)
}

func TestTerminateProposalProject(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	InitBoard(stateDB)
	InitFund(stateDB, testFundAmount)
	// PropProject
	testPropProject(t, env, exec, stateDB, kvdb, true)
	//terminateProposalProject
	terminateProposalProject(t, env, exec, stateDB, kvdb, true)
}

func TestBoardPeriodAmount(t *testing.T) {
	env, exec, stateDB, _ := InitEnv()
	InitFund(stateDB, testProjectAmount)
	act := &auty.ActiveBoard{
		Boards:      boards,
		Amount:      autoCfg.MaxBoardPeriodAmount*types.DefaultCoinPrecision - 100,
		StartHeight: 10,
	}
	_ = stateDB.Set(activeBoardID(), types.Encode(act))

	opt1 := &auty.ProposalProject{
		Year:             2019,
		Month:            7,
		Day:              10,
		Amount:           testProjectAmount,
		ToAddr:           AddrD,
		StartBlockHeight: env.blockHeight + autoCfg.BoardPeriod + 5,
		EndBlockHeight:   env.blockHeight + autoCfg.BoardPeriod + autoCfg.StartEndBlockPeriod + 10,
	}
	pbtx, err := propProjectTx(opt1)
	assert.NoError(t, err)
	pbtx, err = signTx(pbtx, PrivKeyA)
	assert.NoError(t, err)

	exec.SetEnv(env.blockHeight+autoCfg.BoardPeriod+1, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(pbtx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	for _, kv := range receipt.KV {
		_ = stateDB.Set(kv.Key, kv.Value)
	}

	// check
	value, err := stateDB.Get(activeBoardID())
	assert.NoError(t, err)
	nact := &auty.ActiveBoard{}
	_ = types.Decode(value, nact)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), nact.Amount)
}

func testPropProject(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	opt1 := &auty.ProposalProject{
		Year:             2019,
		Month:            7,
		Day:              10,
		Amount:           testProjectAmount,
		ToAddr:           AddrD,
		StartBlockHeight: env.blockHeight + 5,
		EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
	}
	pbtx, err := propProjectTx(opt1)
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

func propProjectTx(parm *auty.ProposalProject) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPropProject,
		Value: &auty.AutonomyAction_PropProject{PropProject: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func revokeProposalProject(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	proposalID := env.txHash
	opt2 := &auty.RevokeProposalProject{
		ProposalID: proposalID,
	}
	rtx, err := revokeProposalProjectTx(opt2)
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
}

func revokeProposalProjectTx(parm *auty.RevokeProposalProject) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionRvkPropProject,
		Value: &auty.AutonomyAction_RvkPropProject{RvkPropProject: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func voteProposalProject(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
		priv string
		appr bool
	}
	records := []record{
		{PrivKeyA, true},
		{PrivKeyB, true},
		{PrivKeyC, true},
		{PrivKeyD, true},

		{PrivKey1, false},
		{PrivKey2, false},
		{PrivKey3, false},
		{PrivKey4, true},
		{PrivKey5, true},
		{PrivKey6, true},
		{PrivKey7, true},
		{PrivKey8, true},
		{PrivKey9, true},
		{PrivKey10, true},
	}

	for _, record := range records {
		opt := &auty.VoteProposalProject{
			ProposalID: proposalID,
		}
		if record.appr {
			opt.Vote = auty.AutonomyVoteOption_APPROVE
		} else {
			opt.Vote = auty.AutonomyVoteOption_OPPOSE
		}
		tx, err := voteProposalProjectTx(opt)
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
}

func voteProposalProjectTx(parm *auty.VoteProposalProject) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionVotePropProject,
		Value: &auty.AutonomyAction_VotePropProject{VotePropProject: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func checkVoteProposalProjectResult(t *testing.T, stateDB dbm.KV, proposalID string) {
	// check
	// status
	value, err := stateDB.Get(propProjectID(proposalID))
	assert.NoError(t, err)
	cur := &auty.AutonomyProposalProject{}
	err = types.Decode(value, cur)
	assert.NoError(t, err)
	assert.Equal(t, int32(auty.AutonomyStatusTmintPropProject), cur.Status)
	assert.Equal(t, AddrA, cur.Address)
	// balance
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	accountAddr := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, int64(0), accountAddr.Frozen)
	accountAddr = accCoin.LoadExecAccount(autonomyAddr, autonomyAddr)
	assert.Equal(t, autoCfg.ProposalAmount*types.DefaultCoinPrecision, accountAddr.Balance)
	accountAddr = accCoin.LoadExecAccount(AddrD, autonomyAddr)
	assert.Equal(t, testProjectAmount, accountAddr.Balance)
	// 更新董事会累计审批金
	value, err = stateDB.Get(activeBoardID())
	assert.NoError(t, err)
	aBd := &auty.ActiveBoard{}
	err = types.Decode(value, aBd)
	assert.NoError(t, err)
	assert.Equal(t, testProjectAmount, aBd.Amount)
}

func pubVoteProposalProject(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
	// 3人参与投票，2人赞成票，1人反对票
	type record struct {
		priv   string
		appr   bool
		origin []string
	}
	records := []record{
		{priv: PrivKeyA, appr: false},
		{priv: PrivKey1, appr: true, origin: []string{AddrB, AddrC}},
	}
	InitMinerAddr(stateDB, []string{AddrB, AddrC}, Addr1)

	for i, record := range records {
		opt := &auty.PubVoteProposalProject{
			ProposalID: proposalID,
			Oppose:     record.appr,
			OriginAddr: record.origin,
		}
		tx, err := pubVoteProposalProjectTx(opt)
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
}

func checkPubVoteProposalProjectResult(t *testing.T, stateDB dbm.KV, proposalID string) {
	// check
	// status
	value, err := stateDB.Get(propProjectID(proposalID))
	assert.NoError(t, err)
	cur := &auty.AutonomyProposalProject{}
	err = types.Decode(value, cur)
	assert.NoError(t, err)
	assert.Equal(t, int32(auty.AutonomyStatusTmintPropProject), cur.Status)
	assert.Equal(t, AddrA, cur.Address)
	// balance
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	accountAddr := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, int64(0), accountAddr.Frozen)
	accountAddr = accCoin.LoadExecAccount(AddrD, autonomyAddr)
	assert.Equal(t, int64(0), accountAddr.Balance)
	accountAddr = accCoin.LoadExecAccount(autonomyAddr, autonomyAddr)
	assert.Equal(t, autoCfg.ProposalAmount*types.DefaultCoinPrecision, accountAddr.Balance)

	// 更新董事会累计审批金
	value, err = stateDB.Get(activeBoardID())
	assert.NoError(t, err)
	aBd := &auty.ActiveBoard{}
	err = types.Decode(value, aBd)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), aBd.Amount)
}

func pubVoteProposalProjectTx(parm *auty.PubVoteProposalProject) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPubVotePropProject,
		Value: &auty.AutonomyAction_PubVotePropProject{PubVotePropProject: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func terminateProposalProject(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
	opt := &auty.TerminateProposalProject{
		ProposalID: proposalID,
	}
	tx, err := terminateProposalProjectTx(opt)
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
}

func terminateProposalProjectTx(parm *auty.TerminateProposalProject) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTmintPropProject,
		Value: &auty.AutonomyAction_TmintPropProject{TmintPropProject: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func TestGetProjectReceiptLog(t *testing.T) {
	pre := &auty.AutonomyProposalProject{
		PropProject:  &auty.ProposalProject{Year: 1800, Month: 1},
		CurRule:      &auty.RuleConfig{BoardApproveRatio: 80},
		Boards:       []string{"111", "222", "333"},
		BoardVoteRes: &auty.VoteResult{TotalVotes: 100},
		Status:       1,
		Address:      "121",
	}
	cur := &auty.AutonomyProposalProject{
		PropProject:  &auty.ProposalProject{Year: 1900, Month: 1},
		CurRule:      &auty.RuleConfig{BoardApproveRatio: 90},
		Boards:       []string{"555", "666", "777"},
		BoardVoteRes: &auty.VoteResult{TotalVotes: 100},
		Status:       2,
		Address:      "123",
	}
	log := getProjectReceiptLog(pre, cur, 2)
	assert.Equal(t, int32(2), log.Ty)
	recpt := &auty.ReceiptProposalProject{}
	err := types.Decode(log.Log, recpt)
	assert.NoError(t, err)
	assert.Equal(t, int32(1800), recpt.Prev.PropProject.Year)
	assert.Equal(t, int32(1900), recpt.Current.PropProject.Year)
	assert.Equal(t, int32(80), recpt.Prev.CurRule.BoardApproveRatio)
	assert.Equal(t, int32(90), recpt.Current.CurRule.BoardApproveRatio)
	assert.Equal(t, []string{"111", "222", "333"}, recpt.Prev.Boards)
	assert.Equal(t, []string{"555", "666", "777"}, recpt.Current.Boards)
}

func TestCopyAutonomyProposalProject(t *testing.T) {
	assert.Nil(t, copyAutonomyProposalProject(nil))
	cur := &auty.AutonomyProposalProject{
		PropProject:  &auty.ProposalProject{Year: 1800, Month: 1},
		CurRule:      &auty.RuleConfig{BoardApproveRatio: 80},
		Boards:       []string{"111", "222", "333"},
		BoardVoteRes: &auty.VoteResult{TotalVotes: 100},
		PubVote:      &auty.PublicVote{Publicity: true},
		Status:       2,
		Address:      "123",
	}
	pre := copyAutonomyProposalProject(cur)
	cur.PropProject.Year = 1900
	cur.PropProject.Month = 2
	cur.CurRule.BoardApproveRatio = 90
	cur.Boards = []string{"555", "666", "777"}
	cur.BoardVoteRes.TotalVotes = 90
	cur.PubVote.Publicity = false
	cur.Address = "234"
	cur.Status = 1

	assert.Equal(t, 1800, int(pre.PropProject.Year))
	assert.Equal(t, 1, int(pre.PropProject.Month))
	assert.Equal(t, []string{"111", "222", "333"}, pre.Boards)
	assert.Equal(t, 80, int(pre.CurRule.BoardApproveRatio))
	assert.Equal(t, "123", pre.Address)
	assert.Equal(t, 2, int(pre.Status))
	assert.Equal(t, 100, int(pre.BoardVoteRes.TotalVotes))
	assert.Equal(t, true, pre.PubVote.Publicity)
	assert.Equal(t, []string{"555", "666", "777"}, cur.Boards)
}
