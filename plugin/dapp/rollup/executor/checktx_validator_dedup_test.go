package executor

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	paratypes "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// 规范平行链title(user.p.<name>.): ForkSecurityHardening 分叉后的 title 校验只接受该形式
const dedupTestTitle = "user.p.para."

// 说明: chain33 在 local/测试配置下会把所有 fork 强制置 0(needSetForkZero, types/fork.go),
// 单测里无法把 ForkSecurityHardening 关掉, 因此"分叉前"那条分支由
// Test_validatorSignThreshold 直接覆盖, 端到端只验证分叉后的行为。

// newDedupTestExecutor 构造 valCount 个验证人的成员集合, 以及一个已可通过顺序校验的 CheckPoint
func newDedupTestExecutor(t *testing.T, valCount int) (*rollup, *rtypes.CheckPoint, []crypto.PrivKey) {
	dir, state, _ := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dir, state) })

	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(cfg)

	r := &rollup{}
	r.SetAPI(api)
	r.SetStateDB(state)

	// 单区块头且高度为 1, 配合空 status 可以通过顺序校验(见 checkCommit 的 parentHash/parentHeight 推导)
	header := &types.Header{Height: 1, ParentHash: []byte("genesis")}
	cp := &rtypes.CheckPoint{
		ChainTitle:  dedupTestTitle,
		CommitRound: 1,
		Batch:       &rtypes.BlockBatch{BlockHeaders: []*types.Header{header}},
	}
	_ = state.Set(formatRollupStatusKey(dedupTestTitle), types.Encode(&rtypes.RollupStatus{}))

	privs := make([]crypto.PrivKey, 0, valCount)
	pubs := make([]string, 0, valCount)
	for i := 0; i < valCount; i++ {
		priv, err := blsDriver.GenKey()
		require.Nil(t, err)
		privs = append(privs, priv)
		pubs = append(pubs, common.ToHex(priv.PubKey().Bytes()))
	}
	api.On("Query", mock.Anything, "GetNodeGroupStatus", mock.Anything).
		Return(&paratypes.ParaNodeGroupStatus{BlsPubKeys: strings.Join(pubs, ",")}, nil)

	return r, cp, privs
}

// signWith 用 privs 为当前 batch 生成聚合签名与公钥列表
func signWith(t *testing.T, cp *rtypes.CheckPoint, privs []crypto.PrivKey) {
	msg := common.Sha256(types.Encode(cp.GetBatch()))
	sigs := make([]crypto.Signature, 0, len(privs))
	pubs := make([][]byte, 0, len(privs))
	for _, priv := range privs {
		sigs = append(sigs, priv.Sign(msg))
		pubs = append(pubs, priv.PubKey().Bytes())
	}
	agg, err := blsDriver.(crypto.AggregateCrypto).Aggregate(sigs)
	require.Nil(t, err)
	cp.ValidatorPubs = pubs
	cp.AggregateValidatorSign = agg.Bytes()
}

// repeatKey 把同一把私钥重复 count 次
func repeatKey(priv crypto.PrivKey, count int) []crypto.PrivKey {
	ret := make([]crypto.PrivKey, 0, count)
	for i := 0; i < count; i++ {
		ret = append(ret, priv)
	}
	return ret
}

// Test_checkCommit_RejectDuplicateValidatorPub 分叉后: 把同一把已注册公钥重复 t 遍
// (t = 成员数*2/3+1)、同一份签名也聚合 t 遍, 必须被拒(分叉前这会通过, 即单把验证人密钥即可伪造 2/3);
// 诚实提交(不同公钥)必须仍然通过。
func Test_checkCommit_RejectDuplicateValidatorPub(t *testing.T) {
	const valCount = 10
	threshold := valCount*2/3 + 1 // 7

	r, cp, privs := newDedupTestExecutor(t, valCount)

	signWith(t, cp, repeatKey(privs[0], threshold))
	require.Equal(t, ErrInvalidValidator, r.checkCommit(cp))

	honest := make([]crypto.PrivKey, 0, threshold)
	for i := 0; i < threshold; i++ {
		honest = append(honest, privs[i])
	}
	signWith(t, cp, honest)
	require.Nil(t, r.checkCommit(cp))
}

// Test_validatorSignThreshold 直接覆盖门槛计数的两条分支, 包括单测无法触达的"分叉前"行为。
func Test_validatorSignThreshold(t *testing.T) {
	const memberCount = 10
	threshold := memberCount*2/3 + 1 // 7

	memberPubs := make(map[string]struct{}, memberCount)
	distinct := make([][]byte, 0, memberCount)
	for i := 0; i < memberCount; i++ {
		pub := []byte{byte(i)}
		distinct = append(distinct, pub)
		memberPubs[hex.EncodeToString(pub)] = struct{}{}
	}

	dup7 := make([][]byte, 0, threshold)
	for i := 0; i < threshold; i++ {
		dup7 = append(dup7, distinct[0]) // 同一把公钥重复 7 次
	}

	// 分叉前: 按条目数计 -> 单把公钥重复 7 次即可达到门槛
	signCount, valCount := validatorSignThreshold(dup7, memberPubs, memberCount, false)
	require.Equal(t, threshold, signCount)
	require.Equal(t, memberCount, valCount)
	require.True(t, signCount >= valCount*2/3+1)

	// 分叉后: 去重后只剩 1 把 -> 达不到门槛
	signCount, valCount = validatorSignThreshold(dup7, memberPubs, memberCount, true)
	require.Equal(t, 1, signCount)
	require.Equal(t, memberCount, valCount)
	require.False(t, signCount >= valCount*2/3+1)

	// 分叉后: 7 把不同公钥 -> 达到门槛
	signCount, valCount = validatorSignThreshold(distinct[:threshold], memberPubs, memberCount, true)
	require.Equal(t, threshold, signCount)
	require.Equal(t, memberCount, valCount)
	require.True(t, signCount >= valCount*2/3+1)
}
