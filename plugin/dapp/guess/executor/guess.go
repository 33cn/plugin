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

func calcGuessGameAddrPrefix(addr string) []byte {
	key := fmt.Sprintf("LODB-guess-addr:%s:", addr)
	return []byte(key)
}

func calcGuessGameAddrKey(addr string, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-addr:%s:%018d", addr, index)
	return []byte(key)
}

func calcGuessGameStatusPrefix(status int32) []byte {
	key := fmt.Sprintf("LODB-guess-status-index:%d:", status)
	return []byte(key)
}

func calcGuessGameStatusKey(status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-status-index:%d:%018d", status, index)
	return []byte(key)
}

func calcGuessGameStatusAndPlayerKey(category string, status, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-category-status:%s:%d:%018d", category, status, index)
	return []byte(key)
}

func calcGuessGameStatusAndPlayerPrefix(category string, status int32) []byte {
	var key string
	key = fmt.Sprintf("LODB-guess-category-status:%s:%d", category, status)

	return []byte(key)
}

func addGuessGameStatusIndexKey(status int32, gameID string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameStatusKey(status, index)
	record := &pkt.PBGameIndexRecord{
		GameId: gameID,
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

func addGuessGameAddrIndexKey(status int32, addr, gameID string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameAddrKey(addr, index)
	record := &pkt.PBGameRecord{
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

func addPBGameStatusAndPlayer(status int32, player int32, value, index int64, gameId string) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameStatusAndPlayerKey(status, player, value, index)
	record := &pkt.PBGameIndexRecord{
		GameId: gameId,
		Index:  index,
	}
	kv.Value = types.Encode(record)
	return kv
}

func delPBGameStatusAndPlayer(status int32, player int32, value, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcGuessGameStatusAndPlayerKey(status, player, value, index)
	kv.Value = nil
	return kv
}
