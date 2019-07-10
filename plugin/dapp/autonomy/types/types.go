// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"reflect"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var name string

var tlog = log.New("module", name)

func init() {
	name = AutonomyX
	types.AllowUserExec = append(types.AllowUserExec, []byte(name))
	// init executor type
	types.RegistorExecutor(name, NewType())
	types.RegisterDappFork(name, "Enable", 1)
}

//getRealExecName
func getRealExecName(paraName string) string {
	return types.ExecName(paraName + AutonomyX)
}

// NewType 生成新的基础类型
func NewType() *AutonomyType {
	c := &AutonomyType{}
	c.SetChild(c)
	return c
}

// AutonomyType 基础类型结构体
type AutonomyType struct {
	types.ExecTypeBase
}

// GetName 获取执行器名称
func (a *AutonomyType) GetName() string {
	return AutonomyX
}

// GetLogMap 获得日志类型列表
func (a *AutonomyType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogPropBoard:    {Ty: reflect.TypeOf(ReceiptProposalBoard{}), Name: "LogPropBoard"},
		TyLogRvkPropBoard:  {Ty: reflect.TypeOf(ReceiptProposalBoard{}), Name: "LogRvkPropBoard"},
		TyLogVotePropBoard:  {Ty: reflect.TypeOf(ReceiptProposalBoard{}), Name: "LogVotePropBoard"},
		TyLogTmintPropBoard: {Ty: reflect.TypeOf(ReceiptProposalBoard{}), Name: "LogTmintPropBoard"},

		TyLogPropProject:    {Ty: reflect.TypeOf(ReceiptProposalProject{}), Name: "LogPropProject"},
		TyLogRvkPropProject:  {Ty: reflect.TypeOf(ReceiptProposalProject{}), Name: "LogRvkPropProject"},
		TyLogVotePropProject:  {Ty: reflect.TypeOf(ReceiptProposalProject{}), Name: "LogVotePropProject"},
		TyLogTmintPropProject: {Ty: reflect.TypeOf(ReceiptProposalProject{}), Name: "LogTmintPropProject"},

		TyLogPropRule:    {Ty: reflect.TypeOf(ReceiptProposalRule{}), Name: "LogPropRule"},
		TyLogRvkPropRule:  {Ty: reflect.TypeOf(ReceiptProposalRule{}), Name: "LogRvkPropRule"},
		TyLogVotePropRule:  {Ty: reflect.TypeOf(ReceiptProposalRule{}), Name: "LogVotePropRule"},
		TyLogTmintPropRule: {Ty: reflect.TypeOf(ReceiptProposalRule{}), Name: "LogTmintPropRule"},
	}
}

// GetPayload 获得空的Unfreeze 的 Payload
func (a *AutonomyType) GetPayload() types.Message {
	return &AutonomyAction{}
}

// GetTypeMap 获得Action 方法列表
func (a *AutonomyType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"PropBoard":    AutonomyActionPropBoard,
		"RvkPropBoard":  AutonomyActionRvkPropBoard,
		"VotePropBoard":  AutonomyActionVotePropBoard,
		"TmintPropBoard": AutonomyActionTmintPropBoard,

		"PropProject":    AutonomyActionPropProject,
		"RvkPropProject":  AutonomyActionRvkPropProject,
		"VotePropProject":  AutonomyActionVotePropProject,
		"TmintPropProject": AutonomyActionTmintPropProject,

		"PropRule":    AutonomyActionPropRule,
		"RvkPropRule":  AutonomyActionRvkPropRule,
		"VotePropRule":  AutonomyActionVotePropRule,
		"TmintPropRule": AutonomyActionTmintPropRule,
	}
}