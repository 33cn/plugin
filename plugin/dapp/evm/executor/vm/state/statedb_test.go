package state

import (
	"testing"

	ctypes "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
)

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
