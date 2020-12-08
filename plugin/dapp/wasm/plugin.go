package wasm

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/wasm/commands"
	"github.com/33cn/plugin/plugin/dapp/wasm/executor"
	"github.com/33cn/plugin/plugin/dapp/wasm/rpc"
	"github.com/33cn/plugin/plugin/dapp/wasm/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.WasmX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.Cmd,
		RPC:      rpc.Init,
	})
}
