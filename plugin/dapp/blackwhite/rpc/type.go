// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"encoding/hex"
	"encoding/json"

	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	bw "github.com/33cn/plugin/plugin/dapp/blackwhite/types"
)

// Jrpc json rpc struct
type Jrpc struct {
	cli *channelClient
}

// Grpc grpc struct
type Grpc struct {
	*channelClient
}

type channelClient struct {
	rpctypes.ChannelClient
}

// Init init grpc param
func Init(name string, s rpctypes.RPCServer) {
	cli := &channelClient{}
	grpc := &Grpc{channelClient: cli}
	cli.Init(name, s, &Jrpc{cli: cli}, grpc)
	bw.RegisterBlackwhiteServer(s.GRPC(), grpc)
}

// BlackwhiteCreateTxRPC ...
type BlackwhiteCreateTxRPC struct{}

// Input for convert struct
func (t *BlackwhiteCreateTxRPC) Input(message json.RawMessage) ([]byte, error) {
	var req bw.BlackwhiteCreateTxReq
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return types.Encode(&req), nil
}

// Output for convert struct
func (t *BlackwhiteCreateTxRPC) Output(reply interface{}) (interface{}, error) {
	if replyData, ok := reply.(*types.Message); ok {
		if tx, ok := (*replyData).(*types.Transaction); ok {
			data := types.Encode(tx)
			return hex.EncodeToString(data), nil
		}
	}
	return nil, types.ErrTypeAsset
}

// BlackwhitePlayTxRPC ...
type BlackwhitePlayTxRPC struct {
}

// Input for convert struct
func (t *BlackwhitePlayTxRPC) Input(message json.RawMessage) ([]byte, error) {
	var req bw.BlackwhitePlayTxReq
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return types.Encode(&req), nil
}

// Output for convert struct
func (t *BlackwhitePlayTxRPC) Output(reply interface{}) (interface{}, error) {
	if replyData, ok := reply.(*types.Message); ok {
		if tx, ok := (*replyData).(*types.Transaction); ok {
			data := types.Encode(tx)
			return hex.EncodeToString(data), nil
		}
	}
	return nil, types.ErrTypeAsset
}

// BlackwhiteShowTxRPC ...
type BlackwhiteShowTxRPC struct {
}

// Input for convert struct
func (t *BlackwhiteShowTxRPC) Input(message json.RawMessage) ([]byte, error) {
	var req bw.BlackwhiteShowTxReq
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return types.Encode(&req), nil
}

// Output for convert struct
func (t *BlackwhiteShowTxRPC) Output(reply interface{}) (interface{}, error) {
	if replyData, ok := reply.(*types.Message); ok {
		if tx, ok := (*replyData).(*types.Transaction); ok {
			data := types.Encode(tx)
			return hex.EncodeToString(data), nil
		}
	}
	return nil, types.ErrTypeAsset
}

// BlackwhiteTimeoutDoneTxRPC ...
type BlackwhiteTimeoutDoneTxRPC struct {
}

// Input for convert struct
func (t *BlackwhiteTimeoutDoneTxRPC) Input(message json.RawMessage) ([]byte, error) {
	var req bw.BlackwhiteTimeoutDoneTxReq
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return types.Encode(&req), nil
}

// Output for convert struct
func (t *BlackwhiteTimeoutDoneTxRPC) Output(reply interface{}) (interface{}, error) {
	if replyData, ok := reply.(*types.Message); ok {
		if tx, ok := (*replyData).(*types.Transaction); ok {
			data := types.Encode(tx)
			return hex.EncodeToString(data), nil
		}
	}
	return nil, types.ErrTypeAsset
}
