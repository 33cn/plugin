// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

// Query_GetProposalBoard 查询提案董事会
func (a *Autonomy) Query_GetProposalBoard(in *types.ReqString) (types.Message, error) {
	return a.getProposalBoard(in)
}

// Query_ListProposalBoard 批量查询
func (a *Autonomy) Query_ListProposalBoard(in *auty.ReqQueryProposalBoard) (types.Message, error) {
	return a.listProposalBoard(in)
}

// Query_GetActiveBoard 查询当前board
func (a *Autonomy) Query_GetActiveBoard(in *types.ReqString) (types.Message, error) {
	return a.getActiveBoard()
}

// Query_GetProposalProject 查询提案项目
func (a *Autonomy) Query_GetProposalProject(in *types.ReqString) (types.Message, error) {
	return a.getProposalProject(in)
}

// Query_ListProposalProject 批量查询
func (a *Autonomy) Query_ListProposalProject(in *auty.ReqQueryProposalProject) (types.Message, error) {
	return a.listProposalProject(in)
}

// Query_GetProposalRule 查询提案规则
func (a *Autonomy) Query_GetProposalRule(in *types.ReqString) (types.Message, error) {
	return a.getProposalRule(in)
}

// Query_ListProposalRule 批量查询
func (a *Autonomy) Query_ListProposalRule(in *auty.ReqQueryProposalRule) (types.Message, error) {
	return a.listProposalRule(in)
}

// Query_GetActiveRule 查询当前rule
func (a *Autonomy) Query_GetActiveRule(in *types.ReqString) (types.Message, error) {
	return a.getActiveRule()
}

// Query_ListProposalComment 批量查询提案评论
func (a *Autonomy) Query_ListProposalComment(in *auty.ReqQueryProposalComment) (types.Message, error) {
	return a.listProposalComment(in)
}

// Query_GetProposalChange 查询提案修改董事会成员
func (a *Autonomy) Query_GetProposalChange(in *types.ReqString) (types.Message, error) {
	return a.getProposalChange(in)
}

// Query_ListProposalChange 批量查询
func (a *Autonomy) Query_ListProposalChange(in *auty.ReqQueryProposalChange) (types.Message, error) {
	return a.listProposalChange(in)
}

// Query_GetProposalItem 查询提案
func (a *Autonomy) Query_GetProposalItem(in *types.ReqString) (types.Message, error) {
	return getProposalItem(a.GetStateDB(), in)
}

// Query_ListProposalItem 批量查询
func (a *Autonomy) Query_ListProposalItem(in *auty.ReqQueryProposalItem) (types.Message, error) {
	return a.listProposalItem(in)
}

// Query_GetProposalChange 查询提案修改董事会成员
func (a *Autonomy) Query_IsAutonomyApprovedItem(in *types.ReqMultiStrings) (types.Message, error) {
	return IsAutonomyApprovedItem(a.GetStateDB(), in)
}
