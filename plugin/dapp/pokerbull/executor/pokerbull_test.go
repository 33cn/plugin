package executor

import (
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pkt "github.com/33cn/plugin/plugin/dapp/pokerbull/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

var (
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	Nodes    = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
	}
)

func TestPokerbull(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	Init(pkt.PokerBullX, cfg, nil)
	total := 1000 * types.DefaultCoinPrecision
	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[0]),
	}
	accountB := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[1]),
	}

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)

	execAddr := dapp.ExecAddress(pkt.PokerBullX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 100)
	_, _, kvdb := util.CreateTestDB()

	accA := account.NewCoinsAccount(cfg)
	accA.SetDB(stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB := account.NewCoinsAccount(cfg)
	accB.SetDB(stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	env := execEnv{
		10,
		cfg.GetDappFork(pkt.PokerBullX, "Enable"),
		1539918074,
	}

	// start game
	p1 := &pkt.PBGameStart{
		Value:     5 * types.DefaultCoinPrecision,
		PlayerNum: 2,
	}
	createTx, err := types.CallCreateTransaction(pkt.PokerBullX, "Start", p1)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = pkt.ExecerPokerBull
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}
	exec := newPBGame()
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	assert.Equal(t, exec.GetCoinsAccount().LoadExecAccount(string(Nodes[0]), execAddr).GetBalance(), total)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	gameID := createTx.Hash()

	// start game p2
	createTx, err = types.CallCreateTransaction(pkt.PokerBullX, "Start", p1)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = pkt.ExecerPokerBull
	createTx, err = signTx(createTx, PrivKeyB)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	// continue game
	p2 := &pkt.PBGameContinue{
		GameId: common.ToHex(gameID),
	}
	createTx, err = types.CallCreateTransaction(pkt.PokerBullX, "Continue", p2)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = pkt.ExecerPokerBull
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	// quit game
	p3 := &pkt.PBGameQuit{
		GameId: common.ToHex(gameID),
	}
	createTx, err = types.CallCreateTransaction(pkt.PokerBullX, "Quit", p3)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = pkt.ExecerPokerBull
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight+2, env.blockTime+2, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	// query
	res, err := exec.Query(pkt.FuncNameQueryGameByID, types.Encode(&pkt.QueryPBGameInfo{GameId: common.ToHex(gameID)}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	res, err = exec.Query(pkt.FuncNameQueryGameByAddr, types.Encode(&pkt.QueryPBGameInfo{Addr: string(Nodes[0])}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	res, err = exec.Query(pkt.FuncNameQueryGameByStatus, types.Encode(&pkt.QueryPBGameInfo{Status: pkt.PBGameActionQuit}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	res, err = exec.Query(pkt.FuncNameQueryGameByRound, types.Encode(&pkt.QueryPBGameByRound{GameId: common.ToHex(gameID), Round: int32(1)}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	var gameIDsS []string
	gameIDsS = append(gameIDsS, common.ToHex(gameID))
	res, err = exec.Query(pkt.FuncNameQueryGameListByIDs, types.Encode(&pkt.QueryPBGameInfos{GameIds: gameIDsS}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName(pkt.PokerBullX, signType), -1)
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
