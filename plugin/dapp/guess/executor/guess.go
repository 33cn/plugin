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

func calcPBGameAddrPrefix(addr string) []byte {
	key := fmt.Sprintf("LODB-guess-addr:%s:", addr)
	return []byte(key)
}

func calcPBGameAddrKey(addr string, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-addr:%s:%018d", addr, index)
	return []byte(key)
}

func calcPBGameStatusPrefix(status int32) []byte {
	key := fmt.Sprintf("LODB-guess-status-index:%d:", status)
	return []byte(key)
}

func calcPBGameStatusKey(status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-status-index:%d:%018d", status, index)
	return []byte(key)
}

func calcPBGameStatusAndPlayerKey(status, player int32, value, index int64) []byte {
	key := fmt.Sprintf("LODB-guess-status:%d:%d:%d:%018d", status, player, value, index)
	return []byte(key)
}

func calcPBGameStatusAndPlayerPrefix(status, player int32, value int64) []byte {
	var key string
	if value == 0 {
		key = fmt.Sprintf("LODB-guess-status:%d:%d:", status, player)
	} else {
		key = fmt.Sprintf("LODB-guess-status:%d:%d:%d", status, player, value)
	}

	return []byte(key)
}

func addPBGameStatusIndexKey(status int32, gameID string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcPBGameStatusKey(status, index)
	record := &pkt.PBGameIndexRecord{
		GameId: gameID,
		Index:  index,
	}
	kv.Value = types.Encode(record)
	return kv
}

func delPBGameStatusIndexKey(status int32, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcPBGameStatusKey(status, index)
	kv.Value = nil
	return kv
}

func addPBGameAddrIndexKey(status int32, addr, gameID string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcPBGameAddrKey(addr, index)
	record := &pkt.PBGameRecord{
		GameId: gameID,
		Status: status,
		Index:  index,
	}
	kv.Value = types.Encode(record)
	return kv
}

func delPBGameAddrIndexKey(addr string, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcPBGameAddrKey(addr, index)
	kv.Value = nil
	return kv
}

func addPBGameStatusAndPlayer(status int32, player int32, value, index int64, gameId string) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcPBGameStatusAndPlayerKey(status, player, value, index)
	record := &pkt.PBGameIndexRecord{
		GameId: gameId,
		Index:  index,
	}
	kv.Value = types.Encode(record)
	return kv
}

func delPBGameStatusAndPlayer(status int32, player int32, value, index int64) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcPBGameStatusAndPlayerKey(status, player, value, index)
	kv.Value = nil
	return kv
}
