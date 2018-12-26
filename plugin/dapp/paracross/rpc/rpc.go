// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

func (c *channelClient) GetTitle(ctx context.Context, req *types.ReqString) (*pt.ParacrossStatus, error) {
	data, err := c.Query(pt.GetExecName(), "GetTitle", req)
	if err != nil {
		return nil, err
	}
	if resp, ok := data.(*pt.ParacrossStatus); ok {
		return resp, nil
	}
	return nil, types.ErrDecode
}

// GetHeight jrpc get consensus height
func (c *Jrpc) GetHeight(req *types.ReqString, result *interface{}) error {
	if req == nil {
		return types.ErrInvalidParam
	}
	data, err := c.cli.GetTitle(context.Background(), req)
	*result = data
	return err
}

func (c *channelClient) ListTitles(ctx context.Context, req *types.ReqNil) (*pt.RespParacrossTitles, error) {
	data, err := c.Query(pt.GetExecName(), "ListTitles", req)
	if err != nil {
		return nil, err
	}
	if resp, ok := data.(*pt.RespParacrossTitles); ok {
		return resp, nil
	}
	return nil, types.ErrDecode
}

// ListTitles get paracross consensus titles list
func (c *Jrpc) ListTitles(req *types.ReqNil, result *interface{}) error {
	data, err := c.cli.ListTitles(context.Background(), req)
	*result = data
	return err
}

func (c *channelClient) GetTitleHeight(ctx context.Context, req *pt.ReqParacrossTitleHeight) (*pt.ReceiptParacrossDone, error) {
	data, err := c.Query(pt.GetExecName(), "GetTitleHeight", req)
	if err != nil {
		return nil, err
	}
	if resp, ok := data.(*pt.ReceiptParacrossDone); ok {
		return resp, nil
	}
	return nil, types.ErrDecode
}

// GetTitleHeight get consensus title height
func (c *Jrpc) GetTitleHeight(req *pt.ReqParacrossTitleHeight, result *interface{}) error {
	if req == nil {
		return types.ErrInvalidParam
	}
	data, err := c.cli.GetTitleHeight(context.Background(), req)
	*result = data
	return err
}

func (c *channelClient) GetAssetTxResult(ctx context.Context, req *types.ReqHash) (*pt.ParacrossAsset, error) {
	data, err := c.Query(pt.GetExecName(), "GetAssetTxResult", req)
	if err != nil {
		return nil, err
	}
	if resp, ok := data.(*pt.ParacrossAsset); ok {
		return resp, nil
	}
	return nil, types.ErrDecode
}

// GetAssetTxResult get asset tx result
func (c *Jrpc) GetAssetTxResult(req *types.ReqHash, result *interface{}) error {
	if req == nil {
		return types.ErrInvalidParam
	}
	data, err := c.cli.GetAssetTxResult(context.Background(), req)
	*result = data
	return err
}

// IsSync query is sync
func (g *channelClient) IsSync(ctx context.Context, in *types.ReqNil) (*types.IsCaughtUp, error) {
	data, err := g.QueryConsensusFunc("para", "IsCaughtUp", &types.ReqNil{})
	if err != nil {
		return nil, err
	}
	return data.(*types.IsCaughtUp), nil
}

// IsSync query is sync
func (c *Jrpc) IsSync(in *types.ReqNil, result *bool) error {
	//TODO consensus and paracross are not the same registered names ?
	data, err := c.cli.QueryConsensusFunc("para", "IsCaughtUp", &types.ReqNil{})
	*result = false
	if err != nil {
		return err
	}
	if reply, ok := data.(*types.IsCaughtUp); ok {
		*result = reply.Iscaughtup
	}
	return nil
}
