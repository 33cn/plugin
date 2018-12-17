// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"encoding/hex"

	"github.com/33cn/chain33/types"
	pb "github.com/33cn/plugin/plugin/dapp/guess/types"
)

func (c *Jrpc) GuessStartTx(parm *pb.GuessStartTxReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	reply, err := c.cli.GuessStart(context.Background(), parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

func (c *Jrpc) GuessBetTx(parm *pb.GuessBetTxReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	reply, err := c.cli.GuessBet(context.Background(), parm)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

func (c *Jrpc) GuessStopBetTx(parm *pb.GuessStopBetTxReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	reply, err := c.cli.GuessStopBet(context.Background(), parm)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

func (c *Jrpc) GuessAbortTx(parm *pb.GuessAbortTxReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	reply, err := c.cli.GuessAbort(context.Background(), parm)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

func (c *Jrpc) GuessPublishTx(parm *pb.GuessPublishTxReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	reply, err := c.cli.GuessPublish(context.Background(), parm)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

/*
func (c *Jrpc) GuessQueryTx(parm *pb.GuessQueryReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	head := &pb.GuessGameQuery{
		GameId: parm.GameId,
	}
	reply, err := c.cli.Show(context.Background(), head)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}*/
