// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvdb

import (
	"github.com/33cn/chain33/common"
	clog "github.com/33cn/chain33/common/log"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/store"
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
)

var klog = log.New("module", "kvdb")

// SetLogLevel set log level
func SetLogLevel(level string) {
	clog.SetLogLevel(level)
}

// DisableLog disable log output
func DisableLog() {
	klog.SetHandler(log.DiscardHandler())
}

func init() {
	drivers.Reg("kvdb", New)
}

// KVStore implementation
type KVStore struct {
	*drivers.BaseStore
	cache map[string]map[string]*types.KeyValue
}

// New KVStore module
func New(cfg *types.Store, sub []byte) queue.Module {
	bs := drivers.NewBaseStore(cfg)
	kvs := &KVStore{bs, make(map[string]map[string]*types.KeyValue)}
	bs.SetChild(kvs)
	return kvs
}

// Close KVStore module
func (kvs *KVStore) Close() {
	kvs.BaseStore.Close()
	klog.Info("store kvdb closed")
}

// Set kvs with statehash to KVStore
func (kvs *KVStore) Set(datas *types.StoreSet, sync bool) ([]byte, error) {
	hash := calcHash(datas)
	kvmap := make(map[string]*types.KeyValue)
	for _, kv := range datas.KV {
		kvmap[string(kv.Key)] = kv
	}
	kvs.save(kvmap)
	return hash, nil
}

// Get kvs with statehash from KVStore
func (kvs *KVStore) Get(datas *types.StoreGet) [][]byte {
	values := make([][]byte, len(datas.Keys))
	if kvmap, ok := kvs.cache[string(datas.StateHash)]; ok {
		for i := 0; i < len(datas.Keys); i++ {
			kv := kvmap[string(datas.Keys[i])]
			if kv != nil {
				values[i] = kv.Value
			}
		}
	} else {
		db := kvs.GetDB()
		for i := 0; i < len(datas.Keys); i++ {
			value, _ := db.Get(datas.Keys[i])
			if value != nil {
				values[i] = value
			}
		}
	}
	return values
}

// MemSet set kvs to the mem of KVStore
func (kvs *KVStore) MemSet(datas *types.StoreSet, sync bool) ([]byte, error) {
	if len(datas.KV) == 0 {
		klog.Info("store kv memset,use preStateHash as stateHash for kvset is null")
		kvmap := make(map[string]*types.KeyValue)
		kvs.cache[string(datas.StateHash)] = kvmap
		return datas.StateHash, nil
	}

	hash := calcHash(datas)
	kvmap := make(map[string]*types.KeyValue)
	for _, kv := range datas.KV {
		kvmap[string(kv.Key)] = kv
	}
	kvs.cache[string(hash)] = kvmap
	if len(kvs.cache) > 100 {
		klog.Error("too many items in cache")
	}
	return hash, nil
}

// Commit kvs in the mem of KVStore
func (kvs *KVStore) Commit(req *types.ReqHash) ([]byte, error) {
	kvmap, ok := kvs.cache[string(req.Hash)]
	if !ok {
		klog.Error("store kvdb commit", "err", types.ErrHashNotFound)
		return nil, types.ErrHashNotFound
	}
	if len(kvmap) == 0 {
		klog.Info("store kvdb commit did nothing for kvset is nil")
		delete(kvs.cache, string(req.Hash))
		return req.Hash, nil
	}
	kvs.save(kvmap)
	delete(kvs.cache, string(req.Hash))
	return req.Hash, nil
}

// MemSetUpgrade set kvs to the mem of KVStore
func (kvs *KVStore) MemSetUpgrade(datas *types.StoreSet, sync bool) ([]byte, error) {
	//not support
	return nil, nil
}

// CommitUpgrade kvs in the mem of KVStore
func (kvs *KVStore) CommitUpgrade(req *types.ReqHash) ([]byte, error) {
	//not support
	return nil, nil
}

// Rollback kvs in the mem of KVStore
func (kvs *KVStore) Rollback(req *types.ReqHash) ([]byte, error) {
	_, ok := kvs.cache[string(req.Hash)]
	if !ok {
		klog.Error("store kvdb rollback", "err", types.ErrHashNotFound)
		return nil, types.ErrHashNotFound
	}
	delete(kvs.cache, string(req.Hash))
	return req.Hash, nil
}

// IterateRangeByStateHash method
func (kvs *KVStore) IterateRangeByStateHash(statehash []byte, start []byte, end []byte, ascending bool, fn func(key, value []byte) bool) {
	panic("empty")
	//TODO:
	//kvs.IterateRangeByStateHash(mavls.GetDB(), statehash, start, end, ascending, fn)
}

// ProcEvent handles supported events
func (kvs *KVStore) ProcEvent(msg *queue.Message) {
	if msg == nil {
		return
	}
	msg.ReplyErr("KVStore", types.ErrActionNotSupport)
}

// Del set kvs to nil with StateHash
func (kvs *KVStore) Del(req *types.StoreDel) ([]byte, error) {
	//not support
	return nil, nil
}

func (kvs *KVStore) save(kvmap map[string]*types.KeyValue) {
	storeBatch := kvs.GetDB().NewBatch(true)
	for _, kv := range kvmap {
		if kv.Value == nil {
			storeBatch.Delete(kv.Key)
		} else {
			storeBatch.Set(kv.Key, kv.Value)
		}
	}
	storeBatch.Write()
}

func calcHash(datas proto.Message) []byte {
	b := types.Encode(datas)
	return common.Sha256(b)
}
