package executor

import (
	"testing"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	pty "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/stretchr/testify/assert"
	//"github.com/33cn/chain33/types/jsonpb"
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

var (
	Symbol         = "TEST"
	AssetExecToken = "token"
	AssetExecPara  = "paracross"

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

func TestToken(t *testing.T) {
	types.SetTitleOnlyForTest("chain33")
	tokenTotal := int64(10000 * 1e8)
	tokenBurn := int64(10 * 1e8)
	tokenMint := int64(20 * 1e8)
	total := int64(100000)
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

	execAddr := address.ExecAddress(pty.TokenX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 100)
	_, _, kvdb := util.CreateTestDB()

	accA, _ := account.NewAccountDB(AssetExecPara, Symbol, stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ := account.NewAccountDB(AssetExecPara, Symbol, stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	env := execEnv{
		10,
		types.GetDappFork(pty.TokenX, pty.ForkTokenCheckX),
		1539918074,
	}

	// set config key
	item := &types.ConfigItem{
		Key: "mavl-manage-token-blacklist",
		Value: &types.ConfigItem_Arr{
			Arr: &types.ArrayConfig{Value: []string{"bty"}},
		},
	}
	stateDB.Set([]byte(item.Key), types.Encode(item))

	item2 := &types.ConfigItem{
		Key: "mavl-manage-token-finisher",
		Value: &types.ConfigItem_Arr{
			Arr: &types.ArrayConfig{Value: []string{string(Nodes[0])}},
		},
	}
	stateDB.Set([]byte(item2.Key), types.Encode(item2))

	// create token
	// 创建
	//ty := pty.TokenType{}
	p1 := &pty.TokenPreCreate{
		Name:         Symbol,
		Symbol:       Symbol,
		Introduction: Symbol,
		Total:        tokenTotal,
		Price:        0,
		Owner:        string(Nodes[0][1:]),
		Category:     pty.CategoryMintBurnSupport,
	}
	//v, _ := types.PBToJSON(p1)
	createTx, err := types.CallCreateTransaction(pty.TokenX, "TokenPreCreate", p1)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}
	exec := newToken()
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(createTx, int(1))
	assert.NotNil(t, err)
	assert.Nil(t, receipt)

	p1 = &pty.TokenPreCreate{
		Name:         Symbol,
		Symbol:       Symbol,
		Introduction: Symbol,
		Total:        tokenTotal,
		Price:        0,
		Owner:        string(Nodes[0]),
		Category:     pty.CategoryMintBurnSupport,
	}
	//v, _ := types.PBToJSON(p1)
	createTx, err = types.CallCreateTransaction(pty.TokenX, "TokenPreCreate", p1)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}
	exec = newToken()
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(createTx, receiptDate, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)

	p2 := &pty.TokenFinishCreate{
		Symbol: Symbol,
		Owner:  string(Nodes[0]),
	}
	//v, _ := types.PBToJSON(p1)
	createTx2, err := types.CallCreateTransaction(pty.TokenX, "TokenFinishCreate", p2)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx2, err = signTx(createTx2, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	receipt, err = exec.Exec(createTx2, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	accDB, _ := account.NewAccountDB(pty.TokenX, Symbol, stateDB)
	accChcek := accDB.LoadAccount(string(Nodes[0]))
	assert.Equal(t, tokenTotal, accChcek.Balance)

	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx2, receiptDate, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)

	// mint burn
	p3 := &pty.TokenMint{
		Symbol: Symbol,
		Amount: tokenMint,
	}
	//v, _ := types.PBToJSON(p1)
	createTx3, err := types.CallCreateTransaction(pty.TokenX, "TokenMint", p3)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx3, err = signTx(createTx3, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight+2, env.blockTime+2, env.difficulty)
	receipt, err = exec.Exec(createTx3, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	accChcek = accDB.LoadAccount(string(Nodes[0]))
	assert.Equal(t, tokenTotal+tokenMint, accChcek.Balance)

	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx3, receiptDate, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)

	p4 := &pty.TokenBurn{
		Symbol: Symbol,
		Amount: tokenBurn,
	}
	//v, _ := types.PBToJSON(p1)
	createTx4, err := types.CallCreateTransaction(pty.TokenX, "TokenBurn", p4)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx4, err = signTx(createTx4, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	receipt, err = exec.Exec(createTx4, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	accChcek = accDB.LoadAccount(string(Nodes[0]))
	assert.Equal(t, tokenTotal+tokenMint-tokenBurn, accChcek.Balance)

	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx4, receiptDate, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.New(types.GetSignName(pty.TokenX, signType))
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
