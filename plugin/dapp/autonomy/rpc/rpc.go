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
func (c *channelClient) propProject(ctx context.Context, head *auty.ProposalProject) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPropProject,
		Value: &auty.AutonomyAction_PropProject{PropProject: head},
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

func (c *channelClient) revokeProposalProject(ctx context.Context, head *auty.RevokeProposalProject) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionRvkPropProject,
		Value: &auty.AutonomyAction_RvkPropProject{RvkPropProject: head},
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

func (c *channelClient) voteProposalProject(ctx context.Context, head *auty.VoteProposalProject) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionVotePropProject,
		Value: &auty.AutonomyAction_VotePropProject{VotePropProject: head},
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

func (c *channelClient) pubVoteProposalProject(ctx context.Context, head *auty.PubVoteProposalProject) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPubVotePropProject,
		Value: &auty.AutonomyAction_PubVotePropProject{PubVotePropProject: head},
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

func (c *channelClient) terminateProposalProject(ctx context.Context, head *auty.TerminateProposalProject) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTmintPropProject,
		Value: &auty.AutonomyAction_TmintPropProject{TmintPropProject: head},
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

// Proposal Rule 相关的接口
func (c *channelClient) propRule(ctx context.Context, head *auty.ProposalRule) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionPropRule,
		Value: &auty.AutonomyAction_PropRule{PropRule: head},
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

func (c *channelClient) revokeProposalRule(ctx context.Context, head *auty.RevokeProposalRule) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionRvkPropRule,
		Value: &auty.AutonomyAction_RvkPropRule{RvkPropRule: head},
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

func (c *channelClient) voteProposalRule(ctx context.Context, head *auty.VoteProposalRule) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionVotePropRule,
		Value: &auty.AutonomyAction_VotePropRule{VotePropRule: head},
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

func (c *channelClient) terminateProposalRule(ctx context.Context, head *auty.TerminateProposalRule) (*types.UnsignTx, error) {
	val := &auty.AutonomyAction{
		Ty:    auty.AutonomyActionTmintPropRule,
		Value: &auty.AutonomyAction_TmintPropRule{TmintPropRule: head},
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