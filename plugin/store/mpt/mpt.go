// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mpt

import (
	"github.com/33cn/chain33/common"
	clog "github.com/33cn/chain33/common/log"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/store"
	"github.com/33cn/chain33/types"
	mpt "github.com/33cn/plugin/plugin/store/mpt/db"
	lru "github.com/hashicorp/golang-lru"
)

var mlog = log.New("module", "mpt")

// SetLogLevel set log level
func SetLogLevel(level string) {
	clog.SetLogLevel(level)
}

// DisableLog disable log
func DisableLog() {
	mlog.SetHandler(log.DiscardHandler())
}

// Store mpt store struct
type Store struct {
	*drivers.BaseStore
	trees map[string]*mpt.TrieEx
	cache *lru.Cache
}

func init() {
	drivers.Reg("mpt", New)
}

// New new mpt store module
func New(cfg *types.Store, sub []byte) queue.Module {
	bs := drivers.NewBaseStore(cfg)
	mpts := &Store{bs, make(map[string]*mpt.TrieEx), nil}
	mpts.cache, _ = lru.New(10)
	bs.SetChild(mpts)
	return mpts
}

// Close close mpt store
func (mpts *Store) Close() {
	mpts.BaseStore.Close()
	mlog.Info("store mavl closed")
}

// Set set k v to mpt store db; sync is true represent write sync
func (mpts *Store) Set(datas *types.StoreSet, sync bool) ([]byte, error) {
	hash, err := mpt.SetKVPair(mpts.GetDB(), datas, sync)
	if err != nil {
		mlog.Error("mpt store error", "err", err)
		return nil, err
	}
	return hash, nil
}

// Get get values by keys
func (mpts *Store) Get(datas *types.StoreGet) [][]byte {
	var tree *mpt.TrieEx
	var err error
	values := make([][]byte, len(datas.Keys))
	search := string(datas.StateHash)
	if data, ok := mpts.cache.Get(search); ok {
		tree = data.(*mpt.TrieEx)
	} else if data, ok := mpts.trees[search]; ok {
		tree = data
	} else {
		tree, err = mpt.NewEx(common.BytesToHash(datas.StateHash), mpt.NewDatabase(mpts.GetDB()))
		if nil != err {
			mlog.Error("Store get can not find a trie")
		}
		if nil == err {
			mpts.cache.Add(search, tree)
		}
		mlog.Debug("store mpt get tree", "err", err, "StateHash", common.ToHex(datas.StateHash))
	}
	if err == nil {
		for i := 0; i < len(datas.Keys); i++ {
			value, err := tree.TryGet(datas.Keys[i])
			if nil == err {
				values[i] = value
			}
		}
	}
	return values
}

// MemSet set keys values to memcory mpt, return root hash and error
func (mpts *Store) MemSet(datas *types.StoreSet, sync bool) ([]byte, error) {
	var err error
	var tree *mpt.TrieEx
	tree, err = mpt.NewEx(common.BytesToHash(datas.StateHash), mpt.NewDatabase(mpts.GetDB()))
	if err != nil {
		mlog.Info("MemSet create a new trie", "err", err)
		return nil, err
	}
	for i := 0; i < len(datas.KV); i++ {
		tree.Update(datas.KV[i].Key, datas.KV[i].Value)
	}
	root, err := tree.Commit(nil)
	if err != nil {
		mlog.Error("MemSet Commit to memory trie fail")
		return nil, err
	}
	hash := root[:]
	mpts.trees[string(hash)] = tree
	if len(mpts.trees) > 1000 {
		mlog.Error("too many trees in cache")
	}
	return hash, nil
}

// Commit convert memcory mpt to storage db
func (mpts *Store) Commit(req *types.ReqHash) ([]byte, error) {
	tree, ok := mpts.trees[string(req.Hash)]
	if !ok {
		mlog.Error("store mpt commit", "err", types.ErrHashNotFound)
		return nil, types.ErrHashNotFound
	}
	err := tree.Commit2Db(common.BytesToHash(req.Hash), true)
	if nil != err {
		mlog.Error("store mpt commit", "err", types.ErrHashNotFound)
		return nil, types.ErrDataBaseDamage
	}
	delete(mpts.trees, string(req.Hash))
	return req.Hash, nil
}

// MemSetUpgrade set keys values to memcory mpt, return root hash and error
func (mpts *Store) MemSetUpgrade(datas *types.StoreSet, sync bool) ([]byte, error) {
	//not support
	return nil, nil
}

// CommitUpgrade convert memcory mpt to storage db
func (mpts *Store) CommitUpgrade(req *types.ReqHash) ([]byte, error) {
	//not support
	return nil, nil
}

// Rollback 回退将缓存的mpt树删除掉
func (mpts *Store) Rollback(req *types.ReqHash) ([]byte, error) {
	_, ok := mpts.trees[string(req.Hash)]
	if !ok {
		mlog.Error("store mavl rollback", "err", types.ErrHashNotFound)
		return nil, types.ErrHashNotFound
	}
	delete(mpts.trees, string(req.Hash))
	return req.Hash, nil
}

// Del ...
func (mpts *Store) Del(req *types.StoreDel) ([]byte, error) {
	//not support
	return nil, nil
}

// IterateRangeByStateHash 迭代实现功能； statehash：当前状态hash, start：开始查找的key, end: 结束的key, ascending：升序，降序, fn 迭代回调函数
func (mpts *Store) IterateRangeByStateHash(statehash []byte, start []byte, end []byte, ascending bool, fn func(key, value []byte) bool) {
	mpt.IterateRangeByStateHash(mpts.GetDB(), statehash, start, end, ascending, fn)
}

// ProcEvent not support message
func (mpts *Store) ProcEvent(msg *queue.Message) {
	if msg == nil {
		return
	}
	msg.ReplyErr("Store", types.ErrActionNotSupport)
}
