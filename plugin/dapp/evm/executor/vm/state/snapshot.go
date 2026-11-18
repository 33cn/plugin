// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package state

import (
	"errors"
	"fmt"
	"sort"

	"github.com/33cn/chain33/common/log/log15"
	cexecutor "github.com/33cn/chain33/executor"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// DataChange 数据状态变更接口
// 所有的数据状态变更事件实现此接口，并且封装各自的变更数据以及回滚动作
// 在调用合约时（具体的Tx执行时），会根据操作生成对应的变更对象并缓存下来
// 如果合约执行出错，会按生成顺序的倒序，依次调用变更对象的回滚接口进行数据回滚，并同步删除变更对象缓存
// 如果合约执行成功，会按生成顺序的正序，依次调用变更对象的数据和日志变更记录，回传给区块链
type DataChange interface {
	revert(mdb *MemoryStateDB) error
	getData(mdb *MemoryStateDB) []*types.KeyValue
	getLog(mdb *MemoryStateDB) []*types.ReceiptLog
}

// Snapshot 版本结构，包含版本号以及当前版本包含的变更对象在变更序列中的开始序号
type Snapshot struct {
	id      int
	entries []DataChange
	statedb *MemoryStateDB
}

// GetID 获取ID
func (ver *Snapshot) GetID() int {
	return ver.id
}

// 回滚当前版本
func (ver *Snapshot) revert() error {
	if ver.entries == nil {
		return nil
	}
	// The interface contract requires undoing changes in reverse creation order. Applying two
	// writes to the same slot forward would restore the intermediate value instead of the
	// pre-frame one. Legacy behaviour is kept before ForkEVMTransferRevert.
	if ver.statedb != nil && ver.statedb.isForkActive(evmtypes.ForkEVMTransferRevert) {
		for i := len(ver.entries) - 1; i >= 0; i-- {
			if err := ver.entries[i].revert(ver.statedb); err != nil {
				return err
			}
		}
		return nil
	}
	for _, entry := range ver.entries {
		if err := entry.revert(ver.statedb); err != nil {
			return err
		}
	}
	return nil
}

// 添加变更数据
func (ver *Snapshot) append(entry DataChange) {
	ver.entries = append(ver.entries, entry)
}

// 获取当前版本变更数据
func (ver *Snapshot) getData() (kvSet []*types.KeyValue, logs []*types.ReceiptLog) {
	// 获取中间的数据变更
	dataMap := make(map[string]*types.KeyValue)

	for _, entry := range ver.entries {
		items := entry.getData(ver.statedb)
		logEntry := entry.getLog(ver.statedb)
		if logEntry != nil {
			logs = append(logs, entry.getLog(ver.statedb)...)
		}

		// 执行去重操作
		for _, kv := range items {
			dataMap[string(kv.Key)] = kv
		}
	}

	// 这里也可能会引起数据顺序不一致的问题，需要修改（目前看KV的顺序不会影响哈希计算，但代码最好保证顺序一致）
	names := make([]string, 0, len(dataMap))
	for name := range dataMap {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		kvSet = append(kvSet, dataMap[name])
	}

	return kvSet, logs
}

type (
	// 基础变更对象，用于封装默认操作
	baseChange struct {
	}

	// 创建合约对象变更事件
	createAccountChange struct {
		baseChange
		account string
	}

	// 自杀事件
	suicideChange struct {
		baseChange
		account string
		prev    bool // whether account had already suicided
	}

	// nonce变更事件
	nonceChange struct {
		baseChange
		account string
		prev    uint64
	}

	// 存储状态变更事件
	storageChange struct {
		baseChange
		account       string
		key, prevalue common.Hash
	}

	// 合约代码状态变更事件
	codeChange struct {
		baseChange
		account            string
		prevcode, prevhash []byte
	}

	// 合约ABI变更事件
	abiChange struct {
		baseChange
		account string
		prevabi string
	}

	// 返还金额变更事件
	refundChange struct {
		baseChange
		prev uint64
	}

	// 转账事件
	// 合约转账动作不执行回滚，失败后数据不会写入区块
	transferChange struct {
		baseChange
		amount int64
		data   []*types.KeyValue
		logs   []*types.ReceiptLog
		// preState carries the raw account KV read before the transfer. It is only populated
		// when ForkEVMTransferRevert is active; empty keeps the legacy no-op revert.
		preState []kvPreState
	}

	// kvPreState is the raw value of one account key before a transfer wrote to it.
	// existed=false means the key was absent and revert must remove it again.
	kvPreState struct {
		key     []byte
		value   []byte
		existed bool
	}

	// 合约生成日志事件
	addLogChange struct {
		baseChange
		txhash common.Hash
		logs   []*types.ReceiptLog
	}

	// 合约生成sha3事件
	addPreimageChange struct {
		baseChange
		hash common.Hash
	}
)

// 在baseChang中定义三个基本操作，子对象中只需要实现必要的操作
func (ch baseChange) revert(s *MemoryStateDB) error {
	return nil
}

func (ch baseChange) getData(s *MemoryStateDB) (kvset []*types.KeyValue) {
	return nil
}

func (ch baseChange) getLog(s *MemoryStateDB) (logs []*types.ReceiptLog) {
	return nil
}

// 创建账户对象的回滚，需要删除缓存中的账户和变更标记
func (ch createAccountChange) revert(s *MemoryStateDB) error {
	delete(s.accounts, ch.account)
	return nil
}

// 创建账户对象的数据集
func (ch createAccountChange) getData(s *MemoryStateDB) (kvset []*types.KeyValue) {
	acc := s.accounts[ch.account]
	if acc != nil {
		kvset = append(kvset, acc.GetDataKV()...)
		kvset = append(kvset, acc.GetStateKV()...)
		return kvset
	}
	return nil
}

func (ch suicideChange) revert(mdb *MemoryStateDB) error {
	// 如果已经自杀过了，不处理
	if ch.prev {
		return nil
	}
	acc := mdb.accounts[ch.account]
	if acc != nil {
		acc.State.Suicided = ch.prev
	}
	return nil
}

func (ch suicideChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	// 如果已经自杀过了，不处理
	if ch.prev {
		return nil
	}
	acc := mdb.accounts[ch.account]
	if acc != nil {
		return acc.GetStateKV()
	}
	return nil
}

func (ch nonceChange) revert(mdb *MemoryStateDB) error {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		acc.State.Nonce = ch.prev
	}
	return nil
}

func (ch nonceChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	// nonce目前没有应用场景，而且每次调用都会变更，暂时先不写到状态数据库中
	//acc := mdb.accounts[ch.account]
	//if acc != nil {
	//	return acc.GetStateKV()
	//}
	return nil
}

func (ch codeChange) revert(mdb *MemoryStateDB) error {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		acc.Data.Code = ch.prevcode
		acc.Data.CodeHash = ch.prevhash
	}
	return nil
}

func (ch codeChange) getData(mdb *MemoryStateDB) (kvset []*types.KeyValue) {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		kvset = append(kvset, acc.GetDataKV()...)
		kvset = append(kvset, acc.GetStateKV()...)
		return kvset
	}
	return nil
}

func (ch abiChange) revert(mdb *MemoryStateDB) error {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		acc.Data.Abi = ch.prevabi
	}
	return nil
}

func (ch abiChange) getData(mdb *MemoryStateDB) (kvset []*types.KeyValue) {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		kvset = append(kvset, acc.GetDataKV()...)
		return kvset
	}
	return nil
}

func (ch storageChange) revert(mdb *MemoryStateDB) error {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		acc.SetState(ch.key, ch.prevalue)
	}
	return nil
}

func (ch storageChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	acc := mdb.accounts[ch.account]
	if _, ok := mdb.stateDirty[ch.account]; ok && acc != nil {
		return acc.GetStateKV()
	}
	return nil
}

func (ch storageChange) getLog(mdb *MemoryStateDB) []*types.ReceiptLog {
	cfg := mdb.api.GetConfig()
	if cfg.IsDappFork(mdb.blockHeight, "evm", evmtypes.ForkEVMState) {
		acc := mdb.accounts[ch.account]
		if acc != nil {
			currentVal := acc.GetState(ch.key)
			receipt := &evmtypes.EVMStateChangeItem{Key: getStateItemKey(ch.account, ch.key.Hex()), PreValue: ch.prevalue.Bytes(), CurrentValue: currentVal.Bytes()}
			return []*types.ReceiptLog{{Ty: evmtypes.TyLogEVMStateChangeItem, Log: types.Encode(receipt)}}
		}
	}
	return nil
}

func (ch refundChange) revert(mdb *MemoryStateDB) error {
	mdb.refund = ch.prev
	return nil
}

func (ch addLogChange) revert(mdb *MemoryStateDB) error {
	logs := mdb.logs[ch.txhash]
	if len(logs) == 1 {
		delete(mdb.logs, ch.txhash)
	} else {
		mdb.logs[ch.txhash] = logs[:len(logs)-1]
	}

	mdb.logSize--
	log15.Info("addLogChange::revert", "mdb.logSize", mdb.logSize)
	return nil
}

func (ch addLogChange) getLog(mdb *MemoryStateDB) []*types.ReceiptLog {
	cfg := mdb.api.GetConfig()
	if !cfg.IsDappFork(mdb.blockHeight, "evm", evmtypes.ForkEVMState) {
		return nil
	}

	return ch.logs
}

func (ch addPreimageChange) revert(mdb *MemoryStateDB) error {
	delete(mdb.preimages, ch.hash)
	return nil
}

func (ch transferChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	return ch.data
}
func (ch transferChange) getLog(mdb *MemoryStateDB) []*types.ReceiptLog {
	return ch.logs
}

// revert restores the account KV captured before the transfer. Before ForkEVMTransferRevert
// preState is empty, so a reverted transfer keeps its old behaviour and leaves the cache as is.
//
// executor.StateDB, the KV used in production, must undo the write through RevertSet rather
// than Set. Under ForkExecRollback Set records the key in GetSetKeys, and every recorded key
// has to show up in the receipt (chain33's checkKV); a restored write that the surviving
// receipt no longer carries would reject the whole transaction. RevertSet removes the key from
// that list as well. A StateDB without RevertSet (the goleveldb used in some tests) still gets
// the direct writes.
func (ch transferChange) revert(mdb *MemoryStateDB) error {
	for _, pre := range ch.preState {
		var original []byte
		if pre.existed {
			original = pre.value
		}
		if err := mdb.revertAccountKey(pre.key, original); err != nil {
			log15.Error("transferChange revert failed", "key", string(pre.key), "err", err)
			return err
		}
	}
	return nil
}

// revertAccountKey undoes one write made by the transfer. original=nil means the key did not
// exist before the transfer and must read as absent afterwards.
func (mdb *MemoryStateDB) revertAccountKey(key, original []byte) error {
	if reverter, ok := mdb.StateDB.(interface {
		RevertSet(key []byte, originalValue []byte) error
	}); ok {
		err := reverter.RevertSet(key, original)
		if err == nil {
			return nil
		}
		if !errors.Is(err, cexecutor.ErrStateDBRevertWithoutTx) {
			return fmt.Errorf("revert state key %x: %w", key, err)
		}
		// No transaction is open, so there is no key list to repair: write through as the
		// legacy path does (the fork that records Set keys is also off).
	}
	if original == nil {
		if deleter, ok := mdb.StateDB.(interface{ Delete([]byte) error }); ok {
			return deleter.Delete(key)
		}
		return mdb.StateDB.Set(key, nil)
	}
	return mdb.StateDB.Set(key, original)
}
