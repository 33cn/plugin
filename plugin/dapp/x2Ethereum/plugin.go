package x2Ethereum

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/commands"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/executor"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/rpc"
	x2ethereumtypes "github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

/*
 * 初始化dapp相关的组件
 */

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     x2ethereumtypes.X2ethereumX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.Cmd,
		RPC:      rpc.Init,
	})
}
