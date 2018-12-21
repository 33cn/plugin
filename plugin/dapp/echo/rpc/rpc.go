package rpc

import (
	"context"

	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	echotypes "github.com/33cn/plugin/plugin/dapp/echo/types/echo"
)

// Jrpc 对外提供服务的RPC接口总体定义
type Jrpc struct {
	cli *channelClient
}

// RPC接口的本地实现
type channelClient struct {
	rpctypes.ChannelClient
}

// Init 注册 rpc 接口
func Init(name string, s rpctypes.RPCServer) {
	cli := &channelClient{}
	// 为了简单起见，这里只注册Jrpc，如果提供grpc的话也在这里注册
	cli.Init(name, s, &Jrpc{cli: cli}, nil)
}

// QueryPing 本合约的查询操作可以使用通用的Query接口，这里单独封装rpc的Query接口只是为了说明实现方式
// 接收客户端请求，并调用本地具体实现逻辑，然后返回结果
func (c *Jrpc) QueryPing(param *echotypes.Query, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	// 将具体的接口实现传递给本地逻辑
	reply, err := c.cli.QueryPing(context.Background(), param)
	if err != nil {
		return err
	}
	*result = reply
	return nil
}

// QueryPing 本地具体实现逻辑
func (c *channelClient) QueryPing(ctx context.Context, queryParam *echotypes.Query) (types.Message, error) {
	return c.Query(echotypes.EchoX, "GetPing", queryParam)
}
