package executor

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/queue"
	oty "github.com/33cn/plugin/plugin/dapp/oracle/types"
	"github.com/stretchr/testify/assert"

	//"github.com/33cn/chain33/types/jsonpb"
	"strings"
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

var (
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4

	Nodes = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}
)
var (
	r *rand.Rand
)

func init() {
	r = rand.New(rand.NewSource(types.Now().UnixNano()))
}
func TestOrace(t *testing.T) {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	Init(oty.OracleX, cfg, nil)
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
	execAddr := address.ExecAddress(oty.OracleX)
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
		cfg.GetDappFork(oty.OracleX, "Enable"),
		1539918074,
	}

	// set config key 授权
	item := &types.ConfigItem{
		Key: "mavl-manage-oracle-publish-event",
		Value: &types.ConfigItem_Arr{
			Arr: &types.ArrayConfig{Value: []string{"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"}},
		},
	}
	stateDB.Set([]byte(item.Key), types.Encode(item))

	// publish event
	ety := types.LoadExecutorType(oty.OracleX)
	tx, err := ety.Create("EventPublish", &oty.EventPublish{Type: "football", SubType: "Premier League", Time: time.Now().AddDate(0, 0, 1).Unix(),
		Content:      fmt.Sprintf("{\"team%d\":\"ChelSea\", \"team%d\":\"Manchester\",\"resultType\":\"score\"}", r.Int()%10, r.Int()%10),
		Introduction: "guess the sore result of football game between ChelSea and Manchester in 2019-01-21 14:00:00"})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.OracleX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	t.Log("tx", tx)
	exec := newOracle()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(1, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	var eventID string
	//获取eventID
	for _, log := range receipt.Logs {
		if log.Ty >= oty.TyLogEventPublish && log.Ty <= oty.TyLogResultPublish {
			status := oty.ReceiptOracle{}
			err := types.Decode(log.Log, &status)
			assert.Nil(t, err)
			eventID = status.EventID
		}
	}
	t.Log("eventID:", eventID)
	//查询事件
	msg, err := exec.Query(oty.FuncNameQueryOracleListByIDs, types.Encode(&oty.QueryOracleInfos{
		EventID: []string{eventID}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply := msg.(*oty.ReplyOracleStatusList)
	assert.Equal(t, int32(oty.EventPublished), reply.Status[0].Status.Status)

	//通过状态查询eventID
	msg, err = exec.Query(oty.FuncNameQueryEventIDByStatus, types.Encode(&oty.QueryEventID{
		Status: int32(oty.EventPublished)}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)
	reply2 := msg.(*oty.ReplyEventIDs)
	assert.Equal(t, eventID, reply2.EventID[0])

	//通过状态和地址查询
	msg, err = exec.Query(oty.FuncNameQueryEventIDByAddrAndStatus, types.Encode(&oty.QueryEventID{
		Status: int32(oty.EventPublished), Addr: string(Nodes[0])}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)
	reply3 := msg.(*oty.ReplyEventIDs)
	assert.Equal(t, eventID, reply3.EventID[0])

	//通过类型和状态查询
	msg, err = exec.Query(oty.FuncNameQueryEventIDByTypeAndStatus, types.Encode(&oty.QueryEventID{
		Status: int32(oty.EventPublished), Type: "football"}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)
	reply4 := msg.(*oty.ReplyEventIDs)
	assert.Equal(t, eventID, reply4.EventID[0])

	//pre publish result
	tx, err = ety.Create("ResultPrePublish", &oty.ResultPrePublish{EventID: eventID,
		Result: fmt.Sprintf("%d:%d", r.Int()%10, r.Int()%10),
		Source: "sina sport"})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.OracleX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)

	exec.SetEnv(env.blockHeight+1, env.blockTime+20, env.difficulty+1)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	msg, err = exec.Query(oty.FuncNameQueryOracleListByIDs, types.Encode(&oty.QueryOracleInfos{
		EventID: []string{eventID}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply = msg.(*oty.ReplyOracleStatusList)
	assert.Equal(t, int32(oty.ResultPrePublished), reply.Status[0].Status.Status)

	//publish result

	tx, err = ety.Create("ResultPublish", &oty.ResultPublish{EventID: eventID,
		Result: fmt.Sprintf("%d:%d", r.Int()%10, r.Int()%10),
		Source: "sina sport"})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.OracleX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+1, env.blockTime+20, env.difficulty+1)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	msg, err = exec.Query(oty.FuncNameQueryOracleListByIDs, types.Encode(&oty.QueryOracleInfos{
		EventID: []string{eventID}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply = msg.(*oty.ReplyOracleStatusList)
	assert.Equal(t, int32(oty.ResultPublished), reply.Status[0].Status.Status)

	// publish event
	tx, err = ety.Create("EventPublish", &oty.EventPublish{Type: "football", SubType: "Premier League", Time: time.Now().AddDate(0, 0, 1).Unix(),
		Content:      fmt.Sprintf("{\"team%d\":\"ChelSea\", \"team%d\":\"Manchester\",\"resultType\":\"score\"}", r.Int()%10, r.Int()%10),
		Introduction: "guess the sore result of football game between ChelSea and Manchester in 2019-03-21 14:00:00"})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.OracleX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	t.Log("tx", tx)

	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	//var eventID string
	//获取eventID
	for _, log := range receipt.Logs {
		if log.Ty >= oty.TyLogEventPublish && log.Ty <= oty.TyLogResultPublish {
			status := oty.ReceiptOracle{}
			err := types.Decode(log.Log, &status)
			assert.Nil(t, err)
			eventID = status.EventID
		}
	}
	t.Log("eventID:", eventID)
	//查询事件
	msg, err = exec.Query(oty.FuncNameQueryOracleListByIDs, types.Encode(&oty.QueryOracleInfos{
		EventID: []string{eventID}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply = msg.(*oty.ReplyOracleStatusList)
	assert.Equal(t, int32(oty.EventPublished), reply.Status[0].Status.Status)

	//EventAbort
	tx, err = ety.Create("EventAbort", &oty.EventAbort{EventID: eventID})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.OracleX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+1, env.blockTime+20, env.difficulty+1)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	msg, err = exec.Query(oty.FuncNameQueryOracleListByIDs, types.Encode(&oty.QueryOracleInfos{
		EventID: []string{eventID}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply = msg.(*oty.ReplyOracleStatusList)
	assert.Equal(t, int32(oty.EventAborted), reply.Status[0].Status.Status)

	//pre publish result
	tx, err = ety.Create("ResultPrePublish", &oty.ResultPrePublish{EventID: eventID,
		Result: fmt.Sprintf("%d:%d", r.Int()%10, r.Int()%10),
		Source: "sina sport"})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.OracleX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+1, env.blockTime+20, env.difficulty+1)
	receipt, err = exec.Exec(tx, int(1))
	assert.Equal(t, oty.ErrResultPrePublishNotAllowed, err)

	// publish event
	tx, err = ety.Create("EventPublish", &oty.EventPublish{Type: "football", SubType: "Premier League", Time: time.Now().AddDate(0, 0, 1).Unix(),
		Content:      fmt.Sprintf("{\"team%d\":\"ChelSea\", \"team%d\":\"Manchester\",\"resultType\":\"score\"}", r.Int()%10, r.Int()%10),
		Introduction: "guess the sore result of football game between ChelSea and Manchester in 2019-03-21 14:00:00"})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.OracleX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	t.Log("tx", tx)

	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	//var eventID string
	//获取eventID
	for _, log := range receipt.Logs {
		if log.Ty >= oty.TyLogEventPublish && log.Ty <= oty.TyLogResultPublish {
			status := oty.ReceiptOracle{}
			err := types.Decode(log.Log, &status)
			assert.Nil(t, err)
			eventID = status.EventID
		}
	}
	t.Log("eventID:", eventID)
	//查询事件
	msg, err = exec.Query(oty.FuncNameQueryOracleListByIDs, types.Encode(&oty.QueryOracleInfos{
		EventID: []string{eventID}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply = msg.(*oty.ReplyOracleStatusList)
	assert.Equal(t, int32(oty.EventPublished), reply.Status[0].Status.Status)
	//pre publish result
	tx, err = ety.Create("ResultPrePublish", &oty.ResultPrePublish{EventID: eventID,
		Result: fmt.Sprintf("%d:%d", r.Int()%10, r.Int()%10),
		Source: "sina sport"})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.OracleX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+1, env.blockTime+20, env.difficulty+1)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	msg, err = exec.Query(oty.FuncNameQueryOracleListByIDs, types.Encode(&oty.QueryOracleInfos{
		EventID: []string{eventID}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply = msg.(*oty.ReplyOracleStatusList)
	assert.Equal(t, int32(oty.ResultPrePublished), reply.Status[0].Status.Status)

	//ResultAbort
	tx, err = ety.Create("ResultAbort", &oty.ResultAbort{EventID: eventID})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.OracleX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+1, env.blockTime+20, env.difficulty+1)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	msg, err = exec.Query(oty.FuncNameQueryOracleListByIDs, types.Encode(&oty.QueryOracleInfos{
		EventID: []string{eventID}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply = msg.(*oty.ReplyOracleStatusList)
	assert.Equal(t, int32(oty.ResultAborted), reply.Status[0].Status.Status)

}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName(oty.OracleX, signType), -1)
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
