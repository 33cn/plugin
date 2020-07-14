package executor

import (
	"fmt"
	"math/big"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	tab "github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

// Action action struct
type Action struct {
	statedb   dbm.KV
	txhash    []byte
	fromaddr  string
	blocktime int64
	height    int64
	execaddr  string
	localDB   dbm.KVDB
	index     int
	api       client.QueueProtocolAPI
}

//NewAction ...
func NewAction(e *exchange, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &Action{e.GetStateDB(), hash, fromaddr,
		e.GetBlockTime(), e.GetHeight(), dapp.ExecAddress(string(tx.Execer)), e.GetLocalDB(), index, e.GetAPI()}
}

//GetIndex get index
func (a *Action) GetIndex() int64 {
	//扩容4个0,用于匹配多个matchorder索引时使用
	return (a.height*types.MaxTxsPerBlock + int64(a.index)) * 1e4
}

//GetKVSet get kv set
func (a *Action) GetKVSet(order *et.Order) (kvset []*types.KeyValue) {
	kvset = append(kvset, &types.KeyValue{Key: calcOrderKey(order.OrderID), Value: types.Encode(order)})
	return kvset
}

//OpSwap 反转
func (a *Action) OpSwap(op int32) int32 {
	if op == et.OpBuy {
		return et.OpSell
	}
	return et.OpBuy
}

//CalcActualCost 计算实际花费
func CalcActualCost(op int32, amount int64, price int64) int64 {
	if op == et.OpBuy {
		return SafeMul(amount, price)
	}
	return amount
}

//CheckPrice price 精度允许范围 1<=price<=1e16 整数
func CheckPrice(price int64) bool {
	if price > 1e16 || price < 1 {
		return false
	}
	return true
}

//CheckOp ...
func CheckOp(op int32) bool {
	if op == et.OpBuy || op == et.OpSell {
		return true
	}
	return false
}

//CheckCount ...
func CheckCount(count int32) bool {
	return count <= 20 && count >= 0
}

//CheckAmount 最小交易1e8
func CheckAmount(amount int64) bool {
	if amount < types.Coin || amount >= types.MaxCoin {
		return false
	}
	return true
}

//CheckDirection ...
func CheckDirection(direction int32) bool {
	if direction == et.ListASC || direction == et.ListDESC {
		return true
	}
	return false
}

//CheckStatus ...
func CheckStatus(status int32) bool {
	if status == et.Ordered || status == et.Completed || status == et.Revoked {
		return true
	}
	return false
}

//CheckExchangeAsset 检查交易得资产是否合法
func CheckExchangeAsset(left, right *et.Asset) bool {
	if left.Execer == "" || left.Symbol == "" || right.Execer == "" || right.Symbol == "" {
		return false
	}
	if (left.Execer == "coins" && right.Execer == "coins") || (left.Symbol == right.Symbol) {
		return false
	}
	return true
}

//LimitOrder ...
func (a *Action) LimitOrder(payload *et.LimitOrder) (*types.Receipt, error) {
	leftAsset := payload.GetLeftAsset()
	rightAsset := payload.GetRightAsset()
	//TODO 参数要合法，必须有严格的校验，后面统一加入到checkTx里面
	//coins执行器下面只有bty
	if !CheckExchangeAsset(leftAsset, rightAsset) {
		return nil, et.ErrAsset
	}
	if !CheckAmount(payload.GetAmount()) {
		return nil, et.ErrAssetAmount
	}
	if !CheckPrice(payload.GetPrice()) {
		return nil, et.ErrAssetPrice
	}
	if !CheckOp(payload.GetOp()) {
		return nil, et.ErrAssetOp
	}
	//TODO 这里symbol
	cfg := a.api.GetConfig()
	leftAssetDB, err := account.NewAccountDB(cfg, leftAsset.GetExecer(), leftAsset.GetSymbol(), a.statedb)
	if err != nil {
		return nil, err
	}
	rightAssetDB, err := account.NewAccountDB(cfg, rightAsset.GetExecer(), rightAsset.GetSymbol(), a.statedb)
	if err != nil {
		return nil, err
	}
	//先检查账户余额
	if payload.GetOp() == et.OpBuy {
		amount := SafeMul(payload.GetAmount(), payload.GetPrice())
		rightAccount := rightAssetDB.LoadExecAccount(a.fromaddr, a.execaddr)
		if rightAccount.Balance < amount {
			elog.Error("limit check right balance", "addr", a.fromaddr, "avail", rightAccount.Balance, "need", amount)
			return nil, et.ErrAssetBalance
		}
		return a.matchLimitOrder(payload, leftAssetDB, rightAssetDB)

	}
	if payload.GetOp() == et.OpSell {
		amount := payload.GetAmount()
		leftAccount := leftAssetDB.LoadExecAccount(a.fromaddr, a.execaddr)
		if leftAccount.Balance < amount {
			elog.Error("limit check left balance", "addr", a.fromaddr, "avail", leftAccount.Balance, "need", amount)
			return nil, et.ErrAssetBalance
		}
		return a.matchLimitOrder(payload, leftAssetDB, rightAssetDB)
	}
	return nil, fmt.Errorf("unknow op")
}

//RevokeOrder ...
func (a *Action) RevokeOrder(payload *et.RevokeOrder) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	order, err := findOrderByOrderID(a.statedb, a.localDB, payload.GetOrderID())
	if err != nil {
		return nil, err
	}
	if order.Addr != a.fromaddr {
		elog.Error("RevokeOrder.OrderCheck", "addr", a.fromaddr, "order.addr", order.Addr, "order.status", order.Status)
		return nil, et.ErrAddr
	}
	if order.Status == et.Completed || order.Status == et.Revoked {
		elog.Error("RevokeOrder.OrderCheck", "addr", a.fromaddr, "order.addr", order.Addr, "order.status", order.Status)
		return nil, et.ErrOrderSatus
	}
	leftAsset := order.GetLimitOrder().GetLeftAsset()
	rightAsset := order.GetLimitOrder().GetRightAsset()
	price := order.GetLimitOrder().GetPrice()
	balance := order.GetBalance()

	cfg := a.api.GetConfig()

	if order.GetLimitOrder().GetOp() == et.OpBuy {
		rightAssetDB, err := account.NewAccountDB(cfg, rightAsset.GetExecer(), rightAsset.GetSymbol(), a.statedb)
		if err != nil {
			return nil, err
		}
		amount := CalcActualCost(et.OpBuy, balance, price)
		rightAccount := rightAssetDB.LoadExecAccount(a.fromaddr, a.execaddr)
		if rightAccount.Frozen < amount {
			elog.Error("revoke check right frozen", "addr", a.fromaddr, "avail", rightAccount.Frozen, "amount", amount)
			return nil, et.ErrAssetBalance
		}
		receipt, err := rightAssetDB.ExecActive(a.fromaddr, a.execaddr, amount)
		if err != nil {
			elog.Error("RevokeOrder.ExecActive", "addr", a.fromaddr, "amount", amount, "err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
	}
	if order.GetLimitOrder().GetOp() == et.OpSell {
		leftAssetDB, err := account.NewAccountDB(cfg, leftAsset.GetExecer(), leftAsset.GetSymbol(), a.statedb)
		if err != nil {
			return nil, err
		}
		amount := CalcActualCost(et.OpSell, balance, price)
		leftAccount := leftAssetDB.LoadExecAccount(a.fromaddr, a.execaddr)
		if leftAccount.Frozen < amount {
			elog.Error("revoke check left frozen", "addr", a.fromaddr, "avail", leftAccount.Frozen, "amount", amount)
			return nil, et.ErrAssetBalance
		}
		receipt, err := leftAssetDB.ExecActive(a.fromaddr, a.execaddr, amount)
		if err != nil {
			elog.Error("RevokeOrder.ExecActive", "addr", a.fromaddr, "amount", amount, "err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
	}

	//更新order状态
	order.Status = et.Revoked
	order.UpdateTime = a.blocktime
	kvs = append(kvs, a.GetKVSet(order)...)
	re := &et.ReceiptExchange{
		Order: order,
		Index: a.GetIndex(),
	}
	receiptlog := &types.ReceiptLog{Ty: et.TyRevokeOrderLog, Log: types.Encode(re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil

}

//撮合交易逻辑方法
// 规则：
//1.买单高于市场价，按价格由低往高撮合。
//2.卖单低于市场价，按价格由高往低进行撮合。
//3.价格相同按先进先出的原则进行撮合
//4.买家获利得原则
func (a *Action) matchLimitOrder(payload *et.LimitOrder, leftAccountDB, rightAccountDB *account.DB) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var orderKey string
	var priceKey string
	var count int

	or := &et.Order{
		OrderID:    a.GetIndex(),
		Value:      &et.Order_LimitOrder{LimitOrder: payload},
		Ty:         et.TyLimitOrderAction,
		Executed:   0,
		AVGPrice:   0,
		Balance:    payload.GetAmount(),
		Status:     et.Ordered,
		Addr:       a.fromaddr,
		UpdateTime: a.blocktime,
		Index:      a.GetIndex(),
	}
	re := &et.ReceiptExchange{
		Order: or,
		Index: a.GetIndex(),
	}

	//单笔交易最多撮合100笔历史订单,最大可撮合得深度，系统得自我防护
	//迭代已有挂单价格
	for {
		//当撮合深度大于最大深度时跳出
		if count >= et.MaxMatchCount {
			break
		}
		//获取现有市场挂单价格信息
		marketDepthList, err := QueryMarketDepth(a.localDB, payload.GetLeftAsset(), payload.GetRightAsset(), a.OpSwap(payload.Op), priceKey, et.Count)
		if err == types.ErrNotFound {
			break
		}
		for _, marketDepth := range marketDepthList.List {
			if count >= et.MaxMatchCount {
				break
			}
			// 卖单价大于买单价
			if payload.Op == et.OpBuy && marketDepth.Price > payload.GetPrice() {
				continue
			}
			// 买单价小于卖单价
			if payload.Op == et.OpSell && marketDepth.Price < payload.GetPrice() {
				continue
			}
			//根据价格进行迭代
			for {
				//当撮合深度大于等于最大深度时跳出
				if count >= et.MaxMatchCount {
					break
				}
				orderList, err := findOrderIDListByPrice(a.localDB, payload.GetLeftAsset(), payload.GetRightAsset(), marketDepth.Price, a.OpSwap(payload.Op), et.ListASC, orderKey)
				if err == types.ErrNotFound {
					break
				}

				for _, matchorder := range orderList.List {
					//当撮合深度大于最大深度时跳出
					if count >= et.MaxMatchCount {
						break
					}
					//同地址不能交易
					if matchorder.Addr == a.fromaddr {
						continue
					}
					//撮合,指针传递
					log, kv, err := a.matchModel(leftAccountDB, rightAccountDB, payload, matchorder, or, re) // payload, or redundant
					if err != nil {
						return nil, err
					}
					logs = append(logs, log...)
					kvs = append(kvs, kv...)
					//订单完成,直接返回，如果没有完成，则继续撮合，直到count等于
					if or.Status == et.Completed {
						receiptlog := &types.ReceiptLog{Ty: et.TyLimitOrderLog, Log: types.Encode(re)}
						logs = append(logs, receiptlog)
						receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
						return receipts, nil
					}
					//TODO 这里得逻辑是否需要调整?当匹配的单数过多，会导致receipt日志数量激增，理论上存在日志存储攻击，需要加下最大匹配深度，防止这种攻击发生
					//撮合深度计数
					count = count + 1
				}
				//查询数据不满足10条说明没有了,跳出循环
				if orderList.PrimaryKey == "" {
					break
				}
				orderKey = orderList.PrimaryKey
			}
		}

		//查询的数据如果没有primaryKey说明没有后续数据了,跳出循环
		if marketDepthList.PrimaryKey == "" {
			break
		}
		priceKey = marketDepthList.PrimaryKey
	}

	//未完成的订单需要冻结剩余未成交的资金
	if payload.Op == et.OpBuy {
		amount := CalcActualCost(et.OpBuy, or.Balance, payload.Price)
		receipt, err := rightAccountDB.ExecFrozen(a.fromaddr, a.execaddr, amount)
		if err != nil {
			elog.Error("LimitOrder.ExecFrozen", "addr", a.fromaddr, "amount", amount, "err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
	}
	if payload.Op == et.OpSell {
		amount := CalcActualCost(et.OpSell, or.Balance, payload.Price)
		receipt, err := leftAccountDB.ExecFrozen(a.fromaddr, a.execaddr, amount)
		if err != nil {
			elog.Error("LimitOrder.ExecFrozen", "addr", a.fromaddr, "amount", amount, "err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
	}
	//更新order状态
	kvs = append(kvs, a.GetKVSet(or)...)
	re.Order = or
	receiptlog := &types.ReceiptLog{Ty: et.TyLimitOrderLog, Log: types.Encode(re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

//交易撮合模型
func (a *Action) matchModel(leftAccountDB, rightAccountDB *account.DB, payload *et.LimitOrder, matchorder *et.Order, or *et.Order, re *et.ReceiptExchange) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var matched int64

	if matchorder.GetBalance() >= or.GetBalance() {
		matched = or.GetBalance()
	} else {
		matched = matchorder.GetBalance()
	}

	elog.Info("try match", "activeId", or.OrderID, "passiveId", matchorder.OrderID, "activeAddr", or.Addr, "passiveAddr",
		matchorder.Addr, "amount", matched, "price", payload.Price)

	if payload.Op == et.OpSell {
		//转移冻结资产
		amount := CalcActualCost(matchorder.GetLimitOrder().Op, matched, payload.Price)
		receipt, err := rightAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, amount)
		if err != nil {
			elog.Error("matchModel.ExecTransferFrozen", "from", matchorder.Addr, "to", a.fromaddr, "amount", amount, "err", err)
			return nil, nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
		//解冻多余资金
		if payload.Price < matchorder.GetLimitOrder().Price {
			amount := CalcActualCost(matchorder.GetLimitOrder().Op, matched, matchorder.GetLimitOrder().Price-payload.Price)
			receipt, err := rightAccountDB.ExecActive(matchorder.Addr, a.execaddr, amount)
			if err != nil {
				elog.Error("matchModel.ExecActive", "addr", matchorder.Addr, "amount", amount, "err", err.Error())
				return nil, nil, err
			}
			logs = append(logs, receipt.Logs...)
			kvs = append(kvs, receipt.KV...)
		}
		//将达成交易的相应资产结算
		amount = CalcActualCost(payload.Op, matched, payload.Price)
		receipt, err = leftAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, amount)
		if err != nil {
			elog.Error("matchModel.ExecTransfer", "from", a.fromaddr, "to", matchorder.Addr, "amount", amount, "err", err.Error())
			return nil, nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)

		//卖单成交得平均价格始终与自身挂单价格相同
		or.AVGPrice = payload.Price
		//计算matchOrder平均成交价格
		matchorder.AVGPrice = caclAVGPrice(matchorder, payload.Price, matched) //TODO
	}
	if payload.Op == et.OpBuy {
		//转移冻结资产
		amount := CalcActualCost(matchorder.GetLimitOrder().Op, matched, matchorder.GetLimitOrder().Price)
		receipt, err := leftAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, amount)
		if err != nil {
			elog.Error("matchModel.ExecTransferFrozen2", "from", matchorder.Addr, "to", a.fromaddr, "amount", amount, "err", err.Error())
			return nil, nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
		//将达成交易的相应资产结算
		amount = CalcActualCost(payload.Op, matched, matchorder.GetLimitOrder().Price)
		receipt, err = rightAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, amount)
		if err != nil {
			elog.Error("matchModel.ExecTransfer2", "from", a.fromaddr, "to", matchorder.Addr, "amount", amount, "err", err.Error())
			return nil, nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)

		//买单得话，价格选取卖单的价格
		or.AVGPrice = matchorder.GetLimitOrder().Price
		//计算matchOrder平均成交价格
		matchorder.AVGPrice = caclAVGPrice(matchorder, matchorder.GetLimitOrder().Price, matched) //TODO
	}

	if matched == matchorder.GetBalance() {
		matchorder.Status = et.Completed
	} else {
		matchorder.Status = et.Ordered
	}

	if matched == or.GetBalance() {
		or.Status = et.Completed
	} else {
		or.Status = et.Ordered
	}

	if matched == or.GetBalance() {
		matchorder.Balance -= matched
		matchorder.Executed = matched
		kvs = append(kvs, a.GetKVSet(matchorder)...)

		or.Executed += matched
		or.Balance = 0
		kvs = append(kvs, a.GetKVSet(or)...) //or complete
	} else {
		or.Balance -= matched
		or.Executed += matched

		matchorder.Executed = matched
		matchorder.Balance = 0
		kvs = append(kvs, a.GetKVSet(matchorder)...) //matchorder complete
	}

	re.Order = or
	re.MatchOrders = append(re.MatchOrders, matchorder)
	return logs, kvs, nil
}

//根据订单号查询，分为两步，优先去localdb中查询，如没有则再去状态数据库中查询
// 1.挂单中得订单信会根据orderID在localdb中存储
// 2.订单撤销，或者成交后，根据orderID在localdb中存储得数据会被删除，这时只能到状态数据库中查询
func findOrderByOrderID(statedb dbm.KV, localdb dbm.KV, orderID int64) (*et.Order, error) {
	table := NewMarketOrderTable(localdb)
	primaryKey := []byte(fmt.Sprintf("%022d", orderID))
	row, err := table.GetData(primaryKey)
	if err != nil {
		data, err := statedb.Get(calcOrderKey(orderID))
		if err != nil {
			elog.Error("findOrderByOrderID.Get", "orderID", orderID, "err", err.Error())
			return nil, err
		}
		var order et.Order
		err = types.Decode(data, &order)
		if err != nil {
			elog.Error("findOrderByOrderID.Decode", "orderID", orderID, "err", err.Error())
			return nil, err
		}
		return &order, nil
	}
	return row.Data.(*et.Order), nil

}

func findOrderIDListByPrice(localdb dbm.KV, left, right *et.Asset, price int64, op, direction int32, primaryKey string) (*et.OrderList, error) {
	table := NewMarketOrderTable(localdb)
	prefix := []byte(fmt.Sprintf("%s:%s:%d:%016d", left.GetSymbol(), right.GetSymbol(), op, price))

	var rows []*tab.Row
	var err error
	if primaryKey == "" { //第一次查询,默认展示最新得成交记录
		rows, err = table.ListIndex("market_order", prefix, nil, et.Count, direction)
	} else {
		rows, err = table.ListIndex("market_order", prefix, []byte(primaryKey), et.Count, direction)
	}
	if err != nil {
		elog.Error("findOrderIDListByPrice.", "left", left, "right", right, "price", price, "err", err.Error())
		return nil, err
	}
	var orderList et.OrderList
	for _, row := range rows {
		order := row.Data.(*et.Order)
		//替换已经成交得量
		order.Executed = order.GetLimitOrder().Amount - order.Balance
		orderList.List = append(orderList.List, order)
	}
	//设置主键索引
	if len(rows) == int(et.Count) {
		orderList.PrimaryKey = string(rows[len(rows)-1].Primary)
	}
	return &orderList, nil
}

//Direction 买单深度是按价格倒序，由高到低
func Direction(op int32) int32 {
	if op == et.OpBuy {
		return et.ListDESC
	}
	return et.ListASC
}

//QueryMarketDepth 这里primaryKey当作主键索引来用，首次查询不需要填值,买单按价格由高往低，卖单按价格由低往高查询
func QueryMarketDepth(localdb dbm.KV, left, right *et.Asset, op int32, primaryKey string, count int32) (*et.MarketDepthList, error) {
	table := NewMarketDepthTable(localdb)
	prefix := []byte(fmt.Sprintf("%s:%s:%d", left.GetSymbol(), right.GetSymbol(), op))
	if count == 0 {
		count = et.Count
	}
	var rows []*tab.Row
	var err error
	if primaryKey == "" { //第一次查询,默认展示最新得成交记录
		rows, err = table.ListIndex("price", prefix, nil, count, Direction(op))
	} else {
		rows, err = table.ListIndex("price", prefix, []byte(primaryKey), count, Direction(op))
	}
	if err != nil {
		//elog.Error("QueryMarketDepth.", "left", left, "right", right, "err", err.Error())
		return nil, err
	}

	var list et.MarketDepthList
	for _, row := range rows {
		list.List = append(list.List, row.Data.(*et.MarketDepth))
	}
	//设置主键索引
	if len(rows) == int(count) {
		list.PrimaryKey = string(rows[len(rows)-1].Primary)
	}
	return &list, nil
}

//QueryHistoryOrderList 只返回成交的订单信息
func QueryHistoryOrderList(localdb dbm.KV, left, right *et.Asset, primaryKey string, count, direction int32) (types.Message, error) {
	table := NewHistoryOrderTable(localdb)
	prefix := []byte(fmt.Sprintf("%s:%s", left.Symbol, right.Symbol))
	indexName := "name"
	if count == 0 {
		count = et.Count
	}
	var rows []*tab.Row
	var err error
	var orderList et.OrderList
HERE:
	if primaryKey == "" { //第一次查询,默认展示最新得成交记录
		rows, err = table.ListIndex(indexName, prefix, nil, count, direction)
	} else {
		rows, err = table.ListIndex(indexName, prefix, []byte(primaryKey), count, direction)
	}
	if err != nil && err != types.ErrNotFound {
		elog.Error("QueryCompletedOrderList.", "left", left, "right", right, "err", err.Error())
		return nil, err
	}
	if err == types.ErrNotFound {
		return &orderList, nil
	}
	for _, row := range rows {
		order := row.Data.(*et.Order)
		//因为这张表里面记录了 completed,revoked 两种状态的订单，所以需要过滤
		if order.Status == et.Revoked {
			continue
		}
		//替换已经成交得量
		order.Executed = order.GetLimitOrder().Amount - order.Balance
		orderList.List = append(orderList.List, order)
		if len(orderList.List) == int(count) {
			//设置主键索引
			orderList.PrimaryKey = string(row.Primary)
			return &orderList, nil
		}
	}
	if len(orderList.List) != int(count) && len(rows) == int(count) {
		primaryKey = string(rows[len(rows)-1].Primary)
		goto HERE
	}
	return &orderList, nil
}

//QueryOrderList 默认展示最新的
func QueryOrderList(localdb dbm.KV, addr string, status, count, direction int32, primaryKey string) (types.Message, error) {
	var table *tab.Table
	if status == et.Completed || status == et.Revoked {
		table = NewHistoryOrderTable(localdb)
	} else {
		table = NewMarketOrderTable(localdb)
	}
	prefix := []byte(fmt.Sprintf("%s:%d", addr, status))
	indexName := "addr_status"
	if count == 0 {
		count = et.Count
	}
	var rows []*tab.Row
	var err error
	if primaryKey == "" { //第一次查询,默认展示最新得成交记录
		rows, err = table.ListIndex(indexName, prefix, nil, count, direction)
	} else {
		rows, err = table.ListIndex(indexName, prefix, []byte(primaryKey), count, direction)
	}
	if err != nil {
		elog.Error("QueryOrderList.", "addr", addr, "err", err.Error())
		return nil, err
	}
	var orderList et.OrderList
	for _, row := range rows {
		order := row.Data.(*et.Order)
		//替换已经成交得量
		order.Executed = order.GetLimitOrder().Amount - order.Balance
		orderList.List = append(orderList.List, order)
	}
	//设置主键索引
	if len(rows) == int(count) {
		orderList.PrimaryKey = string(rows[len(rows)-1].Primary)
	}
	return &orderList, nil
}

func queryMarketDepth(localdb dbm.KV, left, right *et.Asset, op int32, price int64) (*et.MarketDepth, error) {
	table := NewMarketDepthTable(localdb)
	primaryKey := []byte(fmt.Sprintf("%s:%s:%d:%016d", left.GetSymbol(), right.GetSymbol(), op, price))
	row, err := table.GetData(primaryKey)
	if err != nil {
		return nil, err
	}
	return row.Data.(*et.MarketDepth), nil
}

//SafeMul math库中的安全大数乘法，防溢出
func SafeMul(x, y int64) int64 {
	res := big.NewInt(0).Mul(big.NewInt(x), big.NewInt(y))
	res = big.NewInt(0).Div(res, big.NewInt(types.Coin))
	return res.Int64()
}

//计算平均成交价格
func caclAVGPrice(order *et.Order, price int64, amount int64) int64 {
	x := big.NewInt(0).Mul(big.NewInt(order.AVGPrice), big.NewInt(order.GetLimitOrder().Amount-order.GetBalance()))
	y := big.NewInt(0).Mul(big.NewInt(price), big.NewInt(amount))
	total := big.NewInt(0).Add(x, y)
	div := big.NewInt(0).Add(big.NewInt(order.GetLimitOrder().Amount-order.GetBalance()), big.NewInt(amount))
	avg := big.NewInt(0).Div(total, div)
	return avg.Int64()
}
