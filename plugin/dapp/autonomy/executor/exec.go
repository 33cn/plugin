// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/pkg/errors"
)

// 提案董事会相关

// Exec_PropBoard 创建提案
func (a *Autonomy) Exec_PropBoard(payload *auty.ProposalBoard, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.propBoard(payload)
}

// Exec_RvkPropBoard 撤销提案
func (a *Autonomy) Exec_RvkPropBoard(payload *auty.RevokeProposalBoard, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.rvkPropBoard(payload)
}

// Exec_VotePropBoard 投票提案
func (a *Autonomy) Exec_VotePropBoard(payload *auty.VoteProposalBoard, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.votePropBoard(payload)
}

// Exec_TmintPropBoard 终止提案
func (a *Autonomy) Exec_TmintPropBoard(payload *auty.TerminateProposalBoard, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.tmintPropBoard(payload)
}

// 提案项目相关

// Exec_PropProject 创建提案项目
func (a *Autonomy) Exec_PropProject(payload *auty.ProposalProject, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.propProject(payload)
}

// Exec_RvkPropProject 撤销提案项目
func (a *Autonomy) Exec_RvkPropProject(payload *auty.RevokeProposalProject, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.rvkPropProject(payload)
}

// Exec_VotePropProject 投票提案项目
func (a *Autonomy) Exec_VotePropProject(payload *auty.VoteProposalProject, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.votePropProject(payload)
}

// Exec_PubVotePropProject 投票提案项目
func (a *Autonomy) Exec_PubVotePropProject(payload *auty.PubVoteProposalProject, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.pubVotePropProject(payload)
}

// Exec_TmintPropProject 终止提案项目
func (a *Autonomy) Exec_TmintPropProject(payload *auty.TerminateProposalProject, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.tmintPropProject(payload)
}

// 提案规则相关

// Exec_PropRule 创建提案规则
func (a *Autonomy) Exec_PropRule(payload *auty.ProposalRule, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.propRule(payload)
}

// Exec_RvkPropRule 撤销提案规则
func (a *Autonomy) Exec_RvkPropRule(payload *auty.RevokeProposalRule, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.rvkPropRule(payload)
}

// Exec_VotePropRule 投票提案规则
func (a *Autonomy) Exec_VotePropRule(payload *auty.VoteProposalRule, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.votePropRule(payload)
}

// Exec_TmintPropRule 终止提案规则
func (a *Autonomy) Exec_TmintPropRule(payload *auty.TerminateProposalRule, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.tmintPropRule(payload)
}

// Exec_Transfer 资金转入自治系统合约中
func (a *Autonomy) Exec_Transfer(payload *auty.TransferFund, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.transfer(payload)
}

// Exec_CommentProp 评论提案
func (a *Autonomy) Exec_CommentProp(payload *auty.Comment, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.commentProp(payload)
}

// 提案修改董事会成员相关

// Exec_PropChange 创建提案规则
func (a *Autonomy) Exec_PropChange(payload *auty.ProposalChange, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.propChange(payload)
}

// Exec_RvkPropChange 撤销提案规则
func (a *Autonomy) Exec_RvkPropChange(payload *auty.RevokeProposalChange, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.rvkPropChange(payload)
}

// Exec_VotePropChange 投票提案规则
func (a *Autonomy) Exec_VotePropChange(payload *auty.VoteProposalChange, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.votePropChange(payload)
}

// Exec_TmintPropChange 终止提案规则
func (a *Autonomy) Exec_TmintPropChange(payload *auty.TerminateProposalChange, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.tmintPropChange(payload)
}

// Exec_PropChange 创建事项规则
func (a *Autonomy) Exec_PropItem(payload *auty.ProposalItem, tx *types.Transaction, index int) (*types.Receipt, error) {
	if !a.GetAPI().GetConfig().IsDappFork(a.GetHeight(), auty.AutonomyX, auty.ForkAutonomyEnableItem) {
		return nil, errors.Wrapf(types.ErrActionNotSupport, "not after fork")
	}
	action := newAction(a, tx, int32(index))
	return action.propItem(payload)
}

// Exec_RvkPropItem 撤销事项规则
func (a *Autonomy) Exec_RvkPropItem(payload *auty.RevokeProposalItem, tx *types.Transaction, index int) (*types.Receipt, error) {
	if !a.GetAPI().GetConfig().IsDappFork(a.GetHeight(), auty.AutonomyX, auty.ForkAutonomyEnableItem) {
		return nil, errors.Wrapf(types.ErrActionNotSupport, "not after fork")
	}
	action := newAction(a, tx, int32(index))
	return action.rvkPropItem(payload)
}

// Exec_VotePropItem 投票事项规则
func (a *Autonomy) Exec_VotePropItem(payload *auty.VoteProposalItem, tx *types.Transaction, index int) (*types.Receipt, error) {
	if !a.GetAPI().GetConfig().IsDappFork(a.GetHeight(), auty.AutonomyX, auty.ForkAutonomyEnableItem) {
		return nil, errors.Wrapf(types.ErrActionNotSupport, "not after fork")
	}
	action := newAction(a, tx, int32(index))
	return action.votePropItem(payload)
}

// Exec_TmintPropItem 终止事项规则
func (a *Autonomy) Exec_TmintPropItem(payload *auty.TerminateProposalItem, tx *types.Transaction, index int) (*types.Receipt, error) {
	if !a.GetAPI().GetConfig().IsDappFork(a.GetHeight(), auty.AutonomyX, auty.ForkAutonomyEnableItem) {
		return nil, errors.Wrapf(types.ErrActionNotSupport, "not after fork")
	}
	action := newAction(a, tx, int32(index))
	return action.tmintPropItem(payload)
}
