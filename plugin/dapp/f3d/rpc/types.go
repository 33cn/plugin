/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package rpc

import (
	rpctypes "github.com/33cn/chain33/rpc/types"
)

type channelClient struct {
	rpctypes.ChannelClient
}

type Jrpc struct {
	cli *channelClient
}

type Grpc struct {
	*channelClient
}

//func Init(name string, s rpctypes.RPCServer) {
//	cli := &channelClient{}
//	grpc := &Grpc{channelClient: cli}
//	cli.Init(name, s, &Jrpc{cli: cli}, grpc)
//
//}
