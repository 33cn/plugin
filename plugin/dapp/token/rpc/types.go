package rpc

import (
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/types"
)

var log = log15.New("module", "token.rpc")

type Jrpc struct {
	cli *channelClient
}

type Grpc struct {
	*channelClient
}

type channelClient struct {
	types.ChannelClient
}

func Init(name string, s types.RPCServer) {
	cli := &channelClient{}
	grpc := &Grpc{channelClient: cli}
	cli.Init(name, s, &Jrpc{cli: cli}, grpc)
}
