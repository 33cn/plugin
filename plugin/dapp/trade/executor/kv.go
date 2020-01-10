// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

const (
	sellIDPrefix = "mavl-trade-sell-"
	buyIDPrefix  = "mavl-trade-buy-"
)

// ids
func calcTokenSellID(hash string) string {
	return sellIDPrefix + hash
}

func calcTokenBuyID(hash string) string {
	return buyIDPrefix + hash
}

// make a number as token's price whether cheap or dear
// support 1e8 bty pre token or 1/1e8 bty pre token, [1Coins, 1e16Coins]
// the number in key is used to sort buy orders and pages
func calcPriceOfToken(priceBoardlot, AmountPerBoardlot int64) int64 {
	return 1e8 * priceBoardlot / AmountPerBoardlot
}
