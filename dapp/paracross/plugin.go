package paracross

import (
	"gitlab.33.cn/chain33/chain33/pluginmgr"
	"gitlab.33.cn/chain33/plugin/dapp/paracross/commands"
	"gitlab.33.cn/chain33/plugin/dapp/paracross/executor"
	"gitlab.33.cn/chain33/plugin/dapp/paracross/rpc"
	"gitlab.33.cn/chain33/plugin/dapp/paracross/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.ParaX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.ParcCmd,
		RPC:      rpc.Init,
	})
}
