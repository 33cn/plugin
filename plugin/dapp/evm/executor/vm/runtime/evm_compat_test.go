// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// ForkEVMCompat 的运行时边界：fork 前部署合约的 nonce 对齐、CREATE/CREATE2 的
// EIP-150 63/64 转发、CREATE2 递增调用者 nonce、构造函数内的 CREATE，以及
// ForkEVMTransferRevert 关闭时失败 CREATE 保留附带金额。

import (
	"bytes"
	"math/big"
	"strings"
	"testing"

	ctypes "github.com/33cn/chain33/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	vcomm "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	vcommcrypto "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common/crypto"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// newCompatEnv builds a VM environment whose ForkEVMCompat resolves at compatHeight.
// The default configstring keeps every other fork at its registered default.
func newCompatEnv(t *testing.T, compatHeight int64) *gapEnv {
	t.Helper()
	cfgStr := strings.Replace(ctypes.GetDefaultCfgstring(), `Title="local"`, `Title="test"`, 1)
	if !strings.Contains(cfgStr, "[exec.sub.evm]") {
		cfgStr += "\n[exec.sub.evm]\nethMapFromExecutor=\"coins\"\nethMapFromSymbol=\"bty\"\n"
	}
	cfg := ctypes.NewChain33Config(cfgStr)
	require.False(t, cfg.IsLocal(), "a local config would force every fork to height 0")
	cfg.SetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMCompat, compatHeight)
	require.Equal(t, compatHeight, cfg.GetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMCompat))
	require.Equal(t, compatHeight <= gapTestHeight,
		cfg.IsDappFork(gapTestHeight, evmtypes.ExecutorName, evmtypes.ForkEVMCompat))
	return newGapEnvWithCfg(t, cfg, "evm-compat")
}

// gasProbeInit returns init code whose first instruction is GAS. The GAS opcode pushes
// its own post-cost remaining gas, and RETURN hands that word back as the deployed code,
// so the child's initial allowance is readable from the stored code.
//
//	GAS; PUSH1 0; MSTORE; PUSH1 32; PUSH1 0; RETURN
func gasProbeInit() []byte {
	return []byte{0x5a, 0x60, 0x00, 0x52, 0x60, 0x20, 0x60, 0x00, 0xf3}
}

func gasFromCode(t *testing.T, code []byte) uint64 {
	t.Helper()
	require.Len(t, code, 32, "gas probe must deploy a 32-byte word")
	return new(big.Int).SetBytes(code).Uint64()
}

// persistedNonce reads the nonce from the account's state KV in the journaled changes.
func persistedNonce(t *testing.T, env *gapEnv, account vcomm.Address) uint64 {
	t.Helper()
	kvs, _ := env.mdb.GetChangedData(0)
	stateKey := []byte("mavl-" + evmtypes.ExecutorName + "-state: " + account.String())
	for _, kv := range kvs {
		if !bytes.Equal(kv.GetKey(), stateKey) {
			continue
		}
		var state evmtypes.EVMContractState
		require.NoError(t, ctypes.Decode(kv.GetValue(), &state))
		return state.GetNonce()
	}
	t.Fatalf("no state KV for %s", account.String())
	return 0
}

// TestCreateEIP150GasForwarding pins the child allowance for CREATE: the child receives
// 63/64 of the caller's remaining gas (EIP-150); the probe reports that minus GAS's cost.
func TestCreateEIP150GasForwarding(t *testing.T) {
	env := newCompatEnv(t, 0)
	caller, parent := addr(0x11), addr(0x12)
	env.deploy(parent, caller, []byte{0x00})

	const gasBefore = uint64(1_000_000)
	init := gasProbeInit()
	mem := NewMemory()
	mem.Resize(uint64(len(init)))
	mem.Set(0, uint64(len(init)), init)
	stack := newstack()
	stack.push(new(uint256.Int).SetUint64(uint64(len(init)))) // size
	stack.push(new(uint256.Int).SetUint64(0))                 // offset
	stack.push(new(uint256.Int).SetUint64(0))                 // value

	// Open a snapshot so the nonce increment is journaled and visible in the changed data.
	env.mdb.Snapshot()

	contract := NewContract(AccountRef(caller), AccountRef(parent), big.NewInt(0), gasBefore)
	callContext := &callCtx{memory: mem, stack: stack, contract: contract}
	pc := uint64(0)
	_, err := opCreate(&pc, env.evm, callContext)
	require.NoError(t, err)

	child := vcomm.BytesToAddress(ethcrypto.CreateAddress(gethcommon.BytesToAddress(parent.Bytes()), 1).Bytes())
	forwarded := gasBefore - gasBefore/64
	require.Equal(t, forwarded-2, gasFromCode(t, env.mdb.GetCode(child.String())),
		"CREATE must forward 63/64 of the remaining gas")
	// The pre-fork account (code, no stored nonce) counts as nonce 1 and is incremented to 2,
	// and that value reaches the persisted state KV.
	require.Equal(t, uint64(2), env.mdb.GetNonce(parent.String()))
	require.Equal(t, uint64(2), persistedNonce(t, env, parent))
}

// TestContractAddressCollisionRule pins the shared collision predicate: a balance alone
// is not a collision, while a non-zero nonce or code is.
func TestContractAddressCollisionRule(t *testing.T) {
	env := newCompatEnv(t, 0)
	caller := addr(0x61)
	env.fund(caller, 100)
	balanceOnly := addr(0x62)
	env.mdb.AddBalance(balanceOnly.String(), caller.String(), 100)
	require.Equal(t, uint64(100), env.mdb.GetBalance(balanceOnly.String()))
	require.False(t, env.evm.ContractAddressCollision(balanceOnly), "a balance alone must not collide")

	withCode := addr(0x63)
	env.mdb.CreateAccount(withCode.String(), caller.String(), "evm.test", "")
	env.mdb.SetCode(withCode.String(), []byte{0x00})
	require.True(t, env.evm.ContractAddressCollision(withCode), "code must collide")

	withNonce := addr(0x64)
	env.mdb.CreateAccount(withNonce.String(), caller.String(), "evm.test", "")
	env.mdb.SetNonce(withNonce.String(), 1)
	require.True(t, env.evm.ContractAddressCollision(withNonce), "a non-zero nonce must collide")
}

// TestCreate2EIP150GasForwarding pins the same 63/64 rule for CREATE2.
func TestCreate2EIP150GasForwarding(t *testing.T) {
	env := newCompatEnv(t, 0)
	caller, parent := addr(0x21), addr(0x22)
	env.deploy(parent, caller, []byte{0x00})

	const gasBefore = uint64(1_000_000)
	init := gasProbeInit()
	salt := uint256.NewInt(0x1234)
	mem := NewMemory()
	mem.Resize(uint64(len(init)))
	mem.Set(0, uint64(len(init)), init)
	stack := newstack()
	stack.push(salt)                                          // salt
	stack.push(new(uint256.Int).SetUint64(uint64(len(init)))) // size
	stack.push(new(uint256.Int).SetUint64(0))                 // offset
	stack.push(new(uint256.Int).SetUint64(0))                 // endowment

	contract := NewContract(AccountRef(caller), AccountRef(parent), big.NewInt(0), gasBefore)
	callContext := &callCtx{memory: mem, stack: stack, contract: contract}
	pc := uint64(0)
	_, err := opCreate2(&pc, env.evm, callContext)
	require.NoError(t, err)

	saltBytes := salt.Bytes32()
	child := vcommcrypto.CreateAddress2(parent, saltBytes, vcommcrypto.Keccak256(init))
	forwarded := gasBefore - gasBefore/64
	require.Equal(t, forwarded-2, gasFromCode(t, env.mdb.GetCode(child.String())),
		"CREATE2 must forward 63/64 of the remaining gas")
	// CREATE2 also increments the caller nonce.
	require.Equal(t, uint64(2), env.mdb.GetNonce(parent.String()))
}

// TestAccountNonceForPreForkContract pins the shared nonce helper: an account with code
// but no stored nonce counts as 1 under the fork and 0 without it.
func TestAccountNonceForPreForkContract(t *testing.T) {
	for _, tc := range []struct {
		name      string
		compatAt  int64
		wantNonce uint64
	}{
		{"fork off keeps the stored nonce", ctypes.MaxHeight, 0},
		{"fork on treats code-only account as 1", 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := newCompatEnv(t, tc.compatAt)
			account := addr(0x31)
			env.mdb.CreateAccount(account.String(), account.String(), "evm.test", "")
			env.mdb.SetCode(account.String(), []byte{0x00})
			require.Equal(t, uint64(0), env.mdb.GetNonce(account.String()))
			require.Equal(t, tc.wantNonce, env.evm.accountNonce(account))
		})
	}
}

// buildConstructorCreate returns init code that CREATEs childInit from memory and then
// stops with an empty runtime body.
func buildConstructorCreate(childInit []byte) []byte {
	if len(childInit) > 32 {
		panic("childInit longer than one word")
	}
	code := []byte{byte(0x5f + len(childInit))}
	code = append(code, childInit...)
	code = append(code,
		0x60, 0x00, // PUSH1 0
		0x52,                       // MSTORE
		0x60, byte(len(childInit)), // PUSH1 size
		0x60, byte(32-len(childInit)), // PUSH1 offset
		0x60, 0x00, // PUSH1 value
		0xf0, // CREATE
		0x50, // POP
		0x00, // STOP
	)
	return code
}

// TestCreateInsideConstructor covers a CREATE issued while the caller contract is still
// running its constructor: it sees nonce 1 (EIP-161) and the child is deployed.
func TestCreateInsideConstructor(t *testing.T) {
	env := newCompatEnv(t, 0)
	caller, contractA := addr(0x41), addr(0x42)

	// Child init code returns the one-byte runtime 0x2a:
	// PUSH1 0x2a; PUSH1 0; MSTORE8; PUSH1 1; PUSH1 0; RETURN
	childInit := []byte{0x60, 0x2a, 0x60, 0x00, 0x53, 0x60, 0x01, 0x60, 0x00, 0xf3}
	init := buildConstructorCreate(childInit)

	_, _, _, err := env.evm.Create(AccountRef(caller), contractA, init, 5_000_000, "innerContract", "", 0)
	require.NoError(t, err)

	child := vcomm.BytesToAddress(ethcrypto.CreateAddress(gethcommon.BytesToAddress(contractA.Bytes()), 1).Bytes())
	require.Equal(t, []byte{0x2a}, env.mdb.GetCode(child.String()))
	require.Equal(t, uint64(2), env.mdb.GetNonce(contractA.String()),
		"the constructor starts at nonce 1 and its CREATE increments it")
}

// TestCompatCreateKeepsValueWithoutTransferRevert documents the boundary between the two
// forks: ForkEVMCompat alone does not roll back the value attached to a failed CREATE.
func TestCompatCreateKeepsValueWithoutTransferRevert(t *testing.T) {
	env := newCompatEnv(t, 0)
	require.Equal(t, int64(ctypes.MaxHeight),
		env.cfg.GetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMTransferRevert))

	const stash = int64(1000)
	const attached = uint64(5)
	caller, parent := addr(0x51), addr(0x52)
	env.fund(parent, stash)
	env.deploy(parent, caller, buildCreateWithValueCode(attached, []byte{0xfe}))

	_, _, _, err := env.evm.Call(AccountRef(caller), parent, nil, 5_000_000, 0)
	require.NoError(t, err)
	require.Equal(t, stash-int64(attached), env.balance(parent),
		"without ForkEVMTransferRevert the failed CREATE keeps the attached value")
}
