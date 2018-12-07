// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/fingerguessing/types"
)

var glog = log.New("module", "execs.fingeruessing")

var driverName = gt.FguessX

// 初始化时通过反射获取本执行器的方法列表
func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Fingerguessing{}))
}

//本执行器的初始化动作，向系统注册本执行器
func Init(name string, sub []byte) {
	glog.Debug("register fingeruessing execer")
	glog.Debug("The fork height is ============", "height", types.GetDappFork(driverName, "Enable"))
	drivers.Register(GetName(), newFguessing, types.GetDappFork(driverName, "Enable"))
}

// 定义执行器对象
type Fingerguessing struct {
	drivers.DriverBase
}

// 执行器对象初始化包装逻辑
// 后面的两步设置子对象和设置执行器类型必不可少
func newFguessing() drivers.Driver {
	t := &Fingerguessing{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

func GetName() string {
	return newFguessing().GetName()
}

// 返回本执行器驱动名称
func (g *Fingerguessing) GetDriverName() string {
	return driverName
}


