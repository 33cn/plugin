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
	pkt "github.com/33cn/plugin/plugin/dapp/collateralize/types"
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

func addrKeySet(value string, db dbm.KV) {
	var item types.ConfigItem
	item.Addr = value
	item.Ty = pty.ConfigItemArrayConfig
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	item.GetArr().Value = append(item.GetArr().Value, value)

	valueSave := types.Encode(&item)
	db.Set(AddrKey(), valueSave)
}

func initEnv() *execEnv {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	cfg.RegisterDappFork(pkt.CollateralizeX, pkt.ForkCollateralizeTableUpdate, 0)
	Init(pkt.CollateralizeX, cfg, nil)
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
	accountBToken := types.Account{
		Balance: types.DefaultCoinPrecision / 10,
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

	execAddr := dapp.ExecAddress(pkt.CollateralizeX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 100)

	accA := account.NewCoinsAccount(cfg)
	accA.SetDB(stateDB)
	accA.SaveExecAccount(execAddr, &accountA)
	manageKeySet("issuance-manage", accountA.Addr, stateDB)
	addrKeySet(accountA.Addr, stateDB)
	tokenAccA, _ := account.NewAccountDB(cfg, tokenE.GetName(), pkt.CCNYTokenName, stateDB)
	tokenAccA.SaveExecAccount(execAddr, &accountAToken)

	accB := account.NewCoinsAccount(cfg)
	accB.SetDB(stateDB)
	accB.SaveExecAccount(execAddr, &accountB)
	manageKeySet("issuance-price-feed", accountB.Addr, stateDB)
	tokenAccB, _ := account.NewAccountDB(cfg, tokenE.GetName(), pkt.CCNYTokenName, stateDB)
	tokenAccB.SaveExecAccount(execAddr, &accountBToken)

	accC := account.NewCoinsAccount(cfg)
	accC.SetDB(stateDB)
	accC.SaveExecAccount(execAddr, &accountC)
	manageKeySet("issuance-guarantor", accountC.Addr, stateDB)

	return &execEnv{
		blockTime:   time.Now().Unix(),
		blockHeight: cfg.GetDappFork(pkt.CollateralizeX, "Enable"),
		difficulty:  1539918074,
		kvdb:        kvdb,
		api:         api,
		db:          stateDB,
		execAddr:    execAddr,
		cfg:         cfg,
		ldb:         ldb,
	}
}

func TestCollateralize(t *testing.T) {
	env := initEnv()

	// collateralize manage
	p3 := &pkt.CollateralizeManageTx{}
	p3.Period = 5
	p3.LiquidationRatio = 0.25
	p3.DebtCeiling = 100
	p3.StabilityFeeRatio = 0.0001
	p3.TotalBalance = 10000
	createTx, err := pkt.CreateRawCollateralizeManageTx(env.cfg, p3)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
	createTx, err = signTx(createTx, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	exec := newCollateralize()
	exec.SetAPI(env.api)
	exec.SetStateDB(env.db)
	assert.Equal(t, exec.GetCoinsAccount().LoadExecAccount(string(Nodes[0]), env.execAddr).GetBalance(), total)
	exec.SetLocalDB(env.kvdb)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	receipt, err := exec.Exec(createTx, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	t.Log(receipt)
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(createTx, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)

	// collateralize create
	p1 := &pkt.CollateralizeCreateTx{
		TotalBalance: 1000,
	}
	createTx, err = pkt.CreateRawCollateralizeCreateTx(env.cfg, p1)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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
	collateralizeID := createTx.Hash()
	// query collateralize by id
	res, err := exec.Query("CollateralizeInfoByID", types.Encode(&pkt.ReqCollateralizeInfo{CollateralizeId: common.ToHex(collateralizeID)}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query collateralize by status
	res, err = exec.Query("CollateralizeByStatus", types.Encode(&pkt.ReqCollateralizeByStatus{Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query collateralizes by ids
	var collateralizeIDsS []string
	collateralizeIDsS = append(collateralizeIDsS, common.ToHex(collateralizeID))
	res, err = exec.Query("CollateralizeInfoByIDs", types.Encode(&pkt.ReqCollateralizeInfos{CollateralizeIds: collateralizeIDsS}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	// collateralize price
	p2 := &pkt.CollateralizeFeedTx{}
	p2.Price = append(p2.Price, 1)
	p2.Volume = append(p2.Volume, 100)
	createTx, err = pkt.CreateRawCollateralizeFeedTx(env.cfg, p2)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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
	// query collateralize by id
	res, err = exec.Query("CollateralizePrice", nil)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	// collateralize borrow
	p4 := &pkt.CollateralizeBorrowTx{
		CollateralizeID: common.ToHex(collateralizeID),
		Value:           100,
	}
	createTx, err = pkt.CreateRawCollateralizeBorrowTx(env.cfg, p4)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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
	borrowID := createTx.Hash()
	// query collateralize by id
	res, err = exec.Query("CollateralizeRecordByID",
		types.Encode(&pkt.ReqCollateralizeRecord{CollateralizeId: common.ToHex(collateralizeID), RecordId: common.ToHex(borrowID)}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query collateralize by status
	res, err = exec.Query("CollateralizeRecordByStatus",
		types.Encode(&pkt.ReqCollateralizeRecordByStatus{CollateralizeId: common.ToHex(collateralizeID), Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query collateralize by addr
	res, err = exec.Query("CollateralizeRecordByAddr",
		types.Encode(&pkt.ReqCollateralizeRecordByAddr{CollateralizeId: common.ToHex(collateralizeID), Addr: string(Nodes[1]), Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	res, err = exec.Query("CollateralizeRecordByAddr",
		types.Encode(&pkt.ReqCollateralizeRecordByAddr{Addr: string(Nodes[1]), Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query collateralize user balance
	res, err = exec.Query("CollateralizeUserBalance",
		types.Encode(&pkt.ReqCollateralizeRecordByAddr{Addr: string(Nodes[1]), Status: 1}))
	assert.Nil(t, err)
	assert.Equal(t, int64(100)*types.DefaultCoinPrecision, res.(*pkt.RepCollateralizeUserBalance).Balance)

	// collateralize append
	p5 := &pkt.CollateralizeAppendTx{
		CollateralizeID: common.ToHex(collateralizeID),
		RecordID:        common.ToHex(borrowID),
		Value:           100,
	}
	createTx, err = pkt.CreateRawCollateralizeAppendTx(env.cfg, p5)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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
	// query collateralize by id
	res, err = exec.Query("CollateralizeRecordByID",
		types.Encode(&pkt.ReqCollateralizeRecord{CollateralizeId: common.ToHex(collateralizeID), RecordId: common.ToHex(borrowID)}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query collateralize by status
	res, err = exec.Query("CollateralizeRecordByStatus",
		types.Encode(&pkt.ReqCollateralizeRecordByStatus{CollateralizeId: common.ToHex(collateralizeID), Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query collateralize by addr
	res, err = exec.Query("CollateralizeRecordByAddr",
		types.Encode(&pkt.ReqCollateralizeRecordByAddr{CollateralizeId: common.ToHex(collateralizeID), Addr: string(Nodes[1]), Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	// collateralize repay
	p6 := &pkt.CollateralizeRepayTx{
		CollateralizeID: common.ToHex(collateralizeID),
		RecordID:        common.ToHex(borrowID),
	}
	createTx, err = pkt.CreateRawCollateralizeRepayTx(env.cfg, p6)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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
	// query collateralize by status
	res, err = exec.Query("CollateralizeRecordByStatus",
		types.Encode(&pkt.ReqCollateralizeRecordByStatus{CollateralizeId: common.ToHex(collateralizeID), Status: 6}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query collateralize by addr
	res, err = exec.Query("CollateralizeRecordByAddr",
		types.Encode(&pkt.ReqCollateralizeRecordByAddr{CollateralizeId: common.ToHex(collateralizeID), Addr: string(Nodes[1]), Status: 6}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	res, err = exec.Query("CollateralizeRecordByAddr",
		types.Encode(&pkt.ReqCollateralizeRecordByAddr{Addr: string(Nodes[1]), Status: 6}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// query collateralize user balance
	res, err = exec.Query("CollateralizeUserBalance",
		types.Encode(&pkt.ReqCollateralizeRecordByAddr{Addr: string(Nodes[1]), Status: 1}))
	assert.Nil(t, err)
	assert.Equal(t, int64(0), res.(*pkt.RepCollateralizeUserBalance).Balance)

	// collateralize liquidate
	p7 := &pkt.CollateralizeBorrowTx{
		CollateralizeID: common.ToHex(collateralizeID),
		Value:           50,
	}
	createTx, err = pkt.CreateRawCollateralizeBorrowTx(env.cfg, p7)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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

	p71 := &pkt.CollateralizeBorrowTx{
		CollateralizeID: common.ToHex(collateralizeID),
		Value:           50,
	}
	createTx, err = pkt.CreateRawCollateralizeBorrowTx(env.cfg, p71)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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

	p8 := &pkt.CollateralizeFeedTx{}
	p8.Price = append(p8.Price, 0.28)
	p8.Volume = append(p8.Volume, 100)
	createTx, err = pkt.CreateRawCollateralizeFeedTx(env.cfg, p8)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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
	// query collateralize by status
	res, err = exec.Query("CollateralizeRecordByStatus",
		types.Encode(&pkt.ReqCollateralizeRecordByStatus{CollateralizeId: common.ToHex(collateralizeID), Status: 2}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	res, err = exec.Query("CollateralizeRecordByStatus",
		types.Encode(&pkt.ReqCollateralizeRecordByStatus{CollateralizeId: common.ToHex(collateralizeID), Status: 4}))
	assert.Nil(t, res)
	assert.NotNil(t, err)

	p81 := &pkt.CollateralizeFeedTx{}
	p81.Price = append(p81.Price, 0.5)
	p81.Volume = append(p81.Volume, 100)
	createTx, err = pkt.CreateRawCollateralizeFeedTx(env.cfg, p81)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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
	// query collateralize by status
	res, err = exec.Query("CollateralizeRecordByStatus",
		types.Encode(&pkt.ReqCollateralizeRecordByStatus{CollateralizeId: common.ToHex(collateralizeID), Status: 4}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	p9 := &pkt.CollateralizeFeedTx{}
	p9.Price = append(p9.Price, 0.25)
	p9.Volume = append(p9.Volume, 100)
	createTx, err = pkt.CreateRawCollateralizeFeedTx(env.cfg, p9)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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
	// query collateralize by status
	res, err = exec.Query("CollateralizeRecordByStatus",
		types.Encode(&pkt.ReqCollateralizeRecordByStatus{CollateralizeId: common.ToHex(collateralizeID), Status: 3}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	res, err = exec.Query("CollateralizeRecordByStatus",
		types.Encode(&pkt.ReqCollateralizeRecordByStatus{CollateralizeId: common.ToHex(collateralizeID), Status: 4}))
	assert.Nil(t, res)
	assert.NotNil(t, err)
	res, err = exec.Query("CollateralizeRecordByStatus",
		types.Encode(&pkt.ReqCollateralizeRecordByStatus{CollateralizeId: common.ToHex(collateralizeID), Status: 1}))
	assert.Nil(t, res)
	assert.NotNil(t, err)

	// expire liquidate
	p10 := &pkt.CollateralizeBorrowTx{
		CollateralizeID: common.ToHex(collateralizeID),
		Value:           50,
	}
	createTx, err = pkt.CreateRawCollateralizeBorrowTx(env.cfg, p10)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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

	p11 := &pkt.CollateralizeFeedTx{}
	p11.Price = append(p11.Price, 1)
	p11.Volume = append(p11.Volume, 100)
	createTx, err = pkt.CreateRawCollateralizeFeedTx(env.cfg, p11)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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
	// query collateralize by status
	res, err = exec.Query("CollateralizeRecordByStatus",
		types.Encode(&pkt.ReqCollateralizeRecordByStatus{CollateralizeId: common.ToHex(collateralizeID), Status: 5}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	// collateralize retrieve
	p12 := &pkt.CollateralizeRetrieveTx{
		CollateralizeID: common.ToHex(collateralizeID),
		Balance:         100,
	}
	createTx, err = pkt.CreateRawCollateralizeRetrieveTx(env.cfg, p12)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx.Execer = []byte(pkt.CollateralizeX)
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
	// query collateralize by status
	res, err = exec.Query("CollateralizeByStatus", types.Encode(&pkt.ReqCollateralizeByStatus{Status: 1}))
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName(pkt.CollateralizeX, signType), -1)
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
