// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package echo

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/echo/executor"
	echotypes  "github.com/33cn/plugin/plugin/dapp/echo/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     "echo",
		ExecName: echotypes.EchoX,
		Exec:     executor.Init,
		Cmd:      nil,
		RPC:      nil,
	})
}
