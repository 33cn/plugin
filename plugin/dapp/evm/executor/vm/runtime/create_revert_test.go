// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// A failed nested CREATE used to take its attached value from the caller and keep it at the
// child address whose account was rolled back: Create transferred the value before taking the
// snapshot. From ForkEVMTransferRevert on the snapshot is taken first, so the value returns to
// the caller.

import (
	"bytes"
	"strings"
	"testing"

	ctypes "github.com/33cn/chain33/types"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

func newCreateRevertEnv(t *testing.T, forkHeight int64) *gapEnv {
	t.Helper()
	cfgStr := strings.Replace(ctypes.GetDefaultCfgstring(), `Title="local"`, `Title="test"`, 1)
	if !strings.Contains(cfgStr, "[exec.sub.evm]") {
		cfgStr += "\n[exec.sub.evm]\nethMapFromExecutor=\"coins\"\nethMapFromSymbol=\"bty\"\n"
	}
	cfg := ctypes.NewChain33Config(cfgStr)
	require.False(t, cfg.IsLocal(), "a local config would force every fork to height 0")
	cfg.SetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMTransferRevert, forkHeight)
	require.Equal(t, forkHeight, cfg.GetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMTransferRevert))
	require.Equal(t, forkHeight <= gapTestHeight,
		cfg.IsDappFork(gapTestHeight, evmtypes.ExecutorName, evmtypes.ForkEVMTransferRevert))
	return newGapEnvWithCfg(t, cfg, "create-revert")
}

// buildCreateWithValueCode emits CREATE(value, 0, len(init)) with init appended to the code,
// where init is copied into memory with CODECOPY. init is expected to abort, so the create fails.
func buildCreateWithValueCode(value uint64, init []byte) []byte {
	code := []byte{0x60, byte(len(init))} // PUSH1 len(init)
	offsetAt := len(code)
	code = append(code, 0x60, 0x00)            // PUSH1 codeOffset (patched below)
	code = append(code, 0x60, 0x00)            // PUSH1 0 (mem offset)
	code = append(code, 0x39)                  // CODECOPY
	code = append(code, 0x60, byte(len(init))) // PUSH1 size
	code = append(code, 0x60, 0x00)            // PUSH1 0 (mem offset)
	code = append(code, pushUint64(value)...)  // PUSH value
	code = append(code, 0xf0)                  // CREATE
	code = append(code, 0x00)                  // STOP
	code[offsetAt+1] = byte(len(code))
	return append(code, init...)
}

func TestFailedNestedCreateReturnsValue(t *testing.T) {
	const stash = int64(1000)
	const attached = uint64(5)

	for _, tc := range []struct {
		name       string
		forkHeight int64
		wantLeft   int64
	}{
		{"pre-fork keeps the old behaviour", ctypes.MaxHeight, stash - int64(attached)},
		{"post-fork returns the value", 0, stash},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := newCreateRevertEnv(t, tc.forkHeight)
			caller, parent := addr(0x31), addr(0x51)
			env.fund(parent, stash)
			env.deploy(parent, caller, buildCreateWithValueCode(attached, []byte{0xfe}))

			_, _, _, err := env.evm.Call(AccountRef(caller), parent, nil, 5_000_000, 0)
			require.NoError(t, err)
			require.Equal(t, tc.wantLeft, env.balance(parent))
		})
	}
}

// TestSuccessfulCreateWithValueCommitsBalances covers a successful top-level CREATE carrying a
// value. With no snapshot open yet, the pre-fork transfer is dropped by addChange: the balances
// move but the receipt never carries the keys, so chain33's checkKV rejects the transaction (or,
// without ForkExecRollback, commits the value without a receipt entry). From the fork on the
// snapshot opens before the transfer and both keys land in the receipt.
func TestSuccessfulCreateWithValueCommitsBalances(t *testing.T) {
	const stash = int64(1000)
	const attached = uint64(5)

	for _, tc := range []struct {
		name          string
		forkHeight    int64
		wantInReceipt bool
	}{
		{"pre-fork drops the balance keys", ctypes.MaxHeight, false},
		{"post-fork commits the balance keys", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := newCreateRevertEnv(t, tc.forkHeight)
			caller, contract := addr(0x61), addr(0x71)
			env.fund(caller, stash)

			_, snapshot, _, err := env.evm.Create(AccountRef(caller), contract, []byte{0x00}, 5_000_000, "evm.test", "test", attached)
			require.NoError(t, err)
			require.Equal(t, stash-int64(attached), env.balance(caller))
			require.Equal(t, int64(attached), env.balance(contract))

			kvs, _ := env.mdb.GetChangedData(snapshot)
			for _, key := range [][]byte{
				env.coins.AccountKey(caller.String()),
				env.coins.AccountKey(contract.String()),
			} {
				require.Equal(t, tc.wantInReceipt, hasKV(kvs, key), "receipt key %x", key)
			}
		})
	}
}

func hasKV(kvs []*ctypes.KeyValue, key []byte) bool {
	for _, kv := range kvs {
		if bytes.Equal(kv.Key, key) {
			return true
		}
	}
	return false
}
