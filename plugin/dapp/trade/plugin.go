package trade

import (
	"gitlab.33.cn/chain33/plugin/plugin/dapp/trade/commands"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/trade/executor"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/trade/rpc"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/trade/types"
	"gitlab.33.cn/chain33/chain33/pluginmgr"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.TradeX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.TradeCmd,
		RPC:      rpc.Init,
	})
}
