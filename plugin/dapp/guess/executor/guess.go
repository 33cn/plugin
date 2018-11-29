// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/guess/types"
)

var logger = log.New("module", "execs.guess")

func Init(name string, sub []byte) {
	drivers.Register(newGuessGame().GetName(), newGuessGame, types.GetDappFork(driverName, "Enable"))
}

var driverName = pkt.GuessX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Guess{}))
}

type Guess struct {
	drivers.DriverBase
}

func newGuessGame() drivers.Driver {
	t := &Guess{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

func GetName() string {
	return newGuessGame().GetName()
}

func (g *Guess) GetDriverName() string {
	return pkt.GuessX
}

//addr prefix
func calcGuessGameAddrPrefix(addr string) []byte {
	key := fmt.Sprintf("LODB-guess-addr:%s:", addr)
	return []byte(key)
}

//addr index
func calcGuessGameAddrKey(addr string, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-addr:%s:%018d", addr, index)
	return []byte(key)
}

//status prefix
func calcGuessGameStatusPrefix(status int32) []byte {
	key := fmt.Sprintf("LODB-guess-status-index:%d:", status)
	return []byte(key)
}

//status index
func calcGuessGameStatusKey(status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-status-index:%d:%018d", status, index)
	return []byte(key)
}

//addr status prefix
func calcGuessGameAddrStatusPrefix(addr string, status int32) []byte {
	key := fmt.Sprintf("LODB-guess-addr-status-index:%s:%d:", addr, status)
	return []byte(key)
}

//addr status index
func calcGuessGameAddrStatusKey(addr string, status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-addr-status-index:%s:%d:%018d", addr, status, index)
	return []byte(key)
}

//admin prefix
func calcGuessGameAdminPrefix(addr string) []byte {
	key := fmt.Sprintf("LODB-guess-admin:%s:", addr)
	return []byte(key)
}

//admin index
func calcGuessGameAdminKey(addr string, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-admin:%s:%018d", addr, index)
	return []byte(key)
}

//admin status prefix
func calcGuessGameAdminStatusPrefix(admin string, status int32) []byte {
	key := fmt.Sprintf("LODB-guess-admin-status-index:%s:%d:", admin, status)
	return []byte(key)
}

//admin status index
func calcGuessGameAdminStatusKey(admin string, status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-admin-status-index:%s:%d:%018d", admin, status, index)
	return []byte(key)
}

func calcGuessGameCategoryStatusPrefix(category string, status int32) []byte {
	key := fmt.Sprintf("LODB-guess-category-status-index:%s:%d:", category, status)
	return []byte(key)
}

func calcGuessGameCategoryStatusKey(category string, status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-category-status-index:%s:%d:%018d", category, status, index)
	return []byte(key)
}

func addGuessGameAddrIndexKey(status int32, addr, gameID string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameAddrKey(addr, index)
	record := &pkt.GuessGameRecord{
		GameId: gameID,
		Status: status,
		Index:  index,
	}
	kv.Value = types.Encode(record)
	return kv
}

func delGuessGameAddrIndexKey(addr string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameAddrKey(addr, index)
	kv.Value = nil
	return kv
}

func addGuessGameStatusIndexKey(status int32, gameID string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameStatusKey(status, index)
	record := &pkt.GuessGameRecord{
		GameId: gameID,
		Status: status,
		Index:  index,
	}
	kv.Value = types.Encode(record)
	return kv
}

func delGuessGameStatusIndexKey(status int32, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameStatusKey(status, index)
	kv.Value = nil
	return kv
}

func addGuessGameAddrStatusIndexKey(status int32, addr, gameID string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameAddrStatusKey(addr, status, index)
	record := &pkt.GuessGameRecord{
		GameId: gameID,
		Status: status,
		Index:  index,
	}
	kv.Value = types.Encode(record)
	return kv
}

func delGuessGameAddrStatusIndexKey(status int32, addr string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameAddrStatusKey(addr, status, index)
	kv.Value = nil
	return kv
}

func addGuessGameAdminIndexKey(status int32, addr, gameID string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameAdminKey(addr, index)
	record := &pkt.GuessGameRecord{
		GameId: gameID,
		Status: status,
		Index:  index,
	}
	kv.Value = types.Encode(record)
	return kv
}

func delGuessGameAdminIndexKey(addr string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameAdminKey(addr, index)
	kv.Value = nil
	return kv
}

func addGuessGameAdminStatusIndexKey(status int32, addr, gameID string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameAdminStatusKey(addr, status, index)
	record := &pkt.GuessGameRecord{
		GameId: gameID,
		Status: status,
		Index:  index,
	}
	kv.Value = types.Encode(record)
	return kv
}

func delGuessGameAdminStatusIndexKey(status int32, addr string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameAdminStatusKey(addr, status, index)
	kv.Value = nil
	return kv
}

func addGuessGameCategoryStatusIndexKey(status int32, category, gameID string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameCategoryStatusKey(category, status, index)
	record := &pkt.GuessGameRecord{
		GameId: gameID,
		Status: status,
		Index:  index,
	}
	kv.Value = types.Encode(record)
	return kv
}

func delGuessGameCategoryStatusIndexKey(status int32, category string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameCategoryStatusKey(category, status, index)
	kv.Value = nil
	return kv
}