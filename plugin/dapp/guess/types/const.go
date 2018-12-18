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
	JRPCName    = "guess"
	GuessX      = "guess"
	ExecerGuess = []byte(GuessX)
)

const (
	//FuncName_QueryGamesByIds func name
	FuncName_QueryGamesByIds = "QueryGamesByIds"
	//FuncName_QueryGameById func name
	FuncName_QueryGameById = "QueryGameById"
	//FuncName_QueryGameByAddr func name
	FuncName_QueryGameByAddr = "QueryGamesByAddr"
	//FuncName_QueryGameByStatus func name
	FuncName_QueryGameByStatus = "QueryGamesByStatus"
	//FuncName_QueryGameByAdminAddr func name
	FuncName_QueryGameByAdminAddr = "QueryGamesByAdminAddr"
	//FuncName_QueryGameByAddrStatus func name
	FuncName_QueryGameByAddrStatus = "QueryGamesByAddrStatus"
	//FuncName_QueryGameByAdminStatus func name
	FuncName_QueryGameByAdminStatus = "QueryGamesByAdminStatus"
	//FuncName_QueryGameByCategoryStatus func name
	FuncName_QueryGameByCategoryStatus="QueryGamesByCategoryStatus"
)

const (
	//DevShareAddr default value
	DevShareAddr = "1D6RFZNp2rh6QdbcZ1d7RWuBUz61We6SD7"

	//PlatformShareAddr default value
	PlatformShareAddr = "1PHtChNt3UcfssR7v7trKSk3WJtAWjKjjX"
)
