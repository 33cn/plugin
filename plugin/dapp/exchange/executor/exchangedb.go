package executor

import (
	"fmt"
	"math"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	//"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	. "github.com/33cn/chain33/common/db/table"
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

//缓存
func (a *Action) updateStateDBCache(order *et.Order) {
	a.statedb.Set(calcOrderKey(order.OrderID), types.Encode(order))
}

//反转
func (a *Action) OpSwap(op int32) int32 {
	if op == et.OpBuy {
		return et.OpSell
	}
	return et.OpBuy
}

//计算实际花费
func (a *Action) calcActualCost(op int32, amount int64, price float64) int64 {
	if op == et.OpBuy {
		return int64(float64(amount) * Truncate(price))
	}
	return amount
}

//price 精度允许范围小数点后面8位数,0<price<1e8
func CheckPrice(price float64) bool {
	if (Truncate(price) >= 1e8) || (Truncate(price)*float64(1e8) < 1) {
		return false
	}
	return true
}
func CheckOp(op int32) bool {
	if op == et.OpBuy || op == et.OpSell {
		return true
	}
	return false
}

func CheckCount(count int32) bool {
	return count <= 20 && count >= 0
}

//最小交易1e8
func CheckAmount(amount int64) bool {
	if amount < types.Coin || amount >= types.MaxCoin {
		return false
	}
	return true
}

func CheckDirection(direction int32) bool {
	if direction == et.ListASC || direction == et.ListDESC {
		return true
	}
	return false
}

func CheckStatus(status int32) bool {
	if status == et.Ordered || status == et.Completed || status == et.Revoked {
		return true
	}
	return false
}

//检查交易得资产是否合法
func CheckExchangeAsset(left, right *et.Asset) bool {
	if left.Execer == "" || left.Symbol == "" || right.Execer == "" || right.Symbol == "" {
		return false
	}
	if (left.Execer == "coins" && right.Execer == "coins") || (left.Symbol == right.Symbol) {
		return false
	}
	return true
}
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
		amount := int64(float64(payload.GetAmount()) * Truncate(payload.GetPrice()))
		rightAccount := rightAssetDB.LoadExecAccount(a.fromaddr, a.execaddr)
		if rightAccount.Balance < amount {
			elog.Error("LimitOrder.BalanceCheck", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", amount, "err", et.ErrAssetBalance.Error())
			return nil, et.ErrAssetBalance
		}
		return a.matchLimitOrder(payload, leftAssetDB, rightAssetDB)

	}
	if payload.GetOp() == et.OpSell {
		amount := payload.GetAmount()
		leftAccount := leftAssetDB.LoadExecAccount(a.fromaddr, a.execaddr)
		if leftAccount.Balance < amount {
			elog.Error("LimitOrder.BalanceCheck", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", amount, "err", et.ErrAssetBalance.Error())
			return nil, et.ErrAssetBalance
		}
		return a.matchLimitOrder(payload, leftAssetDB, rightAssetDB)
	}
	return nil, fmt.Errorf("unknow op")
}

func (a *Action) RevokeOrder(payload *et.RevokeOrder) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	order, err := findOrderByOrderID(a.statedb, a.localDB, payload.GetOrderID())
	elog.Info("RevokeOrder====", "order", order.Balance)
	if err != nil {
		return nil, err
	}
	if order.Addr != a.fromaddr {
		elog.Error("RevokeOrder.OrderCheck", "addr", a.fromaddr, "order.addr", order.Addr, "order.status", order.Status, "err", et.ErrAddr.Error())
		return nil, et.ErrAddr
	}
	if order.Status == et.Completed || order.Status == et.Revoked {
		elog.Error("RevokeOrder.OrderCheck", "addr", a.fromaddr, "order.addr", order.Addr, "order.status", order.Status, "err", et.ErrOrderSatus.Error())
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
		amount := a.calcActualCost(et.OpBuy, balance, price)
		rightAccount := rightAssetDB.LoadExecAccount(a.fromaddr, a.execaddr)
		if rightAccount.Frozen < amount {
			elog.Error("RevokeOrder.BalanceCheck", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", amount, "err", et.ErrAssetBalance.Error())
			return nil, et.ErrAssetBalance
		}
		receipt, err := rightAssetDB.ExecActive(a.fromaddr, a.execaddr, amount)
		if err != nil {
			elog.Error("RevokeOrder.ExecActive", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", amount, "err", err.Error())
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
		amount := a.calcActualCost(et.OpBuy, balance, price)
		leftAccount := leftAssetDB.LoadExecAccount(a.fromaddr, a.execaddr)
		if leftAccount.Frozen < amount {
			elog.Error("RevokeOrder.BalanceCheck", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", amount, "err", et.ErrAssetBalance.Error())
			return nil, et.ErrAssetBalance
		}
		receipt, err := leftAssetDB.ExecActive(a.fromaddr, a.execaddr, amount)
		if err != nil {
			elog.Error("RevokeOrder.ExecActive", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", amount, "err", err.Error())
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
		if count > et.MaxCount {
			break
		}
		marketDepthList, err := QueryMarketDepth(a.localDB, payload.GetLeftAsset(), payload.GetRightAsset(), a.OpSwap(payload.Op), priceKey, et.Count)
		if err == types.ErrNotFound {
			break
		}
		//elog.Info("matchLimitOrder.QueryMarketDepth", "marketList", marketDepthList)
		for _, marketDepth := range marketDepthList.List {
			if count > et.MaxCount {
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
				orderList, err := findOrderIDListByPrice(a.localDB, payload.GetLeftAsset(), payload.GetRightAsset(), marketDepth.Price, a.OpSwap(payload.Op), et.ListASC, orderKey)
				if err == types.ErrNotFound {
					break
				}
				for _, matchorder := range orderList.List {
					//同地址不能交易
					if matchorder.Addr == a.fromaddr {
						continue
					}
					//TODO 这里得逻辑是否需要调整?当匹配的单数过多，会导致receipt日志数量激增，理论上存在日志存储攻击，需要加下最大匹配深度，防止这种攻击发生
					if matchorder.GetBalance() >= or.GetBalance() {
						if payload.Op == et.OpSell {
							//转移冻结资产
							receipt, err := rightAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), payload.Price))
							if err != nil {
								elog.Error("matchLimitOrder.ExecTransferFrozen", "addr", matchorder.Addr, "execaddr", a.execaddr, "amount", a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), payload.Price), "err", err.Error())
								return nil, err
							}
							logs = append(logs, receipt.Logs...)
							kvs = append(kvs, receipt.KV...)

							//解冻多余资金
							if payload.Price < matchorder.GetLimitOrder().Price {
								receipt, err := rightAccountDB.ExecActive(matchorder.Addr, a.execaddr, a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), matchorder.GetLimitOrder().Price-payload.Price))
								if err != nil {
									elog.Error("matchLimitOrder.ExecActive", "addr", matchorder.Addr, "execaddr", a.execaddr, "amount", a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), matchorder.GetLimitOrder().Price-payload.Price), "err", err.Error())
									return nil, err
								}
								logs = append(logs, receipt.Logs...)
								kvs = append(kvs, receipt.KV...)
							}
							//将达成交易的相应资产结算
							receipt, err = leftAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, a.calcActualCost(payload.Op, or.GetBalance(), payload.Price))
							if err != nil {
								elog.Error("matchLimitOrder.ExecTransfer", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", a.calcActualCost(payload.Op, or.GetBalance(), payload.Price), "err", err.Error())
								return nil, err
							}
							logs = append(logs, receipt.Logs...)
							kvs = append(kvs, receipt.KV...)
						}
						if payload.Op == et.OpBuy {
							//转移冻结资产
							receipt, err := leftAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), matchorder.GetLimitOrder().Price))
							if err != nil {
								elog.Error("matchLimitOrder.ExecTransferFrozen", "addr", matchorder.Addr, "execaddr", a.execaddr, "amount", a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), matchorder.GetLimitOrder().Price), "err", err.Error())
								return nil, err
							}
							logs = append(logs, receipt.Logs...)
							kvs = append(kvs, receipt.KV...)
							//将达成交易的相应资产结算
							receipt, err = rightAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, a.calcActualCost(payload.Op, or.GetBalance(), matchorder.GetLimitOrder().Price))
							if err != nil {
								elog.Error("matchLimitOrder.ExecTransfer", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", a.calcActualCost(payload.Op, or.GetBalance(), matchorder.GetLimitOrder().Price), "err", err.Error())
								return nil, err
							}
							logs = append(logs, receipt.Logs...)
							kvs = append(kvs, receipt.KV...)
						}

						// match receiptorder,涉及赋值先手顺序，代码顺序不可变
						matchorder.Status = func(a, b int64) int32 {
							if a > b {
								return et.Ordered
							}
							return et.Completed
						}(matchorder.GetBalance(), or.GetBalance())
						matchorder.Balance = matchorder.GetBalance() - or.GetBalance()
						//记录本次成交得量
						matchorder.Executed = or.GetBalance()

						a.updateStateDBCache(matchorder)
						kvs = append(kvs, a.GetKVSet(matchorder)...)

						or.Executed = or.Executed + or.GetBalance()
						or.Status = et.Completed
						or.Balance = 0
						//update receipt order
						re.Order = or
						re.MatchOrders = append(re.MatchOrders, matchorder)

						a.updateStateDBCache(or)
						kvs = append(kvs, a.GetKVSet(or)...)
						//statedb 更新
						receiptlog := &types.ReceiptLog{Ty: et.TyLimitOrderLog, Log: types.Encode(re)}
						logs = append(logs, receiptlog)
						receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
						return receipts, nil
					}
					if payload.Op == et.OpSell {
						elog.Info("matchLimitOrder.findOrderByOrderID========", "order", matchorder)
						//转移冻结资产
						receipt, err := rightAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, a.calcActualCost(matchorder.GetLimitOrder().Op, matchorder.GetBalance(), payload.Price))
						if err != nil {
							elog.Error("matchLimitOrder.ExecTransferFrozen", "addr", matchorder.Addr, "execaddr", a.execaddr, "amount", a.calcActualCost(matchorder.GetLimitOrder().Op, matchorder.GetBalance(), payload.Price), "err", err.Error())
							return nil, err
						}
						logs = append(logs, receipt.Logs...)
						kvs = append(kvs, receipt.KV...)

						//解冻成交部分 多余资金
						if payload.Price < matchorder.GetLimitOrder().Price {
							receipt, err := rightAccountDB.ExecActive(matchorder.Addr, a.execaddr, a.calcActualCost(matchorder.GetLimitOrder().Op, matchorder.GetBalance(), matchorder.GetLimitOrder().Price-payload.Price))
							if err != nil {
								elog.Error("matchLimitOrder.ExecActive", "addr", matchorder.Addr, "execaddr", a.execaddr, "amount", a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), matchorder.GetLimitOrder().Price-payload.Price), "err", err.Error())
								return nil, err
							}
							logs = append(logs, receipt.Logs...)
							kvs = append(kvs, receipt.KV...)
						}

						//将达成交易的相应资产结算
						receipt, err = leftAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, a.calcActualCost(payload.Op, matchorder.GetBalance(), payload.Price))
						if err != nil {
							elog.Error("matchLimitOrder.ExecTransfer", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", a.calcActualCost(payload.Op, matchorder.GetBalance(), payload.Price), "err", err.Error())
							return nil, err
						}
						logs = append(logs, receipt.Logs...)
						kvs = append(kvs, receipt.KV...)
					}
					if payload.Op == et.OpBuy {
						elog.Info("matchLimitOrder.findOrderByOrderID++++++", "order", matchorder)
						//转移冻结资产
						receipt, err := leftAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, a.calcActualCost(matchorder.GetLimitOrder().Op, matchorder.GetBalance(), matchorder.GetLimitOrder().Price))
						if err != nil {
							elog.Error("matchLimitOrder.ExecTransferFrozen", "addr", matchorder.Addr, "execaddr", a.execaddr, "amount", a.calcActualCost(matchorder.GetLimitOrder().Op, matchorder.GetBalance(), matchorder.GetLimitOrder().Price), "err", err.Error())
							return nil, err
						}
						logs = append(logs, receipt.Logs...)
						kvs = append(kvs, receipt.KV...)
						//将达成交易的相应资产结算
						receipt, err = rightAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, a.calcActualCost(payload.Op, matchorder.GetBalance(), matchorder.GetLimitOrder().Price))
						if err != nil {
							elog.Error("matchLimitOrder.ExecTransfer", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", a.calcActualCost(payload.Op, matchorder.GetBalance(), matchorder.GetLimitOrder().Price), "err", err.Error())
							return nil, err
						}
						logs = append(logs, receipt.Logs...)
						kvs = append(kvs, receipt.KV...)
					}

					//涉及赋值先后顺序，不可颠倒
					or.Balance = or.Balance - matchorder.Balance
					or.Executed = or.Executed + matchorder.Balance
					or.Status = et.Ordered
					a.updateStateDBCache(or)

					// match receiptorder
					matchorder.Executed = matchorder.Balance
					matchorder.Status = et.Completed
					matchorder.Balance = 0
					a.updateStateDBCache(matchorder)
					kvs = append(kvs, a.GetKVSet(matchorder)...)

					re.Order = or
					re.MatchOrders = append(re.MatchOrders, matchorder)
					//撮合深度计数
					count = count + 1
				}
				//查询数据不满足5条说明没有了,跳出循环
				if orderList.PrimaryKey == "" {
					break
				}
				orderKey = orderList.PrimaryKey
			}
		}

		//查询数据不满足5条说明没有了,跳出循环
		if marketDepthList.PrimaryKey == "" {
			break
		}
		priceKey = marketDepthList.PrimaryKey
	}

	//冻结剩余未成交的资金
	if payload.Op == et.OpBuy {
		receipt, err := rightAccountDB.ExecFrozen(a.fromaddr, a.execaddr, a.calcActualCost(et.OpBuy, or.Balance, payload.Price))
		if err != nil {
			elog.Error("LimitOrder.ExecFrozen", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", a.calcActualCost(et.OpBuy, or.Balance, payload.Price), "err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
	}
	if payload.Op == et.OpSell {
		receipt, err := leftAccountDB.ExecFrozen(a.fromaddr, a.execaddr, a.calcActualCost(et.OpSell, or.Balance, payload.Price))
		if err != nil {
			elog.Error("LimitOrder.ExecFrozen", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", a.calcActualCost(et.OpSell, or.Balance, payload.Price), "err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
	}
	//更新order状态
	kvs = append(kvs, a.GetKVSet(or)...)
	receiptlog := &types.ReceiptLog{Ty: et.TyLimitOrderLog, Log: types.Encode(re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

//根据订单号查询，分为两步，优先去localdb中查询，如没有则再去状态数据库中查询
// 1.挂单中得订单信会根据orderID在localdb中存储
// 2.订单撤销，或者成交后，根据orderID在localdb中存储得数据会被删除，这时只能到状态数据库中查询
func findOrderByOrderID(statedb dbm.KV, localdb dbm.KVDB, orderID int64) (*et.Order, error) {
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

func findOrderIDListByPrice(localdb dbm.KVDB, left, right *et.Asset, price float64, op, direction int32, primaryKey string) (*et.OrderList, error) {
	table := NewMarketOrderTable(localdb)
	query := table.GetQuery(localdb)
	prefix := []byte(fmt.Sprintf("%s:%s:%d:%016d", left.GetSymbol(), right.GetSymbol(), op, int64(Truncate(price*float64(1e8)))))

	var rows []*Row
	var err error
	if primaryKey == "" { //第一次查询,默认展示最新得成交记录
		rows, err = query.ListIndex("market_order", prefix, nil, et.Count, direction)
	} else {
		rows, err = query.ListIndex("market_order", prefix, []byte(primaryKey), et.Count, direction)
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

//买单深度是按价格倒序，由高到低
func Direction(op int32) int32 {
	if op == et.OpBuy {
		return et.ListDESC
	}
	return et.ListASC
}

//这里primaryKey当作主键索引来用，首次查询不需要填值
func QueryMarketDepth(localdb dbm.KVDB, left, right *et.Asset, op int32, primaryKey string, count int32) (*et.MarketDepthList, error) {
	table := NewMarketDepthTable(localdb)
	query := table.GetQuery(localdb)
	prefix := []byte(fmt.Sprintf("%s:%s:%d", left.GetSymbol(), right.GetSymbol(), op))
	if count == 0 {
		count = et.Count
	}
	var rows []*Row
	var err error
	if primaryKey == "" { //第一次查询,默认展示最新得成交记录
		rows, err = query.ListIndex("price", prefix, nil, count, Direction(op))
	} else {
		rows, err = query.ListIndex("price", prefix, []byte(primaryKey), count, Direction(op))
	}
	if err != nil {
		elog.Error("QueryMarketDepth.", "left", left, "right", right, "err", err.Error())
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

//QueryCompletedOrderList
func QueryCompletedOrderList(localdb dbm.KVDB, left, right *et.Asset, primaryKey string, count, direction int32) (types.Message, error) {
	table := NewCompletedOrderTable(localdb)
	query := table.GetQuery(localdb)
	prefix := []byte(fmt.Sprintf("%s:%s", left.Symbol, right.Symbol))
	if count == 0 {
		count = et.Count
	}
	var rows []*Row
	var err error
	if primaryKey == "" { //第一次查询,默认展示最新得成交记录
		rows, err = query.ListIndex("index", prefix, nil, count, direction)
	} else {
		rows, err = query.ListIndex("index", prefix, []byte(primaryKey), count, direction)
	}
	if err != nil {
		elog.Error("QueryCompletedOrderList.", "left", left, "right", right, "err", err.Error())
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

//QueryOrderList,默认展示最新的
func QueryOrderList(localdb dbm.KVDB, statedb dbm.KV, addr string, status, count, direction int32, primaryKey string) (types.Message, error) {
	table := NewUserOrderTable(localdb)
	query := table.GetQuery(localdb)
	prefix := []byte(fmt.Sprintf("%s:%d", addr, status))
	if count == 0 {
		count = et.Count
	}
	var rows []*Row
	var err error
	if primaryKey == "" { //第一次查询,默认展示最新得成交记录
		rows, err = query.ListIndex("index", prefix, nil, count, direction)
	} else {
		rows, err = query.ListIndex("index", prefix, []byte(primaryKey), count, direction)
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

func queryMarketDepth(localdb dbm.KVDB, left, right *et.Asset, op int32, price float64) (*et.MarketDepth, error) {
	table := NewMarketDepthTable(localdb)
	primaryKey := []byte(fmt.Sprintf("%s:%s:%d:%016d", left.GetSymbol(), right.GetSymbol(), op, int64(Truncate(price)*float64(1e8))))
	row, err := table.GetData(primaryKey)
	if err != nil {
		return nil, err
	}
	return row.Data.(*et.MarketDepth), nil
}

//截取小数点后8位
func Truncate(price float64) float64 {
	return math.Trunc(float64(1e8)*price) / float64(1e8)
}
