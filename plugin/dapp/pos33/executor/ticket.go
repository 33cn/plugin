// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

/*
coins 是一个货币的exec。内置货币的执行器。

主要提供两种操作：

EventTransfer -> 转移资产
*/

//package none execer for unknow execer
//all none transaction exec ok, execept nofee
//nofee transaction will not pack into block

import (
	"fmt"
	"strconv"

	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

var clog = log.New("module", "execs.pos33")
var driverName = "pos33"

// Init initial
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, GetName(), newPos33Ticket, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

// InitExecType reg types
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Pos33Ticket{}))
}

// GetName get name
func GetName() string {
	return newPos33Ticket().GetName()
}

// Pos33Ticket driver type
type Pos33Ticket struct {
	drivers.DriverBase
}

func newPos33Ticket() drivers.Driver {
	t := &Pos33Ticket{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetDriverName ...
func (t *Pos33Ticket) GetDriverName() string {
	return driverName
}

func (t *Pos33Ticket) savePos33TicketBind(b *ty.ReceiptPos33TicketBind) (kvs []*types.KeyValue) {
	//解除原来的绑定
	if len(b.OldMinerAddress) > 0 {
		kv := &types.KeyValue{
			Key:   calcBindMinerKey(b.OldMinerAddress, b.ReturnAddress),
			Value: nil,
		}
		//tlog.Warn("tb:del", "key", string(kv.Key))
		kvs = append(kvs, kv)
	}

	kv := &types.KeyValue{Key: calcBindReturnKey(b.ReturnAddress), Value: []byte(b.NewMinerAddress)}
	//tlog.Warn("tb:add", "key", string(kv.Key), "value", string(kv.Value))
	kvs = append(kvs, kv)
	kv = &types.KeyValue{
		Key:   calcBindMinerKey(b.GetNewMinerAddress(), b.ReturnAddress),
		Value: []byte(b.ReturnAddress),
	}
	//tlog.Warn("tb:add", "key", string(kv.Key), "value", string(kv.Value))
	kvs = append(kvs, kv)
	return kvs
}

func (t *Pos33Ticket) delPos33TicketBind(b *ty.ReceiptPos33TicketBind) (kvs []*types.KeyValue) {
	//被取消了，刚好操作反
	kv := &types.KeyValue{
		Key:   calcBindMinerKey(b.NewMinerAddress, b.ReturnAddress),
		Value: nil,
	}
	kvs = append(kvs, kv)
	if len(b.OldMinerAddress) > 0 {
		//恢复旧的绑定
		kv := &types.KeyValue{Key: calcBindReturnKey(b.ReturnAddress), Value: []byte(b.OldMinerAddress)}
		kvs = append(kvs, kv)
		kv = &types.KeyValue{
			Key:   calcBindMinerKey(b.OldMinerAddress, b.ReturnAddress),
			Value: []byte(b.ReturnAddress),
		}
		kvs = append(kvs, kv)
	} else {
		//删除旧的数据
		kv := &types.KeyValue{Key: calcBindReturnKey(b.ReturnAddress), Value: nil}
		kvs = append(kvs, kv)
	}
	return kvs
}

func (t *Pos33Ticket) getAllPos33TicketCount(height int64) (int, error) {
	preH := height - height%ty.Pos33SortitionSize
	if preH == height {
		preH -= ty.Pos33SortitionSize
	}
	key := []byte(ty.Pos33AllPos33TicketCountKeyPrefix + fmt.Sprintf("%d", preH))
	count := 0
	value, err := t.GetLocalDB().Get(key)
	if err != nil {
		clog.Info("savePos33TicketCount error", "error", err)
		return 0, err
	}

	count, err = strconv.Atoi(string(value))
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (t *Pos33Ticket) saveAllPos33TicketCount(openOrClose bool) (kvs []*types.KeyValue) {
	height := t.GetHeight()
	preH := height - height%ty.Pos33SortitionSize
	if preH == height {
		preH -= ty.Pos33SortitionSize
	}
	key := []byte(ty.Pos33AllPos33TicketCountKeyPrefix + fmt.Sprintf("%d", preH))
	count := 0
	value, err := t.GetLocalDB().Get(key)
	if err != nil {
		clog.Info("savePos33TicketCount error", "error", err)
	} else {
		count, err = strconv.Atoi(string(value))
		if err != nil {
			panic(err)
		}
	}
	nxtH := preH + ty.Pos33SortitionSize
	key = []byte(ty.Pos33AllPos33TicketCountKeyPrefix + fmt.Sprintf("%d", nxtH))
	if openOrClose {
		count++
	} else {
		count--
	}
	return []*types.KeyValue{&types.KeyValue{Key: key, Value: []byte(fmt.Sprintf("%d", count))}}
}

func (t *Pos33Ticket) savePos33Ticket(ticketlog *ty.ReceiptPos33Ticket) (kvs []*types.KeyValue) {
	if ticketlog.PrevStatus > 0 {
		kv := delticket(ticketlog.Addr, ticketlog.TicketId, ticketlog.PrevStatus)
		kvs = append(kvs, kv)
	}
	kvs = append(kvs, addticket(ticketlog.Addr, ticketlog.TicketId, ticketlog.Status))
	return kvs
}

func (t *Pos33Ticket) delPos33Ticket(ticketlog *ty.ReceiptPos33Ticket) (kvs []*types.KeyValue) {
	if ticketlog.PrevStatus > 0 {
		kv := addticket(ticketlog.Addr, ticketlog.TicketId, ticketlog.PrevStatus)
		kvs = append(kvs, kv)
	}
	kvs = append(kvs, delticket(ticketlog.Addr, ticketlog.TicketId, ticketlog.Status))
	return kvs
}

func calcPos33TicketKey(addr string, ticketID string, status int32) []byte {
	key := fmt.Sprintf("LODB-pos33-tl:%s:%d:%s", addr, status, ticketID)
	return []byte(key)
}

func calcBindReturnKey(returnAddress string) []byte {
	key := fmt.Sprintf("LODB-pos33-bind:%s", returnAddress)
	return []byte(key)
}

func calcBindMinerKey(minerAddress string, returnAddress string) []byte {
	key := fmt.Sprintf("LODB-pos33-miner:%s:%s", minerAddress, returnAddress)
	return []byte(key)
}

func calcBindMinerKeyPrefix(minerAddress string) []byte {
	key := fmt.Sprintf("LODB-pos33-miner:%s", minerAddress)
	return []byte(key)
}

func calcPos33TicketPrefix(addr string, status int32) []byte {
	key := fmt.Sprintf("LODB-pos33-tl:%s:%d", addr, status)
	return []byte(key)
}

func addticket(addr string, ticketID string, status int32) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcPos33TicketKey(addr, ticketID, status)
	kv.Value = []byte(ticketID)
	return kv
}

func delticket(addr string, ticketID string, status int32) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcPos33TicketKey(addr, ticketID, status)
	kv.Value = nil
	return kv
}

// IsFriend check is fri
func (t *Pos33Ticket) IsFriend(myexec, writekey []byte, tx *types.Transaction) bool {
	clog.Error("ticket  IsFriend", "myex", string(myexec), "writekey", string(writekey))
	//不允许平行链
	return false
}

// CheckTx check tx
func (t *Pos33Ticket) CheckTx(tx *types.Transaction, index int) error {
	//index == -1 only when check in mempool
	if index == -1 {
		var action ty.Pos33TicketAction
		err := types.Decode(tx.Payload, &action)
		if err != nil {
			return err
		}
		if action.Ty == ty.Pos33TicketActionMiner && action.GetMiner() != nil {
			return ty.ErrMinerTx
		}
	}
	return nil
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (t *Pos33Ticket) CheckReceiptExecOk() bool {
	return true
}
