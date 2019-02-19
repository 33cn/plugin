package pos33

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/pos33/executor"
	"github.com/33cn/plugin/plugin/dapp/pos33/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.Pos33X,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      nil,
		RPC:      nil,
	})
}
