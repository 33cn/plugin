package executor

import (
	"bytes"
	"testing"

	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	ctypes "github.com/33cn/chain33/types"
	vcomm "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Factory whose runtime code performs two CREATEs of the same 5-byte init code
// (0x60006000f3: return empty runtime) and returns the two resulting addresses
// as 64 bytes of return data.
//
// runtime:
//
//	PUSH5 0x60006000f3; PUSH1 0x00; MSTORE    ; mem[0..32] = initcode
//	PUSH1 5; PUSH1 27; PUSH1 0; CREATE        ; addr1
//	PUSH1 5; PUSH1 27; PUSH1 0; CREATE        ; addr2
//	PUSH1 0x20; MSTORE                        ; mem[32..64] = addr2
//	PUSH1 0x00; MSTORE                        ; mem[0..32] = addr1
//	PUSH1 0x40; PUSH1 0x00; RETURN
const factoryCreateTwiceRuntime = "6460006000f3" + "600052" + "6005601b6000f0" + "6005601b6000f0" + "602052" + "600052" + "60406000f3"

// deploy wrapper: CODECOPY runtime (34 bytes at offset 12) then RETURN it
const factoryCreateTwiceBytecode = "6022600c60003960226000f3" + factoryCreateTwiceRuntime

// Factory whose runtime code performs a single CREATE and returns the resulting
// address as 32 bytes of return data.
const factoryCreateOnceRuntime = "6460006000f3" + "600052" + "6005601b6000f0" + "600052" + "60206000f3"

// deploy wrapper: CODECOPY runtime (24 bytes at offset 12) then RETURN it
const factoryCreateOnceBytecode = "6018600c60003960186000f3" + factoryCreateOnceRuntime

// CREATE2 with a constant salt and constant init code: the address is by design
// the same on both calls, so the second one must fail as a collision (EIP-1014).
const create2TwiceBytecode = "6026600c60003960266000f3" +
	"6460006000f3" + "600052" +
	"60016005601b6000f5" +
	"60016005601b6000f5" +
	"602052" + "600052" + "60406000f3"

// deployContract deploys bytecode and applies the resulting KV set the way the
// block executor does between transactions.
func deployContract(t *testing.T, cfg *ctypes.Chain33Config, exec *EVMExecutor, bytecode string) string {
	t.Helper()
	execAddr := address.ExecAddress(cfg.ExecName(evmtypes.ExecutorName))
	action := &evmtypes.EVMContractAction{Code: vcomm.FromHex(bytecode), ContractAddr: execAddr}
	tx := &ctypes.Transaction{
		ChainID: cfg.GetChainID(),
		Execer:  []byte(cfg.ExecName(evmtypes.ExecutorName)),
		Payload: ctypes.Encode(action),
		Fee:     1e6, To: execAddr, Nonce: 0,
	}
	signTx(tx, roleDeployer)
	resetEvmExecAddress(cfg)
	r, err := exec.Exec(tx, 0)
	if err != nil || r == nil || r.Ty != ctypes.ExecOk {
		t.Fatalf("deploy: err=%v ty=%d", err, r.GetTy())
	}
	addr := vcomm.NewContractAddress(*vcomm.StringToAddress(tx.From()), tx.Hash()).String()
	for _, kv := range r.KV {
		if err := exec.GetStateDB().Set(kv.Key, kv.Value); err != nil {
			t.Fatalf("apply deploy kv: %v", err)
		}
	}
	return addr
}

// callRet runs a call and returns the receipt plus the EVM return data.
func callRet(t *testing.T, cfg *ctypes.Chain33Config, exec *EVMExecutor, addr string) (*ctypes.Receipt, []byte, error) {
	t.Helper()
	r, err := exec.Exec(makeCallTx(cfg, roleDeployer, addr, nil, 0), 0)
	if err != nil || r == nil {
		return r, nil, err
	}
	var ret []byte
	for _, lg := range r.Logs {
		if lg.Ty != evmtypes.TyLogCallContract {
			continue
		}
		var cr evmtypes.ReceiptEVMContract
		if err := ctypes.Decode(lg.Log, &cr); err != nil {
			t.Fatalf("decode receipt: %v", err)
		}
		ret = cr.Ret
	}
	return r, ret, nil
}

// newTestExecutorFrom builds another executor over the same state/local DB, used
// to prove that state written by one transaction is visible to the next one.
func newTestExecutorFrom(t *testing.T, cfg *ctypes.Chain33Config, height int64, src *EVMExecutor) *EVMExecutor {
	t.Helper()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)

	exec := NewEVMExecutor()
	exec.SetAPI(api)
	exec.SetLocalDB(src.GetLocalDB())
	exec.SetStateDB(src.GetStateDB())
	exec.SetEnv(height, 0, 0)
	exec.CheckInit()
	return exec
}

// contractCreateV2ForkHeight is above the execution height (0) used by the
// tests below, so setting it keeps the fork disabled.
const contractCreateV2ForkHeight = 1000

func enableContractCreateV2(cfg *ctypes.Chain33Config) {
	cfg.SetDappFork("evm", evmtypes.ForkEVMContractCreateV2, 0)
}

func disableContractCreateV2(cfg *ctypes.Chain33Config) {
	cfg.SetDappFork("evm", evmtypes.ForkEVMContractCreateV2, contractCreateV2ForkHeight)
}

// resetEvmExecAddress refreshes the package-level exec address caches. They are
// computed once by Init, and other tests in this package switch the global EVM
// address driver (query_test.go), which leaves the caches stale and makes the
// executor treat a deploy as a plain transfer.
func resetEvmExecAddress(cfg *ctypes.Chain33Config) {
	evmExecAddress = vcomm.StringToAddress(address.ExecAddress(cfg.ExecName(evmtypes.ExecutorName))).String()
	evmExecFormatAddress = address.FormatEthAddress(evmExecAddress)
}

// readContractNonce reads the persisted EVM nonce of a contract account.
func readContractNonce(t *testing.T, exec *EVMExecutor, addr string) uint64 {
	t.Helper()
	key := []byte("mavl-" + evmtypes.ExecutorName + "-state: " + addr)
	raw, err := exec.GetStateDB().Get(key)
	if err != nil {
		return 0
	}
	var st evmtypes.EVMContractState
	if err := ctypes.Decode(raw, &st); err != nil {
		t.Fatalf("decode contract state: %v", err)
	}
	return st.GetNonce()
}

// ethCreateAddress computes the ethereum CREATE address for a chain33-format
// contract address at the given nonce.
func ethCreateAddress(t *testing.T, chain33Addr string, nonce uint64) gethcommon.Address {
	t.Helper()
	raw := vcomm.StringToAddress(chain33Addr)
	if raw == nil {
		t.Fatalf("bad contract address %q", chain33Addr)
	}
	return ethcrypto.CreateAddress(gethcommon.BytesToAddress(raw.Bytes()), nonce)
}

func TestInternalCreateAddressForkOff(t *testing.T) {
	cfg := newTestConfig(t)
	disableContractCreateV2(cfg)
	exec := newTestExecutor(t, cfg, 0)

	factoryAddr := deployContract(t, cfg, exec, factoryCreateTwiceBytecode)
	t.Logf("factory deployed at %s", factoryAddr)

	r, ret, err := callRet(t, cfg, exec, factoryAddr)
	if err != nil {
		t.Fatalf("call factory err: %v", err)
	}
	if r == nil || r.Ty != ctypes.ExecOk {
		t.Fatalf("call factory bad receipt: %+v", r)
	}
	if len(ret) != 64 {
		t.Fatalf("unexpected return length %d: %s", len(ret), vcomm.Bytes2Hex(ret))
	}
	addr1 := vcomm.BytesToAddress(ret[:32])
	addr2 := vcomm.BytesToAddress(ret[32:])
	t.Logf("create #1 -> %s", addr1.String())
	t.Logf("create #2 -> %s", addr2.String())

	// fork 关闭时保持旧行为：同一笔交易内两次 CREATE 得到同一个地址。
	if !bytes.Equal(addr1.Bytes(), addr2.Bytes()) {
		t.Fatalf("fork off should keep the legacy behavior, got %s and %s", addr1.String(), addr2.String())
	}
}

func TestInternalCreateAddressForkOn(t *testing.T) {
	cfg := newTestConfig(t)
	enableContractCreateV2(cfg)
	exec := newTestExecutor(t, cfg, 0)

	factoryAddr := deployContract(t, cfg, exec, factoryCreateTwiceBytecode)
	t.Logf("factory deployed at %s", factoryAddr)

	r, ret, err := callRet(t, cfg, exec, factoryAddr)
	if err != nil {
		t.Fatalf("call factory err: %v", err)
	}
	if r == nil || r.Ty != ctypes.ExecOk {
		t.Fatalf("call factory bad receipt: %+v", r)
	}
	if len(ret) != 64 {
		t.Fatalf("unexpected return length %d: %s", len(ret), vcomm.Bytes2Hex(ret))
	}
	addr1 := vcomm.BytesToAddress(ret[:32])
	addr2 := vcomm.BytesToAddress(ret[32:])
	t.Logf("create #1 -> %s", addr1.String())
	t.Logf("create #2 -> %s", addr2.String())

	if addr1 == addr2 {
		t.Fatalf("BUG: both CREATEs in one tx returned the same address %s", addr1.String())
	}
	want1 := ethCreateAddress(t, factoryAddr, 1)
	want2 := ethCreateAddress(t, factoryAddr, 2)
	if !bytes.Equal(addr1.Bytes(), want1.Bytes()) {
		t.Fatalf("create #1 = %s, want %s", addr1.String(), gethcommon.BytesToAddress(want1.Bytes()).Hex())
	}
	if !bytes.Equal(addr2.Bytes(), want2.Bytes()) {
		t.Fatalf("create #2 = %s, want %s", addr2.String(), gethcommon.BytesToAddress(want2.Bytes()).Hex())
	}
	t.Logf("create #1 matches CreateAddress(factory, 1) = %s", want1.Hex())
	t.Logf("create #2 matches CreateAddress(factory, 2) = %s", want2.Hex())
}

func TestCreate2CollisionForkOn(t *testing.T) {
	cfg := newTestConfig(t)
	enableContractCreateV2(cfg)
	exec := newTestExecutor(t, cfg, 0)
	addr := deployContract(t, cfg, exec, create2TwiceBytecode)

	r, ret, err := callRet(t, cfg, exec, addr)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if r == nil || r.Ty != ctypes.ExecOk {
		t.Fatalf("bad receipt: %+v", r)
	}
	if len(ret) != 64 {
		t.Fatalf("unexpected return length %d", len(ret))
	}
	a1 := vcomm.BytesToAddress(ret[:32])
	a2 := vcomm.BytesToAddress(ret[32:])
	t.Logf("create2 #1 -> %s", a1.String())
	t.Logf("create2 #2 -> %s", a2.String())
	if a1 == (vcomm.Address{}) {
		t.Fatalf("first CREATE2 should succeed")
	}
	if a2 != (vcomm.Address{}) {
		t.Fatalf("second CREATE2 to an occupied address returned %s instead of failing", a2.String())
	}
}

func TestCreateNoncePersistsAcrossTransactions(t *testing.T) {
	cfg := newTestConfig(t)
	enableContractCreateV2(cfg)
	exec := newTestExecutor(t, cfg, 0)

	factoryAddr := deployContract(t, cfg, exec, factoryCreateOnceBytecode)
	t.Logf("factory deployed at %s", factoryAddr)

	callAndApply := func(e *EVMExecutor) vcomm.Address {
		r, ret, err := callRet(t, cfg, e, factoryAddr)
		if err != nil {
			t.Fatalf("call factory err: %v", err)
		}
		if r == nil || r.Ty != ctypes.ExecOk {
			t.Fatalf("call factory bad receipt: %+v", r)
		}
		if len(ret) != 32 {
			t.Fatalf("unexpected return length %d", len(ret))
		}
		applyReceiptKV(t, e, r)
		return vcomm.BytesToAddress(ret)
	}

	// Transaction #1: factory nonce goes 1 -> 2, the new contract is
	// CreateAddress(factory, 1).
	addr1 := callAndApply(exec)
	want1 := ethCreateAddress(t, factoryAddr, 1)
	if !bytes.Equal(addr1.Bytes(), want1.Bytes()) {
		t.Fatalf("tx #1 address = %s, want %s", addr1.String(), want1.Hex())
	}
	t.Logf("tx #1 -> %s (nonce 1)", addr1.String())

	// The incremented nonce must be part of the receipt KV written to state.
	var state evmtypes.EVMContractState
	stateKey := []byte("mavl-" + evmtypes.ExecutorName + "-state: " + factoryAddr)
	stateKV, err := exec.GetStateDB().Get(stateKey)
	if err != nil {
		t.Fatalf("read factory state: %v", err)
	}
	if err := ctypes.Decode(stateKV, &state); err != nil {
		t.Fatalf("decode factory state: %v", err)
	}
	if state.GetNonce() != 2 {
		t.Fatalf("factory nonce persisted = %d, want 2", state.GetNonce())
	}
	t.Logf("factory nonce persisted to state: %d", state.GetNonce())

	// Transaction #2 uses a fresh executor over the same DB: the address must
	// come from the persisted nonce (2), not from the in-memory cache.
	exec2 := newTestExecutorFrom(t, cfg, 0, exec)
	addr2 := callAndApply(exec2)
	want2 := ethCreateAddress(t, factoryAddr, 2)
	if !bytes.Equal(addr2.Bytes(), want2.Bytes()) {
		t.Fatalf("tx #2 address = %s, want %s", addr2.String(), want2.Hex())
	}
	t.Logf("tx #2 -> %s (nonce 2)", addr2.String())

	// Transaction #3 continues the sequence.
	exec3 := newTestExecutorFrom(t, cfg, 0, exec2)
	addr3 := callAndApply(exec3)
	want3 := ethCreateAddress(t, factoryAddr, 3)
	if !bytes.Equal(addr3.Bytes(), want3.Bytes()) {
		t.Fatalf("tx #3 address = %s, want %s", addr3.String(), want3.Hex())
	}
	t.Logf("tx #3 -> %s (nonce 3)", addr3.String())
}

// factory whose single CREATE always fails: the init code reverts.
//
//	PUSH5 0x60006000fd; PUSH1 0; MSTORE
//	PUSH1 5; PUSH1 27; PUSH1 0; CREATE; STOP
const factoryCreateRevertingChildRuntime = "6460006000fd" + "600052" + "6005601b6000f0" + "00"

// deploy wrapper: CODECOPY runtime (17 bytes at offset 12) then RETURN it
const factoryCreateRevertingChildBytecode = "6011600c60003960116000f3" + factoryCreateRevertingChildRuntime

// fork off: the address must stay the legacy hash(txHash||origin), and no nonce
// may be written to the contract state.
func TestInternalCreateAddressForkOffLegacy(t *testing.T) {
	cfg := newTestConfig(t)
	disableContractCreateV2(cfg)
	exec := newTestExecutor(t, cfg, 0)

	factoryAddr := deployContract(t, cfg, exec, factoryCreateTwiceBytecode)

	tx := makeCallTx(cfg, roleDeployer, factoryAddr, nil, 0)
	resetEvmExecAddress(cfg)
	r, err := exec.Exec(tx, 0)
	if err != nil || r == nil || r.Ty != ctypes.ExecOk {
		t.Fatalf("call factory: err=%v ty=%d", err, r.GetTy())
	}
	_, ret := receiptRet(t, r)
	if len(ret) != 64 {
		t.Fatalf("unexpected return length %d", len(ret))
	}

	legacy := vcomm.NewContractAddress(*vcomm.StringToAddress(tx.From()), tx.Hash())
	want := vcomm.LeftPadBytes(legacy.Bytes(), 32)
	for i, got := range [][]byte{ret[:32], ret[32:]} {
		if !bytes.Equal(got, want) {
			t.Fatalf("create #%d = %x, want legacy %x", i+1, got, want)
		}
	}
	if nonce := readContractNonce(t, exec, factoryAddr); nonce != 0 {
		t.Fatalf("fork off wrote nonce %d to contract state", nonce)
	}
}

// fork on: a failed CREATE still consumes one nonce, like go-ethereum.
func TestCreateFailureStillIncrementsNonce(t *testing.T) {
	cfg := newTestConfig(t)
	enableContractCreateV2(cfg)
	exec := newTestExecutor(t, cfg, 0)

	factoryAddr := deployContract(t, cfg, exec, factoryCreateRevertingChildBytecode)
	if nonce := readContractNonce(t, exec, factoryAddr); nonce != 1 {
		t.Fatalf("new contract nonce = %d, want 1 (EIP-161)", nonce)
	}

	r, _, err := callRet(t, cfg, exec, factoryAddr)
	if err != nil || r == nil || r.Ty != ctypes.ExecOk {
		t.Fatalf("call factory: err=%v ty=%d", err, r.GetTy())
	}
	applyReceiptKV(t, exec, r)

	if nonce := readContractNonce(t, exec, factoryAddr); nonce != 2 {
		t.Fatalf("factory nonce after failed CREATE = %d, want 2", nonce)
	}
}

// factory that CREATEs one child with value 1 whose init code reverts.
//
//	PUSH5 0x60006000fd; PUSH1 0; MSTORE
//	PUSH1 5; PUSH1 27; PUSH1 1; CREATE; STOP
const factoryCreateWithValueRuntime = "6460006000fd" + "600052" + "6005601b6001f0" + "00"

// deploy wrapper: CODECOPY runtime (17 bytes at offset 12) then RETURN it
const factoryCreateWithValueBytecode = "6011600c60003960116000f3" + factoryCreateWithValueRuntime

// A failed CREATE consumes a nonce; with ForkEVMTransferRevert enabled the value
// attached to the CREATE is rolled back as well.
func TestFailedCreateRollsBackValueWithTransferRevert(t *testing.T) {
	cfg := newTestConfig(t)
	enableContractCreateV2(cfg)
	cfg.SetDappFork("evm", evmtypes.ForkEVMTransferRevert, 0)
	exec := newTestExecutor(t, cfg, 0)

	factoryAddr := deployContract(t, cfg, exec, factoryCreateWithValueBytecode)
	fundAddr(t, exec, factoryAddr, 1000)
	before := exec.mStateDB.GetBalance(factoryAddr)

	r, _, err := callRet(t, cfg, exec, factoryAddr)
	if err != nil || r == nil || r.Ty != ctypes.ExecOk {
		t.Fatalf("call factory: err=%v ty=%d", err, r.GetTy())
	}
	applyReceiptKV(t, exec, r)

	if after := exec.mStateDB.GetBalance(factoryAddr); after != before {
		t.Fatalf("factory balance %d -> %d, value was not rolled back", before, after)
	}
	if nonce := readContractNonce(t, exec, factoryAddr); nonce != 2 {
		t.Fatalf("factory nonce after failed CREATE = %d, want 2", nonce)
	}
	child := ethCreateAddress(t, factoryAddr, 1)
	if bal := exec.mStateDB.GetBalance(vcomm.BytesToAddress(child.Bytes()).String()); bal != 0 {
		t.Fatalf("failed create address holds %d coins", bal)
	}
}

// receiptRet extracts the EVM return data from a receipt.
func receiptRet(t *testing.T, r *ctypes.Receipt) (*evmtypes.ReceiptEVMContract, []byte) {
	t.Helper()
	for _, lg := range r.Logs {
		if lg.Ty != evmtypes.TyLogCallContract {
			continue
		}
		var cr evmtypes.ReceiptEVMContract
		if err := ctypes.Decode(lg.Log, &cr); err != nil {
			t.Fatalf("decode receipt: %v", err)
		}
		return &cr, cr.Ret
	}
	return nil, nil
}
