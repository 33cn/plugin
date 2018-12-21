package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	echotypes "github.com/33cn/plugin/plugin/dapp/echo/types/echo"
)

var (
	// KeyPrefixPing ping 前缀
	KeyPrefixPing = "mavl-echo-ping:%s"
	// KeyPrefixPang pang 前缀
	KeyPrefixPang = "mavl-echo-pang:%s"

	// KeyPrefixPingLocal local ping 前缀
	KeyPrefixPingLocal = "LODB-echo-ping:%s"
	// KeyPrefixPangLocal local pang 前缀
	KeyPrefixPangLocal = "LODB-echo-pang:%s"
)

// init 初始化时通过反射获取本执行器的方法列表
func init() {
	ety := types.LoadExecutorType(echotypes.EchoX)
	ety.InitFuncList(types.ListMethod(&Echo{}))
}

// Init 本执行器的初始化动作，向系统注册本执行器，这里生效高度暂写为0
func Init(name string, sub []byte) {
	dapp.Register(echotypes.EchoX, newEcho, 0)
}

// Echo 定义执行器对象
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

// GetDriverName 返回本执行器驱动名称
func (h *Echo) GetDriverName() string {
	return echotypes.EchoX
}
