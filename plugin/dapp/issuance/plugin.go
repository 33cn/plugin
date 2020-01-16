// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package issuance

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/issuance/commands"
	"github.com/33cn/plugin/plugin/dapp/issuance/executor"
	"github.com/33cn/plugin/plugin/dapp/issuance/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.IssuanceX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      commands.IssuanceCmd,
	})
}
