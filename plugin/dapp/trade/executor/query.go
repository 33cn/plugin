// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

// 目前设计trade 的query， 有两个部分的大分类
// 1. 按token 分
//    可以用于 token的挂单查询 (按价格排序)： OnBuy/OnSale
//    token 的历史行情 （按价格排序）: SoldOut/BoughtOut--> TODO 是否需要按时间（区块高度排序更合理）
// 2. 按 addr 分。 用于客户个人的钱包
//    自己未完成的交易 （按地址状态来）
//    自己的历史交易 （addr 的所有订单）
//
// 由于现价买/卖是没有orderID的， 用txhash 代替作为key
// key 有两种 orderID， txhash (0xAAAAAAAAAAAAAAA)

// 1.15 both buy/sell order
func (t *trade) Query_GetOnesOrderWithStatus(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOnesOrderWithStatus(req)
}

// get order by id
func (t *trade) Query_GetOneOrder(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOneOrder(req)
}

// query reply utils

const (
	orderStatusInvalid = iota
	orderStatusOn
	orderStatusDone
	orderStatusRevoke
)

const (
	orderTypeInvalid = iota
	orderTypeSell
	orderTypeBuy
)

func fromStatus(status int32) (st, ty int32) {
	switch status {
	case pty.TradeOrderStatusOnSale:
		return orderStatusOn, orderTypeSell
	case pty.TradeOrderStatusSoldOut:
		return orderStatusDone, orderTypeSell
	case pty.TradeOrderStatusRevoked:
		return orderStatusRevoke, orderTypeSell
	case pty.TradeOrderStatusOnBuy:
		return orderStatusOn, orderTypeBuy
	case pty.TradeOrderStatusBoughtOut:
		return orderStatusDone, orderTypeBuy
	case pty.TradeOrderStatusBuyRevoked:
		return orderStatusRevoke, orderTypeBuy
	}
	return orderStatusInvalid, orderTypeInvalid
}

// GetOnesOrderWithStatus by address-status
func (t *trade) GetOnesOrderWithStatus(req *pty.ReqAddrAssets) (types.Message, error) {
	orderStatus, orderType := fromStatus(req.Status)
	if orderStatus == orderStatusInvalid || orderType == orderTypeInvalid {
		return nil, types.ErrInvalidParam
	}

	// 使用 owner isFinished 组合
	var order pty.LocalOrder
	if orderStatus == orderStatusOn {
		order.IsFinished = false
	} else {
		order.IsFinished = true
	}
	order.Owner = req.Addr
	if len(req.FromKey) > 0 {
		order.TxIndex = req.FromKey
	}
	rows, err := listV2(t.GetLocalDB(), "owner_isFinished", &order, req.Count, req.Direction)
	if err != nil {
		tradelog.Error("GetOnesOrderWithStatus", "err", err)
		return nil, err
	}
	return t.toTradeOrders(rows)
}

func fmtReply(cfg *types.Chain33Config, order *pty.LocalOrder) *pty.ReplyTradeOrder {
	priceExec := order.PriceExec
	priceSymbol := order.PriceSymbol
	if priceExec == "" {
		priceExec = cfg.GetCoinExec()
		priceSymbol = cfg.GetCoinSymbol()
	}

	return &pty.ReplyTradeOrder{
		TokenSymbol:       order.AssetSymbol,
		Owner:             order.Owner,
		AmountPerBoardlot: order.AmountPerBoardlot,
		MinBoardlot:       order.MinBoardlot,
		PricePerBoardlot:  order.PricePerBoardlot,
		TotalBoardlot:     order.TotalBoardlot,
		TradedBoardlot:    order.TradedBoardlot,
		BuyID:             order.BuyID,
		Status:            order.Status,
		SellID:            order.SellID,
		TxHash:            order.TxHash[0],
		Height:            order.Height,
		Key:               order.TxIndex,
		BlockTime:         order.BlockTime,
		IsSellOrder:       order.IsSellOrder,
		AssetExec:         order.AssetExec,
		PriceExec:         priceExec,
		PriceSymbol:       priceSymbol,
	}
}

func (t *trade) GetOneOrder(req *pty.ReqAddrAssets) (types.Message, error) {
	query := NewOrderTableV2(t.GetLocalDB())
	tradelog.Debug("query GetData dbg", "primary", req.FromKey)
	row, err := query.GetData([]byte(req.FromKey))
	if err != nil {
		tradelog.Error("query GetData failed", "key", req.FromKey, "err", err)
		return nil, err
	}

	o, ok := row.Data.(*pty.LocalOrder)
	if !ok {
		tradelog.Error("query GetData failed", "err", "bad row type")
		return nil, types.ErrTypeAsset
	}
	cfg := t.GetAPI().GetConfig()
	reply := fmtReply(cfg, o)

	return reply, nil
}
