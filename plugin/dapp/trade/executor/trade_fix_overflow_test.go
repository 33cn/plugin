// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"strings"
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
	"github.com/stretchr/testify/require"
)

func newTradeFixCfg() *types.Chain33Config {
	return types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
}

func newTradeFixDriver(cfg *types.Chain33Config, height int64, stateDB dbm.KV, localDB dbm.KVDB) drivers.Driver {
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	driver := newTrade()
	driver.SetAPI(api)
	driver.SetEnv(height, 1539918074, 1539918074)
	driver.SetStateDB(stateDB)
	driver.SetLocalDB(localDB)
	return driver
}

func applyTradeFixReceipt(stateDB dbm.KV, receipt *types.Receipt) {
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
}

func saveTradeFixExecAccount(t *testing.T, cfg *types.Chain33Config, kvdb dbm.KV, exec, symbol, addr string, balance int64) *account.DB {
	var acc *account.DB
	if exec == "coins" {
		acc = account.NewCoinsAccount(cfg)
		acc.SetDB(kvdb)
	} else {
		var err error
		acc, err = account.NewAccountDB(cfg, exec, symbol, kvdb)
		require.Nil(t, err)
	}
	acc.SaveExecAccount(address.ExecAddress("trade"), &types.Account{Balance: balance, Addr: addr})
	return acc
}

// sell order frozen amount TotalBoardlot*AmountPerBoardlot int64 overflow must be rejected after ForkTradeFixOverflow
func TestTradeFixSellOverflow(t *testing.T) {
	cfg := newTradeFixCfg()
	height := int64(10)
	execAddr := address.ExecAddress("trade")
	total := int64(10000000000)

	_, ldb, kvdb := util.CreateTestDB()
	defer ldb.Close()

	accToken := saveTradeFixExecAccount(t, cfg, kvdb, "token", "TEST", string(Nodes[0]), total)
	driver := newTradeFixDriver(cfg, height, kvdb, kvdb)

	// nominal total (2^32+1)*2^32 overflows int64 and wraps to 2^32
	sell := &pty.TradeSellTx{
		TokenSymbol:       "TEST",
		AssetExec:         "token",
		AmountPerBoardlot: 1 << 32,
		TotalBoardlot:     (1 << 32) + 1,
		MinBoardlot:       0,
		PricePerBoardlot:  2,
		PriceExec:         "coins",
		PriceSymbol:       "bty",
	}
	tx, err := pty.CreateRawTradeSellTx(cfg, sell)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	_, err = driver.Exec(tx, 0)
	assert.Equal(t, types.ErrInvalidParam, err, "overflow sell order must be rejected after fork")

	acc := accToken.LoadExecAccount(string(Nodes[0]), execAddr)
	assert.Equal(t, total, acc.Balance)
	assert.Equal(t, int64(0), acc.Frozen)

	// normal sell order is not affected
	sell = &pty.TradeSellTx{
		TokenSymbol:       "TEST",
		AssetExec:         "token",
		AmountPerBoardlot: 100,
		TotalBoardlot:     100,
		MinBoardlot:       1,
		PricePerBoardlot:  2,
		PriceExec:         "coins",
		PriceSymbol:       "bty",
	}
	tx, err = pty.CreateRawTradeSellTx(cfg, sell)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	receipt, err := driver.Exec(tx, 0)
	require.Nil(t, err, "normal sell order must succeed after fork")
	applyTradeFixReceipt(kvdb, receipt)

	acc = accToken.LoadExecAccount(string(Nodes[0]), execAddr)
	assert.Equal(t, total-100*100, acc.Balance)
	assert.Equal(t, int64(100*100), acc.Frozen)
}

// buy limit order frozen amount PricePerBoardlot*TotalBoardlot int64 overflow must be rejected after ForkTradeFixOverflow
func TestTradeFixBuyLimitOverflow(t *testing.T) {
	cfg := newTradeFixCfg()
	height := int64(10)
	execAddr := address.ExecAddress("trade")
	total := int64(10000000000)

	_, ldb, kvdb := util.CreateTestDB()
	defer ldb.Close()

	accBty := saveTradeFixExecAccount(t, cfg, kvdb, "coins", "bty", string(Nodes[0]), total)
	driver := newTradeFixDriver(cfg, height, kvdb, kvdb)

	// nominal cost 2^32*(2^32+1) overflows int64 and wraps to 2^32
	buy := &pty.TradeBuyLimitTx{
		TokenSymbol:       "TEST",
		AssetExec:         "paracross", // skip token exist check
		AmountPerBoardlot: 2,
		TotalBoardlot:     (1 << 32) + 1,
		MinBoardlot:       0,
		PricePerBoardlot:  1 << 32,
		PriceExec:         "coins",
		PriceSymbol:       "bty",
	}
	tx, err := pty.CreateRawTradeBuyLimitTx(cfg, buy)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	_, err = driver.Exec(tx, 0)
	assert.Equal(t, types.ErrInvalidParam, err, "overflow buy limit order must be rejected after fork")

	acc := accBty.LoadExecAccount(string(Nodes[0]), execAddr)
	assert.Equal(t, total, acc.Balance)
	assert.Equal(t, int64(0), acc.Frozen)

	// normal buy limit order is not affected
	buy = &pty.TradeBuyLimitTx{
		TokenSymbol:       "TEST",
		AssetExec:         "paracross",
		AmountPerBoardlot: 2,
		TotalBoardlot:     100,
		MinBoardlot:       1,
		PricePerBoardlot:  100,
		PriceExec:         "coins",
		PriceSymbol:       "bty",
	}
	tx, err = pty.CreateRawTradeBuyLimitTx(cfg, buy)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	receipt, err := driver.Exec(tx, 0)
	require.Nil(t, err, "normal buy limit order must succeed after fork")
	applyTradeFixReceipt(kvdb, receipt)

	acc = accBty.LoadExecAccount(string(Nodes[0]), execAddr)
	assert.Equal(t, total-100*100, acc.Balance)
	assert.Equal(t, int64(100*100), acc.Frozen)
}

// checkPrice/notSameAsset were wrongly called with PriceExec as the symbol argument,
// after ForkTradeFixOverflow PriceSymbol must be validated instead
func TestTradeFixPriceSymbolArg(t *testing.T) {
	cfg := newTradeFixCfg()
	height := int64(10)
	total := int64(10000000000)

	_, ldb, kvdb := util.CreateTestDB()
	defer ldb.Close()

	saveTradeFixExecAccount(t, cfg, kvdb, "coins", "bty", string(Nodes[0]), total)
	driver := newTradeFixDriver(cfg, height, kvdb, kvdb)

	// same asset order: sell bty priced in bty, must be rejected by notSameAsset
	sell := &pty.TradeSellTx{
		TokenSymbol:       "bty",
		AssetExec:         "coins",
		AmountPerBoardlot: 100,
		TotalBoardlot:     100,
		MinBoardlot:       1,
		PricePerBoardlot:  1,
		PriceExec:         "coins",
		PriceSymbol:       "bty",
	}
	tx, err := pty.CreateRawTradeSellTx(cfg, sell)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	_, err = driver.Exec(tx, 0)
	assert.Equal(t, pty.ErrAssetAndPriceSame, err, "same asset sell order must be rejected after fork")

	// price exec set but price symbol missing, must be rejected by checkPrice
	sell = &pty.TradeSellTx{
		TokenSymbol:       "bty",
		AssetExec:         "coins",
		AmountPerBoardlot: 100,
		TotalBoardlot:     100,
		MinBoardlot:       1,
		PricePerBoardlot:  1,
		PriceExec:         "coins",
		PriceSymbol:       "",
	}
	tx, err = pty.CreateRawTradeSellTx(cfg, sell)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	_, err = driver.Exec(tx, 0)
	assert.Equal(t, types.ErrInvalidParam, err, "order with empty price symbol must be rejected after fork")

	// same asset buy limit order must be rejected as well
	buy := &pty.TradeBuyLimitTx{
		TokenSymbol:       "bty",
		AssetExec:         "coins",
		AmountPerBoardlot: 2,
		TotalBoardlot:     100,
		MinBoardlot:       1,
		PricePerBoardlot:  100,
		PriceExec:         "coins",
		PriceSymbol:       "bty",
	}
	tx, err = pty.CreateRawTradeBuyLimitTx(cfg, buy)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	_, err = driver.Exec(tx, 0)
	assert.Equal(t, pty.ErrAssetAndPriceSame, err, "same asset buy limit order must be rejected after fork")
}

// before ForkTradeFixOverflow the old vulnerable behavior must be kept for consensus
func TestTradeFixPreForkBehavior(t *testing.T) {
	cfg := newTradeFixCfg()
	cfg.SetDappFork(pty.TradeX, pty.ForkTradeFixOverflowX, int64(100))
	height := int64(99)
	execAddr := address.ExecAddress("trade")
	total := int64(10000000000)

	_, ldb, kvdb := util.CreateTestDB()
	defer ldb.Close()

	accToken := saveTradeFixExecAccount(t, cfg, kvdb, "token", "TEST", string(Nodes[0]), total)
	saveTradeFixExecAccount(t, cfg, kvdb, "coins", "bty", string(Nodes[0]), total)
	driver := newTradeFixDriver(cfg, height, kvdb, kvdb)

	// overflow sell order is accepted before the fork, frozen amount wraps to 2^32
	sell := &pty.TradeSellTx{
		TokenSymbol:       "TEST",
		AssetExec:         "token",
		AmountPerBoardlot: 1 << 32,
		TotalBoardlot:     (1 << 32) + 1,
		MinBoardlot:       0,
		PricePerBoardlot:  2,
		PriceExec:         "coins",
		PriceSymbol:       "bty",
	}
	tx, err := pty.CreateRawTradeSellTx(cfg, sell)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	receipt, err := driver.Exec(tx, 0)
	require.Nil(t, err, "overflow sell order keeps old behavior before fork")
	applyTradeFixReceipt(kvdb, receipt)

	acc := accToken.LoadExecAccount(string(Nodes[0]), execAddr)
	assert.Equal(t, int64(1<<32), acc.Frozen, "frozen amount wraps to 2^32 before fork")

	// same asset order is accepted before the fork
	sell = &pty.TradeSellTx{
		TokenSymbol:       "bty",
		AssetExec:         "coins",
		AmountPerBoardlot: 100,
		TotalBoardlot:     100,
		MinBoardlot:       1,
		PricePerBoardlot:  1,
		PriceExec:         "coins",
		PriceSymbol:       "bty",
	}
	tx, err = pty.CreateRawTradeSellTx(cfg, sell)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	_, err = driver.Exec(tx, 0)
	require.Nil(t, err, "same asset sell order keeps old behavior before fork")
}

// normal sell/buy/revoke flow must not be affected by ForkTradeFixOverflow
func TestTradeFixNormalFlow(t *testing.T) {
	cfg := newTradeFixCfg()
	height := int64(10)
	execAddr := address.ExecAddress("trade")
	total := int64(100000)

	_, ldb, kvdb := util.CreateTestDB()
	defer ldb.Close()

	accToken := saveTradeFixExecAccount(t, cfg, kvdb, "token", "TEST", string(Nodes[0]), total)
	accBty := saveTradeFixExecAccount(t, cfg, kvdb, "coins", "bty", string(Nodes[1]), total)
	driver := newTradeFixDriver(cfg, height, kvdb, kvdb)

	sell := &pty.TradeSellTx{
		TokenSymbol:       "TEST",
		AssetExec:         "token",
		AmountPerBoardlot: 100,
		TotalBoardlot:     100,
		MinBoardlot:       1,
		PricePerBoardlot:  2,
		PriceExec:         "coins",
		PriceSymbol:       "bty",
	}
	tx, err := pty.CreateRawTradeSellTx(cfg, sell)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	receipt, err := driver.Exec(tx, 0)
	require.Nil(t, err)
	applyTradeFixReceipt(kvdb, receipt)

	var sellOrder pty.SellOrder
	require.Nil(t, types.Decode(receipt.KV[len(receipt.KV)-1].Value, &sellOrder))
	sellID := sellOrder.SellID[len("mavl-trade-sell-"):]

	// buy 5 boardlots
	buy := &pty.TradeBuyTx{SellID: sellID, BoardlotCnt: 5}
	tx, err = pty.CreateRawTradeBuyTx(cfg, buy)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyB)
	require.Nil(t, err)
	receipt, err = driver.Exec(tx, 0)
	require.Nil(t, err, "buy must succeed after fork")
	applyTradeFixReceipt(kvdb, receipt)

	require.Nil(t, types.Decode(receipt.KV[len(receipt.KV)-1].Value, &sellOrder))
	assert.Equal(t, int64(5), sellOrder.SoldBoardlot)

	// revoke the remaining order
	revoke := &pty.TradeRevokeTx{SellID: sellID}
	tx, err = pty.CreateRawTradeRevokeTx(cfg, revoke)
	require.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	require.Nil(t, err)
	receipt, err = driver.Exec(tx, 0)
	require.Nil(t, err, "revoke must succeed after fork")
	applyTradeFixReceipt(kvdb, receipt)

	acc := accToken.LoadExecAccount(string(Nodes[0]), execAddr)
	assert.Equal(t, total-5*100, acc.Balance)
	assert.Equal(t, int64(0), acc.Frozen)
	assert.Equal(t, total-5*2, accBty.LoadExecAccount(string(Nodes[1]), execAddr).Balance)
}
