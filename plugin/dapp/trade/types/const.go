// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// trade op
const (
	TradeSellLimit = iota
	TradeBuyMarket
	TradeRevokeSell
	TradeSellMarket
	TradeBuyLimit
	TradeRevokeBuy
)

// log
const (
	TyLogTradeSellLimit  = 310
	TyLogTradeBuyMarket  = 311
	TyLogTradeSellRevoke = 312

	TyLogTradeSellMarket = 330
	TyLogTradeBuyLimit   = 331
	TyLogTradeBuyRevoke  = 332
)

// 0->not start, 1->on sale, 2->sold out, 3->revoke, 4->expired
const (
	TradeOrderStatusNotStart = iota //TradeOrderStatusNotStart :
	TradeOrderStatusOnSale
	TradeOrderStatusSoldOut
	TradeOrderStatusRevoked
	TradeOrderStatusExpired
	TradeOrderStatusOnBuy
	TradeOrderStatusBoughtOut
	TradeOrderStatusBuyRevoked
	TradeOrderStatusSellHalfRevoked
	TradeOrderStatusBuyHalfRevoked
	TradeOrderStatusGroupComplete
)

//SellOrderStatus : sell order status map
var SellOrderStatus = map[int32]string{
	TradeOrderStatusNotStart:   "NotStart",
	TradeOrderStatusOnSale:     "OnSale",
	TradeOrderStatusSoldOut:    "SoldOut",
	TradeOrderStatusRevoked:    "Revoked",
	TradeOrderStatusExpired:    "Expired",
	TradeOrderStatusOnBuy:      "OnBuy",
	TradeOrderStatusBoughtOut:  "BoughtOut",
	TradeOrderStatusBuyRevoked: "BuyRevoked",
}

//SellOrderStatus2Int : SellOrderStatus info to value in int32
var SellOrderStatus2Int = map[string]int32{
	"NotStart":   TradeOrderStatusNotStart,
	"OnSale":     TradeOrderStatusOnSale,
	"SoldOut":    TradeOrderStatusSoldOut,
	"Revoked":    TradeOrderStatusRevoked,
	"Expired":    TradeOrderStatusExpired,
	"OnBuy":      TradeOrderStatusOnBuy,
	"BoughtOut":  TradeOrderStatusBoughtOut,
	"BuyRevoked": TradeOrderStatusBuyRevoked,
}

//MapSellOrderStatusStr2Int :
var MapSellOrderStatusStr2Int = map[string]int32{
	"onsale":  TradeOrderStatusOnSale,
	"soldout": TradeOrderStatusSoldOut,
	"revoked": TradeOrderStatusRevoked,
}

//MapBuyOrderStatusStr2Int :
var MapBuyOrderStatusStr2Int = map[string]int32{
	"onbuy":      TradeOrderStatusOnBuy,
	"boughtout":  TradeOrderStatusBoughtOut,
	"buyrevoked": TradeOrderStatusBuyRevoked,
}

const (
	//InvalidStartTime :
	InvalidStartTime = 0
)

const (
	// ForkTradeAssetX support more kinds of asset
	ForkTradeAssetX = "ForkTradeAsset"
	// ForkTradeBuyLimitX support buy limit
	ForkTradeBuyLimitX = "ForkTradeBuyLimit"
	// ForkTradeIDX id without prefix
	ForkTradeIDX = "ForkTradeID"
)
