// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fingerguessing

import (
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/fingerguessing/executor"
	"github.com/33cn/plugin/plugin/dapp/fingerguessing/types"
)

func init() {
	pluginmgr.Register(&pluginmgr.PluginBase{
		Name:     types.FguessX,
		ExecName: executor.GetName(),
		Exec:     executor.Init,
		Cmd:      nil,
		RPC:      nil,
	})
}
