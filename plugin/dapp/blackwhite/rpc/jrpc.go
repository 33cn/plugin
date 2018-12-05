// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"encoding/hex"

	"github.com/33cn/chain33/types"
	bw "github.com/33cn/plugin/plugin/dapp/blackwhite/types"
)

// BlackwhiteCreateTx 创建游戏RPC接口
func (c *Jrpc) BlackwhiteCreateTx(parm *bw.BlackwhiteCreateTxReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	head := &bw.BlackwhiteCreate{
		PlayAmount:  parm.PlayAmount,
		PlayerCount: parm.PlayerCount,
		Timeout:     parm.Timeout,
		GameName:    parm.GameName,
	}
	reply, err := c.cli.Create(context.Background(), head)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// BlackwhiteShowTx 出示游戏密钥的RPC接口
func (c *Jrpc) BlackwhiteShowTx(parm *BlackwhiteShowTx, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	head := &bw.BlackwhiteShow{
		GameID: parm.GameID,
		Secret: parm.Secret,
	}
	reply, err := c.cli.Show(context.Background(), head)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// BlackwhitePlayTx 参与游戏的RPC接口
func (c *Jrpc) BlackwhitePlayTx(parm *BlackwhitePlayTx, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	head := &bw.BlackwhitePlay{
		GameID:     parm.GameID,
		Amount:     parm.Amount,
		HashValues: parm.HashValues,
	}
	reply, err := c.cli.Play(context.Background(), head)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}

// BlackwhiteTimeoutDoneTx 游戏超时RPC接口
func (c *Jrpc) BlackwhiteTimeoutDoneTx(parm *BlackwhiteTimeoutDoneTx, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	head := &bw.BlackwhiteTimeoutDone{
		GameID: parm.GameID,
	}
	reply, err := c.cli.TimeoutDone(context.Background(), head)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply.Data)
	return nil
}
