// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"golang.org/x/net/context"

	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

// Proposal Board 相关的接口
func (c *channelClient) propBoard(ctx context.Context, head *auty.ProposalBoard) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPropBoard,
		Value: &auty.AutonomyAction_PropBoard{PropBoard: head},
	}
	tx := &types.Transaction{
		Payload: types.Encode(val),
	}
	data, err := types.FormatTxEncode(types.ExecName(auty.AutonomyX), tx)
	if err != nil {
		return nil, err
	}
	return &types.UnsignTx{Data: data}, nil
}

func (c *channelClient) revokeProposalBoard(ctx context.Context, head *auty.RevokeProposalBoard) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionRvkPropBoard,
		Value: &auty.AutonomyAction_RvkPropBoard{RvkPropBoard: head},
	}
	tx := &types.Transaction{
		Payload: types.Encode(val),
	}
	data, err := types.FormatTxEncode(types.ExecName(auty.AutonomyX), tx)
	if err != nil {
		return nil, err
	}
	return &types.UnsignTx{Data: data}, nil
}

func (c *channelClient) voteProposalBoard(ctx context.Context, head *auty.VoteProposalBoard) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionVotePropBoard,
		Value: &auty.AutonomyAction_VotePropBoard{VotePropBoard: head},
	}
	tx := &types.Transaction{
		Payload: types.Encode(val),
	}
	data, err := types.FormatTxEncode(types.ExecName(auty.AutonomyX), tx)
	if err != nil {
		return nil, err
	}
	return &types.UnsignTx{Data: data}, nil
}

func (c *channelClient) terminateProposalBoard(ctx context.Context, head *auty.TerminateProposalBoard) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTmintPropBoard,
		Value: &auty.AutonomyAction_TmintPropBoard{TmintPropBoard: head},
	}
	tx := &types.Transaction{
		Payload: types.Encode(val),
	}
	data, err := types.FormatTxEncode(types.ExecName(auty.AutonomyX), tx)
	if err != nil {
		return nil, err
	}
	return &types.UnsignTx{Data: data}, nil
}

// Proposal Project 相关的接口
