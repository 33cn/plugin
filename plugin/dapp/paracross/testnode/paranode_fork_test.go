package testnode

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

// TestNewParaConfigDisablesForkParaFee 固化 newParaConfig 的约定：
// ForkParaFee 必须在**创建节点之前**被置为 MaxHeight（禁用）。原因有两条：
//  1. 测试模式下 chain33 会把所有 fork 置 0，导致 para 链收手续费破坏共识；
//  2. testnode.NewWithConfig 会立即启动 blockchain 等后台 goroutine 并发读取 Forks，
//     之后再写 Forks 会构成数据竞争（chain33 的 Forks 无锁），
//     在 -race 下表现为 TestParaNode 随机失败。
//
// 这里刻意不启动节点：断言配置本身即可，避免测试自身的启停时序引入新的不稳定。
func TestNewParaConfigDisablesForkParaFee(t *testing.T) {
	cfg := newParaConfig("127.0.0.1:0")

	assert.False(t, cfg.IsFork(0, "ForkParaFee"), "ForkParaFee 应保持禁用")
	assert.Equal(t, int64(types.MaxHeight), cfg.GetFork("ForkParaFee"))
	assert.Equal(t, "127.0.0.1:0", cfg.GetModuleConfig().RPC.ParaChain.MainChainGrpcAddr)
}
