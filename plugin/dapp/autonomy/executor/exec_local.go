// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

// 提案董事会相关

// ExecLocal_PropBoard 创建提案
func (a *Autonomy) ExecLocal_PropBoard(payload *auty.ProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalBoard(tx, receiptData)
}

// ExecLocal_RvkPropBoard 撤销提案
func (a *Autonomy) ExecLocal_RvkPropBoard(payload *auty.RevokeProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalBoard(tx, receiptData)
}

// ExecLocal_VotePropBoard 投票提案
func (a *Autonomy) ExecLocal_VotePropBoard(payload *auty.VoteProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalBoard(tx, receiptData)
}

// ExecLocal_TmintPropBoard 终止提案
func (a *Autonomy) ExecLocal_TmintPropBoard(payload *auty.TerminateProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalBoard(tx, receiptData)
}

// 提案项目相关

// ExecLocal_PropProject 创建提案项目
func (a *Autonomy) ExecLocal_PropProject(payload *auty.ProposalProject, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalProject(tx, receiptData)
}

// ExecLocal_RvkPropProject 撤销提案项目
func (a *Autonomy) ExecLocal_RvkPropProject(payload *auty.RevokeProposalProject, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalProject(tx, receiptData)
}

// ExecLocal_VotePropProject 投票提案项目
func (a *Autonomy) ExecLocal_VotePropProject(payload *auty.VoteProposalProject, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalProject(tx, receiptData)
}

// ExecLocal_PubVotePropProject 全体投票提案项目
func (a *Autonomy) ExecLocal_PubVotePropProject(payload *auty.PubVoteProposalProject, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalProject(tx, receiptData)
}

// ExecLocal_TmintPropProject 终止提案项目
func (a *Autonomy) ExecLocal_TmintPropProject(payload *auty.TerminateProposalProject, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalProject(tx, receiptData)
}

// 提案规则相关

// ExecLocal_PropRule 创建提案规则
func (a *Autonomy) ExecLocal_PropRule(payload *auty.ProposalRule, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalRule(tx, receiptData)
}

// ExecLocal_RvkPropRule 撤销提案规则
func (a *Autonomy) ExecLocal_RvkPropRule(payload *auty.RevokeProposalRule, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalRule(tx, receiptData)
}

// ExecLocal_VotePropRule 投票提案规则
func (a *Autonomy) ExecLocal_VotePropRule(payload *auty.VoteProposalRule, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalRule(tx, receiptData)
}

// ExecLocal_TmintPropRule 终止提案规则
func (a *Autonomy) ExecLocal_TmintPropRule(payload *auty.TerminateProposalRule, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalRule(tx, receiptData)
}

// ExecLocal_CommentProp 评论提案
func (a *Autonomy) ExecLocal_CommentProp(payload *auty.Comment, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalCommentProp(tx, receiptData)
}

// 提案修改董事会成员相关

// ExecLocal_PropChange 创建提案规则
func (a *Autonomy) ExecLocal_PropChange(payload *auty.ProposalChange, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalChange(tx, receiptData)
}

// ExecLocal_RvkPropChange 撤销提案规则
func (a *Autonomy) ExecLocal_RvkPropChange(payload *auty.RevokeProposalChange, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalChange(tx, receiptData)
}

// ExecLocal_VotePropChange 投票提案规则
func (a *Autonomy) ExecLocal_VotePropChange(payload *auty.VoteProposalChange, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalChange(tx, receiptData)
}

// ExecLocal_TmintPropChange 终止提案规则
func (a *Autonomy) ExecLocal_TmintPropChange(payload *auty.TerminateProposalChange, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalChange(tx, receiptData)
}

// ExecLocal_PropItem 创建事项规则
func (a *Autonomy) ExecLocal_PropItem(payload *auty.ProposalItem, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalItem(tx, receiptData)
}

// ExecLocal_RvkPropItem 撤销事项规则
func (a *Autonomy) ExecLocal_RvkPropItem(payload *auty.RevokeProposalItem, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalItem(tx, receiptData)
}

// ExecLocal_VotePropItem 投票事项规则
func (a *Autonomy) ExecLocal_VotePropItem(payload *auty.VoteProposalItem, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalItem(tx, receiptData)
}

// ExecLocal_TmintPropItem 终止事项规则
func (a *Autonomy) ExecLocal_TmintPropItem(payload *auty.TerminateProposalItem, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execAutoLocalItem(tx, receiptData)
}
