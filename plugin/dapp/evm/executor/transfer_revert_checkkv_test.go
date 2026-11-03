// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// Framework-level repro of the reverted-transfer attack. The state-level tests use goleveldb,
// which has Delete, so the revert can physically remove the key. Production uses
// executor.StateDB, which only has Set: a key that did not exist before the transfer is
// shadowed with Set(key, nil), and that call still appends the key to GetSetKeys. The receipt
// is assembled from snapshots that no longer hold the reverted transfer, so the key is missing
// from it and chain33's checkKV rejects the whole transaction. execTxOne is unexported, so the
// test reproduces its check against the same StateDB type it inspects.

import (
	"testing"

	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/db"
	cexecutor "github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/system/address/btc"
	ctypes "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	vcomm "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

// newFrameworkStateDB returns the production StateDB backed by a queue that answers every
// store lookup with ErrNotFound, so reads of absent keys do not block on a store module.
func newFrameworkStateDB(t *testing.T, cfg *ctypes.Chain33Config, localDB db.KVDB) *cexecutor.StateDB {
	t.Helper()
	q := queue.New("channel")
	q.SetConfig(cfg)

	cli := q.Client()
	cli.Sub("store")
	done := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-cli.Recv():
				if msg == nil {
					return
				}
				msg.Reply(cli.NewMessage("", ctypes.EventStoreGetReply, &ctypes.StoreReplyValue{Values: [][]byte{nil}}))
			case <-done:
				return
			}
		}
	}()
	t.Cleanup(func() {
		close(done)
		q.Close()
	})

	return cexecutor.NewStateDB(cli, nil, localDB, nil).(*cexecutor.StateDB)
}

// checkReceiptCoversSetKeys mirrors chain33 executor's checkKV: every key recorded by
// StateDB.Set must also be carried by the receipt.
func checkReceiptCoversSetKeys(memset []string, kvs []*ctypes.KeyValue) error {
	keys := make(map[string]struct{}, len(kvs))
	for _, kv := range kvs {
		keys[string(kv.GetKey())] = struct{}{}
	}
	for _, key := range memset {
		if _, ok := keys[key]; !ok {
			return ctypes.ErrNotAllowMemSetKey
		}
	}
	return nil
}

func TestRevertedTransferTxRejectedByCheckKV(t *testing.T) {
	const (
		topUp = uint64(1000)
		moved = uint64(400)
	)

	// query_test.go replaces the package-wide EVM address driver with the ETH one; the executor
	// globals were initialized with the BTC driver, so restore it for the tx envelope.
	driver, err := address.LoadDriver(btc.NormalAddressID, -1)
	require.NoError(t, err)
	vcomm.InitEvmAddressDriver(driver)

	cfg := newTransferRevertConfig(t, 0)
	// checkKV only sees Set keys when the framework begins a per-tx transaction, which it does
	// under ForkExecRollback.
	cfg.SetFork("ForkExecRollback", 0)
	require.True(t, cfg.IsFork(0, "ForkExecRollback"))

	dbDir, rawDB, localDB := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dbDir, rawDB) })

	st := newFrameworkStateDB(t, cfg, localDB)
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)
	if !evmInitOnce {
		Init(cfg.ExecName(evmtypes.ExecutorName), cfg, nil)
		evmInitOnce = true
	}
	exec := NewEVMExecutor()
	exec.SetAPI(api)
	exec.SetLocalDB(localDB)
	exec.SetStateDB(st)
	exec.SetEnv(0, 0, 0)
	exec.CheckInit()

	target := vcomm.BytesToAddress([]byte{0xd0})
	forwarder := deployCode(t, cfg, exec, buildForwarderCode(target, moved))
	caller := deployCode(t, cfg, exec, buildCallerCode(*vcomm.StringToAddress(forwarder), moved))
	fundRole(t, exec, cfg, roleLegitUser, int64(topUp)*2)

	// Model executor.begin(): StartTx plus the in-tx mode that records Set keys.
	st.Begin()
	receipt, err := exec.Exec(makeCallTx(cfg, roleLegitUser, caller, []byte{0x01}, topUp), 0)
	require.NoError(t, err)
	require.Equal(t, int32(ctypes.ExecOk), receipt.GetTy())

	forwarderKey := string(exec.GetMStateDB().CoinsAccount.AccountKey(forwarder))
	require.Contains(t, st.GetSetKeys(), forwarderKey, "the reverted transfer must leave its key behind")

	receiptKeys := make(map[string]struct{}, len(receipt.GetKV()))
	for _, kv := range receipt.GetKV() {
		receiptKeys[string(kv.GetKey())] = struct{}{}
	}
	require.NotContains(t, receiptKeys, forwarderKey, "the reverted transfer must not reach the receipt")

	require.Equal(t, ctypes.ErrNotAllowMemSetKey, checkReceiptCoversSetKeys(st.GetSetKeys(), receipt.GetKV()),
		"checkKV must reject the transaction instead of committing the phantom balance")
}
