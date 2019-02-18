// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	uf "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

var uflog = log.New("module", "execs.unfreeze")

var driverName = uf.UnfreezeX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Unfreeze{}))
}

// Init 重命名执行器名称
func Init(name string, sub []byte) {
	drivers.Register(GetName(), newUnfreeze, types.GetDappFork(driverName, "Enable"))
}

// Unfreeze 执行器结构体
type Unfreeze struct {
	drivers.DriverBase
}

func newUnfreeze() drivers.Driver {
	t := &Unfreeze{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName 获得执行器名字
func GetName() string {
	return newUnfreeze().GetName()
}

// GetDriverName 获得驱动名字
func (u *Unfreeze) GetDriverName() string {
	return driverName
}
