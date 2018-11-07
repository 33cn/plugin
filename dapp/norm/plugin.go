package norm

import (
	"gitlab.33.cn/chain33/chain33/pluginmgr"
	"gitlab.33.cn/chain33/plugin/dapp/norm/executor"
	"gitlab.33.cn/chain33/plugin/dapp/norm/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.NormX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      nil,
		RPC:      nil,
	})
}
