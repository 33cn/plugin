// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvmvccMavl

import (
	clog "github.com/33cn/chain33/common/log"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/store"
	"github.com/33cn/chain33/types"
	"github.com/hashicorp/golang-lru"
	"errors"
)

var (
	kmlog = log.New("module", "kvmvccMavl")
	ErrStateHashLost = errors.New("ErrStateHashLost")
	kvmvccMavlFork int64 = 200 * 10000
)

const (
	canceSize = 2048 //可以缓存2048个roothash, height对
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
	drivers.Reg("kvmvccMavl", New)
}

// KVMVCCMavlStore provide kvmvcc and mavl store interface implementation
type KVMVCCMavlStore struct {
	*drivers.BaseStore
	*KVMVCCStore
	*MavlStore
	cance *lru.Cache

}

type subKVMVCCConfig struct {
	EnableMVCCIter bool `json:"enableMVCCIter"`
}

type subMavlConfig struct {
	EnableMavlPrefix bool  `json:"enableMavlPrefix"`
	EnableMVCC       bool  `json:"enableMVCC"`
	EnableMavlPrune  bool  `json:"enableMavlPrune"`
	PruneHeight      int32 `json:"pruneHeight"`
}

type subConfig struct {
	EnableMVCCIter   bool  `json:"enableMVCCIter"`
	EnableMavlPrefix bool  `json:"enableMavlPrefix"`
	EnableMVCC       bool  `json:"enableMVCC"`
	EnableMavlPrune  bool  `json:"enableMavlPrune"`
	PruneHeight      int32 `json:"pruneHeight"`
}

// New construct KVMVCCStore module
func New(cfg *types.Store, sub []byte) queue.Module {
	bs := drivers.NewBaseStore(cfg)
	var kvms *KVMVCCMavlStore
	var subcfg subConfig
	var subKVMVCCcfg subKVMVCCConfig
	var subMavlcfg subMavlConfig
	if sub != nil {
		types.MustDecode(sub, &subcfg)
		subKVMVCCcfg.EnableMVCCIter = subcfg.EnableMVCCIter
		subMavlcfg.EnableMavlPrefix = subcfg.EnableMavlPrefix
		subMavlcfg.EnableMVCC       = subcfg.EnableMVCC
		subMavlcfg.EnableMavlPrune  = subcfg.EnableMavlPrune
		subMavlcfg.PruneHeight      = subcfg.PruneHeight
	}
	cance, err := lru.New(canceSize)
	if err != nil {
		panic("new KVMVCCMavlStore fail")
	}

	kvms = &KVMVCCMavlStore{bs, NewKVMVCC(&subKVMVCCcfg, bs.GetDB()),
		NewMavl(&subMavlcfg, bs.GetDB()), cance}
	bs.SetChild(kvms)
	return kvms
}

// Close the KVMVCCMavlStore module
func (kvmMavls *KVMVCCMavlStore) Close() {
	kvmMavls.BaseStore.Close()
	kvmMavls.KVMVCCStore.Close()
	kvmMavls.MavlStore.Close()
	kmlog.Info("store kvmMavls closed")
}

// Set kvs with statehash to KVMVCCMavlStore
func (kvmMavls *KVMVCCMavlStore) Set(datas *types.StoreSet, sync bool) ([]byte, error) {
	// 这里后续需要考虑分叉回退
	if datas.Height < kvmvccMavlFork {
		hash, err := kvmMavls.MavlStore.Set(datas, sync)
		if err != nil {
			return hash, err
		}
		_, err = kvmMavls.KVMVCCStore.Set(datas, hash, sync)
		if err != nil {
			return hash, err
		}
		if err == nil {
			kvmMavls.cance.Add(string(hash), datas.Height)
		}
		return hash, err
	}
	// 仅仅做kvmvcc
	hash, err := kvmMavls.KVMVCCStore.Set(datas, nil, sync)
	if err == nil {
		kvmMavls.cance.Add(string(hash), datas.Height)
	}
	return hash, err
}

// Get kvs with statehash from KVMVCCMavlStore
func (kvmMavls *KVMVCCMavlStore) Get(datas *types.StoreGet) [][]byte {
	if value, ok := kvmMavls.cance.Get(string(datas.StateHash)); ok {
		if value.(int64) < kvmvccMavlFork  {
			return kvmMavls.MavlStore.Get(datas)
		}
		return kvmMavls.KVMVCCStore.Get(datas)
	}
	return kvmMavls.KVMVCCStore.Get(datas)
}

// MemSet set kvs to the mem of KVMVCCMavlStore module and return the StateHash
func (kvmMavls *KVMVCCMavlStore) MemSet(datas *types.StoreSet, sync bool) ([]byte, error) {
	// 这里后续需要考虑分叉回退
	if datas.Height < kvmvccMavlFork {
		hash, err := kvmMavls.MavlStore.MemSet(datas, sync)
		if err != nil {
			return hash, err
		}
		_, err = kvmMavls.KVMVCCStore.MemSet(datas, hash, sync)
		if err != nil {
			return hash, err
		}
		if err == nil {
			kvmMavls.cance.Add(string(hash), datas.Height)
		}
		return hash, err
	}
	// 仅仅做kvmvcc
	hash, err := kvmMavls.KVMVCCStore.MemSet(datas, nil, sync)
	if err == nil {
		kvmMavls.cance.Add(string(hash), datas.Height)
	}
	return hash, err
}

// Commit kvs in the mem of KVMVCCMavlStore module to state db and return the StateHash
func (kvmMavls *KVMVCCMavlStore) Commit(req *types.ReqHash) ([]byte, error) {
	if value, ok := kvmMavls.cance.Get(string(req.Hash)); ok {
		if value.(int64) < kvmvccMavlFork {
			hash, err :=  kvmMavls.MavlStore.Commit(req)
			if err != nil {
				return hash, err
			}
			_, err = kvmMavls.KVMVCCStore.Commit(req)
			if err != nil {
				return hash, err
			}
			return hash, err
		}
		return kvmMavls.KVMVCCStore.Commit(req)
	}
	return kvmMavls.KVMVCCStore.Commit(req)
}

// Rollback kvs in the mem of KVMVCCMavlStore module and return the StateHash
func (kvmMavls *KVMVCCMavlStore) Rollback(req *types.ReqHash) ([]byte, error) {
	if value, ok := kvmMavls.cance.Get(string(req.Hash)); ok {
		if value.(int64) < kvmvccMavlFork  {
			hash, err :=  kvmMavls.MavlStore.Rollback(req)
			if err != nil {
				return hash, err
			}
			_, err = kvmMavls.KVMVCCStore.Rollback(req)
			if err != nil {
				return hash, err
			}
			return hash, err
		}
		return kvmMavls.KVMVCCStore.Rollback(req)
	}
	return kvmMavls.KVMVCCStore.Rollback(req)
}

// IterateRangeByStateHash travel with Prefix by StateHash  to get the latest version kvs.
func (kvmMavls *KVMVCCMavlStore) IterateRangeByStateHash(statehash []byte, start []byte, end []byte, ascending bool, fn func(key, value []byte) bool) {
	if value, ok := kvmMavls.cance.Get(string(statehash)); ok {
		if value.(int64) < kvmvccMavlFork  {
			kvmMavls.MavlStore.IterateRangeByStateHash(statehash, start, end, ascending, fn)
			return
		}
		kvmMavls.KVMVCCStore.IterateRangeByStateHash(statehash, start, end, ascending, fn)
		return
	}
	kvmMavls.KVMVCCStore.IterateRangeByStateHash(statehash, start, end, ascending, fn)
	return
}

// ProcEvent handles supported events
func (kvmMavls *KVMVCCMavlStore) ProcEvent(msg queue.Message) {
	msg.ReplyErr("KVMVCCMavlStore", types.ErrActionNotSupport)
}

// Del set kvs to nil with StateHash
func (kvmMavls *KVMVCCMavlStore) Del(req *types.StoreDel) ([]byte, error) {
	// 这里后续需要考虑分叉回退
	if req.Height < kvmvccMavlFork {
		hash, err := kvmMavls.MavlStore.Del(req)
		if err != nil {
			return hash, err
		}
		_, err = kvmMavls.KVMVCCStore.Del(req)
		if err != nil {
			return hash, err
		}
		if err == nil {
			kvmMavls.cance.Remove(string(req.StateHash))
		}
		return hash, err
	}
	// 仅仅做kvmvcc
	hash, err := kvmMavls.KVMVCCStore.Del(req)
	if err == nil {
		kvmMavls.cance.Remove(string(req.StateHash))
	}
	return hash, err
}