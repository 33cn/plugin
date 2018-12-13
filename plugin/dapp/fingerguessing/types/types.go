// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"reflect"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var (
	// 本执行器的名称
	FguessX    = "fingerguessing"
	ExecerGame = []byte(FguessX)
)

// 定义本执行器支持的Action种类
const (
	GameActionCreate = iota + 1
	GameActionMatch
	GameActionCancel
	GameActionClose
)

// 定义本执行器生成的log类型,此logID会在交易中返回，用于区块不同的action
//  建议使用比较大一些数字，避免和系统其它的执行器重合
const (
	//log for game
	TyLogCreateGame = 721
	TyLogMatchGame  = 722
	TyLogCancleGame = 723
	TyLogCloseGame  = 724
)

var tlog = log.New("module", FguessX)

// 初始化方法
func init() {
	// 将本执行器添加到系统白名单
	types.AllowUserExec = append(types.AllowUserExec, []byte(FguessX))
	// 向系统注册本执行器类型
	types.RegistorExecutor(FguessX, NewType())
	types.RegisterDappFork(FguessX, "Enable", 0)
}

// 返回本执行器名称
// chain33有主链和平行链，此方法会判断链的种类返回相应的执行器名称。
// 如果是主链上，返回的是"fingerguessing"这个名称
// 如果是平行链，返回的是"user.p.xxxx.gingerguessing"这样的名称
func getRealExecName(paraName string) string {
	return types.ExecName(paraName + FguessX)
}

// 初始化本执行器类型
func NewType() *GameType {
	c := &GameType{}
	c.SetChild(c)
	return c
}

// exec
type GameType struct {
	types.ExecTypeBase
}

// 返回本执行器的日志类型信息，用于rpc解析日志数据
func (at *GameType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogCreateGame: {reflect.TypeOf(ReceiptGame{}), "LogCreateGame"},
		TyLogCancleGame: {reflect.TypeOf(ReceiptGame{}), "LogCancleGame"},
		TyLogMatchGame:  {reflect.TypeOf(ReceiptGame{}), "LogMatchGame"},
		TyLogCloseGame:  {reflect.TypeOf(ReceiptGame{}), "LogCloseGame"},
	}
}

// 返回本执行器的负载类型
func (g *GameType) GetPayload() types.Message {
	return &FingerguessingAction{}
}

// 返回本执行器中的action字典，支持双向查找
func (g *GameType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create": GameActionCreate,
		"Cancel": GameActionCancel,
		"Close":  GameActionClose,
		"Match":  GameActionMatch,
	}
}
