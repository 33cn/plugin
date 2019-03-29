// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvdb

import (
	"io/ioutil"
	"os"
	"testing"

	drivers "github.com/33cn/chain33/system/store"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

func newStoreCfg(dir string) *types.Store {
	return &types.Store{Name: "kvdb_test", Driver: "leveldb", DbPath: dir, DbCache: 100}
}

func TestKvdbNewClose(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil)
	assert.NotNil(t, store)

	store.Close()
}

func TestKvddbSetGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVStore)
	assert.NotNil(t, store)

	keys0 := [][]byte{[]byte("mk1"), []byte("mk2")}
	get0 := &types.StoreGet{StateHash: drivers.EmptyRoot[:], Keys: keys0}
	values0 := store.Get(get0)
	klog.Info("info", "info", values0)
	// Get exist key, result nil
	assert.Len(t, values0, 2)
	assert.Equal(t, []byte(nil), values0[0])
	assert.Equal(t, []byte(nil), values0[1])

	var kv []*types.KeyValue
	kv = append(kv, &types.KeyValue{Key: []byte("k1"), Value: []byte("v1")})
	kv = append(kv, &types.KeyValue{Key: []byte("k2"), Value: []byte("v2")})
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}
	hash, err := store.Set(datas, true)
	assert.Nil(t, err)
	keys := [][]byte{[]byte("k1"), []byte("k2")}
	get1 := &types.StoreGet{StateHash: hash, Keys: keys}

	values := store.Get(get1)
	assert.Len(t, values, 2)
	assert.Equal(t, []byte("v1"), values[0])
	assert.Equal(t, []byte("v2"), values[1])

	keys = [][]byte{[]byte("k1")}
	get2 := &types.StoreGet{StateHash: hash, Keys: keys}
	values2 := store.Get(get2)
	assert.Len(t, values2, 1)
	assert.Equal(t, []byte("v1"), values2[0])

	get3 := &types.StoreGet{StateHash: drivers.EmptyRoot[:], Keys: keys}
	values3 := store.Get(get3)
	assert.Len(t, values3, 1)
}

func TestKvdbMemSet(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVStore)
	assert.NotNil(t, store)

	var kv []*types.KeyValue
	kv = append(kv, &types.KeyValue{Key: []byte("mk1"), Value: []byte("v1")})
	kv = append(kv, &types.KeyValue{Key: []byte("mk2"), Value: []byte("v2")})
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}
	hash, err := store.MemSet(datas, true)
	assert.Nil(t, err)
	keys := [][]byte{[]byte("mk1"), []byte("mk2")}
	get1 := &types.StoreGet{StateHash: hash, Keys: keys}

	values := store.Get(get1)
	assert.Len(t, values, 2)

	actHash, _ := store.Commit(&types.ReqHash{Hash: hash})
	assert.Equal(t, hash, actHash)

	notExistHash, _ := store.Commit(&types.ReqHash{Hash: drivers.EmptyRoot[:]})
	assert.Nil(t, notExistHash)
}

func TestKvdbMemSetUpgrade(t *testing.T) {
	// not support
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVStore)
	assert.NotNil(t, store)
	store.MemSetUpgrade(nil, false)
}

func TestKvdbCommitUpgrade(t *testing.T) {
	// not support
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVStore)
	assert.NotNil(t, store)
	store.CommitUpgrade(nil)
}

func TestKvdbRollback(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVStore)
	assert.NotNil(t, store)

	var kv []*types.KeyValue
	kv = append(kv, &types.KeyValue{Key: []byte("mk1"), Value: []byte("v1")})
	kv = append(kv, &types.KeyValue{Key: []byte("mk2"), Value: []byte("v2")})
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}
	hash, err := store.MemSet(datas, true)
	assert.Nil(t, err)
	keys := [][]byte{[]byte("mk1"), []byte("mk2")}
	get1 := &types.StoreGet{StateHash: hash, Keys: keys}
	values := store.Get(get1)
	assert.Len(t, values, 2)

	actHash, _ := store.Rollback(&types.ReqHash{Hash: hash})
	assert.Equal(t, hash, actHash)

	notExistHash, _ := store.Rollback(&types.ReqHash{Hash: drivers.EmptyRoot[:]})
	assert.Nil(t, notExistHash)
}
