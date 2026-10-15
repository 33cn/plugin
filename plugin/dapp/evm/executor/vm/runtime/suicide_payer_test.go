// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// SELFDESTRUCT payer semantics, before and after ForkEVMSuicidePayer.
//
// opSuicide reads the amount from the executing account (contract.Address()) but used to take
// the money from the code owner (contract.CodeAddr). Under CALL/CREATE the two addresses are
// the same, so nothing changes there. Under DELEGATECALL/CALLCODE they differ, so a proxy
// holding X could make an unrelated contract hand X to a beneficiary of the proxy's choosing.
//
// The cases below use the blacklist gap harness (gapEnv in blacklist_gap_test.go) for the
// state/coins plumbing, but with a non-local config: under Title="local" needSetForkZero pins
// every dapp fork to 0, which would make the "before the fork" branch unreachable and any
// such test a false pass.

import (
	"math/big"
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	ctypes "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	vmcommon "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/state"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

// newSuicideEnv builds a non-local EVM environment with ForkEVMSuicidePayer pinned to
// forkHeight (0 = active, ctypes.MaxHeight = closed) and asserts the fork really resolves
// that way at the height the VM executes at.
func newSuicideEnv(t *testing.T, forkHeight int64) *gapEnv {
	t.Helper()
	cfgStr := strings.Replace(ctypes.GetDefaultCfgstring(), `Title="local"`, `Title="test"`, 1)
	if !strings.Contains(cfgStr, "[exec.sub.evm]") {
		cfgStr += "\n[exec.sub.evm]\nethMapFromExecutor=\"coins\"\nethMapFromSymbol=\"bty\"\n"
	}
	cfg := ctypes.NewChain33Config(cfgStr)
	require.False(t, cfg.IsLocal(), "a local config would force every fork to height 0")
	cfg.SetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMSuicidePayer, forkHeight)
	require.Equal(t, forkHeight, cfg.GetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMSuicidePayer),
		"the fork height must not be rewritten by the config layer")
	require.Equal(t, forkHeight <= gapTestHeight,
		cfg.IsDappFork(gapTestHeight, evmtypes.ExecutorName, evmtypes.ForkEVMSuicidePayer),
		"the fork must resolve to the intended side at the block height the VM runs at")

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)

	dbDir, stateDB, localDB := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dbDir, stateDB) })
	coins, err := account.NewAccountDB(cfg, "coins", cfg.GetCoinSymbol(), stateDB)
	require.NoError(t, err)

	mdb := state.NewMemoryStateDB(stateDB, localDB, coins, gapTestHeight, api)
	mdb.Prepare(vmcommon.BytesToHash([]byte("suicide-payer")), 0)

	env := &gapEnv{t: t, cfg: cfg, mdb: mdb, coins: coins}
	env.evm = NewEVM(Context{
		CanTransfer: func(db state.EVMStateDB, s vmcommon.Address, v uint64) bool { return db.CanTransfer(s.String(), v) },
		Transfer: func(db state.EVMStateDB, s, r vmcommon.Address, v uint64) bool {
			return db.Transfer(s.String(), r.String(), v)
		},
		GetHash:     func(uint64) vmcommon.Hash { return vmcommon.Hash{} },
		BlockNumber: big.NewInt(gapTestHeight),
	}, mdb, Config{}, cfg)
	return env
}

// buildCallCodeInstruction emits the CALLCODE (0xf2) instruction. Note that buildCallCode in
// log_address_integration_test.go is a plain CALL despite its name.
func buildCallCodeInstruction(callee vmcommon.Address, value uint64) []byte {
	code := []byte{
		0x60, 0x00, // retSize
		0x60, 0x00, // retOffset
		0x60, 0x00, // inSize
		0x60, 0x00, // inOffset
	}
	code = append(code, pushUint64(value)...)
	code = append(code, 0x73) // PUSH20 <callee>
	code = append(code, callee.Bytes()...)
	code = append(code,
		0x61, 0xff, 0xff, // gas
		0xf2, // CALLCODE
		0x00, // STOP
	)
	return code
}

const (
	// proxyStash is both the proxy's balance and the amount the attack moves.
	proxyStash = int64(300_000)
	// victimStash must be >= proxyStash, otherwise the pre-fork transfer fails silently
	// and the drain is not observable.
	victimStash = int64(1_000_000)
)

// TestSuicidePayerDelegatecallPreFork pins today's (defective) behaviour: the amount is read
// from the proxy, but the proxy's own balance stays put and the borrowed-code contract pays.
func TestSuicidePayerDelegatecallPreFork(t *testing.T) {
	user, proxy, victim := addr(0x21), addr(0xb1), addr(0xc1)
	env := newSuicideEnv(t, ctypes.MaxHeight)

	env.fund(proxy, proxyStash)
	env.fund(victim, victimStash)
	env.deploy(proxy, user, buildDelegateCallCode(victim))
	env.deploy(victim, user, buildSelfdestructCode(user))

	_, _, _, err := env.evm.Call(AccountRef(user), proxy, nil, 5_000_000, 0)
	require.NoError(t, err)

	require.Equal(t, proxyStash, env.balance(user), "beneficiary is paid the proxy's balance")
	require.Equal(t, victimStash-proxyStash, env.balance(victim), "the code owner is the payer")
	require.Equal(t, proxyStash, env.balance(proxy), "the proxy keeps its own balance")
	require.True(t, env.mdb.HasSuicided(proxy.String()))
	require.False(t, env.mdb.HasSuicided(victim.String()))
}

// TestSuicidePayerDelegatecallPostFork the proxy can only spend the balance the proxy holds.
func TestSuicidePayerDelegatecallPostFork(t *testing.T) {
	user, proxy, victim := addr(0x21), addr(0xb1), addr(0xc1)
	env := newSuicideEnv(t, 0)

	env.fund(proxy, proxyStash)
	env.fund(victim, victimStash)
	env.deploy(proxy, user, buildDelegateCallCode(victim))
	env.deploy(victim, user, buildSelfdestructCode(user))

	_, _, _, err := env.evm.Call(AccountRef(user), proxy, nil, 5_000_000, 0)
	require.NoError(t, err)

	require.Zero(t, env.balance(proxy), "the proxy pays out its own balance")
	require.Equal(t, proxyStash, env.balance(user))
	require.Equal(t, victimStash, env.balance(victim), "the code owner must be untouched")
	require.Equal(t, int64(0), env.balance(user)-proxyStash, "the attacker nets nothing")
	require.True(t, env.mdb.HasSuicided(proxy.String()))
	require.False(t, env.mdb.HasSuicided(victim.String()))
}

// TestSuicidePayerCallCodeUnchanged CALLCODE diverges the same way as DELEGATECALL.
func TestSuicidePayerCallCodeUnchanged(t *testing.T) {
	user, proxy, victim := addr(0x24), addr(0xb4), addr(0xc4)
	for _, tc := range []struct {
		name       string
		forkHeight int64
	}{{"pre-fork", ctypes.MaxHeight}, {"post-fork", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			env := newSuicideEnv(t, tc.forkHeight)
			env.fund(proxy, proxyStash)
			env.fund(victim, victimStash)
			env.deploy(proxy, user, buildCallCodeInstruction(victim, 0))
			env.deploy(victim, user, buildSelfdestructCode(user))

			_, _, _, err := env.evm.Call(AccountRef(user), proxy, nil, 5_000_000, 0)
			require.NoError(t, err)

			require.Equal(t, proxyStash, env.balance(user))
			require.True(t, env.mdb.HasSuicided(proxy.String()))
			if tc.forkHeight == ctypes.MaxHeight {
				require.Equal(t, victimStash-proxyStash, env.balance(victim), "pre-fork the code owner pays")
				require.Equal(t, proxyStash, env.balance(proxy))
			} else {
				require.Equal(t, victimStash, env.balance(victim), "post-fork the code owner is untouched")
				require.Zero(t, env.balance(proxy))
			}
		})
	}
}

// TestSuicidePayerCallUnchangedAddressEqualsCodeAddr under a plain CALL the payer is the same
// account either way, so the fork must be a no-op.
func TestSuicidePayerCallUnchangedAddressEqualsCodeAddr(t *testing.T) {
	user, callee := addr(0x22), addr(0xc2)
	for _, forkHeight := range []int64{ctypes.MaxHeight, 0} {
		env := newSuicideEnv(t, forkHeight)
		env.fund(callee, victimStash)
		env.deploy(callee, user, buildSelfdestructCode(user))

		_, _, _, err := env.evm.Call(AccountRef(user), callee, nil, 5_000_000, 0)
		require.NoError(t, err)

		require.Equal(t, victimStash, env.balance(user), "fork=%d", forkHeight)
		require.Zero(t, env.balance(callee), "fork=%d", forkHeight)
		require.True(t, env.mdb.HasSuicided(callee.String()), "fork=%d", forkHeight)
	}
}

// TestSuicidePayerCreateUnchangedAddressEqualsCodeAddr the same for CREATE: evm.Create sets
// CodeAddr to the new contract address, so both fork sides must behave identically.
func TestSuicidePayerCreateUnchangedAddressEqualsCodeAddr(t *testing.T) {
	user, created := addr(0x23), addr(0xc3)
	for _, forkHeight := range []int64{ctypes.MaxHeight, 0} {
		env := newSuicideEnv(t, forkHeight)
		env.fund(user, proxyStash)

		_, _, _, err := env.evm.Create(AccountRef(user), created, buildSelfdestructCode(user),
			5_000_000, "evm."+created.String(), created.String(), uint64(proxyStash))
		require.NoError(t, err)

		require.Equal(t, int64(0), env.balance(created), "fork=%d: the new contract pays out its value", forkHeight)
		require.Equal(t, proxyStash, env.balance(user), "fork=%d: the beneficiary gets the value back", forkHeight)
	}
}
