// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// ForkRelayVerifyBtcTx 修复回归测试：
// 修复后伪造交易内容/伪造区块高度的 verify 必须被拒绝，合法 verify 不受影响，
// 分叉前保持原有行为不变。

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/common/db/mocks"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/relay/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func fixEncHeight(h int64) []byte {
	return types.Encode(&types.Int64{Data: h})
}

// 区块 100000 的真实 merkle root 及 SPV 数据(与 relaydb_test.go 既有测试一致)
func fixRealSpv() *ty.BtcSpv {
	strMerkleproof := []string{
		"e9a66845e05d5abc0ad04ec80f774a7e585c6e8db975962d069a522137b80c1d",
		"ccdafb73d8dcd0173d5d5c3c9a0770d0b3953db889dab99ef05b1907518cb815",
	}
	proofs := make([][]byte, len(strMerkleproof))
	for i, kk := range strMerkleproof {
		proofs[i], _ = btcHashStrRevers(kk)
	}
	return &ty.BtcSpv{
		BranchProof: proofs,
		TxIndex:     2,
		BlockHash:   "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506",
		Height:      100000,
		Hash:        relayRealBtcTxHash,
	}
}

func fixRealHeadEnc() []byte {
	head := &ty.BtcHeader{
		Version:    1,
		Height:     100000,
		MerkleRoot: "f3e94742aca4b5ef85488dc37c06c3282295ffec960994b2c0d5ac2a25a95766",
	}
	return types.Encode(head)
}

func fixLegitOrder() *ty.RelayOrder {
	return &ty.RelayOrder{
		XAddr:       addrBtc,
		XAmount:     29900000,
		AcceptTime:  2000,
		ConfirmTime: 3000,
	}
}

func fixLegitBtcTx() *ty.BtcTransaction {
	return &ty.BtcTransaction{
		Vout:        []*ty.Vout{{Address: addrBtc, Value: 0.299 * 1e8}},
		Time:        2500,
		BlockHeight: 100000,
		Hash:        relayRealBtcTxHash,
		RawTx:       relayRealBtcRawTx,
	}
}

// 分叉未激活的配置，用于验证分叉前行为不变
func fixPreForkCfg() *types.Chain33Config {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	cfg.SetDappFork(ty.RelayX, ty.ForkRelayVerifyBtcTx, 1000000)
	return cfg
}

// TestFixRelayVerifyBtcTxLegit 合法 verify(交易内容真实)在分叉后仍然通过
func TestFixRelayVerifyBtcTxLegit(t *testing.T) {
	kvdb := new(mocks.KVDB)
	btc := newBtcStore(kvdb)

	kvdb.On("Get", mock.Anything).Return(fixEncHeight(100100), nil).Once()
	kvdb.On("Get", mock.Anything).Return(fixRealHeadEnc(), nil).Once()

	verify := &ty.RelayVerify{Tx: fixLegitBtcTx(), Spv: fixRealSpv()}
	err := btc.verifyBtcTx(chainTestCfg, 10, verify, fixLegitOrder())
	require.NoError(t, err)
}

// TestFixRelayVerifyBtcTxFakeContent 伪造交易内容在分叉后被拒绝
func TestFixRelayVerifyBtcTxFakeContent(t *testing.T) {
	kvdb := new(mocks.KVDB)
	btc := newBtcStore(kvdb)
	order := fixLegitOrder()

	// 缺失 rawTx，无法重算哈希
	tx := fixLegitBtcTx()
	tx.RawTx = ""
	err := btc.verifyBtcTx(chainTestCfg, 10, &ty.RelayVerify{Tx: tx, Spv: fixRealSpv()}, order)
	require.Equal(t, ty.ErrRelayBtcTxHashErr, err)

	// 自报哈希与 rawTx 重算结果不一致
	tx = fixLegitBtcTx()
	tx.Hash = "7359f0868171b1d194cbee1af2f16ea598ae8fad666d9b012c8ed2b79a236ec4"
	err = btc.verifyBtcTx(chainTestCfg, 10, &ty.RelayVerify{Tx: tx, Spv: fixRealSpv()}, order)
	require.Equal(t, ty.ErrRelayBtcTxHashErr, err)

	// SPV 证明的哈希与 rawTx 重算结果不一致
	tx = fixLegitBtcTx()
	spv := fixRealSpv()
	spv.Hash = "7359f0868171b1d194cbee1af2f16ea598ae8fad666d9b012c8ed2b79a236ec4"
	err = btc.verifyBtcTx(chainTestCfg, 10, &ty.RelayVerify{Tx: tx, Spv: spv}, order)
	require.Equal(t, ty.ErrRelayBtcTxHashErr, err)

	// 订单要求的收款地址并不存在于 rawTx 真实输出中
	fakeOrder := fixLegitOrder()
	fakeOrder.XAddr = "1H8ANdafjpqYntniT3Ddxh4xPBMCSz33pj"
	fakeOrder.XAmount = 20000000 // 该地址真实输出仅为 0.01 BTC(1000000)
	err = btc.verifyBtcTx(chainTestCfg, 10, &ty.RelayVerify{Tx: fixLegitBtcTx(), Spv: fixRealSpv()}, fakeOrder)
	require.Equal(t, ty.ErrRelayVerifyAddrNotFound, err)

	// 订单金额超过 rawTx 真实输出金额
	fakeOrder = fixLegitOrder()
	fakeOrder.XAmount = 1000 * 1e8
	err = btc.verifyBtcTx(chainTestCfg, 10, &ty.RelayVerify{Tx: fixLegitBtcTx(), Spv: fixRealSpv()}, fakeOrder)
	require.Equal(t, ty.ErrRelayVerifyAddrNotFound, err)
}

// TestFixRelayVerifyBtcTxFakeHeight 伪造区块高度在分叉后被拒绝，确认数基于 SPV 证明的区块高度计算
func TestFixRelayVerifyBtcTxFakeHeight(t *testing.T) {
	kvdb := new(mocks.KVDB)
	btc := newBtcStore(kvdb)
	order := fixLegitOrder()

	// 自报 BlockHeight 与 SPV 证明的区块高度(100000)不一致
	tx := fixLegitBtcTx()
	tx.BlockHeight = 1
	kvdb.On("Get", mock.Anything).Return(fixEncHeight(100100), nil).Once()
	kvdb.On("Get", mock.Anything).Return(fixRealHeadEnc(), nil).Once()
	err := btc.verifyBtcTx(chainTestCfg, 10, &ty.RelayVerify{Tx: tx, Spv: fixRealSpv()}, order)
	require.Equal(t, ty.ErrRelayBtcTxHeightErr, err)

	// 自报 Spv.Height 与 SPV 证明的区块高度不一致
	tx = fixLegitBtcTx()
	spv := fixRealSpv()
	spv.Height = 99999
	kvdb.On("Get", mock.Anything).Return(fixEncHeight(100100), nil).Once()
	kvdb.On("Get", mock.Anything).Return(fixRealHeadEnc(), nil).Once()
	err = btc.verifyBtcTx(chainTestCfg, 10, &ty.RelayVerify{Tx: tx, Spv: spv}, order)
	require.Equal(t, ty.ErrRelayBtcTxHeightErr, err)

	// 确认数不足：基于 SPV 证明的区块高度 100000 计算，100000+100 > 100050
	waitOrder := fixLegitOrder()
	waitOrder.XBlockWaits = 100
	kvdb.On("Get", mock.Anything).Return(fixEncHeight(100050), nil).Once()
	kvdb.On("Get", mock.Anything).Return(fixRealHeadEnc(), nil).Once()
	err = btc.verifyBtcTx(chainTestCfg, 10, &ty.RelayVerify{Tx: fixLegitBtcTx(), Spv: fixRealSpv()}, waitOrder)
	require.Equal(t, ty.ErrRelayWaitBlocksErr, err)

	// 确认数足够时正常通过
	kvdb.On("Get", mock.Anything).Return(fixEncHeight(100100), nil).Once()
	kvdb.On("Get", mock.Anything).Return(fixRealHeadEnc(), nil).Once()
	err = btc.verifyBtcTx(chainTestCfg, 10, &ty.RelayVerify{Tx: fixLegitBtcTx(), Spv: fixRealSpv()}, waitOrder)
	require.NoError(t, err)
}

// TestFixRelayVerifyBtcTxPreFork 分叉前保持原行为：不校验 rawTx 与区块高度绑定关系
func TestFixRelayVerifyBtcTxPreFork(t *testing.T) {
	kvdb := new(mocks.KVDB)
	btc := newBtcStore(kvdb)
	preForkCfg := fixPreForkCfg()
	require.True(t, chainTestCfg.IsDappFork(10, ty.RelayX, ty.ForkRelayVerifyBtcTx))
	require.False(t, preForkCfg.IsDappFork(10, ty.RelayX, ty.ForkRelayVerifyBtcTx))
	require.True(t, preForkCfg.IsDappFork(1000000, ty.RelayX, ty.ForkRelayVerifyBtcTx))

	// 分叉前不根据 rawTx 重算哈希，也不校验 BlockHeight 与 SPV 证明区块的一致性，
	// 仅按自报内容做原有校验(保持旧共识行为)
	tx := &ty.BtcTransaction{
		Vout:        []*ty.Vout{{Address: addrBtc, Value: 0.299 * 1e8}},
		Time:        2500,
		BlockHeight: 1,
		Hash:        relayRealBtcTxHash,
	}
	kvdb.On("Get", mock.Anything).Return(fixEncHeight(100100), nil).Once()
	kvdb.On("Get", mock.Anything).Return(fixRealHeadEnc(), nil).Once()
	err := btc.verifyBtcTx(preForkCfg, 10, &ty.RelayVerify{Tx: tx, Spv: fixRealSpv()}, fixLegitOrder())
	require.NoError(t, err)

	// 同样的伪造内容在分叉后被拒绝
	tx.RawTx = relayRealBtcRawTx
	kvdb.On("Get", mock.Anything).Return(fixEncHeight(100100), nil).Once()
	kvdb.On("Get", mock.Anything).Return(fixRealHeadEnc(), nil).Once()
	err = btc.verifyBtcTx(chainTestCfg, 10, &ty.RelayVerify{Tx: tx, Spv: fixRealSpv()}, fixLegitOrder())
	require.Equal(t, ty.ErrRelayBtcTxHeightErr, err)
}
