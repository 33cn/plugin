// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"encoding/hex"

	"github.com/33cn/chain33/types"
	pb "github.com/33cn/plugin/plugin/dapp/pokerbull/types"
)

// PokerBullStartTx 创建游戏开始交易
func (c *Jrpc) PokerBullStartTx(parm *pb.PBStartTxReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	head := &pb.PBGameStart{
		Value:     parm.Value,
		PlayerNum: parm.PlayerNum,
	}
	reply, err := c.cli.Start(context.Background(), head)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// PokerBullContinueTx 创建游戏继续交易
func (c *Jrpc) PokerBullContinueTx(parm *pb.PBContinueTxReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	head := &pb.PBGameContinue{
		GameId: parm.GameId,
	}

	reply, err := c.cli.Continue(context.Background(), head)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

// PokerBullQuitTx 创建游戏推出交易
func (c *Jrpc) PokerBullQuitTx(parm *pb.PBQuitTxReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	head := &pb.PBGameQuit{
		GameId: parm.GameId,
	}
	reply, err := c.cli.Quit(context.Background(), head)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

// PokerBullQueryTx 创建游戏查询交易
func (c *Jrpc) PokerBullQueryTx(parm *pb.PBQueryReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	head := &pb.PBGameQuery{
		GameId: parm.GameId,
	}
	reply, err := c.cli.Show(context.Background(), head)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}
