// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// 安全漏洞修复回归测试(executor 级):
// wasm dapp 向合约暴露的宿主函数 execTransfer/execTransferFrozen
// (resolver.go -> callback.go)此前对合约传入的 from/to 字符串不做任何归一化,
// 直接调用 acc.ExecTransfer/ExecTransferFrozen。account 层用原始字符串比较
// from==to 做自我转账检查, 而存储 key 经 address.FormatAddrKey 归一化为小写,
// 导致同一 eth 地址的大小写变体可绕过自我转账检查, 同 key 覆盖后余额不减反增,
// 凭空造币。
// 修复: 分叉 ForkWasmFixAddrNormalize 开启后, callback 层先用
// common.FmtEthAddressWithFork 将 from/to 归一化, 使大小写变体收敛为同一字符串,
// account 层 from==to 检查必然命中, 返回 ErrSendSameToRecv。

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// 测试用 eth 地址: lower 与 mixed 为同一地址的两种大小写形式
const (
	vulnFixWasmLowerAddr = "0x1a9a5a0bbe37a15e0cf2bd00b65a4a5ce07a3d01"
	vulnFixWasmMixedAddr = "0x1A9A5a0bbe37a15E0cf2bd00b65a4a5ce07a3d01"
	vulnFixWasmOtherAddr = "0x2b8c1e6f4d3a2b1c0f9e8d7c6b5a493827160504"
)

// initVulnFixWasmCB 构造回调测试环境, forkHeight<=0 时分叉取默认注册高度(0, 始终生效)
func initVulnFixWasmCB(t *testing.T, forkHeight int64) (*Wasm, *account.DB, string) {
	testCfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	if forkHeight > 0 {
		testCfg.SetDappFork(types2.WasmX, types2.ForkWasmFixAddrNormalize, forkHeight)
	}
	dir, ldb, kvdb := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dir, ldb) })

	execAddr := address.ExecAddress(testCfg.ExecName(types2.WasmX))
	acc, err := account.NewAccountDB(testCfg, "coins", "bty", kvdb)
	require.Nil(t, err)
	//为测试地址准备主账户余额, 供 transferToExec 充值执行器账户
	acc.SaveAccount(&types.Account{Balance: 100 * types.DefaultCoinPrecision, Addr: vulnFixWasmLowerAddr})
	acc.SaveAccount(&types.Account{Balance: 100 * types.DefaultCoinPrecision, Addr: vulnFixWasmOtherAddr})

	wasm := newWasm().(*Wasm)
	wasm.SetCoinsAccount(acc)
	wasm.SetStateDB(kvdb)
	wasm.SetLocalDB(kvdb)
	wasm.execAddr = execAddr
	api := mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(testCfg)
	api.On("GetRandNum", mock.Anything).Return([]byte("hello"), nil)
	wasm.SetAPI(&api)
	wasmCB = wasm
	t.Cleanup(func() { wasmCB = nil })
	return wasm, acc, execAddr
}

func TestVulnFixWasmExecTransferAddrNormalize(t *testing.T) {
	amount := types.DefaultCoinPrecision

	//分叉后: 同一 eth 地址的大小写变体自我转账被归一化后命中 from==to 检查, 报错且余额不变
	t.Run("post-fork reject mixed-case self transfer", func(t *testing.T) {
		_, acc, execAddr := initVulnFixWasmCB(t, 0)
		require.Nil(t, transferToExec(vulnFixWasmLowerAddr, execAddr, amount))

		err := execTransfer(vulnFixWasmLowerAddr, vulnFixWasmMixedAddr, amount)
		require.Equal(t, types.ErrSendSameToRecv, err, "归一化后应命中 account 层自我转账检查")
		require.Equal(t, amount, acc.LoadExecAccount(vulnFixWasmLowerAddr, execAddr).Balance, "余额不变, 未造币")
		require.Equal(t, amount, acc.LoadExecAccount(vulnFixWasmMixedAddr, execAddr).Balance, "归一化后同一账户, 余额不变")
	})

	//分叉后: execTransferFrozen 同样归一化, 大小写变体自我转账被拒绝
	t.Run("post-fork reject mixed-case self transferFrozen", func(t *testing.T) {
		_, acc, execAddr := initVulnFixWasmCB(t, 0)
		require.Nil(t, transferToExec(vulnFixWasmLowerAddr, execAddr, amount))
		require.Nil(t, execFrozen(vulnFixWasmLowerAddr, amount))

		err := execTransferFrozen(vulnFixWasmLowerAddr, vulnFixWasmMixedAddr, amount)
		require.Equal(t, types.ErrSendSameToRecv, err, "归一化后应命中 account 层自我转账检查")
		execAcc := acc.LoadExecAccount(vulnFixWasmLowerAddr, execAddr)
		require.Equal(t, int64(0), execAcc.Balance, "余额不变")
		require.Equal(t, amount, execAcc.Frozen, "冻结余额不变, 未造币")
	})

	//分叉后: 正常的不同地址转账不受影响, 混合大小写的 to 归一化后入账到同一账户
	t.Run("post-fork normal transfer unaffected", func(t *testing.T) {
		_, acc, execAddr := initVulnFixWasmCB(t, 0)
		require.Nil(t, transferToExec(vulnFixWasmLowerAddr, execAddr, amount))

		otherMixed := "0x2B8c1e6f4d3a2b1c0f9e8d7c6b5a493827160504"
		require.Nil(t, execTransfer(vulnFixWasmLowerAddr, otherMixed, amount/2))
		require.Equal(t, amount/2, acc.LoadExecAccount(vulnFixWasmLowerAddr, execAddr).Balance)
		require.Equal(t, amount/2, acc.LoadExecAccount(vulnFixWasmOtherAddr, execAddr).Balance)
	})

	//分叉前: 保持原行为(不做归一化), 大小写变体自我转账仍可执行, 保证链上共识兼容
	t.Run("pre-fork behavior unchanged", func(t *testing.T) {
		wasm, acc, execAddr := initVulnFixWasmCB(t, 1000000)
		require.False(t, wasm.GetAPI().GetConfig().IsDappFork(wasm.GetHeight(), types2.WasmX, types2.ForkWasmFixAddrNormalize))
		require.Nil(t, transferToExec(vulnFixWasmLowerAddr, execAddr, amount))

		require.Nil(t, execTransfer(vulnFixWasmLowerAddr, vulnFixWasmMixedAddr, amount),
			"分叉前保持原行为, 不做归一化")
		require.Equal(t, 2*amount, acc.LoadExecAccount(vulnFixWasmLowerAddr, execAddr).Balance,
			"分叉前 from==to 字符串检查被绕过, 同 key 覆盖后余额翻倍(原漏洞行为)")
	})
}
