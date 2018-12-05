// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"github.com/33cn/chain33/rpc/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

// Jrpc paracross jrpc interface
type Jrpc struct {
	cli *channelClient
}

// Grpc paracross Grpc interface
type Grpc struct {
	*channelClient
}

type channelClient struct {
	types.ChannelClient
}

// Init paracross rpc register
func Init(name string, s types.RPCServer) {
	cli := &channelClient{}
	grpc := &Grpc{channelClient: cli}
	cli.Init(name, s, &Jrpc{cli: cli}, grpc)

	pt.RegisterParacrossServer(s.GRPC(), grpc)
}
