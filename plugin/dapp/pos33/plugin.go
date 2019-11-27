// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ticket

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/pos3/commands"
	"github.com/33cn/plugin/plugin/dapp/pos3/executor"
	"github.com/33cn/plugin/plugin/dapp/pos3/rpc"
	"github.com/33cn/plugin/plugin/dapp/pos3/types"

	// init wallet
	_ "github.com/33cn/plugin/plugin/dapp/pos3/wallet"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.TicketX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.TicketCmd,
		RPC:      rpc.Init,
	})
}
