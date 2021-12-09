// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package kvmvccmavl kvmvcc+mavl接口
package kvmvccmavl

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"time"

	dbm "github.com/33cn/chain33/common/db"
	clog "github.com/33cn/chain33/common/log"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/store"
	mavl "github.com/33cn/chain33/system/store/mavl/db"
	"github.com/33cn/chain33/types"
	lru "github.com/hashicorp/golang-lru"
)

var (
	kmlog = log.New("module", "kvmvccMavl")
	// ErrStateHashLost ...
	ErrStateHashLost         = errors.New("ErrStateHashLost")
	kvmvccMavlFork     int64 = 200 * 10000
	isDelMavlData            = false
	delMavlDataHeight        = kvmvccMavlFork + 10000
	delMavlDataState   int32
	wg                 sync.WaitGroup
	quit               bool
	isPrunedMavl       bool                       // 是否是被裁剪过的 mavl
	delPrunedMavlState int32 = delPrunedMavlStart // Upgrade时候删除pruned mavl的状态
	isCompactDelMavl   bool                       // 是否对删除mavl后压缩
)

const (
	cacheSize         = 2048 //可以缓存2048个roothash, height对
	batchDataSize     = 1024 * 1024 * 1
	delMavlStateStart = 1
	delMavlStateEnd   = 0

	delPrunedMavlStart    = 0
	delPrunedMavlStarting = 1
	delPruneMavlEnd       = 2
)

// SetLogLevel set log level
func SetLogLevel(level string) {
	clog.SetLogLevel(level)
}

// DisableLog disable log output
func DisableLog() {
	kmlog.SetHandler(log.DiscardHandler())
}

func init() {
	drivers.Reg("kvmvccmavl", New)
	types.RegFork("store-kvmvccmavl", InitFork)
}

//InitFork ...
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork("store-kvmvccmavl", "ForkKvmvccmavl", 0)
}

// KVmMavlStore provide kvmvcc and mavl store interface implementation
type KVmMavlStore struct {
	*drivers.BaseStore
	*KVMVCCStore
	*MavlStore
	cache *lru.Cache
}

type subKVMVCCConfig struct {
	EnableMVCCIter         bool  `json:"enableMVCCIter"`
	EnableMVCCPrune        bool  `json:"enableMVCCPrune"`
	PruneHeight            int32 `json:"pruneHeight"` //每PruneHeight高度做一次状态精简
	EnableEmptyBlockHandle bool  `json:"enableEmptyBlockHandle"`
	ReservedHeight         int64 `json:"reservedHeight"` //保留最新的ReservedHeight个区块的状态数据
}

type subMavlConfig struct {
	EnableMavlPrefix bool  `json:"enableMavlPrefix"`
	EnableMVCC       bool  `json:"enableMVCC"`
	EnableMavlPrune  bool  `json:"enableMavlPrune"`
	PruneHeight      int32 `json:"pruneHeight"`
	// 是否使能内存树
	EnableMemTree bool `json:"enableMemTree"`
	// 是否使能内存树中叶子节点
	EnableMemVal bool `json:"enableMemVal"`
	// 缓存close ticket数目
	TkCloseCacheLen int32 `json:"tkCloseCacheLen"`
}

type subConfig struct {
	EnableMVCCIter   bool  `json:"enableMVCCIter"`
	EnableMavlPrefix bool  `json:"enableMavlPrefix"`
	EnableMVCC       bool  `json:"enableMVCC"`
	EnableMavlPrune  bool  `json:"enableMavlPrune"`
	PruneMavlHeight  int32 `json:"pruneMavlHeight"`
	EnableMVCCPrune  bool  `json:"enableMVCCPrune"`
	PruneMVCCHeight  int32 `json:"pruneMVCCHeight"`
	// 是否使能内存树
	EnableMemTree bool `json:"enableMemTree"`
	// 是否使能内存树中叶子节点
	EnableMemVal bool `json:"enableMemVal"`
	// 缓存close ticket数目
	TkCloseCacheLen int32 `json:"tkCloseCacheLen"`
	// 使能空块处理
	EnableEmptyBlockHandle bool `json:"enableEmptyBlockHandle"`
	//保留的状态数据的高度
	ReservedHeight int64 `json:"reservedHeight"`
}

// New construct KVMVCCStore module
func New(cfg *types.Store, sub []byte, chain33cfg *types.Chain33Config) queue.Module {
	var kvms *KVmMavlStore
	var subcfg subConfig
	var subKVMVCCcfg subKVMVCCConfig
	var subMavlcfg subMavlConfig
	if sub != nil {
		types.MustDecode(sub, &subcfg)
		subKVMVCCcfg.EnableMVCCIter = subcfg.EnableMVCCIter
		subKVMVCCcfg.EnableMVCCPrune = subcfg.EnableMVCCPrune
		subKVMVCCcfg.PruneHeight = subcfg.PruneMVCCHeight
		subKVMVCCcfg.EnableEmptyBlockHandle = subcfg.EnableEmptyBlockHandle
		subKVMVCCcfg.ReservedHeight = subcfg.ReservedHeight

		subMavlcfg.EnableMavlPrefix = subcfg.EnableMavlPrefix
		subMavlcfg.EnableMVCC = subcfg.EnableMVCC
		subMavlcfg.EnableMavlPrune = subcfg.EnableMavlPrune
		subMavlcfg.PruneHeight = subcfg.PruneMavlHeight
		subMavlcfg.EnableMemTree = subcfg.EnableMemTree
		subMavlcfg.EnableMemVal = subcfg.EnableMemVal
		subMavlcfg.TkCloseCacheLen = subcfg.TkCloseCacheLen
	}

	bs := drivers.NewBaseStore(cfg)
	cache, err := lru.New(cacheSize)
	if err != nil {
		panic("new KVmMavlStore fail")
	}

	kvms = &KVmMavlStore{bs, NewKVMVCC(&subKVMVCCcfg, bs.GetDB()),
		NewMavl(&subMavlcfg, bs.GetDB()), cache}
	// 查询是否已经删除mavl
	_, err = bs.GetDB().Get(genDelMavlKey(mvccPrefix))
	if err == nil {
		isDelMavlData = true
	}
	// 查询是否已经压缩过
	_, err = bs.GetDB().Get(genCompactDelMavlKey(mvccPrefix))
	if err == nil {
		isCompactDelMavl = true
	}
	// 查询是否是删除裁剪版mavl
	isPrunedMavl = isPrunedMavlDB(bs.GetDB())
	// 读取fork高度
	if chain33cfg != nil {
		kvmvccMavlFork = chain33cfg.GetDappFork("store-kvmvccmavl", "ForkKvmvccmavl")
	}
	delMavlDataHeight = kvmvccMavlFork + 10000
	bs.SetChild(kvms)
	return kvms
}

// Close the KVmMavlStore module
func (kvmMavls *KVmMavlStore) Close() {
	quit = true
	wg.Wait()
	kmlog.Info("store wait group done")
	kvmMavls.KVMVCCStore.Close()
	kvmMavls.MavlStore.Close()
	kvmMavls.BaseStore.Close()
	kmlog.Info("store kvmMavls closed")
}

// Set kvs with statehash to KVmMavlStore
func (kvmMavls *KVmMavlStore) Set(datas *types.StoreSet, sync bool) ([]byte, error) {
	if datas.Height < kvmvccMavlFork {
		hash, err := kvmMavls.MavlStore.Set(datas, sync)
		if err != nil {
			return hash, err
		}
		if kvmMavls.kvmvccCfg.EnableEmptyBlockHandle {
			_, err = kvmMavls.KVMVCCStore.SetRdm(datas, hash, sync)
			if err != nil {
				return hash, err
			}
		} else {
			_, err = kvmMavls.KVMVCCStore.Set(datas, hash, sync)
			if err != nil {
				return hash, err
			}
		}
		if err == nil {
			kvmMavls.cache.Add(string(hash), datas.Height)
		}
		return hash, err
	}
	// 仅仅做kvmvcc
	var hash []byte
	var err error
	if kvmMavls.kvmvccCfg.EnableEmptyBlockHandle && datas.Height == kvmvccMavlFork { // kvmvccMavlFork高度下前一个区块需要映射
		hash, err = kvmMavls.KVMVCCStore.SetRdm(datas, nil, sync)
	} else {
		hash, err = kvmMavls.KVMVCCStore.Set(datas, nil, sync)
	}

	if err == nil {
		kvmMavls.cache.Add(string(hash), datas.Height)
	}
	return hash, err
}

// Get kvs with statehash from KVmMavlStore
func (kvmMavls *KVmMavlStore) Get(datas *types.StoreGet) [][]byte {
	if kvmMavls.kvmvccCfg.EnableEmptyBlockHandle {
		// 空块情况下只有第一个hash的为非空块
		mvccHash, err := kvmMavls.KVMVCCStore.GetFirstHashRdm(datas.StateHash)
		if err == nil {
			nData := &types.StoreGet{
				StateHash: mvccHash,
				Keys:      datas.Keys,
			}
			return kvmMavls.KVMVCCStore.Get(nData)
		}
		// ForkKvmvccmavl之后无mavl，即不需映射，直接获取
		return kvmMavls.KVMVCCStore.Get(datas)
	}
	return kvmMavls.KVMVCCStore.Get(datas)
}

// MemSet set kvs to the mem of KVmMavlStore module and return the StateHash
func (kvmMavls *KVmMavlStore) MemSet(datas *types.StoreSet, sync bool) ([]byte, error) {
	if datas.Height < kvmvccMavlFork {
		hash, err := kvmMavls.MavlStore.MemSet(datas, sync)
		if err != nil {
			return hash, err
		}
		if kvmMavls.kvmvccCfg.EnableEmptyBlockHandle {
			_, err = kvmMavls.KVMVCCStore.MemSetRdm(datas, hash, sync)
			if err != nil {
				return hash, err
			}
		} else {
			_, err = kvmMavls.KVMVCCStore.MemSet(datas, hash, sync)
			if err != nil {
				return hash, err
			}
		}

		if err == nil {
			kvmMavls.cache.Add(string(hash), datas.Height)
		}
		return hash, err
	}
	// 仅仅做kvmvcc
	var hash []byte
	var err error
	if kvmMavls.kvmvccCfg.EnableEmptyBlockHandle && datas.Height == kvmvccMavlFork { // kvmvccMavlFork高度下前一个区块需要映射
		hash, err = kvmMavls.KVMVCCStore.MemSetRdm(datas, nil, sync)
	} else {
		hash, err = kvmMavls.KVMVCCStore.MemSet(datas, nil, sync)
	}
	if err == nil {
		kvmMavls.cache.Add(string(hash), datas.Height)
	}
	// 删除Mavl数据
	if datas.Height > delMavlDataHeight && !isDelMavlData && !isDelMavling() {
		// 达到该高度时候，将全局的memTree以及tkCloseCache释放掉
		mavl.ReleaseGlobalMem()
		wg.Add(1)
		go DelMavl(kvmMavls.GetDB())
	}
	// 对删除的mavl进行压缩
	if isDelMavlData && !isCompactDelMavl && !isDelMavling() {
		go CompactDelMavl(kvmMavls.GetDB())
		if datas.Height > delMavlDataHeight && datas.Height < delMavlDataHeight*2 {
			// 出于对区块链安全的角度阻塞执行区块压缩之发生在固定高度区间内
			count := 0
			for {
				if quit || isCompactDelMavl {
					break
				}
				if count%100 == 0 {
					kmlog.Info("block compact db", "count time s", count)
				}
				count++
				time.Sleep(time.Second)
			}
		}
	}
	return hash, err
}

// Commit kvs in the mem of KVmMavlStore module to state db and return the StateHash
func (kvmMavls *KVmMavlStore) Commit(req *types.ReqHash) ([]byte, error) {
	if value, ok := kvmMavls.cache.Get(string(req.Hash)); ok {
		if value.(int64) < kvmvccMavlFork {
			hash, err := kvmMavls.MavlStore.Commit(req)
			if err != nil {
				return hash, err
			}
			_, err = kvmMavls.KVMVCCStore.Commit(req)
			return hash, err
		}
		return kvmMavls.KVMVCCStore.Commit(req)
	}
	return kvmMavls.KVMVCCStore.Commit(req)
}

// Rollback kvs in the mem of KVmMavlStore module and return the StateHash
func (kvmMavls *KVmMavlStore) Rollback(req *types.ReqHash) ([]byte, error) {
	if value, ok := kvmMavls.cache.Get(string(req.Hash)); ok {
		if value.(int64) < kvmvccMavlFork {
			hash, err := kvmMavls.MavlStore.Rollback(req)
			if err != nil {
				return hash, err
			}
			realReq := &types.ReqHash{
				Hash:    req.Hash,
				Upgrade: req.Upgrade,
			}
			// 获取kvmvcc的实际statehash
			if kvmMavls.kvmvccCfg.EnableEmptyBlockHandle {
				if value, ok := kvmMavls.cache.Get(string(realReq.Hash)); ok {
					mvccHash, err := kvmMavls.KVMVCCStore.GetHashRdm(realReq.Hash, value.(int64))
					if err == nil {
						realReq.Hash = mvccHash
					}
				}
			}
			_, err = kvmMavls.KVMVCCStore.Rollback(realReq)
			return hash, err
		}
		return kvmMavls.KVMVCCStore.Rollback(req)
	}
	return kvmMavls.KVMVCCStore.Rollback(req)
}

// IterateRangeByStateHash travel with Prefix by StateHash  to get the latest version kvs.
func (kvmMavls *KVmMavlStore) IterateRangeByStateHash(statehash []byte, start []byte, end []byte, ascending bool, fn func(key, value []byte) bool) {
	if value, ok := kvmMavls.cache.Get(string(statehash)); ok && value.(int64) < kvmvccMavlFork {
		kvmMavls.MavlStore.IterateRangeByStateHash(statehash, start, end, ascending, fn)
		return
	}
	hash := statehash
	if kvmMavls.kvmvccCfg.EnableEmptyBlockHandle {
		mvccHash, err := kvmMavls.KVMVCCStore.GetFirstHashRdm(statehash)
		if err == nil {
			hash = mvccHash
		}
	}
	kvmMavls.KVMVCCStore.IterateRangeByStateHash(hash, start, end, ascending, fn)
}

// ProcEvent handles supported events
func (kvmMavls *KVmMavlStore) ProcEvent(msg *queue.Message) {
	if msg == nil {
		return
	}
	msg.ReplyErr("KVmMavlStore", types.ErrActionNotSupport)
}

// MemSetUpgrade set kvs to the mem of KVmMavlStore module  not cache the tree and return the StateHash
func (kvmMavls *KVmMavlStore) MemSetUpgrade(datas *types.StoreSet, sync bool) ([]byte, error) {
	if datas.Height < kvmvccMavlFork {
		var hash []byte
		var err error

		if isPrunedMavl {
			hash, err = kvmMavls.MavlStore.MemSet(datas, sync)
			if err != nil {
				return hash, err
			}
		} else {
			hash, err = kvmMavls.MavlStore.MemSetUpgrade(datas, sync)
			if err != nil {
				return hash, err
			}
		}

		if kvmMavls.kvmvccCfg.EnableEmptyBlockHandle {
			_, err = kvmMavls.KVMVCCStore.MemSetRdm(datas, hash, sync)
			if err != nil {
				return hash, err
			}
		} else {
			_, err = kvmMavls.KVMVCCStore.MemSet(datas, hash, sync)
			if err != nil {
				return hash, err
			}
		}

		if err == nil {
			kvmMavls.cache.Add(string(hash), datas.Height)
		}
		return hash, err
	}
	// 仅仅做kvmvcc
	var hash []byte
	var err error
	if kvmMavls.kvmvccCfg.EnableEmptyBlockHandle && datas.Height == kvmvccMavlFork { // kvmvccMavlFork高度下前一个区块需要映射
		hash, err = kvmMavls.KVMVCCStore.MemSetRdm(datas, nil, sync)
	} else {
		hash, err = kvmMavls.KVMVCCStore.MemSet(datas, nil, sync)
	}
	if err == nil {
		kvmMavls.cache.Add(string(hash), datas.Height)
	}
	return hash, err
}

// CommitUpgrade kvs in the mem of KVmMavlStore module to state db and return the StateHash
func (kvmMavls *KVmMavlStore) CommitUpgrade(req *types.ReqHash) ([]byte, error) {
	var hash []byte
	var err error
	if isPrunedMavl {
		hash, err = kvmMavls.Commit(req)
		if isNeedDelPrunedMavl() {
			wg.Add(1)
			go deletePrunedMavl(kvmMavls.GetDB())
		}
	} else {
		hash, err = kvmMavls.KVMVCCStore.CommitUpgrade(req)
	}
	return hash, err
}

// Del set kvs to nil with StateHash
func (kvmMavls *KVmMavlStore) Del(req *types.StoreDel) ([]byte, error) {
	if req.Height < kvmvccMavlFork {
		hash, err := kvmMavls.MavlStore.Del(req)
		if err != nil {
			return hash, err
		}
		storeDel := &types.StoreDel{
			StateHash: req.StateHash,
			Height:    req.Height,
		}
		// 获取kvmvcc的实际statehash
		if kvmMavls.kvmvccCfg.EnableEmptyBlockHandle {
			mvccHash, err := kvmMavls.KVMVCCStore.GetHashRdm(req.StateHash, req.Height)
			if err == nil {
				storeDel.StateHash = mvccHash
			}
		}
		_, err = kvmMavls.KVMVCCStore.Del(storeDel)
		if err != nil {
			return req.StateHash, err
		}
		if err == nil {
			kvmMavls.cache.Remove(string(req.StateHash))
		}
		return req.StateHash, err
	}
	// 仅仅做kvmvcc
	hash, err := kvmMavls.KVMVCCStore.Del(req)
	if err == nil {
		kvmMavls.cache.Remove(string(req.StateHash))
	}
	return hash, err
}

// DelMavl 数据库中mavl数据清除
// 达到kvmvccMavlFork + 100000 后触发清除
func DelMavl(db dbm.DB) {
	defer wg.Done()
	setDelMavl(delMavlStateStart)
	defer setDelMavl(delMavlStateEnd)
	prefix := ""
	for {
		kmlog.Debug("start once del mavl")
		var loop bool
		loop, prefix = delMavlData(db, prefix)
		if !loop {
			break
		}
		kmlog.Debug("end once del mavl")
		time.Sleep(time.Second * 1)
	}
}

func delMavlData(db dbm.DB, prefix string) (bool, string) {
	it := db.Iterator([]byte(prefix), types.EmptyValue, false)
	defer it.Close()
	batch := db.NewBatch(false)
	count := 0
	const onceCount = 50
	for it.Rewind(); it.Valid(); it.Next() {
		if quit {
			return false, ""
		}
		if !bytes.HasPrefix(it.Key(), mvccPrefix) { // 将非mvcc的mavl数据全部删除
			batch.Delete(it.Key())
			if batch.ValueSize() > batchDataSize {
				dbm.MustWrite(batch)
				batch.Reset()
				count++
			}
		}
		if count > onceCount {
			if it.Next() {
				return true, string(it.Key())
			}
			return true, ""
		}
	}
	batch.Set(genDelMavlKey(mvccPrefix), []byte(""))
	dbm.MustWrite(batch)
	isDelMavlData = true
	kmlog.Info("DelMavl success")
	return false, ""
}

func genDelMavlKey(prefix []byte) []byte {
	delMavl := "--delMavlData--"
	return []byte(fmt.Sprintf("%s%s", string(prefix), delMavl))
}

func isDelMavling() bool {
	return atomic.LoadInt32(&delMavlDataState) == 1
}

func setDelMavl(state int32) {
	atomic.StoreInt32(&delMavlDataState, state)
}

//CompactDelMavl ...
func CompactDelMavl(db dbm.DB) {
	setDelMavl(delMavlStateStart)
	defer setDelMavl(delMavlStateEnd)
	// 开始进行压缩处理
	kmlog.Info("start compact db")
	err := db.CompactRange(nil, nil)
	if err == nil {
		db.Set(genCompactDelMavlKey(mvccPrefix), []byte(""))
		isCompactDelMavl = true
	}
	kmlog.Info("end compact db", "error", err)
}

func genCompactDelMavlKey(prefix []byte) []byte {
	key := "--compactDelMavl--"
	return []byte(fmt.Sprintf("%s%s", string(prefix), key))
}

func isNeedDelPrunedMavl() bool {
	return atomic.LoadInt32(&delPrunedMavlState) == 0
}

func setDelPrunedMavl(state int32) {
	atomic.StoreInt32(&delPrunedMavlState, state)
}

func isPrunedMavlDB(db dbm.DB) bool {
	prefix := []byte(leafNodePrefix)
	it := db.Iterator(prefix, nil, true)
	defer it.Close()
	var isCommit bool
	for it.Rewind(); it.Valid(); it.Next() {
		isCommit = true
		kmlog.Info("need commit mval")
		break
	}
	return isCommit
}

func deletePrunedMavl(db dbm.DB) {
	defer wg.Done()
	setDelPrunedMavl(delPrunedMavlStarting)
	defer setDelPrunedMavl(delPruneMavlEnd)
	prefixS := []string{hashNodePrefix, leafNodePrefix, leafKeyCountPrefix, oldLeafKeyCountPrefix}
	for _, str := range prefixS {
		for {
			stat := deletePrunedMavlData(db, str)
			if stat == 0 {
				return
			} else if stat == 1 {
				break
			} else {
				time.Sleep(time.Millisecond * 100)
			}
		}
	}
}

func deletePrunedMavlData(db dbm.DB, prefix string) (status int) {
	it := db.Iterator([]byte(prefix), nil, false)
	defer it.Close()
	count := 0
	const onceCount = 200
	if it.Rewind() && it.Valid() {
		batch := db.NewBatch(false)
		for it.Next(); it.Valid(); it.Next() { //第一个不做删除
			if quit {
				return 0 // quit
			}
			batch.Delete(it.Key())
			if batch.ValueSize() > batchDataSize {
				dbm.MustWrite(batch)
				batch.Reset()
				count++
			}
			if count > onceCount {
				return 2 //loop
			}
		}
		dbm.MustWrite(batch)
	}
	return 1 // this  prefix Iterator over
}
