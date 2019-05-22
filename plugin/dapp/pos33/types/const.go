// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import types "github.com/33cn/chain33/types"

//game action ty
const (
	Pos33ActionDeposit = iota + 1
	Pos33ActionWithdraw
	Pos33ActionDelegate
	Pos33ActionReword
	Pos33ActionPunish
	Pos33ActionElecte

	//log for game
	TyLogDeposit  = 911
	TyLogWithdraw = 912
	TyLogDelegate = 913
	TyLogReword   = 914
	TyLogPunish   = 915
	TyLogElecte   = 916
)

//包的名字可以通过配置文件来配置
//建议用github的组织名称，或者用户名字开头, 再加上自己的插件的名字
//如果发生重名，可以通过配置文件修改这些名字
const (
	Pos33X = "pos33"
)

// action name
const (
	ActionDeposit  = "deposit"
	ActionWithdraw = "withdraw"
	ActionDelegate = "delegate"
	ActionReword   = "reword"
	ActionPunish   = "punish"
)

// query func name
const (
	FuncNameQuery         = "QueryGameListByIds"
	FuncNameQueryGameByID = "QueryGameById"
)

const (
	// Pos33Miner 抵押的最小单位
	Pos33Miner = types.Coin * 10000
	// Pos33BlockReword 区块奖励
	Pos33BlockReword = types.Coin * 15
	// Pos33SortitionSize 多少区块做一次抽签
	Pos33SortitionSize = 10
	// Pos33VoteReword 每个区块的奖励
	Pos33VoteReword = types.Coin / 2
	// Pos33ProposerSize 候选区块Proposer数量
	Pos33ProposerSize = 7
	// Pos33VerifierSize  候选区块Verifier数量
	Pos33VerifierSize = 10
	// Pos33DepositPeriod 抵押周期
	Pos33DepositPeriod = 40320
	// Pos33FundKeyAddr ycc开发基金地址
	Pos33FundKeyAddr = "1DvAFGqS26Recz22yeoHcovzxN7dUh92ZY"
)

// const var
const (
	KeyPos33AllWeight       = "LODB-pos33-AllWeight:"
	KeyPos33WeightPrefix    = "LODB-pos33-Weight:"
	KeyPos33DelegatePrefix  = "LODB-pos33-Delegate:"
	KeyPos33CommitteePrefix = "LODB-pos33-Committee:"
	KeyPos33RewordPrefix    = "LODB-pos33-Reword:"
)
