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

//const (
//	testBoardAttendRatio   int32 = 60
//	testBoardApproveRatio  int32 = 60
//	testPubOpposeRatio     int32 = 30
//	testProposalAmount     int64 = 0
//	testLargeProjectAmount int64 = 1
//)

const (
	testProjectAmount int64 =  types.Coin * 100 // 工程需要资金
)

func InitBoard(stateDB dbm.KV) {
	// add active board
	board := &auty.ProposalBoard{
		Year: 2019,
		Month: 11,
		Day: 1,
		Boards: []string{AddrA, AddrB, AddrC, AddrD},
		StartBlockHeight:1,
		EndBlockHeight:10,
		RealEndBlockHeight:5,
	}
	stateDB.Set(activeBoardID(), types.Encode(board))
}

func InitRule(stateDB dbm.KV) {
	// add active rule
	rule := &auty.RuleConfig{
		BoardAttendRatio: boardAttendRatio,
		BoardApproveRatio: boardApproveRatio,
		PubOpposeRatio: pubOpposeRatio,
		ProposalAmount: proposalAmount,
		LargeProjectAmount: types.Coin *100,
	}
	stateDB.Set(activeRuleID(), types.Encode(rule))
}

func TestPropProject(t *testing.T) {
	env, exec, _, _ := InitEnv()

	opts := []*auty.ProposalProject{
		&auty.ProposalProject{ // check toaddr
			ToAddr: "1111111111",
			StartBlockHeight:  env.blockHeight + 5,
			EndBlockHeight: env.blockHeight + 10,
		},
		&auty.ProposalProject{ // check amount
			Amount: 0,
			ToAddr: AddrA,
			StartBlockHeight:  env.blockHeight + 5,
			EndBlockHeight: env.blockHeight + 10,
		},
		&auty.ProposalProject{ // check StartBlockHeight EndBlockHeight
			Amount: 10,
			ToAddr: AddrA,
			StartBlockHeight:  env.blockHeight-1,
			EndBlockHeight: env.blockHeight-1,
		},
		&auty.ProposalProject{ // check activeboard
			Amount: 100,
			ToAddr: AddrA,
			StartBlockHeight:  env.blockHeight + 5,
			EndBlockHeight: env.blockHeight + 10,
		},
	}

	result := [] error {
		types.ErrInvalidAddress,
	    types.ErrInvalidParam,
	    types.ErrInvalidParam,
		types.ErrNotFound,
	}

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	for i, tcase := range opts {
		pbtx, err := propProjectTx(tcase)
		require.NoError(t, err)
		pbtx, err = signTx(pbtx, PrivKeyA)
		require.NoError(t, err)
		_, err = exec.Exec(pbtx, int(i))
		require.Error(t, err, result[i])

	}
}

func TestRevokeProposalProject(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	InitBoard(stateDB)
	// PropProject
	testPropProject(t, env, exec, stateDB, kvdb, true)
	//RevokeProposalProject
	revokeProposalProject(t, env, exec, stateDB, kvdb, false)
}

func TestVoteProposalProject(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	InitBoard(stateDB)
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
	// PropProject
	testPropProject(t, env, exec, stateDB, kvdb, true)
	//terminateProposalProject
	terminateProposalProject(t, env, exec, stateDB, kvdb, true)
}

func testPropProject(t *testing.T, env *execEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	opt1 :=  &auty.ProposalProject{
		Year: 2019,
		Month: 7,
		Day:     10,
		Amount: testProjectAmount,
		ToAddr: AddrD,
		StartBlockHeight:  env.blockHeight + 5,
		EndBlockHeight: env.blockHeight + 10,
	}
	pbtx, err := propProjectTx(opt1)
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

func propProjectTx(parm *auty.ProposalProject) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPropProject,
		Value: &auty.AutonomyAction_PropProject{PropProject: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func revokeProposalProject(t *testing.T, env *execEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	proposalID := env.txHash
	opt2 :=  &auty.RevokeProposalProject{
		ProposalID:proposalID,
	}
	rtx, err := revokeProposalProjectTx(opt2)
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
	// check Project
	au := &Autonomy{
		drivers.DriverBase{},
	}
	au.SetStateDB(stateDB)
	au.SetLocalDB(kvdb)
	//action := newAction(au, &types.Transaction{}, 0)
}

func revokeProposalProjectTx(parm *auty.RevokeProposalProject) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionRvkPropProject,
		Value: &auty.AutonomyAction_RvkPropProject{RvkPropProject: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func voteProposalProject(t *testing.T, env *execEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
		{PrivKeyB, true},
		{PrivKeyC, true},
		//{PrivKeyD, true},
	}

	for _, record := range records {
		opt :=  &auty.VoteProposalProject{
			ProposalID:proposalID,
			Approve: record.appr,
		}
		tx, err := voteProposalProjectTx(opt)
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
}

func voteProposalProjectTx(parm *auty.VoteProposalProject) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionVotePropProject,
		Value: &auty.AutonomyAction_VotePropProject{VotePropProject: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func checkVoteProposalProjectResult(t *testing.T, stateDB dbm.KV, proposalID string) {
	// check
	// balance
	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(0), account.Frozen)
	account = accCoin.LoadExecAccount(autonomyAddr, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(proposalAmount) - testProjectAmount, account.Balance)
	// status
	value, err := stateDB.Get(propProjectID(proposalID))
	require.NoError(t, err)
	cur := &auty.AutonomyProposalProject{}
	err = types.Decode(value, cur)
	require.NoError(t, err)
	require.Equal(t, int32(auty.AutonomyStatusTmintPropProject), cur.Status)
	require.Equal(t, AddrA, cur.Address)
}

func pubVoteProposalProject(t *testing.T, env *execEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
		{PrivKeyA, true},
		{PrivKeyB, false},
		{PrivKeyC, true},
		//{PrivKeyD, true},
	}

	for _, record := range records {
		opt :=  &auty.PubVoteProposalProject{
			ProposalID:proposalID,
			Oppose: record.appr,
		}
		tx, err := pubVoteProposalProjectTx(opt)
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
}

func checkPubVoteProposalProjectResult(t *testing.T, stateDB dbm.KV, proposalID string) {
	// check
	// balance
	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(0), account.Frozen)
	account = accCoin.LoadExecAccount(autonomyAddr, address.ExecAddress(auty.AutonomyX))
	require.Equal(t, int64(proposalAmount), account.Balance)
	// status
	value, err := stateDB.Get(propProjectID(proposalID))
	require.NoError(t, err)
	cur := &auty.AutonomyProposalProject{}
	err = types.Decode(value, cur)
	require.NoError(t, err)
	require.Equal(t, int32(auty.AutonomyStatusTmintPropProject), cur.Status)
	require.Equal(t, AddrA, cur.Address)
}

func pubVoteProposalProjectTx(parm *auty.PubVoteProposalProject) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPubVotePropProject,
		Value: &auty.AutonomyAction_PubVotePropProject{PubVotePropProject: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func terminateProposalProject(t *testing.T, env *execEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
	opt :=  &auty.TerminateProposalProject{
		ProposalID:proposalID,
	}
	tx, err := terminateProposalProjectTx(opt)
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

	// check Project
	au := &Autonomy{
		drivers.DriverBase{},
	}
	au.SetStateDB(stateDB)
	au.SetLocalDB(kvdb)
	//action := newAction(au, &types.Transaction{}, 0)
}

func terminateProposalProjectTx(parm *auty.TerminateProposalProject) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTmintPropProject,
		Value: &auty.AutonomyAction_TmintPropProject{TmintPropProject: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func TestGetProjectReceiptLog(t *testing.T) {
	pre := &auty.AutonomyProposalProject{
		PropProject: &auty.ProposalProject{Year: 1800, Month: 1},
		CurRule: &auty.RuleConfig{BoardAttendRatio:80},
		Boards: []string{"111", "222", "333"},
		BoardVoteRes: &auty.VoteResult{TotalVotes: 100},
		Status: 1,
		Address:"121",
	}
	cur := &auty.AutonomyProposalProject{
		PropProject: &auty.ProposalProject{Year: 1900, Month: 1},
		CurRule: &auty.RuleConfig{BoardAttendRatio:90},
		Boards: []string{"555", "666", "777"},
		BoardVoteRes: &auty.VoteResult{TotalVotes: 100},
		Status: 2,
		Address:"123",
	}
	log := getProjectReceiptLog(pre, cur, 2)
	require.Equal(t, int32(2), log.Ty)
	recpt := &auty.ReceiptProposalProject{}
	err := types.Decode(log.Log, recpt)
	require.NoError(t, err)
	require.Equal(t, int32(1800), recpt.Prev.PropProject.Year)
	require.Equal(t, int32(1900), recpt.Current.PropProject.Year)
	require.Equal(t, int32(80), recpt.Prev.CurRule.BoardAttendRatio)
	require.Equal(t, int32(90), recpt.Current.CurRule.BoardAttendRatio)
	require.Equal(t, []string{"111", "222", "333"}, recpt.Prev.Boards)
	require.Equal(t, []string{"555", "666", "777"}, recpt.Current.Boards)
}

func TestCopyAutonomyProposalProject(t *testing.T) {
	require.Nil(t, copyAutonomyProposalProject(nil))
	cur := &auty.AutonomyProposalProject{
		PropProject: &auty.ProposalProject{Year: 1800, Month: 1},
		CurRule: &auty.RuleConfig{BoardAttendRatio:80},
		Boards: []string{"111", "222", "333"},
		BoardVoteRes: &auty.VoteResult{TotalVotes: 100},
		PubVote: &auty.PublicVote{Publicity:true},
		Status: 2,
		Address:"123",
	}
	pre := copyAutonomyProposalProject(cur)
	cur.PropProject.Year = 1900
	cur.PropProject.Month = 2
	cur.CurRule.BoardAttendRatio = 90
	cur.Boards = []string{"555", "666", "777"}
	cur.BoardVoteRes.TotalVotes = 90
	cur.PubVote.Publicity = false
	cur.Address = "234"
	cur.Status = 1

	require.Equal(t, 1800, int(pre.PropProject.Year))
	require.Equal(t, 1, int(pre.PropProject.Month))
	require.Equal(t, []string{"111", "222", "333"}, pre.Boards)
	require.Equal(t, 80, int(pre.CurRule.BoardAttendRatio))
	require.Equal(t, "123", pre.Address)
	require.Equal(t, 2, int(pre.Status))
	require.Equal(t, 100, int(pre.BoardVoteRes.TotalVotes))
	require.Equal(t, true, pre.PubVote.Publicity)
	require.Equal(t, []string{"555", "666", "777"}, cur.Boards)
}