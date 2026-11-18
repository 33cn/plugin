package executor

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// 覆盖 ForkSecurityHardening 分叉后 rollup 提交校验的三条加固。默认(local)配置下 needSetForkZero
// 会把 fork 压成 0, 即分叉默认生效; "分叉前"分支用独立配置
// SetDappFork(..., types.MaxHeight) 关闭, 用于分叉前后行为对照。
const hardeningTestTitle = "user.p.para."

// newHardeningTestRollup 构造 rollup 执行器 + 空状态库, forkHeight 为 ForkSecurityHardening 的高度。
// 注意: local 配置下 SetDappFork 会被 needSetForkZero 强制压成 0, 关闭分叉必须用非 local 配置。
func newHardeningTestRollup(t *testing.T, forkHeight int64) (*rollup, dbm.KV) {
	dir, state, _ := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dir, state) })

	cfgString := strings.Replace(types.GetDefaultCfgstring(), `Title="local"`, `Title="test"`, 1)
	cfg := types.NewChain33Config(cfgString)
	require.False(t, cfg.IsLocal())
	cfg.SetDappFork(rtypes.RollupX, rtypes.ForkSecurityHardening, forkHeight)
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(cfg)
	// 顺序校验通过后才会查询验证人集合, 让该查询失败以便拿到确定性的终止错误
	api.On("Query", mock.Anything, "GetNodeGroupStatus", mock.Anything).Return(nil, types.ErrActionNotSupport)

	r := newRollup().(*rollup)
	r.SetAPI(api)
	r.SetStateDB(state)
	require.NoError(t, state.Set(formatRollupStatusKey(hardeningTestTitle), types.Encode(&rtypes.RollupStatus{})))
	return r, state
}

// execCommit 执行一笔提交, 并把回执里的 KV 落库(框架在区块执行阶段做的事)
func execCommit(t *testing.T, r *rollup, state dbm.KV, cp *rtypes.CheckPoint) {
	tx, err := r.GetExecutorType().CreateTransaction(rtypes.NameCommitAction, cp)
	require.NoError(t, err)
	receipt, err := r.Exec(tx, 0)
	require.NoError(t, err)
	require.Equal(t, int32(types.ExecOk), receipt.Ty)
	for _, kv := range receipt.KV {
		require.NoError(t, state.Set(kv.Key, kv.Value))
	}
}

// TestExecCommitPersistsBlockFragIndex 分段提交的续接批依赖 status.BlockFragIndex, 执行侧必须落盘:
// 分叉前该字段不落盘, 大区块被切分后的下一个批永远过不了顺序校验(ErrOutOfOrderCommit)。
func TestExecCommitPersistsBlockFragIndex(t *testing.T) {
	r, state := newHardeningTestRollup(t, 0)

	h1 := &types.Header{Height: 1, ParentHash: []byte("genesis")}
	execCommit(t, r, state, &rtypes.CheckPoint{
		ChainTitle:  hardeningTestTitle,
		CommitRound: 1,
		Batch: &rtypes.BlockBatch{
			BlockHeaders:   []*types.Header{h1},
			BlockFragIndex: 2,
		},
	})

	status, err := GetRollupStatus(state, hardeningTestTitle)
	require.NoError(t, err)
	require.Equal(t, int32(2), status.BlockFragIndex, "分段索引必须随提交落盘")

	// 续接批: 重复提交被分割块的首个区块头, 后续区块正常续接
	h2 := &types.Header{Height: 2, ParentHash: common.Sha256(types.Encode(h1))}
	continuation := &rtypes.CheckPoint{
		ChainTitle:  hardeningTestTitle,
		CommitRound: 2,
		Batch:       &rtypes.BlockBatch{BlockHeaders: []*types.Header{h1, h2}},
	}
	require.Equal(t, ErrGetValPubs, r.checkCommit(continuation),
		"续接批必须通过顺序校验(分叉前这里会返回 ErrOutOfOrderCommit)")
}

// TestExecCommitPreForkKeepsBlockFragIndexEmpty 分叉前保持历史行为: 不落盘分段索引
func TestExecCommitPreForkKeepsBlockFragIndexEmpty(t *testing.T) {
	r, state := newHardeningTestRollup(t, types.MaxHeight)

	execCommit(t, r, state, &rtypes.CheckPoint{
		ChainTitle:  hardeningTestTitle,
		CommitRound: 1,
		Batch: &rtypes.BlockBatch{
			BlockHeaders:   []*types.Header{{Height: 1, ParentHash: []byte("genesis")}},
			BlockFragIndex: 2,
		},
	})

	status, err := GetRollupStatus(state, hardeningTestTitle)
	require.NoError(t, err)
	require.Equal(t, int32(0), status.BlockFragIndex, "分叉前不落盘分段索引")
}

// TestCheckCommitPubKeyListLengthMismatch 公钥列表比交易列表短时必须返回错误, 不能按下标越界 panic
func TestCheckCommitPubKeyListLengthMismatch(t *testing.T) {
	const valCount = 10
	r, cp, privs := newDedupTestExecutor(t, valCount)

	// 两条交易只带一条公钥
	cp.Batch.TxList = [][]byte{
		types.Encode(&types.Transaction{Payload: []byte("tx1")}),
		types.Encode(&types.Transaction{Payload: []byte("tx2")}),
	}
	cp.Batch.PubKeyList = [][]byte{privs[0].PubKey().Bytes()}
	// 改动交易数据后再签名: 长度检查在验证人门槛之后, 签名本身仍须合法
	signWith(t, cp, privs[:valCount*2/3+1])

	require.Equal(t, ErrTxPubKeyCountNotMatch, r.checkCommit(cp))
}

// TestCheckCommitRejectsNonCanonicalTitle 分叉后只有规范平行链title(user.p.<name>.)能进提交校验:
// 真实链title的前缀(user.p.super)与去掉结尾'.'的形式(user.p.para)都必须被拒
func TestCheckCommitRejectsNonCanonicalTitle(t *testing.T) {
	newCP := func(title string) *rtypes.CheckPoint {
		return &rtypes.CheckPoint{
			ChainTitle:  title,
			CommitRound: 1,
			Batch:       &rtypes.BlockBatch{BlockHeaders: []*types.Header{{Height: 1, ParentHash: []byte("genesis")}}},
		}
	}

	r, _ := newHardeningTestRollup(t, 0)
	for _, title := range []string{"user.p.super", "user.p.para"} {
		require.Equal(t, ErrChainTitle, r.checkCommit(newCP(title)), "title=%s 必须被拒", title)
	}
	// 规范title通过title校验, 停在后续的验证人查询(而不是 ErrChainTitle)
	require.Equal(t, ErrGetValPubs, r.checkCommit(newCP(hardeningTestTitle)))

	// 分叉前沿用前缀校验: 上面的前缀title能通过title检查, 停在后续的验证人查询
	old, _ := newHardeningTestRollup(t, types.MaxHeight)
	require.Equal(t, ErrGetValPubs, old.checkCommit(newCP("user.p.super")))
}
