// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"

	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/trade/types"
)

//CreateRawTradeSellTx :
func (cc *channelClient) CreateRawTradeSellTx(ctx context.Context, in *ptypes.TradeForSell) (*types.UnsignTx, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	sell := &ptypes.Trade{
		Ty:    ptypes.TradeSellLimit,
		Value: &ptypes.Trade_SellLimit{SellLimit: in},
	}
	tx, err := types.CreateFormatTx(types.ExecName(ptypes.TradeX), types.Encode(sell))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

//CreateRawTradeBuyTx :
func (cc *channelClient) CreateRawTradeBuyTx(ctx context.Context, in *ptypes.TradeForBuy) (*types.UnsignTx, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	buy := &ptypes.Trade{
		Ty:    ptypes.TradeBuyMarket,
		Value: &ptypes.Trade_BuyMarket{BuyMarket: in},
	}
	tx, err := types.CreateFormatTx(types.ExecName(ptypes.TradeX), types.Encode(buy))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

//CreateRawTradeRevokeTx :
func (cc *channelClient) CreateRawTradeRevokeTx(ctx context.Context, in *ptypes.TradeForRevokeSell) (*types.UnsignTx, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	buy := &ptypes.Trade{
		Ty:    ptypes.TradeRevokeSell,
		Value: &ptypes.Trade_RevokeSell{RevokeSell: in},
	}
	tx, err := types.CreateFormatTx(types.ExecName(ptypes.TradeX), types.Encode(buy))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

//CreateRawTradeBuyLimitTx :
func (cc *channelClient) CreateRawTradeBuyLimitTx(ctx context.Context, in *ptypes.TradeForBuyLimit) (*types.UnsignTx, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	buy := &ptypes.Trade{
		Ty:    ptypes.TradeBuyLimit,
		Value: &ptypes.Trade_BuyLimit{BuyLimit: in},
	}
	tx, err := types.CreateFormatTx(types.ExecName(ptypes.TradeX), types.Encode(buy))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

//CreateRawTradeSellMarketTx :
func (cc *channelClient) CreateRawTradeSellMarketTx(ctx context.Context, in *ptypes.TradeForSellMarket) (*types.UnsignTx, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	buy := &ptypes.Trade{
		Ty:    ptypes.TradeSellMarket,
		Value: &ptypes.Trade_SellMarket{SellMarket: in},
	}
	tx, err := types.CreateFormatTx(types.ExecName(ptypes.TradeX), types.Encode(buy))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

//CreateRawTradeRevokeBuyTx :
func (cc *channelClient) CreateRawTradeRevokeBuyTx(ctx context.Context, in *ptypes.TradeForRevokeBuy) (*types.UnsignTx, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	buy := &ptypes.Trade{
		Ty:    ptypes.TradeRevokeBuy,
		Value: &ptypes.Trade_RevokeBuy{RevokeBuy: in},
	}
	tx, err := types.CreateFormatTx(types.ExecName(ptypes.TradeX), types.Encode(buy))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}
