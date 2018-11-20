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

	AuthECDSA        = 257
	AuthSM2          = 258
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, ExecerCert)
	types.RegistorExecutor(CertX, NewType())
	// init executor type
	types.RegisterDappFork(CertX, "Enable", 0)
}

// cert执行器类型结构
type CertType struct {
	types.ExecTypeBase
}

// NewType
func NewType() *CertType {
	c := &CertType{}
	c.SetChild(c)
	return c
}

// GetPayload
func (b *CertType) GetPayload() types.Message {
	return &CertAction{}
}

// GetName
func (b *CertType) GetName() string {
	return CertX
}

// GetLogMap
func (b *CertType) GetLogMap() map[int64]*types.LogInfo {
	return nil
}

// GetTypeMap
func (b *CertType) GetTypeMap() map[string]int32 {
	return actionName
}
