package executor

import (
	"fmt"
	"math"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
	"github.com/gogo/protobuf/proto"
)

// Action action struct
type Action struct {
	statedb   dbm.KV
	txhash    []byte
	fromaddr  string
	blocktime int64
	height    int64
	execaddr  string
	localDB   dbm.Lister
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
	} else {
		return et.OpBuy
	}
}

//计算实际花费
func (a *Action) calcActualCost(op int32, amount int64, price float32) int64 {
	if op == et.OpBuy {
		return int64(float32(amount) * Truncate(price))
	}
	return amount
}

//price 精度允许范围小数点后面7位数,0<price<1e8
func CheckPrice(price float32) bool {
	if (Truncate(price) >= 1e8) || (Truncate(price)*float32(1e8) <= 0) {
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
	if count > 20 {
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
	if !types.CheckAmount(payload.GetAmount()) {
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
		amount := int64(float32(payload.GetAmount()) * Truncate(payload.GetPrice()))
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
	order, err := findOrderByOrderID(a.statedb, payload.GetOrderID())
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
func (a *Action) matchLimitOrder(payload *et.LimitOrder, leftAccountDB, rightAccountDB *account.DB) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var index int64

	or := &et.Order{
		OrderID:    common.ToHex(a.txhash),
		Value:      &et.Order_LimitOrder{payload},
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

	//按先进先出的原则进行撮合
	//单笔交易最多撮合100笔历史订单,最大可撮合得深度，系统得自我防护
	for i := 0; i < 20; i++ {
		orderIDs, err := findOrderIDListByPrice(a.localDB, payload.GetLeftAsset(), payload.GetRightAsset(), payload.GetPrice(), a.OpSwap(payload.Op), et.ListASC, index)
		if err == types.ErrNotFound {
			break
		}
		if err != nil {
			return nil, err
		}
		for _, orderID := range orderIDs {
			matchorder, err := findOrderByOrderID(a.statedb, orderID.ID)
			if err != nil {
				elog.Warn("matchLimitOrder.findOrderByOrderID", "order", "err", err.Error())
				continue
			}
			//同地址不能交易
			if matchorder.Addr == a.fromaddr {
				continue
			}
			//TODO 这里得逻辑是否需要调整?当匹配的单数过多，会导致receipt日志数量激增，理论上存在日志存储攻击，需要加下最大匹配深度，防止这种攻击发生
			if matchorder.GetBalance() >= or.GetBalance() {
				if payload.Op == et.OpSell {
					//转移冻结资产
					receipt, err := rightAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), payload.GetPrice()))
					if err != nil {
						elog.Error("matchLimitOrder.ExecTransferFrozen", "addr", matchorder.Addr, "execaddr", a.execaddr, "amount", a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), payload.GetPrice()), "err", err.Error())
						return nil, err
					}
					logs = append(logs, receipt.Logs...)
					kvs = append(kvs, receipt.KV...)
					//将达成交易的相应资产结算
					receipt, err = leftAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, a.calcActualCost(payload.Op, or.GetBalance(), payload.GetPrice()))
					if err != nil {
						elog.Error("matchLimitOrder.ExecTransfer", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", a.calcActualCost(payload.Op, or.GetBalance(), payload.GetPrice()), "err", err.Error())
						return nil, err
					}
					logs = append(logs, receipt.Logs...)
					kvs = append(kvs, receipt.KV...)
				}
				if payload.Op == et.OpBuy {
					//转移冻结资产
					receipt, err := leftAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), payload.GetPrice()))
					if err != nil {
						elog.Error("matchLimitOrder.ExecTransferFrozen", "addr", matchorder.Addr, "execaddr", a.execaddr, "amount", a.calcActualCost(matchorder.GetLimitOrder().Op, or.GetBalance(), payload.GetPrice()), "err", err.Error())
						return nil, err
					}
					logs = append(logs, receipt.Logs...)
					kvs = append(kvs, receipt.KV...)
					//将达成交易的相应资产结算
					receipt, err = rightAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, a.calcActualCost(payload.Op, or.GetBalance(), payload.GetPrice()))
					if err != nil {
						elog.Error("matchLimitOrder.ExecTransfer", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", a.calcActualCost(payload.Op, or.GetBalance(), payload.GetPrice()), "err", err.Error())
						return nil, err
					}
					logs = append(logs, receipt.Logs...)
					kvs = append(kvs, receipt.KV...)
				}

				// match receiptorder,涉及赋值先手顺序，代码顺序不可变
				matchorder.Status = func(a, b int64) int32 {
					if a > b {
						return et.Ordered
					} else {
						return et.Completed
					}
				}(matchorder.GetBalance(), or.GetBalance())
				matchorder.Balance = matchorder.GetBalance() - or.GetBalance()
				matchorder.Executed = matchorder.Executed + or.GetBalance()

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
			} else {
				if payload.Op == et.OpSell {
					//转移冻结资产
					receipt, err := rightAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, a.calcActualCost(matchorder.GetLimitOrder().Op, matchorder.GetBalance(), payload.GetPrice()))
					if err != nil {
						elog.Error("matchLimitOrder.ExecTransferFrozen", "addr", matchorder.Addr, "execaddr", a.execaddr, "amount", a.calcActualCost(matchorder.GetLimitOrder().Op, matchorder.GetBalance(), payload.GetPrice()), "err", err.Error())
						return nil, err
					}
					logs = append(logs, receipt.Logs...)
					kvs = append(kvs, receipt.KV...)
					//将达成交易的相应资产结算
					receipt, err = leftAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, a.calcActualCost(payload.Op, matchorder.GetBalance(), payload.GetPrice()))
					if err != nil {
						elog.Error("matchLimitOrder.ExecTransfer", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", a.calcActualCost(payload.Op, matchorder.GetBalance(), payload.GetPrice()), "err", err.Error())
						return nil, err
					}
					logs = append(logs, receipt.Logs...)
					kvs = append(kvs, receipt.KV...)
				}
				if payload.Op == et.OpBuy {
					//转移冻结资产
					receipt, err := leftAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, a.calcActualCost(matchorder.GetLimitOrder().Op, matchorder.GetBalance(), payload.GetPrice()))
					if err != nil {
						elog.Error("matchLimitOrder.ExecTransferFrozen", "addr", matchorder.Addr, "execaddr", a.execaddr, "amount", a.calcActualCost(matchorder.GetLimitOrder().Op, matchorder.GetBalance(), payload.GetPrice()), "err", err.Error())
						return nil, err
					}
					logs = append(logs, receipt.Logs...)
					kvs = append(kvs, receipt.KV...)
					//将达成交易的相应资产结算
					receipt, err = rightAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, a.calcActualCost(payload.Op, matchorder.GetBalance(), payload.GetPrice()))
					if err != nil {
						elog.Error("matchLimitOrder.ExecTransfer", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", a.calcActualCost(payload.Op, matchorder.GetBalance(), payload.GetPrice()), "err", err.Error())
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
				matchorder.Executed = matchorder.Executed + matchorder.GetBalance()
				matchorder.Status = et.Completed
				matchorder.Balance = 0
				a.updateStateDBCache(matchorder)
				kvs = append(kvs, a.GetKVSet(matchorder)...)

				re.Order = or
				re.MatchOrders = append(re.MatchOrders, matchorder)
			}
		}
		//查询数据不满足5条说明没有了,跳出循环
		if len(orderIDs) < int(et.Count) {
			break
		}
		index = orderIDs[len(orderIDs)-1].Index
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
func findOrderByOrderID(statedb dbm.KV, orderID string) (*et.Order, error) {
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

func findOrderIDListByPrice(localdb dbm.Lister, left, right *et.Asset, price float32, op, direction int32, index int64) ([]*et.OrderID, error) {
	prefix := calcMarketDepthOrderPrefix(left, right, op, price)
	key := calcMarketDepthOrderKey(left, right, op, price, index)
	var values [][]byte
	var err error
	if index == 0 { //第一次查询
		values, err = localdb.List(prefix, nil, et.Count, direction)
	} else {
		values, err = localdb.List(prefix, key, et.Count, direction)
	}
	if err != nil {
		return nil, err
	}
	var list []*et.OrderID
	for _, value := range values {
		var orderID et.OrderID
		err := types.Decode(value, &orderID)
		if err != nil {
			elog.Warn("findOrderIDListByPrice.Decode", "orderID", "err", err.Error())
			continue
		}
		list = append(list, &orderID)
	}
	return list, nil
}

//localdb查询
func findObject(localdb dbm.KVDB, key []byte, msg proto.Message) error {
	value, err := localdb.Get(key)
	if err != nil {
		elog.Warn("findObject.Decode", "key", string(key), "err", err.Error())
		return err
	}
	return types.Decode(value, msg)
}

//买单深度是按价格倒序，由高到低
func Direction(op int32) int32 {
	if op == et.OpBuy {
		return et.ListDESC
	}
	return et.ListASC
}

//这里price当作索引来用，首次查询不需要填值
func QueryMarketDepth(localdb dbm.Lister, left, right *et.Asset, op int32, price float32, count int32) (types.Message, error) {
	prefix := calcMarketDepthPrefix(left, right, op)
	key := calcMarketDepthKey(left, right, op, price)
	var values [][]byte
	var err error
	if price == 0 { //第一次查询，方向卖单由低到高，买单由高到低
		values, err = localdb.List(prefix, nil, count, Direction(op))
	} else {
		values, err = localdb.List(prefix, key, count, Direction(op))
	}
	if err != nil {
		return nil, err
	}
	var list et.MarketDepthList
	for _, value := range values {
		var marketDept et.MarketDepth
		err := types.Decode(value, &marketDept)
		if err != nil {
			elog.Warn("QueryMarketDepth.Decode", "marketDept", "err", err.Error())
			continue
		}
		list.List = append(list.List, &marketDept)
	}
	return &list, nil
}

//QueryCompletedOrderList
func QueryCompletedOrderList(localdb dbm.Lister, statedb dbm.KV, left, right *et.Asset, index int64, count, direction int32) (types.Message, error) {
	prefix := calcCompletedOrderPrefix(left, right)
	key := calcCompletedOrderKey(left, right, index)
	var values [][]byte
	var err error
	if index == 0 { //第一次查询,默认展示最新得成交记录
		values, err = localdb.List(prefix, nil, count, direction)
	} else {
		values, err = localdb.List(prefix, key, count, direction)
	}
	if err != nil {
		return nil, err
	}
	var list []*et.OrderID
	for _, value := range values {
		var orderID et.OrderID
		err := types.Decode(value, &orderID)
		if err != nil {
			elog.Warn("QueryCompletedOrderList.Decode", "marketDept", "err", err.Error())
			continue
		}
		list = append(list, &orderID)
	}
	var orderList et.OrderList
	for _, orderID := range list {
		order, err := findOrderByOrderID(statedb, orderID.ID)
		if err != nil {
			continue
		}
		//替换索引index
		order.Index = orderID.Index
		orderList.List = append(orderList.List, order)
	}

	return &orderList, nil
}

//QueryOrderList,默认展示最新的
func QueryOrderList(localdb dbm.Lister, statedb dbm.KV, addr string, status, count, direction int32, index int64) (types.Message, error) {
	prefix := calcUserOrderIDPrefix(status, addr)
	key := calcUserOrderIDKey(status, addr, index)
	var values [][]byte
	var err error
	if index == 0 { //第一次查询,默认展示最新得成交记录
		values, err = localdb.List(prefix, nil, count, direction)
	} else {
		values, err = localdb.List(prefix, key, count, direction)
	}
	if err != nil {
		return nil, err
	}
	var list []*et.OrderID
	for _, value := range values {
		var orderID et.OrderID
		err := types.Decode(value, &orderID)
		if err != nil {
			elog.Warn("QueryOrderList.Decode", "marketDept", "err", err.Error())
			continue
		}
		list = append(list, &orderID)
	}
	var orderList et.OrderList
	for _, orderID := range list {
		order, err := findOrderByOrderID(statedb, orderID.ID)
		if err != nil {
			continue
		}
		//替换索引index
		order.Index = orderID.Index
		orderList.List = append(orderList.List, order)
	}
	return &orderList, nil
}

//截取小数点后7位
func Truncate(price float32) float32 {
	return float32(math.Trunc(float64(1e8)*float64(price)) / float64(1e8))
}
