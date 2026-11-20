// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// Fork-gated coverage for the opcodes enabled by ForkEVMCompat: PUSH0 (EIP-3855) and
// BASEFEE (EIP-3198).

import (
	"encoding/binary"
	"testing"

	ctypes "github.com/33cn/chain33/types"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

// Runtime bytecode for the enabled-fork cases. Each program stores its result at memory
// offset 0 and returns one 32-byte word.

// PUSH0; ISZERO; MSTORE(0, 1); RETURN(0, 32)
var push0Code = []byte{
	0x5f,       // PUSH0
	0x15,       // ISZERO -> 1
	0x60, 0x00, // PUSH1 0 (memory offset)
	0x52,       // MSTORE
	0x60, 0x20, // PUSH1 32 (return size)
	0x60, 0x00, // PUSH1 0 (return offset)
	0xf3, // RETURN
}

// BASEFEE; ISZERO; MSTORE(0, 1); RETURN(0, 32)
var baseFeeCode = []byte{
	0x48,       // BASEFEE
	0x15,       // ISZERO -> 1
	0x60, 0x00, // PUSH1 0
	0x52,       // MSTORE
	0x60, 0x20, // PUSH1 32
	0x60, 0x00, // PUSH1 0
	0xf3, // RETURN
}

func callReturnData(t *testing.T, receipt *ctypes.Receipt) []byte {
	t.Helper()
	for _, log := range receipt.GetLogs() {
		if log.GetTy() != evmtypes.TyLogCallContract {
			continue
		}
		var contract evmtypes.ReceiptEVMContract
		require.NoError(t, ctypes.Decode(log.GetLog(), &contract))
		return contract.Ret
	}
	t.Fatal("receipt has no TyLogCallContract log")
	return nil
}

func callContract(t *testing.T, cfg *ctypes.Chain33Config, exec *EVMExecutor, contractAddr string, input []byte) *ctypes.Receipt {
	t.Helper()
	receipt, err := exec.Exec(makeCallTx(cfg, roleLegitUser, contractAddr, input, 0), 0)
	require.NoError(t, err)
	require.Equal(t, int32(ctypes.ExecOk), receipt.GetTy())
	applyReceiptKV(t, exec, receipt)
	return receipt
}

func uint64Word(v uint64) []byte {
	word := make([]byte, 32)
	binary.BigEndian.PutUint64(word[24:], v)
	return word
}

func TestEVMCompatOpcodesDisabled(t *testing.T) {
	restoreBtcAddressDriver(t)
	cfg := newTestConfig(t)
	disableEVMCompat(cfg)
	exec := newTestExecutor(t, cfg, 0)

	// The new bytes must stay invalid opcodes when the fork is off. Only the byte that is
	// actually executed matters, so a single unknown opcode followed by STOP is enough.
	for _, tc := range []struct {
		name string
		op   byte
	}{
		{"PUSH0", 0x5f},
		{"BASEFEE", 0x48},
	} {
		t.Run(tc.name, func(t *testing.T) {
			contractAddr := deployCode(t, cfg, exec, []byte{tc.op, 0x00})
			_, err := exec.Exec(makeCallTx(cfg, roleLegitUser, contractAddr, nil, 0), 0)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid opcode")
		})
	}
}

func TestEVMCompatOpcodesEnabled(t *testing.T) {
	restoreBtcAddressDriver(t)
	cfg := newTestConfig(t)
	enableEVMCompat(cfg)
	exec := newTestExecutor(t, cfg, 0)

	t.Run("PUSH0", func(t *testing.T) {
		contractAddr := deployCode(t, cfg, exec, push0Code)
		ret := callReturnData(t, callContract(t, cfg, exec, contractAddr, nil))
		require.Equal(t, uint64Word(1), ret, "PUSH0 must push zero, so ISZERO yields one")
	})

	t.Run("BASEFEE", func(t *testing.T) {
		contractAddr := deployCode(t, cfg, exec, baseFeeCode)
		ret := callReturnData(t, callContract(t, cfg, exec, contractAddr, nil))
		require.Equal(t, uint64Word(1), ret, "BASEFEE returns zero because chain33 has no base fee")
	})
}
