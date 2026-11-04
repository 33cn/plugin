// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// A failed nested CREATE used to take its attached value from the caller and keep it at the
// child address whose account was rolled back: Create transferred the value before taking the
// snapshot. The snapshot is now taken first, so the value returns to the caller.

import (
	"bytes"
	"testing"

	ctypes "github.com/33cn/chain33/types"
	"github.com/stretchr/testify/require"
)

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

	env := newGapEnv(t, nil)
	caller, parent := addr(0x31), addr(0x51)
	env.fund(parent, stash)
	env.deploy(parent, caller, buildCreateWithValueCode(attached, []byte{0xfe}))

	_, _, _, err := env.evm.Call(AccountRef(caller), parent, nil, 5_000_000, 0)
	require.NoError(t, err)
	require.Equal(t, stash, env.balance(parent))
}

// TestSuccessfulCreateWithValueCommitsBalances covers a successful top-level CREATE carrying a
// value. With no snapshot open yet, the old code dropped the transfer in addChange: the balances
// moved but the receipt never carried the keys, so chain33's checkKV rejected the transaction.
// Both keys must now land in the receipt.
func TestSuccessfulCreateWithValueCommitsBalances(t *testing.T) {
	const stash = int64(1000)
	const attached = uint64(5)

	env := newGapEnv(t, nil)
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
		require.True(t, hasKV(kvs, key), "receipt key %x", key)
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
