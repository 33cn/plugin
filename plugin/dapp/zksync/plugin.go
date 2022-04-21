package wasm

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/zksync/commands"
	"github.com/33cn/plugin/plugin/dapp/zksync/executor"
	"github.com/33cn/plugin/plugin/dapp/zksync/rpc"
	"github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.Zksync,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.ZksyncCmd,
		RPC:      rpc.Init,
	})
}
