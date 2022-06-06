// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/util"

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

var (
	Addr18    = "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
	PrivKey18 = "0x56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138"
	Addr19    = "1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
	PrivKey19 = "0x2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989"
)

func InitChange(t *testing.T, stateDB dbm.KV) {
	act := &auty.ActiveBoard{
		Boards:    boards,
		Revboards: []string{Addr18},
	}
	err := stateDB.Set(activeBoardID(), types.Encode(act))
	assert.NoError(t, err)
}

func TestRevokeProposalChange(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	InitChange(t, stateDB)
	// PropChange
	testPropChange(t, env, exec, stateDB, kvdb, true)
	//RevokeProposalChange
	revokeProposalChange(t, env, exec, stateDB, kvdb, false)
}

func TestVoteProposalChange(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	InitChange(t, stateDB)
	// PropChange
	testPropChange(t, env, exec, stateDB, kvdb, true)
	//voteProposalChange
	voteProposalChange(t, env, exec, stateDB, kvdb, true)
}

func TestErrorVoteProposalChange(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	InitChange(t, stateDB)
	// PropChange
	testPropChange(t, env, exec, stateDB, kvdb, true)
	//voteProposalChange
	voteErrorProposalChange(t, env, exec, stateDB, kvdb, true)
}

func TestTerminateProposalChange(t *testing.T) {
	env, exec, stateDB, kvdb := InitEnv()
	InitChange(t, stateDB)
	// PropChange
	testPropChange(t, env, exec, stateDB, kvdb, true)
	//terminateProposalChange
	terminateProposalChange(t, env, exec, stateDB, kvdb, true)
}

func testPropChange(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	opt1 := &auty.ProposalChange{
		Year:             2019,
		Month:            7,
		Day:              10,
		Changes:          []*auty.Change{{Cancel: true, Addr: Addr19}},
		StartBlockHeight: env.blockHeight + 5,
		EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
	}
	pbtx, err := propChangeTx(opt1)
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

	// local
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

func propChangeTx(parm *auty.ProposalChange) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPropChange,
		Value: &auty.AutonomyAction_PropChange{PropChange: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func revokeProposalChange(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	proposalID := env.txHash
	opt2 := &auty.RevokeProposalChange{
		ProposalID: proposalID,
	}
	rtx, err := revokeProposalChangeTx(opt2)
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

func revokeProposalChangeTx(parm *auty.RevokeProposalChange) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionRvkPropChange,
		Value: &auty.AutonomyAction_RvkPropChange{RvkPropChange: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func voteProposalChange(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
	type record struct {
		priv string
		appr bool
	}
	records := []record{
		//{PrivKeyA, false},
		{PrivKeyB, true},
		{PrivKeyC, true},
		{PrivKeyD, true},

		{PrivKey1, false},
		{PrivKey2, false},
		{PrivKey3, false},
		{PrivKey4, false},
		{PrivKey5, true},
		{PrivKey6, true},
		{PrivKey7, true},
		{PrivKey8, true},
		{PrivKey9, true},
		{PrivKey10, true},
		{PrivKey11, true},
		{PrivKey12, true},
	}

	for _, record := range records {
		opt := &auty.VoteProposalChange{
			ProposalID: proposalID,
		}
		if record.appr {
			opt.Vote = auty.AutonomyVoteOption_APPROVE
		} else {
			opt.Vote = auty.AutonomyVoteOption_OPPOSE
		}
		tx, err := voteProposalChangeTx(opt)
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
	// check
	// balance
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	accountAddr := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, int64(0), accountAddr.Frozen)
	accountAddr = accCoin.LoadExecAccount(autonomyAddr, autonomyAddr)
	assert.Equal(t, autoCfg.ProposalAmount*types.DefaultCoinPrecision, accountAddr.Balance)
	// status
	value, err := stateDB.Get(propChangeID(proposalID))
	assert.NoError(t, err)
	cur := &auty.AutonomyProposalChange{}
	err = types.Decode(value, cur)
	assert.NoError(t, err)
	assert.Equal(t, int32(auty.AutonomyStatusTmintPropChange), cur.Status)
	assert.Equal(t, AddrA, cur.Address)
	assert.Equal(t, true, cur.VoteResult.Pass)

	value, err = stateDB.Get(activeBoardID())
	assert.NoError(t, err)
	act := &auty.ActiveBoard{}
	err = types.Decode(value, act)
	assert.NoError(t, err)
	assert.Equal(t, act.Revboards[0], Addr18)
	assert.Equal(t, len(act.Boards), len(boards))
}

func voteErrorProposalChange(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
	type record struct {
		priv string
		appr bool
	}
	records := []record{
		{PrivKey18, false},
		{PrivKey19, false},
		{PrivKeyA, false},
		{PrivKeyB, true},
		{PrivKeyC, true},
		{PrivKeyD, true},

		{PrivKey1, false},
		{PrivKey2, false},
		{PrivKey3, false},
		{PrivKey4, false},
		{PrivKey5, true},
		{PrivKey6, true},
		{PrivKey7, true},
		{PrivKey8, true},
		{PrivKey9, true},
		{PrivKey10, true},
		{PrivKey11, true},
		{PrivKey12, true},
	}

	for i, record := range records {
		opt := &auty.VoteProposalChange{
			ProposalID: proposalID,
		}
		if record.appr {
			opt.Vote = auty.AutonomyVoteOption_APPROVE
		} else {
			opt.Vote = auty.AutonomyVoteOption_OPPOSE
		}
		tx, err := voteProposalChangeTx(opt)
		assert.NoError(t, err)
		tx, err = signTx(tx, record.priv)
		assert.NoError(t, err)
		// 设定当前高度为投票高度
		exec.SetEnv(env.startHeight, env.blockTime, env.difficulty)

		receipt, err := exec.Exec(tx, 1)
		if i < 2 {
			assert.Equal(t, err, auty.ErrNoActiveBoard)
		} else {
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
	// check
	// balance
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	accountAddr := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, int64(0), accountAddr.Frozen)
	accountAddr = accCoin.LoadExecAccount(autonomyAddr, autonomyAddr)
	assert.Equal(t, autoCfg.ProposalAmount*types.DefaultCoinPrecision, accountAddr.Balance)
	// status
	value, err := stateDB.Get(propChangeID(proposalID))
	assert.NoError(t, err)
	cur := &auty.AutonomyProposalChange{}
	err = types.Decode(value, cur)
	assert.NoError(t, err)
	assert.Equal(t, int32(auty.AutonomyStatusTmintPropChange), cur.Status)
	assert.Equal(t, AddrA, cur.Address)
	assert.Equal(t, true, cur.VoteResult.Pass)

	value, err = stateDB.Get(activeBoardID())
	assert.NoError(t, err)
	act := &auty.ActiveBoard{}
	err = types.Decode(value, act)
	assert.NoError(t, err)
	assert.Equal(t, act.Revboards[0], Addr18)
	assert.Equal(t, len(act.Boards), len(boards))
}

func voteProposalChangeTx(parm *auty.VoteProposalChange) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionVotePropChange,
		Value: &auty.AutonomyAction_VotePropChange{VotePropChange: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func terminateProposalChange(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
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
	opt := &auty.TerminateProposalChange{
		ProposalID: proposalID,
	}
	tx, err := terminateProposalChangeTx(opt)
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
	assert.Equal(t, int64(0), accountAddr.Frozen)
}

func terminateProposalChangeTx(parm *auty.TerminateProposalChange) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTmintPropChange,
		Value: &auty.AutonomyAction_TmintPropChange{TmintPropChange: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func TestGetChangeReceiptLog(t *testing.T) {
	pre := &auty.AutonomyProposalChange{
		PropChange: &auty.ProposalChange{Year: 1800, Month: 1},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status:     1,
		Address:    "121",
	}
	cur := &auty.AutonomyProposalChange{
		PropChange: &auty.ProposalChange{Year: 1900, Month: 1},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status:     2,
		Address:    "123",
	}
	log := getChangeReceiptLog(pre, cur, 2)
	assert.Equal(t, int32(2), log.Ty)
	recpt := &auty.ReceiptProposalChange{}
	err := types.Decode(log.Log, recpt)
	assert.NoError(t, err)
	assert.Equal(t, int32(1800), recpt.Prev.PropChange.Year)
	assert.Equal(t, int32(1900), recpt.Current.PropChange.Year)
}

func TestCheckChangeable(t *testing.T) {
	at := newTestAutonomy()
	tx := &types.Transaction{}
	action := newAction(at, tx, 0)

	act := &auty.ActiveBoard{
		Boards: boards,
	}

	// 正常撤销一个地址
	changes := []*auty.Change{{Cancel: true, Addr: AddrA}}
	cur, err := action.checkChangeable(act, changes)
	assert.NoError(t, err)
	assert.Equal(t, len(cur.Boards), len(boards)-1)
	assert.Equal(t, cur.Revboards[0], AddrA)

	// 恢复撤销地址
	changes = []*auty.Change{
		{Cancel: false, Addr: AddrA},
	}
	ncur, err := action.checkChangeable(cur, changes)
	assert.NoError(t, err)
	assert.Equal(t, len(ncur.Boards), len(boards))
	assert.Equal(t, len(ncur.Revboards), 0)

	// 撤销两个地址，撤销不够最小minBoards
	changes = []*auty.Change{
		{Cancel: true, Addr: AddrA},
		{Cancel: true, Addr: AddrB},
	}
	_, err = action.checkChangeable(act, changes)
	assert.Equal(t, err, auty.ErrBoardNumber)

	// 恢复一个没有被撤销的地址
	changes = []*auty.Change{
		{Cancel: false, Addr: AddrA},
	}
	_, err = action.checkChangeable(act, changes)
	assert.Equal(t, err, auty.ErrChangeBoardAddr)

	// 撤销一个不存在地址
	changes = []*auty.Change{
		{Cancel: true, Addr: "1111111111"},
	}
	_, err = action.checkChangeable(act, changes)
	assert.Equal(t, err, auty.ErrChangeBoardAddr)
}

func TestReplaceBoard(t *testing.T) {
	at := newTestAutonomy()
	signer := util.HexToPrivkey(PrivKey17)
	tx := &types.Transaction{}
	tx.Sign(types.SECP256K1, signer)
	action := newAction(at, tx, 0)

	act := &auty.ActiveBoard{
		Boards: boards,
	}

	// 一个成员只允许替换一个新的
	changes := []*auty.Change{
		{Cancel: true, Addr: Addr18},
		{Cancel: true, Addr: Addr19},
	}
	_, err := action.replaceBoard(act, changes)
	assert.ErrorIs(t, err, types.ErrInvalidParam)

	// 只允许替换，不允许恢复操作
	changes = []*auty.Change{{Cancel: false, Addr: Addr18}}
	_, err = action.replaceBoard(act, changes)
	assert.ErrorIs(t, err, types.ErrInvalidParam)

	// 替换一个不存在地址
	changes = []*auty.Change{{Cancel: true, Addr: "0x1111111111"}}
	_, err = action.replaceBoard(act, changes)
	assert.NotNil(t, err)

	// 正常替换一个地址
	changes = []*auty.Change{{Cancel: true, Addr: Addr18}}
	cur, err := action.replaceBoard(act, changes)
	assert.NoError(t, err)
	assert.Equal(t, cur.Boards[20], Addr18)
	assert.Equal(t, cur.Revboards[0], Addr17)
}

func TestCopyAutonomyProposalChange(t *testing.T) {
	assert.Nil(t, copyAutonomyProposalChange(nil))
	cur := &auty.AutonomyProposalChange{
		PropChange: &auty.ProposalChange{
			Year:    1900,
			Month:   1,
			Changes: []*auty.Change{{Cancel: true, Addr: "11"}, {Cancel: false, Addr: "12"}},
		},
		CurRule: &auty.RuleConfig{BoardApproveRatio: 100},
		Board: &auty.ActiveBoard{
			Boards:    []string{"11", "12"},
			Revboards: []string{"13", "15"},
			Amount:    100,
		},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status:     2,
		Address:    "123",
	}
	pre := copyAutonomyProposalChange(cur)
	cur.PropChange.Year = 1800
	cur.PropChange.Month = 2
	cur.PropChange.Changes[0].Cancel = false
	cur.PropChange.Changes[0].Addr = "21"
	cur.PropChange.Changes[1].Cancel = true
	cur.PropChange.Changes[1].Addr = "22"
	cur.CurRule.BoardApproveRatio = 90
	cur.Board.Boards[0] = "21"
	cur.Board.Boards[1] = "22"
	cur.Board.Revboards[0] = "23"
	cur.Board.Revboards[0] = "25"
	cur.Board.Amount = 90
	cur.VoteResult.TotalVotes = 50
	cur.Address = "234"
	cur.Status = 1

	assert.Equal(t, 1900, int(pre.PropChange.Year))
	assert.Equal(t, 1, int(pre.PropChange.Month))
	assert.Equal(t, &auty.Change{Cancel: true, Addr: "11"}, pre.PropChange.Changes[0])
	assert.Equal(t, &auty.Change{Cancel: false, Addr: "12"}, pre.PropChange.Changes[1])

	assert.Equal(t, "11", pre.Board.Boards[0])
	assert.Equal(t, "12", pre.Board.Boards[1])
	assert.Equal(t, "13", pre.Board.Revboards[0])
	assert.Equal(t, "15", pre.Board.Revboards[1])
	assert.Equal(t, 100, int(pre.Board.Amount))

	assert.Equal(t, 100, int(pre.VoteResult.TotalVotes))
	assert.Equal(t, "123", pre.Address)
	assert.Equal(t, 2, int(pre.Status))
}
