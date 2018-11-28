// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"

	"github.com/33cn/chain33/types"
)

// blackwhite action type
const (
	// BlackwhiteActionCreate blackwhite create action
	BlackwhiteActionCreate = iota
	// BlackwhiteActionPlay blackwhite play action
	BlackwhiteActionPlay
	// BlackwhiteActionShow blackwhite show action
	BlackwhiteActionShow
	// BlackwhiteActionTimeoutDone blackwhite timeout action
	BlackwhiteActionTimeoutDone
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, ExecerBlackwhite)
	// init executor type
	types.RegistorExecutor(BlackwhiteX, NewType())
	types.RegisterDappFork(BlackwhiteX, "ForkBlackWhiteV2", 900000)
	types.RegisterDappFork(BlackwhiteX, "Enable", 850000)
}

// BlackwhiteType 执行器基类结构体
type BlackwhiteType struct {
	types.ExecTypeBase
}

// NewType 创建执行器类型
func NewType() *BlackwhiteType {
	c := &BlackwhiteType{}
	c.SetChild(c)
	return c
}

// GetPayload 获取blackwhite action
func (b *BlackwhiteType) GetPayload() types.Message {
	return &BlackwhiteAction{}
}

// GetName 获取执行器名称
func (b *BlackwhiteType) GetName() string {
	return BlackwhiteX
}

// GetLogMap 获取log的映射对应关系
func (b *BlackwhiteType) GetLogMap() map[int64]*types.LogInfo {
	return logInfo
}

// GetTypeMap 根据action的name获取type
func (b *BlackwhiteType) GetTypeMap() map[string]int32 {
	return actionName
}

// ActionName 根据交易的payLoad获取blackwhite的action的name
func (b BlackwhiteType) ActionName(tx *types.Transaction) string {
	var g BlackwhiteAction
	err := types.Decode(tx.Payload, &g)
	if err != nil {
		return "unknown-Blackwhite-action-err"
	}
	if g.Ty == BlackwhiteActionCreate && g.GetCreate() != nil {
		return "BlackwhiteCreate"
	} else if g.Ty == BlackwhiteActionShow && g.GetShow() != nil {
		return "BlackwhiteShow"
	} else if g.Ty == BlackwhiteActionPlay && g.GetPlay() != nil {
		return "BlackwhitePlay"
	} else if g.Ty == BlackwhiteActionTimeoutDone && g.GetTimeoutDone() != nil {
		return "BlackwhiteTimeoutDone"
	}
	return "unknown"
}

// Amount ...
func (b BlackwhiteType) Amount(tx *types.Transaction) (int64, error) {
	return 0, nil
}

// CreateTx ...
// TODO 暂时不修改实现， 先完成结构的重构
func (b BlackwhiteType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	glog.Debug("Blackwhite.CreateTx", "action", action)
	var tx *types.Transaction
	return tx, nil
}

// BlackwhiteRoundInfo ...
type BlackwhiteRoundInfo struct {
}

// Input for convert struct
func (t *BlackwhiteRoundInfo) Input(message json.RawMessage) ([]byte, error) {
	var req ReqBlackwhiteRoundInfo
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return types.Encode(&req), nil
}

// Output for convert struct
func (t *BlackwhiteRoundInfo) Output(reply interface{}) (interface{}, error) {
	return reply, nil
}

// BlackwhiteByStatusAndAddr ...
type BlackwhiteByStatusAndAddr struct {
}

// Input for convert struct
func (t *BlackwhiteByStatusAndAddr) Input(message json.RawMessage) ([]byte, error) {
	var req ReqBlackwhiteRoundList
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return types.Encode(&req), nil
}

// Output for convert struct
func (t *BlackwhiteByStatusAndAddr) Output(reply interface{}) (interface{}, error) {
	return reply, nil
}

// BlackwhiteloopResult ...
type BlackwhiteloopResult struct {
}

// Input for convert struct
func (t *BlackwhiteloopResult) Input(message json.RawMessage) ([]byte, error) {
	var req ReqLoopResult
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return types.Encode(&req), nil
}

// Output for convert struct
func (t *BlackwhiteloopResult) Output(reply interface{}) (interface{}, error) {
	return reply, nil
}
