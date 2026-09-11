// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor_test

import (
	"errors"
	"math/rand"
	"strings"
	"testing"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	rpctypes "github.com/33cn/chain33/rpc/types"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	_ "github.com/33cn/plugin/plugin/dapp/evm"
	"github.com/stretchr/testify/require"
)

// 本文件走 testnode 完整节点栈，依赖 plugin/go.mod 中的 github.com/33cn/chain33
// （当前为 v1.71.0，已含 mver 多版本黑名单）。
//
// 场景：节点按 mver 启动后
//   V1 拦截指定地址 → V2 空名单解除拦截 → V3 把同一地址重新拉黑。
// coins 与 EVM（ContractAddr 命中）都走 queue SendTx 和 JSON-RPC SendTransaction：
// 拦截时 mempool 拒绝，放行时 WaitTx 确认上链。
// solo 只在 mempool 有交易时出块，跨高度用 none 交易推进，不能空等。
// 创世地址 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt 不能进名单，否则出块和 SendHot 会被自己拦住。

const (
	forkV1Height = int64(2)
	forkV2Height = int64(5)
	forkV3Height = int64(15)

	// TestPrivkeyList[2]，不是 genesis / hot
	blockedBTCAddr = "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"
	blockedETHAddr = "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	defaultBlacklistSection = "[mver.blacklist]\naccountBlacklist=[]\n"
)

func TestNodeBlacklistV2ClearThenV3Restore(t *testing.T) {
	cfg := newNodeBlacklistCfg(t)
	mock := testnode.NewWithConfig(cfg, nil)
	defer mock.Close()
	mock.Listen()

	require.NoError(t, mock.SendHot(), "genesis 转给 hot 必须在名单之外")

	t.Run("节点启动后配置快照", func(t *testing.T) {
		assertCfgSnapshot(t, cfg)
	})

	t.Run("V1 生效后拦截", func(t *testing.T) {
		mineTo(t, mock, forkV1Height-1)
		assertNextInWindow(t, mock, forkV1Height, forkV2Height)
		assertSendBlocked(t, mock, true)
	})

	t.Run("V2 空名单后放行", func(t *testing.T) {
		mineTo(t, mock, forkV2Height)
		assertNextInWindow(t, mock, forkV2Height, forkV3Height)
		assertSendBlocked(t, mock, false)
	})

	t.Run("V3 重新拉黑后拦截", func(t *testing.T) {
		mineTo(t, mock, forkV3Height)
		require.GreaterOrEqual(t, nextHeight(t, mock), forkV3Height)
		assertSendBlocked(t, mock, true)
	})
}

func newNodeBlacklistCfg(t *testing.T) *types.Chain33Config {
	t.Helper()
	cfgstring := types.GetDefaultCfgstring()
	cfgstring = strings.Replace(cfgstring, `Title="local"`, `Title="chain33"`+"\nDisableForkCheck=true", 1)
	cfgstring = strings.Replace(cfgstring, defaultBlacklistSection, "", 1)
	cfgstring += nodeBlacklistTOML()
	// V2/V3 未在 RegisterSystemFork 里登记，toml 里写出高度后由 initForkConfig 注入。
	// Title 不能是 local，否则 SetAllFork(0) 会把三个版本压到同一高度，只留下 V3。
	return types.NewChain33Config(cfgstring)
}

func nodeBlacklistTOML() string {
	return `
[mver.blacklist]
accountBlacklist=[]
[mver.blacklist.ForkAccountBlacklist]
accountBlacklist=["` + blockedBTCAddr + `","` + blockedETHAddr + `"]
[mver.blacklist.ForkAccountBlacklistV2]
accountBlacklist=[]
[mver.blacklist.ForkAccountBlacklistV3]
accountBlacklist=["` + blockedBTCAddr + `","` + blockedETHAddr + `"]

[fork.system]
ForkAccountBlacklist=2
ForkAccountBlacklistV2=5
ForkAccountBlacklistV3=15

[fork.sub.coins]
Enable=0
ForkFriendExecer=0

[fork.sub.evm]
Enable=0
`
}

func assertCfgSnapshot(t *testing.T, cfg *types.Chain33Config) {
	t.Helper()
	hot := address.PubKeyToAddr(address.DefaultID, util.TestPrivkeyList[0].PubKey().Bytes())
	priv := util.TestPrivkeyList[0]

	cases := []struct {
		name    string
		height  int64
		blocked bool
	}{
		{"V1 前放行", forkV1Height - 1, false},
		{"V1 拦截", forkV1Height, true},
		{"V2 前仍拦截", forkV2Height - 1, true},
		{"V2 空名单放行", forkV2Height, false},
		{"V3 前仍放行", forkV3Height - 1, false},
		{"V3 重新拦截", forkV3Height, true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.blocked, cfg.IsBlockedAccount(blockedBTCAddr, c.height), "btc height=%d", c.height)
			require.Equal(t, c.blocked, cfg.IsBlockedAccount(blockedETHAddr, c.height), "eth height=%d", c.height)
			require.False(t, cfg.IsBlockedAccount(hot, c.height), "hot 不得进名单")

			coinsErr := types.CheckTxBlockedAccount(cfg, c.height, coinsTx(cfg, priv, blockedBTCAddr))
			evmErr := types.CheckTxBlockedAccount(cfg, c.height, evmTx(cfg, priv, blockedETHAddr))
			if c.blocked {
				require.Error(t, coinsErr)
				require.Error(t, evmErr)
				require.True(t, errors.Is(coinsErr, types.ErrBlockedAccount))
				require.True(t, errors.Is(evmErr, types.ErrBlockedAccount))
				return
			}
			require.NoError(t, coinsErr)
			require.NoError(t, evmErr)
		})
	}
}

func assertSendBlocked(t *testing.T, mock *testnode.Chain33Mock, wantBlocked bool) {
	t.Helper()
	next := nextHeight(t, mock)
	var onChain []namedHash
	t.Run("queue", func(t *testing.T) {
		onChain = append(onChain, assertSendVia(t, mock, next, wantBlocked, "queue", sendTx)...)
	})
	t.Run("jsonrpc", func(t *testing.T) {
		onChain = append(onChain, assertSendVia(t, mock, next, wantBlocked, "jsonrpc", sendTxRPC)...)
	})
	if wantBlocked {
		return
	}
	for _, item := range onChain {
		waitOnChain(t, mock, item.hash, item.kind)
	}
}

type namedHash struct {
	kind string
	hash []byte
}

func assertSendVia(t *testing.T, mock *testnode.Chain33Mock, next int64, wantBlocked bool, via string, send sendTxFn) []namedHash {
	t.Helper()
	cfg := mock.GetClient().GetConfig()
	hot := mock.GetHotKey()

	coinsHash, coinsErr := send(t, mock, coinsTx(cfg, hot, blockedBTCAddr))
	evmHash, evmErr := send(t, mock, evmTx(cfg, hot, blockedETHAddr))
	if wantBlocked {
		require.True(t, isBlockedErr(coinsErr), "%s height=%d coins 应拦截, err=%v", via, next, coinsErr)
		require.True(t, isBlockedErr(evmErr), "%s height=%d EVM 应拦截, err=%v", via, next, evmErr)
		return nil
	}
	require.NoError(t, coinsErr, "%s height=%d coins 应放行", via, next)
	require.NoError(t, evmErr, "%s height=%d EVM 应放行", via, next)
	return []namedHash{{via + " coins", coinsHash}, {via + " EVM", evmHash}}
}

func waitOnChain(t *testing.T, mock *testnode.Chain33Mock, hash []byte, kind string) {
	t.Helper()
	detail, err := mock.WaitTx(hash)
	require.NoError(t, err, "%s 必须能上链", kind)
	require.NotNil(t, detail, "%s WaitTx 无详情", kind)
	require.NotNil(t, detail.Receipt, "%s 上链后必须有 receipt", kind)
}

func isBlockedErr(err error) bool {
	return err != nil && (errors.Is(err, types.ErrBlockedAccount) ||
		strings.Contains(err.Error(), types.ErrBlockedAccount.Error()))
}

func mineTo(t *testing.T, mock *testnode.Chain33Mock, lastHeight int64) {
	t.Helper()
	cfg := mock.GetClient().GetConfig()
	for {
		header, err := mock.GetAPI().GetLastHeader()
		require.NoError(t, err)
		if header.Height >= lastHeight {
			return
		}
		tx := util.CreateNoneTx(cfg, mock.GetHotKey())
		reply, err := mock.GetAPI().SendTx(tx)
		require.NoError(t, err, "推进高度失败 current=%d target=%d", header.Height, lastHeight)
		mock.SetLastSend(reply.GetMsg())
		_, err = mock.WaitTx(reply.GetMsg())
		require.NoError(t, err)
	}
}

func assertNextInWindow(t *testing.T, mock *testnode.Chain33Mock, atLeast, before int64) {
	t.Helper()
	next := nextHeight(t, mock)
	require.GreaterOrEqual(t, next, atLeast, "尚未到达目标高度")
	require.Less(t, next, before, "已经跨过当前分叉窗口 next=%d", next)
}

func nextHeight(t *testing.T, mock *testnode.Chain33Mock) int64 {
	t.Helper()
	header, err := mock.GetAPI().GetLastHeader()
	require.NoError(t, err)
	return header.Height + 1
}

func sendTx(t *testing.T, mock *testnode.Chain33Mock, tx *types.Transaction) ([]byte, error) {
	t.Helper()
	reply, err := mock.GetAPI().SendTx(tx)
	if err != nil {
		return nil, err
	}
	mock.SetLastSend(reply.GetMsg())
	return reply.GetMsg(), nil
}

type sendTxFn func(*testing.T, *testnode.Chain33Mock, *types.Transaction) ([]byte, error)

func sendTxRPC(t *testing.T, mock *testnode.Chain33Mock, tx *types.Transaction) ([]byte, error) {
	t.Helper()
	jsonc := mock.GetJSONC()
	require.NotNil(t, jsonc, "JSON-RPC 客户端未就绪")
	var txhash string
	err := jsonc.Call("Chain33.SendTransaction", &rpctypes.RawParm{Data: common.ToHex(types.Encode(tx))}, &txhash)
	if err != nil {
		return nil, err
	}
	hash, err := common.FromHex(txhash)
	if err != nil {
		return nil, err
	}
	mock.SetLastSend(hash)
	return hash, nil
}

func coinsTx(cfg *types.Chain33Config, priv crypto.PrivKey, to string) *types.Transaction {
	return util.CreateCoinsTx(cfg, priv, to, types.DefaultCoinPrecision)
}

func evmTx(cfg *types.Chain33Config, priv crypto.PrivKey, contractAddr string) *types.Transaction {
	execAddr := address.ExecAddress("evm")
	action := &types.EVMContractAction4Chain33{
		Amount:       0,
		GasLimit:     10000,
		GasPrice:     1,
		ContractAddr: contractAddr,
	}
	tx := &types.Transaction{
		ChainID: cfg.GetChainID(),
		Execer:  []byte("evm"),
		Payload: types.Encode(action),
		Fee:     1e6,
		To:      execAddr,
		Nonce:   rand.Int63(),
	}
	tx.Sign(types.SECP256K1, priv)
	return tx
}
