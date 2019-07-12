// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
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

// Exec_RvkPropBoard 撤销提案项目
func (a *Autonomy) Exec_RvkPropProject(payload *auty.RevokeProposalProject, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.rvkPropProject(payload)
}

// Exec_VotePropBoard 投票提案项目
func (a *Autonomy) Exec_VotePropProject(payload *auty.VoteProposalProject, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.votePropProject(payload)
}

// Exec_VotePropBoard 投票提案项目
func (a *Autonomy) Exec_PubVotePropProject(payload *auty.PubVoteProposalProject, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.pubVotePropProject(payload)
}

// Exec_TmintPropBoard 终止提案项目
func (a *Autonomy) Exec_TmintPropProject(payload *auty.TerminateProposalProject, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(a, tx, int32(index))
	return action.tmintPropProject(payload)
}