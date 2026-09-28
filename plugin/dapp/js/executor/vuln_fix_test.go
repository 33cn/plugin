// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// 安全漏洞修复回归测试(executor 级):
// js dapp 的 exec_transfer 绑定(account.go execTransferFunc)此前对合约传入的
// from/to 仅做 address.CheckAddress 校验, 不做 eth 地址归一化。account 层
// ExecTransfer 用原始字符串比较 from==to 做自我转账检查, 而存储 key 经
// address.FormatAddrKey 归一化为小写, 导致同一 eth 地址的大小写变体可绕过
// 自我转账检查, 同 key 覆盖后余额不减反增, 凭空造币。
// 修复: 分叉 ForkJSFixAddrNormalize 开启后, execTransferFunc 在调用 account 层前
// 先用 common.FmtEthAddressWithFork 将 from/to 归一化, 使大小写变体收敛为同一字符串,
// account 层 from==to 检查必然命中, 返回 ErrSendSameToRecv。

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// 测试用 eth 地址: lower 与 mixed 为同一地址的两种大小写形式
const (
	vulnFixJSLowerAddr = "0x1a9a5a0bbe37a15e0cf2bd00b65a4a5ce07a3d01"
	vulnFixJSMixedAddr = "0x1A9A5a0bbe37a15E0cf2bd00b65a4a5ce07a3d01"
	vulnFixJSOtherAddr = "0x2b8c1e6f4d3a2b1c0f9e8d7c6b5a493827160504"
)

var vulnFixJSCode = `
function Init(context) {
    this.kvc = new kvcreator("init")
    this.context = context
    return this.kvc.receipt()
}

function ExecInit() {
    this.acc = new account(this.kvc, "coins", "bty")
}

Exec.prototype.mint = function(args) {
    var err = this.acc.execTransfer(this.name, args.from, args.to, args.amount)
    throwerr(err, "execTransfer")
    return this.kvc.receipt()
}
`

// newVulnFixJSConfig 构造启用 eth 地址的测试配置, forkHeight<=0 时分叉取默认注册高度(0, 始终生效)
func newVulnFixJSConfig(forkHeight int64) *types.Chain33Config {
	//Title=local 的配置所有分叉强制为高度0(needSetForkZero), 无法测试分叉前行为,
	//参照 wasm_test.go 将 Title 替换为 chain33, 使 SetDappFork 生效
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	//默认配置中 eth 地址驱动禁用([address.enableHeight] eth=-2), 测试启用 eth 地址,
	//模拟支持 eth 地址格式的链(如开启 EVM 的链)
	cfg.GetModuleConfig().Address.EnableHeight["eth"] = 0
	if forkHeight > 0 {
		cfg.SetDappFork(ptypes.JsX, ptypes.ForkJSFixAddrNormalize, forkHeight)
	}
	//地址驱动启用高度为全局配置, 需显式生效
	address.Init(cfg.GetModuleConfig().Address)
	return cfg
}

// initVulnFixJSExec 部署测试合约, 与 jsvm_test.go 的 initExec 逻辑一致, 但使用自定义配置
func initVulnFixJSExec(t *testing.T, cfg *types.Chain33Config, ldb db.DB, kvdb db.KVDB) *js {
	Init(ptypes.JsX, cfg, nil)

	e := newjs().(*js)
	e.SetEnv(1, time.Now().Unix(), 1)
	mockapi := &mocks.QueueProtocolAPI{}
	mockapi.On("Query", "ticket", "RandNumHash", mock.Anything).Return(&types.ReplyHash{Hash: []byte("hello")}, nil)
	mockapi.On("GetConfig", mock.Anything).Return(cfg, nil)
	e.SetAPI(mockapi)
	gclient, err := grpcclient.NewMainChainClient(cfg, "")
	require.Nil(t, err)
	e.SetExecutorAPI(mockapi, gclient)
	e.SetLocalDB(kvdb)
	e.SetStateDB(kvdb)
	//合约名需全局唯一: 包级 codecache 以合约名为 key 缓存 VM, 与其他测试重名会加载到错误代码
	c, tx := createCodeTx("vulnfixmint", vulnFixJSCode)

	// set config key
	item := &types.ConfigItem{
		Key:  "mavl-manage-js-creator",
		Addr: tx.From(),
		Value: &types.ConfigItem_Arr{
			Arr: &types.ArrayConfig{Value: []string{tx.From()}},
		},
	}
	kvdb.Set([]byte(item.Key), types.Encode(item))

	receipt, err := e.Exec_Create(c, tx, 0)
	require.Nil(t, err)
	util.SaveKVList(ldb, receipt.KV)
	return e
}

// vulnFixJSFundExec 向合约执行器账户充值(amount 同时作为主账户初始余额)
func vulnFixJSFundExec(t *testing.T, cfg *types.Chain33Config, kvdb db.KVDB, execer, addr string, amount int64) *account.DB {
	acc, err := account.NewAccountDB(cfg, "coins", "bty", kvdb)
	require.Nil(t, err)
	acc.SaveAccount(&types.Account{Balance: amount, Addr: addr})
	_, err = acc.TransferToExec(addr, address.ExecAddress(execer), amount)
	require.Nil(t, err)
	return acc
}

func vulnFixJSCallMint(t *testing.T, e *js, from, to string, amount int64) error {
	args, err := json.Marshal(map[string]interface{}{"from": from, "to": to, "amount": amount})
	require.Nil(t, err)
	call, tx := callCodeTx("vulnfixmint", "mint", string(args))
	_, err = e.Exec_Call(call, tx, 0)
	return err
}

func TestVulnFixJSExecTransferAddrNormalize(t *testing.T) {
	userExec := "user." + ptypes.JsX + ".vulnfixmint"
	execAddr := address.ExecAddress(userExec)
	amount := types.DefaultCoinPrecision

	//分叉后: 同一 eth 地址的大小写变体自我转账被归一化后命中 from==to 检查, 交易报错且余额不变
	t.Run("post-fork reject mixed-case self transfer", func(t *testing.T) {
		dir, ldb, kvdb := util.CreateTestDB()
		defer util.CloseTestDB(dir, ldb)
		cfg := newVulnFixJSConfig(0)
		e := initVulnFixJSExec(t, cfg, ldb, kvdb)
		acc := vulnFixJSFundExec(t, cfg, kvdb, userExec, vulnFixJSLowerAddr, amount)

		err := vulnFixJSCallMint(t, e, vulnFixJSLowerAddr, vulnFixJSMixedAddr, amount)
		require.NotNil(t, err, "大小写变体自我转账应被拒绝")
		require.True(t, strings.Contains(err.Error(), types.ErrSendSameToRecv.Error()),
			"归一化后应命中 account 层自我转账检查, 实际错误: %v", err)
		require.Equal(t, amount, acc.LoadExecAccount(vulnFixJSLowerAddr, execAddr).Balance, "余额不变, 未造币")
		require.Equal(t, amount, acc.LoadExecAccount(vulnFixJSMixedAddr, execAddr).Balance, "归一化后同一账户, 余额不变")
	})

	//分叉后: 正常的不同地址转账不受影响, 混合大小写的 to 归一化后入账到同一账户
	t.Run("post-fork normal transfer unaffected", func(t *testing.T) {
		dir, ldb, kvdb := util.CreateTestDB()
		defer util.CloseTestDB(dir, ldb)
		cfg := newVulnFixJSConfig(0)
		e := initVulnFixJSExec(t, cfg, ldb, kvdb)
		acc := vulnFixJSFundExec(t, cfg, kvdb, userExec, vulnFixJSLowerAddr, amount)

		otherMixed := "0x2B8c1e6f4d3a2b1c0f9e8d7c6b5a493827160504"
		err := vulnFixJSCallMint(t, e, vulnFixJSLowerAddr, otherMixed, amount/2)
		require.Nil(t, err)
		require.Equal(t, amount/2, acc.LoadExecAccount(vulnFixJSLowerAddr, execAddr).Balance)
		require.Equal(t, amount/2, acc.LoadExecAccount(vulnFixJSOtherAddr, execAddr).Balance)
	})

	//分叉前: 保持原行为(不做归一化), 大小写变体自我转账仍可执行, 保证链上共识兼容
	t.Run("pre-fork behavior unchanged", func(t *testing.T) {
		dir, ldb, kvdb := util.CreateTestDB()
		defer util.CloseTestDB(dir, ldb)
		cfg := newVulnFixJSConfig(1000000)
		require.False(t, cfg.IsDappFork(1, ptypes.JsX, ptypes.ForkJSFixAddrNormalize))
		e := initVulnFixJSExec(t, cfg, ldb, kvdb)
		acc := vulnFixJSFundExec(t, cfg, kvdb, userExec, vulnFixJSLowerAddr, amount)

		err := vulnFixJSCallMint(t, e, vulnFixJSLowerAddr, vulnFixJSMixedAddr, amount)
		require.Nil(t, err, "分叉前保持原行为, 不做归一化")
		require.Equal(t, 2*amount, acc.LoadExecAccount(vulnFixJSLowerAddr, execAddr).Balance,
			"分叉前 from==to 字符串检查被绕过, 同 key 覆盖后余额翻倍(原漏洞行为)")
	})
}
