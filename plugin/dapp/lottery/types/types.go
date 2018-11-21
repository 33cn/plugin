// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

//Lottery op
const (
	LotteryActionCreate = 1 + iota
	LotteryActionBuy
	LotteryActionShow
	LotteryActionDraw
	LotteryActionClose

	//log for lottery
	TyLogLotteryCreate = 801
	TyLogLotteryBuy    = 802
	TyLogLotteryDraw   = 803
	TyLogLotteryClose  = 804
)

// Lottery name
const (
	LotteryX = "lottery"
)

//Lottery status
const (
	LotteryCreated = 1 + iota
	LotteryPurchase
	LotteryDrawed
	LotteryClosed
)
