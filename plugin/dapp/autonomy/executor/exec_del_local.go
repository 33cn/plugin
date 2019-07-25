// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

// 提案董事会相关

// ExecDelLocal_PropBoard 创建提案董事会
func (a *Autonomy) ExecDelLocal_PropBoard(payload *auty.ProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalBoard(receiptData)
}

// ExecDelLocal_RvkPropBoard 撤销提案
func (a *Autonomy) ExecDelLocal_RvkPropBoard(payload *auty.RevokeProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalBoard(receiptData)
}

// ExecDelLocal_VotePropBoard 投票提案
func (a *Autonomy) ExecDelLocal_VotePropBoard(payload *auty.VoteProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalBoard(receiptData)
}

// ExecDelLocal_TmintPropBoard 终止提案
func (a *Autonomy) ExecDelLocal_TmintPropBoard(payload *auty.TerminateProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalBoard(receiptData)
}

// 提案项目相关

// ExecDelLocal_PropProject 创建提案项目
func (a *Autonomy) ExecDelLocal_PropProject(payload *auty.ProposalProject, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalProject(receiptData)
}

// ExecDelLocal_RvkPropProject 撤销提案
func (a *Autonomy) ExecDelLocal_RvkPropProject(payload *auty.RevokeProposalProject, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalProject(receiptData)
}

// ExecDelLocal_VotePropProject 投票提案
func (a *Autonomy) ExecDelLocal_VotePropProject(payload *auty.VoteProposalProject, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalProject(receiptData)
}

// ExecDelLocal_PubVotePropProject 投票提案
func (a *Autonomy) ExecDelLocal_PubVotePropProject(payload *auty.PubVoteProposalProject, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalProject(receiptData)
}

// ExecDelLocal_TmintPropProject 终止提案
func (a *Autonomy) ExecDelLocal_TmintPropProject(payload *auty.TerminateProposalProject, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalProject(receiptData)
}

// 提案规则相关

// ExecDelLocal_PropRule 创建提案规则
func (a *Autonomy) ExecDelLocal_PropRule(payload *auty.ProposalRule, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalRule(receiptData)
}

// ExecDelLocal_RvkPropRule 撤销提案规则
func (a *Autonomy) ExecDelLocal_RvkPropRule(payload *auty.RevokeProposalRule, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalRule(receiptData)
}

// ExecDelLocal_VotePropRule 投票提案规则
func (a *Autonomy) ExecDelLocal_VotePropRule(payload *auty.VoteProposalRule, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalRule(receiptData)
}

// ExecDelLocal_TmintPropRule 终止提案规则
func (a *Autonomy) ExecDelLocal_TmintPropRule(payload *auty.TerminateProposalRule, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalRule(receiptData)
}

// ExecDelLocal_CommentProp 终止提案规则
func (a *Autonomy) ExecDelLocal_CommentProp(payload *auty.Comment, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execDelLocalCommentProp(receiptData)
}
