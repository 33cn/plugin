// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/issuance/types"
)

var clog = log.New("module", "execs.issuance")
var driverName = pty.IssuanceX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Issuance{}))
}

type subConfig struct {
	ParaRemoteGrpcClient string `json:"paraRemoteGrpcClient"`
}

var cfg subConfig

// Init issuance
func Init(name string, sub []byte) {
	driverName := GetName()
	if name != driverName {
		panic("system dapp can't be rename")
	}
	if sub != nil {
		types.MustDecode(sub, &cfg)
	}
	drivers.Register(driverName, newIssuance, types.GetDappFork(driverName, "Enable"))
}

// GetName for Issuance
func GetName() string {
	return newIssuance().GetName()
}

// Issuance driver
type Issuance struct {
	drivers.DriverBase
}

func newIssuance() drivers.Driver {
	c := &Issuance{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(driverName))
	return c
}

// GetDriverName for Issuance
func (c *Issuance) GetDriverName() string {
	return pty.IssuanceX
}

func (c *Issuance) addIssuanceID(issuancelog *pty.ReceiptIssuance) (kvs []*types.KeyValue) {
	key := calcIssuanceKey(issuancelog.IssuanceId, issuancelog.Index)
	record := &pty.IssuanceRecord{
		IssuanceId:issuancelog.IssuanceId,
		Index: issuancelog.Index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) deleteIssuanceID(issuancelog *pty.ReceiptIssuance) (kvs []*types.KeyValue) {
	key := calcIssuanceKey(issuancelog.IssuanceId, issuancelog.Index)
	kv := &types.KeyValue{Key: key, Value: nil}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) addIssuanceStatus(issuancelog *pty.ReceiptIssuance) (kvs []*types.KeyValue) {
	key := calcIssuanceStatusKey(issuancelog.Status, issuancelog.Index)
	record := &pty.IssuanceRecord{
		IssuanceId:issuancelog.IssuanceId,
		Index: issuancelog.Index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) deleteIssuanceStatus(issuancelog *pty.ReceiptIssuance) (kvs []*types.KeyValue) {
	key := calcIssuanceStatusKey(issuancelog.Status, issuancelog.Index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) addIssuanceAddr(issuancelog *pty.ReceiptIssuance) (kvs []*types.KeyValue) {
	key := calcIssuanceAddrKey(issuancelog.AccountAddr, issuancelog.Index)
	record := &pty.IssuanceRecord{
		IssuanceId:issuancelog.IssuanceId,
		Index: issuancelog.Index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) deleteIssuanceAddr(issuancelog *pty.ReceiptIssuance) (kvs []*types.KeyValue) {
	key := calcIssuanceAddrKey(issuancelog.AccountAddr, issuancelog.Index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) addIssuanceRecordStatus(issuancelog *pty.ReceiptIssuance) (kvs []*types.KeyValue) {
	key := calcIssuanceRecordStatusKey(issuancelog.RecordStatus, issuancelog.Index)

	record := &pty.IssuanceRecord{
		IssuanceId:issuancelog.IssuanceId,
		Addr:  issuancelog.AccountAddr,
		Index: issuancelog.Index,
	}

	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) deleteIssuanceRecordStatus(issuancelog *pty.ReceiptIssuance) (kvs []*types.KeyValue) {
	key := calcIssuanceRecordStatusKey(issuancelog.RecordStatus, issuancelog.Index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (c *Issuance) CheckReceiptExecOk() bool {
	return true
}

// ExecutorOrder 设置localdb的EnableRead
func (c *Issuance) ExecutorOrder() int64 {
	if types.IsFork(c.GetHeight(), "ForkLocalDBAccess") {
		return drivers.ExecLocalSameTime
	}
	return c.DriverBase.ExecutorOrder()
}