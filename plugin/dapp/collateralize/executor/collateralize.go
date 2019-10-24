// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/collateralize/types"
)

var clog = log.New("module", "execs.collateralize")
var driverName = pty.CollateralizeX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Collateralize{}))
}

type subConfig struct {
	ParaRemoteGrpcClient string `json:"paraRemoteGrpcClient"`
}

var cfg subConfig

// Init collateralize
func Init(name string, sub []byte) {
	driverName := GetName()
	if name != driverName {
		panic("system dapp can't be rename")
	}
	if sub != nil {
		types.MustDecode(sub, &cfg)
	}
	drivers.Register(driverName, newCollateralize, types.GetDappFork(driverName, "Enable"))
}

// GetName for Collateralize
func GetName() string {
	return newCollateralize().GetName()
}

// Collateralize driver
type Collateralize struct {
	drivers.DriverBase
}

func newCollateralize() drivers.Driver {
	c := &Collateralize{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(driverName))
	return c
}

// GetDriverName for Collateralize
func (c *Collateralize) GetDriverName() string {
	return pty.CollateralizeX
}

func (c *Collateralize) addCollateralizeID(collateralizeId string, index int64) (kvs []*types.KeyValue) {
	key := calcCollateralizeKey(collateralizeId, index)
	record := &pty.CollateralizeRecord{
		CollateralizeId:collateralizeId,
		Index: index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Collateralize) deleteCollateralizeID(collateralizeId string, index int64) (kvs []*types.KeyValue) {
	key := calcCollateralizeKey(collateralizeId, index)
	kv := &types.KeyValue{Key: key, Value: nil}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Collateralize) addCollateralizeStatus(status int32, collateralizeId string, index int64) (kvs []*types.KeyValue) {
	key := calcCollateralizeStatusKey(status, index)
	record := &pty.CollateralizeRecord{
		CollateralizeId:collateralizeId,
		Index: index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Collateralize) deleteCollateralizeStatus(status int32, index int64) (kvs []*types.KeyValue) {
	key := calcCollateralizeStatusKey(status, index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Collateralize) addCollateralizeAddr(addr string, collateralizeId string, index int64) (kvs []*types.KeyValue) {
	key := calcCollateralizeAddrKey(addr, index)
	record := &pty.CollateralizeRecord{
		CollateralizeId:collateralizeId,
		Index: index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (c *Collateralize) deleteCollateralizeAddr(addr string, index int64) (kvs []*types.KeyValue) {
	key := calcCollateralizeAddrKey(addr, index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Collateralize) addCollateralizeRecordStatus(recordStatus int32, collateralizeId string, recordId string, index int64) (kvs []*types.KeyValue) {
	key := calcCollateralizeRecordStatusKey(recordStatus, index)

	record := &pty.CollateralizeRecord{
		CollateralizeId:collateralizeId,
		RecordId:recordId,
		Index: index,
	}

	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Collateralize) deleteCollateralizeRecordStatus(recordStatus int32, index int64) (kvs []*types.KeyValue) {
	key := calcCollateralizeRecordStatusKey(recordStatus, index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Collateralize) addCollateralizeRecordAddr(recordAddr string, collateralizeId string, recordId string, index int64) (kvs []*types.KeyValue) {
	key := calcCollateralizeRecordAddrKey(recordAddr, index)

	record := &pty.CollateralizeRecord{
		CollateralizeId:collateralizeId,
		RecordId:recordId,
		Index: index,
	}

	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}
	kvs = append(kvs, kv)
	return kvs
}

func (c *Collateralize) deleteCollateralizeRecordAddr(recordAddr string, index int64) (kvs []*types.KeyValue) {
	key := calcCollateralizeRecordAddrKey(recordAddr, index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (c *Collateralize) CheckReceiptExecOk() bool {
	return true
}

// ExecutorOrder 设置localdb的EnableRead
func (c *Collateralize) ExecutorOrder() int64 {
	if types.IsFork(c.GetHeight(), "ForkLocalDBAccess") {
		return drivers.ExecLocalSameTime
	}
	return c.DriverBase.ExecutorOrder()
}