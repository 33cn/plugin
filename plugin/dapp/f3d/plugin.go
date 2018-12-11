package unfreeze

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/f3d/commands"
	"github.com/33cn/plugin/plugin/dapp/f3d/executor"
	pt "github.com/33cn/plugin/plugin/dapp/f3d/ptypes"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     pt.F3DX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.Cmd,
		RPC:      nil,
	})
}
