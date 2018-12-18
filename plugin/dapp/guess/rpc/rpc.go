// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	pb "github.com/33cn/plugin/plugin/dapp/guess/types"
)

//构造start原始交易
func (c *channelClient) GuessStart(ctx context.Context, parm *pb.GuessStartTxReq) (*types.UnsignTx, error) {
	v := &pb.GuessGameStart{
		Topic:          parm.Topic,
		Options:        parm.Options,
		Category:       parm.Category,
		MaxBetHeight:   parm.MaxBetHeight,
		MaxBetsOneTime: parm.MaxBetsOneTime,
		MaxBetsNumber:  parm.MaxBetsNumber,
		DevFeeFactor:   parm.DevFeeFactor,
		DevFeeAddr:     parm.DevFeeAddr,
		PlatFeeFactor:  parm.PlatFeeFactor,
		PlatFeeAddr:    parm.PlatFeeAddr,
		ExpireHeight:   parm.ExpireHeight,
	}

	val := &pb.GuessGameAction{
		Ty:    pb.GuessGameActionStart,
		Value: &pb.GuessGameAction_Start{Start: v},
	}

	name := types.ExecName(pb.GuessX)
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(pb.GuessX)),
		Payload: types.Encode(val),
		Fee:     parm.Fee,
		To:      address.ExecAddress(name),
	}

	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

//构造bet原始交易
func (c *channelClient) GuessBet(ctx context.Context, parm *pb.GuessBetTxReq) (*types.UnsignTx, error) {
	v := &pb.GuessGameBet{
		GameId:  parm.GameId,
		Option:  parm.Option,
		BetsNum: parm.Bets,
	}

	val := &pb.GuessGameAction{
		Ty:    pb.GuessGameActionBet,
		Value: &pb.GuessGameAction_Bet{Bet: v},
	}

	name := types.ExecName(pb.GuessX)
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(pb.GuessX)),
		Payload: types.Encode(val),
		Fee:     parm.Fee,
		To:      address.ExecAddress(name),
	}

	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

//构造stopBet原始交易
func (c *channelClient) GuessStopBet(ctx context.Context, parm *pb.GuessStopBetTxReq) (*types.UnsignTx, error) {
	v := &pb.GuessGameStopBet{
		GameId: parm.GameId,
	}

	val := &pb.GuessGameAction{
		Ty:    pb.GuessGameActionStopBet,
		Value: &pb.GuessGameAction_StopBet{StopBet: v},
	}

	name := types.ExecName(pb.GuessX)
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(pb.GuessX)),
		Payload: types.Encode(val),
		Fee:     parm.Fee,
		To:      address.ExecAddress(name),
	}

	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

//构造abort原始交易
func (c *channelClient) GuessAbort(ctx context.Context, parm *pb.GuessAbortTxReq) (*types.UnsignTx, error) {
	v := &pb.GuessGameAbort{
		GameId: parm.GameId,
	}

	val := &pb.GuessGameAction{
		Ty:    pb.GuessGameActionAbort,
		Value: &pb.GuessGameAction_Abort{Abort: v},
	}
	name := types.ExecName(pb.GuessX)
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(pb.GuessX)),
		Payload: types.Encode(val),
		Fee:     parm.Fee,
		To:      address.ExecAddress(name),
	}

	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

//构造publish原始交易
func (c *channelClient) GuessPublish(ctx context.Context, parm *pb.GuessPublishTxReq) (*types.UnsignTx, error) {
	v := &pb.GuessGamePublish{
		GameId: parm.GameId,
		Result: parm.Result,
	}

	val := &pb.GuessGameAction{
		Ty:    pb.GuessGameActionPublish,
		Value: &pb.GuessGameAction_Publish{Publish: v},
	}

	name := types.ExecName(pb.GuessX)
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(pb.GuessX)),
		Payload: types.Encode(val),
		Fee:     parm.Fee,
		To:      address.ExecAddress(name),
	}

	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}

//构造查询交易
func (c *channelClient) Show(ctx context.Context, head *pb.GuessGameQuery) (*types.UnsignTx, error) {
	val := &pb.GuessGameAction{
		Ty:    pb.GuessGameActionQuery,
		Value: &pb.GuessGameAction_Query{Query: head},
	}
	tx, err := types.CreateFormatTx(pb.GuessX, types.Encode(val))
	if err != nil {
		return nil, err
	}
	data := types.Encode(tx)
	return &types.UnsignTx{Data: data}, nil
}
