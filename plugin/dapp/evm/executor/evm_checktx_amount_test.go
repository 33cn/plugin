// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"math"
	"testing"

	ctypes "github.com/33cn/chain33/types"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

// makeCallTxWithAmount builds a signed evm call tx carrying the given amount.
func makeCallTxWithAmount(cfg *ctypes.Chain33Config, amount uint64) *ctypes.Transaction {
	action := &evmtypes.EVMContractAction{
		Amount: amount, GasLimit: 100000, GasPrice: 1,
		Code: nil, Para: []byte("test"),
		ContractAddr: "0x0000000000000000000000000000000000000000",
	}
	tx := &ctypes.Transaction{
		ChainID: cfg.GetChainID(),
		Execer:  []byte(cfg.ExecName(evmtypes.ExecutorName)),
		Payload: ctypes.Encode(action),
		Fee:     1e6, Nonce: 0,
	}
	signTx(tx, roleAttacker)
	return tx
}

// After ForkEVMFixOverflow, CheckTx must reject evm txs whose uint64 Amount
// exceeds math.MaxInt64 (the value that used to wrap to -1 and bypass the
// balance check, enabling the fake msg.value attack). Before the fork the
// historical behavior is preserved (only replay check runs).
func TestCheckTxRejectsAmountOverflow(t *testing.T) {
	cfg := newTestConfig(t) // ForkEVMFixOverflow = 1000

	// post-fork: overflow amount rejected at CheckTx
	exec := newTestExecutor(t, cfg, 1000)
	tx := makeCallTxWithAmount(cfg, math.MaxUint64)
	require.Equal(t, ctypes.ErrAmount, exec.CheckTx(tx, 0))

	tx = makeCallTxWithAmount(cfg, math.MaxInt64+1)
	require.Equal(t, ctypes.ErrAmount, exec.CheckTx(tx, 0))

	// post-fork: boundary and normal amounts pass
	tx = makeCallTxWithAmount(cfg, math.MaxInt64)
	require.NoError(t, exec.CheckTx(tx, 0))
	tx = makeCallTxWithAmount(cfg, 1e8)
	require.NoError(t, exec.CheckTx(tx, 0))
	tx = makeCallTxWithAmount(cfg, 0)
	require.NoError(t, exec.CheckTx(tx, 0))

	// pre-fork: historical behavior preserved (overflow amount NOT rejected here;
	// the exec-layer fork gate handles it on such chains)
	execPre := newTestExecutor(t, cfg, 999)
	tx = makeCallTxWithAmount(cfg, math.MaxUint64)
	require.NoError(t, execPre.CheckTx(tx, 0))
}
