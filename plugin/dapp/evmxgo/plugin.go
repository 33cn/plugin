package types

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/evmxgo/commands"
	"github.com/33cn/plugin/plugin/dapp/evmxgo/executor"
	"github.com/33cn/plugin/plugin/dapp/evmxgo/rpc"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
)

/*
 * 初始化dapp相关的组件
 */

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     evmxgotypes.EvmxgoX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.Cmd,
		RPC:      rpc.Init,
	})
}
