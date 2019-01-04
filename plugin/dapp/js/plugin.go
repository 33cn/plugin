package unfreeze

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/js/executor"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     ptypes.JsX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      nil,
		RPC:      nil,
	})
}
