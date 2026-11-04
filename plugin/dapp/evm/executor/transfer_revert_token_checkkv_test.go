// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// TransferToToken goes through the same transferChange revert path as Transfer, so it gets the
// same pair of checks on the production StateDB: a legitimate token transfer that survives a
// reverted one still passes checkKV, and the credit of a reverted transfer cannot be spent.

import (
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	cexecutor "github.com/33cn/chain33/executor"
	ctypes "github.com/33cn/chain33/types"
	evmstate "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/state"
	tokenty "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/stretchr/testify/require"
)

const (
	tokenRevertFrom   = "14KEKbY3kNFLfQEGJbNweV4whre7NpqzuB"
	tokenRevertMiddle = "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
	tokenRevertSink   = "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
)

type tokenRevertFixture struct {
	st      *cexecutor.StateDB
	mdb     *evmstate.MemoryStateDB
	tokenDB *account.DB
}

// newTokenRevertFixture builds a MemoryStateDB over the production StateDB, a created token in
// LocalDB and one funded token account. Funding happens before Begin, as in a real block.
func newTokenRevertFixture(t *testing.T) *tokenRevertFixture {
	t.Helper()
	cfg, st, localDB := productionStateDBFixture(t)
	require.NoError(t, localDB.Set([]byte("LODB-token-create-sto-1-BTY-0"),
		ctypes.Encode(&tokenty.LocalToken{Symbol: "BTY", Status: 1})))

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)
	mdb := evmstate.NewMemoryStateDB(st, localDB, nil, 0, api)

	tokenDB, err := account.NewAccountDB(cfg, "token", "BTY", st)
	require.NoError(t, err)
	acc := tokenDB.LoadAccount(tokenRevertFrom)
	acc.Balance = 1000
	tokenDB.SaveAccount(acc)

	return &tokenRevertFixture{st: st, mdb: mdb, tokenDB: tokenDB}
}

func (f *tokenRevertFixture) balance(t *testing.T, addr string) int64 {
	t.Helper()
	return f.tokenDB.LoadAccount(addr).Balance
}

// applyKV commits the receipt KV into the state, completing the checkKV path the way the
// framework does before the transaction is accepted.
func (f *tokenRevertFixture) applyKV(t *testing.T, kvs []*ctypes.KeyValue) {
	t.Helper()
	for _, kv := range kvs {
		require.NoError(t, f.st.Set(kv.GetKey(), kv.GetValue()))
	}
}

func TestTransferToTokenRevertCheckKVPair(t *testing.T) {
	t.Run("positive: transfers after a revert pass checkKV", func(t *testing.T) {
		f := newTokenRevertFixture(t)
		f.st.Begin()
		outer := f.mdb.Snapshot()

		ok, err := f.mdb.TransferToToken(tokenRevertFrom, tokenRevertSink, "BTY", 100)
		require.NoError(t, err)
		require.True(t, ok)

		inner := f.mdb.Snapshot()
		ok, err = f.mdb.TransferToToken(tokenRevertFrom, tokenRevertMiddle, "BTY", 400)
		require.NoError(t, err)
		require.True(t, ok)
		f.mdb.RevertToSnapshot(inner)

		// The outer frame keeps executing after the reverted inner transfer.
		ok, err = f.mdb.TransferToToken(tokenRevertFrom, tokenRevertSink, "BTY", 50)
		require.NoError(t, err)
		require.True(t, ok)

		kvs, _ := f.mdb.GetChangedData(outer)
		require.NoError(t, checkReceiptCoversSetKeys(f.st.GetSetKeys(), kvs),
			"the reverted token transfer must not leave keys the receipt lacks")
		f.applyKV(t, kvs)

		require.Equal(t, int64(850), f.balance(t, tokenRevertFrom))
		require.Zero(t, f.balance(t, tokenRevertMiddle), "the reverted transfer must not credit the middle account")
		require.Equal(t, int64(150), f.balance(t, tokenRevertSink))
	})

	t.Run("negative: the reverted credit cannot be spent", func(t *testing.T) {
		f := newTokenRevertFixture(t)
		f.st.Begin()
		outer := f.mdb.Snapshot()

		ok, err := f.mdb.TransferToToken(tokenRevertFrom, tokenRevertSink, "BTY", 100)
		require.NoError(t, err)
		require.True(t, ok)

		inner := f.mdb.Snapshot()
		ok, err = f.mdb.TransferToToken(tokenRevertFrom, tokenRevertMiddle, "BTY", 400)
		require.NoError(t, err)
		require.True(t, ok)
		f.mdb.RevertToSnapshot(inner)

		ok, err = f.mdb.TransferToToken(tokenRevertMiddle, tokenRevertSink, "BTY", 400)
		require.False(t, ok, "the phantom credit must not be spendable")
		require.ErrorIs(t, err, ctypes.ErrNoBalance)

		kvs, _ := f.mdb.GetChangedData(outer)
		require.NoError(t, checkReceiptCoversSetKeys(f.st.GetSetKeys(), kvs))
		f.applyKV(t, kvs)

		require.Equal(t, int64(900), f.balance(t, tokenRevertFrom))
		require.Zero(t, f.balance(t, tokenRevertMiddle))
		require.Equal(t, int64(100), f.balance(t, tokenRevertSink))
	})
}
