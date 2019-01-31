// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"github.com/33cn/chain33/types"
)

// ValNodeX define
var ValNodeX = "valnode"

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(ValNodeX))
	types.RegistorExecutor(ValNodeX, NewType())
	types.RegisterDappFork(ValNodeX, "Enable", 0)
}

// GetExecName get exec name
func GetExecName() string {
	return types.ExecName(ValNodeX)
}

// ValNodeType stuct
type ValNodeType struct {
	types.ExecTypeBase
}

// NewType method
func NewType() *ValNodeType {
	c := &ValNodeType{}
	c.SetChild(c)
	return c
}

// GetName 获取执行器名称
func (t *ValNodeType) GetName() string {
	return ValNodeX
}

// GetPayload method
func (t *ValNodeType) GetPayload() types.Message {
	return &ValNodeAction{}
}

// GetTypeMap method
func (t *ValNodeType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Node":      ValNodeActionUpdate,
		"BlockInfo": ValNodeActionBlockInfo,
	}
}

// GetLogMap method
func (t *ValNodeType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{}
}
