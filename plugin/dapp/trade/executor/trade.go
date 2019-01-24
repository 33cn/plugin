// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

/*
trade执行器支持trade的创建和交易，

主要提供操作有以下几种：
1）挂单出售；
2）购买指定的卖单；
3）撤销卖单；
4）挂单购买；
5）出售指定的买单；
6）撤销买单；
*/

import (
	log "github.com/33cn/chain33/common/log/log15"

	"github.com/33cn/chain33/common/db/table"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

var (
	tradelog         = log.New("module", "execs.trade")
	defaultAssetExec = "token"
	driverName       = "trade"
)

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&trade{}))
}

// Init : 注册当前trade合约
func Init(name string, sub []byte) {
	drivers.Register(GetName(), newTrade, types.GetDappFork(driverName, "Enable"))
}

// GetName : 获取trade合约名字
func GetName() string {
	return newTrade().GetName()
}

type trade struct {
	drivers.DriverBase
}

func newTrade() drivers.Driver {
	t := &trade{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

func (t *trade) GetDriverName() string {
	return driverName
}

func (t *trade) getSellOrderFromDb(sellID []byte) *pty.SellOrder {
	value, err := t.GetStateDB().Get(sellID)
	if err != nil {
		panic(err)
	}
	var sellorder pty.SellOrder
	types.Decode(value, &sellorder)
	return &sellorder
}

func genSaveSellKv(sellorder *pty.SellOrder) []*types.KeyValue {
	status := sellorder.Status
	var kv []*types.KeyValue
	kv = saveSellOrderKeyValue(kv, sellorder, status)
	if pty.TradeOrderStatusSoldOut == status || pty.TradeOrderStatusRevoked == status {
		tradelog.Debug("trade saveSell ", "remove old status onsale to soldout or revoked with sellid", sellorder.SellID)
		kv = deleteSellOrderKeyValue(kv, sellorder, pty.TradeOrderStatusOnSale)
	}
	return kv
}

func (t *trade) saveSell(base *pty.ReceiptSellBase, ty int32, tx *types.Transaction, txIndex string, ldb *table.Table) []*types.KeyValue {
	sellorder := t.getSellOrderFromDb([]byte(base.SellID))

	if ty == pty.TyLogTradeSellLimit && sellorder.SoldBoardlot == 0 {
		newOrder := t.genSellLimit(tx, base, sellorder, txIndex)
		tradelog.Info("Table", "sell-add", newOrder)
		ldb.Add(newOrder)
	} else {
		t.updateSellLimit(tx, base, sellorder, txIndex, ldb)
	}
	return genSaveSellKv(sellorder)
}

func deleteSellOrderKeyValue(kv []*types.KeyValue, sellorder *pty.SellOrder, status int32) []*types.KeyValue {
	return genSellOrderKeyValue(kv, sellorder, status, nil)
}

func saveSellOrderKeyValue(kv []*types.KeyValue, sellorder *pty.SellOrder, status int32) []*types.KeyValue {
	sellID := []byte(sellorder.SellID)
	return genSellOrderKeyValue(kv, sellorder, status, sellID)
}

func genDeleteSellKv(sellorder *pty.SellOrder) []*types.KeyValue {
	status := sellorder.Status
	var kv []*types.KeyValue
	kv = deleteSellOrderKeyValue(kv, sellorder, status)
	if pty.TradeOrderStatusSoldOut == status || pty.TradeOrderStatusRevoked == status {
		tradelog.Debug("trade saveSell ", "remove old status onsale to soldout or revoked with sellID", sellorder.SellID)
		kv = saveSellOrderKeyValue(kv, sellorder, pty.TradeOrderStatusOnSale)
	}
	return kv
}

func (t *trade) deleteSell(base *pty.ReceiptSellBase, ty int32, tx *types.Transaction, txIndex string, ldb *table.Table, tradedBoardlot int64) []*types.KeyValue {
	sellorder := t.getSellOrderFromDb([]byte(base.SellID))
	if ty == pty.TyLogTradeSellLimit && sellorder.SoldBoardlot == 0 {
		ldb.Del([]byte(txIndex))
	} else {
		t.rollBackSellLimit(tx, base, sellorder, txIndex, ldb, tradedBoardlot)
	}
	return genDeleteSellKv(sellorder)
}

func (t *trade) saveBuy(receiptTradeBuy *pty.ReceiptBuyBase, tx *types.Transaction, txIndex string, ldb *table.Table) []*types.KeyValue {
	//tradelog.Info("save", "buy", receiptTradeBuy)

	var kv []*types.KeyValue
	order := t.genBuyMarket(tx, receiptTradeBuy, txIndex)
	tradelog.Debug("trade BuyMarket save local", "order", order)
	ldb.Add(order)
	return saveBuyMarketOrderKeyValue(kv, receiptTradeBuy, pty.TradeOrderStatusBoughtOut, t.GetHeight())
}

func (t *trade) deleteBuy(receiptTradeBuy *pty.ReceiptBuyBase, txIndex string, ldb *table.Table) []*types.KeyValue {
	var kv []*types.KeyValue
	ldb.Del([]byte(txIndex))
	return deleteBuyMarketOrderKeyValue(kv, receiptTradeBuy, pty.TradeOrderStatusBoughtOut, t.GetHeight())
}

// BuyLimit Local
func (t *trade) getBuyOrderFromDb(buyID []byte) *pty.BuyLimitOrder {
	value, err := t.GetStateDB().Get(buyID)
	if err != nil {
		panic(err)
	}
	var buyOrder pty.BuyLimitOrder
	types.Decode(value, &buyOrder)
	return &buyOrder
}

func genSaveBuyLimitKv(buyOrder *pty.BuyLimitOrder) []*types.KeyValue {
	status := buyOrder.Status
	var kv []*types.KeyValue
	kv = saveBuyLimitOrderKeyValue(kv, buyOrder, status)
	if pty.TradeOrderStatusBoughtOut == status || pty.TradeOrderStatusBuyRevoked == status {
		tradelog.Debug("trade saveBuyLimit ", "remove old status with Buyid", buyOrder.BuyID)
		kv = deleteBuyLimitKeyValue(kv, buyOrder, pty.TradeOrderStatusOnBuy)
	}
	return kv
}

func (t *trade) saveBuyLimit(buy *pty.ReceiptBuyBase, ty int32, tx *types.Transaction, txIndex string, ldb *table.Table) []*types.KeyValue {
	buyOrder := t.getBuyOrderFromDb([]byte(buy.BuyID))
	tradelog.Debug("Table", "buy-add", buyOrder)
	if buyOrder.Status == pty.TradeOrderStatusOnBuy && buy.BoughtBoardlot == 0 {
		order := t.genBuyLimit(tx, buy, txIndex)
		tradelog.Info("Table", "buy-add", order)
		ldb.Add(order)
	} else {
		t.updateBuyLimit(tx, buy, buyOrder, txIndex, ldb)
	}

	return genSaveBuyLimitKv(buyOrder)
}

func saveBuyLimitOrderKeyValue(kv []*types.KeyValue, buyOrder *pty.BuyLimitOrder, status int32) []*types.KeyValue {
	buyID := []byte(buyOrder.BuyID)
	return genBuyLimitOrderKeyValue(kv, buyOrder, status, buyID)
}

func deleteBuyLimitKeyValue(kv []*types.KeyValue, buyOrder *pty.BuyLimitOrder, status int32) []*types.KeyValue {
	return genBuyLimitOrderKeyValue(kv, buyOrder, status, nil)
}

func genDeleteBuyLimitKv(buyOrder *pty.BuyLimitOrder) []*types.KeyValue {
	status := buyOrder.Status
	var kv []*types.KeyValue
	kv = deleteBuyLimitKeyValue(kv, buyOrder, status)
	if pty.TradeOrderStatusBoughtOut == status || pty.TradeOrderStatusBuyRevoked == status {
		tradelog.Debug("trade saveSell ", "remove old status onsale to soldout or revoked with sellid", buyOrder.BuyID)
		kv = saveBuyLimitOrderKeyValue(kv, buyOrder, pty.TradeOrderStatusOnBuy)
	}
	return kv
}

func (t *trade) deleteBuyLimit(buy *pty.ReceiptBuyBase, ty int32, tx *types.Transaction, txIndex string, ldb *table.Table, traded int64) []*types.KeyValue {
	buyOrder := t.getBuyOrderFromDb([]byte(buy.BuyID))
	if ty == pty.TyLogTradeBuyLimit && buy.BoughtBoardlot == 0 {
		ldb.Del([]byte(txIndex))
	} else {
		t.rollbackBuyLimit(tx, buy, buyOrder, txIndex, ldb, traded)
	}
	return genDeleteBuyLimitKv(buyOrder)
}

func (t *trade) saveSellMarket(receiptTradeBuy *pty.ReceiptSellBase, tx *types.Transaction, txIndex string, ldb *table.Table) []*types.KeyValue {
	var kv []*types.KeyValue
	order := t.genSellMarket(tx, receiptTradeBuy, txIndex)
	ldb.Add(order)
	return saveSellMarketOrderKeyValue(kv, receiptTradeBuy, pty.TradeOrderStatusSoldOut, t.GetHeight())
}

func (t *trade) deleteSellMarket(receiptTradeBuy *pty.ReceiptSellBase, txIndex string, ldb *table.Table) []*types.KeyValue {
	var kv []*types.KeyValue
	ldb.Del([]byte(txIndex))
	return deleteSellMarketOrderKeyValue(kv, receiptTradeBuy, pty.TradeOrderStatusSoldOut, t.GetHeight())
}

func saveSellMarketOrderKeyValue(kv []*types.KeyValue, receipt *pty.ReceiptSellBase, status int32, height int64) []*types.KeyValue {
	txhash := []byte(receipt.TxHash)
	return genSellMarketOrderKeyValue(kv, receipt, status, height, txhash)
}

func deleteSellMarketOrderKeyValue(kv []*types.KeyValue, receipt *pty.ReceiptSellBase, status int32, height int64) []*types.KeyValue {
	return genSellMarketOrderKeyValue(kv, receipt, status, height, nil)
}

func saveBuyMarketOrderKeyValue(kv []*types.KeyValue, receipt *pty.ReceiptBuyBase, status int32, height int64) []*types.KeyValue {
	txhash := []byte(receipt.TxHash)
	return genBuyMarketOrderKeyValue(kv, receipt, status, height, txhash)
}

func deleteBuyMarketOrderKeyValue(kv []*types.KeyValue, receipt *pty.ReceiptBuyBase, status int32, height int64) []*types.KeyValue {
	return genBuyMarketOrderKeyValue(kv, receipt, status, height, nil)
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (t *trade) CheckReceiptExecOk() bool {
	return true
}
