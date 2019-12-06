package types

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/exchange/commands"
	"github.com/33cn/plugin/plugin/dapp/exchange/executor"
	"github.com/33cn/plugin/plugin/dapp/exchange/rpc"
	exchangetypes "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

/*
 * 初始化dapp相关的组件
 */

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     exchangetypes.ExchangeX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.Cmd,
		RPC:      rpc.Init,
	})
}
