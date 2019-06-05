// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"time"

	"fmt"

	"github.com/33cn/chain33/common"
	drivers "github.com/33cn/chain33/system/store"
	"github.com/33cn/chain33/types"
	mavl "github.com/33cn/chain33/system/store/mavl"
)

const MaxKeylenth int = 64

func newStoreCfg(dir string) *types.Store {
	return &types.Store{Name: "mavl_test", Driver: "leveldb", DbPath: dir, DbCache: 100}
}

func GetRandomString(length int) string {
	return common.GetRandPrintString(20, length)
}
func main() {
	dir, _ := ioutil.TempDir("", "mavl")
	//defer os.RemoveAll(dir) // clean up
	os.RemoveAll(dir)       //删除已存在目录

	var storeCfg = newStoreCfg(dir)
	store := mavl.New(storeCfg, nil).(*mavl.Store)

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < 20; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("vabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrst%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{Key: []byte(string(key)), Value: []byte(string(value))})
	}
	datas := &types.StoreSet{
		StateHash: drivers.EmptyRoot[:],
		KV:        kv,
		Height:    0}

	var hashes [][]byte
	blocks := 50000
	times := 10000
	start1 := time.Now()
	for i := 0; i < blocks; i++ {
		datas.Height = int64(i)
		value = fmt.Sprintf("vvabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrst%d", i)
		for j := 0; j < 20; j++ {
			datas.KV[j].Value = []byte(value)
		}
		hash, err := store.MemSet(datas, true)
		if err != nil {
			fmt.Println("MemSet failed", err)
		}

		req := &types.ReqHash{
			Hash: hash,
		}

		_, err = store.Commit(req)
		if err != nil {
			fmt.Println("Commit failed", err)
		}

		datas.StateHash = hash
		if i < times {
			hashes = append(hashes, hash)
		}
		fmt.Println("Block number:", i)
	}
	end1 := time.Now()

	start := time.Now()

	getData := &types.StoreGet{
		StateHash: hashes[0],
		Keys:      keys}

	for i := 0; i < times; i++ {
		getData.StateHash = hashes[i]
		store.Get(getData)
		fmt.Println("read times:", i, " kv numbers:", i * 20)
	}
	end := time.Now()
	fmt.Println("mavl BenchmarkStoreGetKvsFor100million MemSet&Commit cost time is ", end1.Sub(start1), "blocks is", blocks)
	fmt.Println("mavl BenchmarkStoreGetKvsFor100million Get cost time is", end.Sub(start), "num is ", times, ",blocks is ", blocks)
}
