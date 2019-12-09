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
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
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

func TestExchange(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	Init(et.ExchangeX, cfg, nil)
	total := 100 * types.Coin
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
	execAddr := address.ExecAddress(et.ExchangeX)
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

	accA1, _ := account.NewAccountDB(cfg, "token", "CCNY", stateDB)
	accA1.SaveExecAccount(execAddr, &accountA)

	accB1, _ := account.NewAccountDB(cfg, "paracross", "coins.bty", stateDB)
	accB1.SaveExecAccount(execAddr, &accountB)

	accC1, _ := account.NewAccountDB(cfg, "paracross", "token.CCNY", stateDB)
	accC1.SaveExecAccount(execAddr, &accountC)

	accD1, _ := account.NewAccountDB(cfg, "token", "para", stateDB)
	accD1.SaveExecAccount(execAddr, &accountD)

	env := execEnv{
		10,
		cfg.GetDappFork(et.ExchangeX, "Enable"),
		1539918074,
	}

	// orderlimit  bty:CCNY  买bty
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, err := ety.Create("LimitOrder", &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 4, Amount: 10 * types.Coin, Op: et.OpBuy})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, et.ExchangeX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	exec := newExchange()
	e := exec.(*exchange)
	err = e.CheckTx(tx, 1)
	assert.Nil(t, err)
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	env.blockHeight = env.blockHeight + 1
	env.blockTime = env.blockTime + 20
	env.difficulty = env.difficulty + 1
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(tx, receiptData, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	orderID1 := common.ToHex(tx.Hash())
	//根据订单号，查询订单详情
	msg, err := exec.Query(et.FuncNameQueryOrder, types.Encode(&et.QueryOrder{OrderID: orderID1}))
	if err != nil {
		t.Error(err)
	}
	reply := msg.(*et.Order)
	t.Log(reply)
	assert.Equal(t, int32(et.Ordered), reply.Status)
	assert.Equal(t, 10*types.Coin, reply.GetBalance())
	//查看账户余额
	acc := accA1.LoadExecAccount(string(Nodes[0]), execAddr)
	t.Log(acc)
	//根据op查询市场深度
	msg, err = exec.Query(et.FuncNameQueryMarketDepth, types.Encode(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpBuy}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply1 := msg.(*et.MarketDepthList)
	assert.Equal(t, 10*types.Coin, reply1.List[0].GetAmount())

	//根据状态和地址查询
	msg, err = exec.Query(et.FuncNameQueryOrderList, types.Encode(&et.QueryOrderList{Status: et.Ordered, Address: string(Nodes[0])}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply2 := msg.(*et.OrderList)
	assert.Equal(t, orderID1, reply2.List[0].OrderID)

	// orderlimit  bty:CCNY 卖bty
	tx, err = ety.Create("LimitOrder", &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 4, Amount: 5 * types.Coin, Op: et.OpSell})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, et.ExchangeX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	err = e.CheckTx(tx, 1)
	assert.Nil(t, err)
	env.blockHeight = env.blockHeight + 1
	env.blockTime = env.blockTime + 20
	env.difficulty = env.difficulty + 1
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptData, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	orderID2 := common.ToHex(tx.Hash())
	//根据订单号，查询订单详情
	msg, err = exec.Query(et.FuncNameQueryOrder, types.Encode(&et.QueryOrder{OrderID: orderID1}))
	if err != nil {
		t.Error(err)
	}
	//订单1的状态应该还是ordered
	reply = msg.(*et.Order)
	assert.Equal(t, int32(et.Ordered), reply.Status)
	t.Log(reply)

	msg, err = exec.Query(et.FuncNameQueryOrder, types.Encode(&et.QueryOrder{OrderID: orderID2}))
	if err != nil {
		t.Error(err)
	}

	reply = msg.(*et.Order)
	assert.Equal(t, int32(et.Completed), reply.Status)
	//根据op查询市场深度
	msg, err = exec.Query(et.FuncNameQueryMarketDepth, types.Encode(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpBuy}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply1 = msg.(*et.MarketDepthList)
	t.Log(reply1.List)
	//市场深度应该改变
	assert.Equal(t, 5*types.Coin, reply1.List[0].GetAmount())

	//QueryCompletedOrderList
	msg, err = exec.Query(et.FuncNameQueryCompletedOrderList, types.Encode(&et.QueryCompletedOrderList{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}}))
	if err != nil {
		t.Error(err)
	}
	reply2 = msg.(*et.OrderList)
	assert.Equal(t, orderID2, reply2.List[0].OrderID)
	//撤回之前的订单
	// orderlimit  bty:CCNY
	tx, err = ety.Create("RevokeOrder", &et.RevokeOrder{OrderID: orderID1})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, et.ExchangeX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	err = e.CheckTx(tx, 1)
	assert.Nil(t, err)

	env.blockHeight = env.blockHeight + 1
	env.blockTime = env.blockTime + 20
	env.difficulty = env.difficulty + 1
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptData, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	//根据订单号，查询订单详情
	msg, err = exec.Query(et.FuncNameQueryOrder, types.Encode(&et.QueryOrder{OrderID: orderID1}))
	if err != nil {
		t.Error(err)
	}
	reply = msg.(*et.Order)
	assert.Equal(t, int32(et.Revoked), reply.Status)
	t.Log(reply)
	//根据op查询市场深度
	msg, err = exec.Query(et.FuncNameQueryMarketDepth, types.Encode(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpBuy}))
	if err != nil {
		t.Error(err)
	}
	reply1 = msg.(*et.MarketDepthList)
	t.Log(reply1.GetList())
	t.Log(len(reply1.GetList()))

	//反向测试
	// orderlimit  bty:CCNY 卖bty
	tx, err = ety.Create("LimitOrder", &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Price: 0.5, Amount: 10 * types.Coin, Op: et.OpSell})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, et.ExchangeX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	err = e.CheckTx(tx, 1)
	assert.Nil(t, err)
	env.blockHeight = env.blockHeight + 1
	env.blockTime = env.blockTime + 20
	env.difficulty = env.difficulty + 1
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptData, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	orderID3 := common.ToHex(tx.Hash())
	msg, err = exec.Query(et.FuncNameQueryOrder, types.Encode(&et.QueryOrder{OrderID: orderID3}))
	if err != nil {
		t.Error(err)
	}
	//订单1的状态应该还是ordered
	reply = msg.(*et.Order)
	assert.Equal(t, int32(et.Ordered), reply.Status)
	t.Log(reply)

	tx, err = ety.Create("LimitOrder", &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Price: 0.5, Amount: 10 * types.Coin, Op: et.OpSell})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, et.ExchangeX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	err = e.CheckTx(tx, 1)
	assert.Nil(t, err)
	env.blockHeight = env.blockHeight + 1
	env.blockTime = env.blockTime + 20
	env.difficulty = env.difficulty + 1
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptData, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	orderID4 := common.ToHex(tx.Hash())
	//根据订单号，查询订单详情
	msg, err = exec.Query(et.FuncNameQueryOrder, types.Encode(&et.QueryOrder{OrderID: orderID4}))
	if err != nil {
		t.Error(err)
	}
	//订单1的状态应该还是ordered
	reply = msg.(*et.Order)
	assert.Equal(t, int32(et.Ordered), reply.Status)
	t.Log(reply)

	//根据op查询市场深度
	msg, err = exec.Query(et.FuncNameQueryMarketDepth, types.Encode(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Op: et.OpSell}))
	if err != nil {
		t.Error(err)
	}
	reply1 = msg.(*et.MarketDepthList)
	t.Log(reply1.List)
	//市场深度应该改变
	assert.Equal(t, 20*types.Coin, reply1.List[0].GetAmount())
	//根据状态和地址查询
	msg, err = exec.Query(et.FuncNameQueryOrderList, types.Encode(&et.QueryOrderList{Status: et.Ordered, Address: string(Nodes[0])}))
	if err != nil {
		t.Error(err)
	}
	reply2 = msg.(*et.OrderList)
	t.Log(reply2)
	//默认倒序查询
	assert.Equal(t, orderID3, reply2.List[1].OrderID)
	assert.Equal(t, orderID4, reply2.List[0].OrderID)

	tx, err = ety.Create("LimitOrder", &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Price: 0.5, Amount: 20 * types.Coin, Op: et.OpBuy})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, et.ExchangeX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	err = e.CheckTx(tx, 1)
	assert.Nil(t, err)
	env.blockHeight = env.blockHeight + 1
	env.blockTime = env.blockTime + 20
	env.difficulty = env.difficulty + 1
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptData, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	orderID5 := common.ToHex(tx.Hash())

	//根据订单号，查询订单详情
	msg, err = exec.Query(et.FuncNameQueryOrder, types.Encode(&et.QueryOrder{OrderID: orderID5}))
	if err != nil {
		t.Error(err)
	}
	//订单1的状态应该还是ordered
	reply = msg.(*et.Order)
	assert.Equal(t, int32(et.Completed), reply.Status)
	//根据op查询市场深度
	msg, err = exec.Query(et.FuncNameQueryMarketDepth, types.Encode(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Op: et.OpSell}))
	if err != nil {
		t.Error(err)
	}
	reply1 = msg.(*et.MarketDepthList)
	t.Log(reply1.List)
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.New(types.GetSignName("", signType))
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

func TestTruncate(t *testing.T) {
	a := float32(1.00000212000000000001)
	b := float32(0.34567)
	c := float32(1234)
	t.Log(Truncate(a))
	t.Log(Truncate(b))
	t.Log(Truncate(c))
}

func TestCheckPrice(t *testing.T) {
	t.Log(CheckPrice(0.25))
}
