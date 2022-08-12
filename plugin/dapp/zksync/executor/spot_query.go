package executor

import (
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/executor/spot"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

//查询市场深度
func (e *zksync) Query_QueryMarketDepth(in *et.SpotQueryMarketDepth) (types.Message, error) {
	if !et.CheckCount(in.Count) {
		return nil, et.ErrCount
	}
	if !et.CheckExchangeAsset(e.GetAPI().GetConfig().GetCoinExec(), in.LeftAsset, in.RightAsset) {
		return nil, et.ErrAsset
	}

	if !et.CheckOp(in.Op) {
		return nil, et.ErrAssetOp
	}
	return spot.QueryMarketDepth(e.GetLocalDB(), &dbprefix{}, in)
}

//查询已经完成得订单
func (e *zksync) Query_QueryHistoryOrderList(in *et.SpotQueryHistoryOrderList) (types.Message, error) {
	if !et.CheckExchangeAsset(e.GetAPI().GetConfig().GetCoinExec(), in.LeftAsset, in.RightAsset) {
		return nil, et.ErrAsset
	}
	if !et.CheckCount(in.Count) {
		return nil, et.ErrCount
	}

	if !et.CheckDirection(in.Direction) {
		return nil, et.ErrDirection
	}
	return spot.QueryHistoryOrderList(e.GetLocalDB(), &dbprefix{}, in)
}

//根据orderID查询订单信息
func (e *zksync) Query_QueryOrder(in *et.SpotQueryOrder) (types.Message, error) {
	if in.OrderID == 0 {
		return nil, et.ErrOrderID
	}
	return spot.FindOrderByOrderID(e.GetStateDB(), e.GetLocalDB(), &dbprefix{}, in.OrderID)
}

//根据订单状态，查询订单信息（这里面包含所有交易对）
func (e *zksync) Query_QueryOrderList(in *et.SpotQueryOrderList) (types.Message, error) {
	if !et.CheckStatus(in.Status) {
		return nil, et.ErrStatus
	}
	if !et.CheckCount(in.Count) {
		return nil, et.ErrCount
	}

	if !et.CheckDirection(in.Direction) {
		return nil, et.ErrDirection
	}

	if in.Address == "" {
		return nil, et.ErrAddr
	}
	return spot.QueryOrderList(e.GetLocalDB(), &dbprefix{}, in)
}

//根据orderID查询订单信息
func (e *zksync) Query_QueryNftOrder(in *et.SpotQueryOrder) (types.Message, error) {
	defer func() {
		if err := recover(); err != nil {
			zlog.Error("Query_QueryNftOrder", "err", err, "stack", et.GetStack())
		}
	}()
	if in.OrderID == 0 {
		return nil, et.ErrOrderID
	}
	return spot.FindOrderByOrderNftID(e.GetStateDB(), e.GetLocalDB(), &dbprefix{}, in.OrderID)
}
