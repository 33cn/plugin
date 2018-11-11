package rpc

import (
	"github.com/33cn/chain33/client/mocks"
	rpctypes "github.com/33cn/chain33/rpc/types"
)

func newTestChannelClient() *channelClient {
	api := &mocks.QueueProtocolAPI{}
	return &channelClient{
		ChannelClient: rpctypes.ChannelClient{QueueProtocolAPI: api},
	}
}

func newTestJrpcClient() *Jrpc {
	return &Jrpc{cli: newTestChannelClient()}
}
