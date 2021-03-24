// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

var log = log15.New("module", "mix.rpc")

// Jrpc json rpc class
type Jrpc struct {
	cli *channelClient
}

// Grpc grpc class
type Grpc struct {
	*channelClient
}

type channelClient struct {
	types.ChannelClient
}

// Init init rpc server
func Init(name string, s types.RPCServer) {
	cli := &channelClient{}
	grpc := &Grpc{channelClient: cli}
	cli.Init(name, s, &Jrpc{cli: cli}, grpc)
	mixTy.RegisterMixPrivacyServer(s.GRPC(), grpc)
}
