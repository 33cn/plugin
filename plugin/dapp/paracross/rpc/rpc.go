// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

// IsSync query is sync
func (c *channelClient) IsSync(ctx context.Context, in *types.ReqNil) (*types.IsCaughtUp, error) {
	data, err := c.QueryConsensusFunc("para", "IsCaughtUp", &types.ReqNil{})
	if err != nil {
		return nil, err
	}
	return data.(*types.IsCaughtUp), nil
}

// IsSync query is sync
func (c *Jrpc) IsSync(in *types.ReqNil, result *interface{}) error {
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

// GetParaLocalBlockInfo query para chain the download layer's local height
func (c *channelClient) GetParaLocalBlockInfo(ctx context.Context, in *types.ReqInt) (*pt.ParaLocalDbBlockInfo, error) {
	data, err := c.QueryConsensusFunc("para", "LocalBlockInfo", in)
	if err != nil {
		return nil, err
	}
	return data.(*pt.ParaLocalDbBlockInfo), nil
}

// GetParaLocalBlockInfo query para local height
func (c *Jrpc) GetParaLocalBlockInfo(in *types.ReqInt, result *interface{}) error {
	data, err := c.cli.GetParaLocalBlockInfo(context.Background(), in)
	if err != nil {
		return err
	}
	*result = data
	return nil
}

// GetParaNodeLeaderInfo query para bls sign leader info
func (c *channelClient) GetParaNodeLeaderInfo(ctx context.Context, in *types.ReqNil) (*pt.ElectionStatus, error) {
	data, err := c.QueryConsensusFunc("para", "LeaderInfo", in)
	if err != nil {
		return nil, err
	}
	return data.(*pt.ElectionStatus), nil
}

// GetParaNodeLeaderInfo query para bls sign leader info
func (c *Jrpc) GetParaNodeLeaderInfo(in *types.ReqNil, result *interface{}) error {
	data, err := c.cli.GetParaNodeLeaderInfo(context.Background(), in)
	if err != nil {
		return err
	}
	*result = data
	return nil
}

// GetParaCmtTxInfo query para chain commit tx info to bls sign
func (c *channelClient) GetParaCmtTxInfo(ctx context.Context, in *types.ReqNil) (*pt.ParaBlsSignSumInfo, error) {
	data, err := c.QueryConsensusFunc("para", "CommitTxInfo", in)
	if err != nil {
		return nil, err
	}
	return data.(*pt.ParaBlsSignSumInfo), nil
}

// GetParaCmtTxInfo query para commit tx info
func (c *Jrpc) GetParaCmtTxInfo(in *types.ReqNil, result *interface{}) error {
	data, err := c.cli.GetParaCmtTxInfo(context.Background(), in)
	if err != nil {
		return err
	}
	*result = data
	return nil
}

// GetParaBlsPubKey query para chain bls pubkey
func (c *channelClient) GetParaBlsPubKey(ctx context.Context, in *types.ReqString) (*pt.BlsPubKey, error) {
	ins := *in
	data, err := c.QueryConsensusFunc("para", "BlsPubKey", &ins)
	if err != nil {
		return nil, err
	}
	return data.(*pt.BlsPubKey), nil
}

// GetParaBlsPubKey query para local height
func (c *Jrpc) GetParaBlsPubKey(in *types.ReqString, result *interface{}) error {
	data, err := c.cli.GetParaBlsPubKey(context.Background(), in)
	if err != nil {
		return err
	}
	*result = data
	return nil
}
