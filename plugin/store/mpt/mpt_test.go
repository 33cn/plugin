// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mpt

import (
	"io/ioutil"
	"os"
	"testing"

	"fmt"
	"time"

	"github.com/33cn/chain33/common"
	drivers "github.com/33cn/chain33/system/store"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

const MaxKeylenth int = 64

func newStoreCfg(dir string) *types.Store {
	return &types.Store{Name: "mpt_test", Driver: "leveldb", DbPath: dir, DbCache: 100}
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
	store := New(storeCfg, nil).(*Store)
	assert.NotNil(t, store)

	keys0 := [][]byte{[]byte("mk1"), []byte("mk2")}
	get0 := &types.StoreGet{StateHash: drivers.EmptyRoot[:], Keys: keys0}
	values0 := store.Get(get0)
	mlog.Info("info", "info", values0)
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
	assert.Equal(t, []byte(nil), values3[0])
}

func TestKvdbMemSet(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*Store)
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

func TestKvmvccdbMemSetUpgrade(t *testing.T) {
	// not support
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*Store)
	assert.NotNil(t, store)
	store.MemSetUpgrade(nil, false)
}

func TestKvmvccdbCommitUpgrade(t *testing.T) {
	// not support
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*Store)
	assert.NotNil(t, store)
	store.CommitUpgrade(nil)
}

func TestKvdbRollback(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*Store)
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

var checkKVResult []*types.KeyValue

func checkKV(k, v []byte) bool {
	checkKVResult = append(checkKVResult,
		&types.KeyValue{Key: k, Value: v})
	//mlog.Debug("checkKV", "key", string(k), "value", string(v))
	return false
}

func GetRandomString(length int) string {
	return common.GetRandPrintString(20, length)
}

//目前正常情况下get 都是一个一个获取的
func BenchmarkGet(b *testing.B) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*Store)
	assert.NotNil(b, store)

	var kv []*types.KeyValue
	var keys [][]byte
	var hash = drivers.EmptyRoot[:]
	for i := 0; i < b.N; i++ {
		key := GetRandomString(MaxKeylenth)
		value := fmt.Sprintf("%s%d", key, i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{Key: []byte(string(key)), Value: []byte(string(value))})
		if i%10000 == 0 {
			datas := &types.StoreSet{StateHash: hash, KV: kv, Height: 0}
			hash, err = store.Set(datas, true)
			assert.Nil(b, err)
			kv = nil
		}
	}
	if kv != nil {
		datas := &types.StoreSet{StateHash: hash, KV: kv, Height: 0}
		hash, err = store.Set(datas, true)
		assert.Nil(b, err)
	}
	start := time.Now()
	b.ResetTimer()
	for _, key := range keys {
		getData := &types.StoreGet{
			StateHash: hash,
			Keys:      [][]byte{key}}
		store.Get(getData)
	}
	end := time.Now()
	fmt.Println("mpt BenchmarkGet cost time is", end.Sub(start), "num is", b.N)
}

func BenchmarkSet(b *testing.B) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*Store)
	assert.NotNil(b, store)

	var kv []*types.KeyValue
	var keys [][]byte
	var hash = drivers.EmptyRoot[:]
	start := time.Now()
	for i := 0; i < b.N; i++ {
		key := GetRandomString(MaxKeylenth)
		value := fmt.Sprintf("%s%d", key, i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{Key: []byte(string(key)), Value: []byte(string(value))})
		if i%10000 == 0 {
			datas := &types.StoreSet{StateHash: hash, KV: kv, Height: 0}
			hash, err = store.Set(datas, true)
			assert.Nil(b, err)
			kv = nil
		}
	}
	if kv != nil {
		datas := &types.StoreSet{StateHash: hash, KV: kv, Height: 0}
		_, err = store.Set(datas, true)
		assert.Nil(b, err)
	}
	end := time.Now()
	fmt.Println("mpt BenchmarkSet cost time is", end.Sub(start), "num is", b.N)
}

func BenchmarkMemSet(b *testing.B) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*Store)
	assert.NotNil(b, store)

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < b.N; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{Key: []byte(string(key)), Value: []byte(string(value))})
	}
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}
	start := time.Now()
	b.ResetTimer()
	hash, err := store.MemSet(datas, true)
	assert.Nil(b, err)
	assert.NotNil(b, hash)
	end := time.Now()
	fmt.Println("mpt BenchmarkMemSet cost time is", end.Sub(start), "num is", b.N)
}

func BenchmarkCommit(b *testing.B) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*Store)
	assert.NotNil(b, store)

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < b.N; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{Key: []byte(string(key)), Value: []byte(string(value))})
	}
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}
	hash, err := store.MemSet(datas, true)
	assert.Nil(b, err)
	req := &types.ReqHash{
		Hash: hash,
	}

	start := time.Now()
	b.ResetTimer()
	_, err = store.Commit(req)
	assert.NoError(b, err, "NoError")
	end := time.Now()
	fmt.Println("mpt BenchmarkCommit cost time is", end.Sub(start), "num is", b.N)
	b.StopTimer()
}
