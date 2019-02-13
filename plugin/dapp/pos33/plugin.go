package pos33

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/pos33/executor"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     "pos33",
		ExecName: "pos33",
		Exec:     executor.Init,
		Cmd:      nil,
		RPC:      nil,
	})
}
