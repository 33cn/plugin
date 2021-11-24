package executor

import (
	"testing"
	"time"

	"github.com/33cn/chain33/client"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	pty "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pkt "github.com/33cn/plugin/plugin/dapp/issuance/types"
	tokenE "github.com/33cn/plugin/plugin/dapp/token/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
	kvdb        dbm.KVDB
	api         client.QueueProtocolAPI
	db          dbm.KV
	execAddr    string
	cfg         *types.Chain33Config
	ldb         dbm.DB
}

var (
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0xc2b31057b8692a56c7dd18199df71c1d21b781c0b6858c52997c9dbf778e8550" // 12evczYyX9ZKPYvwSEvRkRyTjpSrJuLudg
	Nodes    = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("12evczYyX9ZKPYvwSEvRkRyTjpSrJuLudg"),
	}
	total      = 10000 * types.DefaultCoinPrecision
	totalToken = 100000 * types.DefaultCoinPrecision
)

func manageKeySet(key string, value string, db dbm.KV) {
	var item types.ConfigItem
	item.Key = key
	item.Addr = value
	item.Ty = pty.ConfigItemArrayConfig
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	item.GetArr().Value = append(item.GetArr().Value, value)

	manageKey := types.ManageKey(key)
	valueSave := types.Encode(&item)
	db.Set([]byte(manageKey), valueSave)
}

func initEnv() *execEnv {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	cfg.RegisterDappFork(pkt.IssuanceX, pkt.ForkIssuanceTableUpdate, 0)
	Init(pkt.IssuanceX, cfg, nil)
	_, ldb, kvdb := util.CreateTestDB()

	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[0]),
	}
	accountAToken := types.Account{
		Balance: totalToken,
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

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)

	execAddr := dapp.ExecAddress(pkt.IssuanceX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 100)

	accA := account.NewCoinsAccount(cfg)
	accA.SetDB(stateDB)
	accA.SaveExecAccount(execAddr, &accountA)
	manageKeySet("issuance-manage", accountA.Addr, stateDB)
	manageKeySet("issuance-fund", accountA.Addr, stateDB)
	tokenAccA, _ := account.NewAccountDB(cfg, tokenE.GetName(), pkt.CCNYTokenName, stateDB)
	tokenAccA.SaveExecAccount(execAddr, &accountAToken)

	accB := account.NewCoinsAccount(cfg)
	accB.SetDB(stateDB)
	accB.SaveExecAccount(execAddr, &accountB)
	manageKeySet("issuance-price-feed", accountB.Addr, stateDB)

	accC := account.NewCoinsAccount(cfg)
	accC.SetDB(stateDB)
	accC.SaveExecAccount(execAddr, &accountC)
	manageKeySet("issuance-guarantor", accountC.Addr, stateDB)

	return &execEnv{
		blockTime:   time.Now().Unix(),
		blockHeight: cfg.GetDappFork(pkt.IssuanceX, "Enable"),
		difficulty:  1539918074,
		kvdb:        kvdb,
		api:         api,
		db:          stateDB,
		execAddr:    execAddr,
		cfg:         cfg,
		ldb:         ldb,
	}
}

func TestIssuance(t *testing.T) {
	env := initEnv()

	// issuance create
	p1 := &pkt.IssuanceCreateTx{
		TotalBalance:     1000,
		DebtCeiling:      200,
		LiquidationRatio: 0.25,
		Period:           5,
	}
	createTx, err := pkt.CreateRawIssuanceCreateTx(env.cfg, p1)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	exec := newIssuance()
	exec.SetAPI(env.api)
	exec.SetStateDB(env.db)
	assert.Equal(t, exec.GetCoinsAccount().LoadExecAccount(string(Nodes[0]), env.execAddr).GetBalance(), total)
	exec.SetLocalDB(env.kvdb)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
	issuanceID := createTx.Hash()
	// query issuance by id
	res, err := exec.Query("IssuanceInfoByID", types.Encode(&pkt.ReqIssuanceInfo{IssuanceId: common.ToHex(issuanceID)}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query issuance by status
	res, err = exec.Query("IssuanceByStatus", types.Encode(&pkt.ReqIssuanceByStatus{Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query issuances by ids
	var issuanceIDsS []string
	issuanceIDsS = append(issuanceIDsS, common.ToHex(issuanceID))
	res, err = exec.Query("IssuanceInfoByIDs", types.Encode(&pkt.ReqIssuanceInfos{IssuanceIds: issuanceIDsS}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	// issuance price
	p2 := &pkt.IssuanceFeedTx{}
	p2.Price = append(p2.Price, 1)
	p2.Volume = append(p2.Volume, 100)
	createTx, err = pkt.CreateRawIssuanceFeedTx(env.cfg, p2)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}

	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
	// query issuance by id
	res, err = exec.Query("IssuancePrice", nil)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	// issuance manage
	p3 := &pkt.IssuanceManageTx{}
	p3.Addr = append(p3.Addr, string(Nodes[1]))
	p3.Addr = append(p3.Addr, string(Nodes[2]))
	createTx, err = pkt.CreateRawIssuanceManageTx(env.cfg, p3)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)

	// issuance debt
	p4 := &pkt.IssuanceDebtTx{
		IssuanceID: common.ToHex(issuanceID),
		Value:      100,
	}
	createTx, err = pkt.CreateRawIssuanceDebtTx(env.cfg, p4)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
	debtID := createTx.Hash()
	// query issuance by id
	res, err = exec.Query("IssuanceRecordByID",
		types.Encode(&pkt.ReqIssuanceRecords{IssuanceId: common.ToHex(issuanceID), DebtId: common.ToHex(debtID)}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query issuance by status
	res, err = exec.Query("IssuanceRecordsByStatus",
		types.Encode(&pkt.ReqIssuanceRecords{Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query issuance by addr
	res, err = exec.Query("IssuanceRecordsByAddr",
		types.Encode(&pkt.ReqIssuanceRecords{Addr: string(Nodes[1])}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	res, err = exec.Query("IssuanceRecordsByAddr",
		types.Encode(&pkt.ReqIssuanceRecords{Addr: string(Nodes[1]), Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query issuance user balance
	res, err = exec.Query("IssuanceUserBalance",
		types.Encode(&pkt.ReqIssuanceRecords{Addr: string(Nodes[1]), Status: 1}))
	assert.Nil(t, err)
	assert.Equal(t, 100*types.DefaultCoinPrecision, res.(*pkt.RepIssuanceUserBalance).Balance)

	// issuance repay
	p5 := &pkt.IssuanceRepayTx{
		IssuanceID: common.ToHex(issuanceID),
		DebtID:     common.ToHex(debtID),
	}
	createTx, err = pkt.CreateRawIssuanceRepayTx(env.cfg, p5)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
	// query issuance by status
	res, err = exec.Query("IssuanceRecordsByStatus",
		types.Encode(&pkt.ReqIssuanceRecords{Status: 6}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query issuance by addr
	res, err = exec.Query("IssuanceRecordsByAddr",
		types.Encode(&pkt.ReqIssuanceRecords{Addr: string(Nodes[1])}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	res, err = exec.Query("IssuanceRecordsByAddr",
		types.Encode(&pkt.ReqIssuanceRecords{Addr: string(Nodes[1]), Status: 6}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query issuance user balance
	res, err = exec.Query("IssuanceUserBalance",
		types.Encode(&pkt.ReqIssuanceRecords{Addr: string(Nodes[1]), Status: 1}))
	assert.Nil(t, err)
	assert.Equal(t, int64(0), res.(*pkt.RepIssuanceUserBalance).Balance)

	// issuance liquidate
	p6 := &pkt.IssuanceDebtTx{
		IssuanceID: common.ToHex(issuanceID),
		Value:      50,
	}
	createTx, err = pkt.CreateRawIssuanceDebtTx(env.cfg, p6)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)

	p61 := &pkt.IssuanceDebtTx{
		IssuanceID: common.ToHex(issuanceID),
		Value:      50,
	}
	createTx, err = pkt.CreateRawIssuanceDebtTx(env.cfg, p61)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
	createTx, err = signTx(createTx, PrivKeyC)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	t.Log(receipt)
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)

	p7 := &pkt.IssuanceFeedTx{}
	p7.Price = append(p7.Price, 0.28)
	p7.Volume = append(p7.Volume, 100)
	createTx, err = pkt.CreateRawIssuanceFeedTx(env.cfg, p7)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
	// query issuance by status
	res, err = exec.Query("IssuanceRecordsByStatus",
		types.Encode(&pkt.ReqIssuanceRecords{Status: 2}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	res, err = exec.Query("IssuanceRecordsByStatus",
		types.Encode(&pkt.ReqIssuanceRecords{Status: 4}))
	assert.Nil(t, res)
	assert.NotNil(t, err)

	p8 := &pkt.IssuanceFeedTx{}
	p8.Price = append(p8.Price, 0.5)
	p8.Volume = append(p8.Volume, 100)
	createTx, err = pkt.CreateRawIssuanceFeedTx(env.cfg, p8)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
	// query issuance by status
	res, err = exec.Query("IssuanceRecordsByStatus", types.Encode(&pkt.ReqIssuanceRecords{Status: 4}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	p81 := &pkt.IssuanceFeedTx{}
	p81.Price = append(p81.Price, 0.25)
	p81.Volume = append(p81.Volume, 100)
	createTx, err = pkt.CreateRawIssuanceFeedTx(env.cfg, p81)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
	// query issuance by status
	res, err = exec.Query("IssuanceRecordsByStatus", types.Encode(&pkt.ReqIssuanceRecords{Status: 3}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	res, err = exec.Query("IssuanceRecordsByStatus", types.Encode(&pkt.ReqIssuanceRecords{Status: 4}))
	assert.NotNil(t, err)
	assert.Nil(t, res)
	res, err = exec.Query("IssuanceRecordsByStatus", types.Encode(&pkt.ReqIssuanceRecords{Status: 1}))
	assert.Nil(t, res)
	assert.NotNil(t, err)

	// expire liquidate
	p9 := &pkt.IssuanceDebtTx{
		IssuanceID: common.ToHex(issuanceID),
		Value:      100,
	}
	createTx, err = pkt.CreateRawIssuanceDebtTx(env.cfg, p9)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)

	p10 := &pkt.IssuanceFeedTx{}
	p10.Price = append(p10.Price, 1)
	p10.Volume = append(p10.Volume, 100)
	createTx, err = pkt.CreateRawIssuanceFeedTx(env.cfg, p10)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
	createTx, err = signTx(createTx, PrivKeyB)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}
	exec.SetEnv(env.blockHeight+1, env.blockTime+6, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	t.Log(receipt)
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
	// query issuance by status
	res, err = exec.Query("IssuanceRecordsByStatus", types.Encode(&pkt.ReqIssuanceRecords{Status: 5}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	// issuance close
	p11 := &pkt.IssuanceCloseTx{
		IssuanceID: common.ToHex(issuanceID),
	}
	createTx, err = pkt.CreateRawIssuanceCloseTx(env.cfg, p11)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight+2, env.blockTime+2, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
	// query issuance by status
	res, err = exec.Query("IssuanceByStatus", types.Encode(&pkt.ReqIssuanceByStatus{Status: 2}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	// issuance create
	p12 := &pkt.IssuanceCreateTx{
		TotalBalance:     200,
		DebtCeiling:      100,
		LiquidationRatio: 0.25,
		Period:           5,
	}
	createTx, err = pkt.CreateRawIssuanceCreateTx(env.cfg, p12)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	receipt, err = exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
	issuanceID = createTx.Hash()
	// query issuance by id
	res, err = exec.Query("IssuanceInfoByID", types.Encode(&pkt.ReqIssuanceInfo{IssuanceId: common.ToHex(issuanceID)}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query issuance by status
	res, err = exec.Query("IssuanceByStatus", types.Encode(&pkt.ReqIssuanceByStatus{Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	p13 := &pkt.IssuanceDebtTx{
		IssuanceID: common.ToHex(issuanceID),
		Value:      100,
	}
	createTx, err = pkt.CreateRawIssuanceDebtTx(env.cfg, p13)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)

	p14 := &pkt.IssuanceFeedTx{}
	p14.Price = append(p14.Price, 0.25)
	p14.Volume = append(p14.Volume, 100)
	createTx, err = pkt.CreateRawIssuanceFeedTx(env.cfg, p14)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
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
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)

	// issuance close
	p15 := &pkt.IssuanceCloseTx{
		IssuanceID: common.ToHex(issuanceID),
	}
	createTx, err = pkt.CreateRawIssuanceCloseTx(env.cfg, p15)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.IssuanceX)
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight+2, env.blockTime+2, env.difficulty)
	receipt, err = exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName(pkt.IssuanceX, signType), -1)
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
