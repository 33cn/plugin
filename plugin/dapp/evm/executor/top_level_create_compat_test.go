// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"testing"

	"github.com/33cn/chain33/common/address"
	ctypes "github.com/33cn/chain33/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	vcomm "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// TestTopLevelETHSignedCreateUsesRawAddressUnderFork deploys with an ETH-signed CREATE
// while the BTC address driver is active. Under ForkEVMCompat the contract address must be
// the raw 20-byte CreateAddress(sender, txNonce); the driver round trip would zero it.
func TestTopLevelETHSignedCreateUsesRawAddressUnderFork(t *testing.T) {
	restoreBtcAddressDriver(t)
	cfg := newTestConfig(t)
	enableEVMCompat(cfg)
	exec := newTestExecutor(t, cfg, 0)

	execAddr := address.ExecAddress(cfg.ExecName(evmtypes.ExecutorName))
	action := &evmtypes.EVMContractAction{Code: vcomm.FromHex("60006000f3"), ContractAddr: execAddr}
	tx := &ctypes.Transaction{
		ChainID: cfg.GetChainID(),
		Execer:  []byte(cfg.ExecName(evmtypes.ExecutorName)),
		Payload: ctypes.Encode(action),
		Fee:     1e6, To: execAddr, Nonce: 0,
	}
	signTx(tx, roleDeployer)
	// Use the ETH signature id the executor dispatches on (signTx only sets the crypto id).
	tx.Signature.Ty = ctypes.EncodeSignID(ctypes.SECP256K1ETH, ctypes.EthAddressID)
	resetEvmExecAddress(cfg)

	receipt, err := exec.Exec(tx, 0)
	require.NoError(t, err)
	require.Equal(t, int32(ctypes.ExecOk), receipt.GetTy())

	sender := vcomm.StringToAddress(tx.From())
	require.NotNil(t, sender)
	raw := ethcrypto.CreateAddress(gethcommon.BytesToAddress(sender.Bytes()), uint64(tx.Nonce))
	want := vcomm.BytesToAddress(raw.Bytes())

	wantKey := []byte("mavl-" + evmtypes.ExecutorName + "-data: " + want.String())
	for _, kv := range receipt.GetKV() {
		if bytes.Equal(kv.GetKey(), wantKey) {
			return
		}
	}
	t.Fatalf("no contract data KV at the raw ETH address %s", want.String())
}

// TestTopLevelCreateAllowsBalanceOnlyAddressUnderFork pre-funds the derived contract
// address with coins only: under ForkEVMCompat that is not a collision (only nonce or
// code is), so the deploy succeeds.
func TestTopLevelCreateAllowsBalanceOnlyAddressUnderFork(t *testing.T) {
	restoreBtcAddressDriver(t)
	cfg := newTestConfig(t)
	enableEVMCompat(cfg)
	exec := newTestExecutor(t, cfg, 0)

	execAddr := address.ExecAddress(cfg.ExecName(evmtypes.ExecutorName))
	action := &evmtypes.EVMContractAction{Code: vcomm.FromHex("60006000f3"), ContractAddr: execAddr}
	tx := &ctypes.Transaction{
		ChainID: cfg.GetChainID(),
		Execer:  []byte(cfg.ExecName(evmtypes.ExecutorName)),
		Payload: ctypes.Encode(action),
		Fee:     1e6, To: execAddr, Nonce: 0,
	}
	signTx(tx, roleDeployer)
	tx.Signature.Ty = ctypes.EncodeSignID(ctypes.SECP256K1ETH, ctypes.EthAddressID)
	resetEvmExecAddress(cfg)

	sender := vcomm.StringToAddress(tx.From())
	require.NotNil(t, sender)
	raw := ethcrypto.CreateAddress(gethcommon.BytesToAddress(sender.Bytes()), uint64(tx.Nonce))
	target := vcomm.BytesToAddress(raw.Bytes())
	fundAddr(t, exec, target.String(), 100)

	receipt, err := exec.Exec(tx, 0)
	require.NoError(t, err)
	require.Equal(t, int32(ctypes.ExecOk), receipt.GetTy())
}
