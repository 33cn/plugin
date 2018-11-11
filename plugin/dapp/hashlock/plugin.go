package hashlock

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/hashlock/commands"
	"github.com/33cn/plugin/plugin/dapp/hashlock/executor"
	"github.com/33cn/plugin/plugin/dapp/hashlock/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.HashlockX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.HashlockCmd,
		RPC:      nil,
	})
}
