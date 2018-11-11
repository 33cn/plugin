package ticket

import (
	"gitlab.33.cn/chain33/chain33/pluginmgr"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/ticket/commands"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/ticket/executor"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/ticket/rpc"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/ticket/types"
	_ "gitlab.33.cn/chain33/plugin/plugin/dapp/ticket/wallet"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.TicketX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.TicketCmd,
		RPC:      rpc.Init,
	})
}
