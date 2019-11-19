// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "github.com/33cn/chain33/types"

//cert
const (
	CertActionNew    = 1
	CertActionUpdate = 2
	CertActionNormal = 3

	AuthECDSA = 257
	AuthSM2   = 258
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, ExecerCert)
	// init executor type
	types.RegFork(CertX, InitFork)
	types.RegExec(CertX, InitExecutor)
}

func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(CertX, "Enable", 0)
}

func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(CertX, NewType(cfg))
}

// CertType cert执行器类型结构
type CertType struct {
	types.ExecTypeBase
}

// NewType 新建cert类型结构
func NewType(cfg *types.Chain33Config) *CertType {
	c := &CertType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取payload
func (b *CertType) GetPayload() types.Message {
	return &CertAction{}
}

// GetName 获取执行器名
func (b *CertType) GetName() string {
	return CertX
}

// GetLogMap 获取logmap
func (b *CertType) GetLogMap() map[int64]*types.LogInfo {
	return nil
}

// GetTypeMap 获取类型map
func (b *CertType) GetTypeMap() map[string]int32 {
	return actionName
}
