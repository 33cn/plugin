// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvmvccmavl

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"fmt"

	"bytes"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/store"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const MaxKeylenth int = 64

func newStoreCfg(dir string) *types.Store {
	return &types.Store{Name: "kvmvccMavl_test", Driver: "leveldb", DbPath: dir, DbCache: 100}
}

func newStoreCfgIter(dir string) (*types.Store, []byte) {
	return &types.Store{Name: "kvmvccMavl_test", Driver: "leveldb", DbPath: dir, DbCache: 100}, enableConfig()
}

func TestKvmvccMavlNewClose(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(t, store)

	store.Close()
}

func TestKvmvccMavlSetGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(t, store)

	kvmvccMavlFork = 50
	defer func() {
		kvmvccMavlFork = 200 * 10000
	}()
	hash := drivers.EmptyRoot[:]
	for i := 0; i < 100; i++ {
		var kvs []*types.KeyValue
		kvs = append(kvs, &types.KeyValue{Key: []byte(fmt.Sprintf("k%d", i)), Value: []byte(fmt.Sprintf("v%d", i))})
		kvs = append(kvs, &types.KeyValue{Key: []byte(fmt.Sprintf("key%d", i)), Value: []byte(fmt.Sprintf("value%d", i))})
		datas := &types.StoreSet{
			StateHash: hash,
			KV:        kvs,
			Height:    int64(i)}
		hash, err = store.Set(datas, true)
		assert.Nil(t, err)
		keys := [][]byte{[]byte(fmt.Sprintf("k%d", i)), []byte(fmt.Sprintf("key%d", i))}
		get := &types.StoreGet{StateHash: hash, Keys: keys}
		values := store.Get(get)
		assert.Len(t, values, 2)
		assert.Equal(t, []byte(fmt.Sprintf("v%d", i)), values[0])
		assert.Equal(t, []byte(fmt.Sprintf("value%d", i)), values[1])
	}
}

func TestKvmvccMavlMemSet(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(t, store)

	kvmvccMavlFork = 50
	defer func() {
		kvmvccMavlFork = 200 * 10000
	}()
	hash := drivers.EmptyRoot[:]
	for i := 0; i < 100; i++ {
		var kvs []*types.KeyValue
		kvs = append(kvs, &types.KeyValue{Key: []byte(fmt.Sprintf("k%d", i)), Value: []byte(fmt.Sprintf("v%d", i))})
		kvs = append(kvs, &types.KeyValue{Key: []byte(fmt.Sprintf("key%d", i)), Value: []byte(fmt.Sprintf("value%d", i))})
		datas := &types.StoreSet{
			StateHash: hash,
			KV:        kvs,
			Height:    int64(i)}

		hash, err = store.MemSet(datas, true)
		assert.Nil(t, err)
		actHash, _ := store.Commit(&types.ReqHash{Hash: hash})
		assert.Equal(t, hash, actHash)
		keys := [][]byte{[]byte(fmt.Sprintf("k%d", i)), []byte(fmt.Sprintf("key%d", i))}
		get := &types.StoreGet{StateHash: hash, Keys: keys}
		values := store.Get(get)
		assert.Len(t, values, 2)
		assert.Equal(t, []byte(fmt.Sprintf("v%d", i)), values[0])
		assert.Equal(t, []byte(fmt.Sprintf("value%d", i)), values[1])
	}
	notExistHash, _ := store.Commit(&types.ReqHash{Hash: drivers.EmptyRoot[:]})
	assert.Nil(t, notExistHash)
}

func TestKvmvccMavlMemSetUpgrade(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(t, store)

	kvmvccMavlFork = 50
	defer func() {
		kvmvccMavlFork = 200 * 10000
	}()
	hash := drivers.EmptyRoot[:]
	for i := 0; i < 1; i++ {
		var kvs []*types.KeyValue
		kvs = append(kvs, &types.KeyValue{Key: []byte(fmt.Sprintf("k%d", i)), Value: []byte(fmt.Sprintf("v%d", i))})
		kvs = append(kvs, &types.KeyValue{Key: []byte(fmt.Sprintf("key%d", i)), Value: []byte(fmt.Sprintf("value%d", i))})
		datas := &types.StoreSet{
			StateHash: hash,
			KV:        kvs,
			Height:    int64(i)}

		hash, err = store.MemSetUpgrade(datas, true)
		assert.Nil(t, err)
		actHash, _ := store.CommitUpgrade(&types.ReqHash{Hash: hash})
		assert.Equal(t, hash, actHash)
		keys := [][]byte{[]byte(fmt.Sprintf("k%d", i)), []byte(fmt.Sprintf("key%d", i))}
		get := &types.StoreGet{StateHash: hash, Keys: keys}
		values := store.Get(get)
		assert.Len(t, values, 2)
		assert.Equal(t, []byte(fmt.Sprintf("v%d", i)), values[0])
		assert.Equal(t, []byte(fmt.Sprintf("value%d", i)), values[1])
	}
	notExistHash, _ := store.CommitUpgrade(&types.ReqHash{Hash: drivers.EmptyRoot[:]})
	assert.Nil(t, notExistHash)
}

func TestKvmvccMavlCommit(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(t, store)

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < 30; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{Key: []byte(string(key)), Value: []byte(string(value))})
	}
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}

	// 设置分叉高度
	forkHeight := 100
	kvmvccMavlFork = int64(forkHeight)
	defer func() {
		kvmvccMavlFork = 200 * 10000
	}()
	frontHash := make([]byte, 0, 32)
	var hash []byte
	for i := 0; i < 200; i++ {
		datas.Height = int64(i)
		hash, err = store.MemSet(datas, true)
		assert.Nil(t, err)
		req := &types.ReqHash{
			Hash: hash,
		}
		if i+1 == forkHeight {
			frontHash = append(frontHash, hash...)
		}
		_, err = store.Commit(req)
		assert.NoError(t, err, "NoError")
		datas.StateHash = hash
	}

	if len(frontHash) > 0 {
		get := &types.StoreGet{StateHash: frontHash, Keys: keys}
		values := store.Get(get)
		require.Equal(t, len(values), len(keys))
		for i := range keys {
			require.Equal(t, kv[i].Value, values[i])
		}
	}

	if len(hash) > 0 {
		get := &types.StoreGet{StateHash: hash, Keys: keys}
		values := store.Get(get)
		require.Equal(t, len(values), len(keys))
		for i := range keys {
			require.Equal(t, kv[i].Value, values[i])
		}
	}
}

func TestKvmvccMavlRollback(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
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
	get := &types.StoreGet{StateHash: hash, Keys: keys}
	values := store.Get(get)
	assert.Len(t, values, 2)

	actHash, _ := store.Rollback(&types.ReqHash{Hash: hash})
	assert.Equal(t, hash, actHash)

	notExistHash, err := store.Rollback(&types.ReqHash{Hash: drivers.EmptyRoot[:]})
	assert.Nil(t, notExistHash)
	assert.Equal(t, types.ErrHashNotFound.Error(), err.Error())

	// 分叉之后
	kvmvccMavlFork = 1
	defer func() {
		kvmvccMavlFork = 200 * 10000
	}()

	hash, err = store.MemSet(datas, true)
	assert.Nil(t, err)
	actHash, _ = store.Commit(&types.ReqHash{Hash: hash})
	assert.Equal(t, hash, actHash)

	var kv1 []*types.KeyValue
	kv1 = append(kv1, &types.KeyValue{Key: []byte("mk3"), Value: []byte("v3")})
	kv1 = append(kv1, &types.KeyValue{Key: []byte("mk4"), Value: []byte("v4")})
	datas1 := &types.StoreSet{
		StateHash: hash,
		KV:        kv1,
		Height:    1}
	hash1, err := store.MemSet(datas1, true)
	assert.Nil(t, err)
	keys1 := [][]byte{[]byte("mk3"), []byte("mk4")}
	get1 := &types.StoreGet{StateHash: hash1, Keys: keys1}
	values1 := store.Get(get1)
	assert.Len(t, values1, 2)

	actHash, _ = store.Rollback(&types.ReqHash{Hash: hash1})
	assert.Equal(t, hash1, actHash)

	notExistHash, err = store.Rollback(&types.ReqHash{Hash: drivers.EmptyRoot[:]})
	assert.Nil(t, notExistHash)
	assert.Equal(t, types.ErrHashNotFound.Error(), err.Error())
}

func TestKvmvccdbRollbackBatch(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
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
	var kvset []*types.KeyValue
	req := &types.ReqHash{Hash: hash}
	hash1 := make([]byte, len(hash))
	copy(hash1, hash)
	store.Commit(req)
	// 设置分叉高度
	kvmvccMavlFork = 50
	defer func() {
		kvmvccMavlFork = 200 * 10000
	}()
	for i := 1; i <= 202; i++ {
		kvset = nil
		datas1 := &types.StoreSet{StateHash: hash1, KV: datas.KV, Height: datas.Height + int64(i)}
		s1 := fmt.Sprintf("v1-%03d", datas.Height+int64(i))
		s2 := fmt.Sprintf("v2-%03d", datas.Height+int64(i))
		datas.KV[0].Value = []byte(s1)
		datas.KV[1].Value = []byte(s2)
		hash1 = calcHash(datas1)
		//zzh
		//kmlog.Debug("KVMVCCStore MemSet AddMVCC", "prestatehash", common.ToHex(datas.StateHash), "hash", common.ToHex(hash), "height", datas.Height)
		kmlog.Info("KVMVCCStore MemSet AddMVCC for 202", "prestatehash", common.ToHex(datas1.StateHash), "hash", common.ToHex(hash1), "height", datas1.Height)
		kvlist, err := store.mvcc.AddMVCC(datas1.KV, hash1, datas1.StateHash, datas1.Height)
		if err != nil {
			kmlog.Info("KVMVCCStore MemSet AddMVCC failed for 202, continue")
			continue
		}

		if len(kvlist) > 0 {
			kvset = append(kvset, kvlist...)
		}
		store.kvsetmap[string(hash1)] = kvset
		req := &types.ReqHash{Hash: hash1}
		store.Commit(req)
	}

	maxVersion, err := store.mvcc.GetMaxVersion()
	assert.Equal(t, err, nil)
	assert.Equal(t, int64(202), maxVersion)

	keys := [][]byte{[]byte("mk1"), []byte("mk2")}
	get1 := &types.StoreGet{StateHash: hash, Keys: keys}
	values := store.Get(get1)
	assert.Len(t, values, 2)
	assert.Equal(t, []byte("v1"), values[0])
	assert.Equal(t, []byte("v2"), values[1])

	var kv2 []*types.KeyValue
	kv2 = append(kv2, &types.KeyValue{Key: []byte("mk1"), Value: []byte("v11")})
	kv2 = append(kv2, &types.KeyValue{Key: []byte("mk2"), Value: []byte("v22")})

	//触发批量回滚
	datas2 := &types.StoreSet{StateHash: hash, KV: kv2, Height: 1}
	hash, err = store.MemSet(datas2, true)
	assert.Nil(t, err)
	req = &types.ReqHash{Hash: hash}
	store.Commit(req)

	maxVersion, err = store.mvcc.GetMaxVersion()
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(3), maxVersion)

	get2 := &types.StoreGet{StateHash: hash, Keys: keys}
	values2 := store.Get(get2)
	assert.Len(t, values, 2)
	assert.Equal(t, values2[0], kv2[0].Value)
	assert.Equal(t, values2[1], kv2[1].Value)

	datas3 := &types.StoreSet{StateHash: hash, KV: kv2, Height: 2}
	hash, err = store.MemSet(datas3, true)
	assert.Nil(t, err)
	req = &types.ReqHash{Hash: hash}
	store.Commit(req)

	maxVersion, err = store.mvcc.GetMaxVersion()
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(2), maxVersion)
}

func enableConfig() []byte {
	data, _ := json.Marshal(&subConfig{EnableMVCCIter: true})
	return data
}

func TestIterateRangeByStateHash(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	storeCfg, sub := newStoreCfgIter(dir)
	store := New(storeCfg, sub).(*KVmMavlStore)
	assert.NotNil(t, store)

	execaddr := "0111vcBNSEA7fZhAdLJphDwQRQJa111"
	addr := "06htvcBNSEA7fZhAdLJphDwQRQJaHpy"
	addr1 := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
	addr2 := "26htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
	addr3 := "36htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
	addr4 := "46htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
	accCoin := account.NewCoinsAccount()

	account1 := &types.Account{
		Balance: 1000 * 1e8,
		Addr:    addr1,
	}

	account2 := &types.Account{
		Balance: 900 * 1e8,
		Addr:    addr2,
	}

	account3 := &types.Account{
		Balance: 800 * 1e8,
		Addr:    addr3,
	}

	account4 := &types.Account{
		Balance: 700 * 1e8,
		Addr:    addr4,
	}
	set1 := accCoin.GetKVSet(account1)
	set2 := accCoin.GetKVSet(account2)
	set3 := accCoin.GetKVSet(account3)
	set4 := accCoin.GetKVSet(account4)

	set5 := accCoin.GetExecKVSet(execaddr, account4)

	fmt.Println("---test case1-1 ---")
	var kv []*types.KeyValue
	kv = append(kv, &types.KeyValue{Key: set4[0].GetKey(), Value: set4[0].GetValue()})
	kv = append(kv, &types.KeyValue{Key: set3[0].GetKey(), Value: set3[0].GetValue()})
	kv = append(kv, &types.KeyValue{Key: set1[0].GetKey(), Value: set1[0].GetValue()})
	kv = append(kv, &types.KeyValue{Key: set2[0].GetKey(), Value: set2[0].GetValue()})
	kv = append(kv, &types.KeyValue{Key: set5[0].GetKey(), Value: set5[0].GetValue()})
	for i := 0; i < len(kv); i++ {
		fmt.Println("key:", string(kv[i].Key), "value:", string(kv[i].Value))
	}
	datas := &types.StoreSet{StateHash: drivers.EmptyRoot[:], KV: kv, Height: 0}
	hash, err := store.MemSet(datas, true)
	assert.Nil(t, err)
	var kvset []*types.KeyValue
	req := &types.ReqHash{Hash: hash}
	hash1 := make([]byte, len(hash))
	copy(hash1, hash)
	store.Commit(req)

	resp := &types.ReplyGetTotalCoins{}
	resp.Count = 100000

	store.IterateRangeByStateHash(hash, []byte("mavl-coins-bty-"), []byte("mavl-coins-bty-exec"), true, resp.IterateRangeByStateHash)
	fmt.Println("resp.Num=", resp.Num)
	fmt.Println("resp.Amount=", resp.Amount)

	assert.Equal(t, int64(4), resp.Num)
	assert.Equal(t, int64(340000000000), resp.Amount)

	// 设置分叉高度
	kvmvccMavlFork = 5
	defer func() {
		kvmvccMavlFork = 200 * 10000
	}()
	fmt.Println("---test case1-2 ---")
	firstForkHash := drivers.EmptyRoot[:]
	for i := 1; i <= 10; i++ {
		kvset = nil

		s1 := fmt.Sprintf("%03d", 11-i)
		addrx := addr + s1
		account := &types.Account{
			Balance: ((1000 + int64(i)) * 1e8),
			Addr:    addrx,
		}
		set := accCoin.GetKVSet(account)
		fmt.Println("key:", string(set[0].GetKey()), "value:", set[0].GetValue())
		kvset = append(kvset, &types.KeyValue{Key: set[0].GetKey(), Value: set[0].GetValue()})
		datas1 := &types.StoreSet{StateHash: hash1, KV: kvset, Height: datas.Height + int64(i)}
		hash1, err = store.MemSet(datas1, true)
		assert.Nil(t, err)
		req := &types.ReqHash{Hash: hash1}
		store.Commit(req)
		if int(kvmvccMavlFork) == i {
			firstForkHash = hash1
		}
	}

	resp = &types.ReplyGetTotalCoins{}
	resp.Count = 100000
	store.IterateRangeByStateHash(hash1, []byte("mavl-coins-bty-"), []byte("mavl-coins-bty-exec"), true, resp.IterateRangeByStateHash)
	fmt.Println("resp.Num=", resp.Num)
	fmt.Println("resp.Amount=", resp.Amount)
	assert.Equal(t, int64(14), resp.Num)
	assert.Equal(t, int64(1345500000000), resp.Amount)

	fmt.Println("---test case1-3 ---")

	resp = &types.ReplyGetTotalCoins{}
	resp.Count = 100000
	store.IterateRangeByStateHash(hash1, []byte("mavl-coins-bty-06htvcBNSEA7fZhAdLJphDwQRQJaHpy003"), []byte("mavl-coins-bty-exec"), true, resp.IterateRangeByStateHash)
	fmt.Println("resp.Num=", resp.Num)
	fmt.Println("resp.Amount=", resp.Amount)
	assert.Equal(t, int64(12), resp.Num)
	assert.Equal(t, int64(1143600000000), resp.Amount)

	fmt.Println("---test case1-4 ---")

	resp = &types.ReplyGetTotalCoins{}
	resp.Count = 2
	store.IterateRangeByStateHash(hash1, []byte("mavl-coins-bty-06htvcBNSEA7fZhAdLJphDwQRQJaHpy003"), []byte("mavl-coins-bty-exec"), true, resp.IterateRangeByStateHash)
	fmt.Println("resp.Num=", resp.Num)
	fmt.Println("resp.Amount=", resp.Amount)
	assert.Equal(t, int64(2), resp.Num)
	assert.Equal(t, int64(201500000000), resp.Amount)

	fmt.Println("---test case1-5 ---")

	resp = &types.ReplyGetTotalCoins{}
	resp.Count = 2
	store.IterateRangeByStateHash(hash1, []byte("mavl-coins-bty-"), []byte("mavl-coins-bty-exec"), true, resp.IterateRangeByStateHash)
	fmt.Println("resp.Num=", resp.Num)
	fmt.Println("resp.Amount=", resp.Amount)
	assert.Equal(t, int64(2), resp.Num)
	assert.Equal(t, int64(201900000000), resp.Amount)

	fmt.Println("---test case1-6 ---")

	resp = &types.ReplyGetTotalCoins{}
	resp.Count = 10000
	store.IterateRangeByStateHash(firstForkHash, []byte("mavl-coins-bty-"), []byte("mavl-coins-bty-exec"), true, resp.IterateRangeByStateHash)
	fmt.Println("resp.Num=", resp.Num)
	fmt.Println("resp.Amount=", resp.Amount)
	assert.Equal(t, int64(0), resp.Num)
	assert.Equal(t, int64(0), resp.Amount)
}

func TestProcEvent(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	storeCfg, sub := newStoreCfgIter(dir)
	store := New(storeCfg, sub).(*KVmMavlStore)
	assert.NotNil(t, store)

	store.ProcEvent(nil)
	store.ProcEvent(&queue.Message{})
}

func GetRandomString(length int) string {
	return common.GetRandPrintString(20, length)
}

func TestDelMavlData(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	storeCfg := newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(t, store)

	db := store.GetDB()

	db.Set([]byte(mvccPrefix), []byte("value1"))
	db.Set([]byte(fmt.Sprintf("%s123", mvccPrefix)), []byte("value2"))
	db.Set([]byte(fmt.Sprintf("%s546", mvccPrefix)), []byte("value3"))
	db.Set([]byte(fmt.Sprintf("123%s", mvccPrefix)), []byte("value4"))
	db.Set([]byte("key11"), []byte("value11"))
	db.Set([]byte("key22"), []byte("value22"))

	quit = false
	delMavlData(db)

	v, err := db.Get([]byte(mvccPrefix))
	require.NoError(t, err)
	require.Equal(t, []byte("value1"), v)
	v, err = db.Get([]byte(fmt.Sprintf("%s123", mvccPrefix)))
	require.NoError(t, err)
	require.Equal(t, []byte("value2"), v)
	v, err = db.Get([]byte(fmt.Sprintf("%s546", mvccPrefix)))
	require.NoError(t, err)
	require.Equal(t, []byte("value3"), v)
	_, err = db.Get([]byte(fmt.Sprintf("123%s", mvccPrefix)))
	require.Error(t, err)
	_, err = db.Get([]byte("key11"))
	require.Error(t, err)
	_, err = db.Get([]byte("key22"))
	require.Error(t, err)
	_, err = db.Get(genDelMavlKey(mvccPrefix))
	require.NoError(t, err)
}

func TestPruning(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	storeCfg := newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(t, store)

	kvmvccStore := NewKVMVCC(&subKVMVCCConfig{}, store.GetDB())

	SetPruneHeight(10)
	defer SetPruneHeight(0)

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < 30; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{Key: []byte(string(key)), Value: []byte(string(value))})
	}
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}

	var hashes [][]byte
	for i := 0; i < 100; i++ {
		datas.Height = int64(i)
		value = fmt.Sprintf("vv%d", i)
		for j := 0; j < 30; j++ {
			datas.KV[j].Value = []byte(value)
		}
		hash, err := kvmvccStore.MemSet(datas, nil, true)
		require.NoError(t, err)
		req := &types.ReqHash{
			Hash: hash,
		}
		_, err = kvmvccStore.Commit(req)
		require.NoError(t, err)
		datas.StateHash = hash
		hashes = append(hashes, hash)
	}

	pruningMVCC(store.GetDB(), 99)

	//check
	getDatas := &types.StoreGet{
		StateHash: drivers.EmptyRoot[:],
		Keys:      keys,
	}

	for i := 0; i < len(hashes); i++ {
		getDatas.StateHash = hashes[i]
		values := store.Get(getDatas)
		value = fmt.Sprintf("vv%d", i)

		if i < 80 {
			for _, v := range values {
				require.Equal(t, []byte(nil), v)
			}
		}

		if i > 90 {
			for _, v := range values {
				require.Equal(t, []byte(value), v)
			}
		}
	}
}

func TestGetKeyVersion(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	storeCfg := newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(t, store)

	mvcc := dbm.NewMVCC(store.GetDB())
	kvs := []*types.KeyValue{
		{Key: []byte("5"), Value: []byte("11")},
		{Key: []byte("123"), Value: []byte("111")},
		{Key: []byte(""), Value: []byte("1111")},
	}
	hash := []byte("12345678901234567890123456789012")
	vsnkv, err := mvcc.AddMVCC(kvs, hash, nil, 0)
	require.NoError(t, err)
	for _, kv := range vsnkv {
		if bytes.Contains(kv.Key, mvccData) && bytes.Contains(kv.Key, kvs[0].Key) {
			k, h, err := getKeyVersion(kv.Key)
			require.NoError(t, err)
			require.Equal(t, k, kvs[0].Key)
			require.Equal(t, h, int64(0))
			continue
		}
		if bytes.Contains(kv.Key, mvccData) && bytes.Contains(kv.Key, kvs[1].Key) {
			k, h, err := getKeyVersion(kv.Key)
			require.NoError(t, err)
			require.Equal(t, k, kvs[1].Key)
			require.Equal(t, h, int64(0))
			continue
		}
		if bytes.Contains(kv.Key, mvccData) {
			k, h, err := getKeyVersion(kv.Key)
			require.NoError(t, err)
			require.Equal(t, k, kvs[2].Key)
			require.Equal(t, h, int64(0))
		}
	}
}

func BenchmarkGetkmvccMavl(b *testing.B) { benchmarkGet(b, false) }
func BenchmarkGetkmvcc(b *testing.B)     { benchmarkGet(b, true) }

func benchmarkGet(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录

	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

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
		//kv = nil
	}
	assert.Nil(b, err)
	start := time.Now()
	b.ResetTimer()
	for _, key := range keys {
		getData := &types.StoreGet{
			StateHash: hash,
			Keys:      [][]byte{key}}
		store.Get(getData)
	}
	end := time.Now()
	fmt.Println("kvmvcc BenchmarkGet cost time is", end.Sub(start), "num is", b.N)
}

func BenchmarkStoreGetKvs4NkmvccMavl(b *testing.B) { benchmarkStoreGetKvs4N(b, false) }
func BenchmarkStoreGetKvs4Nkmvcc(b *testing.B)     { benchmarkStoreGetKvs4N(b, true) }

func benchmarkStoreGetKvs4N(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(b, store)

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	kvnum := 30
	for i := 0; i < kvnum; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{Key: []byte(string(key)), Value: []byte(string(value))})
	}
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}
	hash, err := store.Set(datas, true)
	assert.Nil(b, err)
	getData := &types.StoreGet{
		StateHash: hash,
		Keys:      keys}

	start := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values := store.Get(getData)
		assert.Len(b, values, kvnum)
	}

	end := time.Now()
	fmt.Println("kvmvcc BenchmarkStoreGetKvs4N cost time is", end.Sub(start), "num is", b.N)

	b.StopTimer()
}

func BenchmarkStoreGetKvsForNNkmvccMavl(b *testing.B) { benchmarkStoreGetKvsForNN(b, false) }
func BenchmarkStoreGetKvsForNNkmvcc(b *testing.B)     { benchmarkStoreGetKvsForNN(b, true) }

func benchmarkStoreGetKvsForNN(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录

	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < 30; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{Key: []byte(string(key)), Value: []byte(string(value))})
	}
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}

	var hashes [][]byte
	for i := 0; i < b.N; i++ {
		datas.Height = int64(i)
		value = fmt.Sprintf("vv%d", i)
		for j := 0; j < 30; j++ {
			datas.KV[j].Value = []byte(value)
		}
		hash, err := store.MemSet(datas, true)
		assert.Nil(b, err)
		req := &types.ReqHash{
			Hash: hash,
		}
		_, err = store.Commit(req)
		assert.NoError(b, err, "NoError")
		datas.StateHash = hash
		hashes = append(hashes, hash)
	}

	start := time.Now()
	b.ResetTimer()

	getData := &types.StoreGet{
		StateHash: hashes[0],
		Keys:      keys}

	for i := 0; i < b.N; i++ {
		getData.StateHash = hashes[i]
		store.Get(getData)
	}
	end := time.Now()
	fmt.Println("kvmvcc BenchmarkStoreGetKvsForNN cost time is", end.Sub(start), "num is", b.N)
	b.StopTimer()
}

func BenchmarkStoreGetKvsFor10000kmvccMavl(b *testing.B) { benchmarkStoreGetKvsFor10000(b, false) }
func BenchmarkStoreGetKvsFor10000kmvcc(b *testing.B)     { benchmarkStoreGetKvsFor10000(b, true) }

func benchmarkStoreGetKvsFor10000(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录

	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < 30; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{Key: []byte(string(key)), Value: []byte(string(value))})
	}
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}

	var hashes [][]byte
	blocks := 10000
	times := 10000
	start1 := time.Now()
	for i := 0; i < blocks; i++ {
		datas.Height = int64(i)
		value = fmt.Sprintf("vv%d", i)
		for j := 0; j < 30; j++ {
			datas.KV[j].Value = []byte(value)
		}
		hash, err := store.MemSet(datas, true)
		assert.Nil(b, err)
		req := &types.ReqHash{
			Hash: hash,
		}
		_, err = store.Commit(req)
		assert.NoError(b, err, "NoError")
		datas.StateHash = hash
		hashes = append(hashes, hash)
	}
	end1 := time.Now()

	start := time.Now()
	b.ResetTimer()

	getData := &types.StoreGet{
		StateHash: hashes[0],
		Keys:      keys}

	for i := 0; i < times; i++ {
		getData.StateHash = hashes[i]
		store.Get(getData)
	}
	end := time.Now()
	fmt.Println("kvmvcc BenchmarkStoreGetKvsFor10000 MemSet&Commit cost time is ", end1.Sub(start1), "blocks is", blocks)
	fmt.Println("kvmvcc BenchmarkStoreGetKvsFor10000 Get cost time is", end.Sub(start), "num is ", times, ",blocks is ", blocks)
	b.StopTimer()
}

func BenchmarkGetIterkmvccMavl(b *testing.B) { benchmarkGetIter(b, false) }
func BenchmarkGetIterkmvcc(b *testing.B)     { benchmarkGetIter(b, true) }

func benchmarkGetIter(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录

	storeCfg, sub := newStoreCfgIter(dir)
	store := New(storeCfg, sub).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

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
		//kv = nil
	}
	assert.Nil(b, err)
	start := time.Now()
	b.ResetTimer()
	for _, key := range keys {
		getData := &types.StoreGet{
			StateHash: hash,
			Keys:      [][]byte{key}}
		store.Get(getData)
	}
	end := time.Now()
	fmt.Println("kvmvcc BenchmarkGet cost time is", end.Sub(start), "num is", b.N)
}

func BenchmarkSetkmvccMavl(b *testing.B) { benchmarkSet(b, false) }
func BenchmarkSetkmvcc(b *testing.B)     { benchmarkSet(b, true) }

func benchmarkSet(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(b, store)
	b.Log(dir)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

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
		//kv = nil
	}
	end := time.Now()
	fmt.Println("mpt BenchmarkSet cost time is", end.Sub(start), "num is", b.N)
}

//上一个用例，一次性插入多对kv；本用例每次插入30对kv，分多次插入，测试性能表现。
func BenchmarkStoreSetkmvccMavl(b *testing.B) { benchmarkStoreSet(b, false) }
func BenchmarkStoreSetkmvcc(b *testing.B)     { benchmarkStoreSet(b, true) }

func benchmarkStoreSet(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < 30; i++ {
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
	for i := 0; i < b.N; i++ {
		hash, err := store.Set(datas, true)
		assert.Nil(b, err)
		assert.NotNil(b, hash)
	}
	end := time.Now()
	fmt.Println("kvmvcc BenchmarkSet cost time is", end.Sub(start), "num is", b.N)
}

func BenchmarkSetIterkmvccMavl(b *testing.B) { benchmarkSetIter(b, false) }
func BenchmarkSetIterkmvcc(b *testing.B)     { benchmarkSetIter(b, true) }

func benchmarkSetIter(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	storeCfg, sub := newStoreCfgIter(dir)
	store := New(storeCfg, sub).(*KVmMavlStore)
	assert.NotNil(b, store)
	b.Log(dir)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

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
		//kv = nil
	}
	end := time.Now()
	fmt.Println("kvmvcc BenchmarkSet cost time is", end.Sub(start), "num is", b.N)
}

//一次设定多对kv，测试一次的时间/多少对kv，来算平均一对kv的耗时。
func BenchmarkMemSetkmvccMavl(b *testing.B) { benchmarkMemSet(b, false) }
func BenchmarkMemSetkmvcc(b *testing.B)     { benchmarkMemSet(b, true) }

func benchmarkMemSet(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

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
	fmt.Println("kvmvcc BenchmarkMemSet cost time is", end.Sub(start), "num is", b.N)
}

//一次设定30对kv，设定N次，计算每次设定30对kv的耗时。
func BenchmarkStoreMemSetkmvccMavl(b *testing.B) { benchmarkStoreMemSet(b, false) }
func BenchmarkStoreMemSetkmvcc(b *testing.B)     { benchmarkStoreMemSet(b, true) }

func benchmarkStoreMemSet(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < 30; i++ {
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
	for i := 0; i < b.N; i++ {
		hash, err := store.MemSet(datas, true)
		assert.Nil(b, err)
		assert.NotNil(b, hash)
		req := &types.ReqHash{
			Hash: hash}
		store.Rollback(req)
	}
	end := time.Now()
	fmt.Println("kvmvcc BenchmarkStoreMemSet cost time is", end.Sub(start), "num is", b.N)
}

func BenchmarkCommitkmvccMavl(b *testing.B) { benchmarkCommit(b, false) }
func BenchmarkCommitkmvcc(b *testing.B)     { benchmarkCommit(b, true) }

func benchmarkCommit(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}
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
	req := &types.ReqHash{
		Hash: hash,
	}
	_, err = store.Commit(req)
	assert.NoError(b, err, "NoError")

	end := time.Now()
	fmt.Println("kvmvcc BenchmarkCommit cost time is", end.Sub(start), "num is", b.N)
	b.StopTimer()
}

func BenchmarkStoreCommitkmvccMavl(b *testing.B) { benchmarkStoreCommit(b, false) }
func BenchmarkStoreCommitkmvcc(b *testing.B)     { benchmarkStoreCommit(b, true) }

func benchmarkStoreCommit(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	var storeCfg = newStoreCfg(dir)
	store := New(storeCfg, nil).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < 30; i++ {
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
	for i := 0; i < b.N; i++ {
		datas.Height = int64(i)
		hash, err := store.MemSet(datas, true)
		assert.Nil(b, err)
		req := &types.ReqHash{
			Hash: hash,
		}
		_, err = store.Commit(req)
		assert.NoError(b, err, "NoError")
		datas.StateHash = hash
	}
	end := time.Now()
	fmt.Println("kvmvcc BenchmarkStoreCommit cost time is", end.Sub(start), "num is", b.N)
	b.StopTimer()
}

func BenchmarkIterMemSetkmvccMavl(b *testing.B) { benchmarkIterMemSet(b, false) }
func BenchmarkIterMemSetkmvcc(b *testing.B)     { benchmarkIterMemSet(b, true) }

//一次设定多对kv，测试一次的时间/多少对kv，来算平均一对kv的耗时。
func benchmarkIterMemSet(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	storeCfg, sub := newStoreCfgIter(dir)
	store := New(storeCfg, sub).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

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
	fmt.Println("kvmvcc BenchmarkMemSet cost time is", end.Sub(start), "num is", b.N)
}

func BenchmarkIterCommitkmvccMavl(b *testing.B) { benchmarkIterCommit(b, false) }
func BenchmarkIterCommitkmvcc(b *testing.B)     { benchmarkIterCommit(b, true) }

func benchmarkIterCommit(b *testing.B, isResetForkHeight bool) {
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(b, err)
	defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录
	storeCfg, sub := newStoreCfgIter(dir)
	store := New(storeCfg, sub).(*KVmMavlStore)
	assert.NotNil(b, store)

	if isResetForkHeight {
		kvmvccMavlFork = 0
		defer func() {
			kvmvccMavlFork = 200 * 10000
		}()
	}

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
	req := &types.ReqHash{
		Hash: hash,
	}
	_, err = store.Commit(req)
	assert.NoError(b, err, "NoError")

	end := time.Now()
	fmt.Println("kvmvcc BenchmarkCommit cost time is", end.Sub(start), "num is", b.N)
	b.StopTimer()
}
