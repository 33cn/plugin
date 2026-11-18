// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// Framework-level tests for reverted value transfers on the production StateDB shape:
// a concrete executor.StateDB with no Delete, ForkExecRollback open (Begin records every Set
// key in GetSetKeys) and the receipt/checkKV path that decides whether a transaction is
// accepted. The two shapes are a pair: a legitimate transaction that survives a reverted call
// must commit, and the mint shape must not create value.

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

// newProductionExec builds an EVM executor over the production StateDB with the transfer-revert
// fork and ForkExecRollback both open at height 0.
func newProductionExec(t *testing.T, cfg *ctypes.Chain33Config, st *cexecutor.StateDB, localDB db.KVDB) *EVMExecutor {
	t.Helper()
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
	return exec
}

func restoreBtcAddressDriver(t *testing.T) {
	t.Helper()
	// query_test.go replaces the package-wide EVM address driver with the ETH one; the executor
	// globals were initialized with the BTC driver, so restore it for the tx envelope.
	driver, err := address.LoadDriver(btc.NormalAddressID, -1)
	require.NoError(t, err)
	vcomm.InitEvmAddressDriver(driver)
}

// buildForwardAndStoreCode calls callee with value (the callee aborts) and then writes slot 0,
// so the outer frame continues and changes state after the inner value transfer is reverted.
func buildForwardAndStoreCode(callee vcomm.Address, value uint64) []byte {
	code := buildCall(callee, value, 0)
	code = append(code, 0x60, 0x01) // PUSH1 1
	code = append(code, 0x60, 0x00) // PUSH1 0
	code = append(code, 0x55)       // SSTORE
	code = append(code, 0x00)       // STOP
	return code
}

func productionStateDBFixture(t *testing.T) (*ctypes.Chain33Config, *cexecutor.StateDB, db.KVDB) {
	t.Helper()
	restoreBtcAddressDriver(t)
	cfg := newTransferRevertConfig(t, 0)
	// checkKV only sees Set keys when the framework begins a per-tx transaction, which it does
	// under ForkExecRollback.
	cfg.SetFork("ForkExecRollback", 0)
	require.True(t, cfg.IsFork(0, "ForkExecRollback"))

	dbDir, rawDB, localDB := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dbDir, rawDB) })
	return cfg, newFrameworkStateDB(t, cfg, localDB), localDB
}

// TestRevertedCallValueTxSucceedsUnderCheckKV is the reviewer's positive case: the outer call
// succeeds, an inner CALL{value} reverts, execution continues and writes state. Post-fork the
// transaction must be ExecOk, pass checkKV, and commit exactly the post-revert balances.
func TestRevertedCallValueTxSucceedsUnderCheckKV(t *testing.T) {
	const (
		topUp = uint64(1000)
		moved = uint64(400)
	)

	cfg, st, localDB := productionStateDBFixture(t)
	exec := newProductionExec(t, cfg, st, localDB)

	target := vcomm.BytesToAddress([]byte{0xd0})
	forwarder := deployCode(t, cfg, exec, buildForwarderCode(target, moved))
	caller := deployCode(t, cfg, exec, buildForwardAndStoreCode(*vcomm.StringToAddress(forwarder), moved))
	fundRole(t, exec, cfg, roleLegitUser, int64(topUp)*2)
	user := addrFromRole(cfg, roleLegitUser)
	initial := balanceOf(exec, user)

	// Model executor.begin(): StartTx plus the in-tx mode that records Set keys.
	st.Begin()
	receipt, err := exec.Exec(makeCallTx(cfg, roleLegitUser, caller, []byte{0x01}, topUp), 0)
	require.NoError(t, err)
	require.Equal(t, int32(ctypes.ExecOk), receipt.GetTy())
	require.NoError(t, checkReceiptCoversSetKeys(st.GetSetKeys(), receipt.GetKV()),
		"a reverted inner value transfer must not leave keys the receipt lacks")

	applyReceiptKV(t, exec, receipt)
	require.Equal(t, initial-int64(topUp), balanceOf(exec, user))
	require.Equal(t, int64(topUp), balanceOf(exec, caller), "the outer transfer survives")
	require.Zero(t, balanceOf(exec, forwarder), "the reverted transfer must not credit the callee")
	require.Zero(t, balanceOf(exec, target.String()))
	require.Equal(t, vcomm.BytesToHash([]byte{0x01}), exec.GetMStateDB().GetState(caller, vcomm.BytesToHash([]byte{0x00})),
		"execution must continue past the reverted call and commit its state write")
}

// TestRevertedCallValueMintTxCreatesNoValue keeps the attack shape under the same production
// StateDB: the phantom credit from a reverted transfer is spent in a later call. The transaction
// now succeeds (it no longer trips checkKV) but must not mint the reverted amount.
func TestRevertedCallValueMintTxCreatesNoValue(t *testing.T) {
	const (
		topUp = uint64(1000)
		moved = uint64(400)
	)

	cfg, st, localDB := productionStateDBFixture(t)
	exec := newProductionExec(t, cfg, st, localDB)

	target := vcomm.BytesToAddress([]byte{0xd0})
	forwarder := deployCode(t, cfg, exec, buildForwarderCode(target, moved))
	caller := deployCode(t, cfg, exec, buildCallerCode(*vcomm.StringToAddress(forwarder), moved))
	fundRole(t, exec, cfg, roleLegitUser, int64(topUp)*2)
	user := addrFromRole(cfg, roleLegitUser)
	initial := balanceOf(exec, user)

	st.Begin()
	receipt, err := exec.Exec(makeCallTx(cfg, roleLegitUser, caller, []byte{0x01}, topUp), 0)
	require.NoError(t, err)
	require.Equal(t, int32(ctypes.ExecOk), receipt.GetTy())
	require.NoError(t, checkReceiptCoversSetKeys(st.GetSetKeys(), receipt.GetKV()))

	applyReceiptKV(t, exec, receipt)
	require.Equal(t, initial-int64(topUp), balanceOf(exec, user))
	require.Equal(t, int64(topUp), balanceOf(exec, caller))
	require.Zero(t, balanceOf(exec, forwarder), "the reverted credit must not be spendable")
	require.Zero(t, balanceOf(exec, target.String()), "the mint must not reach the target")
}
