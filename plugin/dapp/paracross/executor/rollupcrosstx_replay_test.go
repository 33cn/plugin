// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	dbmock "github.com/33cn/chain33/common/db/mocks"
	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

// 回归 paracross RollupCrossTx 结算的「同轮重放」漏洞(见 rollup types.ForkSecurityHardening 之后新增的
// paracross 侧水位 mavl-paracross-rollupcrossdone-<title>):
//
// 此前 rollupCrossTx 的准入只校验 status.CommitRound == commit.CommitRound, 而 status.CommitRound 要到
// 【下一轮】rollup commit 落库时(rollup/executor/exec.go 的 Exec_Commit)才推进。因此在被下一轮提交打断之前,
// 任何人都可以复制链上公开的 RollupCrossTx payload、改 nonce 重新签名(无需是验证人, 执行器不校验签名者),
// 对同一轮反复结算 —— 反复从该 title 的托管池支取(execWithdraw) / 重复铸币(execCreateAsset)。
//
// 本文件断言的关键点不是"返回了错误", 而是"托管池余额没有被第二次拿走"。
const (
	// rollupReplayTitle 规范平行链 title, 与跨链交易 execer(user.p.test.paracross)精确对应
	rollupReplayTitle = "user.p.test."
	// rollupReplayCoin 一个主币最小精度单位
	rollupReplayCoin = int64(100000000)
)

// rollupReplayEnv 一条 title 的 rollup 结算测试环境
type rollupReplayEnv struct {
	t    *testing.T
	cfg  *types.Chain33Config
	exec *Paracross
	db   dbm.KV
	// acc 与 execWithdraw 内部使用的账户库一致(createAccount(cfg, db, "coins", "bty"))
	acc *account.DB

	crossTx  *types.Transaction // 该轮里真实存在的跨链提现交易
	execAddr string             // paracross 执行器地址
	poolAddr string             // 该 title 的托管池: paracross 执行器地址下的 user.p.test.paracross 子账户
	amount   int64
	initial  int64
}

// newRollupReplayEnv 构造一条 title 的 rollup 结算环境:
// 主链配置、按标题建好一轮 rollup 状态、并给该 title 的托管池预置余额
func newRollupReplayEnv(t *testing.T, initial, amount int64) *rollupReplayEnv {
	cfg := chain33TestMainCfg
	// 本用例只覆盖"分叉生效"后的行为: 单测配置下 RollupX.ForkSecurityHardening 默认即生效。
	// (分叉关闭的分支在单测里不可达 —— needSetForkZero/SetDappFork 会把高度压成 0)
	require.True(t, cfg.IsDappFork(0, rtypes.RollupX, rtypes.ForkSecurityHardening),
		"precondition: rollup ForkSecurityHardening 必须在高度 0 已生效")

	stateDB, err := dbm.NewGoMemDB("state", "state", 1024)
	require.NoError(t, err)

	env := &rollupReplayEnv{
		t:        t,
		cfg:      cfg,
		db:       stateDB,
		execAddr: address.ExecAddress(pt.ParaX),                     // ExecAddress("paracross")
		poolAddr: address.ExecAddress(rollupReplayTitle + pt.ParaX), // ExecAddress("user.p.test.paracross")
		amount:   amount,
		initial:  initial,
	}
	env.crossTx = env.buildCrossTx()

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	// TxIndices 的 BlockHeight 为 0 时, 框架回退到按交易哈希查询交易
	api.On("QueryTx", mock.Anything).Return(&types.TransactionDetail{Tx: env.crossTx}, nil)

	env.exec = newParacross().(*Paracross)
	env.exec.SetAPI(api)
	env.exec.SetLocalDB(new(dbmock.KVDB))
	env.exec.SetStateDB(stateDB)
	env.exec.SetEnv(0, 0, 0)
	env.exec.SetBlockInfo([]byte(""), []byte(""), 0)

	env.acc, err = account.NewAccountDB(cfg, "coins", "bty", stateDB)
	require.NoError(t, err)
	return env
}

// buildCrossTx 构造该轮里一笔真实存在的跨链提现交易: 主链从该 title 的托管池向 Account14K 支取 amount。
// AssetExec=user.p.test.paracross 会被 getCrossAction 判定为 MainAssetWithdraw, 主链侧最终走 execWithdraw。
func (e *rollupReplayEnv) buildCrossTx() *types.Transaction {
	action := &pt.ParacrossAction{
		Ty: pt.ParacrossActionCrossAssetTransfer,
		Value: &pt.ParacrossAction_CrossAssetTransfer{
			CrossAssetTransfer: &pt.CrossAssetTransfer{
				AssetExec:   rollupReplayTitle + pt.ParaX,
				AssetSymbol: "coins.bty",
				Amount:      e.amount,
				ToAddr:      Account14K,
			},
		},
	}
	return &types.Transaction{
		Execer:  []byte(rollupReplayTitle + pt.ParaX),
		Payload: types.Encode(action),
		Nonce:   1,
	}
}

// seedPool 给该 title 的托管池预置余额
func (e *rollupReplayEnv) seedPool() {
	_, err := e.acc.ExecDeposit(e.poolAddr, e.execAddr, e.initial)
	require.NoError(e.t, err)
}

// poolBalance 读该 title 的托管池余额
func (e *rollupReplayEnv) poolBalance() int64 {
	return e.acc.LoadExecAccount(e.poolAddr, e.execAddr).Balance
}

// seedRound 写入该 title 的 rollup 状态与指定轮次的 roundinfo, 使 rollupCrossTx 能通过准入
func (e *rollupReplayEnv) seedRound(round int64) {
	status := &rtypes.RollupStatus{CommitRound: round}
	require.NoError(e.t, e.db.Set(rollupReplayStatusKey(rollupReplayTitle), types.Encode(status)))

	roundInfo := &rtypes.CommitRoundInfo{
		CommitRound:      round,
		CrossTxCheckHash: common.ToHex(CalcTxHashsHash([][]byte{e.crossTx.Hash()})),
		CrossTxResults:   "01", // 唯一一笔跨链交易执行成功
	}
	require.NoError(e.t, e.db.Set(rollupReplayRoundInfoKey(rollupReplayTitle, round), types.Encode(roundInfo)))
}

// newRollupCrossTx 构造一笔 paracross/RollupCrossTx 交易(与 rollup 节点 commit2MainChain 提交的 payload 同构),
// 用指定私钥与 nonce 签名 —— 复制同一 payload、换 nonce/签名即"另一笔"交易(mempool 不会去重, 执行器也不校验签名者)
func (e *rollupReplayEnv) newRollupCrossTx(round int64, privKey string, nonce int64) *types.Transaction {
	commit := &pt.RollupCrossTx{
		ChainTitle:  rollupReplayTitle,
		CommitRound: round,
		TxIndices: []*pt.CrossTxIndex{{
			BlockHeight: 0, // 为 0 时回退到按交易哈希 QueryTx 查询
			FilterIndex: 0,
			TxHash:      e.crossTx.Hash(),
		}},
	}
	tx := &types.Transaction{
		Execer: []byte(pt.ParaX),
		Payload: types.Encode(&pt.ParacrossAction{
			Ty:    pt.TyActionRollupCrossTx,
			Value: &pt.ParacrossAction_RollupCrossTx{RollupCrossTx: commit},
		}),
		To:    address.ExecAddress(pt.ParaX),
		Nonce: nonce,
	}
	signed, err := signTitleTx(e.t, tx, privKey)
	require.NoError(e.t, err)
	return signed
}

// execOnce 执行一次结算: 走 executor 的 Exec 分发入口(CheckTx + Exec_RollupCrossTx),
// 并在成功时把 receipt 的 KV 应用到 stateDB。
// 框架在区块执行阶段会统一落库 receipt.KV(水位就带在其中), 直接调用 rollupCrossTx 不会落库,
// 因此这里必须补齐该语义, 才能观察到水位随交易前进、失败回滚时水位不前进。
func (e *rollupReplayEnv) execOnce(tx *types.Transaction) (*types.Receipt, error) {
	receipt, err := e.exec.Exec(tx, 0)
	if err != nil || receipt == nil {
		return receipt, err
	}
	for _, kv := range receipt.KV {
		require.NoError(e.t, e.db.Set(kv.Key, kv.Value))
	}
	return receipt, nil
}

// doneRound 读该 title 已结算的最大轮次水位(生产代码同一入口)
func (e *rollupReplayEnv) doneRound() int64 {
	round, err := getRollupCrossDone(e.db, rollupReplayTitle)
	require.NoError(e.t, err)
	return round
}

// rollup 状态的存储 key 由 rollup 执行器定义(rollup/executor/kv.go), 这里按约定字面量复刻,
// 让测试独立于 rollup 包的内部实现
func rollupReplayStatusKey(title string) []byte {
	return []byte(fmt.Sprintf("%s%s-status", "mavl-rollup-", title))
}

func rollupReplayRoundInfoKey(title string, round int64) []byte {
	return []byte(fmt.Sprintf("%s%s-roundinfo-%018d", "mavl-rollup-", title, round))
}

// TestRollupCrossTxReplayRejected 同一轮(CommitRound=1)的结算此前可以被任意账户重放:
// 第二笔(复制 payload、改 nonce 换私钥重新签名)在修复后必须被拒, 且托管池余额不能再被支取一次。
func TestRollupCrossTxReplayRejected(t *testing.T) {
	env := newRollupReplayEnv(t, 100*rollupReplayCoin, 10*rollupReplayCoin)
	env.seedRound(1)
	env.seedPool()
	require.Equal(t, env.initial, env.poolBalance(), "预置的托管池余额")

	// 第一笔: 正常结算该轮
	first := env.newRollupCrossTx(1, PrivKeyA, 1)
	receipt, err := env.execOnce(first)
	require.NoError(t, err)
	require.NotNil(t, receipt)
	require.Equal(t, int32(types.ExecOk), receipt.Ty)

	assert.Equal(t, env.initial-env.amount, env.poolBalance(), "第一笔结算后托管池应只被支取一次")
	assert.Equal(t, int64(1), env.doneRound(), "结算成功后水位应推进到本轮")

	// 水位落在约定的 key 上: mavl-paracross-rollupcrossdone-<title>
	// (以下两条为附加契约检查, 用 assert 而非 require, 保证后面的关键断言始终会跑到)
	raw, err := env.db.Get([]byte("mavl-paracross-rollupcrossdone-" + rollupReplayTitle))
	if assert.NoError(t, err, "水位必须写在 mavl-paracross-rollupcrossdone-<title>") {
		var watermark types.Int64
		require.NoError(t, types.Decode(raw, &watermark))
		assert.Equal(t, int64(1), watermark.Data)
	}

	// 第二笔: 攻击者复制同一 payload, 改 nonce、换私钥重新签名 —— 一笔"不同"的交易
	second := env.newRollupCrossTx(1, PrivKeyB, 2)
	require.NotEqual(t, first.Hash(), second.Hash(), "两笔必须是不同的交易(不同 nonce/签名)")

	receipt, err = env.execOnce(second)
	assert.Nil(t, receipt)
	assert.Equal(t, ErrRollupRoundSettled, errors.Cause(err))
	// 关键断言: 钱没有再被拿走
	assert.Equal(t, env.initial-env.amount, env.poolBalance(), "重放必须不能再动托管池余额")
	assert.Equal(t, int64(1), env.doneRound(), "重放不应改变水位")
}

// TestRollupCrossTxNextRoundStillSettles 正向用例: 水位只挡"已结算过的轮次", 不能把合法流程堵死 ——
// 第 1 轮结算后水位推进到 1, 第 2 轮(rollup 侧 status/roundinfo 已前移)仍能正常结算。
func TestRollupCrossTxNextRoundStillSettles(t *testing.T) {
	env := newRollupReplayEnv(t, 100*rollupReplayCoin, 10*rollupReplayCoin)
	env.seedRound(1)
	env.seedPool()

	// 第 1 轮
	_, err := env.execOnce(env.newRollupCrossTx(1, PrivKeyA, 1))
	require.NoError(t, err)
	assert.Equal(t, int64(1), env.doneRound())

	// rollup 侧推进到第 2 轮(下一轮 commit 落库, CommitRound 前移)
	env.seedRound(2)

	// 第 2 轮仍能正常结算
	receipt, err := env.execOnce(env.newRollupCrossTx(2, PrivKeyA, 2))
	require.NoError(t, err, "下一轮结算必须仍能成功")
	require.NotNil(t, receipt)
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)
	assert.Equal(t, env.initial-2*env.amount, env.poolBalance())
	assert.Equal(t, int64(2), env.doneRound())

	// 第 2 轮同样不能被重放
	receipt, err = env.execOnce(env.newRollupCrossTx(2, PrivKeyB, 3))
	assert.Nil(t, receipt)
	assert.Equal(t, ErrRollupRoundSettled, errors.Cause(err))
	assert.Equal(t, env.initial-2*env.amount, env.poolBalance(), "重放必须不能再动托管池余额")
}
