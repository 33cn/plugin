package multisig

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/multisig/commands"
	"github.com/33cn/plugin/plugin/dapp/multisig/executor"
	"github.com/33cn/plugin/plugin/dapp/multisig/rpc"
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     mty.MultiSigX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.MultiSigCmd,
		RPC:      rpc.Init,
	})
}
