// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package state

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	ctypes "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

const transferRevertTestHeight = int64(10)

func newTransferRevertMDB(t *testing.T) *MemoryStateDB {
	t.Helper()
	cfgStr := ctypes.GetDefaultCfgstring()
	if !strings.Contains(cfgStr, "[exec.sub.evm]") {
		cfgStr += "\n[exec.sub.evm]\nethMapFromExecutor=\"coins\"\nethMapFromSymbol=\"bty\"\n"
	}
	cfg := ctypes.NewChain33Config(cfgStr)
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)

	dbDir, stateDB, localDB := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dbDir, stateDB) })
	coinsAccount, err := account.NewAccountDB(cfg, "coins", cfg.GetCoinSymbol(), stateDB)
	require.NoError(t, err)

	mdb := NewMemoryStateDB(stateDB, localDB, coinsAccount, transferRevertTestHeight, api)
	mdb.evmPlatformAddr = address.ExecAddress(cfg.ExecName(evmtypes.ExecutorName))
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

// TestTransferRevertRestoresBalances: a reverted transfer leaves both balances untouched.
func TestTransferRevertRestoresBalances(t *testing.T) {
	const (
		sender    = "14KEKbY3kNFLfQEGJbNweV4whre7NpqzuB"
		recipient = "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
		initial   = int64(1000)
		moved     = uint64(400)
	)

	mdb := newTransferRevertMDB(t)
	fundCoinAccount(mdb, sender, initial)

	snapshot := mdb.Snapshot()
	require.True(t, mdb.Transfer(sender, recipient, moved))
	require.Equal(t, initial-int64(moved), coinBalance(mdb, sender))
	require.Equal(t, int64(moved), coinBalance(mdb, recipient))

	mdb.RevertToSnapshot(snapshot)
	require.Equal(t, initial, coinBalance(mdb, sender))
	require.Zero(t, coinBalance(mdb, recipient), "recipient key did not exist before the transfer and must be removed")
}

// TestSnapshotRevertSameSlotReverseOrder pins the reverse-order revert: two SSTOREs to the
// same slot followed by a revert must restore the pre-frame value, not the intermediate one.
func TestSnapshotRevertSameSlotReverseOrder(t *testing.T) {
	contract := common.BytesToAddress([]byte{0x42})
	slot := common.BytesToHash([]byte("slot"))
	pre := common.BytesToHash([]byte("pre"))
	first := common.BytesToHash([]byte("first"))
	second := common.BytesToHash([]byte("second"))

	mdb := newTransferRevertMDB(t)
	mdb.CreateAccount(contract.String(), "creator", "evm.test", "test")
	mdb.SetState(contract.String(), slot, pre)

	snapshot := mdb.Snapshot()
	mdb.SetState(contract.String(), slot, first)
	mdb.SetState(contract.String(), slot, second)
	require.Equal(t, second, mdb.GetState(contract.String(), slot))

	mdb.RevertToSnapshot(snapshot)
	require.Equal(t, pre, mdb.GetState(contract.String(), slot), "reverse revert restores the pre-frame value")
}
