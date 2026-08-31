// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor_test

// 链上端到端回归测试（testnode）：验证 finishCreate 的全局 symbol 唯一性修复。
//
// 漏洞：tokendb.go finishCreate 只按 (symbol, owner) 查 per-owner 记录并检查状态，
// 缺少像 preCreate 那样对全局 symbol 唯一性（checkTokenExist）的检查。同一 symbol
// 被两个不同 owner 先后 preCreate 后，两个 owner 都能被 finishCreate，每次 finish 都
// 执行 GenesisInit(owner, Total)，导致同一 symbol 实际发行 N * Total。
//
// 与 token_finishcheck_test.go（函数调用级单测）的区别：单测直接 exec.Exec(tx, 1)
// 并手工 apply receipt 的 KV，绕过了真实链的 mempool -> 共识 -> executor 完整链路。
// 本测试用 util/testnode 拉起完整链节点，走真实交易上链路径验证：
//   - fork ForkTokenFinishCheck 生效后，同一 symbol 的第二次 finishCreate 被拒绝（不超发）；
//   - fork 高度之前，旧行为保留（第二次 finishCreate 成功，超发，用于对照复现 bug）。

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/executor"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	pty "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/stretchr/testify/require"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
)

// e2eTokenCfg 构造 token fork 配置。Title 由 "local" 替换为 "chain33"，否则
// needSetForkZero 会把所有 fork 高度强制置 0，无法测试 fork 之前的行为。
func e2eTokenCfg(forkHeight int64) *types.Chain33Config {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	cfg.SetDappFork(pty.TokenX, pty.ForkTokenFinishCheckX, forkHeight)
	return cfg
}

// e2eSendTokenRaw 通过 token 的 jrpc 构造未签名交易，再签名上链，等待并返回交易明细。
func e2eSendTokenRaw(t *testing.T, mocker *testnode.Chain33Mock, method string,
	param interface{}, priv crypto.PrivKey) *rpctypes.TransactionDetail {
	t.Helper()
	var txhex string
	require.Nil(t, mocker.GetJSONC().Call(method, param, &txhex))
	hash, err := mocker.SendAndSign(priv, txhex)
	require.Nil(t, err)
	detail, err := mocker.WaitTx(hash)
	require.Nil(t, err)
	return detail
}

// e2eSendManage 通过 manage 交易配置 token 的 manage 类配置（如 token-blacklist）。
func e2eSendManage(t *testing.T, mocker *testnode.Chain33Mock, cfg *types.Chain33Config,
	key, op, value string) {
	t.Helper()
	tx := util.CreateManageTx(cfg, mocker.GetHotKey(), key, op, value)
	require.NotNil(t, tx)
	reply, err := mocker.GetAPI().SendTx(tx)
	require.Nil(t, err)
	detail, err := mocker.WaitTx(reply.GetMsg())
	require.Nil(t, err)
	require.Equal(t, int32(types.ExecOk), detail.Receipt.Ty, "manage %s %s failed", key, op)
}

// e2eTokenBalance 查询 token 账户命名空间 mavl-token-<symbol>- 下 addr 的余额。
func e2eTokenBalance(t *testing.T, mocker *testnode.Chain33Mock, cfg *types.Chain33Config,
	stateHash []byte, symbol, addr string) int64 {
	t.Helper()
	statedb := executor.NewStateDB(mocker.GetClient(), stateHash, nil, nil)
	accDB, err := account.NewAccountDB(cfg, "token", symbol, statedb)
	require.Nil(t, err)
	return accDB.LoadAccount(addr).Balance
}

// e2ePreCreate 两个 owner 分别 preCreate 同一 symbol（价格 0），返回最后一个交易的区块 StateHash。
func e2ePreCreate(t *testing.T, mocker *testnode.Chain33Mock, symbol string,
	total int64, ownerA, ownerB string) {
	t.Helper()
	preA := &pty.TokenPreCreate{Name: symbol, Symbol: symbol, Introduction: symbol,
		Total: total, Price: 0, Owner: ownerA}
	detail := e2eSendTokenRaw(t, mocker, "token.CreateRawTokenPreCreateTx", preA, mocker.GetGenesisKey())
	require.Equal(t, int32(types.ExecOk), detail.Receipt.Ty, "ownerA preCreate should succeed")

	preB := &pty.TokenPreCreate{Name: symbol, Symbol: symbol, Introduction: symbol,
		Total: total, Price: 0, Owner: ownerB}
	detail = e2eSendTokenRaw(t, mocker, "token.CreateRawTokenPreCreateTx", preB, mocker.GetGenesisKey())
	require.Equal(t, int32(types.ExecOk), detail.Receipt.Ty, "ownerB preCreate same symbol should succeed")
}

// TestVulnE2ETokenDupFinishCreateRejected 验证 ForkTokenFinishCheck 生效后：
// 两个 owner precreate 同一 symbol，第一次 finishCreate 成功，第二次被拒绝，
// 只有第一个 owner 持有 Total，不超发。
func TestVulnE2ETokenDupFinishCreateRejected(t *testing.T) {
	cfg := e2eTokenCfg(0) // fork 高度 0，立即生效
	mocker := testnode.NewWithConfig(cfg, nil)
	defer mocker.Close()
	mocker.Listen()

	// 给热键（审批人/超级管理员）拨款，用于支付 manage 交易与 finishCreate 的手续费
	require.Nil(t, mocker.SendHot())

	symbol := "DUPREJECT"
	total := int64(10000 * types.DefaultCoinPrecision)
	ownerA := mocker.GetGenesisAddress()
	ownerB := mocker.GetHotAddress()

	// 初始化 token-blacklist manage 配置（ForkTokenBlackListX 高度 0 生效，preCreate 必须能读到）
	e2eSendManage(t, mocker, cfg, "token-blacklist", "add", "BTY")
	e2ePreCreate(t, mocker, symbol, total, ownerA, ownerB)

	// 第一次 finishCreate 成功，写入全局 token 记录
	finA := &pty.TokenFinishCreate{Symbol: symbol, Owner: ownerA}
	detail := e2eSendTokenRaw(t, mocker, "token.CreateRawTokenFinishTx", finA, mocker.GetHotKey())
	require.Equal(t, int32(types.ExecOk), detail.Receipt.Ty, "first finishCreate should succeed")

	// 第二次 finishCreate 同一 symbol，被全局唯一性检查拒绝
	finB := &pty.TokenFinishCreate{Symbol: symbol, Owner: ownerB}
	detail = e2eSendTokenRaw(t, mocker, "token.CreateRawTokenFinishTx", finB, mocker.GetHotKey())
	require.NotEqual(t, int32(types.ExecOk), detail.Receipt.Ty, "duplicate finishCreate must be rejected")
	stateHash := mocker.GetBlock(detail.Height).StateHash

	// 只有 ownerA 持有 Total，ownerB 为 0，无超发
	balA := e2eTokenBalance(t, mocker, cfg, stateHash, symbol, ownerA)
	balB := e2eTokenBalance(t, mocker, cfg, stateHash, symbol, ownerB)
	t.Logf("symbol=%s declared Total=%d, ownerA=%d, ownerB=%d", symbol, total, balA, balB)
	require.Equal(t, total, balA)
	require.Equal(t, int64(0), balB)
}

// TestVulnE2ETokenDupFinishCreatePreFork 验证 fork 高度之前的旧行为：
// 同一 symbol 被重复 finishCreate 仍然成功，实际发行 2 * Total（超发对照，复现 bug）。
func TestVulnE2ETokenDupFinishCreatePreFork(t *testing.T) {
	cfg := e2eTokenCfg(2000000) // fork 生效高度高于测试链高度，旧行为保留
	mocker := testnode.NewWithConfig(cfg, nil)
	defer mocker.Close()
	mocker.Listen()

	require.Nil(t, mocker.SendHot())

	symbol := "DUPPREFORK"
	total := int64(10000 * types.DefaultCoinPrecision)
	ownerA := mocker.GetGenesisAddress()
	ownerB := mocker.GetHotAddress()

	e2eSendManage(t, mocker, cfg, "token-blacklist", "add", "BTY")
	e2ePreCreate(t, mocker, symbol, total, ownerA, ownerB)

	finA := &pty.TokenFinishCreate{Symbol: symbol, Owner: ownerA}
	detail := e2eSendTokenRaw(t, mocker, "token.CreateRawTokenFinishTx", finA, mocker.GetHotKey())
	require.Equal(t, int32(types.ExecOk), detail.Receipt.Ty, "first finishCreate should succeed")

	// fork 之前：无全局 symbol 唯一性检查，第二次 finishCreate 仍然成功
	finB := &pty.TokenFinishCreate{Symbol: symbol, Owner: ownerB}
	detail = e2eSendTokenRaw(t, mocker, "token.CreateRawTokenFinishTx", finB, mocker.GetHotKey())
	require.Equal(t, int32(types.ExecOk), detail.Receipt.Ty, "pre-fork duplicate finishCreate keeps old behavior")
	stateHash := mocker.GetBlock(detail.Height).StateHash

	// 两个 owner 各持有 Total，实际发行 2 * Total（超发）
	balA := e2eTokenBalance(t, mocker, cfg, stateHash, symbol, ownerA)
	balB := e2eTokenBalance(t, mocker, cfg, stateHash, symbol, ownerB)
	t.Logf("symbol=%s declared Total=%d, ownerA=%d, ownerB=%d, actual issued=%d (2x Total)",
		symbol, total, balA, balB, balA+balB)
	require.Equal(t, total, balA)
	require.Equal(t, total, balB)
}
