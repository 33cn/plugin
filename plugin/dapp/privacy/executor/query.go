// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/privacy/types"
)

func (this *privacy) Query_ShowAmountsOfUTXO(param *pty.ReqPrivacyToken) (types.Message, error) {
	return this.ShowAmountsOfUTXO(param)
}

func (this *privacy) Query_ShowUTXOs4SpecifiedAmount(param *pty.ReqPrivacyToken) (types.Message, error) {
	return this.ShowUTXOs4SpecifiedAmount(param)
}

func (this *privacy) Query_GetUTXOGlobalIndex(param *pty.ReqUTXOGlobalIndex) (types.Message, error) {
	return this.getGlobalUtxoIndex(param)
}

func (this *privacy) Query_GetTxsByAddr(param *types.ReqAddr) (types.Message, error) {
	return this.GetTxsByAddr(param)
}
