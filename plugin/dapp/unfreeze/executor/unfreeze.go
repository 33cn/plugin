// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	log "github.com/inconshreveable/log15"
	uf "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	drivers "gitlab.33.cn/chain33/chain33/system/dapp"
	"gitlab.33.cn/chain33/chain33/types"
)

var uflog = log.New("module", "execs.unfreeze")

var driverName = uf.UnfreezeX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Unfreeze{}))
}

func Init(name string, sub []byte) {
	drivers.Register(GetName(), newUnfreeze, 0)
}

type Unfreeze struct {
	drivers.DriverBase
}

func newUnfreeze() drivers.Driver {
	t := &Unfreeze{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

func GetName() string {
	return newUnfreeze().GetName()
}

func (u *Unfreeze) GetDriverName() string {
	return driverName
}
