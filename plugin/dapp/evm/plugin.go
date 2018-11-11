package evm

import (
	"gitlab.33.cn/chain33/chain33/pluginmgr"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/evm/commands"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/evm/executor"
	"gitlab.33.cn/chain33/plugin/plugin/dapp/evm/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.ExecutorName,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.EvmCmd,
		RPC:      nil,
	})
}
