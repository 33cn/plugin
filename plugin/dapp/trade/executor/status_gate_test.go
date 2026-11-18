// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const gateTestBlockTime = int64(1539918074)

type statusGateEnv struct {
	driver  drivers.Driver
	stateDB dbm.KV
	localDB dbm.KVDB
}

func newStatusGateEnv(t *testing.T) *statusGateEnv {
	t.Helper()
	return newStatusGateEnvAt(t, 0)
}

func newStatusGateEnvAt(t *testing.T, height int64) *statusGateEnv {
	t.Helper()
	stateDB, err := dbm.NewGoMemDB("status-gate", "status-gate", 100)
	assert.Nil(t, err)
	_, ldb, kvdb := util.CreateTestDB()
	t.Cleanup(func() { ldb.Close() })
	return &statusGateEnv{
		driver:  newStatusGateDriver(t, stateDB, kvdb, height),
		stateDB: stateDB,
		localDB: kvdb,
	}
}

// atHeight returns an env at another block height on the same state, so a test
// can create an order before a fork and fill it after the fork activates.
func (env *statusGateEnv) atHeight(t *testing.T, height int64) *statusGateEnv {
	t.Helper()
	return &statusGateEnv{
		driver:  newStatusGateDriver(t, env.stateDB, env.localDB, height),
		stateDB: env.stateDB,
		localDB: env.localDB,
	}
}

func newStatusGateDriver(t *testing.T, stateDB dbm.KV, localDB dbm.KVDB, height int64) drivers.Driver {
	t.Helper()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(chain33TestCfg, nil)
	driver := newTrade()
	driver.SetAPI(api)
	driver.SetEnv(height, gateTestBlockTime, 2)
	driver.SetStateDB(stateDB)
	driver.SetLocalDB(localDB)
	return driver
}

// setTradeForks pins the trade forks a status gate test depends on. A gate
// height above the block height keeps the gate off, so the same scenario also
// pins the pre-fork behaviour.
func setTradeForks(t *testing.T, gateHeight int64) {
	t.Helper()
	forks := map[string]int64{
		pty.ForkTradeAssetX:      0,
		pty.ForkTradeIDX:         0,
		pty.ForkTradeFixAssetDBX: 0,
		pty.ForkTradePriceX:      0,
		pty.ForkTradeStatusGateX: gateHeight,
	}
	for fork, height := range forks {
		fork, height := fork, height
		old := chain33TestCfg.GetDappFork(pty.TradeX, fork)
		chain33TestCfg.SetDappFork(pty.TradeX, fork, height)
		t.Cleanup(func() { chain33TestCfg.SetDappFork(pty.TradeX, fork, old) })
	}
}

func seedExecAccount(t *testing.T, exec, symbol, addr string, balance, frozen int64, db dbm.KV) {
	t.Helper()
	accDB, err := account.NewAccountDB(chain33TestCfg, exec, symbol, db)
	assert.Nil(t, err)
	accDB.SaveExecAccount(address.ExecAddress("trade"), &types.Account{Addr: addr, Balance: balance, Frozen: frozen})
}

func loadExecAccount(t *testing.T, exec, symbol, addr string, db dbm.KV) *types.Account {
	t.Helper()
	accDB, err := account.NewAccountDB(chain33TestCfg, exec, symbol, db)
	assert.Nil(t, err)
	return accDB.LoadExecAccount(addr, address.ExecAddress("trade"))
}

func execBuyLimit(t *testing.T, env *statusGateEnv, price, total int64) string {
	t.Helper()
	buy := &pty.TradeBuyLimitTx{
		TokenSymbol:       Symbol,
		AmountPerBoardlot: 100,
		MinBoardlot:       2,
		PricePerBoardlot:  price,
		TotalBoardlot:     total,
		AssetExec:         AssetExecPara,
		PriceExec:         AssetExecToken,
		PriceSymbol:       SymbolA,
	}
	tx, err := pty.CreateRawTradeBuyLimitTx(chain33TestCfg, buy)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	receipt, err := env.driver.Exec(tx, 0)
	if err != nil {
		t.Fatalf("buy limit exec failed: %v", err)
	}
	var order pty.BuyLimitOrder
	assert.Nil(t, types.Decode(receipt.KV[1].Value, &order))
	return order.BuyID[len(buyIDPrefix):]
}

func execRevokeBuy(t *testing.T, env *statusGateEnv, buyID string) {
	t.Helper()
	tx, err := pty.CreateRawTradeRevokeBuyTx(chain33TestCfg, &pty.TradeRevokeBuyTx{BuyID: buyID})
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	_, err = env.driver.Exec(tx, 0)
	if err != nil {
		t.Fatalf("revoke buy exec failed: %v", err)
	}
}

func execSellMarket(t *testing.T, env *statusGateEnv, buyID string, cnt int64) (*types.Receipt, error) {
	t.Helper()
	tx, err := pty.CreateRawTradeSellMarketTx(chain33TestCfg, &pty.TradeSellMarketTx{BuyID: buyID, BoardlotCnt: cnt})
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	return env.driver.Exec(tx, 0)
}

// setupRevokedBuyOrder seeds buyer B's price account with enough frozen funds
// for a second live buy order, then revokes the first order. The revoked order's
// own collateral has been refunded; the frozen funds left belong to the live one.
func setupRevokedBuyOrder(t *testing.T, env *statusGateEnv) (revokedID, liveID string) {
	t.Helper()
	seedExecAccount(t, AssetExecToken, SymbolA, string(Nodes[1]), 100000, 0, env.stateDB)
	seedExecAccount(t, AssetExecPara, Symbol, string(Nodes[0]), 100000, 0, env.stateDB)

	revokedID = execBuyLimit(t, env, 2, 100)
	liveID = execBuyLimit(t, env, 2, 200)
	execRevokeBuy(t, env, revokedID)
	return revokedID, liveID
}

func loadBuyOrder(t *testing.T, env *statusGateEnv, buyID string) *pty.BuyLimitOrder {
	t.Helper()
	order, err := getBuyOrderFromID([]byte(buyIDPrefix+buyID), env.stateDB)
	assert.Nil(t, err)
	return order
}

func TestTradeSellMarketRevokedBuyOrderForkActive(t *testing.T) {
	setTradeForks(t, 0)
	env := newStatusGateEnv(t)
	revokedID, _ := setupRevokedBuyOrder(t, env)

	// B: 100000 - 200 (revoked order) - 400 (live order) + 200 (revoke refund)
	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 99600, Frozen: 400},
		normalizeAccount(loadExecAccount(t, AssetExecToken, SymbolA, string(Nodes[1]), env.stateDB)))

	receipt, err := execSellMarket(t, env, revokedID, 5)
	assert.Nil(t, receipt)
	assert.Equal(t, pty.ErrTBuyOrderRevoked, err)

	// The rejected fill must not move any funds.
	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 99600, Frozen: 400},
		normalizeAccount(loadExecAccount(t, AssetExecToken, SymbolA, string(Nodes[1]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[0])},
		normalizeAccount(loadExecAccount(t, AssetExecToken, SymbolA, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 100000},
		normalizeAccount(loadExecAccount(t, AssetExecPara, Symbol, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[1])},
		normalizeAccount(loadExecAccount(t, AssetExecPara, Symbol, string(Nodes[1]), env.stateDB)))

	order := loadBuyOrder(t, env, revokedID)
	assert.Equal(t, int32(pty.TradeOrderStatusBuyRevoked), order.Status)
	assert.Equal(t, int64(0), order.BoughtBoardlot)
}

func TestTradeSellMarketRevokedBuyOrderForkInactive(t *testing.T) {
	setTradeForks(t, 100) // gate stays off at block height 0
	env := newStatusGateEnv(t)
	revokedID, _ := setupRevokedBuyOrder(t, env)

	receipt, err := execSellMarket(t, env, revokedID, 5)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)

	// The revoked order is filled from the live order's collateral: pre-fork
	// behaviour stays byte-for-byte the same.
	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 99600, Frozen: 390},
		normalizeAccount(loadExecAccount(t, AssetExecToken, SymbolA, string(Nodes[1]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 10},
		normalizeAccount(loadExecAccount(t, AssetExecToken, SymbolA, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 99500},
		normalizeAccount(loadExecAccount(t, AssetExecPara, Symbol, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 500},
		normalizeAccount(loadExecAccount(t, AssetExecPara, Symbol, string(Nodes[1]), env.stateDB)))

	order := loadBuyOrder(t, env, revokedID)
	assert.Equal(t, int32(pty.TradeOrderStatusBuyRevoked), order.Status)
	assert.Equal(t, int64(5), order.BoughtBoardlot)
}

func TestTradeSellMarketLiveBuyOrderForkActive(t *testing.T) {
	setTradeForks(t, 0)
	env := newStatusGateEnv(t)
	seedExecAccount(t, AssetExecToken, SymbolA, string(Nodes[1]), 100000, 0, env.stateDB)
	seedExecAccount(t, AssetExecPara, Symbol, string(Nodes[0]), 100000, 0, env.stateDB)
	buyID := execBuyLimit(t, env, 2, 100)

	receipt, err := execSellMarket(t, env, buyID, 5)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)

	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 99800, Frozen: 190},
		normalizeAccount(loadExecAccount(t, AssetExecToken, SymbolA, string(Nodes[1]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 10},
		normalizeAccount(loadExecAccount(t, AssetExecToken, SymbolA, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 99500},
		normalizeAccount(loadExecAccount(t, AssetExecPara, Symbol, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 500},
		normalizeAccount(loadExecAccount(t, AssetExecPara, Symbol, string(Nodes[1]), env.stateDB)))

	order := loadBuyOrder(t, env, buyID)
	assert.Equal(t, int32(pty.TradeOrderStatusOnBuy), order.Status)
	assert.Equal(t, int64(5), order.BoughtBoardlot)
}

func execSellLimit(t *testing.T, env *statusGateEnv, starttime, totalBoardlot int64) string {
	t.Helper()
	sell := &pty.Trade{
		Ty: pty.TradeSellLimit,
		Value: &pty.Trade_SellLimit{SellLimit: &pty.TradeForSell{
			TokenSymbol:       Symbol,
			AmountPerBoardlot: 100,
			MinBoardlot:       2,
			PricePerBoardlot:  2,
			TotalBoardlot:     totalBoardlot,
			Starttime:         starttime,
			AssetExec:         AssetExecToken,
			PriceExec:         AssetExecPara,
			PriceSymbol:       SymbolA,
		}},
	}
	tx, err := types.CreateFormatTx(chain33TestCfg, chain33TestCfg.ExecName(pty.TradeX), types.Encode(sell))
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	receipt, err := env.driver.Exec(tx, 0)
	if err != nil {
		t.Fatalf("sell limit exec failed: %v", err)
	}
	var order pty.SellOrder
	assert.Nil(t, types.Decode(receipt.KV[1].Value, &order))
	return order.SellID[len(sellIDPrefix):]
}

func execRevokeSell(t *testing.T, env *statusGateEnv, sellID string) error {
	t.Helper()
	tx, err := pty.CreateRawTradeRevokeTx(chain33TestCfg, &pty.TradeRevokeTx{SellID: sellID})
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	_, err = env.driver.Exec(tx, 0)
	return err
}

func execBuySellOrder(t *testing.T, env *statusGateEnv, sellID string, cnt int64) (*types.Receipt, error) {
	t.Helper()
	tx, err := pty.CreateRawTradeBuyTx(chain33TestCfg, &pty.TradeBuyTx{SellID: sellID, BoardlotCnt: cnt})
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	return env.driver.Exec(tx, 0)
}

func loadSellOrder(t *testing.T, env *statusGateEnv, sellID string) *pty.SellOrder {
	t.Helper()
	order, err := getSellOrderFromID([]byte(sellIDPrefix+sellID), env.stateDB)
	assert.Nil(t, err)
	return order
}

func TestTradeRevokeScheduledSellOrderForkActive(t *testing.T) {
	setTradeForks(t, 0)
	env := newStatusGateEnv(t)
	seedExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), 100000, 0, env.stateDB)

	sellID := execSellLimit(t, env, 1, 10)
	assert.Equal(t, int32(pty.TradeOrderStatusNotStart), loadSellOrder(t, env, sellID).Status)

	assert.Nil(t, execRevokeSell(t, env, sellID))

	order := loadSellOrder(t, env, sellID)
	assert.Equal(t, int32(pty.TradeOrderStatusRevoked), order.Status)
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 100000},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))
}

func TestTradeRevokeScheduledSellOrderForkInactive(t *testing.T) {
	setTradeForks(t, 100) // gate stays off at block height 0
	env := newStatusGateEnv(t)
	seedExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), 100000, 0, env.stateDB)

	sellID := execSellLimit(t, env, 1, 10)
	assert.Nil(t, execRevokeSell(t, env, sellID))

	// Pre-fork save wrapper still resets a scheduled order to NotStart.
	order := loadSellOrder(t, env, sellID)
	assert.Equal(t, int32(pty.TradeOrderStatusNotStart), order.Status)
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 100000},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))
}

func TestTradeBuyRevokedScheduledSellOrderForkActive(t *testing.T) {
	setTradeForks(t, 0)
	env := newStatusGateEnv(t)
	seedExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), 100000, 0, env.stateDB)
	seedExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), 100000, 0, env.stateDB)

	sellID := execSellLimit(t, env, 1, 10)
	assert.Nil(t, execRevokeSell(t, env, sellID))

	receipt, err := execBuySellOrder(t, env, sellID, 5)
	assert.Nil(t, receipt)
	assert.Equal(t, pty.ErrTSellOrderRevoked, err)

	// The rejected fill must not move any funds.
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 100000},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 100000},
		normalizeAccount(loadExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), env.stateDB)))

	order := loadSellOrder(t, env, sellID)
	assert.Equal(t, int32(pty.TradeOrderStatusRevoked), order.Status)
	assert.Equal(t, int64(0), order.SoldBoardlot)
}

// A scheduled order is seeded with extra frozen tokens standing for another
// live order. A fill larger than the remainder must not reach that shared pool.
func TestTradeBuyScheduledSellOrderNotEnoughForkActive(t *testing.T) {
	setTradeForks(t, 0)
	env := newStatusGateEnv(t)
	seedExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), 100000, 5000, env.stateDB)
	seedExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), 100000, 0, env.stateDB)

	sellID := execSellLimit(t, env, 1, 10)
	// This order freezes 10 * 100 on top of the 5000 already frozen.
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 99000, Frozen: 6000},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))

	receipt, err := execBuySellOrder(t, env, sellID, 8)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 99000, Frozen: 5200},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))

	receipt, err = execBuySellOrder(t, env, sellID, 5) // only 2 boardlot remain
	assert.Nil(t, receipt)
	assert.Equal(t, pty.ErrTSellOrderNotEnough, err)
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 99000, Frozen: 5200},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))

	receipt, err = execBuySellOrder(t, env, sellID, 1) // below MinBoardlot
	assert.Nil(t, receipt)
	assert.Equal(t, pty.ErrTCntLessThanMinBoardlot, err)
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 99000, Frozen: 5200},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))

	order := loadSellOrder(t, env, sellID)
	assert.Equal(t, int32(pty.TradeOrderStatusNotStart), order.Status)
	assert.Equal(t, int64(8), order.SoldBoardlot)
}

func TestTradeBuyScheduledSellOrderSoldOutForkActive(t *testing.T) {
	setTradeForks(t, 0)
	env := newStatusGateEnv(t)
	seedExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), 100000, 5000, env.stateDB)
	seedExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), 100000, 0, env.stateDB)

	sellID := execSellLimit(t, env, 1, 10)

	receipt, err := execBuySellOrder(t, env, sellID, 10)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)

	order := loadSellOrder(t, env, sellID)
	assert.Equal(t, int32(pty.TradeOrderStatusSoldOut), order.Status)
	assert.Equal(t, int64(10), order.SoldBoardlot)
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 99000, Frozen: 5000},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))

	receipt, err = execBuySellOrder(t, env, sellID, 2)
	assert.Nil(t, receipt)
	assert.Equal(t, pty.ErrTSellOrderSoldout, err)
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 99000, Frozen: 5000},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))
}

func TestTradeBuyOnSaleSellOrderForkActive(t *testing.T) {
	setTradeForks(t, 0)
	env := newStatusGateEnv(t)
	seedExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), 100000, 0, env.stateDB)
	seedExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), 100000, 0, env.stateDB)

	sellID := execSellLimit(t, env, pty.InvalidStartTime, 10)
	assert.Equal(t, int32(pty.TradeOrderStatusOnSale), loadSellOrder(t, env, sellID).Status)

	receipt, err := execBuySellOrder(t, env, sellID, 4)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)

	order := loadSellOrder(t, env, sellID)
	assert.Equal(t, int32(pty.TradeOrderStatusOnSale), order.Status)
	assert.Equal(t, int64(4), order.SoldBoardlot)
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 99000, Frozen: 600},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 99992},
		normalizeAccount(loadExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), env.stateDB)))
}

// A scheduled order revoked before the fork was saved as NotStart and its
// collateral refunded, so it is indistinguishable from a live one. After the
// fork no fill may reach it: the seller keeps the collateral of the other
// orders sharing the frozen pool.
func TestTradeBuyScheduledSellOrderRevokedBeforeForkForkActive(t *testing.T) {
	setTradeForks(t, 10)
	env := newStatusGateEnvAt(t, 0)
	seedExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), 100000, 5000, env.stateDB)
	seedExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), 100000, 0, env.stateDB)

	sellID := execSellLimit(t, env, 1, 10)
	assert.Nil(t, execRevokeSell(t, env, sellID))
	order := loadSellOrder(t, env, sellID)
	assert.Equal(t, int32(pty.TradeOrderStatusNotStart), order.Status)
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 100000, Frozen: 5000},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))

	atFork := env.atHeight(t, 10)
	receipt, err := execBuySellOrder(t, atFork, sellID, 4)
	assert.Nil(t, receipt)
	assert.Equal(t, pty.ErrTSellOrderScheduledBeforeFork, err)

	// The rejected fill must not move any funds.
	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 100000, Frozen: 5000},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 100000},
		normalizeAccount(loadExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), env.stateDB)))

	order = loadSellOrder(t, env, sellID)
	assert.Equal(t, int32(pty.TradeOrderStatusNotStart), order.Status)
	assert.Equal(t, int64(0), order.SoldBoardlot)
}

// Pre-fork behaviour stays byte-for-byte the same: the stale row is still
// fillable and pays out of the seller's shared frozen pool.
func TestTradeBuyScheduledSellOrderRevokedBeforeForkForkInactive(t *testing.T) {
	setTradeForks(t, 100) // gate stays off at block height 0
	env := newStatusGateEnv(t)
	seedExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), 100000, 5000, env.stateDB)
	seedExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), 100000, 0, env.stateDB)

	sellID := execSellLimit(t, env, 1, 10)
	assert.Nil(t, execRevokeSell(t, env, sellID))

	receipt, err := execBuySellOrder(t, env, sellID, 4)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)

	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 100000, Frozen: 4600},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 99992},
		normalizeAccount(loadExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), env.stateDB)))

	order := loadSellOrder(t, env, sellID)
	assert.Equal(t, int32(pty.TradeOrderStatusNotStart), order.Status)
	assert.Equal(t, int64(4), order.SoldBoardlot)
}

func TestTradeBuyScheduledSellOrderCreatedAfterForkForkActive(t *testing.T) {
	setTradeForks(t, 10)
	env := newStatusGateEnvAt(t, 10)
	seedExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), 100000, 0, env.stateDB)
	seedExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), 100000, 0, env.stateDB)

	sellID := execSellLimit(t, env, 1, 10)
	assert.Equal(t, int64(10), loadSellOrder(t, env, sellID).Height)

	receipt, err := execBuySellOrder(t, env, sellID, 4)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)

	assert.Equal(t, &types.Account{Addr: string(Nodes[0]), Balance: 99000, Frozen: 600},
		normalizeAccount(loadExecAccount(t, AssetExecToken, Symbol, string(Nodes[0]), env.stateDB)))
	assert.Equal(t, &types.Account{Addr: string(Nodes[1]), Balance: 99992},
		normalizeAccount(loadExecAccount(t, AssetExecPara, SymbolA, string(Nodes[1]), env.stateDB)))

	order := loadSellOrder(t, env, sellID)
	assert.Equal(t, int32(pty.TradeOrderStatusNotStart), order.Status)
	assert.Equal(t, int64(4), order.SoldBoardlot)
}

func normalizeAccount(acc *types.Account) *types.Account {
	// LoadExecAccount returns a zero-valued Frozen for accounts that were never
	// written; drop only the fields the test does not care about.
	return &types.Account{Addr: acc.Addr, Balance: acc.Balance, Frozen: acc.Frozen}
}
