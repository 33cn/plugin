package executor

import (
	"testing"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/queue"
	pty "github.com/33cn/plugin/plugin/dapp/game/types"
	"github.com/stretchr/testify/assert"
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

var (
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	Nodes    = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}
)

func TestGame(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	Init(pty.GameX, cfg, nil)
	total := 100 * types.DefaultCoinPrecision
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

	accountC := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[2]),
	}
	accountD := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[3]),
	}
	execAddr := address.ExecAddress(pty.GameX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 1000)
	_, _, kvdb := util.CreateTestDB()

	accA, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	accC, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accC.SaveExecAccount(execAddr, &accountC)

	accD, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accD.SaveExecAccount(execAddr, &accountD)
	env := execEnv{
		10,
		cfg.GetDappFork(pty.GameX, "Enable"),
		1539918074,
	}

	// create game
	createParam := &pty.GamePreCreateTx{Amount: 2 * types.DefaultCoinPrecision,
		HashType:  "sha256",
		HashValue: common.Sha256([]byte("harrylee" + string(Rock))),
		Fee:       100000}
	createTx, err := pty.CreateRawGamePreCreateTx(cfg, createParam)
	if err != nil {
		t.Error(err)
	}
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error(err)
	}
	exec := newGame()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(1, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(createTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(createTx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	gameID := common.ToHex(createTx.Hash())

	//match game
	matchParam := &pty.GamePreMatchTx{GameID: gameID, Guess: Scissor, Fee: 100000}
	matchTx, err := pty.CreateRawGamePreMatchTx(cfg, matchParam)
	if err != nil {
		t.Error(err)
	}
	matchTx, err = signTx(matchTx, PrivKeyB)
	if err != nil {
		t.Error(err)
	}
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(2, env.blockTime+20, env.difficulty)
	receipt, err = exec.Exec(matchTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(matchTx, receiptDate, int(1))
	assert.Nil(t, err)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	msg, err := exec.Query(pty.FuncNameQueryGameListByIds, types.Encode(&pty.QueryGameInfos{
		GameIds: []string{gameID},
	}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	msg, err = exec.Query(pty.FuncNameQueryGameByID, types.Encode(&pty.QueryGameInfo{
		GameId: gameID}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	_, err = exec.Query(pty.FuncNameQueryGameListByStatusAndAddr, types.Encode(&pty.QueryGameListByStatusAndAddr{
		Status: pty.GameActionMatch}))
	if err != nil {
		t.Error(err)
	}

	//close game
	closeParam := &pty.GamePreCloseTx{GameID: gameID, Secret: "harrylee", Result: Rock, Fee: 100000}
	closeTx, err := pty.CreateRawGamePreCloseTx(cfg, closeParam)
	if err != nil {
		t.Error(err)
	}
	closeTx, err = signTx(closeTx, PrivKeyA)
	if err != nil {
		t.Error(err)
	}
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(2, env.blockTime+20, env.difficulty)
	receipt, err = exec.Exec(closeTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(closeTx, receiptDate, int(1))
	assert.Nil(t, err)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	acA := accA.LoadExecAccount(string(Nodes[0]), execAddr)
	acB := accB.LoadExecAccount(string(Nodes[1]), execAddr)
	t.Log(acA)
	t.Log(acB)

	msg, err = exec.Query(pty.FuncNameQueryGameByID, types.Encode(&pty.QueryGameInfo{
		GameId: gameID}))
	if err != nil {
		t.Error(err)
	}
	reply := msg.(*pty.ReplyGame)
	assert.Equal(t, int32(pty.GameActionClose), reply.Game.Status)
	assert.Equal(t, IsCreatorWin, reply.Game.Result)

	// create game
	createParam = &pty.GamePreCreateTx{Amount: 2 * types.DefaultCoinPrecision,
		HashType:  "sha256",
		HashValue: common.Sha256([]byte("123456" + string(Rock))),
		Fee:       100000}
	createTx, err = pty.CreateRawGamePreCreateTx(cfg, createParam)
	if err != nil {
		t.Error(err)
	}
	createTx, err = signTx(createTx, PrivKeyC)
	if err != nil {
		t.Error(err)
	}
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(1, env.blockTime, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	gameID = common.ToHex(createTx.Hash())

	//cancle game
	cancleParam := &pty.GamePreCancelTx{Fee: 1e5, GameID: gameID}
	cancelTx, err := pty.CreateRawGamePreCancelTx(cfg, cancleParam)
	if err != nil {
		t.Error(err)
	}
	createTx, err = signTx(cancelTx, PrivKeyC)
	if err != nil {
		t.Error(err)
	}
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(1, env.blockTime, env.difficulty)
	receipt, err = exec.Exec(cancelTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(cancelTx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	msg, err = exec.Query(pty.FuncNameQueryGameByID, types.Encode(&pty.QueryGameInfo{
		GameId: gameID}))
	if err != nil {
		t.Error(err)
	}
	reply = msg.(*pty.ReplyGame)
	assert.Equal(t, int32(pty.GameActionCancel), reply.Game.Status)

	//create game
	createParam = &pty.GamePreCreateTx{Amount: 2 * types.DefaultCoinPrecision,
		HashType:  "sha256",
		HashValue: common.Sha256([]byte("123456" + string(Rock))),
		Fee:       100000}
	createTx, err = pty.CreateRawGamePreCreateTx(cfg, createParam)
	if err != nil {
		t.Error(err)
	}
	createTx, err = signTx(createTx, PrivKeyC)
	if err != nil {
		t.Error(err)
	}
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(1, env.blockTime, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	gameID = common.ToHex(createTx.Hash())

	//match game
	matchParam = &pty.GamePreMatchTx{GameID: gameID, Guess: Rock, Fee: 100000}
	matchTx, err = pty.CreateRawGamePreMatchTx(cfg, matchParam)
	if err != nil {
		t.Error(err)
	}
	matchTx, err = signTx(matchTx, PrivKeyB)
	if err != nil {
		t.Error(err)
	}
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(2, env.blockTime+20, env.difficulty)
	receipt, err = exec.Exec(matchTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(matchTx, receiptDate, int(1))
	assert.Nil(t, err)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	//close game
	closeParam = &pty.GamePreCloseTx{GameID: gameID, Secret: "123456", Result: Rock, Fee: 100000}
	closeTx, err = pty.CreateRawGamePreCloseTx(cfg, closeParam)
	if err != nil {
		t.Error(err)
	}
	closeTx, err = signTx(closeTx, PrivKeyC)
	if err != nil {
		t.Error(err)
	}
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(2, env.blockTime+20, env.difficulty)
	receipt, err = exec.Exec(closeTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(closeTx, receiptDate, int(1))
	assert.Nil(t, err)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	msg, err = exec.Query(pty.FuncNameQueryGameByID, types.Encode(&pty.QueryGameInfo{
		GameId: gameID}))
	if err != nil {
		t.Error(err)
	}
	reply = msg.(*pty.ReplyGame)
	assert.Equal(t, int32(pty.GameActionClose), reply.Game.Status)
	assert.Equal(t, IsDraw, reply.Game.Result)
	//create game
	createParam = &pty.GamePreCreateTx{Amount: 2 * types.DefaultCoinPrecision,
		HashType:  "sha256",
		HashValue: common.Sha256([]byte("123456" + string(Rock))),
		Fee:       100000}
	createTx, err = pty.CreateRawGamePreCreateTx(cfg, createParam)
	if err != nil {
		t.Error(err)
	}
	createTx, err = signTx(createTx, PrivKeyC)
	if err != nil {
		t.Error(err)
	}
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(1, env.blockTime, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	gameID = common.ToHex(createTx.Hash())

	//match game
	matchParam = &pty.GamePreMatchTx{GameID: gameID, Guess: Paper, Fee: 100000}
	matchTx, err = pty.CreateRawGamePreMatchTx(cfg, matchParam)
	if err != nil {
		t.Error(err)
	}
	matchTx, err = signTx(matchTx, PrivKeyB)
	if err != nil {
		t.Error(err)
	}
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(2, env.blockTime+20, env.difficulty)
	receipt, err = exec.Exec(matchTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(matchTx, receiptDate, int(1))
	assert.Nil(t, err)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	//close game
	closeParam = &pty.GamePreCloseTx{GameID: gameID, Secret: "123456", Result: Rock, Fee: 100000}
	closeTx, err = pty.CreateRawGamePreCloseTx(cfg, closeParam)
	if err != nil {
		t.Error(err)
	}
	closeTx, err = signTx(closeTx, PrivKeyC)
	if err != nil {
		t.Error(err)
	}
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(2, env.blockTime+20, env.difficulty)
	receipt, err = exec.Exec(closeTx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(closeTx, receiptDate, int(1))
	assert.Nil(t, err)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	msg, err = exec.Query(pty.FuncNameQueryGameByID, types.Encode(&pty.QueryGameInfo{
		GameId: gameID}))
	if err != nil {
		t.Error(err)
	}
	reply = msg.(*pty.ReplyGame)
	assert.Equal(t, int32(pty.GameActionClose), reply.Game.Status)
	assert.Equal(t, IsMatcherWin, reply.Game.Result)
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName("", signType), -1)
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
