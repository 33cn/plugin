// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

func (c *channelClient) GetTitle(ctx context.Context, req *types.ReqString) (*pt.ParacrossConsensusStatus, error) {
	data, err := c.Query(pt.GetExecName(), "GetTitle", req)
	if err != nil {
		return nil, err
	}
	header, err := c.GetLastHeader()
	if err != nil {
		return nil, err
	}
	chainHeight := header.Height

	if resp, ok := data.(*pt.ParacrossStatus); ok {
		// 如果主链上查询平行链的高度，chain height应该是平行链的高度而不是主链高度， 平行链的真实高度需要在平行链侧查询
		if !types.IsPara() {
			chainHeight = resp.Height
		}
		return &pt.ParacrossConsensusStatus{
			Title:            resp.Title,
			ChainHeight:      chainHeight,
			ConsensHeight:    resp.Height,
			ConsensBlockHash: common.ToHex(resp.BlockHash),
		}, nil
	}
	return nil, types.ErrDecode
}

// GetHeight jrpc get consensus height
func (c *Jrpc) GetHeight(req *types.ReqString, result *interface{}) error {
	if req == nil || req.Data == "" {
		if types.IsPara() {
			req = &types.ReqString{Data: types.GetTitle()}
		} else {
			return types.ErrInvalidParam
		}
	}

	data, err := c.cli.GetTitle(context.Background(), req)
	*result = *data

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

func (c *channelClient) GetTitleHeight(ctx context.Context, req *pt.ReqParacrossTitleHeight) (*pt.RespParacrossDone, error) {
	data, err := c.Query(pt.GetExecName(), "GetTitleHeight", req)
	if err != nil {
		return nil, err
	}
	if resp, ok := data.(*pt.RespParacrossDone); ok {
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

func (c *channelClient) GetBlock2MainInfo(ctx context.Context, req *types.ReqBlocks) (*pt.ParaBlock2MainInfo, error) {
	ret := &pt.ParaBlock2MainInfo{}
	details, err := c.GetBlocks(req)
	if err != nil {
		return nil, err
	}

	for _, item := range details.Items {
		data := &pt.ParaBlock2MainMap{
			Height:     item.Block.Height,
			BlockHash:  common.ToHex(item.Block.Hash()),
			MainHeight: item.Block.MainHeight,
			MainHash:   common.ToHex(item.Block.MainHash),
		}
		ret.Items = append(ret.Items, data)
	}

	return ret, nil
}

// GetBlock2MainInfo jrpc get para block info with main chain map
func (c *Jrpc) GetBlock2MainInfo(req *types.ReqBlocks, result *interface{}) error {
	if req == nil {
		return types.ErrInvalidParam
	}

	ret, err := c.cli.GetBlock2MainInfo(context.Background(), req)
	*result = *ret
	return err
}

// GetNodeGroup get super node group
func (c *channelClient) GetNodeGroup(ctx context.Context, req *types.ReqString) (*types.ReplyConfig, error) {
	r := *req
	data, err := c.Query(pt.GetExecName(), "GetNodeGroup", &r)
	if err != nil {
		return nil, err
	}
	if resp, ok := data.(*types.ReplyConfig); ok {
		return resp, nil
	}
	return nil, types.ErrDecode
}

// GetNodeGroup get super node group
func (c *Jrpc) GetNodeGroup(req *types.ReqString, result *interface{}) error {
	data, err := c.cli.GetNodeGroup(context.Background(), req)
	*result = data
	return err
}

// GetNodeStatus get super node status
func (c *channelClient) GetNodeStatus(ctx context.Context, req *pt.ReqParacrossNodeInfo) (*pt.ParaNodeAddrStatus, error) {
	r := *req
	data, err := c.Query(pt.GetExecName(), "GetNodeAddrInfo", &r)
	if err != nil {
		return nil, err
	}
	if resp, ok := data.(*pt.ParaNodeAddrStatus); ok {
		return resp, nil
	}
	return nil, types.ErrDecode
}

// GetNodeStatus get super node status
func (c *Jrpc) GetNodeStatus(req *pt.ReqParacrossNodeInfo, result *interface{}) error {
	data, err := c.cli.GetNodeStatus(context.Background(), req)
	*result = data
	return err
}

//ListNodeStatus list super node by status
func (c *channelClient) ListNodeStatus(ctx context.Context, req *pt.ReqParacrossNodeInfo) (*pt.RespParacrossNodeAddrs, error) {
	r := *req
	data, err := c.Query(pt.GetExecName(), "ListNodeStatusInfo", &r)
	if err != nil {
		return nil, err
	}
	if resp, ok := data.(*pt.RespParacrossNodeAddrs); ok {
		return resp, nil
	}
	return nil, types.ErrDecode
}

//ListNodeStatus list super node by status
func (c *Jrpc) ListNodeStatus(req *pt.ReqParacrossNodeInfo, result *interface{}) error {
	data, err := c.cli.ListNodeStatus(context.Background(), req)
	*result = data
	return err
}
