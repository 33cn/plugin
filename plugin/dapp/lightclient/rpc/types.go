package rpc

import (
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/log/log15"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	lightclienttypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	"github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient"
	_ "github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient/btc"
	_ "github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient/neutrino"
)

/*
 * rpc相关结构定义和初始化
 */

var (
	log = log15.New("module", "lightclient.rpc")
)

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
	lightclienttypes.RegisterLightclientServer(s.GRPC(), grpc)

	api, _ := client.New(s.GetQueueClient(), nil)

	lightCfg := &lightclient.Config{}
	types.MustDecode(s.GetQueueClient().GetConfig().GetSubConfig().RPC["light"], lightCfg)
	for _, c := range lightCfg.EnableClients {
		log.Info("Init", "new light client", c)
		if lc := lightclient.New(c); lc != nil {
			err := lc.Init(s.Context(), api, lightCfg)
			if err != nil {
				log.Error("Init", "light", c, "lightClient init err", err)
				panic(err)
			}
			go lc.Start()
		}
	}
}
