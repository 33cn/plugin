// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvmvccmavl

import (
	"bytes"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
)

const (
	pruningStateStart  = 1
	pruningStateEnd    = 0
	defaultPruneHeight = 10000 // 每个10000裁剪一次
)

var (
	pruningState int32
	batch        dbm.Batch
)

var (
	//同common/db中的mvcc相关的定义保持一致
	mvccPrefix = []byte(".-mvcc-.")
	mvccMeta   = append(mvccPrefix, []byte("m.")...)
	mvccData   = append(mvccPrefix, []byte("d.")...)
	//mvccLast               = append(mvccPrefix, []byte("l.")...)
	mvccMetaVersion        = append(mvccMeta, []byte("version.")...)
	mvccMetaVersionKeyList = append(mvccMeta, []byte("versionkl.")...)

	// for empty block
	rdmHashPrefix = append(mvccPrefix, []byte("rdm.")...)
)

// KVMVCCStore provide kvmvcc store interface implementation
type KVMVCCStore struct {
	db        dbm.DB
	mvcc      dbm.MVCC
	kvsetmap  map[string][]*types.KeyValue
	sync      bool
	kvmvccCfg *subKVMVCCConfig
}

// NewKVMVCC construct KVMVCCStore module
func NewKVMVCC(sub *subKVMVCCConfig, db dbm.DB) *KVMVCCStore {
	var kvs *KVMVCCStore
	if sub == nil {
		panic("sub is nil memory")
	}
	if sub.PruneHeight == 0 {
		sub.PruneHeight = defaultPruneHeight
	}
	if sub.EnableMVCCIter {
		kvs = &KVMVCCStore{db: db, mvcc: dbm.NewMVCCIter(db), kvsetmap: make(map[string][]*types.KeyValue), kvmvccCfg: sub}
	} else {
		kvs = &KVMVCCStore{db: db, mvcc: dbm.NewMVCC(db), kvsetmap: make(map[string][]*types.KeyValue), kvmvccCfg: sub}
	}
	return kvs
}

// Close the KVMVCCStore module
func (mvccs *KVMVCCStore) Close() {
	kmlog.Info("store kvdb closed")
}

// Set kvs with statehash to KVMVCCStore
func (mvccs *KVMVCCStore) Set(datas *types.StoreSet, hash []byte, sync bool) ([]byte, error) {
	if hash == nil {
		hash = calcHash(datas)
	}
	kvlist, err := mvccs.mvcc.AddMVCC(datas.KV, hash, datas.StateHash, datas.Height)
	if err != nil {
		return nil, err
	}
	mvccs.saveKVSets(kvlist, sync)
	return hash, nil
}

// Get kvs with statehash from KVMVCCStore
func (mvccs *KVMVCCStore) Get(datas *types.StoreGet) [][]byte {
	values := make([][]byte, len(datas.Keys))
	version, err := mvccs.mvcc.GetVersion(datas.StateHash)
	if err != nil {
		kmlog.Error("Get version by hash failed.", "hash", common.ToHex(datas.StateHash), "error:", err)
		return values
	}
	for i := 0; i < len(datas.Keys); i++ {
		value, err := mvccs.mvcc.GetV(datas.Keys[i], version)
		if err != nil {
			//kmlog.Error("GetV by Keys failed.", "Key", string(datas.Keys[i]), "version", version)
		} else if value != nil {
			values[i] = value
		}
	}
	return values
}

// MemSet set kvs to the mem of KVMVCCStore module and return the StateHash
func (mvccs *KVMVCCStore) MemSet(datas *types.StoreSet, hash []byte, sync bool) ([]byte, error) {
	beg := types.Now()
	defer func() {
		kmlog.Debug("kvmvcc MemSet", "cost", types.Since(beg))
	}()
	kvset, err := mvccs.checkVersion(datas.Height)
	if err != nil {
		return nil, err
	}
	if hash == nil {
		hash = calcHash(datas)
	}
	//kmlog.Debug("KVMVCCStore MemSet AddMVCC", "prestatehash", common.ToHex(datas.StateHash), "hash", common.ToHex(hash), "height", datas.Height)
	kvlist, err := mvccs.mvcc.AddMVCC(datas.KV, hash, datas.StateHash, datas.Height)
	if err != nil {
		return nil, err
	}
	if len(kvlist) > 0 {
		kvset = append(kvset, kvlist...)
	}
	mvccs.kvsetmap[string(hash)] = kvset
	mvccs.sync = sync
	// 进行裁剪
	if mvccs.kvmvccCfg != nil && mvccs.kvmvccCfg.EnableMVCCPrune &&
		!isPruning() && datas.Height%int64(mvccs.kvmvccCfg.PruneHeight) == 0 {
		wg.Add(1)
		go mvccs.pruningMVCC(datas.Height)
	}
	return hash, nil
}

// Commit kvs in the mem of KVMVCCStore module to state db and return the StateHash
func (mvccs *KVMVCCStore) Commit(req *types.ReqHash) ([]byte, error) {
	beg := types.Now()
	defer func() {
		kmlog.Debug("kvmvcc Commit", "cost", types.Since(beg))
	}()
	_, ok := mvccs.kvsetmap[string(req.Hash)]
	if !ok {
		kmlog.Error("store kvmvcc commit", "err", types.ErrHashNotFound)
		return nil, types.ErrHashNotFound
	}
	//kmlog.Debug("KVMVCCStore Commit saveKVSets", "hash", common.ToHex(req.Hash))
	mvccs.saveKVSets(mvccs.kvsetmap[string(req.Hash)], mvccs.sync)
	delete(mvccs.kvsetmap, string(req.Hash))
	return req.Hash, nil
}

// CommitUpgrade kvs in the mem of KVMVCCStore module to state db and re
func (mvccs *KVMVCCStore) CommitUpgrade(req *types.ReqHash) ([]byte, error) {
	_, ok := mvccs.kvsetmap[string(req.Hash)]
	if !ok {
		kmlog.Error("store kvmvcc commit", "err", types.ErrHashNotFound)
		return nil, types.ErrHashNotFound
	}
	//kmlog.Debug("KVMVCCStore Commit saveKVSets", "hash", common.ToHex(req.Hash))
	if batch == nil {
		batch = mvccs.db.NewBatch(mvccs.sync)
	}
	batch.Reset()
	kvset := mvccs.kvsetmap[string(req.Hash)]
	for i := 0; i < len(kvset); i++ {
		if kvset[i].Value == nil {
			batch.Delete(kvset[i].Key)
		} else {
			batch.Set(kvset[i].Key, kvset[i].Value)
		}
	}
	dbm.MustWrite(batch)
	delete(mvccs.kvsetmap, string(req.Hash))
	return req.Hash, nil
}

// Rollback kvs in the mem of KVMVCCStore module and return the StateHash
func (mvccs *KVMVCCStore) Rollback(req *types.ReqHash) ([]byte, error) {
	_, ok := mvccs.kvsetmap[string(req.Hash)]
	if !ok {
		kmlog.Error("store kvmvcc rollback", "err", types.ErrHashNotFound)
		return nil, types.ErrHashNotFound
	}

	//kmlog.Debug("KVMVCCStore Rollback", "hash", common.ToHex(req.Hash))

	delete(mvccs.kvsetmap, string(req.Hash))
	return req.Hash, nil
}

// IterateRangeByStateHash travel with Prefix by StateHash  to get the latest version kvs.
func (mvccs *KVMVCCStore) IterateRangeByStateHash(statehash []byte, start []byte, end []byte, ascending bool, fn func(key, value []byte) bool) {
	if !mvccs.kvmvccCfg.EnableMVCCIter {
		panic("call IterateRangeByStateHash when disable mvcc iter")
	}
	//按照kv最新值来进行遍历处理，要求statehash必须是最新区块的statehash，否则不支持该接口
	maxVersion, err := mvccs.mvcc.GetMaxVersion()
	if err != nil {
		kmlog.Error("KVMVCCStore IterateRangeByStateHash can't get max version, ignore the call.", "err", err)
		return
	}

	version, err := mvccs.mvcc.GetVersion(statehash)
	if err != nil {
		kmlog.Error("KVMVCCStore IterateRangeByStateHash can't get version, ignore the call.", "stateHash", common.ToHex(statehash), "err", err)
		return
	}

	if version != maxVersion {
		kmlog.Error("KVMVCCStore IterateRangeByStateHash call failed for maxVersion does not match version.", "maxVersion", maxVersion, "version", version, "stateHash", common.ToHex(statehash))
		return
	}

	//kmlog.Info("KVMVCCStore do the IterateRangeByStateHash")
	listhelper := dbm.NewListHelper(mvccs.mvcc.(*dbm.MVCCIter))
	listhelper.IteratorCallback(start, end, 0, 1, fn)
}

// ProcEvent handles supported events
func (mvccs *KVMVCCStore) ProcEvent(msg queue.Message) {
	msg.ReplyErr("KVStore", types.ErrActionNotSupport)
}

// Del set kvs to nil with StateHash
func (mvccs *KVMVCCStore) Del(req *types.StoreDel) ([]byte, error) {
	maxVersion, err := mvccs.mvcc.GetMaxVersion()
	if err != nil {
		kmlog.Error("store kvmvcc GetMaxVersion failed", "err", err)
		if err != types.ErrNotFound {
			panic(err)
		} else {
			return nil, err
		}
	}
	var kvset []*types.KeyValue
	for i := maxVersion; i >= req.Height; i-- {
		hash, err := mvccs.mvcc.GetVersionHash(i)
		if err != nil {
			kmlog.Warn("store kvmvcc Del GetVersionHash failed", "height", i, "maxVersion", maxVersion)
			continue
		}
		kvlist, err := mvccs.mvcc.DelMVCC(hash, i, true)
		if err != nil {
			kmlog.Warn("store kvmvcc Del DelMVCC failed", "height", i, "err", err)
			continue
		}
		kvset = append(kvset, kvlist...)
		kmlog.Debug("store kvmvcc Del DelMVCC4Height", "height", i, "maxVersion", maxVersion)
	}
	if len(kvset) > 0 {
		mvccs.saveKVSets(kvset, mvccs.sync)
	}
	return req.StateHash, nil
}

func (mvccs *KVMVCCStore) saveKVSets(kvset []*types.KeyValue, sync bool) {
	if len(kvset) == 0 {
		return
	}

	storeBatch := mvccs.db.NewBatch(sync)

	for i := 0; i < len(kvset); i++ {
		if kvset[i].Value == nil {
			storeBatch.Delete(kvset[i].Key)
		} else {
			storeBatch.Set(kvset[i].Key, kvset[i].Value)
		}
	}
	dbm.MustWrite(storeBatch)
}

//GetMaxVersion GetMaxVersion 获取当前最大高度
func (mvccs *KVMVCCStore) GetMaxVersion() (int64, error) {
	return mvccs.mvcc.GetMaxVersion()
}

func (mvccs *KVMVCCStore) checkVersion(height int64) ([]*types.KeyValue, error) {
	//检查新加入区块的height和现有的version的关系，来判断是否要回滚数据
	maxVersion, err := mvccs.mvcc.GetMaxVersion()
	if err != nil {
		if err != types.ErrNotFound {
			kmlog.Error("store kvmvcc checkVersion GetMaxVersion failed", "err", err)
			panic(err)
		} else {
			maxVersion = -1
			kmlog.Error("store kvmvcc checkVersion GetMaxVersion failed", "err", err, "maxVersion", maxVersion)
		}
	}

	//kmlog.Debug("store kvmvcc checkVersion ", "maxVersion", maxVersion, "currentVersion", height)

	var kvset []*types.KeyValue
	if maxVersion < height-1 {
		kmlog.Error("store kvmvcc checkVersion found statehash lost", "maxVersion", maxVersion, "height", height)
		return nil, ErrStateHashLost
	} else if maxVersion == height-1 {
		return nil, nil
	} else {
		for i := maxVersion; i >= height; i-- {
			hash, err := mvccs.mvcc.GetVersionHash(i)
			if err != nil {
				kmlog.Warn("store kvmvcc checkVersion GetVersionHash failed", "height", i, "maxVersion", maxVersion)
				continue
			}
			kvlist, err := mvccs.mvcc.DelMVCC(hash, i, false)
			if err != nil {
				kmlog.Warn("store kvmvcc checkVersion DelMVCC failed", "height", i, "err", err)
				continue
			}
			kvset = append(kvset, kvlist...)

			kmlog.Debug("store kvmvcc checkVersion DelMVCC4Height", "height", i, "maxVersion", maxVersion)
		}
	}

	return kvset, nil
}

func calcHash(datas proto.Message) []byte {
	b := types.Encode(datas)
	return common.Sha256(b)
}

//SetRdm ...
func (mvccs *KVMVCCStore) SetRdm(datas *types.StoreSet, mavlHash []byte, sync bool) ([]byte, error) {
	mvccHash := calcHash(datas)
	// 取出前一个hash映射
	var preMvccHash []byte
	var err error
	if datas.Height > 0 {
		preMvccHash, err = mvccs.GetHashRdm(datas.StateHash, datas.Height-1)
		if err != nil {
			kmlog.Error("kvmvcc GetHashRdm", "error", err, "height", datas.Height-1)
			return nil, err
		}
	}
	kvlist, err := mvccs.mvcc.AddMVCC(datas.KV, mvccHash, preMvccHash, datas.Height)
	if err != nil {
		return nil, err
	}

	hash := mvccHash
	if mavlHash != nil {
		hash = mavlHash
		// add rdm
		key := calcRdmKey(mavlHash, datas.Height)
		kvlist = append(kvlist, &types.KeyValue{Key: key, Value: mvccHash})
	}
	mvccs.saveKVSets(kvlist, sync)
	return hash, nil
}

//MemSetRdm ...
func (mvccs *KVMVCCStore) MemSetRdm(datas *types.StoreSet, mavlHash []byte, sync bool) ([]byte, error) {
	beg := types.Now()
	defer func() {
		kmlog.Debug("kvmvcc MemSetRdm", "cost", types.Since(beg))
	}()
	kvset, err := mvccs.checkVersion(datas.Height)
	if err != nil {
		return nil, err
	}

	//kmlog.Debug("KVMVCCStore MemSet AddMVCC", "prestatehash", common.ToHex(datas.StateHash), "hash", common.ToHex(hash), "height", datas.Height)
	mvcchash := calcHash(datas)

	// 取出前一个hash映射
	var preMvccHash []byte
	if datas.Height > 0 {
		preMvccHash, err = mvccs.GetHashRdm(datas.StateHash, datas.Height-1)
		if err != nil {
			kmlog.Error("kvmvcc GetHashRdm", "error", err, "height", datas.Height-1)
			return nil, err
		}
	}

	kvlist, err := mvccs.mvcc.AddMVCC(datas.KV, mvcchash, preMvccHash, datas.Height)
	if err != nil {
		return nil, err
	}
	if len(kvlist) > 0 {
		kvset = append(kvset, kvlist...)
	}

	hash := mvcchash
	if mavlHash != nil {
		// 如果mavlHash是nil,即返回mvcchash
		hash = mavlHash
		// add rdm
		kv := &types.KeyValue{Key: calcRdmKey(mavlHash, datas.Height), Value: mvcchash}
		kvset = append(kvset, kv)
	}
	mvccs.kvsetmap[string(hash)] = kvset
	mvccs.sync = sync

	// 进行裁剪
	if mvccs.kvmvccCfg != nil && mvccs.kvmvccCfg.EnableMVCCPrune &&
		!isPruning() && datas.Height%int64(mvccs.kvmvccCfg.PruneHeight) == 0 {
		wg.Add(1)
		go mvccs.pruningMVCC(datas.Height)
	}
	return hash, nil
}

//GetHashRdm ...
func (mvccs *KVMVCCStore) GetHashRdm(hash []byte, height int64) ([]byte, error) {
	key := calcRdmKey(hash, height)
	return mvccs.db.Get(key)
}

//GetFirstHashRdm ...
func (mvccs *KVMVCCStore) GetFirstHashRdm(hash []byte) ([]byte, error) {
	prefix := append(rdmHashPrefix, hash...)
	list := dbm.NewListHelper(mvccs.db)
	values := list.IteratorScanFromFirst(prefix, 1, dbm.ListASC)
	if len(values) == 1 {
		return values[0], nil
	}
	return nil, types.ErrNotFound
}

func calcRdmKey(hash []byte, height int64) []byte {
	key := append(rdmHashPrefix, hash...)
	key = append(key, []byte(".")...)
	key = append(key, pad(height)...)
	return key
}

/*裁剪-------------------------------------------*/
func (mvccs *KVMVCCStore) pruningMVCC(curHeight int64) {
	defer wg.Done()
	safeHeight := curHeight - mvccs.kvmvccCfg.ReservedHeight
	if safeHeight <= 0 {
		return
	}
	setPruning(pruningStateStart)
	defer setPruning(pruningStateEnd)
	start := time.Now()
	pruningMVCCDappExpired(mvccs.db, safeHeight)
	kmlog.Info("pruningMVCCDappExpired", "current height", curHeight, "cost", time.Since(start))
	pruningMVCCData(mvccs.db, safeHeight)
	kmlog.Info("pruningMVCCData", "current height", curHeight, "cost", time.Since(start))
	pruningMVCCMeta(mvccs.db, safeHeight)
	kmlog.Info("pruningMVCCMeta", "current height", curHeight, "cost", time.Since(start))
}

func pruningMVCCData(db dbm.DB, safeHeight int64) {
	it := db.Iterator(mvccData, nil, true)
	defer it.Close()
	newKey := []byte("--.xxx.--")
	batch := db.NewBatch(false)
	defer dbm.MustWrite(batch)
	for it.Rewind(); it.Valid(); it.Next() {
		if quit {
			//该处退出
			return
		}
		key, height, err := getKeyVersion(it.Key())
		if err != nil {
			continue
		}
		if height >= safeHeight {
			continue
		}
		if bytes.Compare(key, newKey) != 0 {
			newKey = make([]byte, len(key))
			copy(newKey, key)
			continue
		}
		batch.Delete(it.Key())
		if batch.ValueSize() > 1<<20 {
			dbm.MustWrite(batch)
			batch.Reset()
		}
	}
}

// TODO:
// 合约里自定义规则用于检查哪些kv对是已经废弃的
// 对于定义过规则的合约，这里会遍历该合约所有的kv对，然后合约内部检查该kv对是否已经废弃
// 更高效的做法是仅遍历指定合约里可能废弃的那些key（通过进一步指定prefix实现），但通用性会变差
// 现阶段效率差不多，暂时不做进一步优化
func pruningMVCCDappExpired(db dbm.DB, safeHeight int64) {
	names := dapp.KVExpiredCheckerList()
	for _, name := range names {
		pruneDapp(db, name, safeHeight)
	}
}

func pruneDapp(db dbm.DB, name string, safeHeight int64) {
	checkFunc, ok := dapp.LoadKVExpiredChecker(name)
	if !ok {
		return
	}
	var prefix []byte
	prefix = append(prefix, mvccData...)
	prefix = append(prefix, "mavl-"+name...)
	it := db.Iterator(prefix, nil, true)
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		if quit {
			//该处退出
			return
		}
		key, height, err := getKeyVersion(it.Key())
		if err != nil {
			continue
		}
		if height > safeHeight {
			continue
		}
		if checkFunc(key, it.Value()) {
			deleteKeyAllVersion(db, key)
		}
	}
}

func deleteKeyAllVersion(db dbm.DB, key []byte) {
	start := append(mvccData, key...)
	it := db.Iterator(start, nil, false)
	defer it.Close()
	batch := db.NewBatch(false)
	for it.Rewind(); it.Valid(); it.Next() {
		batch.Delete(it.Key())
		if batch.ValueSize() > 1<<20 {
			dbm.MustWrite(batch)
			batch.Reset()
		}
	}
	dbm.MustWrite(batch)
}

func pruningMVCCMeta(db dbm.DB, height int64) {
	pruningMVCCMetaVersion(db, height)
	pruningMVCCMetaVersionKeyList(db, height)
}

func pruningMVCCMetaVersion(db dbm.DB, height int64) {
	startPrefix := append(mvccMetaVersion, pad(0)...)
	endPrefix := append(mvccMetaVersion, pad(height)...)
	it := db.Iterator(startPrefix, endPrefix, false)
	defer it.Close()
	batch := db.NewBatch(false)
	for it.Rewind(); it.Valid(); it.Next() {
		if quit {
			//该处退出
			return
		}
		batch.Delete(it.Key())
		batch.Delete(append(mvccMeta, it.Value()...))
		if batch.ValueSize() > 1<<20 {
			dbm.MustWrite(batch)
			batch.Reset()
		}
	}
	dbm.MustWrite(batch)
	_ = db.CompactRange(startPrefix, endPrefix)
}

func pruningMVCCMetaVersionKeyList(db dbm.DB, height int64) {
	startPrefix := append(mvccMetaVersionKeyList, pad(0)...)
	endPrefix := append(mvccMetaVersionKeyList, pad(height)...)
	it := db.Iterator(startPrefix, endPrefix, false)
	defer it.Close()
	batch := db.NewBatch(false)
	for it.Rewind(); it.Valid(); it.Next() {
		if quit {
			//该处退出
			return
		}
		batch.Delete(it.Key())
		if batch.ValueSize() > 1<<20 {
			dbm.MustWrite(batch)
			batch.Reset()
		}
	}
	dbm.MustWrite(batch)
	_ = db.CompactRange(startPrefix, endPrefix)
}

func getKeyVersion(vsnKey []byte) ([]byte, int64, error) {
	if len(vsnKey) < len(mvccData)+1+20 {
		return nil, 0, types.ErrSize
	}
	sLen := vsnKey[len(vsnKey)-20:]
	iLen, err := strconv.Atoi(string(sLen))
	if err != nil {
		return nil, 0, types.ErrSize
	}
	k := bytes.TrimPrefix(vsnKey, mvccData)
	key := k[:len(k)-1-20]
	return key, int64(iLen), nil
}

func pad(version int64) []byte {
	//equals to `[]byte(fmt.Sprintf("%020d", version))`
	sInt := strconv.FormatInt(version, 10)
	result := []byte("00000000000000000000")
	copy(result[20-len(sInt):], sInt)
	return result
}

func isPruning() bool {
	return atomic.LoadInt32(&pruningState) == pruningStateStart
}

func setPruning(state int32) {
	atomic.StoreInt32(&pruningState, state)
}
