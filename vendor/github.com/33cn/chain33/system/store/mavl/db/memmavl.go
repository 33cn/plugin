// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mavl

import (
	"github.com/hashicorp/golang-lru"
	"fmt"
	"time"
	. "github.com/dgryski/go-farm"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
)

type MemTreeOpera interface {
	Add(key, value interface{})
	Get(key interface{}) (value interface{}, ok bool)
	Delete(key interface{})
	Contains(key interface{}) bool
	Len() int
}

type TreeMap struct {
	mpCache map[interface{}]interface{}
}

func NewTreeMap(size int) *TreeMap {
	mp := &TreeMap{}
	mp.mpCache = make(map[interface{}]interface{}, size)
	return mp
}

func (tm *TreeMap) Add(key, value interface{}) {
	if _, ok := tm.mpCache[key]; ok {
		panic(fmt.Sprintln("*********for test map*******", key))
		delete(tm.mpCache, key)
		return
	}
	tm.mpCache[key] = value
}

func (tm *TreeMap) Get(key interface{}) (value interface{}, ok bool) {
	if value, ok := tm.mpCache[key]; ok {
		return value, ok
	}
	return nil, false
}

func (tm *TreeMap) Delete(key interface{}) {
	if _, ok := tm.mpCache[key]; ok {
		delete(tm.mpCache, key)
	}
}

func (tm *TreeMap) Contains(key interface{}) bool {
	if _, ok := tm.mpCache[key]; ok {
		return true
	}
	return false
}

func (tm *TreeMap) Len() int {
	return len(tm.mpCache)
}

type TreeARC struct {
	arcCache *lru.ARCCache
}

func NewTreeARC(size int) *TreeARC {
	ma := &TreeARC{}
	ma.arcCache, _ = lru.NewARC(size)
	return ma
}

func (ta *TreeARC) Add(key, value interface{}) {
	if ta.arcCache.Contains(key) {
		panic(fmt.Sprintln("*********for test TreeARC*******", key))
		ta.arcCache.Remove(key)
		return
	}
	ta.arcCache.Add(key, value)
}

func (ta *TreeARC) Get(key interface{}) (value interface{}, ok bool) {
	return ta.arcCache.Get(key)
}

func (ta *TreeARC) Delete(key interface{}) {
	ta.arcCache.Remove(key)
}

func (ta *TreeARC) Contains(key interface{}) bool {
	return ta.arcCache.Contains(key)
}

func (ta *TreeARC) Len() int {
	return ta.arcCache.Len()
}

type hashNode struct {
	leftHash  []byte
	rightHash []byte
}

func LoadTree2MemDb(db dbm.DB, hash []byte, trMem MemTreeOpera) {
	if trMem == nil {
		return
	}
	nDb := newNodeDB(db, true)
	node, err := nDb.GetLightNode(nil, hash)
	if err != nil {
		fmt.Println("err", err)
		return
	}
	pri := ""
	if len(node.hash) > 32 {
		pri = string(node.hash[:16])
	}
	treelog.Info("hash node", "hash pri", pri, "hash", common.ToHex(node.hash), "height", node.height)
	start := time.Now()
	leftHash := make([]byte, len(node.leftHash))
	copy(leftHash, node.leftHash)
	rightHash := make([]byte, len(node.rightHash))
	copy(rightHash, node.rightHash)
	trMem.Add(Hash64(node.hash), &hashNode{leftHash: leftHash, rightHash: rightHash})
	node.LoadNodeInfo(nDb, trMem)
	end := time.Now()
	treelog.Info("hash node", "cost time", end.Sub(start), "node count", trMem.Len())
	PrintMemStats(1)
}

func (node *Node) LoadNodeInfo(db *nodeDB, trMem MemTreeOpera) {
	if node.height == 0 {
		//trMem.Add(Hash64(node.hash), &hashNode{leftHash: node.leftHash, rightHash: node.rightHash})
		leftHash := make([]byte, len(node.leftHash))
		copy(leftHash, node.leftHash)
		rightHash := make([]byte, len(node.rightHash))
		copy(rightHash, node.rightHash)
		trMem.Add(Hash64(node.hash), &hashNode{leftHash: leftHash, rightHash: rightHash})
		return
	}
	if node.leftHash != nil {
		left, err := db.GetLightNode(nil, node.leftHash)
		if err != nil {
			return
		}
		//trMem.Add(Hash64(left.hash), &hashNode{leftHash: left.leftHash, rightHash: left.rightHash})
		leftHash := make([]byte, len(left.leftHash))
		copy(leftHash, left.leftHash)
		rightHash := make([]byte, len(left.rightHash))
		copy(rightHash, left.rightHash)
		trMem.Add(Hash64(left.hash), &hashNode{leftHash: leftHash, rightHash: rightHash})
		left.LoadNodeInfo(db, trMem)
	}
	if node.rightHash != nil {
		right, err := db.GetLightNode(nil, node.rightHash)
		if err != nil {
			return
		}
		//trMem.Add(Hash64(right.hash), &hashNode{leftHash: right.leftHash, rightHash: right.rightHash})
		leftHash := make([]byte, len(right.leftHash))
		copy(leftHash, right.leftHash)
		rightHash := make([]byte, len(right.rightHash))
		copy(rightHash, right.rightHash)
		trMem.Add(Hash64(right.hash), &hashNode{leftHash: leftHash, rightHash: rightHash})
		right.LoadNodeInfo(db, trMem)
	}
}

func (ndb *nodeDB) GetLightNode(t *Tree, hash []byte) (*Node, error) {
	// Doesn't exist, load from db.
	var buf []byte
	buf, err := ndb.db.Get(hash)

	if len(buf) == 0 || err != nil {
		return nil, ErrNodeNotExist
	}
	node, err := MakeNode(buf, t)
	if err != nil {
		panic(fmt.Sprintf("Error reading IAVLNode. bytes: %X  error: %v", buf, err))
	}
	node.hash = hash
	node.key = nil
	node.value = nil
	return node, nil
}

func copyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)
	return copiedBytes
}