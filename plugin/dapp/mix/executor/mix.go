// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"

	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

var (
	mlog       = log.New("module", "execs.mix")
	driverName = mixTy.MixX
)

// Mix exec
type Mix struct {
	drivers.DriverBase
}

//Init paracross exec register
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, GetName(), newMix, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
	setPrefix()
}

//InitExecType ...
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Mix{}))
}

//GetName return paracross name
func GetName() string {
	return newMix().GetName()
}

func newMix() drivers.Driver {
	c := &Mix{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(driverName))
	return c
}

// GetDriverName return paracross driver name
func (c *Mix) GetDriverName() string {
	return pt.ParaX
}
