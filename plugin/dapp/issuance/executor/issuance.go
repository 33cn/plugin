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

func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Issuance{}))
}

type subConfig struct {
	ParaRemoteGrpcClient string `json:"paraRemoteGrpcClient"`
}

var cfg subConfig

// Init issuance
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	driverName := GetName()
	if name != driverName {
		panic("system dapp can't be rename")
	}
	if sub != nil {
		types.MustDecode(sub, &cfg)
	}
	drivers.Register(cfg, driverName, newIssuance, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
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

func (c *Issuance) addIssuanceID(index int64, issuanceId string) (kvs []*types.KeyValue) {
	key := calcIssuanceKey(issuanceId, index)
	record := &pty.IssuanceRecord{
		IssuanceId: issuanceId,
		Index:      index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) deleteIssuanceID(index int64, issuanceId string) (kvs []*types.KeyValue) {
	key := calcIssuanceKey(issuanceId, index)
	kv := &types.KeyValue{Key: key, Value: nil}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) addIssuanceStatus(status int32, index int64, issuanceId string) (kvs []*types.KeyValue) {
	key := calcIssuanceStatusKey(status, index)
	record := &pty.IssuanceRecord{
		IssuanceId: issuanceId,
		Index:      index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) deleteIssuanceStatus(status int32, index int64) (kvs []*types.KeyValue) {
	key := calcIssuanceStatusKey(status, index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) addIssuanceRecordAddr(accountAddr string, index int64, debtId string, issuanceId string) (kvs []*types.KeyValue) {
	key := calcIssuanceRecordAddrKey(accountAddr, index)
	record := &pty.IssuanceRecord{
		IssuanceId: issuanceId,
		DebtId:     debtId,
		Index:      index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) deleteIssuanceRecordAddr(accountAddr string, index int64) (kvs []*types.KeyValue) {
	key := calcIssuanceRecordAddrKey(accountAddr, index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) addIssuanceRecordStatus(recordStatus int32, accountAddr string, index int64, debtId string, issuanceId string) (kvs []*types.KeyValue) {
	key := calcIssuanceRecordStatusKey(recordStatus, index)

	record := &pty.IssuanceRecord{
		IssuanceId: issuanceId,
		DebtId:     debtId,
		Addr:       accountAddr,
		Index:      index,
	}

	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) deleteIssuanceRecordStatus(status int32, index int64) (kvs []*types.KeyValue) {
	key := calcIssuanceRecordStatusKey(status, index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) addIssuancePriceRecord(recordTime int64, price float64) (kvs []*types.KeyValue) {
	key := calcIssuancePriceKey(string(recordTime))

	record := &pty.IssuanceAssetPriceRecord{
		RecordTime: recordTime,
		BtyPrice:   price,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Issuance) deleteIssuancePriceRecord(recordTime int64) (kvs []*types.KeyValue) {
	key := calcIssuancePriceKey(string(recordTime))

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
	cfg := c.GetAPI().GetConfig()
	if cfg.IsFork(c.GetHeight(), "ForkLocalDBAccess") {
		return drivers.ExecLocalSameTime
	}
	return c.DriverBase.ExecutorOrder()
}
