// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"errors"
	"math/big"
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const blockedRuntimeAddr = "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

func newBlockedEVM(t *testing.T, blockedAddrs []string) *EVM {
	t.Helper()
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	// 名单自高度 0 起生效，覆盖下面 Context 的区块高度 1
	restore := cfg.SetBlockedAccountsForTest(0, blockedAddrs)
	t.Cleanup(restore)
	ctx := Context{BlockNumber: big.NewInt(1)}
	return NewEVM(ctx, &state.MemoryStateDB{}, Config{}, cfg)
}

func TestCheckBlockedAccount(t *testing.T) {
	blocked := common.BytesToAddress(common.FromHex(blockedRuntimeAddr))
	normal := common.BytesToAddress(common.FromHex("0x0000000000000000000000000000000000000001"))

	t.Run("hit caller", func(t *testing.T) {
		evm := newBlockedEVM(t, []string{blockedRuntimeAddr})
		err := checkBlockedAccount(evm, blocked, normal)
		require.Error(t, err)
		assert.True(t, errors.Is(err, types.ErrBlockedAccount))
	})

	t.Run("hit target", func(t *testing.T) {
		evm := newBlockedEVM(t, []string{blockedRuntimeAddr})
		err := checkBlockedAccount(evm, normal, blocked)
		require.Error(t, err)
		assert.True(t, errors.Is(err, types.ErrBlockedAccount))
	})

	t.Run("normal pass", func(t *testing.T) {
		evm := newBlockedEVM(t, []string{blockedRuntimeAddr})
		assert.NoError(t, checkBlockedAccount(evm, normal, normal))
	})

	t.Run("empty blocklist pass", func(t *testing.T) {
		cfg := types.NewChain33Config(types.GetDefaultCfgstring())
		defer cfg.SetBlockedAccountsForTest(0, []string{})()
		evm := NewEVM(Context{BlockNumber: big.NewInt(1)}, &state.MemoryStateDB{}, Config{}, cfg)
		assert.NoError(t, checkBlockedAccount(evm, blocked, blocked))
	})

	// 名单自高度 H 起生效时，H 之前的区块必须放行，保证历史回放结果不变
	t.Run("before fork height pass", func(t *testing.T) {
		cfg := types.NewChain33Config(types.GetDefaultCfgstring())
		defer cfg.SetBlockedAccountsForTest(100, []string{blockedRuntimeAddr})()
		evm := NewEVM(Context{BlockNumber: big.NewInt(99)}, &state.MemoryStateDB{}, Config{}, cfg)
		assert.NoError(t, checkBlockedAccount(evm, blocked, blocked))

		evm = NewEVM(Context{BlockNumber: big.NewInt(100)}, &state.MemoryStateDB{}, Config{}, cfg)
		assert.Error(t, checkBlockedAccount(evm, blocked, blocked))
	})
}

// TestCallBlockedAccount 验证 EVM.Call 在黑名单地址下返回 error（触发上层 revert）
func TestCallBlockedAccount(t *testing.T) {
	evm := newBlockedEVM(t, []string{blockedRuntimeAddr})
	caller := AccountRef(common.BytesToAddress(common.FromHex(blockedRuntimeAddr)))
	target := common.BytesToAddress(common.FromHex("0x0000000000000000000000000000000000000001"))

	_, _, _, err := evm.Call(caller, target, nil, 100000, 0)
	require.Error(t, err)
	assert.True(t, errors.Is(err, types.ErrBlockedAccount))
}

// TestCreateBlockedAccount 验证 EVM.Create 在黑名单地址下返回 error（触发上层 revert）。
// 与 TestCallBlockedAccount 对称，覆盖 Create 路径的 checkBlockedAccount 分支。
func TestCreateBlockedAccount(t *testing.T) {
	evm := newBlockedEVM(t, []string{blockedRuntimeAddr})
	caller := AccountRef(common.BytesToAddress(common.FromHex(blockedRuntimeAddr)))
	contractAddr := common.BytesToAddress(common.FromHex("0x0000000000000000000000000000000000000002"))

	_, _, _, err := evm.Create(caller, contractAddr, nil, 100000, "evm", "test", 0)
	require.Error(t, err)
	assert.True(t, errors.Is(err, types.ErrBlockedAccount))
}
