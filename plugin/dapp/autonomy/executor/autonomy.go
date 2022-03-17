// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

type subConfig struct {
	Total      string `json:"total"`
	UseBalance bool   `json:"useBalance"`
	Execer     string `json:"execer"`
	BindKey    string `json:"bindKey"`
}

var (
	alog         = log.New("module", "execs.autonomy")
	driverName   = auty.AutonomyX
	autonomyAddr string
	subcfg       subConfig
	ticketName   = auty.TicketX
)

// Init 重命名执行器名称
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	if sub != nil {
		types.MustDecode(sub, &subcfg)
	}
	autonomyAddr = address.ExecAddress(cfg.ExecName(auty.AutonomyX))
	ticketName = cfg.ExecName(auty.TicketX)
	drivers.Register(cfg, GetName(), newAutonomy, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

//InitExecType ...
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Autonomy{}))
}

// Autonomy 执行器结构体
type Autonomy struct {
	drivers.DriverBase
}

func newAutonomy() drivers.Driver {
	t := &Autonomy{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName 获得执行器名字
func GetName() string {
	return newAutonomy().GetName()
}

// GetDriverName 获得驱动名字
func (u *Autonomy) GetDriverName() string {
	return driverName
}
