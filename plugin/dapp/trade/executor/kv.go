// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
)

const (
	sellIDPrefix = "mavl-trade-sell-"
	buyIDPrefix  = "mavl-trade-buy-"
)

// 下个版本可以删除
const (
	sellOrderSHTAS = "LODB-trade-sellorder-shtas:"
	sellOrderASTS  = "LODB-trade-sellorder-asts:"
	sellOrderATSS  = "LODB-trade-sellorder-atss:"
	sellOrderTSPAS = "LODB-trade-sellorder-tspas:"
	buyOrderSHTAS  = "LODB-trade-buyorder-shtas:"
	buyOrderASTS   = "LODB-trade-buyorder-asts:"
	buyOrderATSS   = "LODB-trade-buyorder-atss:"
	buyOrderTSPAS  = "LODB-trade-buyorder-tspas:"
	// Addr-Status-Type-Height-Key
	orderASTHK = "LODB-trade-order-asthk:"
)

// ids
func calcTokenSellID(hash string) string {
	return sellIDPrefix + hash
}

func calcTokenBuyID(hash string) string {
	return buyIDPrefix + hash
}

// 特定帐号下的订单
// 这里状态进行转化, 分成 状态和类型， 状态三种， 类型 两种
//  on:  OnSale OnBuy
//  done:  Soldout boughtOut
//  revoke:  RevokeSell RevokeBuy
// buy/sell 两种类型
//  目前页面是按addr， 状态来

// make a number as token's price whether cheap or dear
// support 1e8 bty pre token or 1/1e8 bty pre token, [1Coins, 1e16Coins]
// the number in key is used to sort buy orders and pages
func calcPriceOfToken(priceBoardlot, AmountPerBoardlot int64) int64 {
	return 1e8 * priceBoardlot / AmountPerBoardlot
}

// UpdateLocalDBPart1 手动生成KV，需要在原有数据库中删除
// TODO
func UpdateLocalDBPart1() {
	prefix := []string{
		sellOrderSHTAS,
		sellOrderASTS,
		sellOrderATSS,
		sellOrderTSPAS,
		buyOrderSHTAS,
		buyOrderASTS,
		buyOrderATSS,
		buyOrderTSPAS,
		orderASTHK,
	}
	fmt.Printf("%+v", prefix)
}
