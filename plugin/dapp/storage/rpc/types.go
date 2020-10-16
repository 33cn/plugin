package rpc

import (
	rpctypes "github.com/33cn/chain33/rpc/types"
)

/*
 * rpc相关结构定义和初始化
 */

// 实现grpc的service接口
type channelClient struct {
	rpctypes.ChannelClient
}

// Jrpc 实现json rpc调用实例
type Jrpc struct {
	cli *channelClient
}

// Grpc grpc
type Grpc struct {
	*channelClient
}

// Init init rpc
func Init(name string, s rpctypes.RPCServer) {
	cli := &channelClient{}
	grpc := &Grpc{channelClient: cli}
	cli.Init(name, s, &Jrpc{cli: cli}, grpc)
	//存在grpc service时注册grpc server，需要生成对应的pb.go文件
	//storagetypes.RegisterStorageServer(s.GRPC(), grpc)
}
