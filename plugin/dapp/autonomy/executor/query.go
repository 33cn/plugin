// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

// Query_GetUnfreezeWithdraw 查询合约可提币量
func (a *Autonomy) Query_GetProposalBoard(in *auty.ReqQueryProposalBoard) (types.Message, error) {
	return a.getProposalBoard(in)
}