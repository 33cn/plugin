// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"github.com/33cn/chain33/types"
)

// NormX name
var NormX = "norm"

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(NormX))
	types.RegFork(NormX, InitFork)
	types.RegExec(NormX, InitExecutor)
}

//InitFork ...
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(NormX, "Enable", 0)
}

//InitExecutor ...
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(NormX, NewType(cfg))
}

// NormType def
type NormType struct {
	types.ExecTypeBase
}

// NewType method
func NewType(cfg *types.Chain33Config) *NormType {
	c := &NormType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetName 获取执行器名称
func (norm *NormType) GetName() string {
	return NormX
}

// GetPayload method
func (norm *NormType) GetPayload() types.Message {
	return &NormAction{}
}

// GetTypeMap method
func (norm *NormType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Nput": NormActionPut,
	}
}

// GetLogMap method
func (norm *NormType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{}
}
