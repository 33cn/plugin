package rpc

import (
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/types"
)

var log = log15.New("module", "multisig.rpc")

// Jrpc 申明Jrpc结构体
type Jrpc struct {
	cli *channelClient
}

// Jrpc 申明Grpc结构体
type Grpc struct {
	*channelClient
}

type channelClient struct {
	types.ChannelClient
}

// Init 初始化rpc实例
func Init(name string, s types.RPCServer) {
	cli := &channelClient{}
	grpc := &Grpc{channelClient: cli}
	cli.Init(name, s, &Jrpc{cli: cli}, grpc)
}
