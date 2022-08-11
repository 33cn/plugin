package spot

import (
	"fmt"

	dbm "github.com/33cn/chain33/common/db"
	tab "github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

type assetPair struct {
	left  *et.ZkAsset
	right *et.ZkAsset
}

func newAssetPair1(left, right uint64) *assetPair {
	//ZkAssetidL := et.ZkAssetid{}
	return &assetPair{
		left: &et.ZkAsset{
			Ty:    et.AssetType_L1Erc20,
			Value: &et.ZkAsset_ZkAssetid{ZkAssetid: left},
		},

		right: &et.ZkAsset{
			Ty: et.AssetType_L1Erc20,
			Value: &et.ZkAsset_ZkAssetid{
				ZkAssetid: right,
			},
		},
	}
}

func (pair *assetPair) l1historyPrefix() []byte {
	return []byte(fmt.Sprintf("%08d:%08d", pair.left.GetZkAssetid(), pair.right.GetZkAssetid()))
}

func (pair *assetPair) l1MarketDepthPrefix(op int32, price int64) []byte {
	left := pair.left.GetZkAssetid()
	right := pair.right.GetZkAssetid()
	return []byte(fmt.Sprintf("%08d:%08d:%d:%016d", left, right, op, price))
}

func SymbolStr(a *et.ZkAsset) string {
	switch a.Ty {
	case et.AssetType_L1Erc20:
		return fmt.Sprintf("%08d", a.GetZkAssetid())
	case et.AssetType_TokenType:
		return a.GetTokenAsset().Symbol
	}
	return "unknow"
}

//QueryHistoryOrderList Only the order information is returned
func QueryHistoryOrderList(localdb dbm.KV, dbprefix et.DBprefix, in *et.SpotQueryHistoryOrderList) (types.Message, error) {
	left, right, primaryKey, count, direction := in.LeftAsset, in.RightAsset, in.PrimaryKey, in.Count, in.Direction

	pair := newAssetPair1(left, right)
	table := NewHistoryOrderTable(localdb, dbprefix)
	prefix := pair.l1historyPrefix()
	indexName := "name"
	if count == 0 {
		count = et.Count
	}
	var rows []*tab.Row
	var err error
	var orderList et.SpotOrderList
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
		order := row.Data.(*et.SpotOrder)
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
func QueryOrderList(localdb dbm.KV, dbprefix et.DBprefix, in *et.SpotQueryOrderList) (types.Message, error) {
	var table *tab.Table
	if in.Status == et.Completed || in.Status == et.Revoked {
		table = NewHistoryOrderTable(localdb, dbprefix)
	} else {
		table = NewMarketOrderTable(localdb, dbprefix)
	}
	prefix := []byte(fmt.Sprintf("%s:%d", in.Address, in.Status))
	indexName := "addr_status"
	count := in.Count
	if count == 0 {
		count = et.Count
	}
	var rows []*tab.Row
	var err error
	if in.PrimaryKey == "" {
		rows, err = table.ListIndex(indexName, prefix, nil, count, in.Direction)
	} else {
		rows, err = table.ListIndex(indexName, prefix, []byte(in.PrimaryKey), count, in.Direction)
	}
	if err != nil {
		elog.Error("QueryOrderList.", "addr", in.Address, "err", err.Error())
		return nil, err
	}
	var orderList et.SpotOrderList
	for _, row := range rows {
		order := row.Data.(*et.SpotOrder)
		order.Executed = order.GetLimitOrder().Amount - order.Balance
		orderList.List = append(orderList.List, order)
	}
	if len(rows) == int(count) {
		orderList.PrimaryKey = string(rows[len(rows)-1].Primary)
	}
	return &orderList, nil
}

func getMarketDepth(marketTable *tab.Table, left, right uint64, op int32, price int64) (*et.SpotMarketDepth, error) {
	pair := newAssetPair1(left, right)
	primaryKey := pair.l1MarketDepthPrefix(op, price)
	row, err := marketTable.GetData(primaryKey)
	if err != nil {
		// In localDB, delete is set to nil first and deleted last
		if err == types.ErrDecode && row == nil {
			err = types.ErrNotFound
		}
		return nil, err
	}
	return row.Data.(*et.SpotMarketDepth), nil
}

func updateIndex(marketTable, orderTable, historyTable *tab.Table, receipt *et.ReceiptSpotMatch) (kvs []*types.KeyValue) {
	elog.Info("updateIndex", "order.status", receipt.Order.Status)
	switch receipt.Order.Status {
	case et.Ordered:
		err := updateOrder(marketTable, orderTable, historyTable, receipt.GetOrder(), receipt.GetIndex())
		if err != nil {
			return nil
		}
		err = updateMatchedOrders(marketTable, orderTable, historyTable, receipt.GetOrder(), receipt.GetMatchOrders(), receipt.GetIndex())
		if err != nil {
			return nil
		}
	case et.Completed:
		err := updateOrder(marketTable, orderTable, historyTable, receipt.GetOrder(), receipt.GetIndex())
		if err != nil {
			return nil
		}
		err = updateMatchedOrders(marketTable, orderTable, historyTable, receipt.GetOrder(), receipt.GetMatchOrders(), receipt.GetIndex())
		if err != nil {
			return nil
		}
	case et.Revoked:
		err := updateOrder(marketTable, orderTable, historyTable, receipt.GetOrder(), receipt.GetIndex())
		if err != nil {
			return nil
		}
	}

	return
}

func updateOrder(marketTable, orderTable, historyTable *tab.Table, order *et.SpotOrder, index int64) error {
	left := order.GetLimitOrder().GetLeftAsset()
	right := order.GetLimitOrder().GetRightAsset()
	op := order.GetLimitOrder().GetOp()
	price := order.GetLimitOrder().GetPrice()
	switch order.Status {
	case et.Ordered:
		var markDepth et.SpotMarketDepth
		depth, err := getMarketDepth(marketTable, left, right, op, price)
		if err == types.ErrNotFound {
			markDepth.Price = price
			markDepth.LeftAsset = left
			markDepth.RightAsset = right
			markDepth.Op = op
			markDepth.Amount = order.Balance
		} else {
			markDepth.Price = price
			markDepth.LeftAsset = left
			markDepth.RightAsset = right
			markDepth.Op = op
			markDepth.Amount = depth.Amount + order.Balance
		}
		err = marketTable.Replace(&markDepth)
		if err != nil {
			elog.Error("updateIndex", "marketTable.Replace", err.Error())
		}
		err = orderTable.Replace(order)
		if err != nil {
			elog.Error("updateIndex", "orderTable.Replace", err.Error())
		}

	case et.Completed:
		err := historyTable.Replace(order)
		if err != nil {
			elog.Error("updateIndex", "historyTable.Replace", err.Error())
		}
	case et.Revoked:
		var marketDepth et.SpotMarketDepth
		depth, err := getMarketDepth(marketTable, left, right, op, price)
		if err == nil {
			marketDepth.Price = price
			marketDepth.LeftAsset = left
			marketDepth.RightAsset = right
			marketDepth.Op = op
			marketDepth.Amount = depth.Amount - order.Balance

			if marketDepth.Amount > 0 {
				err = marketTable.Replace(&marketDepth)
				if err != nil {
					elog.Error("updateIndex", "marketTable.Replace", err.Error())
				}
			}
			if marketDepth.Amount <= 0 {
				err = marketTable.DelRow(&marketDepth)
				if err != nil {
					elog.Error("updateIndex", "marketTable.DelRow", err.Error())
				}
			}
		}

		primaryKey := []byte(fmt.Sprintf("%022d", order.OrderID))
		err = orderTable.Del(primaryKey)
		if err != nil {
			elog.Error("updateIndex", "orderTable.Del", err.Error())
		}
		order.Status = et.Revoked
		order.Index = index
		err = historyTable.Replace(order)
		if err != nil {
			elog.Error("updateIndex", "historyTable.Replace", err.Error())
		}
	}
	return nil
}

func updateMatchedOrders(marketTable, orderTable, historyTable *tab.Table, order *et.SpotOrder, matchOrders []*et.SpotOrder, index int64) error {
	left := order.GetLimitOrder().GetLeftAsset()
	right := order.GetLimitOrder().GetRightAsset()
	op := order.GetLimitOrder().GetOp()
	if len(matchOrders) > 0 {
		cache := make(map[int64]int64)
		for i, matchOrder := range matchOrders {
			if matchOrder.Balance == 0 && matchOrder.Executed == 0 {
				var matchDepth et.SpotMarketDepth
				matchDepth.Price = matchOrder.AVGPrice
				matchDepth.LeftAsset = left
				matchDepth.RightAsset = right
				matchDepth.Op = OpSwap(op)
				matchDepth.Amount = 0
				err := marketTable.DelRow(&matchDepth)
				if err != nil && err != types.ErrNotFound {
					elog.Error("updateIndex", "marketTable.DelRow", err.Error())
				}
				continue
			}
			if matchOrder.Status == et.Completed {
				err := orderTable.DelRow(matchOrder)
				if err != nil {
					elog.Error("updateIndex", "orderTable.DelRow", err.Error())
				}
				matchOrder.Index = index + int64(i+1)
				err = historyTable.Replace(matchOrder)
				if err != nil {
					elog.Error("updateIndex", "historyTable.Replace", err.Error())
				}
			} else if matchOrder.Status == et.Ordered {
				err := orderTable.Replace(matchOrder)
				if err != nil {
					elog.Error("updateIndex", "orderTable.Replace", err.Error())
				}
			}
			executed := cache[matchOrder.GetLimitOrder().Price]
			executed = executed + matchOrder.Executed
			cache[matchOrder.GetLimitOrder().Price] = executed
		}

		for pr, executed := range cache {
			var matchDepth et.SpotMarketDepth
			depth, err := getMarketDepth(marketTable, left, right, OpSwap(op), pr)
			if err != nil {
				continue
			} else {
				matchDepth.Price = pr
				matchDepth.LeftAsset = left
				matchDepth.RightAsset = right
				matchDepth.Op = OpSwap(op)
				matchDepth.Amount = depth.Amount - executed
			}
			if matchDepth.Amount > 0 {
				err = marketTable.Replace(&matchDepth)
				if err != nil {
					elog.Error("updateIndex", "marketTable.Replace", err.Error())
				}
			}
			if matchDepth.Amount <= 0 {
				err = marketTable.DelRow(&matchDepth)
				if err != nil {
					elog.Error("updateIndex", "marketTable.DelRow", err.Error())
				}
			}
		}
	}
	return nil
}

type orderLRepo struct {
	table *tab.Table
}

func newOrderLRepo(localdb dbm.KV, p et.DBprefix) *orderLRepo {
	table := NewMarketOrderTable(localdb, p)
	return &orderLRepo{
		table: table,
	}
}

func (db *orderLRepo) pricePrefix(left, right string, price int64, op int32) []byte {
	return []byte(fmt.Sprintf("%s:%s:%d:%016d", left, right, op, price))
}

// asset to string as part of key
func (db *orderLRepo) findOrderIDListByPrice(leftStr, rightSrt string, price int64, op, direction int32, primaryKey string) (*et.SpotOrderList, error) {
	table := db.table
	prefix := db.pricePrefix(leftStr, rightSrt, price, op)

	var rows []*tab.Row
	var err error
	if primaryKey == "" { // First query, the default display of the latest transaction record
		rows, err = table.ListIndex("market_order", prefix, nil, et.Count, direction)
	} else {
		rows, err = table.ListIndex("market_order", prefix, []byte(primaryKey), et.Count, direction)
	}
	if err != nil {
		if primaryKey == "" {
			elog.Error("findOrderIDListByPrice.", "left", leftStr, "right", rightSrt, "price", price, "err", err.Error())
		}
		return nil, err
	}
	var orderList et.SpotOrderList
	for _, row := range rows {
		order := row.Data.(*et.SpotOrder)
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

func SymbolStrL1(i uint64) string {
	return fmt.Sprintf("%08d", i)
}

//QueryMarketDepth 这里primaryKey当作主键索引来用，
//The first query does not need to fill in the value, pay according to the price from high to low, selling orders according to the price from low to high query
func QueryMarketDepth(localdb dbm.KV, dbprefix et.DBprefix, in *et.SpotQueryMarketDepth) (*et.SpotMarketDepthList, error) {
	left, right, op := in.LeftAsset, in.RightAsset, in.Op
	count, primaryKey := in.Count, in.PrimaryKey
	marketTable := NewMarketDepthTable(localdb, dbprefix)

	return queryMarketDepthList(marketTable, SymbolStrL1(left), SymbolStrL1(right), op, primaryKey, count)
}

func queryMarketDepthList(table *tab.Table, left, right string, op int32, primaryKey string, count int32) (*et.SpotMarketDepthList, error) {
	prefix := []byte(fmt.Sprintf("%s:%s:%d", left, right, op))
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
		elog.Error("QueryMarketDepth.", "prefix", string(prefix), "left", left, "right", right, "err", err.Error())
		return nil, err
	}

	var list et.SpotMarketDepthList
	for _, row := range rows {
		list.List = append(list.List, row.Data.(*et.SpotMarketDepth))
	}
	if len(rows) == int(count) {
		list.PrimaryKey = string(rows[len(rows)-1].Primary)
	}
	return &list, nil
}
