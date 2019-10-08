// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

//Collateralize op
const (
	CollateralizeActionCreate = 1 + iota
	CollateralizeActionBorrow
	CollateralizeActionRepay
	CollateralizeActionAppend
	CollateralizeActionFeed
	CollateralizeActionClose

	//log for Collateralize
	TyLogCollateralizeCreate    = 801
	TyLogCollateralizeBorrow    = 802
	TyLogCollateralizeRepay     = 803
	TyLogCollateralizeAppend    = 804
	TyLogCollateralizeFeed      = 805
	TyLogCollateralizeClose     = 806
)

// Collateralize name
const (
	CollateralizeX = "collateralize"
	CCNYTokenName = "ccny"
	CollateralizeRepayRatio = 1.1 //TODO 清算比例，抵押物价值跌到借出ccny价值110%的时候开始清算
)

//Collateralize status
const (
	CollateralizeStatusCreated = 1 + iota
	CollateralizeStatusClose
)

const (
	CollateralizeAssetTypeBty = 1 + iota
	CollateralizeAssetTypeBtc
	CollateralizeAssetTypeEth
)

const (
	CollateralizeUserStatusCreate = 1 + iota
	CollateralizeUserStatusWarning
	CollateralizeUserStatusSystemRepayed
	CollateralizeUserStatusClose
)