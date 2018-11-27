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
)

const (
	// PlayStyleDefault 默认游戏类型
	PlayStyleDefault = iota + 1
	// PlayStyleDealer 庄家玩法
	PlayStyleDealer
)

const (
	// TyLogPBGameStart log for start PBgame
	TyLogPBGameStart = 721
	// TyLogPBGameContinue log for continue PBgame
	TyLogPBGameContinue = 722
	// TyLogPBGameQuit log for quit PBgame
	TyLogPBGameQuit = 723
	// TyLogPBGameQuery log for query PBgame
	TyLogPBGameQuery = 724
)

//包的名字可以通过配置文件来配置
//建议用github的组织名称，或者用户名字开头, 再加上自己的插件的名字
//如果发生重名，可以通过配置文件修改这些名字
var (
	JRPCName        = "pokerbull"
	PokerBullX      = "pokerbull"
	ExecerPokerBull = []byte(PokerBullX)
)

const (
	// FuncNameQueryGameListByIDs 根据id列表查询game列表
	FuncNameQueryGameListByIDs = "QueryGameListByIDs"
	// FuncNameQueryGameByID 根据id查询game
	FuncNameQueryGameByID = "QueryGameByID"
	// FuncNameQueryGameByAddr 根据地址查询game
	FuncNameQueryGameByAddr = "QueryGameByAddr"
	// FuncNameQueryGameByStatus 根据status查询game
	FuncNameQueryGameByStatus = "QueryGameByStatus"
	// FuncNameQueryGameByRound 查询某一回合游戏结果
	FuncNameQueryGameByRound = "QueryGameByRound"
)
