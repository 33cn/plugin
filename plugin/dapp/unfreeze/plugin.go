// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unfreeze

import (
	"gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/commands"
	"gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/executor"
	"gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/rpc"
	uf "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/pluginmgr"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     uf.PackageName,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.Cmd,
		RPC:      rpc.Init,
	})
}
