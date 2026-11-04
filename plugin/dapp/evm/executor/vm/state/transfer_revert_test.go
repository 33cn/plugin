// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/db"
	cexecutor "github.com/33cn/chain33/executor"
	ctypes "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

const transferRevertTestHeight = int64(10)

// newTransferRevertConfig builds an EVM config whose ForkEVMTransferRevert resolves to
// forkHeight. The config must not be local: Title="local" pins every dapp fork to 0 via
// needSetForkZero, which would make the pre-fork side unreachable.
func newTransferRevertConfig(t *testing.T, forkHeight int64) *ctypes.Chain33Config {
	t.Helper()
	cfgStr := strings.Replace(ctypes.GetDefaultCfgstring(), `Title="local"`, `Title="evm-transfer-revert-test"`, 1)
	if !strings.Contains(cfgStr, "[exec.sub.evm]") {
		cfgStr += "\n[exec.sub.evm]\nethMapFromExecutor=\"coins\"\nethMapFromSymbol=\"bty\"\n"
	}
	cfg := ctypes.NewChain33Config(cfgStr)
	require.False(t, cfg.IsLocal())
	cfg.SetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMTransferRevert, forkHeight)
	require.Equal(t, forkHeight, cfg.GetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMTransferRevert))
	return cfg
}

// newTransferRevertMDBWith builds a MemoryStateDB over kv, so tests can pick a KV without
// Delete to reach transferChange.revert's fallback branch.
func newTransferRevertMDBWith(t *testing.T, cfg *ctypes.Chain33Config, kv db.KV) *MemoryStateDB {
	t.Helper()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)
	coinsAccount, err := account.NewAccountDB(cfg, "coins", cfg.GetCoinSymbol(), kv)
	require.NoError(t, err)

	mdb := NewMemoryStateDB(kv, nil, coinsAccount, transferRevertTestHeight, api)
	mdb.evmPlatformAddr = address.ExecAddress(cfg.ExecName(evmtypes.ExecutorName))
	return mdb
}

func newTransferRevertMDB(t *testing.T, forkHeight int64) *MemoryStateDB {
	t.Helper()
	cfg := newTransferRevertConfig(t, forkHeight)
	dbDir, stateDB, localDB := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dbDir, stateDB) })

	mdb := newTransferRevertMDBWith(t, cfg, stateDB)
	mdb.LocalDB = localDB
	return mdb
}

func fundCoinAccount(mdb *MemoryStateDB, addr string, amount int64) {
	acc := mdb.CoinsAccount.LoadAccount(addr)
	acc.Balance = amount
	mdb.CoinsAccount.SaveAccount(acc)
}

func coinBalance(mdb *MemoryStateDB, addr string) int64 {
	return mdb.CoinsAccount.LoadAccount(addr).Balance
}

// TestTransferRevertRestoresBalances pins both sides of ForkEVMTransferRevert: after the fork a
// reverted transfer leaves both balances untouched, before it the cache keeps the transfer.
func TestTransferRevertRestoresBalances(t *testing.T) {
	const (
		sender    = "14KEKbY3kNFLfQEGJbNweV4whre7NpqzuB"
		recipient = "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
		initial   = int64(1000)
		moved     = uint64(400)
	)

	t.Run("pre-fork keeps the old behaviour", func(t *testing.T) {
		mdb := newTransferRevertMDB(t, ctypes.MaxHeight)
		fundCoinAccount(mdb, sender, initial)

		snapshot := mdb.Snapshot()
		require.True(t, mdb.Transfer(sender, recipient, moved))
		require.Equal(t, initial-int64(moved), coinBalance(mdb, sender))
		require.Equal(t, int64(moved), coinBalance(mdb, recipient))

		mdb.RevertToSnapshot(snapshot)
		require.Equal(t, initial-int64(moved), coinBalance(mdb, sender), "pre-fork revert must not restore the sender")
		require.Equal(t, int64(moved), coinBalance(mdb, recipient), "pre-fork revert must leave the phantom credit")
	})

	t.Run("post-fork restores both balances", func(t *testing.T) {
		mdb := newTransferRevertMDB(t, 0)
		fundCoinAccount(mdb, sender, initial)

		snapshot := mdb.Snapshot()
		require.True(t, mdb.Transfer(sender, recipient, moved))
		require.Equal(t, initial-int64(moved), coinBalance(mdb, sender))
		require.Equal(t, int64(moved), coinBalance(mdb, recipient))

		mdb.RevertToSnapshot(snapshot)
		require.Equal(t, initial, coinBalance(mdb, sender))
		require.Zero(t, coinBalance(mdb, recipient), "recipient key did not exist before the transfer and must be removed")
	})
}

// TestSnapshotRevertSameSlotReverseOrder pins the reverse-order revert: two SSTOREs to the
// same slot followed by a revert must restore the pre-frame value, not the intermediate one.
func TestSnapshotRevertSameSlotReverseOrder(t *testing.T) {
	contract := common.BytesToAddress([]byte{0x42})
	slot := common.BytesToHash([]byte("slot"))
	pre := common.BytesToHash([]byte("pre"))
	first := common.BytesToHash([]byte("first"))
	second := common.BytesToHash([]byte("second"))

	t.Run("pre-fork restores the intermediate value", func(t *testing.T) {
		mdb := newTransferRevertMDB(t, ctypes.MaxHeight)
		mdb.CreateAccount(contract.String(), "creator", "evm.test", "test")
		mdb.SetState(contract.String(), slot, pre)

		snapshot := mdb.Snapshot()
		mdb.SetState(contract.String(), slot, first)
		mdb.SetState(contract.String(), slot, second)
		require.Equal(t, second, mdb.GetState(contract.String(), slot))

		mdb.RevertToSnapshot(snapshot)
		require.Equal(t, first, mdb.GetState(contract.String(), slot), "forward revert restores the intermediate write")
	})

	t.Run("post-fork restores the pre-frame value", func(t *testing.T) {
		mdb := newTransferRevertMDB(t, 0)
		mdb.CreateAccount(contract.String(), "creator", "evm.test", "test")
		mdb.SetState(contract.String(), slot, pre)

		snapshot := mdb.Snapshot()
		mdb.SetState(contract.String(), slot, first)
		mdb.SetState(contract.String(), slot, second)
		require.Equal(t, second, mdb.GetState(contract.String(), slot))

		mdb.RevertToSnapshot(snapshot)
		require.Equal(t, pre, mdb.GetState(contract.String(), slot), "reverse revert restores the pre-frame value")
	})
}

// TestTransferRevertWithoutDeleteHidesKey covers transferChange.revert's fallback when the KV
// has no transaction open: executor.StateDB exposes RevertSet, but outside Begin there is no
// key list to repair, so revert writes directly. A key that did not exist before the transfer
// is then shadowed with a nil value, which Get reports as absent. The goleveldb used by the
// other tests has Delete and never reaches this branch.
func TestTransferRevertWithoutDeleteHidesKey(t *testing.T) {
	const (
		sender    = "14KEKbY3kNFLfQEGJbNweV4whre7NpqzuB"
		recipient = "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
		initial   = int64(1000)
		moved     = uint64(400)
	)

	st := cexecutor.NewStateDB(nil, nil, nil, nil).(*cexecutor.StateDB)
	_, hasDelete := interface{}(st).(interface{ Delete([]byte) error })
	require.False(t, hasDelete, "executor.StateDB must not expose Delete or the fallback stays untested")
	_, hasRevertSet := interface{}(st).(interface{ RevertSet([]byte, []byte) error })
	require.True(t, hasRevertSet, "executor.StateDB should undo writes through RevertSet inside a tx")

	mdb := newTransferRevertMDBWith(t, newTransferRevertConfig(t, 0), st)
	fundCoinAccount(mdb, sender, initial)
	recipientKey := mdb.CoinsAccount.AccountKey(recipient)

	snapshot := mdb.Snapshot()
	require.True(t, mdb.Transfer(sender, recipient, moved))
	_, err := mdb.StateDB.Get(recipientKey)
	require.NoError(t, err, "the transfer must have written the recipient key")

	mdb.RevertToSnapshot(snapshot)
	_, err = mdb.StateDB.Get(recipientKey)
	require.Equal(t, ctypes.ErrNotFound, err, "the nil Set must shadow the key as absent")
	require.Equal(t, initial, coinBalance(mdb, sender))
}

// faultyKV wraps a KV and fails the read of one key, standing in for a store timeout.
type faultyKV struct {
	db.KV
	failKey []byte
	err     error
}

func (f faultyKV) Get(key []byte) ([]byte, error) {
	if bytes.Equal(key, f.failKey) {
		return nil, f.err
	}
	return f.KV.Get(key)
}

// revertFailingKV is a KV whose RevertSet always fails, so the error path of the producer can
// be observed without a real storage fault.
type revertFailingKV struct {
	db.KV
	err error
}

func (f revertFailingKV) RevertSet(key, originalValue []byte) error {
	return f.err
}

// TestRevertErrorIsRecorded pins the propagation of a failed RevertSet: the error reaches the
// MemoryStateDB instead of being swallowed, and the next execution can clear it.
func TestRevertErrorIsRecorded(t *testing.T) {
	const (
		sender    = "14KEKbY3kNFLfQEGJbNweV4whre7NpqzuB"
		recipient = "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
		initial   = int64(1000)
	)

	dbDir, stateDB, _ := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dbDir, stateDB) })
	revertErr := errors.New("revert failed")
	mdb := newTransferRevertMDBWith(t, newTransferRevertConfig(t, 0), revertFailingKV{KV: stateDB, err: revertErr})
	fundCoinAccount(mdb, sender, initial)

	snapshot := mdb.Snapshot()
	require.True(t, mdb.Transfer(sender, recipient, 400))
	mdb.RevertToSnapshot(snapshot)
	require.ErrorIs(t, mdb.RevertError(), revertErr, "the producer error must reach the caller")

	mdb.ResetRevertError()
	require.NoError(t, mdb.RevertError())
}

// TestCapturePreStateFailsTransferOnReadError pins the reviewer's Suggestion: only ErrNotFound
// may mean "the key did not exist". Any other read error leaves the pre-transfer value unknown,
// so the transfer must fail instead of recording a state that was never read.
func TestCapturePreStateFailsTransferOnReadError(t *testing.T) {
	const (
		sender    = "14KEKbY3kNFLfQEGJbNweV4whre7NpqzuB"
		recipient = "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
		initial   = int64(1000)
		moved     = uint64(400)
	)

	cfg := newTransferRevertConfig(t, 0)
	dbDir, stateDB, _ := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dbDir, stateDB) })

	// The key is built from the same config the coins DB uses.
	coinsAccount, err := account.NewAccountDB(cfg, "coins", cfg.GetCoinSymbol(), stateDB)
	require.NoError(t, err)
	storeErr := errors.New("store read timeout")
	mdb := newTransferRevertMDBWith(t, cfg, faultyKV{
		KV:      stateDB,
		failKey: coinsAccount.AccountKey(recipient),
		err:     storeErr,
	})
	fundCoinAccount(mdb, sender, initial)

	snapshot := mdb.Snapshot()
	require.False(t, mdb.Transfer(sender, recipient, moved), "a pre-state read error must fail the transfer")
	mdb.RevertToSnapshot(snapshot)
	require.Equal(t, initial, coinBalance(mdb, sender))
	require.Zero(t, coinBalance(mdb, recipient))
}
