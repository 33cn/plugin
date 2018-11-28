// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/pokerbull/types"
)

var logger = log.New("module", "execs.pokerbull")

// Init 执行器初始化
func Init(name string, sub []byte) {
	drivers.Register(newPBGame().GetName(), newPBGame, types.GetDappFork(driverName, "Enable"))
}

var driverName = pkt.PokerBullX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&PokerBull{}))
}

// PokerBull 斗牛执行器结构
type PokerBull struct {
	drivers.DriverBase
}

func newPBGame() drivers.Driver {
	t := &PokerBull{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName 获取斗牛执行器名
func GetName() string {
	return newPBGame().GetName()
}

// GetDriverName 获取斗牛执行器名
func (g *PokerBull) GetDriverName() string {
	return pkt.PokerBullX
}

func calcPBGameAddrPrefix(addr string) []byte {
	key := fmt.Sprintf("LODB-pokerbull-addr:%s:", addr)
	return []byte(key)
}

func calcPBGameAddrKey(addr string, index int64) []byte {
	key := fmt.Sprintf("LODB-pokerbull-addr:%s:%018d", addr, index)
	return []byte(key)
}

func calcPBGameStatusPrefix(status int32) []byte {
	key := fmt.Sprintf("LODB-pokerbull-status-index:%d:", status)
	return []byte(key)
}

func calcPBGameStatusKey(status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-pokerbull-status-index:%d:%018d", status, index)
	return []byte(key)
}

func calcPBGameStatusAndPlayerKey(status, player int32, value, index int64) []byte {
	key := fmt.Sprintf("LODB-pokerbull-status:%d:%d:%015d:%018d", status, player, value, index)
	return []byte(key)
}

func calcPBGameStatusAndPlayerPrefix(status, player int32, value int64) []byte {
	var key string
	if value == 0 {
		key = fmt.Sprintf("LODB-pokerbull-status:%d:%d:", status, player)
	} else {
		key = fmt.Sprintf("LODB-pokerbull-status:%d:%d:%015d", status, player, value)
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

func addPBGameStatusAndPlayer(status int32, player int32, value, index int64, gameID string) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcPBGameStatusAndPlayerKey(status, player, value, index)
	record := &pkt.PBGameIndexRecord{
		GameId: gameID,
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

// CheckReceiptExecOk return true to check if receipt ty is ok
func (g *PokerBull) CheckReceiptExecOk() bool {
	return true
}
