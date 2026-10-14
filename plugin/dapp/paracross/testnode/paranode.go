package testnode

import (
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	_ "github.com/33cn/plugin/plugin/mempool/init"
)

/*
1. solo 模式，后台启动一个 主节点
2. 启动一个平行链节点：注意，这个要测试的话，会依赖平行链插件
*/

// ParaNode 平行链节点由两个节点组成
type ParaNode struct {
	Main *testnode.Chain33Mock
	Para *testnode.Chain33Mock
}

// NewParaNode 创建一个平行链节点
func NewParaNode(main *testnode.Chain33Mock, para *testnode.Chain33Mock) *ParaNode {
	if main == nil {
		main = testnode.New("", nil)
		main.Listen()
	}
	if para == nil {
		cfg := types.NewChain33Config(DefaultConfig)
		cfg.GetModuleConfig().RPC.ParaChain.MainChainGrpcAddr = main.GetCfg().RPC.GrpcBindAddr
		// chain33 ForkParaFee 默认-1，测试模式 SetAllFork(0) 激活收费
		// 导致 para 链处理同步区块时收手续费破坏共识，手动覆盖。
		// 必须在创建节点之前设置：testnode.NewWithConfig 会立即启动 blockchain 等后台 goroutine，
		// 之后任何对配置里 Forks 的写入都会与它们并发读 Forks 形成数据竞争（chain33 的 Forks 无锁），
		// 在 -race 下表现为 TestParaNode 随机失败。
		cfg.SetFork("ForkParaFee", types.MaxHeight)
		para = testnode.NewWithConfig(cfg, nil)
		para.Listen()
	}
	return &ParaNode{Main: main, Para: para}
}

// Close 关闭系统
func (node *ParaNode) Close() {
	node.Para.Close()
	node.Main.Close()
}
