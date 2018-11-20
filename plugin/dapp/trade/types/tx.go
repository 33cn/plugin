// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

//TradeSellTx : info for sell order
type TradeSellTx struct {
	TokenSymbol       string `json:"tokenSymbol"`
	AmountPerBoardlot int64  `json:"amountPerBoardlot"`
	MinBoardlot       int64  `json:"minBoardlot"`
	PricePerBoardlot  int64  `json:"pricePerBoardlot"`
	TotalBoardlot     int64  `json:"totalBoardlot"`
	Fee               int64  `json:"fee"`
	AssetExec         string `json:"assetExec"`
}

//TradeBuyTx :info for buy order to speficied order
type TradeBuyTx struct {
	SellID      string `json:"sellID"`
	BoardlotCnt int64  `json:"boardlotCnt"`
	Fee         int64  `json:"fee"`
}

//TradeRevokeTx :用于取消卖单的信息
type TradeRevokeTx struct {
	SellID string `json:"sellID,"`
	Fee    int64  `json:"fee"`
}

//TradeBuyLimitTx :用于挂买单的信息
type TradeBuyLimitTx struct {
	TokenSymbol       string `json:"tokenSymbol"`
	AmountPerBoardlot int64  `json:"amountPerBoardlot"`
	MinBoardlot       int64  `json:"minBoardlot"`
	PricePerBoardlot  int64  `json:"pricePerBoardlot"`
	TotalBoardlot     int64  `json:"totalBoardlot"`
	Fee               int64  `json:"fee"`
	AssetExec         string `json:"assetExec"`
}

//TradeSellMarketTx :用于向指定买单出售token的信息
type TradeSellMarketTx struct {
	BuyID       string `json:"buyID"`
	BoardlotCnt int64  `json:"boardlotCnt"`
	Fee         int64  `json:"fee"`
}

//TradeRevokeBuyTx :取消指定买单
type TradeRevokeBuyTx struct {
	BuyID string `json:"buyID,"`
	Fee   int64  `json:"fee"`
}
