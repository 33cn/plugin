package privacy

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/privacy/commands"
	_ "github.com/33cn/plugin/plugin/dapp/privacy/crypto"
	"github.com/33cn/plugin/plugin/dapp/privacy/executor"
	"github.com/33cn/plugin/plugin/dapp/privacy/rpc"
	"github.com/33cn/plugin/plugin/dapp/privacy/types"
	_ "github.com/33cn/plugin/plugin/dapp/privacy/wallet"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.PrivacyX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.PrivacyCmd,
		RPC:      rpc.Init,
	})
}
