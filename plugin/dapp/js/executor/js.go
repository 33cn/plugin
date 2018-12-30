package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
)

var (
	ptylog = log.New("module", "execs.js")
)

var driverName = ptypes.JsX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&js{}))
}

//Init 插件初始化
func Init(name string, sub []byte) {
	drivers.Register(GetName(), newjs, 0)
}

type js struct {
	drivers.DriverBase
}

func newjs() drivers.Driver {
	t := &js{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

//GetName 获取名字
func GetName() string {
	return newjs().GetName()
}

//GetDriverName 获取插件的名字
func (u *js) GetDriverName() string {
	return driverName
}
