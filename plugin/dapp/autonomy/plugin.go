// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autonomy

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/autonomy/commands"
	"github.com/33cn/plugin/plugin/dapp/autonomy/executor"
	"github.com/33cn/plugin/plugin/dapp/autonomy/rpc"
	ptypes "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     ptypes.PackageName,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.AutonomyCmd,
		RPC:      rpc.Init,
	})
}
