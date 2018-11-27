// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"fmt"

	"github.com/33cn/chain33/types"
	pb "github.com/33cn/plugin/plugin/dapp/guess/types"
	"github.com/33cn/plugin/plugin/dapp/guess/executor"
)

func (c *channelClient) GuessStart(ctx context.Context, head *pb.GuessGameStart) (*types.UnsignTx, error) {
	if head.MaxBetsNumber > executor.MaxBetsNumber {
		return nil, fmt.Errorf("Max Bets Number Should Be Maximum %d", executor.MaxBetsNumber)
	}

	val := &pb.GuessGameAction{
		Ty:    pb.GuessGameActionStart,
			Value: &pb.GuessGameAction_Start{head},
	}
	tx, err := types.CreateFormatTx(pb.GuessX, types.Encode(val))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

func (c *channelClient) GuessBet(ctx context.Context, head *pb.GuessGameBet) (*types.UnsignTx, error) {
	val := &pb.GuessGameAction{
		Ty:    pb.GuessGameActionBet,
		Value: &pb.GuessGameAction_Bet{head},
	}
	tx, err := types.CreateFormatTx(pb.GuessX, types.Encode(val))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

func (c *channelClient) GuessAbort(ctx context.Context, head *pb.GuessGameAbort) (*types.UnsignTx, error) {
	val := &pb.GuessGameAction{
		Ty:    pb.GuessGameActionAbort,
		Value: &pb.GuessGameAction_Abort{head},
	}
	tx, err := types.CreateFormatTx(pb.GuessX, types.Encode(val))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

func (c *channelClient) GuessPublish(ctx context.Context, head *pb.GuessGamePublish) (*types.UnsignTx, error) {
	val := &pb.GuessGameAction{
		Ty:    pb.GuessGameActionPublish,
		Value: &pb.GuessGameAction_Publish{head},
	}
	tx, err := types.CreateFormatTx(pb.GuessX, types.Encode(val))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

func (c *channelClient) Show(ctx context.Context, head *pb.PBGameQuery) (*types.UnsignTx, error) {
	val := &pb.PBGameAction{
		Ty:    pb.PBGameActionQuery,
		Value: &pb.PBGameAction_Query{head},
	}
	tx, err := types.CreateFormatTx(pb.PokerBullX, types.Encode(val))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}
