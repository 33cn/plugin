package rpc

import (
	"gitlab.33.cn/chain33/chain33/common/log/log15"
	pty "gitlab.33.cn/chain33/plugin/plugin/dapp/privacy/types"
	"gitlab.33.cn/chain33/chain33/rpc/types"
)

var log = log15.New("module", "privacy.rpc")

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
	pty.RegisterPrivacyServer(s.GRPC(), grpc)
}
