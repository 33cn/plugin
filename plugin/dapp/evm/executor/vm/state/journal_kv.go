// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package state

import (
	"sort"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
)

// journalKV 是 MemoryStateDB 与执行器 StateDB 之间的写缓冲，作用等同 chain33 的 txcache，
// 但支持与 Snapshot 版本一一对应的嵌套回滚。
//
// EVM 执行期间所有经由 MemoryStateDB 的状态写入（coins/token 账户等）只进入本缓冲：
// 回滚某个快照时，按 undo 日志倒序恢复写之前的值，调用方无需为每种写入单独实现 revert。
// 执行成功后由 GetChangedData 把留下的写入一次性交给回执，并写回底层 StateDB。
type journalKV struct {
	parent db.KV
	dirty  map[string][]byte // nil 值表示该 key 在本交易内不存在
	undo   []kvUndo
	marks  []int // marks[v] 为快照 v 创建时 undo 的长度
}

type kvUndo struct {
	key     string
	prev    []byte
	hadPrev bool // 写之前 dirty 中是否已有该 key
}

func newJournalKV(parent db.KV) *journalKV {
	return &journalKV{parent: parent, dirty: make(map[string][]byte)}
}

// Get 先读缓冲，再读底层 StateDB
func (j *journalKV) Get(key []byte) ([]byte, error) {
	if v, ok := j.dirty[string(key)]; ok {
		if v == nil {
			return nil, types.ErrNotFound
		}
		return v, nil
	}
	return j.parent.Get(key)
}

// Set 只写缓冲，并记录 undo
func (j *journalKV) Set(key []byte, value []byte) error {
	k := string(key)
	prev, had := j.dirty[k]
	j.undo = append(j.undo, kvUndo{key: k, prev: prev, hadPrev: had})
	if value != nil {
		value = append([]byte(nil), value...)
	}
	j.dirty[k] = value
	return nil
}

// Begin/Commit/Rollback 满足 db.KV 接口；交易级的提交与回滚由执行器框架负责
func (j *journalKV) Begin()        {}
func (j *journalKV) Commit() error { return nil }
func (j *journalKV) Rollback()     { j.revert(0) }

// snapshot 记录快照 version 的回滚点，version 必须等于当前快照数
func (j *journalKV) snapshot(version int) {
	j.marks = append(j.marks[:version], len(j.undo))
}

// revertTo 把缓冲恢复到快照 version 创建时的状态
func (j *journalKV) revertTo(version int) {
	if version < 0 || version >= len(j.marks) {
		return
	}
	j.revert(j.marks[version])
	j.marks = j.marks[:version]
}

func (j *journalKV) revert(n int) {
	for i := len(j.undo) - 1; i >= n; i-- {
		u := j.undo[i]
		if u.hadPrev {
			j.dirty[u.key] = u.prev
		} else {
			delete(j.dirty, u.key)
		}
	}
	j.undo = j.undo[:n]
}

// changes 返回缓冲中的全部写入，按 key 排序保证各节点一致
func (j *journalKV) changes() []*types.KeyValue {
	keys := make([]string, 0, len(j.dirty))
	for k := range j.dirty {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	kvs := make([]*types.KeyValue, 0, len(keys))
	for _, k := range keys {
		kvs = append(kvs, &types.KeyValue{Key: []byte(k), Value: j.dirty[k]})
	}
	return kvs
}
