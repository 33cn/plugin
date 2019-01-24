// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

/*
现有接口
 1.  查询地址对应的买单 （无分页）
   1.1 只指定地址   -> owner
   1.2 同时指定地址和token  -> owner_asset
   1.3 显示一个用户成交的所有买单 -> owner
   1.4 显示一个用户成交的指定一个或者多个token所有买单 -> owner_asset 不支持多个
 2. 分状态查询地址的买单： 状态 地址 （无分页） -> owner_status
 3. 显示一个token 指定数量的买单 GetTokenBuyOrderByStatus  -> asset_inBuy_status
 4. 显示指定token出售者的一个或多个token 或 不指定token 的卖单 （无分页） -> owner_asset/owner_asset_isSell 不支持多个
 5. 显示指定状态下的某地址卖单 （无分页）  -> owner_isSell_status
 6. 显示一个token 指定数量的卖单    -> asset_isSell
 7. 根据状态分页列出某地址的订单（包括买单卖单） owner_status
*/
var opt_order_table = &table.Option{
	Prefix:  "LODB-trade",
	Name:    "order",
	Primary: "txIndex",
	// asset = asset_exec+asset_symbol
	//
	// status: 设计为可以同时查询几种的并集 , 存储为前缀， 需要提前设计需要合并的， 用前缀表示
	//    进行中，  撤销，  部分成交 ， 全部成交，  完成状态统一前缀. 数字和原来不一样
	//      00     10     11          12         1*
	// 排序过滤条件： 可以组合，status&isSell 和前缀冲突
	Index: []string{
		"key",                 // 内部查询用
		"asset",               // 按资产统计订单
		"asset_isSell_status", // 接口 3
		// "asset_status", 可能需求， 用于资产的交易历史
		// "asset_isSell",
		"owner",              // 接口 1.1， 1.3
		"owner_asset",        // 接口 1.2, 1.4, 4, 7
		"owner_asset_isSell", // 接口 4
		"owner_asset_status", // 新需求， 在
		"owner_isSell",       // 接口 6
		// "owner_isSell_status",  可能需求， 界面分开显示订单
		// "owner_isSell_statusPrefix", // 状态可以定制组合, 成交历史需求
		"owner_status",             // 接口 2
		"assset_isSell_isFinished", // 用 isFinish, 进行订单是否完成的列表功能
		"owner_asset_isFinished",
		"owner_isFinished",
		// "owner_statusPrefix", // 状态可以定制组合 , 成交历史需求
	},
}

// OrderRow order row
type OrderRow struct {
	*pty.LocalOrder
}

// NewOrderRow create row
func NewOrderRow() *OrderRow {
	return &OrderRow{LocalOrder: nil}
}

// CreateRow create row
func (r *OrderRow) CreateRow() *table.Row {
	return &table.Row{Data: &pty.LocalOrder{}}
}

// SetPayload set payload
func (r *OrderRow) SetPayload(data types.Message) error {
	if d, ok := data.(*pty.LocalOrder); ok {
		r.LocalOrder = d
		return nil
	}
	return types.ErrTypeAsset
}

// Get get index key
func (r *OrderRow) Get(key string) ([]byte, error) {
	switch key {
	case "txIndex":
		return []byte(r.TxIndex), nil
	case "key":
		return []byte(r.Key), nil
	case "asset":
		return []byte(r.asset()), nil
	case "asset_isSell_status":
		return []byte(fmt.Sprintf("%s_%d_%s", r.asset(), r.isSell(), r.status())), nil
	case "owner":
		return []byte(r.Owner), nil
	case "owner_asset":
		return []byte(fmt.Sprintf("%s_%s", r.Owner, r.asset())), nil
	case "owner_asset_isSell":
		return []byte(fmt.Sprintf("%s_%s_%d", r.Owner, r.asset(), r.isSell())), nil
	case "owner_asset_status":
		return []byte(fmt.Sprintf("%s_%s_%s", r.Owner, r.asset(), r.status())), nil
	case "owner_isSell":
		return []byte(fmt.Sprintf("%s_%d", r.Owner, r.isSell())), nil
	//case "owner_isSell_statusPrefix":
	//	return []byte(fmt.Sprintf("%s_%d_%s", r.Owner, r.asset(), r.isSell())), nil
	case "owner_status":
		return []byte(fmt.Sprintf("%s_%s", r.Owner, r.status())), nil
	//case "owner_statusPrefix":
	//	return []byte(fmt.Sprintf("%s_%d", r.Owner, r.isSell())), nil
	case "assset_isSell_isFinished":
		return []byte(fmt.Sprintf("%s_%d_%d", r.Owner, r.isSell(), r.isFinished())), nil
	case "owner_asset_isFinished":
		return []byte(fmt.Sprintf("%s_%s_%d", r.Owner, r.asset(), r.isFinished())), nil
	case "owner_isFinished":
		return []byte(fmt.Sprintf("%s_%d", r.Owner, r.isFinished())), nil
	default:
		return nil, types.ErrNotFound
	}
}

func (r *OrderRow) asset() string {
	return r.LocalOrder.AssetExec + "." + r.LocalOrder.AssetSymbol
}

func (r *OrderRow) isSell() int {
	if r.IsSellOrder {
		return 1
	}
	return 0
}

func (r *OrderRow) isFinished() int {
	if r.IsFinished {
		return 1
	}
	return 0
}

// status: 设计为可以同时查询几种的并集 , 存储为前缀， 需要提前设计需要合并的， 用前缀表示
//    进行中，  撤销，  部分成交 ， 全部成交，  完成状态统一前缀. 数字和原来不一样
//      01     10     11          12        19 -> 1*
func (r *OrderRow) status() string {
	if r.Status == pty.TradeOrderStatusOnBuy || r.Status == pty.TradeOrderStatusOnSale {
		return "01" // 试图用1 可以匹配所有完成的
	} else if r.Status == pty.TradeOrderStatusSoldOut || r.Status == pty.TradeOrderStatusBoughtOut {
		return "12"
	} else if r.Status == pty.TradeOrderStatusRevoked || r.Status == pty.TradeOrderStatusBuyRevoked {
		return "10"
	} else if r.Status == pty.TradeOrderStatusSellHalfRevoked || r.Status == pty.TradeOrderStatusBuyHalfRevoked {
		return "11"
	} else if r.Status == pty.TradeOrderStatusGroupComplete {
		return "1" // 1* match complete
	}

	return "XX"
}

// NewOrderTable create order table
func NewOrderTable(kvdb dbm.KV) *table.Table {
	rowMeta := NewOrderRow()
	rowMeta.SetPayload(&pty.LocalOrder{})
	t, err := table.NewTable(rowMeta, kvdb, opt_order_table)
	if err != nil {
		panic(err)
	}
	return t
}

func (t *trade) genSellLimit(tx *types.Transaction, sell *pty.ReceiptSellBase,
	sellorder *pty.SellOrder, txIndex string) *pty.LocalOrder {

	order := &pty.LocalOrder{
		AssetSymbol:       sellorder.TokenSymbol,
		TxIndex:           txIndex,
		Owner:             sellorder.Address,
		AmountPerBoardlot: sellorder.AmountPerBoardlot,
		MinBoardlot:       sellorder.MinBoardlot,
		PricePerBoardlot:  sellorder.PricePerBoardlot,
		TotalBoardlot:     sellorder.TotalBoardlot,
		TradedBoardlot:    sellorder.SoldBoardlot,
		BuyID:             "",
		Status:            sellorder.Status,
		SellID:            sell.SellID,
		TxHash:            []string{common.ToHex(tx.Hash())},
		Height:            sell.Height,
		Key:               sell.SellID,
		BlockTime:         t.GetBlockTime(),
		IsSellOrder:       true,
		AssetExec:         sellorder.AssetExec,
		IsFinished:        false,
	}
	return order
}

func (t *trade) updateSellLimit(tx *types.Transaction, sell *pty.ReceiptSellBase,
	sellorder *pty.SellOrder, txIndex string, ldb *table.Table) *pty.LocalOrder {

	xs, err := ldb.ListIndex("key", []byte(sell.SellID), nil, 1, 0)
	if err != nil || len(xs) != 1 {
		return nil
	}
	order, ok := xs[0].Data.(*pty.LocalOrder)
	tradelog.Debug("Table dbg", "sell-update", order, "data", xs[0].Data)
	if !ok {
		tradelog.Error("Table failed", "sell-update", order)
		return nil

	}
	status := sellorder.Status
	if status == pty.TradeOrderStatusRevoked && sell.SoldBoardlot > 0 {
		status = pty.TradeOrderStatusSellHalfRevoked
	}
	order.Status = status
	order.TxHash = append(order.TxHash, common.ToHex(tx.Hash()))
	order.TradedBoardlot = sellorder.SoldBoardlot
	order.IsFinished = (status != pty.TradeOrderStatusOnSale)

	tradelog.Debug("Table", "sell-update", order)

	ldb.Replace(order)

	return order
}

func (t *trade) rollBackSellLimit(tx *types.Transaction, sell *pty.ReceiptSellBase,
	sellorder *pty.SellOrder, txIndex string, ldb *table.Table, tradedBoardlot int64) *pty.LocalOrder {

	xs, err := ldb.ListIndex("key", []byte(sell.SellID), nil, 1, 0)
	if err != nil || len(xs) != 1 {
		return nil
	}
	order, ok := xs[0].Data.(*pty.LocalOrder)
	if !ok {
		return nil

	}
	// 撤销订单回滚, 只需要修改状态
	// 其他的操作需要还修改数量
	order.Status = pty.TradeOrderStatusOnSale
	order.TxHash = order.TxHash[:len(order.TxHash)-1]
	order.TradedBoardlot = order.TradedBoardlot - tradedBoardlot
	order.IsFinished = (order.Status != pty.TradeOrderStatusOnSale)

	ldb.Replace(order)

	return order
}

func parseOrderAmountFloat(s string) int64 {
	x, err := strconv.ParseFloat(s, 64)
	if err != nil {
		tradelog.Error("parseOrderAmountFloat", "decode receipt", err)
		return 0
	}
	return int64(x * float64(types.TokenPrecision))
}

func parseOrderPriceFloat(s string) int64 {
	x, err := strconv.ParseFloat(s, 64)
	if err != nil {
		tradelog.Error("parseOrderPriceFloat", "decode receipt", err)
		return 0
	}
	return int64(x * float64(types.Coin))
}

func (t *trade) genSellMarket(tx *types.Transaction, sell *pty.ReceiptSellBase, txIndex string) *pty.LocalOrder {

	order := &pty.LocalOrder{
		AssetSymbol:       sell.TokenSymbol,
		TxIndex:           txIndex,
		Owner:             sell.Owner,
		AmountPerBoardlot: parseOrderAmountFloat(sell.AmountPerBoardlot),
		MinBoardlot:       sell.MinBoardlot,
		PricePerBoardlot:  parseOrderPriceFloat(sell.PricePerBoardlot),
		TotalBoardlot:     sell.TotalBoardlot,
		TradedBoardlot:    sell.SoldBoardlot,
		BuyID:             sell.BuyID,
		Status:            pty.TradeOrderStatusSoldOut,
		SellID:            calcTokenSellID(hex.EncodeToString(tx.Hash())),
		TxHash:            []string{common.ToHex(tx.Hash())},
		Height:            sell.Height,
		Key:               calcTokenSellID(hex.EncodeToString(tx.Hash())),
		BlockTime:         t.GetBlockTime(),
		IsSellOrder:       true,
		AssetExec:         sell.AssetExec,

		IsFinished: true,
	}
	return order
}

func (t *trade) genBuyLimit(tx *types.Transaction, buy *pty.ReceiptBuyBase, txIndex string) *pty.LocalOrder {

	order := &pty.LocalOrder{
		AssetSymbol:       buy.TokenSymbol,
		TxIndex:           txIndex,
		Owner:             buy.Owner,
		AmountPerBoardlot: parseOrderAmountFloat(buy.AmountPerBoardlot),
		MinBoardlot:       buy.MinBoardlot,
		PricePerBoardlot:  parseOrderPriceFloat(buy.PricePerBoardlot),
		TotalBoardlot:     buy.TotalBoardlot,
		TradedBoardlot:    buy.BoughtBoardlot,
		BuyID:             buy.BuyID,
		Status:            pty.TradeOrderStatusOnBuy,
		SellID:            "",
		TxHash:            []string{common.ToHex(tx.Hash())},
		Height:            buy.Height,
		Key:               buy.BuyID,
		BlockTime:         t.GetBlockTime(),
		IsSellOrder:       false,
		AssetExec:         buy.AssetExec,
		IsFinished:        false,
	}
	return order
}

func (t *trade) updateBuyLimit(tx *types.Transaction, buy *pty.ReceiptBuyBase,
	buyorder *pty.BuyLimitOrder, txIndex string, ldb *table.Table) *pty.LocalOrder {

	xs, err := ldb.ListIndex("key", []byte(buy.BuyID), nil, 1, 0)
	if err != nil || len(xs) != 1 {
		return nil
	}
	order, ok := xs[0].Data.(*pty.LocalOrder)
	if !ok {
		return nil

	}
	status := buyorder.Status
	if status == pty.TradeOrderStatusBuyRevoked && buy.BoughtBoardlot > 0 {
		status = pty.TradeOrderStatusBuyHalfRevoked
	}
	order.Status = status
	order.TxHash = append(order.TxHash, common.ToHex(tx.Hash()))
	order.TradedBoardlot = buyorder.BoughtBoardlot
	order.IsFinished = (status != pty.TradeOrderStatusOnBuy)

	ldb.Replace(order)

	return order
}

func (t *trade) rollbackBuyLimit(tx *types.Transaction, buy *pty.ReceiptBuyBase,
	buyorder *pty.BuyLimitOrder, txIndex string, ldb *table.Table, traded int64) *pty.LocalOrder {

	xs, err := ldb.ListIndex("key", []byte(buy.BuyID), nil, 1, 0)
	if err != nil || len(xs) != 1 {
		return nil
	}
	order, ok := xs[0].Data.(*pty.LocalOrder)
	if !ok {
		return nil
	}

	order.Status = pty.TradeOrderStatusOnBuy
	order.TxHash = order.TxHash[:len(order.TxHash)-1]
	order.TradedBoardlot = order.TradedBoardlot - traded
	order.IsFinished = false

	ldb.Replace(order)

	return order
}

func (t *trade) genBuyMarket(tx *types.Transaction, buy *pty.ReceiptBuyBase, txIndex string) *pty.LocalOrder {

	order := &pty.LocalOrder{
		AssetSymbol:       buy.TokenSymbol,
		TxIndex:           txIndex,
		Owner:             buy.Owner,
		AmountPerBoardlot: parseOrderAmountFloat(buy.AmountPerBoardlot),
		MinBoardlot:       buy.MinBoardlot,
		PricePerBoardlot:  parseOrderPriceFloat(buy.PricePerBoardlot),
		TotalBoardlot:     buy.TotalBoardlot,
		TradedBoardlot:    buy.BoughtBoardlot,
		BuyID:             calcTokenBuyID(hex.EncodeToString(tx.Hash())),
		Status:            pty.TradeOrderStatusBoughtOut,
		SellID:            buy.SellID,
		TxHash:            []string{common.ToHex(tx.Hash())},
		Height:            buy.Height,
		Key:               calcTokenBuyID(hex.EncodeToString(tx.Hash())),
		BlockTime:         t.GetBlockTime(),
		IsSellOrder:       true,
		AssetExec:         buy.AssetExec,
		IsFinished:        true,
	}
	return order
}

func list(db dbm.KVDB, indexName string, data *pty.LocalOrder, count, direction int32) ([]*table.Row, error) {
	query := NewOrderTable(db).GetQuery(db)
	var primary []byte
	if len(data.TxIndex) > 0 {
		primary = []byte(data.TxIndex)
	}

	cur := &OrderRow{LocalOrder: data}
	index, err := cur.Get(indexName)
	if err != nil {
		tradelog.Error("query List failed", "key", string(primary), "param", data, "err", err)
		return nil, err
	}
	tradelog.Debug("query List dbg", "indexName", indexName, "index", string(index), "primary", primary, "count", count, "direction", direction)
	rows, err := query.ListIndex(indexName, index, primary, count, direction)
	if err != nil {
		tradelog.Error("query List failed", "key", string(primary), "param", data, "err", err)
		return nil, err
	}
	if len(rows) == 0 {
		return nil, types.ErrNotFound
	}
	return rows, nil
}

/*



按 资产 查询 ：
按 资产 & 地址 查询
按 地址

排序和分类
 1. 时间顺序   txindex
 1. 分类， 不同的状态 & 不同的性质： 买/卖

交易 -> 订单 按订单来 (交易和订单是多对多的关系，不适合joinTable)

交易 T1 Create -> T2 part-take -> T3 Revoke

订单左为进行中， 右为完成，
订单   （C1) | () ->  (C1m) | (C2) -> () | (C2, C1r)


查询交易 / 查询订单
  C ->   C/M -> C/D
  \
   \ ->R


状态 1, TradeOrderStatusOnSale, 在售
状态 2： TradeOrderStatusSoldOut，售完
状态 3： TradeOrderStatusRevoked， 卖单被撤回
状态 4： TradeOrderStatusExpired， 订单超时(目前不支持订单超时)
状态 5： TradeOrderStatusOnBuy， 求购
状态 6： TradeOrderStatusBoughtOut， 购买完成
状态 7： TradeOrderStatusBuyRevoked， 买单被撤回
*/
