package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	echotypes "github.com/33cn/plugin/plugin/dapp/echo/types/echo"
)

var (
	// 执行交易生成的数据KEY
	KeyPrefixPing = "mavl-echo-ping:%s"
	KeyPrefixPang = "mavl-echo-pang:%s"
	// 本地执行生成的数据KEY
	KeyPrefixPingLocal = "LODB-echo-ping:%s"
	KeyPrefixPangLocal = "LODB-echo-pang:%s"
)

// 初始化时通过反射获取本执行器的方法列表
func init() {
	ety := types.LoadExecutorType(echotypes.EchoX)
	ety.InitFuncList(types.ListMethod(&Echo{}))
}

//本执行器的初始化动作，向系统注册本执行器，这里生效高度暂写为0
func Init(name string, sub []byte) {
	dapp.Register(echotypes.EchoX, newEcho, 0)
}

// 定义执行器对象
type Echo struct {
	dapp.DriverBase
}

// 执行器对象初始化包装逻辑，后面的两步设置子对象和设置执行器类型必不可少
func newEcho() dapp.Driver {
	c := &Echo{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(echotypes.EchoX))
	return c
}

// 返回本执行器驱动名称
func (h *Echo) GetDriverName() string {
	return echotypes.EchoX
}
