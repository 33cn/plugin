// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/pokerbull/executor"
	pb "github.com/33cn/plugin/plugin/dapp/pokerbull/types"
	"github.com/pkg/errors"
)

func (c *channelClient) Start(ctx context.Context, head *pb.PBGameStart) (*types.UnsignTx, error) {
	if head.PlayerNum > executor.MaxPlayerNum {
		return nil, errors.New("Player number should be maximum 5")
	}

	val := &pb.PBGameAction{
		Ty:    pb.PBGameActionStart,
		Value: &pb.PBGameAction_Start{Start: head},
	}
	tx, err := types.CreateFormatTx(types.ExecName(pb.PokerBullX), types.Encode(val))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

func (c *channelClient) Continue(ctx context.Context, head *pb.PBGameContinue) (*types.UnsignTx, error) {
	val := &pb.PBGameAction{
		Ty:    pb.PBGameActionContinue,
		Value: &pb.PBGameAction_Continue{Continue: head},
	}
	tx, err := types.CreateFormatTx(types.ExecName(pb.PokerBullX), types.Encode(val))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

func (c *channelClient) Quit(ctx context.Context, head *pb.PBGameQuit) (*types.UnsignTx, error) {
	val := &pb.PBGameAction{
		Ty:    pb.PBGameActionQuit,
		Value: &pb.PBGameAction_Quit{Quit: head},
	}
	tx, err := types.CreateFormatTx(types.ExecName(pb.PokerBullX), types.Encode(val))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

func (c *channelClient) Show(ctx context.Context, head *pb.PBGameQuery) (*types.UnsignTx, error) {
	val := &pb.PBGameAction{
		Ty:    pb.PBGameActionQuery,
		Value: &pb.PBGameAction_Query{Query: head},
	}
	tx, err := types.CreateFormatTx(types.ExecName(pb.PokerBullX), types.Encode(val))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}
