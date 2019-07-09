// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

// ExecLocal_PropBoard 创建提案
func (a *Autonomy) ExecLocal_PropBoard(payload *auty.ProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execLocalBoard(receiptData)
}

// ExecLocal_RvkPropBoard 撤销提案
func (a *Autonomy) ExecLocal_RvkPropBoard(payload *auty.RevokeProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error){
	return a.execLocalBoard(receiptData)
}

// ExecLocal_VotePropBoard 投票提案
func (a *Autonomy) ExecLocal_VotePropBoard(payload *auty.VoteProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execLocalBoard(receiptData)
}

// ExecLocal_TmintPropBoard 终止提案
func (a *Autonomy) ExecLocal_TmintPropBoard(payload *auty.TerminateProposalBoard, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return a.execLocalBoard(receiptData)
}
