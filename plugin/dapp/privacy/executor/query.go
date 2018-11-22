// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of p source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/privacy/types"
)

// Query_ShowAmountsOfUTXO show amount of utxo
func (p *privacy) Query_ShowAmountsOfUTXO(param *pty.ReqPrivacyToken) (types.Message, error) {
	return p.ShowAmountsOfUTXO(param)
}

// Query_ShowUTXOs4SpecifiedAmount shwo utxos for specified amount
func (p *privacy) Query_ShowUTXOs4SpecifiedAmount(param *pty.ReqPrivacyToken) (types.Message, error) {
	return p.ShowUTXOs4SpecifiedAmount(param)
}

// Query_GetUTXOGlobalIndex get utxo global index
func (p *privacy) Query_GetUTXOGlobalIndex(param *pty.ReqUTXOGlobalIndex) (types.Message, error) {
	return p.getGlobalUtxoIndex(param)
}

// Query_GetTxsByAddr get transactions by address
func (p *privacy) Query_GetTxsByAddr(param *types.ReqAddr) (types.Message, error) {
	return p.GetTxsByAddr(param)
}
