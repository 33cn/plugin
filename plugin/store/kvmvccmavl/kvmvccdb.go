// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvmvccmavl

import (
	"bytes"
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
)

const (
	pruningStateStart  = 1
	pruningStateEnd    = 0
	onceScanCount      = 10000 // 单次扫描数目
	onceCount          = 1000  // 容器长度
	levelPruningHeight = 100 * 10000
)

var (
	maxRollbackNum = 200
	// 是否开启裁剪
	enablePrune bool
	// 每个10000裁剪一次
	pruneHeight  = 10000
	pruningState int32
	batch        dbm.Batch
)

var (
	//同common/db中的mvcc相关的定义保持一致
	mvccPrefix = []byte(".-mvcc-.")
	//mvccMeta               = append(mvccPrefix, []byte("m.")...)
	mvccData = append(mvccPrefix, []byte("d.")...)
	//mvccLast               = append(mvccPrefix, []byte("l.")...)
	//mvccMetaVersion        = append(mvccMeta, []byte("version.")...)
	//mvccMetaVersionKeyList = append(mvccMeta, []byte("versionkl.")...)
)

// KVMVCCStore provide kvmvcc store interface implementation
type KVMVCCStore struct {
	db              dbm.DB
	mvcc            dbm.MVCC
	kvsetmap        map[string][]*types.KeyValue
	enableMVCCIter  bool
	enableMavlPrune bool
	pruneHeight     int32
	sync            bool
}

// NewKVMVCC construct KVMVCCStore module
func NewKVMVCC(sub *subKVMVCCConfig, db dbm.DB) *KVMVCCStore {
	var kvs *KVMVCCStore
	enable := false
	if sub != nil {
		enable = sub.EnableMVCCIter
	}
	if enable {
		kvs = &KVMVCCStore{db, dbm.NewMVCCIter(db), make(map[string][]*types.KeyValue),
			true, sub.EnableMavlPrune, sub.PruneHeight, false}
	} else {
		kvs = &KVMVCCStore{db, dbm.NewMVCC(db), make(map[string][]*types.KeyValue),
			false, sub.EnableMavlPrune, sub.PruneHeight, false}
	}
	EnablePrune(sub.EnableMavlPrune)
	SetPruneHeight(int(sub.PruneHeight))
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
		kmlog.Error("Get version by hash failed.", "hash", common.ToHex(datas.StateHash))
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
		kmlog.Info("kvmvcc MemSet", "cost", types.Since(beg))
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
	if enablePrune && !isPruning() &&
		pruneHeight != 0 &&
		datas.Height%int64(pruneHeight) == 0 &&
		datas.Height/int64(pruneHeight) > 1 {
		wg.Add(1)
		go pruning(mvccs.db, datas.Height)
	}
	return hash, nil
}

// Commit kvs in the mem of KVMVCCStore module to state db and return the StateHash
func (mvccs *KVMVCCStore) Commit(req *types.ReqHash) ([]byte, error) {
	beg := types.Now()
	defer func() {
		kmlog.Info("kvmvcc Commit", "cost", types.Since(beg))
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
	batch.Write()
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
	if !mvccs.enableMVCCIter {
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
	kvset, err := mvccs.mvcc.DelMVCC(req.StateHash, req.Height, true)
	if err != nil {
		kmlog.Error("store kvmvcc del", "err", err)
		return nil, err
	}

	kmlog.Info("KVMVCCStore Del", "hash", common.ToHex(req.StateHash), "height", req.Height)
	mvccs.saveKVSets(kvset, mvccs.sync)
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
	storeBatch.Write()
}

// GetMaxVersion 获取当前最大高度
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
		count := 1
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
			//为避免高度差过大时出现异常，做一个保护，一次最多回滚200个区块
			count++
			if count >= maxRollbackNum {
				break
			}
		}
	}

	return kvset, nil
}

func calcHash(datas proto.Message) []byte {
	b := types.Encode(datas)
	return common.Sha256(b)
}

/*裁剪-------------------------------------------*/

// EnablePrune 使能裁剪
func EnablePrune(enable bool) {
	enablePrune = enable
}

// SetPruneHeight 设置每次裁剪高度
func SetPruneHeight(height int) {
	pruneHeight = height
}

func pruning(db dbm.DB, height int64) {
	defer wg.Done()
	pruningMVCC(db, height)
}

func pruningMVCC(db dbm.DB, height int64) {
	setPruning(pruningStateStart)
	defer setPruning(pruningStateEnd)
	pruningFirst(db, height)
}

func pruningFirst(db dbm.DB, curHeight int64) {
	it := db.Iterator(mvccData, nil, true)
	defer it.Close()

	var mp map[string][]int64
	count := 0
	batch := db.NewBatch(true)
	for it.Rewind(); it.Valid(); it.Next() {
		if quit {
			//该处退出
			return
		}
		if mp == nil {
			mp = make(map[string][]int64, onceCount)
		}

		key, height, err := getKeyVersion(it.Key())
		if err != nil {
			continue
		}

		if curHeight < height+levelPruningHeight &&
			curHeight >= height+int64(pruneHeight) {
			mp[string(key)] = append(mp[string(key)], height)
			count++
		}
		if len(mp) >= onceCount-1 || count > onceScanCount {
			deleteOldKV(mp, curHeight, batch)
			mp = nil
			count = 0
		}
	}
	if len(mp) > 0 {
		deleteOldKV(mp, curHeight, batch)
		mp = nil
		_ = mp
	}
}

func deleteOldKV(mp map[string][]int64, curHeight int64, batch dbm.Batch) {
	if len(mp) == 0 {
		return
	}
	batch.Reset()
	for key, vals := range mp {
		if len(vals) > 1 && vals[1] != vals[0] { //防止相同高度时候出现的误删除
			for _, val := range vals[1:] { //从第二个开始判断
				if curHeight >= val+int64(pruneHeight) {
					batch.Delete(genKeyVersion([]byte(key), val)) // 删除老版本key
					if batch.ValueSize() > batchDataSize {
						batch.Write()
						batch.Reset()
					}
				}
			}
		}
		delete(mp, key)
	}
	batch.Write()
}

func genKeyVersion(key []byte, height int64) []byte {
	b := append([]byte{}, mvccData...)
	newkey := append(b, key...)
	newkey = append(newkey, []byte(".")...)
	newkey = append(newkey, pad(height)...)
	return newkey
}

func getKeyVersion(vsnKey []byte) ([]byte, int64, error) {
	if !bytes.Contains(vsnKey, mvccData) {
		return nil, 0, types.ErrSize
	}
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
	s := fmt.Sprintf("%020d", version)
	return []byte(s)
}

func isPruning() bool {
	return atomic.LoadInt32(&pruningState) == 1
}

func setPruning(state int32) {
	atomic.StoreInt32(&pruningState, state)
}
