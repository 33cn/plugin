package executor

import (
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

//查询市场深度
func (s *exchange) Query_QueryMarketDepth(in *et.QueryMarketDepth) (types.Message, error) {
	if !CheckCount(in.Count) {
		return nil, et.ErrCount
	}
	if in.Count == 0 {
		in.Count = 10
	}
	if !CheckExchangeAsset(in.LeftAsset, in.RightAsset) {
		return nil, et.ErrAsset
	}

	if !CheckOp(in.Op) {
		return nil, et.ErrAssetOp
	}
	return QueryMarketDepth(s.GetLocalDB(), in.LeftAsset, in.RightAsset, in.Op, in.Price, in.Count)
}

//查询已经完成得订单
func (s *exchange) Query_QueryCompletedOrderList(in *et.QueryCompletedOrderList) (types.Message, error) {
	if !CheckExchangeAsset(in.LeftAsset, in.RightAsset) {
		return nil, et.ErrAsset
	}
	if in.Count > 20 {
		return nil, et.ErrCount
	}

	if !CheckDirection(in.Direction) {
		return nil, et.ErrDirection
	}
	return QueryCompletedOrderList(s.GetLocalDB(), s.GetStateDB(), in.LeftAsset, in.RightAsset, in.Index, in.Count, in.Direction)
}

//根据orderID查询订单信息
func (s *exchange) Query_QueryOrder(in *et.QueryOrder) (types.Message, error) {
	if in.OrderID == "" {
		return nil, et.ErrOrderID
	}
	return findOrderByOrderID(s.GetStateDB(), in.OrderID)
}

//根据订单状态，查询订单信息（这里面包含所有交易对）
func (s *exchange) Query_QueryOrderList(in *et.QueryOrderList) (types.Message, error) {
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
	return QueryOrderList(s.GetLocalDB(), s.GetStateDB(), in.Address, in.Status, in.Count, in.Direction, in.Index)
}
