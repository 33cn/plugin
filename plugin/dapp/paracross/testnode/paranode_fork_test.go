package testnode

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

// TestParaNodeForkParaFeeDisabled 固化 NewParaNode 的配置约定：
// ForkParaFee 必须在节点创建**之前**被置为 MaxHeight（禁用）。原因有两条：
//  1. 测试模式下 chain33 会把所有 fork 置 0，导致 para 链收手续费破坏共识；
//  2. testnode.NewWithConfig 会立即启动 blockchain 等后台 goroutine，
//     之后再写配置里的 Forks 会与这些 goroutine 并发读 Forks 形成数据竞争
//     （chain33 的 Forks 无锁），在 -race 下表现为 TestParaNode 随机失败。
func TestParaNodeForkParaFeeDisabled(t *testing.T) {
	node := NewParaNode(nil, nil)
	defer node.Close()

	paraCfg := node.Para.GetAPI().GetConfig()
	assert.False(t, paraCfg.IsFork(0, "ForkParaFee"), "ForkParaFee 应保持禁用")
	assert.Equal(t, int64(types.MaxHeight), paraCfg.GetFork("ForkParaFee"))
}
