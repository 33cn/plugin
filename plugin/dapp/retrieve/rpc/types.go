// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"github.com/33cn/chain33/rpc/types"
)

// Jrpc def
type Jrpc struct {
	cli *channelClient
}

// Grpc def
type Grpc struct {
	*channelClient
}

type channelClient struct {
	types.ChannelClient
}

// Init retrieve rpc
func Init(name string, s types.RPCServer) {
	cli := &channelClient{}
	grpc := &Grpc{channelClient: cli}
	cli.Init(name, s, &Jrpc{cli: cli}, grpc)
}
