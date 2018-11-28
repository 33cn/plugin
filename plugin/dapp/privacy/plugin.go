// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package privacy

import (
	"github.com/33cn/chain33/pluginmgr"
	_ "github.com/33cn/plugin/plugin/dapp/privacy/autotest" // register autotest package
	"github.com/33cn/plugin/plugin/dapp/privacy/commands"
	_ "github.com/33cn/plugin/plugin/dapp/privacy/crypto" // register crypto package
	"github.com/33cn/plugin/plugin/dapp/privacy/executor"
	"github.com/33cn/plugin/plugin/dapp/privacy/rpc"
	"github.com/33cn/plugin/plugin/dapp/privacy/types"
	_ "github.com/33cn/plugin/plugin/dapp/privacy/wallet" // register wallet package
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
