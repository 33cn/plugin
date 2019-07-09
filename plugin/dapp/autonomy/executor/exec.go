// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

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