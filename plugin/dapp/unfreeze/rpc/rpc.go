// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"encoding/hex"

	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

func (c *channelClient) GetUnfreeze(ctx context.Context, in *types.ReqString) (*pty.Unfreeze, error) {
	v, err := c.Query(pty.UnfreezeX, "GetUnfreeze", in)
	if err != nil {
		return nil, err
	}
	if resp, ok := v.(*pty.Unfreeze); ok {
		return resp, nil
	}
	return nil, types.ErrDecode
}

func (c *channelClient) GetUnfreezeWithdraw(ctx context.Context, in *types.ReqString) (*pty.ReplyQueryUnfreezeWithdraw, error) {
	v, err := c.Query(pty.UnfreezeX, "GetUnfreezeWithdraw", in)
	if err != nil {
		return nil, err
	}
	if resp, ok := v.(*pty.ReplyQueryUnfreezeWithdraw); ok {
		return resp, nil
	}
	return nil, types.ErrDecode
}

func (c *Jrpc) GetUnfreeze(in *types.ReqString, result *interface{}) error {
	v, err := c.cli.GetUnfreeze(context.Background(), in)
	if err != nil {
		return err
	}

	*result = v
	return nil
}

func (c *Jrpc) GetUnfreezeWithdraw(in *types.ReqString, result *interface{}) error {
	v, err := c.cli.GetUnfreezeWithdraw(context.Background(), in)
	if err != nil {
		return err
	}

	*result = v
	return nil
}

func (c *Jrpc) CreateRawUnfreezeCreate(param *pty.UnfreezeCreate, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	data, err := types.CallCreateTx(types.ExecName(pty.UnfreezeX), "UnfreezeCreateTX", param)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}

func (c *Jrpc) CreateRawUnfreezeWithdraw(param *pty.UnfreezeWithdraw, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	data, err := types.CallCreateTx(types.ExecName(pty.UnfreezeX), "UnfreezeWithdrawTx", param)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}

func (c *Jrpc) CreateRawUnfreezeTerminate(param *pty.UnfreezeTerminate, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	data, err := types.CallCreateTx(types.ExecName(pty.UnfreezeX), "UnfreezeTerminateTx", param)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}
