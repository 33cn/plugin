package spot

import (
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

// market depth:
// price - list
// order - list for each price
type matcher struct {
	localdb  dbm.KV
	statedb  dbm.KV
	api      client.QueueProtocolAPI
	dbprefix et.DBprefix

	matchCount int
	maxMatch   int
	done       bool

	// price list
	pricekey     string
	endPriceList bool

	// order list
	lastOrderPrice int64
	orderKey       string
	endOrderList   bool
}

func newMatcher(statedb, localdb dbm.KV, api client.QueueProtocolAPI, dbprefix et.DBprefix) *matcher {
	return &matcher{
		localdb:  localdb,
		statedb:  statedb,
		api:      api,
		dbprefix: dbprefix,

		pricekey:     "",
		matchCount:   0,
		maxMatch:     et.MaxMatchCount,
		done:         false,
		endPriceList: false,
	}
}

// set the transaction logic method
// rules:
//1. The purchase price is higher than the market price, and the price is matched from low to high.
//2. Sell orders are matched at prices lower than market prices.
//3. Match the same prices on a first-in, first-out basis

func (m *matcher) isDone() bool {
	return (m.done || m.matchCount >= m.maxMatch)
}

func (m *matcher) recordMatchCount() {
	m.matchCount = m.matchCount + 1
	if m.matchCount >= m.maxMatch {
		m.done = true
	}
}

func (m *matcher) priceDone(op int32, price int64, marketDepth *et.SpotMarketDepth) bool {
	if priceDone(op, price, marketDepth) {
		m.done = true
		return true
	}
	return false
}

func priceDone(op int32, price int64, marketDepth *et.SpotMarketDepth) bool {
	if op == et.OpBuy && marketDepth.Price > price {
		return true
	}
	if op == et.OpSell && marketDepth.Price < price {
		return true
	}
	return false
}

func (m *matcher) QueryMarketDepth(left, right *et.ZkAsset, op int32) (*et.SpotMarketDepthList, error) {
	if m.endPriceList {
		m.done = true
		return nil, nil
	}
	marketTable := NewMarketDepthTable(m.localdb, m.dbprefix)
	marketDepthList, _ := queryMarketDepthList(marketTable, SymbolStr(left), SymbolStr(right), OpSwap(op), m.pricekey, et.Count)
	if marketDepthList == nil || len(marketDepthList.List) == 0 {
		return nil, nil
	}

	// reatch the last price list
	if marketDepthList.PrimaryKey == "" {
		m.endPriceList = true
	}

	// set next key
	m.pricekey = marketDepthList.PrimaryKey
	return marketDepthList, nil
}

func (m *matcher) findOrderIDListByPrice(left, right *et.ZkAsset, op int32, marketDepth *et.SpotMarketDepth) (*et.SpotOrderList, error) {
	direction := et.ListASC // 撮合按时间先后顺序
	price := marketDepth.Price
	if price != m.lastOrderPrice {
		m.orderKey = ""
		m.endOrderList = false
	}

	orderLdb := newOrderLRepo(m.localdb, m.dbprefix)
	orderList, err := orderLdb.findOrderIDListByPrice(SymbolStr(left), SymbolStr(right), price, OpSwap(op), direction, m.orderKey)
	if err != nil {
		if err == types.ErrNotFound {
			return &et.SpotOrderList{List: []*et.SpotOrder{}, PrimaryKey: ""}, nil
		}
		elog.Error("findOrderIDListByPrice error" /*"height", a.height, */, "symbol", SymbolStr(left), "price", marketDepth.Price, "op", OpSwap(op), "error", err)
		return nil, err
	}
	// reatch the last order list for price
	if orderList.PrimaryKey == "" {
		m.endOrderList = true
	}

	// set next key
	m.orderKey = orderList.PrimaryKey
	return orderList, nil
}

func (m *matcher) isEndOrderList(price int64) bool {
	return price == m.lastOrderPrice && m.endOrderList
}

func (matcher1 *matcher) MatchOrder(order *Order, taker *SpotTrader, orderdb *orderSRepo, s *Spot) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	op := order.GetOp()
	price := order.GetPrice()
	left, right := order.GetAsset()
	for {
		if matcher1.isDone() {
			break
		}

		//Obtain price information of existing market listing

		marketDepthList, _ := matcher1.QueryMarketDepth(left, right, op)
		if marketDepthList == nil || len(marketDepthList.List) == 0 {
			break
		}
		for _, marketDepth := range marketDepthList.List {
			elog.Info("LimitOrder debug find depth", "amount", marketDepth.Amount, "price", marketDepth.Price, "order-price", price, "op", OpSwap(op), "index", taker.order.order.GetOrderID())
			if matcher1.isDone() || matcher1.priceDone(op, price, marketDepth) {
				break
			}

			for {
				if matcher1.isDone() {
					break
				}

				orderList, err := matcher1.findOrderIDListByPrice(left, right, op, marketDepth)
				if err != nil || orderList == nil || len(orderList.List) == 0 {
					break
				}
				// got orderlist to trade
				for _, matchorder := range orderList.List {
					if matcher1.isDone() {
						break
					}
					// Check the order status
					order, err := orderdb.findOrderBy(matchorder.GetOrderID())
					if err != nil || order.Status != et.Ordered {
						continue
					}
					orderx := NewOrder(order, orderdb)
					log, kv, err := taker.matchModel(orderx, matcher1.statedb, s)
					if err != nil {
						elog.Error("matchModel", "height", "orderID", order.GetOrderID(), "payloadID", taker.order.order.GetOrderID(), "error", err)
						return nil, err
					}
					logs = append(logs, log...)
					kvs = append(kvs, kv...)
					if taker.order.order.Status == et.Completed {
						matcher1.done = true
						break
					}
					// match depth count
					matcher1.recordMatchCount()
				}
				if matcher1.isEndOrderList(marketDepth.Price) {
					break
				}
			}
		}
	}

	kvs = append(kvs, orderdb.GetOrderKvSet(taker.order.order)...)
	receiptlog := &types.ReceiptLog{Ty: et.TyLimitOrderLog, Log: types.Encode(taker.matches)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}
