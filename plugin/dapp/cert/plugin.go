package cert

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/cert/executor"
	"github.com/33cn/plugin/plugin/dapp/cert/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.CertX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      nil,
		RPC:      nil,
	})
}
