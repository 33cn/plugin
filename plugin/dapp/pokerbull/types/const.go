// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "github.com/33cn/chain33/types"

//game action ty
const (
	PBGameActionStart = iota + 1
	PBGameActionContinue
	PBGameActionQuit
	PBGameActionQuery
	PBGameActionPlay
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
	// TyLogPBGamePlay log for play PBgame
	TyLogPBGamePlay = 725
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
	// CreateStartTx 创建开始交易
	CreateStartTx = "Start"
	// CreateContinueTx 创建继续交易
	CreateContinueTx = "Continue"
	// CreateQuitTx 创建退出交易
	CreateQuitTx = "Quit"
	// CreatePlayTx 创建已匹配玩家交易
	CreatePlayTx = "Play"
)

const (
	// ListDESC 降序
	ListDESC = int32(0)
	// DefaultCount 默认一次取多少条记录
	DefaultCount = int32(20)
	// MaxPlayerNum 最大玩家数
	MaxPlayerNum = 5
	// MinPlayerNum 最小玩家数
	MinPlayerNum = 2
	// MinPlayValue 最小赌注
	MinPlayValue = 10 * types.Coin
	// DefaultStyle 默认游戏类型
	DefaultStyle = PlayStyleDefault
	// PlatformAddress 平台地址
	PlatformAddress = "1PHtChNt3UcfssR7v7trKSk3WJtAWjKjjX"
	// PlatformFee 平台佣金
	PlatformFee = int64(0.005 * float64(types.Coin))
	// DeveloperAddress 开发着地址
	DeveloperAddress = "1D6RFZNp2rh6QdbcZ1d7RWuBUz61We6SD7"
	// DeveloperFee 开发者佣金
	DeveloperFee = int64(0.005 * float64(types.Coin))
	// WinnerReturn 赢家回报率
	WinnerReturn = types.Coin - DeveloperFee - PlatformFee
	// PlatformSignAddress 平台签名地址
	PlatformSignAddress = "1Geb4ppNiAwMKKyrJgcis3JA57FkqsXvdR"
)
