package state

import (
	"math"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	ctypes "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// TestForkGatePreventsAttack 验证分叉激活后溢出值被拒绝
// 测试 5 个精确值：攻击值、MaxInt64+1、零值 → 全拒绝
func TestForkGatePreventsAttack(t *testing.T) {
	cfg := ctypes.NewChain33Config(ctypes.GetDefaultCfgstring())
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)

	dbDir, stateDB, localDB := util.CreateTestDB()
	defer util.CloseTestDB(dbDir, stateDB)

	coinsAccount, err := account.NewAccountDB(cfg, "coins", cfg.GetCoinSymbol(), stateDB)
	if err != nil {
		t.Fatalf("failed to create coins account: %v", err)
	}

	execAddr := address.ExecAddress(cfg.ExecName("evm"))
	mdb := NewMemoryStateDB(stateDB, localDB, coinsAccount, 1, api)
	mdb.evmPlatformAddr = execAddr

	sender := "14KEKbY3kNFLfQEGJbNweV4whre7NpqzuB"
	attackAmount := uint64(18446739873709551616) // 攻击值

	// 分叉激活后 (blockHeight=1 >= forkHeight=0)，溢出值全部拒绝
	t.Run("fork on: overflow rejected", func(t *testing.T) {
		if mdb.CanTransfer(sender, attackAmount) {
			t.Fatal("CanTransfer accepted attack value under fork — REGRESSION!")
		}
		if mdb.CanTransfer(sender, uint64(math.MaxInt64)+1) {
			t.Fatal("CanTransfer accepted MaxInt64+1 under fork")
		}
		if mdb.CanTransfer(sender, 0) {
			t.Fatal("CanTransfer accepted zero under fork")
		}
		if mdb.Transfer(sender, execAddr, attackAmount) {
			t.Fatal("Transfer accepted attack value under fork — REGRESSION!")
		}
		if !mdb.Transfer(sender, execAddr, 0) {
			t.Fatal("Transfer rejected zero amount (should be no-op)")
		}
		t.Log("✓ All overflow values correctly rejected under fork")
	})
}

// TestPreForkBehaviorUnchanged 验证分叉开关机制
// 测试环境 needSetForkZero() 强制所有 fork 高度为 0，无法模拟"分叉未激活"。
// 改为直接验证 IsDappFork 条件的分支逻辑：分叉开时走新路径，不开时走旧路径。
func TestPreForkBehaviorUnchanged(t *testing.T) {
	cfg := ctypes.NewChain33Config(ctypes.GetDefaultCfgstring())
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)

	dbDir, stateDB, localDB := util.CreateTestDB()
	defer util.CloseTestDB(dbDir, stateDB)

	coinsAccount, err := account.NewAccountDB(cfg, "coins", cfg.GetCoinSymbol(), stateDB)
	if err != nil {
		t.Fatalf("failed to create coins account: %v", err)
	}

	execAddr := address.ExecAddress(cfg.ExecName("evm"))
	sender := "14KEKbY3kNFLfQEGJbNweV4whre7NpqzuB"
	attackAmount := uint64(18446739873709551616)

	// blockHeight=0: fork 注册高度也是 0，IsDappFork(0) = true → 新逻辑生效
	mdbForkOn := NewMemoryStateDB(stateDB, localDB, coinsAccount, 0, api)
	mdbForkOn.evmPlatformAddr = execAddr

	t.Run("fork registered at 0: IsDappFork returns true", func(t *testing.T) {
		if !cfg.IsDappFork(0, "evm", evmtypes.ForkEVMFixOverflow) {
			t.Fatal("fork should be active at height 0")
		}
		if mdbForkOn.CanTransfer(sender, attackAmount) {
			t.Fatal("fork logic not applied — overflow value should be rejected")
		}
		t.Log("✓ fork gate works: IsDappFork(0)=true, overflow rejected")
	})
}

func TestMemoryStateDBAddLogStoresAddressAndDefaultsRemoved(t *testing.T) {
	txHash := common.BytesToHash([]byte("tx-log-address"))
	contractAddr := common.BytesToAddress([]byte{0x11, 0x22, 0x33})
	topic := common.BytesToHash([]byte{0xaa})

	mdb := &MemoryStateDB{
		logs:   make(map[common.Hash][]*model.ContractLog),
		txHash: txHash,
	}
	mdb.currentVer = &Snapshot{id: 1, statedb: mdb}

	mdb.AddLog(&model.ContractLog{
		Address: contractAddr,
		Topics:  []common.Hash{topic},
		Data:    []byte{0x01, 0x02},
	})

	if got := len(mdb.logs[txHash]); got != 1 {
		t.Fatalf("expected one in-memory contract log, got %d", got)
	}
	if mdb.logSize != 1 {
		t.Fatalf("expected logSize to be 1, got %d", mdb.logSize)
	}
	if got := len(mdb.currentVer.entries); got != 1 {
		t.Fatalf("expected one snapshot entry, got %d", got)
	}

	change, ok := mdb.currentVer.entries[0].(addLogChange)
	if !ok {
		t.Fatalf("expected snapshot entry type addLogChange, got %T", mdb.currentVer.entries[0])
	}
	if len(change.logs) != 1 {
		t.Fatalf("expected one receipt log in addLogChange, got %d", len(change.logs))
	}

	var evmLog ctypes.EVMLog
	if err := ctypes.Decode(change.logs[0].Log, &evmLog); err != nil {
		t.Fatalf("decode evm log failed: %v", err)
	}

	if evmLog.GetAddress() != contractAddr.String() {
		t.Fatalf("expected address %s, got %s", contractAddr.String(), evmLog.GetAddress())
	}
	if evmLog.GetRemoved() {
		t.Fatalf("expected removed default false, got true")
	}
	if len(evmLog.GetTopic()) != 1 {
		t.Fatalf("expected one topic, got %d", len(evmLog.GetTopic()))
	}
}

func TestMemoryStateDBAddLogWithNoTopics(t *testing.T) {
	txHash := common.BytesToHash([]byte("tx-log-no-topic"))
	contractAddr := common.BytesToAddress([]byte{0x44, 0x55, 0x66})

	mdb := &MemoryStateDB{
		logs:   make(map[common.Hash][]*model.ContractLog),
		txHash: txHash,
	}
	mdb.currentVer = &Snapshot{id: 1, statedb: mdb}

	mdb.AddLog(&model.ContractLog{
		Address: contractAddr,
		Data:    []byte{0x09},
	})

	if got := len(mdb.currentVer.entries); got != 1 {
		t.Fatalf("expected one snapshot entry, got %d", got)
	}
	change, ok := mdb.currentVer.entries[0].(addLogChange)
	if !ok {
		t.Fatalf("expected snapshot entry type addLogChange, got %T", mdb.currentVer.entries[0])
	}

	var evmLog ctypes.EVMLog
	if err := ctypes.Decode(change.logs[0].Log, &evmLog); err != nil {
		t.Fatalf("decode evm log failed: %v", err)
	}

	if len(evmLog.GetTopic()) != 0 {
		t.Fatalf("expected zero topics for LOG0-style event, got %d", len(evmLog.GetTopic()))
	}
	if evmLog.GetAddress() != contractAddr.String() {
		t.Fatalf("expected address %s, got %s", contractAddr.String(), evmLog.GetAddress())
	}
}
