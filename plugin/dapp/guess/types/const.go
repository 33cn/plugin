// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

//game action ty
const (
	PBGameActionStart = iota + 1
	PBGameActionContinue
	PBGameActionQuit
	PBGameActionQuery

	GuessGameActionStart = iota + 1
	GuessGameActionBet
	GuessGameActionStopBet
	GuessGameActionAbort
	GuessGameActionPublish
	GuessGameActionQuery

	GuessGameStatusStart = iota + 1
	GuessGameStatusBet
	GuessGameStatusStopBet
	GuessGameStatusAbort
	GuessGameStatusPublish
	GuessGameStatusTimeOut
)

//game log ty
const (
	TyLogGuessGameStart   = 901
	TyLogGuessGameBet     = 902
	TyLogGuessGameStopBet = 903
	TyLogGuessGameAbort   = 904
	TyLogGuessGamePublish = 905
	TyLogGuessGameTimeout = 906
)

//包的名字可以通过配置文件来配置
//建议用github的组织名称，或者用户名字开头, 再加上自己的插件的名字
//如果发生重名，可以通过配置文件修改这些名字
var (
	GuessX      = "guess"
	ExecerGuess = []byte(GuessX)
)

const (
	//FuncNameQueryGamesByIDs func name
	FuncNameQueryGamesByIDs = "QueryGamesByIDs"

	//FuncNameQueryGameByID func name
	FuncNameQueryGameByID = "QueryGameByID"

	//FuncNameQueryGameByAddr func name
	FuncNameQueryGameByAddr = "QueryGamesByAddr"

	//FuncNameQueryGameByStatus func name
	FuncNameQueryGameByStatus = "QueryGamesByStatus"

	//FuncNameQueryGameByAdminAddr func name
	FuncNameQueryGameByAdminAddr = "QueryGamesByAdminAddr"

	//FuncNameQueryGameByAddrStatus func name
	FuncNameQueryGameByAddrStatus = "QueryGamesByAddrStatus"

	//FuncNameQueryGameByAdminStatus func name
	FuncNameQueryGameByAdminStatus = "QueryGamesByAdminStatus"

	//FuncNameQueryGameByCategoryStatus func name
	FuncNameQueryGameByCategoryStatus = "QueryGamesByCategoryStatus"

	//CreateStartTx 创建开始交易
	CreateStartTx = "Start"

	//CreateBetTx 创建下注交易
	CreateBetTx = "Bet"

	//CreateStopBetTx 创建停止下注交易
	CreateStopBetTx = "StopBet"

	//CreatePublishTx 创建公布结果交易
	CreatePublishTx = "Publish"

	//CreateAbortTx 创建撤销游戏交易
	CreateAbortTx = "Abort"
)

const (
	//DevShareAddr default value
	DevShareAddr = "1D6RFZNp2rh6QdbcZ1d7RWuBUz61We6SD7"

	//PlatformShareAddr default value
	PlatformShareAddr = "1PHtChNt3UcfssR7v7trKSk3WJtAWjKjjX"
)
