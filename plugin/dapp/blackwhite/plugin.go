// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package blackwhite 黑白配游戏插件
package blackwhite

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/blackwhite/commands"
	"github.com/33cn/plugin/plugin/dapp/blackwhite/executor"
	"github.com/33cn/plugin/plugin/dapp/blackwhite/rpc"
	"github.com/33cn/plugin/plugin/dapp/blackwhite/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.BlackwhiteX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.BlackwhiteCmd,
		RPC:      rpc.Init,
	})
}
