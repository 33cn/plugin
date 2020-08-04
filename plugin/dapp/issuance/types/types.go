// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

//Issuance op
const (
	IssuanceActionCreate = 1 + iota // 创建借贷
	IssuanceActionDebt              // 大户抵押
	IssuanceActionRepay             // 大户清算
	IssuanceActionFeed              // 发行合约喂价
	IssuanceActionClose             // 关闭借贷
	IssuanceActionManage            // 借贷管理

	//log for Issuance
	TyLogIssuanceCreate = 741
	TyLogIssuanceDebt   = 742
	TyLogIssuanceRepay  = 743
	TyLogIssuanceFeed   = 745
	TyLogIssuanceClose  = 746
)

// Issuance name
const (
	IssuanceX                   = "issuance"
	CCNYTokenName               = "CCNY"
	IssuancePreLiquidationRatio = 11000 //TODO 预清算比例，抵押物价值跌到借出ccny价值110%的时候开始清算
)

//Issuance status
const (
	IssuanceStatusCreated = 1 + iota
	IssuanceStatusClose
)

//status ...
const (
	IssuanceUserStatusCreate = 1 + iota
	IssuanceUserStatusWarning
	IssuanceUserStatusSystemLiquidate
	IssuanceUserStatusExpire
	IssuanceUserStatusExpireLiquidate
	IssuanceUserStatusClose
)

//type ...
const (
	PriceFeedKey = "issuance-price-feed"
	GuarantorKey = "issuance-guarantor"
	ManageKey    = "issuance-manage"
	FundKey      = "issuance-fund"
)

//fork ...
var (
	ForkIssuanceTableUpdate = "ForkIssuanceTableUpdate"
)
