// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paracross

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/mix/commands"
	"github.com/33cn/plugin/plugin/dapp/mix/executor"
	"github.com/33cn/plugin/plugin/dapp/mix/rpc"
	"github.com/33cn/plugin/plugin/dapp/mix/types"
	_ "github.com/33cn/plugin/plugin/dapp/mix/wallet" // register wallet package
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.MixX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.MixCmd,
		RPC:      rpc.Init,
	})
}
