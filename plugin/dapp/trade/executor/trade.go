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
	defaultPriceExec = "coins"
)

// Init : 注册当前trade合约
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, GetName(), newTrade, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

//InitExecType ...
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&trade{}))
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

// sell limit
func (t *trade) saveSell(base *pty.ReceiptSellBase, ty int32, tx *types.Transaction, txIndex string, ldb *table.Table) {
	sellorder := t.getSellOrderFromDb([]byte(base.SellID))

	if ty == pty.TyLogTradeSellLimit && sellorder.SoldBoardlot == 0 {
		newOrder := t.genSellLimit(tx, base, sellorder, txIndex)
		tradelog.Info("Table", "sell-add", newOrder)
		ldb.Add(newOrder)
	} else {
		t.updateSellLimit(tx, base, sellorder, txIndex, ldb)
	}
}

func (t *trade) deleteSell(base *pty.ReceiptSellBase, ty int32, tx *types.Transaction, txIndex string, ldb *table.Table, tradedBoardlot int64) {
	sellorder := t.getSellOrderFromDb([]byte(base.SellID))
	if ty == pty.TyLogTradeSellLimit && sellorder.SoldBoardlot == 0 {
		ldb.Del([]byte(txIndex))
	} else {
		t.rollBackSellLimit(tx, base, sellorder, txIndex, ldb, tradedBoardlot)
	}
}

func (t *trade) saveBuy(receiptTradeBuy *pty.ReceiptBuyBase, tx *types.Transaction, txIndex string, ldb *table.Table) {
	order := t.genBuyMarket(tx, receiptTradeBuy, txIndex)
	tradelog.Debug("trade BuyMarket save local", "order", order)
	ldb.Add(order)
}

func (t *trade) deleteBuy(receiptTradeBuy *pty.ReceiptBuyBase, txIndex string, ldb *table.Table) {
	ldb.Del([]byte(txIndex))
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

func (t *trade) saveBuyLimit(buy *pty.ReceiptBuyBase, ty int32, tx *types.Transaction, txIndex string, ldb *table.Table) {
	buyOrder := t.getBuyOrderFromDb([]byte(buy.BuyID))
	tradelog.Debug("Table", "buy-add", buyOrder)
	if buyOrder.Status == pty.TradeOrderStatusOnBuy && buy.BoughtBoardlot == 0 {
		order := t.genBuyLimit(tx, buy, txIndex)
		tradelog.Info("Table", "buy-add", order)
		ldb.Add(order)
	} else {
		t.updateBuyLimit(tx, buy, buyOrder, txIndex, ldb)
	}
}

func (t *trade) deleteBuyLimit(buy *pty.ReceiptBuyBase, ty int32, tx *types.Transaction, txIndex string, ldb *table.Table, traded int64) {
	buyOrder := t.getBuyOrderFromDb([]byte(buy.BuyID))
	if ty == pty.TyLogTradeBuyLimit && buy.BoughtBoardlot == 0 {
		ldb.Del([]byte(txIndex))
	} else {
		t.rollbackBuyLimit(tx, buy, buyOrder, txIndex, ldb, traded)
	}
}

func (t *trade) saveSellMarket(receiptTradeBuy *pty.ReceiptSellBase, tx *types.Transaction, txIndex string, ldb *table.Table) {
	order := t.genSellMarket(tx, receiptTradeBuy, txIndex)
	ldb.Add(order)

}

func (t *trade) deleteSellMarket(receiptTradeBuy *pty.ReceiptSellBase, txIndex string, ldb *table.Table) {
	ldb.Del([]byte(txIndex))
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (t *trade) CheckReceiptExecOk() bool {
	return true
}
