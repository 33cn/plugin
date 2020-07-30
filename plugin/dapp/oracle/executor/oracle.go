/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	oty "github.com/33cn/plugin/plugin/dapp/oracle/types"
)

var olog = log.New("module", "execs.oracle")
var driverName = oty.OracleX

// Init 执行器初始化
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, newOracle().GetName(), newOracle, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

//InitExecType ...
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&oracle{}))
}

// GetName 获取oracle执行器名
func GetName() string {
	return newOracle().GetName()
}

func newOracle() drivers.Driver {
	t := &oracle{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// oracle driver
type oracle struct {
	drivers.DriverBase
}

func (ora *oracle) GetDriverName() string {
	return oty.OracleX
}
