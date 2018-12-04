// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// LotteryCreateTx for construction
type LotteryCreateTx struct {
	PurBlockNum    int64 `json:"purBlockNum"`
	DrawBlockNum   int64 `json:"drawBlockNum"`
	Fee            int64 `json:"fee"`
	OpRewardRatio  int64 `json:"opRewardRatio"`
	DevRewardRatio int64 `json:"devRewardRatio"`
}

// LotteryBuyTx for construction
type LotteryBuyTx struct {
	LotteryID string `json:"lotteryId"`
	Amount    int64  `json:"amount"`
	Number    int64  `json:"number"`
	Way       int64  `json:"way"`
	Fee       int64  `json:"fee"`
}

// LotteryDrawTx for construction
type LotteryDrawTx struct {
	LotteryID string `json:"lotteryId"`
	Fee       int64  `json:"fee"`
}

// LotteryCloseTx for construction
type LotteryCloseTx struct {
	LotteryID string `json:"lotteryId"`
	Fee       int64  `json:"fee"`
}
