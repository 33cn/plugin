package state

import (
	"fmt"

	"github.com/33cn/chain33/common/db"

	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/golang/protobuf/proto"
)

// EvmAccount 合约账户对象
type EvmAccount struct {
	mdb *MemoryStateDB

	// Addr 用户地址
	Addr string

	// State 合约状态数据
	State evmtypes.EVMContractState

	// 当前的状态数据缓存
	stateCache map[string]common.Hash
}

// NewEvmAccount 创建一个新的合约对象
// 注意，此时合约对象有可能已经存在也有可能不存在
// 需要通过LoadContract进行判断
func NewEvmAccount(addr string, db *MemoryStateDB) *EvmAccount {
	if len(addr) == 0 || db == nil {
		log15.Error("NewContractAccount error, something is missing", "contract addr", addr, "db", db)
		return nil
	}
	ca := &EvmAccount{Addr: addr, mdb: db}
	ca.State.Storage = make(map[string][]byte)
	ca.stateCache = make(map[string]common.Hash)
	return ca
}

// GetState 获取状态数据；
// 获取数据分为两层，一层是从当前的缓存中获取，如果获取不到，再从localdb中获取
func (ca *EvmAccount) GetState(key common.Hash) common.Hash {
	// 从ForkV19开始，状态数据使用单独的KEY存储
	cfg := ca.mdb.api.GetConfig()
	if cfg.IsDappFork(ca.mdb.blockHeight, "evm", evmtypes.ForkEVMState) {
		if val, ok := ca.stateCache[key.Hex()]; ok {
			return val
		}
		keyStr := getStateItemKey(ca.Addr, key.Hex())
		// 如果缓存中取不到数据，则只能到本地数据库中查询
		val, err := ca.mdb.LocalDB.Get([]byte(keyStr))
		if err != nil {
			log15.Debug("GetState error!", "key", key, "error", err)
			return common.Hash{}
		}
		valHash := common.BytesToHash(val)
		ca.stateCache[key.Hex()] = valHash
		return valHash
	}
	return common.BytesToHash(ca.State.GetStorage()[key.Hex()])
}

// SetState 设置状态数据
func (ca *EvmAccount) SetState(key, value common.Hash) {
	ca.mdb.addChange(storageChange{
		baseChange: baseChange{},
		account:    ca.Addr,
		key:        key,
		prevalue:   ca.GetState(key),
	})
	cfg := ca.mdb.api.GetConfig()
	if cfg.IsDappFork(ca.mdb.blockHeight, "evm", evmtypes.ForkEVMState) {
		ca.stateCache[key.Hex()] = value
		//需要设置到localdb中，以免同一个区块中同一个合约多次调用时，状态数据丢失
		keyStr := getStateItemKey(ca.Addr, key.Hex())
		ca.mdb.LocalDB.Set([]byte(keyStr), value.Bytes())
	} else {
		ca.State.GetStorage()[key.Hex()] = value.Bytes()
		ca.updateStorageHash()
	}
}

func (ca *EvmAccount) updateStorageHash() {
	// 从ForkV20开始，状态数据使用单独KEY存储
	cfg := ca.mdb.api.GetConfig()
	if cfg.IsDappFork(ca.mdb.blockHeight, "evm", evmtypes.ForkEVMState) {
		return
	}
	var state = &evmtypes.EVMContractState{Suicided: ca.State.Suicided, Nonce: ca.State.Nonce}
	state.Storage = make(map[string][]byte)
	for k, v := range ca.State.GetStorage() {
		state.Storage[k] = v
	}
	ret := types.Encode(state)

	ca.State.StorageHash = common.ToHash(ret).Bytes()
}

// LoadContract 从数据库中加载合约信息（在只有合约地址的情况下）
func (ca *EvmAccount) LoadContract(db db.KV) {

	// 加载状态数据
	data, err := db.Get(ca.GetStateKey())
	if err != nil {
		return
	}
	ca.resotreState(data)
}

// 从外部恢复合约状态
func (ca *EvmAccount) resotreState(data []byte) {
	var content evmtypes.EVMContractState
	err := proto.Unmarshal(data, &content)
	if err != nil {
		log15.Error("read contract state error", ca.Addr)
		return
	}
	ca.State = content
	if ca.State.Storage == nil {
		ca.State.Storage = make(map[string][]byte)
	}
}

// GetStateKV 获取合约状态数据，包含nonce、是否自杀、存储哈希、存储数据
func (ca *EvmAccount) GetStateKV() (kvSet []*types.KeyValue) {
	datas := types.Encode(&ca.State)
	kvSet = append(kvSet, &types.KeyValue{Key: ca.GetStateKey(), Value: datas})
	return
}

// GetDataKey 获取数据KEY
func (ca *EvmAccount) GetDataKey() []byte {
	return []byte("mavl-" + evmtypes.ExecutorName + "-data: " + ca.Addr)
}

// GetStateKey 获取状态key
func (ca *EvmAccount) GetStateKey() []byte {
	return []byte("mavl-" + evmtypes.ExecutorName + "-state: " + ca.Addr)
}

// SetNonce 设置nonce值
func (ca *EvmAccount) SetNonce(nonce uint64) {
	fmt.Println("updateStorageHashupdateStorageHashupdateStorageHashupdateStorageHashupdateStorageHash", "SetNonce", nonce)
	ca.mdb.addChange(nonceChange{
		baseChange: baseChange{},
		account:    ca.Addr,
		prev:       ca.State.GetNonce(),
	})
	ca.State.Nonce = nonce

}

// GetNonce 获取nonce值
func (ca *EvmAccount) GetNonce() uint64 {
	fmt.Println("updateStorageHashupdateStorageHashupdateStorageHashupdateStorageHashupdateStorageHash", "GetNonce", "")
	return ca.State.GetNonce()
}
