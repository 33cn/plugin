// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

// Query_GetProposalBoard 查询提案董事会
func (a *Autonomy) Query_GetProposalBoard(in *auty.ReqQueryProposalBoard) (types.Message, error) {
	return a.getProposalBoard(in)
}

// Query_GetProposalProject 查询提案项目
func (a *Autonomy) Query_GetProposalProject(in *auty.ReqQueryProposalProject) (types.Message, error) {
	return a.getProposalProject(in)
}