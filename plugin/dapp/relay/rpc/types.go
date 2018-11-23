// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"github.com/33cn/chain33/rpc/types"
)

type jrpc struct {
	cli *channelClient
}

type grpc struct {
	*channelClient
}

type channelClient struct {
	types.ChannelClient
}

// Init relay rpc register
func Init(name string, s types.RPCServer) {
	cli := &channelClient{}
	grpc := &grpc{channelClient: cli}
	cli.Init(name, s, &jrpc{cli: cli}, grpc)
}
