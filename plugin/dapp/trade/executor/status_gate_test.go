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
}

func newStatusGateEnv(t *testing.T) *statusGateEnv {
	t.Helper()
	stateDB, err := dbm.NewGoMemDB("status-gate", "status-gate", 100)
	assert.Nil(t, err)
	_, ldb, kvdb := util.CreateTestDB()
	t.Cleanup(func() { ldb.Close() })

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(chain33TestCfg, nil)
	driver := newTrade()
	driver.SetAPI(api)
	driver.SetEnv(0, gateTestBlockTime, 2)
	driver.SetStateDB(stateDB)
	driver.SetLocalDB(kvdb)
	return &statusGateEnv{driver: driver, stateDB: stateDB}
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

func normalizeAccount(acc *types.Account) *types.Account {
	// LoadExecAccount returns a zero-valued Frozen for accounts that were never
	// written; drop only the fields the test does not care about.
	return &types.Account{Addr: acc.Addr, Balance: acc.Balance, Frozen: acc.Frozen}
}
