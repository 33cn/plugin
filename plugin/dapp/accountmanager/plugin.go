package types

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/accountmanager/commands"
	"github.com/33cn/plugin/plugin/dapp/accountmanager/executor"
	"github.com/33cn/plugin/plugin/dapp/accountmanager/rpc"
	accountmanagertypes "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

/*
 * 初始化dapp相关的组件
 */

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     accountmanagertypes.AccountmanagerX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.Cmd,
		RPC:      rpc.Init,
	})
}
