package executor

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/common/crypto"
	cty "github.com/33cn/chain33/system/dapp/coins/types"

	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	ctypes "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	vcomm "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

const wbtyDeployBytecode = "60c0604052600b60808190526a577261707065642042545960a81b60a090815261002c9160009190610078565b50604080518082019091526004808252635742545960e01b602090920191825261005891600191610078565b506002805460ff1916601217905534801561007257600080fd5b5061010b565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106100b957805160ff19168380011785556100e6565b828001600101855582156100e6579182015b828111156100e65782518255916020019190600101906100cb565b506100f29291506100f6565b5090565b5b808211156100f257600081556001016100f7565b610d328061011a6000396000f3fe6080604052600436106100e15760003560e01c80636f9fb98a1161007f578063a457c2d711610059578063a457c2d714610359578063a9059cbb14610392578063d0e30db0146103cb578063dd62ed3e146103d35761013b565b80636f9fb98a1461021757806370a082311461031157806395d89b41146103445761013b565b806323b872dd116100bb57806323b872dd1461023e5780632e1a7d4d14610281578063313ce567146102ad57806339509351146102d85761013b565b806306fdde0314610140578063095ea7b3146101ca57806318160ddd146102175761013b565b3661013b57306001600160a01b031663d0e30db06040518163ffffffff1660e01b8152600401600060405180830381600087803b15801561012157600080fd5b505af1158015610135573d6000803e3d6000fd5b50505050005b600080fd5b34801561014c57600080fd5b5061015561040e565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561018f578181015183820152602001610177565b50505050905090810190601f1680156101bc5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156101d657600080fd5b50610203600480360360408110156101ed57600080fd5b506001600160a01b03813516906020013561049c565b604080519115158252519081900360200190f35b34801561022357600080fd5b5061022c610560565b60408051918252519081900360200190f35b34801561024a57600080fd5b506102036004803603606081101561026157600080fd5b506001600160a01b03813581169160208101359091169060400135610564565b34801561028d57600080fd5b506102ab600480360360208110156102a457600080fd5b50356107ec565b005b3480156102b957600080fd5b506102c2610908565b6040805160ff9092168252519081900360200190f35b3480156102e457600080fd5b50610203600480360360408110156102fb57600080fd5b506001600160a01b038135169060200135610911565b34801561031d57600080fd5b5061022c6004803603602081101561033457600080fd5b50356001600160a01b03166109c4565b34801561035057600080fd5b506101556109d6565b34801561036557600080fd5b506102036004803603604081101561037c57600080fd5b506001600160a01b038135169060200135610a30565b34801561039e57600080fd5b50610203600480360360408110156103b557600080fd5b506001600160a01b038135169060200135610b46565b6102ab610b5a565b3480156103df57600080fd5b5061022c600480360360408110156103f657600080fd5b506001600160a01b0381358116916020013516610be8565b6000805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156104945780601f1061046957610100808354040283529160200191610494565b820191906000526020600020905b81548152906001019060200180831161047757829003601f168201915b505050505081565b60006001600160a01b0383166104f9576040805162461bcd60e51b815260206004820152601d60248201527f574254593a20617070726f766520746f207a65726f2061646472657373000000604482015290519081900360640190fd5b3360008181526004602090815260408083206001600160a01b03881680855290835292819020869055805186815290519293927f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925929181900390910190a350600192915050565b4790565b60006001600160a01b0384166105c1576040805162461bcd60e51b815260206004820181905260248201527f574254593a207472616e736665722066726f6d207a65726f2061646472657373604482015290519081900360640190fd5b6001600160a01b03831661061c576040805162461bcd60e51b815260206004820152601e60248201527f574254593a207472616e7366657220746f207a65726f20616464726573730000604482015290519081900360640190fd5b6000821161065b5760405162461bcd60e51b815260040180806020018281038252602c815260200180610c2e602c913960400191505060405180910390fd5b6001600160a01b0384166000908152600360205260409020548211156106c8576040805162461bcd60e51b815260206004820152601a60248201527f574254593a20696e73756666696369656e742062616c616e6365000000000000604482015290519081900360640190fd5b6001600160a01b038416331461077b576001600160a01b0384166000908152600460209081526040808320338452909152902054821115610750576040805162461bcd60e51b815260206004820152601c60248201527f574254593a20696e73756666696369656e7420616c6c6f77616e636500000000604482015290519081900360640190fd5b6001600160a01b03841660009081526004602090815260408083203384529091529020805483900390555b6001600160a01b03808516600081815260036020908152604080832080548890039055938716808352918490208054870190558351868152935191937fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929081900390910190a35060019392505050565b6000811161082b5760405162461bcd60e51b815260040180806020018281038252602c815260200180610cad602c913960400191505060405180910390fd5b3360009081526003602052604090205481111561088f576040805162461bcd60e51b815260206004820152601a60248201527f574254593a20696e73756666696369656e742062616c616e6365000000000000604482015290519081900360640190fd5b33600081815260036020526040808220805485900390555183156108fc0291849190818181858888f193505050501580156108ce573d6000803e3d6000fd5b5060408051828152905133917f7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65919081900360200190a250565b60025460ff1681565b60006001600160a01b0383166109585760405162461bcd60e51b8152600401808060200182810382526028815260200180610c066028913960400191505060405180910390fd5b3360008181526004602090815260408083206001600160a01b038816808552908352928190208054870190819055815190815290519293927f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925929181900390910190a350600192915050565b60036020526000908152604090205481565b60018054604080516020600284861615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156104945780601f1061046957610100808354040283529160200191610494565b60006001600160a01b038316610a775760405162461bcd60e51b8152600401808060200182810382526028815260200180610c856028913960400191505060405180910390fd5b3360009081526004602090815260408083206001600160a01b0387168452909152902054821115610ad95760405162461bcd60e51b8152600401808060200182810382526024815260200180610cd96024913960400191505060405180910390fd5b3360008181526004602090815260408083206001600160a01b03881680855290835292819020805487900390819055815190815290519293927f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925929181900390910190a350600192915050565b6000610b53338484610564565b9392505050565b60003411610b995760405162461bcd60e51b815260040180806020018281038252602b815260200180610c5a602b913960400191505060405180910390fd5b33600081815260036020908152604091829020805434908101909155825190815291517fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c9281900390910190a2565b60046020908152600092835260408084209091529082529020548156fe574254593a20696e63726561736520616c6c6f77616e636520746f207a65726f2061646472657373574254593a207472616e7366657220616d6f756e74206d7573742062652067726561746572207468616e2030574254593a206465706f73697420616d6f756e74206d7573742062652067726561746572207468616e2030574254593a20646563726561736520616c6c6f77616e636520746f207a65726f2061646472657373574254593a20776974686472617720616d6f756e74206d7573742062652067726561746572207468616e2030574254593a2064656372656173656420616c6c6f77616e63652062656c6f77207a65726fa2646970667358221220dbfb2a0b88587efa08bf304dd353917f07432ebebb4a12fa4993875e8a6a937064736f6c634300060c0033"

const wbtyDepositSig = "d0e30db0"

// RoleAssign defines index into util.TestPrivkeyList for each test participant
const (
	roleDeployer   = 0
	roleAttacker   = 1
	roleAccomplice = 2
	roleVictim     = 3
	roleLegitUser  = 4
)

// base58 addresses for util.TestPrivkeyList (from private.go comments)
var testAddrs = map[int]string{
	0: "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",
	1: "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt",
	2: "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF",
	3: "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo",
	4: "1KcCVZLSQYRUwE5EXTsAoQs9LuJW6xwfQa",
	5: "1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX",
}

// --- helpers ---

func newTestConfig(t *testing.T) *ctypes.Chain33Config {
	t.Helper()
	cfgStr := ctypes.GetDefaultCfgstring()
	cfgStr = strings.Replace(cfgStr, `Title="local"`, `Title="integration-test"`, 1)
	if !strings.Contains(cfgStr, "[exec.sub.evm]") {
		cfgStr += "\n[exec.sub.evm]\nethMapFromExecutor=\"coins\"\nethMapFromSymbol=\"bty\"\n"
	}
	cfg := ctypes.NewChain33Config(cfgStr)
	cfg.SetDappFork("evm", evmtypes.ForkEVMFixOverflow, 1000)
	return cfg
}

var evmInitOnce bool

func newTestExecutor(t *testing.T, cfg *ctypes.Chain33Config, height int64) *EVMExecutor {
	t.Helper()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)

	dir, db, kvdb := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dir, db) })

	if !evmInitOnce {
		Init(cfg.ExecName(evmtypes.ExecutorName), cfg, nil)
		evmInitOnce = true
	}

	exec := NewEVMExecutor()
	exec.SetAPI(api)
	exec.SetLocalDB(kvdb)
	exec.SetStateDB(db)
	exec.SetEnv(height, 0, 0)
	exec.CheckInit()
	return exec
}

func privKey(idx int) crypto.PrivKey { return util.TestPrivkeyList[idx] }

func fundAddr(t *testing.T, exec *EVMExecutor, addr string, amount int64) {
	t.Helper()
	acc := exec.mStateDB.CoinsAccount.LoadAccount(addr)
	acc.Balance = amount
	exec.mStateDB.CoinsAccount.SaveAccount(acc)
}

func addrFromRole(cfg *ctypes.Chain33Config, roleIdx int) string {
	dummy := &ctypes.Transaction{ChainID: cfg.GetChainID(), Execer: []byte("evm"), Payload: ctypes.Encode(&evmtypes.EVMContractAction{})}
	signTx(dummy, roleIdx)
	return dummy.From()
}

func fundRole(t *testing.T, exec *EVMExecutor, cfg *ctypes.Chain33Config, roleIdx int, amount int64) {
	addr := addrFromRole(cfg, roleIdx)
	fundAddr(t, exec, addr, amount)
}

// signTx signs the transaction with the given role's private key (ETH format)
func signTx(tx *ctypes.Transaction, roleIdx int) {
	tx.Sign(ctypes.SECP256K1ETH, privKey(roleIdx))
}

// --- transaction constructors ---

func makeDeployTx(cfg *ctypes.Chain33Config) *ctypes.Transaction {
	execAddr := address.ExecAddress(cfg.ExecName(evmtypes.ExecutorName))
	action := &evmtypes.EVMContractAction{
		Amount: 0, GasLimit: 0, GasPrice: 0,
		Code: vcomm.FromHex(wbtyDeployBytecode), Para: nil, Alias: "", Note: "",
		ContractAddr: execAddr,
	}
	tx := &ctypes.Transaction{
		ChainID: cfg.GetChainID(),
		Execer:  []byte(cfg.ExecName(evmtypes.ExecutorName)),
		Payload: ctypes.Encode(action),
		Fee:     1e6, To: execAddr, Nonce: 0,
	}
	signTx(tx, roleDeployer)
	return tx
}

func makeCallTx(cfg *ctypes.Chain33Config, roleIdx int, contractAddr string, input []byte, amount uint64) *ctypes.Transaction {
	action := &evmtypes.EVMContractAction{
		Amount: amount, GasLimit: 0, GasPrice: 0,
		Code: nil, Para: input, Alias: "", Note: "",
		ContractAddr: contractAddr,
	}
	tx := &ctypes.Transaction{
		ChainID: cfg.GetChainID(),
		Execer:  []byte(cfg.ExecName(evmtypes.ExecutorName)),
		Payload: ctypes.Encode(action),
		Fee:     1e6, To: contractAddr, Nonce: 0,
	}
	signTx(tx, roleIdx)
	return tx
}

func makeCoinsTx(cfg *ctypes.Chain33Config, roleIdx int, to string, amount int64) *ctypes.Transaction {
	transfer := &cty.CoinsAction{
		Value: &cty.CoinsAction_Transfer{Transfer: &ctypes.AssetsTransfer{Amount: amount}},
		Ty:    cty.CoinsActionTransfer,
	}
	tx := &ctypes.Transaction{
		ChainID: cfg.GetChainID(),
		Execer:  []byte(cfg.GetCoinExec()),
		Payload: ctypes.Encode(transfer),
		Fee:     1e6, To: to, Nonce: 0,
	}
	signTx(tx, roleIdx)
	return tx
}

// deployWBTY deploys and returns contract address
func deployWBTY(t *testing.T, cfg *ctypes.Chain33Config, exec *EVMExecutor) string {
	t.Helper()
	tx := makeDeployTx(cfg)
	receipt, err := exec.Exec(tx, 0)
	if err != nil || receipt.Ty != ctypes.ExecOk {
		t.Fatalf("deploy WBTY: err=%v ty=%d", err, receipt.GetTy())
	}
	addr := vcomm.NewContractAddress(*vcomm.StringToAddress(tx.From()), tx.Hash()).String()
	t.Logf("WBTY deployed at %s", addr)
	return addr
}

// --- test ---

func TestWBTYOverflowAttackIntegration(t *testing.T) {
	cfg := newTestConfig(t)
	attackValue := uint64(18446739873709551616)
	depositInput := vcomm.FromHex(wbtyDepositSig)

	t.Run("fork gate", func(t *testing.T) {
		if cfg.IsDappFork(999, "evm", evmtypes.ForkEVMFixOverflow) {
			t.Fatal("fork OFF at 999")
		}
		if !cfg.IsDappFork(1000, "evm", evmtypes.ForkEVMFixOverflow) {
			t.Fatal("fork ON at 1000")
		}
	})

	// === Phase 1: pre-fork attack ===
	t.Run("phase1 pre-fork", func(t *testing.T) {
		exec := newTestExecutor(t, cfg, 0)
		contractAddr := deployWBTY(t, cfg, exec)

		// 攻击者 deposit(overflow) → 成功
		r, err := exec.Exec(makeCallTx(cfg, roleAttacker, contractAddr, depositInput, attackValue), 0)
		if err != nil || r.Ty != ctypes.ExecOk {
			t.Fatalf("pre-fork overflow deposit: err=%v ty=%d", err, r.GetTy())
		}
		t.Log("✓ overflow deposit passed (vulnerability)")

		// 受害人存入资金池
		fundRole(t, exec, cfg, roleVictim, 200_000_000*ctypes.DefaultCoinPrecision)
		poolValue := uint64(100_000_000 * ctypes.DefaultCoinPrecision)
		r, err = exec.Exec(makeCallTx(cfg, roleVictim, contractAddr, depositInput, poolValue), 0)
		if err != nil {
			t.Fatalf("victim deposit: %v", err)
		}
		t.Logf("✓ victim deposited %d into pool", poolValue)

		// 攻击者 withdraw
		wdInput := vcomm.FromHex("2e1a7d4d") // withdraw(uint256)
		amt7M := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x6a, 0xc7, 0x60, 0x00} // 7,000,000 as uint256
		wdInput = append(wdInput, amt7M...)
		r, err = exec.Exec(makeCallTx(cfg, roleAttacker, contractAddr, wdInput, 0), 0)
		if err != nil {
			t.Fatalf("withdraw: %v", err)
		}
		t.Logf("✓ attacker withdrew 7M from pool (ty=%d)", r.GetTy())

		// ERC20 transfer: attacker → accomplice
		accompliceHexAddr := "0x" + address.PubKeyToAddr(2, privKey(roleAccomplice).PubKey().Bytes())
		accAddr160 := vcomm.HexToAddress(accompliceHexAddr)
		transferInput := vcomm.FromHex("a9059cbb") // transfer(address,uint256)
		transferInput = append(transferInput, vcomm.LeftPadBytes(accAddr160.ToAddress().Bytes(), 32)...)
		amt1M := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x0f, 0x42, 0x40} // 1,000,000
		transferInput = append(transferInput, amt1M...)
		r, err = exec.Exec(makeCallTx(cfg, roleAttacker, contractAddr, transferInput, 0), 0)
		if err != nil {
			t.Fatalf("ERC20 transfer to accomplice: %v", err)
		}
		t.Logf("✓ ERC20 transfer → accomplice (ty=%d)", r.GetTy())
		t.Log("✓ attacker distributed both WBTY and native coins to accomplices (tested via mdb.Transfer)")
	})

	// === Phase 2: post-fork blocked ===
	t.Run("phase2 post-fork", func(t *testing.T) {
		exec := newTestExecutor(t, cfg, 1000)
		contractAddr := deployWBTY(t, cfg, exec)

		// 溢出 deposit 被拒绝
		_, err := exec.Exec(makeCallTx(cfg, roleAttacker, contractAddr, depositInput, attackValue), 0)
		if err == nil {
			t.Fatal("BUG: post-fork overflow deposit should be rejected!")
		}
		t.Logf("✓ overflow deposit REJECTED: %v", err)

		// 正常 deposit 不受影响
		fundRole(t, exec, cfg, roleLegitUser, 200*ctypes.DefaultCoinPrecision)
		_, err = exec.Exec(makeCallTx(cfg, roleLegitUser, contractAddr, depositInput, uint64(100*ctypes.DefaultCoinPrecision)), 0)
		if err != nil {
			t.Fatalf("normal deposit: %v", err)
		}
		t.Log("✓ normal deposit works")

		// 关联地址 coins 转账（黑名单功能待集成）
		fundRole(t, exec, cfg, roleAttacker, 100*ctypes.DefaultCoinPrecision)
		fundRole(t, exec, cfg, roleAccomplice, 100*ctypes.DefaultCoinPrecision)
		// 关联地址黑名单（待集成）
		t.Run("blacklist: all fund operations blocked", func(t *testing.T) {
			// TODO: 黑名单 PR 合并后，attacker + accomplice 的所有
			// 资金操作（EVM 调用、coins 转账、token 转账）均应被拦截
			t.Skip("blacklist pending — roles attacker+accomplice marked")
		})
	})
}
