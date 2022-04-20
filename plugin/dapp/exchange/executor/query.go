package executor

import (
	"math"

	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

//查询市场深度
func (e *exchange) Query_QueryMarketDepth(in *et.QueryMarketDepth) (types.Message, error) {
	if !CheckCount(in.Count) {
		return nil, et.ErrCount
	}
	if !CheckExchangeAsset(e.GetAPI().GetConfig().GetCoinExec(), in.LeftAsset, in.RightAsset) {
		return nil, et.ErrAsset
	}

	if !CheckOp(in.Op) {
		return nil, et.ErrAssetOp
	}
	return QueryMarketDepth(e.GetLocalDB(), in.LeftAsset, in.RightAsset, in.Op, in.PrimaryKey, in.Count)
}

//查询市场深度(分档位)
func (e *exchange) Query_QueryDepth(in *et.QueryMarketDepth) (types.Message, error) {
	if in.Count == 0 {
		in.Count = et.Count
	}
	if !Check5Count(in.Count) {
		return nil, et.Err5Count
	}
	if !CheckExchangeAsset(e.GetAPI().GetConfig().GetCoinExec(), in.LeftAsset, in.RightAsset) {
		return nil, et.ErrAsset
	}

	tCfg, err := ParseConfig(e.GetAPI().GetConfig(), e.GetHeight())
	if err != nil {
		elog.Error("Query_QueryDepth.ParseConfig", "err", err)
		return nil, err
	}

	priceDigits := tCfg.GetTrade(in.LeftAsset, in.RightAsset).GetPriceDigits()

	if !CheckDepth(in.Depth, priceDigits) {
		return nil, et.ErrAssetDepth
	} else if in.Depth == 1 {
		return QueryAllMarketDepth(e.GetLocalDB(), in.LeftAsset, in.RightAsset, in.Count)
	}

	return QueryAllDept(e.GetLocalDB(), in.LeftAsset, in.RightAsset, in.Op, in.Count, int64(1e10/math.Pow10(int(priceDigits-(in.Depth-1)+2))))
}

//查询已经完成得订单
func (e *exchange) Query_QueryHistoryOrderList(in *et.QueryHistoryOrderList) (types.Message, error) {
	if !CheckExchangeAsset(e.GetAPI().GetConfig().GetCoinExec(), in.LeftAsset, in.RightAsset) {
		return nil, et.ErrAsset
	}
	if !CheckCount(in.Count) {
		return nil, et.ErrCount
	}

	if !CheckDirection(in.Direction) {
		return nil, et.ErrDirection
	}
	return QueryHistoryOrderList(e.GetLocalDB(), in.LeftAsset, in.RightAsset, in.PrimaryKey, in.Count, in.Direction)
}

//根据orderID查询订单信息
func (e *exchange) Query_QueryOrder(in *et.QueryOrder) (types.Message, error) {
	if in.OrderID == 0 {
		return nil, et.ErrOrderID
	}
	return findOrderByOrderID(e.GetStateDB(), e.GetLocalDB(), in.OrderID)
}

//根据订单状态，查询订单信息（这里面包含所有交易对）
func (e *exchange) Query_QueryOrderList(in *et.QueryOrderList) (types.Message, error) {
	if !CheckStatus(in.Status) {
		return nil, et.ErrStatus
	}
	if !CheckCount(in.Count) {
		return nil, et.ErrCount
	}

	if !CheckDirection(in.Direction) {
		return nil, et.ErrDirection
	}

	if in.Address == "" {
		return nil, et.ErrAddr
	}
	return QueryOrderList(e.GetLocalDB(), in.Address, in.Status, in.Count, in.Direction, in.PrimaryKey)
}
