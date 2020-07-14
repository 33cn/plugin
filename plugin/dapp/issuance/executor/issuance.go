// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/issuance/types"
)

var clog = log.New("module", "execs.issuance")
var driverName = pty.IssuanceX

//InitExecType ...
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Issuance{}))
}

// Init issuance
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	driverName := GetName()
	if name != driverName {
		panic("system dapp can't be rename")
	}
	if sub != nil {
		types.MustDecode(sub, &cfg)
	}
	drivers.Register(cfg, driverName, newIssuance, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

// GetName for Issuance
func GetName() string {
	return newIssuance().GetName()
}

// Issuance driver
type Issuance struct {
	drivers.DriverBase
}

func newIssuance() drivers.Driver {
	c := &Issuance{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(driverName))
	return c
}

// GetDriverName for Issuance
func (c *Issuance) GetDriverName() string {
	return pty.IssuanceX
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (c *Issuance) CheckReceiptExecOk() bool {
	return true
}

// ExecutorOrder 设置localdb的EnableRead
func (c *Issuance) ExecutorOrder() int64 {
	cfg := c.GetAPI().GetConfig()
	if cfg.IsFork(c.GetHeight(), "ForkLocalDBAccess") {
		return drivers.ExecLocalSameTime
	}
	return c.DriverBase.ExecutorOrder()
}
