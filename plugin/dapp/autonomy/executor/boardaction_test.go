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
	dbmock "github.com/33cn/chain33/common/db/mocks"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common"
	commonlog "github.com/33cn/chain33/common/log"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/account"
	_ "github.com/33cn/chain33/system"
	"github.com/stretchr/testify/mock"
	"github.com/33cn/chain33/common/address"
)

type execEnv struct {
	blockTime   int64 // 1539918074
	blockHeight int64
	index       int
	difficulty  uint64
	txHash      string
}

var (
	Symbol = "BTY"
	Asset  = "coins"

	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	AddrA    = "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
	AddrB    = "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
	AddrC    = "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
	AddrD    = "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"

	AddrBWeight           uint64 = 1
	AddrCWeight           uint64 = 4
	AddrDWeight           uint64 = 10
	NewWeight             uint64 = 2
	Requiredweight        uint64 = 5
	NewRequiredweight     uint64 = 4
	CoinsBtyDailylimit    uint64 = 100
	NewCoinsBtyDailylimit uint64 = 10
	PrintFlag                    = false
	InAmount              int64  = 10
	OutAmount             int64  = 5

	boards = []string{"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4", "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR", "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"}
)

func init() {
	commonlog.SetLogLevel("error")
	//types.AllowUserExec = append(types.AllowUserExec, []byte("coins"))
}

func TestProposalBoard(t *testing.T) {
	total := types.Coin * 30000
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

	env := execEnv{
		1539918074,
		10,
		2,
		1539918074,
		"hash",
	}

	stateDB, _ := dbm.NewGoMemDB("state", "state", 100)
	localDB := new(dbmock.KVDB)
	api := new(apimock.QueueProtocolAPI)

	accCoin := account.NewCoinsAccount()
	accCoin.SetDB(stateDB)
	accCoin.SaveAccount(&accountA)
	accCoin.SaveExecAccount(address.ExecAddress(auty.AutonomyX), &accountA)
	accCoin.SaveAccount(&accountB)
	accCoin.SaveAccount(&accountC)
	accCoin.SaveAccount(&accountD)

	driver := newAutonomy()
	driver.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	driver.SetStateDB(stateDB)
	driver.SetLocalDB(localDB)
	driver.SetAPI(api)

	// PropBoardTx
	opt1 :=  &auty.ProposalBoard{
		Year: 2019,
		Month: 7,
		Day:     10,
		Boards:    boards,
		StartBlockHeight:  env.blockHeight + 5,
		EndBlockHeight: env.blockHeight + 10,
	}
	pbtx, err := propBoardTx(opt1)
	require.NoError(t, err)
	pbtx, err = signTx(pbtx, PrivKeyA)
	require.NoError(t, err)
	exec := newAutonomy()
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(localDB)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(pbtx, int(1))
	require.NoError(t, err)
	require.NotNil(t, receipt)

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
	api.On("GetLastHeader", mock.Anything).Return(&types.Header{StateHash: []byte("111111111111111111111")}, nil)
	api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values: make([][]byte, 1)}, nil)
	hear := &types.Header{StateHash: []byte("111111111111111111111")}
	api.On("GetHeaders", mock.Anything).
		Return(&types.Headers{
		Items:[]*types.Header{hear}}, nil)
	_, err := action.getStartHeightVoteAccount(addr, 0)
	require.NoError(t, err)
}

func TestGetReceiptLog(t *testing.T) {
	pre := &auty.AutonomyProposalBoard{
		PropBoard: &auty.ProposalBoard{Year: 1800, Month: 1},
		Res: &auty.VotesResult{TotalVotes: 100},
		Status: 1,
		Address:"121",
	}
	cur := &auty.AutonomyProposalBoard{
		PropBoard: &auty.ProposalBoard{Year: 1900, Month: 1},
		Res: &auty.VotesResult{TotalVotes: 100},
		Status: 2,
		Address:"123",
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
	cur := &auty.AutonomyProposalBoard{
		PropBoard: &auty.ProposalBoard{Year: 1900, Month: 1},
		Res: &auty.VotesResult{TotalVotes: 100},
		Status: 2,
		Address:"123",
	}
	new := copyAutonomyProposalBoard(cur)
	cur.PropBoard.Year = 1800
	cur.PropBoard.Month = 2
	cur.Res.TotalVotes = 50
	cur.Address = "234"
	cur.Status = 1

	require.Equal(t, 1900, int(new.PropBoard.Year))
	require.Equal(t, 1, int(new.PropBoard.Month))
	require.Equal(t, 100, int(new.Res.TotalVotes))
	require.Equal(t, "123", new.Address)
	require.Equal(t, 2, int(new.Status))
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