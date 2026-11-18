// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// Executor-level test for the guard in innerExec that turns a failed snapshot revert into a
// failed transaction. A failed revert leaves the account cache and the receipt describing
// different state, so the transaction must not be committed.

import (
	"errors"
	"testing"

	"github.com/33cn/chain33/common/db"
	ctypes "github.com/33cn/chain33/types"
	vcomm "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/stretchr/testify/require"
)

// failingRevertKV wraps the production StateDB and fails every RevertSet, standing in for a
// store fault during the undo of a reverted value transfer.
type failingRevertKV struct {
	db.KV
	err error
}

func (f failingRevertKV) RevertSet(key, originalValue []byte) error {
	return f.err
}

// TestInnerExecFailsWhenSnapshotRevertFails drives the same CALL{value} revert shape as the
// transfer-revert integration tests, but with a KV whose RevertSet always fails. The inner call
// is popped by the caller contract, so the outer frame reports success; only the RevertError
// guard can fail the transaction. The next execution must still succeed: ResetRevertError clears
// the recorded error at the start of every innerExec call.
func TestInnerExecFailsWhenSnapshotRevertFails(t *testing.T) {
	const (
		topUp = uint64(1000)
		moved = uint64(400)
	)

	cfg, st, localDB := productionStateDBFixture(t)
	revertErr := errors.New("injected RevertSet failure")

	// Wrap the production StateDB so the snapshot revert of the value transfer fails. SetStateDB
	// also points the coins account at the wrapper, and CheckInit rebuilds the MemoryStateDB.
	exec := newProductionExec(t, cfg, st, localDB)
	exec.SetStateDB(failingRevertKV{KV: st, err: revertErr})
	exec.CheckInit()

	target := vcomm.BytesToAddress([]byte{0xd0})
	forwarder := deployCode(t, cfg, exec, buildForwarderCode(target, moved))
	caller := deployCode(t, cfg, exec, buildCallerCode(*vcomm.StringToAddress(forwarder), moved))
	noop := deployCode(t, cfg, exec, []byte{0x00})

	user := addrFromRole(cfg, roleLegitUser)
	fundRole(t, exec, cfg, roleLegitUser, int64(topUp)*2)
	initial := balanceOf(exec, user)

	// First transaction: the forwarder aborts on the value call, the reverted transfer fails to
	// restore its account keys, and the outer contract continues to success.
	st.Begin()
	receipt, err := exec.Exec(makeCallTx(cfg, roleLegitUser, caller, []byte{0x01}, topUp), 0)
	require.ErrorIs(t, err, revertErr, "innerExec must fail the transaction on a failed snapshot revert")
	require.Empty(t, receipt.GetKV(), "the failed transaction must carry no state to commit")
	require.ErrorIs(t, exec.GetMStateDB().RevertError(), revertErr,
		"the guard must observe the recorded revert error")
	st.Rollback()

	require.Equal(t, initial, balanceOf(exec, user), "no balance may move out of the user")
	require.Zero(t, balanceOf(exec, forwarder), "the reverted credit must not be committed")
	require.Zero(t, balanceOf(exec, target.String()), "the phantom credit must not be committed")

	// Second execution on the same MemoryStateDB (Exec would recreate it through CheckInit).
	// Nothing reverts here, so a stale error would fail at the guard; ResetRevertError at the
	// start of innerExec must clear it before execution starts.
	tx := makeCallTx(cfg, roleLegitUser, noop, nil, 0)
	msg, err := exec.GetMessage(tx, 0, nil)
	require.NoError(t, err)

	st.Begin()
	receipt, err = exec.innerExec(msg, tx.Hash(), tx.GetSignature().GetTy(), 0, msg.GasLimit(), false)
	require.NoError(t, err, "a failed revert must not poison the next execution")
	require.Equal(t, int32(ctypes.ExecOk), receipt.GetTy())
	require.NoError(t, exec.GetMStateDB().RevertError(), "ResetRevertError must clear the recorded error")
	st.Rollback()
}
