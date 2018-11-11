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

	SignNameAuthECDSA = "auth_ecdsa"
	AUTH_ECDSA        = 257
	SignNameAuthSM2   = "auth_sm2"
	AUTH_SM2          = 258
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, ExecerCert)
	types.RegistorExecutor(CertX, NewType())
	// init executor type
	types.RegisterDappFork(CertX, "Enable", 0)
}

type CertType struct {
	types.ExecTypeBase
}

func NewType() *CertType {
	c := &CertType{}
	c.SetChild(c)
	return c
}

func (b *CertType) GetPayload() types.Message {
	return &CertAction{}
}

func (b *CertType) GetName() string {
	return CertX
}

func (b *CertType) GetLogMap() map[int64]*types.LogInfo {
	return nil
}

func (b *CertType) GetTypeMap() map[string]int32 {
	return actionName
}
