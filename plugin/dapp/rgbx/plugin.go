package types

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/rgbx/commands"
	"github.com/33cn/plugin/plugin/dapp/rgbx/executor"
	"github.com/33cn/plugin/plugin/dapp/rgbx/rpc"
	rgbxtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
)

/*
 * 初始化dapp相关的组件
 */

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     rgbxtypes.RgbxX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.Cmd,
		RPC:      rpc.Init,
	})
}
