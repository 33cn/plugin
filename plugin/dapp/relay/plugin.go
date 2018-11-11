package relay

import (
	"gitlab.33.cn/chain33/chain33/pluginmgr"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/relay/commands"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/relay/executor"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/relay/rpc"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/relay/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.RelayX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.RelayCmd,
		RPC:      rpc.Init,
	})
}
