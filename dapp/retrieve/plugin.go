package retrieve

import (
	"gitlab.33.cn/chain33/chain33/pluginmgr"
	"gitlab.33.cn/chain33/plugin/dapp/retrieve/commands"
	"gitlab.33.cn/chain33/plugin/dapp/retrieve/executor"
	"gitlab.33.cn/chain33/plugin/dapp/retrieve/rpc"
	"gitlab.33.cn/chain33/plugin/dapp/retrieve/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.RetrieveX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.RetrieveCmd,
		RPC:      rpc.Init,
	})
}
