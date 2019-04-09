// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvmvccdb

import (
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	clog "github.com/33cn/chain33/common/log"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/store"
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
)

var klog = log.New("module", "kvmvccdb")
var maxRollbackNum = 200

// SetLogLevel set log level
func SetLogLevel(level string) {
	clog.SetLogLevel(level)
}

// DisableLog disable log output
func DisableLog() {
	klog.SetHandler(log.DiscardHandler())
}

func init() {
	drivers.Reg("kvmvcc", New)
}

// KVMVCCStore provide kvmvcc store interface implementation
type KVMVCCStore struct {
	*drivers.BaseStore
	mvcc           dbm.MVCC
	kvsetmap       map[string][]*types.KeyValue
	enableMVCCIter bool
	sync           bool
}

type subConfig struct {
	EnableMVCCIter bool `json:"enableMVCCIter"`
}

// New construct KVMVCCStore module
func New(cfg *types.Store, sub []byte) queue.Module {
	bs := drivers.NewBaseStore(cfg)
	var kvs *KVMVCCStore
	enable := false
	if sub != nil {
		var subcfg subConfig
		types.MustDecode(sub, &subcfg)
		enable = subcfg.EnableMVCCIter
	}
	if enable {
		kvs = &KVMVCCStore{bs, dbm.NewMVCCIter(bs.GetDB()), make(map[string][]*types.KeyValue), true, false}
	} else {
		kvs = &KVMVCCStore{bs, dbm.NewMVCC(bs.GetDB()), make(map[string][]*types.KeyValue), false, false}
	}
	bs.SetChild(kvs)
	return kvs
}

// Close the KVMVCCStore module
func (mvccs *KVMVCCStore) Close() {
	mvccs.BaseStore.Close()
	klog.Info("store kvdb closed")
}

// Set kvs with statehash to KVMVCCStore
func (mvccs *KVMVCCStore) Set(datas *types.StoreSet, sync bool) ([]byte, error) {
	hash := calcHash(datas)
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
		klog.Error("Get version by hash failed.", "hash", common.ToHex(datas.StateHash))
		return values
	}
	for i := 0; i < len(datas.Keys); i++ {
		value, err := mvccs.mvcc.GetV(datas.Keys[i], version)
		if err != nil {
			klog.Error("GetV by Keys failed.", "Key", string(datas.Keys[i]), "version", version)
		} else if value != nil {
			values[i] = value
		}
	}
	return values
}

// MemSet set kvs to the mem of KVMVCCStore module and return the StateHash
func (mvccs *KVMVCCStore) MemSet(datas *types.StoreSet, sync bool) ([]byte, error) {
	beg := types.Now()
	defer func() {
		klog.Info("kvmvcc MemSet", "cost", types.Since(beg))
	}()
	kvset, err := mvccs.checkVersion(datas.Height)
	if err != nil {
		return nil, err
	}
	hash := calcHash(datas)
	//klog.Debug("KVMVCCStore MemSet AddMVCC", "prestatehash", common.ToHex(datas.StateHash), "hash", common.ToHex(hash), "height", datas.Height)
	kvlist, err := mvccs.mvcc.AddMVCC(datas.KV, hash, datas.StateHash, datas.Height)
	if err != nil {
		return nil, err
	}
	if len(kvlist) > 0 {
		kvset = append(kvset, kvlist...)
	}
	mvccs.sync = sync
	mvccs.kvsetmap[string(hash)] = kvset
	return hash, nil
}

// Commit kvs in the mem of KVMVCCStore module to state db and return the StateHash
func (mvccs *KVMVCCStore) Commit(req *types.ReqHash) ([]byte, error) {
	beg := types.Now()
	defer func() {
		klog.Info("kvmvcc Commit", "cost", types.Since(beg))
	}()
	_, ok := mvccs.kvsetmap[string(req.Hash)]
	if !ok {
		klog.Error("store kvmvcc commit", "err", types.ErrHashNotFound)
		return nil, types.ErrHashNotFound
	}
	//klog.Debug("KVMVCCStore Commit saveKVSets", "hash", common.ToHex(req.Hash))
	mvccs.saveKVSets(mvccs.kvsetmap[string(req.Hash)], mvccs.sync)
	delete(mvccs.kvsetmap, string(req.Hash))
	return req.Hash, nil
}

// MemSetUpgrade set kvs to the mem of KVMVCCStore module and return the StateHash
func (mvccs *KVMVCCStore) MemSetUpgrade(datas *types.StoreSet, sync bool) ([]byte, error) {
	//not support
	return nil, nil
}

// CommitUpgrade kvs in the mem of KVMVCCStore module to state db and return the StateHash
func (mvccs *KVMVCCStore) CommitUpgrade(req *types.ReqHash) ([]byte, error) {
	//not support
	return nil, nil
}

// Rollback kvs in the mem of KVMVCCStore module and return the StateHash
func (mvccs *KVMVCCStore) Rollback(req *types.ReqHash) ([]byte, error) {
	_, ok := mvccs.kvsetmap[string(req.Hash)]
	if !ok {
		klog.Error("store kvmvcc rollback", "err", types.ErrHashNotFound)
		return nil, types.ErrHashNotFound
	}

	//klog.Debug("KVMVCCStore Rollback", "hash", common.ToHex(req.Hash))

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
		klog.Error("KVMVCCStore IterateRangeByStateHash can't get max version, ignore the call.", "err", err)
		return
	}

	version, err := mvccs.mvcc.GetVersion(statehash)
	if err != nil {
		klog.Error("KVMVCCStore IterateRangeByStateHash can't get version, ignore the call.", "stateHash", common.ToHex(statehash), "err", err)
		return
	}

	if version != maxVersion {
		klog.Error("KVMVCCStore IterateRangeByStateHash call failed for maxVersion does not match version.", "maxVersion", maxVersion, "version", version, "stateHash", common.ToHex(statehash))
		return
	}

	//klog.Info("KVMVCCStore do the IterateRangeByStateHash")
	listhelper := dbm.NewListHelper(mvccs.mvcc.(*dbm.MVCCIter))
	listhelper.IteratorCallback(start, end, 0, 1, fn)
}

// ProcEvent handles supported events
func (mvccs *KVMVCCStore) ProcEvent(msg *queue.Message) {
	if msg == nil {
		return
	}
	msg.ReplyErr("KVStore", types.ErrActionNotSupport)
}

// Del set kvs to nil with StateHash
func (mvccs *KVMVCCStore) Del(req *types.StoreDel) ([]byte, error) {
	kvset, err := mvccs.mvcc.DelMVCC(req.StateHash, req.Height, true)
	if err != nil {
		klog.Error("store kvmvcc del", "err", err)
		return nil, err
	}

	klog.Info("KVMVCCStore Del", "hash", common.ToHex(req.StateHash), "height", req.Height)
	mvccs.saveKVSets(kvset, mvccs.sync)
	return req.StateHash, nil
}

func (mvccs *KVMVCCStore) saveKVSets(kvset []*types.KeyValue, sync bool) {
	if len(kvset) == 0 {
		return
	}

	storeBatch := mvccs.GetDB().NewBatch(sync)

	for i := 0; i < len(kvset); i++ {
		if kvset[i].Value == nil {
			storeBatch.Delete(kvset[i].Key)
		} else {
			storeBatch.Set(kvset[i].Key, kvset[i].Value)
		}
	}
	err := storeBatch.Write()
	if err != nil {
		klog.Error("store kvmvcc saveKVSets to db failed")
	}
}

func (mvccs *KVMVCCStore) checkVersion(height int64) ([]*types.KeyValue, error) {
	//检查新加入区块的height和现有的version的关系，来判断是否要回滚数据
	maxVersion, err := mvccs.mvcc.GetMaxVersion()
	if err != nil {
		if err != types.ErrNotFound {
			klog.Error("store kvmvcc checkVersion GetMaxVersion failed", "err", err)
			panic(err)
		} else {
			maxVersion = -1
		}
	}

	//klog.Debug("store kvmvcc checkVersion ", "maxVersion", maxVersion, "currentVersion", height)

	var kvset []*types.KeyValue
	if maxVersion < height-1 {
		klog.Error("store kvmvcc checkVersion found statehash lost", "maxVersion", maxVersion, "height", height)
		return nil, ErrStateHashLost
	} else if maxVersion == height-1 {
		return nil, nil
	} else {
		count := 1
		for i := maxVersion; i >= height; i-- {
			hash, err := mvccs.mvcc.GetVersionHash(i)
			if err != nil {
				klog.Warn("store kvmvcc checkVersion GetVersionHash failed", "height", i, "maxVersion", maxVersion)
				continue
			}
			kvlist, err := mvccs.mvcc.DelMVCC(hash, i, false)
			if err != nil {
				klog.Warn("store kvmvcc checkVersion DelMVCC failed", "height", i, "err", err)
				continue
			}
			kvset = append(kvset, kvlist...)

			klog.Debug("store kvmvcc checkVersion DelMVCC4Height", "height", i, "maxVersion", maxVersion)
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
