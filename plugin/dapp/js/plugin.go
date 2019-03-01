package js

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/js/cmd"
	"github.com/33cn/plugin/plugin/dapp/js/executor"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"

	// init auto test
	_ "github.com/33cn/plugin/plugin/dapp/js/autotest"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     ptypes.JsX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      cmd.JavaScriptCmd,
		RPC:      nil,
	})
}
