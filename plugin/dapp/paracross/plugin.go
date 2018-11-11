package paracross

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/paracross/commands"
	"github.com/33cn/plugin/plugin/dapp/paracross/executor"
	"github.com/33cn/plugin/plugin/dapp/paracross/rpc"
	"github.com/33cn/plugin/plugin/dapp/paracross/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.ParaX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.ParcCmd,
		RPC:      rpc.Init,
	})
}
