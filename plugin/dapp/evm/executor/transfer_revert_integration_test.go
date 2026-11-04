// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// Transaction-level repro of the reverted-transfer mint: A calls B with value, B reverts, then
// the phantom credit B kept in the account cache is spent by a later call in the same tx. The
// committed state is assembled from the receipt KV only, so before ForkEVMTransferRevert the
// transaction mints the moved amount. From the fork on the first transfer is rolled back and
// the later spend has no balance to take.

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/address/btc"
	ctypes "github.com/33cn/chain33/types"
	vcomm "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

func newTransferRevertConfig(t *testing.T, forkHeight int64) *ctypes.Chain33Config {
	t.Helper()
	cfgStr := strings.Replace(ctypes.GetDefaultCfgstring(), `Title="local"`, `Title="integration-test"`, 1)
	if !strings.Contains(cfgStr, "[exec.sub.evm]") {
		cfgStr += "\n[exec.sub.evm]\nethMapFromExecutor=\"coins\"\nethMapFromSymbol=\"bty\"\n"
	}
	cfg := ctypes.NewChain33Config(cfgStr)
	require.False(t, cfg.IsLocal())
	cfg.SetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMTransferRevert, forkHeight)
	require.Equal(t, forkHeight, cfg.GetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMTransferRevert))
	return cfg
}

func pushUint64(v uint64) []byte {
	if v == 0 {
		return []byte{0x60, 0x00}
	}
	var b []byte
	for n := v; n > 0; n >>= 8 {
		b = append([]byte{byte(n)}, b...)
	}
	return append([]byte{0x60 + byte(len(b)-1)}, b...)
}

// buildCall emits CALL(callee, value) with the given calldata size (memory offset 0).
func buildCall(callee vcomm.Address, value, inSize uint64) []byte {
	code := pushUint64(0)                      // retSize
	code = append(code, pushUint64(0)...)      // retOffset
	code = append(code, pushUint64(inSize)...) // inSize
	code = append(code, pushUint64(0)...)      // inOffset
	code = append(code, pushUint64(value)...)  // value
	code = append(code, 0x73)
	code = append(code, callee.Bytes()...)
	code = append(code, pushUint64(0xffff)...)
	code = append(code, 0xf1) // CALL
	code = append(code, 0x50) // POP
	return code
}

// buildCallerCode calls callee twice: first with value (the callee aborts), then with one byte
// of calldata, which makes the callee forward its balance.
func buildCallerCode(callee vcomm.Address, value uint64) []byte {
	code := buildCall(callee, value, 0)
	code = append(code, buildCall(callee, 0, 1)...)
	code = append(code, 0x00) // STOP
	return code
}

// buildForwarderCode aborts when calldata is empty, otherwise CALLs target with value.
func buildForwarderCode(target vcomm.Address, value uint64) []byte {
	code := []byte{
		0x36,       // CALLDATASIZE
		0x60, 0x08, // PUSH1 8
		0x57,             // JUMPI
		0xfe,             // INVALID
		0x00, 0x00, 0x00, // padding
		0x5b, // JUMPDEST @8
	}
	code = append(code, buildCall(target, value, 0)...)
	code = append(code, 0x00) // STOP
	return code
}

func makeDeployCodeTx(cfg *ctypes.Chain33Config, roleIdx int, code []byte) *ctypes.Transaction {
	execAddr := address.ExecAddress(cfg.ExecName(evmtypes.ExecutorName))
	action := &evmtypes.EVMContractAction{
		Amount: 0, GasLimit: 0, GasPrice: 0,
		Code: code, ContractAddr: execAddr,
	}
	tx := &ctypes.Transaction{
		ChainID: cfg.GetChainID(),
		Execer:  []byte(cfg.ExecName(evmtypes.ExecutorName)),
		Payload: ctypes.Encode(action),
		Fee:     1e6, To: execAddr, Nonce: 0,
	}
	signTx(tx, roleIdx)
	return tx
}

func applyReceiptKV(t *testing.T, exec *EVMExecutor, receipt *ctypes.Receipt) {
	t.Helper()
	for _, kv := range receipt.GetKV() {
		require.NoError(t, exec.GetStateDB().Set(kv.GetKey(), kv.GetValue()))
	}
}

// wrapInit turns runtime code into CREATE init code that returns it.
func wrapInit(runtime []byte) []byte {
	code := []byte{0x60, byte(len(runtime))} // PUSH1 len
	offsetAt := len(code)
	code = append(code, 0x60, 0x00)               // PUSH1 codeOffset (patched below)
	code = append(code, 0x60, 0x00)               // PUSH1 0 (mem offset)
	code = append(code, 0x39)                     // CODECOPY
	code = append(code, 0x60, byte(len(runtime))) // PUSH1 return size
	code = append(code, 0x60, 0x00)               // PUSH1 0 (return offset)
	code = append(code, 0xf3)                     // RETURN
	code[offsetAt+1] = byte(len(code))
	return append(code, runtime...)
}

func deployCode(t *testing.T, cfg *ctypes.Chain33Config, exec *EVMExecutor, code []byte) string {
	t.Helper()
	tx := makeDeployCodeTx(cfg, roleDeployer, wrapInit(code))
	receipt, err := exec.Exec(tx, 0)
	require.NoError(t, err)
	require.Equal(t, int32(ctypes.ExecOk), receipt.GetTy())
	applyReceiptKV(t, exec, receipt)
	return vcomm.NewContractAddress(*vcomm.StringToAddress(tx.From()), tx.Hash()).String()
}

func TestRevertedCallValueCannotBeSpent(t *testing.T) {
	const (
		topUp = uint64(1000)
		moved = uint64(400)
	)

	// Other tests in this package (query_test.go) replace the package-wide EVM address driver
	// with the ETH one. The executor globals were initialized with the default BTC driver, so
	// restore it or the tx envelope address no longer matches the EVM exec address.
	driver, err := address.LoadDriver(btc.NormalAddressID, -1)
	require.NoError(t, err)
	vcomm.InitEvmAddressDriver(driver)

	for _, tc := range []struct {
		name       string
		forkHeight int64
		minted     uint64
	}{
		{"pre-fork mints the reverted value", ctypes.MaxHeight, moved},
		{"post-fork mints nothing", 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newTransferRevertConfig(t, tc.forkHeight)
			exec := newTestExecutor(t, cfg, 0)

			target := vcomm.BytesToAddress([]byte{0xd0})
			forwarder := deployCode(t, cfg, exec, buildForwarderCode(target, moved))
			caller := deployCode(t, cfg, exec, buildCallerCode(*vcomm.StringToAddress(forwarder), moved))

			user := addrFromRole(cfg, roleLegitUser)
			fundRole(t, exec, cfg, roleLegitUser, int64(topUp)*2)

			sum := func() int64 {
				return balanceOf(exec, user) + balanceOf(exec, caller) +
					balanceOf(exec, forwarder) + balanceOf(exec, target.String())
			}
			before := sum()

			receipt, err := exec.Exec(makeCallTx(cfg, roleLegitUser, caller, []byte{0x01}, topUp), 0)
			require.NoError(t, err)
			require.Equal(t, int32(ctypes.ExecOk), receipt.GetTy())
			applyReceiptKV(t, exec, receipt)

			require.Equal(t, before+int64(tc.minted), sum(), "committed total must not grow beyond the minted amount")
			require.Equal(t, int64(tc.minted), balanceOf(exec, target.String()), "the forwarder must not spend a phantom credit")
		})
	}
}

func balanceOf(exec *EVMExecutor, addr string) int64 {
	return exec.GetMStateDB().CoinsAccount.LoadAccount(addr).Balance
}
