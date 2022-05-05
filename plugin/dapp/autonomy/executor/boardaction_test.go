// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	dbmock "github.com/33cn/chain33/common/db/mocks"
	commonlog "github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/queue"
	_ "github.com/33cn/chain33/system"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	ticket "github.com/33cn/plugin/plugin/dapp/ticket/executor"
	ticketTy "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	PrivKey1  = "0x9d4f8ab11361be596468b265cb66946c87873d4a119713fd0c3d8302eae0a8e4"
	PrivKey2  = "0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5"
	PrivKey3  = "0xc21d38be90493512a5c2417d565269a8b23ce8152010e404ff4f75efead8183a"
	PrivKey4  = "0xfdf2bbff853ecff2e7b86b2a8b45726c6538ca7d1403dc94e50131ef379bdca0"
	PrivKey5  = "0x794443611e7369a57b078881445b93b754cbc9b9b8f526535ab9c6d21d29203d"
	PrivKey6  = "0xf2cc48d30560e4c92e84821df68cf1086de82ee6a5725fc2a590a58d6ffe4fc5"
	PrivKey7  = "0xeb4738a7c685a7ccf5471c3335a2d7ebe284b11d8a1717d814904b8d1ba936d9"
	PrivKey8  = "0x9d315182e56fde7fadb94408d360203894e5134216944e858f9b31f70e9ecf40"
	PrivKey9  = "0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1"
	PrivKey10 = "0x1c3e6cac2f887e1ab9180e2d5772dc4ba01accb8d4df434faba097003eb35482"

	Addr1  = "12HKLEn6g4FH39yUbHh4EVJWcFo5CXg22d"
	Addr2  = "1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj"
	Addr3  = "12cjnN5D4DPdBQSwh6vjwJbtsW4EJALTMv"
	Addr4  = "1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH"
	Addr5  = "1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB"
	Addr6  = "1L1puAUjfmtDECKo2C1qLWsAMZtDGTBWf6"
	Addr7  = "1LNf9AVXzUMQkQM5hgGLhkdrVtD8UMBSUm"
	Addr8  = "1PcGKYYoLn1PLLJJodc1UpgWGeFAQasAkx"
	Addr9  = "1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu"
	Addr10 = "1Q9sQwothzM1gKSzkVZ8Dt1tqKX1uzSagx"

	PrivKey11 = "0xfd0c4a8a1efcd221ee0f36b7d4f57d8ff843cb8bc193b39c7863332d355acafa"
	PrivKey12 = "0x4c9691bf6acc908ef5c07abcad23cf7f98e46e84101aa5059322aa53eb4dc471"
	PrivKey13 = "0x50b9c6a4358ef8ffc96d5831a8dfd5e0fae504d49e20c5eafd12b6015423de33"
	PrivKey14 = "0x96e3c766850a915fe4718b890d96208d5d1a3694b2597e08165480b5b48b84cb"
	PrivKey15 = "0xeac5e45243c3920cf8a98f3d3a2e3a9b43f30a21769b57f734213913511e7575"
	PrivKey16 = "0xd2aaa6f050a4db13fbd2c8bf87cbb96e217289172baca6c12e8a8b0680b9aa1a"
	PrivKey17 = "0x33b3b977c657435a49773b5605a704ad5fdca0fa34fe36a02ea0f13a49099832"

	Addr11 = "15VUiygdxMSZ3rykwe742yomp2cPJ9Tfve"
	Addr12 = "1DyR84CU5AHbGXLEnhHMwMvWNMeunLZsuJ"
	Addr13 = "132pBvrgSYgHASxzoeL3bqnsqUpaBbUktm"
	Addr14 = "1DEV4XwdBUWRkMuy4ARRiEAoxQ2LoDByNG"
	Addr15 = "18Y87cw2hiYC71bvpD872oYMYXtw66Qp6o"
	Addr16 = "1Fghq6cgdJEDr6gQBmvba3t6aXAkyZyjr2"
	Addr17 = "142KsfJLvEA5FJxAgKm9ZDtFVjkRaPdu82"

	boards = []string{
		AddrA,
		AddrB,
		AddrC,
		AddrD,

		Addr1,
		Addr2,
		Addr3,
		Addr4,
		Addr5,
		Addr6,
		Addr7,
		Addr8,
		Addr9,
		Addr10,

		Addr11,
		Addr12,
		Addr13,
		Addr14,
		Addr15,
		Addr16,
		Addr17,
	}
	total = types.DefaultCoinPrecision * 30000
)

func init() {
	commonlog.SetLogLevel("error")
	Init(auty.AutonomyX, chainTestCfg, nil)
}

// InitEnv 初始化环境
func InitEnv() (*ExecEnv, drivers.Driver, dbm.KV, dbm.KVDB) {
	//cfg := types.NewChain33Config(types.GetDefaultCfgstring())
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

	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	accCoin.SaveAccount(&accountA)
	accCoin.SaveExecAccount(autonomyAddr, &accountA)
	accCoin.SaveAccount(&accountB)
	accCoin.SaveAccount(&accountC)
	accCoin.SaveAccount(&accountD)
	//total ticket balance
	accCoin.SaveAccount(&types.Account{Balance: total * 4,
		Frozen: 0,
		Addr:   "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"})

	exec := newAutonomy()
	q := queue.New("channel")
	q.SetConfig(chainTestCfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	return env, exec, stateDB, kvdb
}

func InitMinerAddr(stateDB dbm.KV, addrs []string, bind string) {
	for _, addr := range addrs {
		tkBind := &ticketTy.TicketBind{
			MinerAddress:  bind,
			ReturnAddress: addr,
		}
		stateDB.Set(bindKey(addr), types.Encode(tkBind))
	}
}

func TestPropBoard(t *testing.T) {
	env, exec, stateDB, _ := InitEnv()

	opts := []*auty.ProposalBoard{
		{ // ErrRepeatAddr
			BoardUpdate:      auty.BoardUpdate_ADDBoard,
			Boards:           []string{"18e1nfiux7aVSfN2zYUZhbidMRokbBSPA6", "18e1nfiux7aVSfN2zYUZhbidMRokbBSPA6"},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // ErrRepeatAddr
			BoardUpdate:      auty.BoardUpdate_ADDBoard,
			Boards:           []string{"18e1nfiux7aVSfN2zYUZhbidMRokbBSPA6", AddrA},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // ErrBoardNumber
			BoardUpdate:      auty.BoardUpdate_ADDBoard,
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // 正常
			BoardUpdate:      auty.BoardUpdate_ADDBoard,
			Boards:           []string{"18e1nfiux7aVSfN2zYUZhbidMRokbBSPA6"},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},

		{ // ErrRepeatAddr
			BoardUpdate:      auty.BoardUpdate_REPLACEALL,
			Boards:           []string{"18e1nfiux7aVSfN2zYUZhbidMRokbBSPA6", "18e1nfiux7aVSfN2zYUZhbidMRokbBSPA6"},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // ErrBoardNumber
			BoardUpdate:      auty.BoardUpdate_REPLACEALL,
			Boards:           []string{"18e1nfiux7aVSfN2zYUZhbidMRokbBSPA6", AddrA},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // 正常
			BoardUpdate:      auty.BoardUpdate_REPLACEALL,
			Boards:           boards,
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},

		{ // ErrNotFound
			BoardUpdate:      auty.BoardUpdate_DELBoard,
			Boards:           []string{"18e1nfiux7aVSfN2zYUZhbidMRokbBSPA6"},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // ErrNotFound
			BoardUpdate:      auty.BoardUpdate_DELBoard,
			Boards:           []string{Addr17, "18e1nfiux7aVSfN2zYUZhbidMRokbBSPA6"},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // ErrBoardNumber
			BoardUpdate:      auty.BoardUpdate_DELBoard,
			Boards:           []string{Addr16, Addr17},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // ErrRepeatAddr
			BoardUpdate:      auty.BoardUpdate_DELBoard,
			Boards:           []string{Addr17, Addr17},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},
		{ // 正常
			BoardUpdate:      auty.BoardUpdate_DELBoard,
			Boards:           []string{Addr17},
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
		},

		{ // ErrSetBlockHeight
			BoardUpdate:      auty.BoardUpdate_REPLACEALL,
			Boards:           boards,
			StartBlockHeight: env.blockHeight + 5,
			EndBlockHeight:   env.blockHeight + autoCfg.PropEndBlockPeriod + 10,
		},
	}
	result := []error{
		auty.ErrRepeatAddr,
		auty.ErrRepeatAddr,
		auty.ErrBoardNumber,
		nil,

		auty.ErrRepeatAddr,
		auty.ErrBoardNumber,
		nil,

		types.ErrNotFound,
		types.ErrNotFound,
		auty.ErrBoardNumber,
		auty.ErrRepeatAddr,
		nil,
		auty.ErrSetBlockHeight,
	}
	lenBoards := []int{0, 0, 0, 22, 0, 0, 21, 0, 0, 0, 0, 20, 0}

	InitBoard(stateDB)
	exec.SetStateDB(stateDB)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	for i, tcase := range opts {
		pbtx, err := propBoardTx(tcase)
		assert.NoError(t, err)
		pbtx, err = signTx(pbtx, PrivKeyA)
		assert.NoError(t, err)
		receipt, err := exec.Exec(pbtx, i)
		assert.Equal(t, errors.Cause(err), result[i])
		if receipt != nil {
			var stat auty.AutonomyProposalBoard
			err := types.Decode(receipt.KV[1].Value, &stat)
			assert.NoError(t, err)
			assert.Equal(t, len(stat.Board.Boards), lenBoards[i])
		}
	}
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
		BoardUpdate:      auty.BoardUpdate_REPLACEALL,
		StartBlockHeight: env.blockHeight + 5,
		EndBlockHeight:   env.blockHeight + autoCfg.StartEndBlockPeriod + 10,
	}
	pbtx, err := propBoardTx(opt1)
	assert.NoError(t, err)
	pbtx, err = signTx(pbtx, PrivKeyA)
	assert.NoError(t, err)

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(pbtx, int(1))
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	if save {
		for _, kv := range receipt.KV {
			stateDB.Set(kv.Key, kv.Value)
		}
	}

	// local
	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(pbtx, receiptData, int(1))
	assert.NoError(t, err)
	assert.NotNil(t, set)
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
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, autoCfg.ProposalAmount*types.DefaultCoinPrecision, account.Frozen)
}

func propBoardTx(parm *auty.ProposalBoard) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPropBoard,
		Value: &auty.AutonomyAction_PropBoard{PropBoard: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func revokeProposalBoard(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	proposalID := env.txHash
	opt2 := &auty.RevokeProposalBoard{
		ProposalID: proposalID,
	}
	rtx, err := revokeProposalBoardTx(opt2)
	assert.NoError(t, err)
	rtx, err = signTx(rtx, PrivKeyA)
	assert.NoError(t, err)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(rtx, int(1))
	assert.NoError(t, err)
	assert.NotNil(t, receipt)
	if save {
		for _, kv := range receipt.KV {
			stateDB.Set(kv.Key, kv.Value)
		}
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(rtx, receiptData, int(1))
	assert.NoError(t, err)
	assert.NotNil(t, set)
	if save {
		for _, kv := range set.KV {
			kvdb.Set(kv.Key, kv.Value)
		}
	}
	// del
	set, err = exec.ExecDelLocal(rtx, receiptData, int(1))
	assert.NoError(t, err)
	assert.NotNil(t, set)
	// check
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, int64(0), account.Frozen)
}

func revokeProposalBoardTx(parm *auty.RevokeProposalBoard) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionRvkPropBoard,
		Value: &auty.AutonomyAction_RvkPropBoard{RvkPropBoard: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func voteProposalBoard(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	api := new(apimock.QueueProtocolAPI)
	api.On("StoreList", mock.Anything).Return(&types.StoreListReply{}, nil)
	api.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)
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
		vote   auty.AutonomyVoteOption
		origin []string
	}
	records := []record{
		{priv: PrivKeyA, vote: auty.AutonomyVoteOption_OPPOSE},
		{priv: PrivKey1, vote: auty.AutonomyVoteOption_APPROVE, origin: []string{AddrB, AddrC, AddrD}},
	}
	InitMinerAddr(stateDB, []string{AddrB, AddrC, AddrD}, Addr1)

	for i, record := range records {
		opt := &auty.VoteProposalBoard{
			ProposalID: proposalID,
			VoteOption: record.vote,
			OriginAddr: record.origin,
		}
		tx, err := voteProposalBoardTx(opt)
		assert.NoError(t, err)
		tx, err = signTx(tx, record.priv)
		assert.NoError(t, err)
		// 设定当前高度为投票高度
		exec.SetEnv(env.startHeight, env.blockTime, env.difficulty)

		receipt, err := exec.Exec(tx, int(1))
		assert.NoError(t, err)
		assert.NotNil(t, receipt)
		if save {
			for _, kv := range receipt.KV {
				stateDB.Set(kv.Key, kv.Value)
			}
		}
		receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
		set, err := exec.ExecLocal(tx, receiptData, int(1))
		assert.NoError(t, err)
		assert.NotNil(t, set)
		if save {
			for _, kv := range set.KV {
				kvdb.Set(kv.Key, kv.Value)
			}
		}
		// del
		set, err = exec.ExecDelLocal(tx, receiptData, int(1))
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
	account := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, int64(0), account.Frozen)
	account = accCoin.LoadExecAccount(autonomyAddr, autonomyAddr)
	assert.Equal(t, autoCfg.ProposalAmount*types.DefaultCoinPrecision, account.Balance)
	// status
	value, err := stateDB.Get(propBoardID(proposalID))
	assert.NoError(t, err)
	cur := &auty.AutonomyProposalBoard{}
	err = types.Decode(value, cur)
	assert.NoError(t, err)
	assert.Equal(t, int32(auty.AutonomyStatusTmintPropBoard), cur.Status)
	assert.Equal(t, AddrA, cur.Address)
	assert.Equal(t, true, cur.VoteResult.Pass)
}

func voteProposalBoardTx(parm *auty.VoteProposalBoard) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionVotePropBoard,
		Value: &auty.AutonomyAction_VotePropBoard{VotePropBoard: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func terminateProposalBoard(t *testing.T, env *ExecEnv, exec drivers.Driver, stateDB dbm.KV, kvdb dbm.KVDB, save bool) {
	api := new(apimock.QueueProtocolAPI)
	api.On("StoreList", mock.Anything).Return(&types.StoreListReply{}, nil)
	api.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)
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
	assert.NoError(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.NoError(t, err)
	exec.SetEnv(env.endHeight+1, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(tx, int(1))
	assert.NoError(t, err)
	assert.NotNil(t, receipt)
	if save {
		for _, kv := range receipt.KV {
			stateDB.Set(kv.Key, kv.Value)
		}
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(tx, receiptData, int(1))
	assert.NoError(t, err)
	assert.NotNil(t, set)
	if save {
		for _, kv := range set.KV {
			kvdb.Set(kv.Key, kv.Value)
		}
	}
	// del
	set, err = exec.ExecDelLocal(tx, receiptData, int(1))
	assert.NoError(t, err)
	assert.NotNil(t, set)
	// check
	accCoin := account.NewCoinsAccount(chainTestCfg)
	accCoin.SetDB(stateDB)
	account := accCoin.LoadExecAccount(AddrA, autonomyAddr)
	assert.Equal(t, int64(0), account.Frozen)
	account = accCoin.LoadExecAccount(autonomyAddr, autonomyAddr)
	assert.Equal(t, int64(0), account.Frozen)
}

func terminateProposalBoardTx(parm *auty.TerminateProposalBoard) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTmintPropBoard,
		Value: &auty.AutonomyAction_TmintPropBoard{TmintPropBoard: parm},
	}
	return types.CreateFormatTx(chainTestCfg, chainTestCfg.ExecName(auty.AutonomyX), types.Encode(val))
}

func TestGetStartHeightVoteAccount(t *testing.T) {
	at := newAutonomy().(*Autonomy)
	at.SetLocalDB(new(dbmock.KVDB))

	api := new(apimock.QueueProtocolAPI)
	addr := "1JmFaA6unrCFYEWPGRi7uuXY1KthTJxJEP"
	api.On("StoreList", mock.Anything).Return(&types.StoreListReply{}, nil)
	api.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)
	api.On("GetLastHeader", mock.Anything).Return(&types.Header{StateHash: []byte("")}, nil)

	at.SetAPI(api)
	tx := &types.Transaction{}
	action := newAction(at, tx, 0)

	acc := &types.Account{
		Currency: 0,
		Balance:  types.DefaultCoinPrecision,
	}
	val := types.Encode(acc)
	values := [][]byte{val}
	api.On("StoreGet", mock.Anything).Return(&types.StoreReplyValue{Values: values}, nil)
	hear := &types.Header{StateHash: []byte("")}
	api.On("GetHeaders", mock.Anything).
		Return(&types.Headers{
			Items: []*types.Header{hear}}, nil)
	account, err := action.getStartHeightVoteAccount(addr, "", 0)
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, types.DefaultCoinPrecision, account.Balance)
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
	assert.Equal(t, int32(2), log.Ty)
	recpt := &auty.ReceiptProposalBoard{}
	err := types.Decode(log.Log, recpt)
	assert.NoError(t, err)
	assert.Equal(t, int32(1800), recpt.Prev.PropBoard.Year)
	assert.Equal(t, int32(1900), recpt.Current.PropBoard.Year)
}

func TestCopyAutonomyProposalBoard(t *testing.T) {
	assert.Nil(t, copyAutonomyProposalBoard(nil))
	cur := &auty.AutonomyProposalBoard{
		PropBoard:  &auty.ProposalBoard{Year: 1900, Month: 1},
		Board:      &auty.ActiveBoard{Boards: []string{"111", "112"}, Revboards: []string{"113", "114"}},
		CurRule:    &auty.RuleConfig{BoardApproveRatio: 100},
		VoteResult: &auty.VoteResult{TotalVotes: 100},
		Status:     2,
		Address:    "123",
	}
	pre := copyAutonomyProposalBoard(cur)
	cur.PropBoard.Year = 1800
	cur.PropBoard.Month = 2
	cur.Board.Boards = []string{"211", "212"}
	cur.Board.Revboards = []string{"113", "114"}
	cur.CurRule.BoardApproveRatio = 90
	cur.VoteResult.TotalVotes = 50
	cur.Address = "234"
	cur.Status = 1

	assert.Equal(t, 1900, int(pre.PropBoard.Year))
	assert.Equal(t, []string{"111", "112"}, pre.Board.Boards)
	assert.Equal(t, []string{"113", "114"}, pre.Board.Revboards)
	assert.Equal(t, 100, int(pre.CurRule.BoardApproveRatio))
	assert.Equal(t, 100, int(pre.VoteResult.TotalVotes))
	assert.Equal(t, "123", pre.Address)
	assert.Equal(t, 2, int(pre.Status))
}

func TestVerifyMinerAddr(t *testing.T) {
	at := newTestAutonomy()
	stateDB, _ := dbm.NewGoMemDB("state", "state", 100)
	at.SetStateDB(stateDB)
	tx := &types.Transaction{}
	action := newAction(at, tx, 0)
	addrs := []string{
		AddrA,
		AddrB,
		AddrC,
	}
	// 授权地址AddrD
	for _, addr := range addrs {
		tkBind := &ticketTy.TicketBind{
			MinerAddress:  AddrD,
			ReturnAddress: addr,
		}
		stateDB.Set(bindKey(addr), types.Encode(tkBind))
	}
	_, err := action.verifyMinerAddr(addrs, AddrD)
	assert.NoError(t, err)
	// ErrRepeatAddr
	addrss := []string{
		AddrA,
		AddrB,
		AddrC,
		AddrA,
	}
	add, err := action.verifyMinerAddr(addrss, AddrD)
	assert.Equal(t, auty.ErrRepeatAddr, err)
	assert.Equal(t, add, AddrA)

	// ErrMinerAddr
	testf := "12HKLEn6g4FH39yUbHh4EVJWcFo5CXg22d"
	addrs = []string{testf}
	addr, err := action.verifyMinerAddr(addrs, AddrD)
	assert.Equal(t, auty.ErrMinerAddr, err)
	assert.Equal(t, testf, addr)

	// ErrBindAddr
	testf = "1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj"
	tkBind := &ticketTy.TicketBind{
		MinerAddress:  AddrA,
		ReturnAddress: testf,
	}
	stateDB.Set(bindKey(testf), types.Encode(tkBind))
	addrs = []string{testf}
	addr, err = action.verifyMinerAddr(addrs, AddrD)
	assert.Equal(t, auty.ErrBindAddr, err)
	assert.Equal(t, testf, addr)
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName(auty.AutonomyX, signType), -1)
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

func TestBindKey(t *testing.T) {
	ids := []string{
		"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",
		"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k",
		"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs",
		"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR",
	}
	for _, id := range ids {
		subcfg.BindKey = ""
		assert.Equal(t, bindKey(id), ticket.BindKey(id))

		subcfg.BindKey = "mavl-ticket-tbind-"
		assert.NotNil(t, bindKey(id))
		assert.Equal(t, bindKey(id), ticket.BindKey(id))

		subcfg.BindKey = "mavl-pos33-bind-"
		assert.NotNil(t, bindKey(id))
		assert.False(t, string(bindKey(id)) == string(ticket.BindKey(id)))
	}
}
