package valnode

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/valnode/executor"
	"github.com/33cn/plugin/plugin/dapp/valnode/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.ValNodeX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      nil,
		RPC:      nil,
	})
}
