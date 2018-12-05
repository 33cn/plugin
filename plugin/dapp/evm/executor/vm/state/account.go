// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

// ContractAccount 合约账户对象
type ContractAccount struct {
	mdb *MemoryStateDB

	// Addr 合约代码地址
	Addr string

	// Data 合约固定数据
	Data evmtypes.EVMContractData

	// State 合约状态数据
	State evmtypes.EVMContractState

	// 当前的状态数据缓存
	stateCache map[string]common.Hash
}

// NewContractAccount 创建一个新的合约对象
// 注意，此时合约对象有可能已经存在也有可能不存在
// 需要通过LoadContract进行判断
func NewContractAccount(addr string, db *MemoryStateDB) *ContractAccount {
	if len(addr) == 0 || db == nil {
		log15.Error("NewContractAccount error, something is missing", "contract addr", addr, "db", db)
		return nil
	}
	ca := &ContractAccount{Addr: addr, mdb: db}
	ca.State.Storage = make(map[string][]byte)
	ca.stateCache = make(map[string]common.Hash)
	return ca
}

// GetState 获取状态数据；
// 获取数据分为两层，一层是从当前的缓存中获取，如果获取不到，再从localdb中获取
func (ca *ContractAccount) GetState(key common.Hash) common.Hash {
	// 从ForkV19开始，状态数据使用单独的KEY存储
	if types.IsDappFork(ca.mdb.blockHeight, "evm", evmtypes.ForkEVMState) {
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
func (ca *ContractAccount) SetState(key, value common.Hash) {
	ca.mdb.addChange(storageChange{
		baseChange: baseChange{},
		account:    ca.Addr,
		key:        key,
		prevalue:   ca.GetState(key),
	})
	if types.IsDappFork(ca.mdb.blockHeight, "evm", evmtypes.ForkEVMState) {
		ca.stateCache[key.Hex()] = value
		//需要设置到localdb中，以免同一个区块中同一个合约多次调用时，状态数据丢失
		keyStr := getStateItemKey(ca.Addr, key.Hex())
		ca.mdb.LocalDB.Set([]byte(keyStr), value.Bytes())
	} else {
		ca.State.GetStorage()[key.Hex()] = value.Bytes()
		ca.updateStorageHash()
	}
}

// TransferState 从原有的存储在一个对象，将状态数据分散存储到多个KEY，保证合约可以支撑大量状态数据
func (ca *ContractAccount) TransferState() {
	if len(ca.State.Storage) > 0 {
		storage := ca.State.Storage
		// 为了保证不会造成新、旧数据并存的情况，需要将旧的状态数据清空
		ca.State.Storage = make(map[string][]byte)
		ca.State.StorageHash = common.ToHash([]byte{}).Bytes()

		// 从旧的区块迁移状态数据到新的区块，模拟状态数据变更的操作
		for key, value := range storage {
			ca.SetState(common.BytesToHash(common.FromHex(key)), common.BytesToHash(value))
		}
		// 更新本合约的状态数据（删除旧的map存储信息）
		ca.mdb.UpdateState(ca.Addr)
		return
	}
}

func (ca *ContractAccount) updateStorageHash() {
	// 从ForkV20开始，状态数据使用单独KEY存储
	if types.IsDappFork(ca.mdb.blockHeight, "evm", evmtypes.ForkEVMState) {
		return
	}
	var state = &evmtypes.EVMContractState{Suicided: ca.State.Suicided, Nonce: ca.State.Nonce}
	state.Storage = make(map[string][]byte)
	for k, v := range ca.State.GetStorage() {
		state.Storage[k] = v
	}
	ret, err := proto.Marshal(state)
	if err != nil {
		log15.Error("marshal contract state data error", "error", err)
		return
	}

	ca.State.StorageHash = common.ToHash(ret).Bytes()
}

// 从外部恢复合约数据
func (ca *ContractAccount) resotreData(data []byte) {
	var content evmtypes.EVMContractData
	err := proto.Unmarshal(data, &content)
	if err != nil {
		log15.Error("read contract data error", ca.Addr)
		return
	}

	ca.Data = content
}

// 从外部恢复合约状态
func (ca *ContractAccount) resotreState(data []byte) {
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

// LoadContract 从数据库中加载合约信息（在只有合约地址的情况下）
func (ca *ContractAccount) LoadContract(db db.KV) {
	// 加载代码数据
	data, err := db.Get(ca.GetDataKey())
	if err != nil {
		return
	}
	ca.resotreData(data)

	// 加载状态数据
	data, err = db.Get(ca.GetStateKey())
	if err != nil {
		return
	}
	ca.resotreState(data)
}

// SetCode 设置合约二进制代码
// 会同步生成代码哈希
func (ca *ContractAccount) SetCode(code []byte) {
	prevcode := ca.Data.GetCode()
	ca.mdb.addChange(codeChange{
		baseChange: baseChange{},
		account:    ca.Addr,
		prevhash:   ca.Data.GetCodeHash(),
		prevcode:   prevcode,
	})
	ca.Data.Code = code
	ca.Data.CodeHash = common.ToHash(code).Bytes()
}

// SetAbi 设置合约绑定的ABI数据
func (ca *ContractAccount) SetAbi(abi string) {
	if types.IsDappFork(ca.mdb.GetBlockHeight(), "evm", evmtypes.ForkEVMABI) {
		ca.mdb.addChange(abiChange{
			baseChange: baseChange{},
			account:    ca.Addr,
			prevabi:    ca.Data.Abi,
		})
		ca.Data.Abi = abi
	}
}

// SetCreator 设置创建者
func (ca *ContractAccount) SetCreator(creator string) {
	if len(creator) == 0 {
		log15.Error("SetCreator error", "creator", creator)
		return
	}
	ca.Data.Creator = creator
}

// SetExecName 设置合约名称
func (ca *ContractAccount) SetExecName(execName string) {
	if len(execName) == 0 {
		log15.Error("SetExecName error", "execName", execName)
		return
	}
	ca.Data.Name = execName
}

// SetAliasName 设置合约别名
func (ca *ContractAccount) SetAliasName(alias string) {
	if len(alias) == 0 {
		log15.Error("SetAliasName error", "aliasName", alias)
		return
	}
	ca.Data.Alias = alias
}

// GetAliasName 获取合约别名
func (ca *ContractAccount) GetAliasName() string {
	return ca.Data.Alias
}

// GetCreator 获取创建者
func (ca *ContractAccount) GetCreator() string {
	return ca.Data.Creator
}

// GetExecName 获取合约明名称
func (ca *ContractAccount) GetExecName() string {
	return ca.Data.Name
}

// GetDataKV 合约固定数据，包含合约代码，以及代码哈希
func (ca *ContractAccount) GetDataKV() (kvSet []*types.KeyValue) {
	ca.Data.Addr = ca.Addr
	datas, err := proto.Marshal(&ca.Data)
	if err != nil {
		log15.Error("marshal contract data error!", "addr", ca.Addr, "error", err)
		return
	}
	kvSet = append(kvSet, &types.KeyValue{Key: ca.GetDataKey(), Value: datas})
	return
}

// GetStateKV 获取合约状态数据，包含nonce、是否自杀、存储哈希、存储数据
func (ca *ContractAccount) GetStateKV() (kvSet []*types.KeyValue) {
	datas, err := proto.Marshal(&ca.State)
	if err != nil {
		log15.Error("marshal contract state error!", "addr", ca.Addr, "error", err)
		return
	}
	kvSet = append(kvSet, &types.KeyValue{Key: ca.GetStateKey(), Value: datas})
	return
}

// BuildDataLog 构建变更日志
func (ca *ContractAccount) BuildDataLog() (log *types.ReceiptLog) {
	datas, err := proto.Marshal(&ca.Data)
	if err != nil {
		log15.Error("marshal contract data error!", "addr", ca.Addr, "error", err)
		return
	}
	return &types.ReceiptLog{Ty: evmtypes.TyLogContractData, Log: datas}
}

// BuildStateLog 构建变更日志
func (ca *ContractAccount) BuildStateLog() (log *types.ReceiptLog) {
	datas, err := proto.Marshal(&ca.State)
	if err != nil {
		log15.Error("marshal contract state log error!", "addr", ca.Addr, "error", err)
		return
	}

	return &types.ReceiptLog{Ty: evmtypes.TyLogContractState, Log: datas}
}

// GetDataKey 获取数据KEY
func (ca *ContractAccount) GetDataKey() []byte {
	return []byte("mavl-" + evmtypes.ExecutorName + "-data: " + ca.Addr)
}

// GetStateKey 获取状态key
func (ca *ContractAccount) GetStateKey() []byte {
	return []byte("mavl-" + evmtypes.ExecutorName + "-state: " + ca.Addr)
}

// 这份数据是存在LocalDB中的
func getStateItemKey(addr, key string) string {
	return fmt.Sprintf("LODB-"+evmtypes.ExecutorName+"-state:%v:%v", addr, key)
}

// Suicide 自杀
func (ca *ContractAccount) Suicide() bool {
	ca.State.Suicided = true
	return true
}

// HasSuicided 是否已经自杀
func (ca *ContractAccount) HasSuicided() bool {
	return ca.State.GetSuicided()
}

// Empty 是否为空对象
func (ca *ContractAccount) Empty() bool {
	return ca.Data.GetCodeHash() == nil || len(ca.Data.GetCodeHash()) == 0
}

// SetNonce 设置nonce值
func (ca *ContractAccount) SetNonce(nonce uint64) {
	ca.mdb.addChange(nonceChange{
		baseChange: baseChange{},
		account:    ca.Addr,
		prev:       ca.State.GetNonce(),
	})
	ca.State.Nonce = nonce
}

// GetNonce 获取nonce值
func (ca *ContractAccount) GetNonce() uint64 {
	return ca.State.GetNonce()
}
