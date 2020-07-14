// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/collateralize/types"
)

var clog = log.New("module", "execs.collateralize")
var driverName = pty.CollateralizeX

//InitExecType ...
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Collateralize{}))
}

// Init collateralize
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	driverName := GetName()
	if name != driverName {
		panic("system dapp can't be rename")
	}
	if sub != nil {
		types.MustDecode(sub, &cfg)
	}
	drivers.Register(cfg, driverName, newCollateralize, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

// GetName for Collateralize
func GetName() string {
	return newCollateralize().GetName()
}

// Collateralize driver
type Collateralize struct {
	drivers.DriverBase
}

func newCollateralize() drivers.Driver {
	c := &Collateralize{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(driverName))
	return c
}

// GetDriverName for Collateralize
func (c *Collateralize) GetDriverName() string {
	return pty.CollateralizeX
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (c *Collateralize) CheckReceiptExecOk() bool {
	return true
}

// ExecutorOrder 设置localdb的EnableRead
func (c *Collateralize) ExecutorOrder() int64 {
	cfg := c.GetAPI().GetConfig()
	if cfg.IsFork(c.GetHeight(), "ForkLocalDBAccess") {
		return drivers.ExecLocalSameTime
	}
	return c.DriverBase.ExecutorOrder()
}
