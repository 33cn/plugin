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
func (Coll *Collateralize) GetDriverName() string {
	return pty.CollateralizeX
}

func (Coll *Collateralize) saveCollateralizeBorrow(collateralizelog *pty.ReceiptCollateralize) (kvs []*types.KeyValue) {
	key := calcCollateralizeBorrowKey(collateralizelog.CollateralizeId, collateralizelog.AccountAddr)
	record := &pty.CollateralizeBorrowRecord{CollateralizeId:collateralizelog.CollateralizeId,}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (Coll *Collateralize) deleteCollateralizeBorrow(collateralizelog *pty.ReceiptCollateralize) (kvs []*types.KeyValue) {
	key := calcCollateralizeBorrowKey(collateralizelog.CollateralizeId, collateralizelog.AccountAddr)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (Coll *Collateralize) saveCollateralizeRepay(collateralizelog *pty.ReceiptCollateralize) (kvs []*types.KeyValue) {
	key := calcCollateralizeRepayKey(collateralizelog.CollateralizeId)
	record := &pty.CollateralizeRepayRecord{}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}
	kvs = append(kvs, kv)
	return kvs
}

func (Coll *Collateralize) deleteCollateralizeRepay(collateralizelog *pty.ReceiptCollateralize) (kvs []*types.KeyValue) {
	key := calcCollateralizeRepayKey(collateralizelog.CollateralizeId)
	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (Coll *Collateralize) saveCollateralize(collateralizelog *pty.ReceiptCollateralize) (kvs []*types.KeyValue) {
	if collateralizelog.PreStatus > 0 {
		kv := delCollateralize(collateralizelog.CollateralizeId, collateralizelog.PreStatus)
		kvs = append(kvs, kv)
	}
	kvs = append(kvs, addCollateralize(collateralizelog.CollateralizeId, collateralizelog.Status))
	return kvs
}

func (Coll *Collateralize) deleteCollateralize(collateralizelog *pty.ReceiptCollateralize) (kvs []*types.KeyValue) {
	if collateralizelog.PreStatus > 0 {
		kv := addCollateralize(collateralizelog.CollateralizeId, collateralizelog.PreStatus)
		kvs = append(kvs, kv)
	}
	kvs = append(kvs, delCollateralize(collateralizelog.CollateralizeId, collateralizelog.Status))
	return kvs
}

func addCollateralize(collateralizeID string, status int32) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcCollateralizeKey(collateralizeID)
	kv.Value = []byte(collateralizeID)
	return kv
}

func delCollateralize(collateralizeID string, status int32) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcCollateralizeKey(collateralizeID)
	kv.Value = nil
	return kv
}

// GetPayloadValue CollateralizeAction
func (Coll *Collateralize) GetPayloadValue() types.Message {
	return &pty.CollateralizeAction{}
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (Coll *Collateralize) CheckReceiptExecOk() bool {
	return true
}
