// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

type relayOrder2Show struct {
	OrderID       string `json:"orderid"`
	Status        string `json:"status"`
	Creator       string `json:"address"`
	Amount        string `json:"amount"`
	CoinOperation string `json:"coinoperation"`
	Coin          string `json:"coin"`
	CoinAmount    string `json:"coinamount"`
	CoinAddr      string `json:"coinaddr"`
	CoinWaits     uint32 `json:"coinwaits"`
	CreateTime    int64  `json:"createtime"`
	AcceptAddr    string `json:"acceptaddr"`
	AcceptTime    int64  `json:"accepttime"`
	ConfirmTime   int64  `json:"confirmtime"`
	FinishTime    int64  `json:"finishtime"`
	FinishTxHash  string `json:"finishtxhash"`
	Height        int64  `json:"height"`
}
