// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// 本文件验证的是 chain33 框架层（types.CheckTxBlockedAccount）看不见、
// 只能由 EVM 内部检查拦截的资产打出路径。每个用例都：
//   1. 用真实 MemoryStateDB + coins 账户，断言余额而不是只断言 error；
//   2. 先证明交易信封上的地址全部干净（chain33 放行），再证明 EVM 层拦下；
//   3. 以 fork 关闭作为对照，证明放行时资产确实会被打出，测试不是空转。
//
// 场景编号对应 docs/security/evm-account-blacklist.md 的 B1-B4。

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
	"github.com/stretchr/testify/require"
)

const gapTestHeight = int64(700000)

type gapEnv struct {
	t     *testing.T
	cfg   *ctypes.Chain33Config
	mdb   *state.MemoryStateDB
	coins *account.DB
	evm   *EVM
}

// newGapEnv 构建带真实 coins 账户的 EVM 运行环境。
// 默认 cfgstring 的 Title="local" 会把所有 fork（含 ForkAccountBlacklist）置为 0，
// 需要关闭 fork 做对照时由调用方显式 SetFork。
// 追加 ethMapFromExecutor="coins" 使 statedb 直接读写 coins 主账户，
// 否则余额走 evm 子账户（LoadExecAccount），fund/balance 会与实际转账路径不一致。
func newGapEnv(t *testing.T, blocked []string) *gapEnv {
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

	coins, err := account.NewAccountDB(cfg, "coins", cfg.GetCoinSymbol(), stateDB)
	require.NoError(t, err)

	mdb := state.NewMemoryStateDB(stateDB, localDB, coins, gapTestHeight, api)
	mdb.Prepare(vmcommon.BytesToHash([]byte("blacklist-gap")), 0)

	restore := ctypes.SetBlockedAccountsForTest(blocked)
	t.Cleanup(restore)

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

func (e *gapEnv) fund(addr vmcommon.Address, amount int64) {
	acc := e.coins.LoadAccount(addr.String())
	acc.Balance = amount
	e.coins.SaveAccount(acc)
}

func (e *gapEnv) balance(addr vmcommon.Address) int64 {
	return e.coins.LoadAccount(addr.String()).Balance
}

func (e *gapEnv) deploy(addr, creator vmcommon.Address, code []byte) {
	e.mdb.CreateAccount(addr.String(), creator.String(), "evm."+addr.String(), addr.String())
	e.mdb.SetCode(addr.String(), code)
}

// assertEnvelopeClean 证明交易信封上的地址 chain33 都放行：
// 这就是框架层 CheckTxBlockedAccount 对这类内部交易"看不见"的原因。
func (e *gapEnv) assertEnvelopeClean(addrs ...vmcommon.Address) {
	e.t.Helper()
	for _, a := range addrs {
		require.False(e.t, ctypes.IsBlockedAccount(a.String()), "envelope address %s must be clean", a)
	}
}

func addr(b byte) vmcommon.Address { return vmcommon.BytesToAddress([]byte{b}) }

// buildValueCallCode 生成 "CALL <callee> 携带 value" 的合约代码（inputs 全 0）。
func buildValueCallCode(callee vmcommon.Address, value uint64) []byte {
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
		0xf1, // CALL
		0x00, // STOP
	)
	return code
}

// buildDelegateCallCode 生成 "DELEGATECALL <callee>" 的合约代码。
func buildDelegateCallCode(callee vmcommon.Address) []byte {
	code := []byte{
		0x60, 0x00, // retSize
		0x60, 0x00, // retOffset
		0x60, 0x00, // inSize
		0x60, 0x00, // inOffset
		0x73, // PUSH20 <callee>
	}
	code = append(code, callee.Bytes()...)
	code = append(code,
		0x61, 0xff, 0xff, // gas
		0xf4, // DELEGATECALL
		0x00, // STOP
	)
	return code
}

// buildSelfdestructCode 生成 "SELFDESTRUCT <beneficiary>" 的合约代码。
func buildSelfdestructCode(beneficiary vmcommon.Address) []byte {
	code := []byte{0x73} // PUSH20 <beneficiary>
	code = append(code, beneficiary.Bytes()...)
	return append(code, 0xff) // SELFDESTRUCT
}

func pushUint64(v uint64) []byte {
	b := new(big.Int).SetUint64(v).Bytes()
	if len(b) == 0 {
		b = []byte{0}
	}
	return append([]byte{0x60 + byte(len(b)-1)}, b...)
}

// TestGapB1_BlockedContractWokenByInnerCall 场景 B1：
//
//	干净用户 U → 干净合约 B → CALL → 黑名单合约 A → A 向 U 转出自身余额
//
// chain33 只看到 from=U、to=B，A 只出现在 B 的运行时栈上。
// evm.Call 的 target 检查是唯一能阻止 A 被唤醒的位置。
func TestGapB1_BlockedContractWokenByInnerCall(t *testing.T) {
	user, relay, blockedC := addr(0x11), addr(0xb1), addr(0xa1)
	env := newGapEnv(t, []string{blockedC.String()})
	env.assertEnvelopeClean(user, relay)

	const stash = int64(1_000_000)
	env.fund(blockedC, stash)
	env.deploy(relay, user, buildValueCallCode(blockedC, 0))
	env.deploy(blockedC, user, buildValueCallCode(user, uint64(stash)))

	_, _, _, err := env.evm.Call(AccountRef(user), relay, nil, 5_000_000, 0)
	require.NoError(t, err, "outer call is legal; inner failure is a revert, not a tx error")
	require.Equal(t, stash, env.balance(blockedC), "blocked contract must keep its balance")
	require.Zero(t, env.balance(user), "user must not receive blocked funds")

	// 对照：fork 关闭时同一条链路会把钱打出去，证明上面的断言不是空转。
	env.cfg.SetFork(ctypes.ForkAccountBlacklist, ctypes.MaxHeight)
	_, _, _, err = env.evm.Call(AccountRef(user), relay, nil, 5_000_000, 0)
	require.NoError(t, err)
	require.Zero(t, env.balance(blockedC), "with fork off the funds must leave (proves the path is live)")
	require.Equal(t, stash, env.balance(user))
}

// TestGapB2_TokenPrecompileThirdPartyFrom 场景 B2：
// token 预编译 transfer(from,to,amount) 的 from 取自 calldata，与 caller 无绑定，
// 第三方合约可任意指定 from=黑名单。chain33 侧 Para 长度 100 字节，
// IsBlockedAccountRaw 要求精确 20 字节，完全看不见。
// statedb.TransferToToken 的 from 检查是这条链路的唯一防线。
func TestGapB2_TokenPrecompileThirdPartyFrom(t *testing.T) {
	victim, receiver, manager, caller := addr(0xa2), addr(0x22), addr(0x33), addr(0xc2)
	env := newGapEnv(t, []string{victim.String()})
	env.assertEnvelopeClean(receiver, manager, caller)

	// 由 manage 名单内的创建者部署的第三方合约作为 precompile 的 caller
	env.deploy(caller, manager, nil)
	precompile := vmcommon.BytesToAddress(vmcommon.FromHex(TokenPrecompileAddr))
	saved, had := CustomizePrecompiledContracts[precompile.ToHash160()]
	CustomizePrecompiledContracts[precompile.ToHash160()] = NewTokenPrecompile(&TokenContract{SuperManager: []string{manager.String()}})
	t.Cleanup(func() {
		if had {
			CustomizePrecompiledContracts[precompile.ToHash160()] = saved
		} else {
			delete(CustomizePrecompiledContracts, precompile.ToHash160())
		}
	})

	// transfer(address,address,uint256) 的 ABI 编码 = 4 + 32*3 字节，from 落在 [4:36] 的后 20 字节
	calldata := vmcommon.FromHex("0x" + transfer)
	calldata = append(calldata, make([]byte, 32*3)...)
	copy(calldata[4+12:], victim.Bytes())
	copy(calldata[36+12:], receiver.Bytes())
	calldata[len(calldata)-1] = 100
	require.False(t, ctypes.IsBlockedAccountRaw(calldata), "chain33 Para check must be blind to 100-byte calldata")

	ret, _, err := RunStateFulPrecompiledContract(env.evm, AccountRef(caller), CustomizePrecompiledContracts[precompile.ToHash160()], calldata, 100000)
	require.ErrorIs(t, err, ctypes.ErrBlockedAccount, "ret=%s", ret)

	// 检查一旦被移除，同一调用会落到 tokenStatus 的 ErrNotFound（本测试不注册 token），
	// 不再是 ErrBlockedAccount —— 用它区分"被黑名单拦下"和"因别的原因失败"。
	env.cfg.SetFork(ctypes.ForkAccountBlacklist, ctypes.MaxHeight)
	_, _, err = RunStateFulPrecompiledContract(env.evm, AccountRef(caller), CustomizePrecompiledContracts[precompile.ToHash160()], calldata, 100000)
	require.Error(t, err)
	require.NotErrorIs(t, err, ctypes.ErrBlockedAccount)
	require.ErrorIs(t, err, ctypes.ErrNotFound, "with fork off the call reaches tokenStatus, i.e. past the blacklist gate")
}

// TestGapB3_SelfdestructBeneficiary 场景 B3：
// 黑名单合约 SELFDESTRUCT，把余额打给受益人。
// opSuicide → AddBalance → statedb.Transfer(sender=合约)，Transfer 的 sender 检查拦住。
// 这里直接以黑名单合约为 Call 目标会先被 Call 的 target 检查拦下，
// 为了单独验证 Transfer 这道闸，用 DELEGATECALL 绕过 Call 检查（见 B3b），
// 本用例只固定"经由 CALL 唤醒"这条完整链路的最终结果。
func TestGapB3_SelfdestructBeneficiary(t *testing.T) {
	user, relay, relay2, blockedC := addr(0x13), addr(0xb3), addr(0xb6), addr(0xa3)
	env := newGapEnv(t, []string{blockedC.String()})
	env.assertEnvelopeClean(user, relay, relay2)

	const stash = int64(500_000)
	env.fund(blockedC, stash)
	env.deploy(relay, user, buildValueCallCode(blockedC, 0))
	env.deploy(relay2, user, buildValueCallCode(blockedC, 0))
	env.deploy(blockedC, user, buildSelfdestructCode(user))

	_, _, _, err := env.evm.Call(AccountRef(user), relay, nil, 5_000_000, 0)
	require.NoError(t, err)
	require.Equal(t, stash, env.balance(blockedC))
	require.Zero(t, env.balance(user))

	env.cfg.SetFork(ctypes.ForkAccountBlacklist, ctypes.MaxHeight)
	_, _, _, err = env.evm.Call(AccountRef(user), relay2, nil, 5_000_000, 0)
	require.NoError(t, err)
	require.Zero(t, env.balance(blockedC), "with fork off SELFDESTRUCT pays the beneficiary")
	require.Equal(t, stash, env.balance(user))
}

// TestGapB3b_DelegatecallSelfdestructDrainsCodeAddr 场景 B3 的变体，
// 也是 Call/Create 检查覆盖不到、只有 statedb.Transfer 能拦的路径：
//
//	干净用户 U → 干净合约 P → DELEGATECALL → 黑名单合约 A 的代码执行 SELFDESTRUCT
//
// evm.DelegateCall 没有黑名单检查（DELEGATECALL 借代码在调用者上下文执行，
// 通常不构成打出）。但 opSuicide 的付款方取 contract.CodeAddr（= A），
// 金额取 contract.Address()（= P）的余额：攻击者给 P 充值 X，即可让 A 向受益人付出 X。
// 这使 statedb.Transfer 的 sender 检查成为真正的最后一道闸，而非冗余。
func TestGapB3b_DelegatecallSelfdestructDrainsCodeAddr(t *testing.T) {
	user, proxy, proxy2, blockedC := addr(0x14), addr(0xb4), addr(0xb5), addr(0xa4)
	env := newGapEnv(t, []string{blockedC.String()})
	env.assertEnvelopeClean(user, proxy, proxy2)

	const stash = int64(300_000)
	env.fund(blockedC, stash)
	env.fund(proxy, stash)
	env.fund(proxy2, stash)
	env.deploy(proxy, user, buildDelegateCallCode(blockedC))
	env.deploy(proxy2, user, buildDelegateCallCode(blockedC))
	env.deploy(blockedC, user, buildSelfdestructCode(user))

	_, _, _, err := env.evm.Call(AccountRef(user), proxy, nil, 5_000_000, 0)
	require.NoError(t, err)
	require.Equal(t, stash, env.balance(blockedC), "delegatecall+selfdestruct must not drain the blocked code address")
	require.Zero(t, env.balance(user))

	// 对照：fork 关闭时 A 的余额确实经由 CodeAddr 被打出（proxy 已自毁，换 proxy2 触发）。
	env.cfg.SetFork(ctypes.ForkAccountBlacklist, ctypes.MaxHeight)
	_, _, _, err = env.evm.Call(AccountRef(user), proxy2, nil, 5_000_000, 0)
	require.NoError(t, err)
	require.Zero(t, env.balance(blockedC), "with fork off, opSuicide pays from CodeAddr (=A)")
	require.Equal(t, stash, env.balance(user))
	require.Equal(t, stash, env.balance(proxy2), "the proxy itself is untouched: the payer is A, not P")
}

// TestGapB4_BlockedContractInnerValueCall 场景 B4：
// 黑名单合约 A 被 Call 拦下之前的"最后一道闸"——直接以 A 为 caller 发起带 value 的内部转账，
// 模拟任何绕过 Call target 检查（例如未来新增的调用路径）的情形，
// statedb.Transfer 的 sender 检查仍然拒绝。
func TestGapB4_BlockedContractInnerValueCall(t *testing.T) {
	user, blockedC := addr(0x15), addr(0xa5)
	env := newGapEnv(t, []string{blockedC.String()})

	const stash = int64(200_000)
	env.fund(blockedC, stash)

	require.False(t, env.mdb.Transfer(blockedC.String(), user.String(), uint64(stash)))
	require.Equal(t, stash, env.balance(blockedC))
	require.Zero(t, env.balance(user))

	// Transfer 对"打入"也拒绝（chain33 原始设计：禁止收发），保留但不作为主要目标。
	env.fund(user, stash)
	require.False(t, env.mdb.Transfer(user.String(), blockedC.String(), uint64(stash)))
	require.Equal(t, stash, env.balance(user))
}

// TestGap_ForkGateBlocksNothingBeforeHeight 验证 fork 门控：
// 未到 ForkAccountBlacklist 高度时，statedb / runtime 两层都不得改变执行结果，
// 否则会与未升级节点分链。
func TestGap_ForkGateBlocksNothingBeforeHeight(t *testing.T) {
	user, blockedC := addr(0x16), addr(0xa6)
	env := newGapEnv(t, []string{blockedC.String()})
	env.cfg.SetFork(ctypes.ForkAccountBlacklist, gapTestHeight+1)

	const stash = int64(100_000)
	env.fund(blockedC, stash)
	require.True(t, env.mdb.CanTransfer(blockedC.String(), uint64(stash)))
	require.True(t, env.mdb.Transfer(blockedC.String(), user.String(), uint64(stash)))
	require.Equal(t, stash, env.balance(user))

	require.NoError(t, checkBlockedAccount(env.evm, blockedC, user))
}

// TestGap_ParaLengthAsymmetry 记录 chain33 Para 维度与 EVM 地址解析之间的不对称：
// vmcommon.Address.SetBytes 对超长输入取后 20 字节，chain33 IsBlockedAccountRaw 要求精确 20 字节。
// 用 32 字节左填充的 Para 可绕过 chain33 的 Para 维度，但只能"打入"、不能"打出"，
// 由 innerExec 的 receiver 检查兜底。此处固定该行为，防止两侧被无意改成不一致。
func TestGap_ParaLengthAsymmetry(t *testing.T) {
	blockedC := addr(0xa7)
	restore := ctypes.SetBlockedAccountsForTest([]string{blockedC.String()})
	t.Cleanup(restore)

	padded := make([]byte, 32)
	copy(padded[12:], blockedC.Bytes())

	require.False(t, ctypes.IsBlockedAccountRaw(padded), "chain33 is blind to 32-byte padded Para")
	require.Equal(t, blockedC, vmcommon.BytesToAddress(padded), "EVM resolves the same bytes to the blocked address")
	require.True(t, ctypes.IsBlockedAccount(vmcommon.BytesToAddress(padded).String()),
		"so the EVM-side receiver check catches what chain33 misses")
}
