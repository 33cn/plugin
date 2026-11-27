// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// Lottery op
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

// Fork for lottery
const (
	// ForkLotteryFixOverflow 修复购买金额溢出漏洞，LotteryBuy 在转账/冻结前增加金额上限与溢出检查
	ForkLotteryFixOverflow = "ForkLotteryFixOverflow"
)

// Lottery status
const (
	LotteryCreated = 1 + iota
	LotteryPurchase
	LotteryDrawed
	LotteryClosed
)
