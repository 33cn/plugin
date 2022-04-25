package executor

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"reflect"

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
	return &Action{
		statedb:   e.GetStateDB(),
		txhash:    hash,
		fromaddr:  fromaddr,
		blocktime: e.GetBlockTime(),
		height:    e.GetHeight(),
		execaddr:  dapp.ExecAddress(string(tx.Execer)),
		localDB:   e.GetLocalDB(),
		index:     index,
		api:       e.GetAPI(),
	}
}

//GetIndex get index
func (a *Action) GetIndex() int64 {
	// Add four zeros to match multiple MatchOrder indexes
	return (a.height*types.MaxTxsPerBlock + int64(a.index)) * 1e4
}

//GetKVSet get kv set
func (a *Action) GetKVSet(order *et.Order) (kvset []*types.KeyValue) {
	kvset = append(kvset, &types.KeyValue{Key: calcOrderKey(order.OrderID), Value: types.Encode(order)})
	return kvset
}

//OpSwap reverse
func (a *Action) OpSwap(op int32) int32 {
	if op == et.OpBuy {
		return et.OpSell
	}
	return et.OpBuy
}

//CalcActualCost Calculate actual cost
func CalcActualCost(op int32, amount int64, price, coinPrecision int64) int64 {
	if op == et.OpBuy {
		return SafeMul(amount, price, coinPrecision)
	}
	return amount
}

//CheckPrice price  1<=price<=1e16
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

func Check5Count(count int32) bool {
	return count <= 50 && count >= 0
}

//CheckAmount 最小交易 1coin
func CheckAmount(amount, coinPrecision int64) bool {
	if amount < 1 || amount >= types.MaxCoin*coinPrecision {
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

//CheckExchangeAsset
func CheckExchangeAsset(coinExec string, left, right *et.Asset) bool {
	if left.Execer == "" || left.Symbol == "" || right.Execer == "" || right.Symbol == "" {
		return false
	}
	if (left.Execer == coinExec && right.Execer == coinExec) || (left.Symbol == right.Symbol) {
		return false
	}
	return true
}

//CheckDepth 1:价格精度；priceDigits+3：精度为百位
func CheckDepth(depth, priceDigits int32) bool {
	return depth <= priceDigits+3 && depth >= 1
}

//LimitOrder ...
func (a *Action) LimitOrder(payload *et.LimitOrder, entrustAddr string) (*types.Receipt, error) {
	leftAsset := payload.GetLeftAsset()
	rightAsset := payload.GetRightAsset()
	cfg := a.api.GetConfig()
	if !CheckExchangeAsset(cfg.GetCoinExec(), leftAsset, rightAsset) {
		return nil, et.ErrAsset
	}
	if !CheckAmount(payload.GetAmount(), cfg.GetCoinPrecision()) {
		return nil, et.ErrAssetAmount
	}
	if !CheckPrice(payload.GetPrice()) {
		return nil, et.ErrAssetPrice
	}
	if !CheckOp(payload.GetOp()) {
		return nil, et.ErrAssetOp
	}

	leftAssetDB, err := account.NewAccountDB(cfg, leftAsset.GetExecer(), leftAsset.GetSymbol(), a.statedb)
	if err != nil {
		return nil, err
	}
	rightAssetDB, err := account.NewAccountDB(cfg, rightAsset.GetExecer(), rightAsset.GetSymbol(), a.statedb)
	if err != nil {
		return nil, err
	}
	//Check your account balance first
	if payload.GetOp() == et.OpBuy {
		amount := SafeMul(payload.GetAmount(), payload.GetPrice(), cfg.GetCoinPrecision())
		rightAccount := rightAssetDB.LoadExecAccount(a.fromaddr, a.execaddr)
		if rightAccount.Balance < amount {
			elog.Error("limit check right balance", "addr", a.fromaddr, "avail", rightAccount.Balance, "need", amount)
			return nil, et.ErrAssetBalance
		}
		return a.matchLimitOrder(payload, leftAssetDB, rightAssetDB, entrustAddr)

	}
	if payload.GetOp() == et.OpSell {
		amount := payload.GetAmount()
		leftAccount := leftAssetDB.LoadExecAccount(a.fromaddr, a.execaddr)
		if leftAccount.Balance < amount {
			elog.Error("limit check left balance", "addr", a.fromaddr, "avail", leftAccount.Balance, "need", amount)
			return nil, et.ErrAssetBalance
		}
		return a.matchLimitOrder(payload, leftAssetDB, rightAssetDB, entrustAddr)
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
		amount := CalcActualCost(et.OpBuy, balance, price, cfg.GetCoinPrecision())
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
		amount := CalcActualCost(et.OpSell, balance, price, cfg.GetCoinPrecision())
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

	order.Status = et.Revoked
	order.UpdateTime = a.blocktime
	order.RevokeHash = hex.EncodeToString(a.txhash)
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

// set the transaction logic method
// rules:
//1. The purchase price is higher than the market price, and the price is matched from low to high.
//2. Sell orders are matched at prices lower than market prices.
//3. Match the same prices on a first-in, first-out basis
func (a *Action) matchLimitOrder(payload *et.LimitOrder, leftAccountDB, rightAccountDB *account.DB, entrustAddr string) (*types.Receipt, error) {
	var (
		logs     []*types.ReceiptLog
		kvs      []*types.KeyValue
		priceKey string
		count    int
		taker    int32
		maker    int32
		minFee   int64
	)

	cfg := a.api.GetConfig()
	tCfg, err := ParseConfig(a.api.GetConfig(), a.height)
	if err != nil {
		elog.Error("executor/exchangedb matchLimitOrder.ParseConfig", "err", err)
		return nil, err
	}

	if cfg.IsDappFork(a.height, et.ExchangeX, et.ForkFix1) && tCfg.IsBankAddr(a.fromaddr) {
		return nil, et.ErrAddrIsBank
	}

	if !tCfg.IsFeeFreeAddr(a.fromaddr) {
		trade := tCfg.GetTrade(payload.GetLeftAsset(), payload.GetRightAsset())
		taker = trade.GetTaker()
		maker = trade.GetMaker()
		minFee = trade.GetMinFee()
	}

	or := &et.Order{
		OrderID:     a.GetIndex(),
		Value:       &et.Order_LimitOrder{LimitOrder: payload},
		Ty:          et.TyLimitOrderAction,
		Executed:    0,
		AVGPrice:    0,
		Balance:     payload.GetAmount(),
		Status:      et.Ordered,
		EntrustAddr: entrustAddr,
		Addr:        a.fromaddr,
		UpdateTime:  a.blocktime,
		Index:       a.GetIndex(),
		Rate:        maker,
		MinFee:      minFee,
		Hash:        hex.EncodeToString(a.txhash),
		CreateTime:  a.blocktime,
	}
	re := &et.ReceiptExchange{
		Order: or,
		Index: a.GetIndex(),
	}

	// A single transaction can match up to 100 historical orders, the maximum depth can be matched, the system has to protect itself
	// Iteration has listing price
	var done bool
	for {
		if count >= et.MaxMatchCount {
			break
		}
		if done {
			break
		}
		//Obtain price information of existing market listing
		marketDepthList, _ := QueryMarketDepth(a.localDB, payload.GetLeftAsset(), payload.GetRightAsset(), a.OpSwap(payload.Op), priceKey, et.Count)
		if marketDepthList == nil || len(marketDepthList.List) == 0 {
			break
		}
		for _, marketDepth := range marketDepthList.List {
			elog.Info("LimitOrder debug find depth", "height", a.height, "amount", marketDepth.Amount, "price", marketDepth.Price, "order-price", payload.GetPrice(), "op", a.OpSwap(payload.Op), "index", a.GetIndex())
			if count >= et.MaxMatchCount {
				done = true
				break
			}
			if payload.Op == et.OpBuy && marketDepth.Price > payload.GetPrice() {
				done = true
				break
			}
			if payload.Op == et.OpSell && marketDepth.Price < payload.GetPrice() {
				done = true
				break
			}

			var hasOrder = false
			var orderKey string
			for {
				if count >= et.MaxMatchCount {
					done = true
					break
				}
				orderList, err := findOrderIDListByPrice(a.localDB, payload.GetLeftAsset(), payload.GetRightAsset(), marketDepth.Price, a.OpSwap(payload.Op), et.ListASC, orderKey)
				if orderList != nil && !hasOrder {
					hasOrder = true
				}
				if err != nil {
					if err == types.ErrNotFound {
						break
					}
					elog.Error("findOrderIDListByPrice error", "height", a.height, "symbol", payload.GetLeftAsset().Symbol, "price", marketDepth.Price, "op", a.OpSwap(payload.Op), "error", err)
					return nil, err
				}
				for _, matchorder := range orderList.List {
					if count >= et.MaxMatchCount {
						done = true
						break
					}
					// Check the order status
					order, err := findOrderByOrderID(a.statedb, a.localDB, matchorder.GetOrderID())
					if err != nil || order.Status != et.Ordered {
						if len(orderList.List) == 1 {
							hasOrder = true
						}
						continue
					}
					log, kv, err := a.matchModel(leftAccountDB, rightAccountDB, payload, order, or, re, tCfg.GetFeeAddr(), taker) // payload, or redundant
					if err != nil {
						if err == types.ErrNoBalance {
							elog.Warn("matchModel RevokeOrder", "height", a.height, "orderID", order.GetOrderID(), "payloadID", or.GetOrderID(), "error", err)
							continue
						}
						return nil, err
					}
					logs = append(logs, log...)
					kvs = append(kvs, kv...)
					if or.Status == et.Completed {
						receiptlog := &types.ReceiptLog{Ty: et.TyLimitOrderLog, Log: types.Encode(re)}
						logs = append(logs, receiptlog)
						receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
						return receipts, nil
					}
					// match depth count
					count = count + 1
				}
				if orderList.PrimaryKey == "" {
					break
				}
				orderKey = orderList.PrimaryKey
			}
			if !hasOrder {
				var matchorder et.Order
				matchorder.UpdateTime = a.blocktime
				matchorder.Status = et.Completed
				matchorder.Balance = 0
				matchorder.Executed = 0
				matchorder.AVGPrice = marketDepth.Price
				elog.Info("make empty match to del depth", "height", a.height, "price", marketDepth.Price, "amount", marketDepth.Amount)
				re.MatchOrders = append(re.MatchOrders, &matchorder)
			}
		}

		if marketDepthList.PrimaryKey == "" {
			break
		}
		priceKey = marketDepthList.PrimaryKey
	}

	//Outstanding orders require freezing of the remaining unclosed funds
	if payload.Op == et.OpBuy {
		amount := CalcActualCost(et.OpBuy, or.Balance, payload.Price, cfg.GetCoinPrecision())
		receipt, err := rightAccountDB.ExecFrozen(a.fromaddr, a.execaddr, amount)
		if err != nil {
			elog.Error("LimitOrder.ExecFrozen OpBuy", "addr", a.fromaddr, "amount", amount, "err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
	}
	if payload.Op == et.OpSell {
		amount := CalcActualCost(et.OpSell, or.Balance, payload.Price, cfg.GetCoinPrecision())
		receipt, err := leftAccountDB.ExecFrozen(a.fromaddr, a.execaddr, amount)
		if err != nil {
			elog.Error("LimitOrder.ExecFrozen OpSell", "addr", a.fromaddr, "amount", amount, "err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
	}
	kvs = append(kvs, a.GetKVSet(or)...)
	re.Order = or
	receiptlog := &types.ReceiptLog{Ty: et.TyLimitOrderLog, Log: types.Encode(re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) matchModel(leftAccountDB, rightAccountDB *account.DB, payload *et.LimitOrder, matchorder *et.Order, or *et.Order, re *et.ReceiptExchange, feeAddr string, taker int32) ([]*types.ReceiptLog, []*types.KeyValue, error) {
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

	cfg := a.api.GetConfig()
	var receipt *types.Receipt
	var err error
	if payload.Op == et.OpSell {
		//Transfer of frozen assets
		amount := CalcActualCost(matchorder.GetLimitOrder().Op, matched, matchorder.GetLimitOrder().Price, cfg.GetCoinPrecision())
		if matchorder.Addr != a.fromaddr {
			receipt, err = rightAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, amount)
		} else {
			receipt, err = rightAccountDB.ExecActive(a.fromaddr, a.execaddr, amount)
		}
		if err != nil {
			elog.Error("matchModel.ExecTransferFrozen", "from", matchorder.Addr, "to", a.fromaddr, "amount", amount, "err", err)
			return nil, nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)

		//Charge fee
		activeFee := calcMtfFee(amount, taker) //Transaction fee of the active party
		if activeFee != 0 {
			receipt, err = rightAccountDB.ExecTransfer(a.fromaddr, feeAddr, a.execaddr, activeFee)
			if err != nil {
				elog.Error("matchModel.ExecTransfer sell", "from", a.fromaddr, "to", feeAddr,
					"amount", amount, "rate", taker, "activeFee", activeFee, "err", err.Error())
				return nil, nil, err
			}
			or.DigestedFee += activeFee
			logs = append(logs, receipt.Logs...)
			kvs = append(kvs, receipt.KV...)
		}

		//The settlement of the corresponding assets for the transaction to be concluded
		amount = CalcActualCost(payload.Op, matched, matchorder.GetLimitOrder().Price, cfg.GetCoinPrecision())
		if a.fromaddr != matchorder.Addr {
			receipt, err = leftAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, amount)
			if err != nil {
				elog.Error("matchModel.ExecTransfer", "from", a.fromaddr, "to", matchorder.Addr, "amount", amount, "err", err.Error())
				return nil, nil, err
			}
			logs = append(logs, receipt.Logs...)
			kvs = append(kvs, receipt.KV...)
		}

		//Charge fee
		passiveFee := calcMtfFee(amount, matchorder.GetRate()) //Passive transaction fees
		if passiveFee != 0 {
			receipt, err = leftAccountDB.ExecTransfer(matchorder.Addr, feeAddr, a.execaddr, passiveFee)
			if err != nil {
				elog.Error("matchModel.ExecTransfer sell", "from", matchorder.Addr, "to", feeAddr,
					"amount", amount, "rate", matchorder.GetRate(), "passiveFee", passiveFee, "err", err.Error())
				return nil, nil, err
			}
			matchorder.DigestedFee += passiveFee
			logs = append(logs, receipt.Logs...)
			kvs = append(kvs, receipt.KV...)
		}

		or.AVGPrice = caclAVGPrice(or, matchorder.GetLimitOrder().Price, matched)
		//Calculate the average transaction price
		matchorder.AVGPrice = caclAVGPrice(matchorder, matchorder.GetLimitOrder().Price, matched)
	}
	if payload.Op == et.OpBuy {
		amount := CalcActualCost(matchorder.GetLimitOrder().Op, matched, matchorder.GetLimitOrder().Price, cfg.GetCoinPrecision())
		if a.fromaddr != matchorder.Addr {
			receipt, err = leftAccountDB.ExecTransferFrozen(matchorder.Addr, a.fromaddr, a.execaddr, amount)
		} else {
			receipt, err = leftAccountDB.ExecActive(a.fromaddr, a.execaddr, amount)
		}
		if err != nil {
			elog.Error("matchModel.ExecTransferFrozen2", "from", matchorder.Addr, "to", a.fromaddr, "amount", amount, "err", err.Error())
			return nil, nil, err
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)

		activeFee := calcMtfFee(amount, taker)
		if activeFee != 0 {
			receipt, err = leftAccountDB.ExecTransfer(a.fromaddr, feeAddr, a.execaddr, activeFee)
			if err != nil {
				elog.Error("matchModel.ExecTransfer buy", "from", a.fromaddr, "to", feeAddr,
					"amount", amount, "rate", taker, "activeFee", activeFee, "err", err.Error())
				return nil, nil, err
			}
			or.DigestedFee += activeFee
			logs = append(logs, receipt.Logs...)
			kvs = append(kvs, receipt.KV...)
		}

		amount = CalcActualCost(payload.Op, matched, matchorder.GetLimitOrder().Price, cfg.GetCoinPrecision())
		if a.fromaddr != matchorder.Addr {
			receipt, err = rightAccountDB.ExecTransfer(a.fromaddr, matchorder.Addr, a.execaddr, amount)
			if err != nil {
				elog.Error("matchModel.ExecTransfer2", "from", a.fromaddr, "to", matchorder.Addr, "amount", amount, "err", err.Error())
				return nil, nil, err
			}
			logs = append(logs, receipt.Logs...)
			kvs = append(kvs, receipt.KV...)
		}

		passiveFee := calcMtfFee(amount, matchorder.GetRate())
		if passiveFee != 0 {
			receipt, err = rightAccountDB.ExecTransfer(matchorder.Addr, feeAddr, a.execaddr, passiveFee)
			if err != nil {
				elog.Error("matchModel.ExecTransfer buy", "from", matchorder.Addr, "to", feeAddr,
					"amount", amount, "rate", matchorder.GetRate(), "passiveFee", passiveFee, "err", err.Error())
				return nil, nil, err
			}
			matchorder.DigestedFee += passiveFee
			logs = append(logs, receipt.Logs...)
			kvs = append(kvs, receipt.KV...)
		}

		or.AVGPrice = caclAVGPrice(or, matchorder.GetLimitOrder().Price, matched)
		matchorder.AVGPrice = caclAVGPrice(matchorder, matchorder.GetLimitOrder().Price, matched)
	}

	matchorder.UpdateTime = a.blocktime

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

// Query the status database according to the order number
// Localdb deletion sequence: delete the cache in real time first, and modify the DB uniformly during block generation.
// The cache data will be deleted. However, if the cache query fails, the deleted data can still be queried in the DB
func findOrderByOrderID(statedb dbm.KV, localdb dbm.KV, orderID int64) (*et.Order, error) {
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
	order.Executed = order.GetLimitOrder().Amount - order.Balance
	return &order, nil
}

func findOrderIDListByPrice(localdb dbm.KV, left, right *et.Asset, price int64, op, direction int32, primaryKey string) (*et.OrderList, error) {
	table := NewMarketOrderTable(localdb)
	prefix := []byte(fmt.Sprintf("%s:%s:%d:%016d", left.GetSymbol(), right.GetSymbol(), op, price))

	var rows []*tab.Row
	var err error
	if primaryKey == "" { // First query, the default display of the latest transaction record
		rows, err = table.ListIndex("market_order", prefix, nil, et.Count, direction)
	} else {
		rows, err = table.ListIndex("market_order", prefix, []byte(primaryKey), et.Count, direction)
	}
	if err != nil {
		if primaryKey == "" {
			elog.Error("findOrderIDListByPrice.", "left", left.Symbol, "right", right.Symbol, "price", price, "err", err.Error())
		}
		return nil, err
	}
	var orderList et.OrderList
	for _, row := range rows {
		order := row.Data.(*et.Order)
		// The replacement has been done
		order.Executed = order.GetLimitOrder().Amount - order.Balance
		orderList.List = append(orderList.List, order)
	}
	// Set the primary key index
	if len(rows) == int(et.Count) {
		orderList.PrimaryKey = string(rows[len(rows)-1].Primary)
	}
	return &orderList, nil
}

//Direction
//Buying depth is in reverse order by price, from high to low
func Direction(op int32) int32 {
	if op == et.OpBuy {
		return et.ListDESC
	}
	return et.ListASC
}

//QueryMarketDepth 这里primaryKey当作主键索引来用，
//The first query does not need to fill in the value, pay according to the price from high to low, selling orders according to the price from low to high query
func QueryMarketDepth(localdb dbm.KV, left, right *et.Asset, op int32, primaryKey string, count int32) (*et.MarketDepthList, error) {
	table := NewMarketDepthTable(localdb)
	prefix := []byte(fmt.Sprintf("%s:%s:%d", left.GetSymbol(), right.GetSymbol(), op))
	if count == 0 {
		count = et.Count
	}
	var rows []*tab.Row
	var err error
	if primaryKey == "" { // First query, the default display of the latest transaction record
		rows, err = table.ListIndex("price", prefix, nil, count, Direction(op))
	} else {
		rows, err = table.ListIndex("price", prefix, []byte(primaryKey), count, Direction(op))
	}
	if err != nil {
		return nil, err
	}

	var list et.MarketDepthList
	for _, row := range rows {
		list.List = append(list.List, row.Data.(*et.MarketDepth))
	}
	if len(rows) == int(count) {
		list.PrimaryKey = string(rows[len(rows)-1].Primary)
	}
	return &list, nil
}

//QueryMarketDepth 查询市场深度(买卖一起返回,不做聚合)
func QueryAllMarketDepth(localdb dbm.KV, left, right *et.Asset, count int32) (*et.MarketAllDepth, error) {
	var list et.MarketAllDepth

	bid, err := QueryMarketDepth(localdb, left, right, et.OpBuy, "", count)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	if bid != nil {
		list.Bids = bid.List
	}

	ask, err := QueryMarketDepth(localdb, left, right, et.OpSell, "", count)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	if ask != nil {
		list.Asks = ask.List
	}

	return &list, nil
}

func QueryAllDept(localdb dbm.KV, left, right *et.Asset, op, count int32, depth int64) (*et.MarketAllDepth, error) {
	var list et.MarketAllDepth

	bid, err := QueryBidDepth(localdb, left, right, count, depth)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	if bid != nil && op != et.OpSell {
		list.Bids = bid.List
	}

	ask, err := QueryAskDepth(localdb, left, right, count, depth)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	if ask != nil && op != et.OpBuy {
		list.Asks = ask.List
	}
	return &list, nil
}
func QueryBidDepth(localdb dbm.KV, left, right *et.Asset, count int32, depth int64) (*et.MarketDepthList, error) {
	table := NewMarketDepthTable(localdb)
	prefix := []byte(fmt.Sprintf("%s:%s:%d", left.GetSymbol(), right.GetSymbol(), et.OpBuy))
	var rows []*tab.Row
	var err error

	rows, err = table.ListIndex("price", prefix, nil, 20, Direction(et.OpBuy))
	if err != nil {
		return nil, err
	}

	var list et.MarketDepthList
	first := (rows[0].Data.(*et.MarketDepth).Price) / depth
	for index := 0; index < int(count); index++ {
		if (first - int64(index)) < 0 {
			break
		}
		marketDepth := &et.MarketDepth{
			LeftAsset:  left,
			RightAsset: right,
			Price:      (first - int64(index)) * depth,
			Amount:     0,
			Op:         et.OpBuy,
		}
		list.List = append(list.List, marketDepth)
	}

	var amount, price, nPrice, temp int64
	for {
		for _, row := range rows {
			amount = row.Data.(*et.MarketDepth).Amount
			price = row.Data.(*et.MarketDepth).Price
			if len(list.List) > 0 {
				nPrice = list.List[len(list.List)-1].Price
			}
			if price < nPrice {
				return &list, nil
			}
			for {
				if price >= list.List[temp].Price {
					list.List[temp].Amount += amount
					list.PrimaryKey = string(row.Primary)
					break
				}
				temp++
			}
		}
		rows, err = table.ListIndex("price", prefix, []byte(list.PrimaryKey), 20, Direction(et.OpBuy))
		if err == types.ErrNotFound {
			return &list, nil
		}
		if err != nil {
			return nil, err
		}
	}
}
func QueryAskDepth(localdb dbm.KV, left, right *et.Asset, count int32, depth int64) (*et.MarketDepthList, error) {
	table := NewMarketDepthTable(localdb)
	prefix := []byte(fmt.Sprintf("%s:%s:%d", left.GetSymbol(), right.GetSymbol(), et.OpSell))
	var rows []*tab.Row
	var err error

	rows, err = table.ListIndex("price", prefix, nil, 20, Direction(et.OpSell))
	if err != nil {
		return nil, err
	}

	var list et.MarketDepthList
	first := math.Ceil(float64(rows[0].Data.(*et.MarketDepth).Price) / float64(depth))
	for index := 0; index < int(count); index++ {
		marketDepth := &et.MarketDepth{
			LeftAsset:  left,
			RightAsset: right,
			Price:      (int64(first) + int64(index)) * depth,
			Amount:     0,
			Op:         et.OpSell,
		}
		list.List = append(list.List, marketDepth)
	}

	var amount, price, temp int64
	for {
		for _, row := range rows {
			amount = row.Data.(*et.MarketDepth).Amount
			price = row.Data.(*et.MarketDepth).Price
			if price > list.List[len(list.List)-1].Price {
				return &list, nil
			}
			for {
				if price <= list.List[temp].Price {
					list.List[temp].Amount += amount
					list.PrimaryKey = string(row.Primary)
					break
				}
				temp++
			}
		}
		rows, err = table.ListIndex("price", prefix, []byte(list.PrimaryKey), 20, Direction(et.OpSell))
		if err == types.ErrNotFound {
			return &list, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

//QueryHistoryOrderList Only the order information is returned
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
	if primaryKey == "" { // First query, the default display of the latest transaction record
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
		// This table contains orders completed,revoked so filtering is required
		if order.Status == et.Revoked {
			continue
		}
		// The replacement has been done
		order.Executed = order.GetLimitOrder().Amount - order.Balance
		orderList.List = append(orderList.List, order)
		if len(orderList.List) == int(count) {
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

//QueryOrderList Displays the latest by default
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
	if primaryKey == "" {
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
		order.Executed = order.GetLimitOrder().Amount - order.Balance
		orderList.List = append(orderList.List, order)
	}
	if len(rows) == int(count) {
		orderList.PrimaryKey = string(rows[len(rows)-1].Primary)
	}
	return &orderList, nil
}

func queryMarketDepth(marketTable *tab.Table, left, right *et.Asset, op int32, price int64) (*et.MarketDepth, error) {
	primaryKey := []byte(fmt.Sprintf("%s:%s:%d:%016d", left.GetSymbol(), right.GetSymbol(), op, price))
	row, err := marketTable.GetData(primaryKey)
	if err != nil {
		// In localDB, delete is set to nil first and deleted last
		if err == types.ErrDecode && row == nil {
			err = types.ErrNotFound
		}
		return nil, err
	}
	return row.Data.(*et.MarketDepth), nil
}

//SafeMul Safe multiplication of large numbers, prevent overflow
func SafeMul(x, y, coinPrecision int64) int64 {
	res := big.NewInt(0).Mul(big.NewInt(x), big.NewInt(y))
	res = big.NewInt(0).Div(res, big.NewInt(coinPrecision))
	return res.Int64()
}

//Calculate the average transaction price
func caclAVGPrice(order *et.Order, price int64, amount int64) int64 {
	x := big.NewInt(0).Mul(big.NewInt(order.AVGPrice), big.NewInt(order.GetLimitOrder().Amount-order.GetBalance()))
	y := big.NewInt(0).Mul(big.NewInt(price), big.NewInt(amount))
	total := big.NewInt(0).Add(x, y)
	div := big.NewInt(0).Add(big.NewInt(order.GetLimitOrder().Amount-order.GetBalance()), big.NewInt(amount))
	avg := big.NewInt(0).Div(total, div)
	return avg.Int64()
}

//计Calculation fee
func calcMtfFee(cost int64, rate int32) int64 {
	fee := big.NewInt(0).Mul(big.NewInt(cost), big.NewInt(int64(rate)))
	fee = big.NewInt(0).Div(fee, big.NewInt(types.DefaultCoinPrecision))
	return fee.Int64()
}

func ParseConfig(cfg *types.Chain33Config, height int64) (*et.Econfig, error) {
	banks, err := ParseStrings(cfg, "banks", height)
	if err != nil || len(banks) == 0 {
		return nil, err
	}

	robots, err := ParseStrings(cfg, "robots", height)
	if err != nil || len(banks) == 0 {
		return nil, err
	}
	robotMap := make(map[string]bool)
	for _, v := range robots {
		robotMap[v] = true
	}

	coins, err := ParseCoins(cfg, "coins", height)
	if err != nil {
		return nil, err
	}
	exchanges, err := ParseSymbols(cfg, "exchanges", height)
	if err != nil {
		return nil, err
	}
	return &et.Econfig{
		Banks:     banks,
		RobotMap:  robotMap,
		Coins:     coins,
		Exchanges: exchanges,
	}, nil
}

func ParseStrings(cfg *types.Chain33Config, tradeKey string, height int64) (ret []string, err error) {
	val, err := cfg.MG(et.MverPrefix+"."+tradeKey, height)
	if err != nil {
		return nil, err
	}

	datas, ok := val.([]interface{})
	if !ok {
		elog.Error("invalid val", "val", val, "key", tradeKey)
		return nil, et.ErrCfgFmt
	}

	for _, v := range datas {
		one, ok := v.(string)
		if !ok {
			elog.Error("invalid one", "one", one, "key", tradeKey)
			return nil, et.ErrCfgFmt
		}
		ret = append(ret, one)
	}
	return
}

func ParseCoins(cfg *types.Chain33Config, tradeKey string, height int64) (coins []et.CoinCfg, err error) {
	coins = make([]et.CoinCfg, 0)

	val, err := cfg.MG(et.MverPrefix+"."+tradeKey, height)
	if err != nil {
		return nil, err
	}

	datas, ok := val.([]interface{})
	if !ok {
		elog.Error("invalid coins", "val", val, "type", reflect.TypeOf(val))
		return nil, et.ErrCfgFmt
	}

	for _, e := range datas {
		v, ok := e.(map[string]interface{})
		if !ok {
			elog.Error("invalid coins one", "one", v, "key", tradeKey)
			return nil, et.ErrCfgFmt
		}

		coin := et.CoinCfg{
			Coin:   v["coin"].(string),
			Execer: v["execer"].(string),
			Name:   v["name"].(string),
		}
		coins = append(coins, coin)
	}
	return
}

func ParseSymbols(cfg *types.Chain33Config, tradeKey string, height int64) (symbols map[string]*et.Trade, err error) {
	symbols = make(map[string]*et.Trade)

	val, err := cfg.MG(et.MverPrefix+"."+tradeKey, height)
	if err != nil {
		return nil, err
	}

	datas, ok := val.([]interface{})
	if !ok {
		elog.Error("invalid Symbols", "val", val, "type", reflect.TypeOf(val))
		return nil, et.ErrCfgFmt
	}

	for _, e := range datas {
		v, ok := e.(map[string]interface{})
		if !ok {
			elog.Error("invalid Symbols one", "one", v, "key", tradeKey)
			return nil, et.ErrCfgFmt
		}

		symbol := v["symbol"].(string)
		symbols[symbol] = &et.Trade{
			Symbol:       symbol,
			PriceDigits:  int32(formatInterface(v["priceDigits"])),
			AmountDigits: int32(formatInterface(v["amountDigits"])),
			Taker:        int32(formatInterface(v["taker"])),
			Maker:        int32(formatInterface(v["maker"])),
			MinFee:       formatInterface(v["minFee"]),
		}
	}
	return
}
func formatInterface(data interface{}) int64 {
	switch data.(type) {
	case int64:
		return data.(int64)
	case int32:
		return int64(data.(int32))
	case int:
		return int64(data.(int))
	default:
		return 0
	}
}
