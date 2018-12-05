// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"reflect"

	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

// status
const (
	BlackwhiteStatusCreate = iota + 1
	BlackwhiteStatusPlay
	BlackwhiteStatusShow
	BlackwhiteStatusTimeout
	BlackwhiteStatusDone
)

const (
	// TyLogBlackwhiteCreate log for blackwhite create game
	TyLogBlackwhiteCreate = 750
	// TyLogBlackwhitePlay log for blackwhite play game
	TyLogBlackwhitePlay = 751
	// TyLogBlackwhiteShow log for blackwhite show game
	TyLogBlackwhiteShow = 752
	// TyLogBlackwhiteTimeout log for blackwhite timeout game
	TyLogBlackwhiteTimeout = 753
	// TyLogBlackwhiteDone log for blackwhite down game
	TyLogBlackwhiteDone = 754
	// TyLogBlackwhiteLoopInfo log for blackwhite LoopInfo game
	TyLogBlackwhiteLoopInfo = 755
)

const (
	// GetBlackwhiteRoundInfo 用于在cmd里面的区分不同的查询
	GetBlackwhiteRoundInfo = "GetBlackwhiteRoundInfo"
	// GetBlackwhiteByStatusAndAddr 用于在cmd里面的区分不同的查询
	GetBlackwhiteByStatusAndAddr = "GetBlackwhiteByStatusAndAddr"
	// GetBlackwhiteloopResult 用于在cmd里面的区分不同的查询
	GetBlackwhiteloopResult = "GetBlackwhiteloopResult"
)

var (
	// BlackwhiteX 执行器名字
	BlackwhiteX = "blackwhite"
	glog        = log15.New("module", BlackwhiteX)
	// JRPCName json RPC name
	JRPCName = "Blackwhite"
	// ExecerBlackwhite 执行器名字byte形式
	ExecerBlackwhite = []byte(BlackwhiteX)
	actionName       = map[string]int32{
		"Create":      BlackwhiteActionCreate,
		"Play":        BlackwhiteActionPlay,
		"Show":        BlackwhiteActionShow,
		"TimeoutDone": BlackwhiteActionTimeoutDone,
	}
	logInfo = map[int64]*types.LogInfo{
		TyLogBlackwhiteCreate:   {Ty: reflect.TypeOf(ReceiptBlackwhite{}), Name: "LogBlackwhiteCreate"},
		TyLogBlackwhitePlay:     {Ty: reflect.TypeOf(ReceiptBlackwhite{}), Name: "LogBlackwhitePlay"},
		TyLogBlackwhiteShow:     {Ty: reflect.TypeOf(ReceiptBlackwhite{}), Name: "LogBlackwhiteShow"},
		TyLogBlackwhiteTimeout:  {Ty: reflect.TypeOf(ReceiptBlackwhite{}), Name: "LogBlackwhiteTimeout"},
		TyLogBlackwhiteDone:     {Ty: reflect.TypeOf(ReceiptBlackwhite{}), Name: "LogBlackwhiteDone"},
		TyLogBlackwhiteLoopInfo: {Ty: reflect.TypeOf(ReplyLoopResults{}), Name: "LogBlackwhiteLoopInfo"},
	}
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, ExecerBlackwhite)
}
