// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

type tradeOrderResult struct {
	TokenSymbol       string `json:"tokenSymbol"`
	Owner             string `json:"owner"`
	AmountPerBoardlot string `json:"amountPerBoardlot"`
	MinBoardlot       int64  `json:"minBoardlot"`
	PricePerBoardlot  string `json:"pricePerBoardlot"`
	TotalBoardlot     int64  `json:"totalBoardlot"`
	TradedBoardlot    int64  `json:"tradedBoardlot"`
	BuyID             string `json:"buyID"`
	Status            int32  `json:"status"`
	SellID            string `json:"sellID"`
	TxHash            string `json:"txHash"`
	Height            int64  `json:"height"`
	Key               string `json:"key"`
	BlockTime         int64  `json:"blockTime"`
	IsSellOrder       bool   `json:"isSellOrder"`
}

type replySellOrdersResult struct {
	SellOrders []*tradeOrderResult `json:"sellOrders"`
}

type replyBuyOrdersResult struct {
	BuyOrders []*tradeOrderResult `json:"buyOrders"`
}

type replyTradeOrdersResult struct {
	Orders []*tradeOrderResult `json:"orders"`
}
