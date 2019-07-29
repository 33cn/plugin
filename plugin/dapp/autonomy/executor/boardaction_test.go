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
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	dbmock "github.com/33cn/chain33/common/db/mocks"
	commonlog "github.com/33cn/chain33/common/log"
	_ "github.com/33cn/chain33/system"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ExecEnv exec environment
type ExecEnv struct {
	blockTime   int64 // 1539918074
	blockHeight int64
	index       int
	difficulty  uint64
	txHash      string
	startHeight int64
	endHeight   int64
}

var (
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	AddrA    = "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
	AddrB    = "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
	AddrC    = "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
	AddrD    = "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"

	boards = []string{"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4", "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR", "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"}
	total  = types.Coin * 30000
)

func init() {
	commonlog.SetLogLevel("error")
}

// InitEnv 初始化环境
func InitEnv() (*ExecEnv, drivers.Driver, dbm.KV, dbm.KVDB) {
	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    AddrA,
	}

	accountB := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    AddrB,
	}

	accountC := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    AddrC,
	}

	accountD := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    AddrD,
	}

	env := &ExecEnv{
		blockTime:   1539918074,
		blockHeight: 10,
		index:       2,
		difficulty:  1539918074,
		txHash:      "",
	}

	stateDB, _ := dbm.NewGoMemDB("state", "state", 100)
	_, _, kvdb := util.CreateTestDB()

	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	accCoin.SaveAccount(&accountA)
	accCoin.SaveExecAccount(address.ExecAddress(auty.AutonomyX), &accountA)
	accCoin.SaveAccount(&accountB)
	accCoin.SaveAccount(&accountC)
	accCoin.SaveAccount(&accountD)
	//total ticket balance
	accCoin.SaveAccount(&types.Account{Balance: total * 4,
		Frozen: 0,
		Addr:   "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"})

	exec := newAutonomy()
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	return env, exec, stateDB, kvdb
}

func TestRevokeProposalBoard(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	// PropBoard
	testPropBoard(t, env, exec, stateDB, kvdb, true)
	//RevokeProposalBoard
	revokeProposalBoard(t, env, exec, stateDB, kvdb, false)
}

func TestVoteProposalBoard(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	// PropBoard
	testPropBoard(t, env, exec, stateDB, kvdb, true)
	//voteProposalBoard
	voteProposalBoard(t, env, exec, stateDB, kvdb, true)
}

func TestTerminateProposalBoard(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	// PropBoard
	testPropBoard(t, env, exec, stateDB, kvdb, true)
	//terminateProposalBoard
	terminateProposalBoard(t, env, exec, stateDB, kvdb, true)
}

func testPropBoard(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	opt1 := &auty.ProposalBoard{
		Year:             2019,
		Month:            7,
		Day:              10,
		Boards:           boards,
		StartBlockHeight: env.blockHeight + 5,
		EndBlockHeight:   env.blockHeight + 10,
	}
	pbtx, err := propBoardTx(opt1)
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

	// local
	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(pbtx, receiptData, int(1))
	require.NoError(t, err)
	require.NotNil(t, set)
	if save {
		for _, kv := range set.KV {
			kvdb.Set(kv.Key, kv.Value)
		}
	}
	// del
	set, err = exec.ExecDelLocal(pbtx, receiptData, int(1))
	require.NoError(t, err)
	require.NotNil(t, set)

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

func propBoardTx(parm *auty.ProposalBoard) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPropBoard,
		Value: &auty.AutonomyAction_PropBoard{PropBoard: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func revokeProposalBoard(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	proposalID := env.txHash
	opt2 := &auty.RevokeProposalBoard{
		ProposalID: proposalID,
	}
	rtx, err := revokeProposalBoardTx(opt2)
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
}

func revokeProposalBoardTx(parm *auty.RevokeProposalBoard) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionRvkPropBoard,
		Value: &auty.AutonomyAction_RvkPropBoard{RvkPropBoard: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func voteProposalBoard(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
		opt := &auty.VoteProposalBoard{
			ProposalID: proposalID,
			Approve:    record.appr,
		}
		tx, err := voteProposalBoardTx(opt)
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
	require.Equal(t, int64(proposalAmount), account.Balance)
	// status
	value, err := stateDB.Get(propBoardID(proposalID))
	require.NoError(t, err)
	cur := &auty.AutonomyProposalBoard{}
	err = types.Decode(value, cur)
	require.NoError(t, err)
	require.Equal(t, int32(auty.AutonomyStatusTmintPropBoard), cur.Status)
	require.Equal(t, AddrA, cur.Address)
	require.Equal(t, true, cur.VoteResult.Pass)
}

func voteProposalBoardTx(parm *auty.VoteProposalBoard) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionVotePropBoard,
		Value: &auty.AutonomyAction_VotePropBoard{VotePropBoard: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func terminateProposalBoard(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
	opt := &auty.TerminateProposalBoard{
		ProposalID: proposalID,
	}
	tx, err := terminateProposalBoardTx(opt)
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
}

func terminateProposalBoardTx(parm *auty.TerminateProposalBoard) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTmintPropBoard,
		Value: &auty.AutonomyAction_TmintPropBoard{TmintPropBoard: parm},
	}
	return types.CreateFormatTx(types.ExecName(auty.AutonomyX), types.Encode(val))
}

func TestGetStartHeightVoteAccount(t *testing.T) {
	at := newAutonomy().(*Autonomy)
	at.SetLocalDB(new(dbmock.KVDB))

	api := new(apimock.QueueProtocolAPI)
	at.SetAPI(api)
	tx := &types.Transaction{}
	action := newAction(at, tx, 0)

	addr := "1JmFaA6unrCFYEWPGRi7uuXY1KthTJxJEP"
	api.On("StoreList", mock.Anything).Return(&types.StoreListReply{}, nil)
	api.On("GetLastHeader", mock.Anything).Return(&types.Header{StateHash: []byte("")}, nil)
	acc := &types.Account{
		Currency: 0,
		Balance:  types.Coin,
	}
	val := types.Encode(acc)
	values := [][]byte{val}
	api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values: values}, nil)
	hear := &types.Header{StateHash: []byte("")}
	api.On("GetHeaders", mock.Anything).
		Return(&types.Headers{
			Items: []*types.Header{hear}}, nil)
	account, err := action.getStartHeightVoteAccount(addr, "", 0)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, types.Coin, account.Balance)
}

func TestGetReceiptLog(t *testing.T) {
	pre := &auty.AutonomyProposalBoard{
		PropBoard:  &auty.ProposalBoard{Year: 1800, Month: 1},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status:     1,
		Address:    "121",
	}
	cur := &auty.AutonomyProposalBoard{
		PropBoard:  &auty.ProposalBoard{Year: 1900, Month: 1},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status:     2,
		Address:    "123",
	}
	log := getReceiptLog(pre, cur, 2)
	require.Equal(t, int32(2), log.Ty)
	recpt := &auty.ReceiptProposalBoard{}
	err := types.Decode(log.Log, recpt)
	require.NoError(t, err)
	require.Equal(t, int32(1800), recpt.Prev.PropBoard.Year)
	require.Equal(t, int32(1900), recpt.Current.PropBoard.Year)
}

func TestCopyAutonomyProposalBoard(t *testing.T) {
	require.Nil(t, copyAutonomyProposalBoard(nil))
	cur := &auty.AutonomyProposalBoard{
		PropBoard:  &auty.ProposalBoard{Year: 1900, Month: 1},
		CurRule:    &auty.RuleConfig{BoardAttendRatio: 100},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status:     2,
		Address:    "123",
	}
	pre := copyAutonomyProposalBoard(cur)
	cur.PropBoard.Year = 1800
	cur.PropBoard.Month = 2
	cur.CurRule.BoardAttendRatio = 90
	cur.VoteResult.TotalVotes = 50
	cur.Address = "234"
	cur.Status = 1

	require.Equal(t, 1900, int(pre.PropBoard.Year))
	require.Equal(t, 1, int(pre.PropBoard.Month))
	require.Equal(t, 100, int(pre.CurRule.BoardAttendRatio))
	require.Equal(t, 100, int(pre.VoteResult.TotalVotes))
	require.Equal(t, "123", pre.Address)
	require.Equal(t, 2, int(pre.Status))
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.New(types.GetSignName(auty.AutonomyX, signType))
	if err != nil {
		return tx, err
	}

	bytes, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		return tx, err
	}

	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return tx, err
	}

	tx.Sign(int32(signType), privKey)
	return tx, nil
}
