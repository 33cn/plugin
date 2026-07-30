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
	// local 标题下 SetAllFork(0)，ForkAccountBlacklist 从高度 0 启用
	restore := types.SetBlockedAccountsForTest(blockedAddrs)
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
		restore := types.SetBlockedAccountsForTest([]string{})
		defer restore()
		evm := NewEVM(Context{BlockNumber: big.NewInt(1)}, &state.MemoryStateDB{}, Config{}, cfg)
		assert.NoError(t, checkBlockedAccount(evm, blocked, blocked))
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
