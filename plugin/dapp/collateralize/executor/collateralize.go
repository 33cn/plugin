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

func (Coll *Collateralize) addCollateralizeStatus(collateralizelog *pty.ReceiptCollateralize) (kvs []*types.KeyValue) {
	key := calcCollateralizeStatusKey(collateralizelog.Status, collateralizelog.Index)
	record := &pty.CollateralizeRecord{
		CollateralizeId:collateralizelog.CollateralizeId,
		Index: collateralizelog.Index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (Coll *Collateralize) deleteCollateralizeStatus(collateralizelog *pty.ReceiptCollateralize) (kvs []*types.KeyValue) {
	key := calcCollateralizeStatusKey(collateralizelog.Status, collateralizelog.Index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (Coll *Collateralize) addCollateralizeAddr(collateralizelog *pty.ReceiptCollateralize) (kvs []*types.KeyValue) {
	key := calcCollateralizeAddrKey(collateralizelog.AccountAddr, collateralizelog.Index)
	record := &pty.CollateralizeRecord{
		CollateralizeId:collateralizelog.CollateralizeId,
		Index: collateralizelog.Index,
	}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (Coll *Collateralize) deleteCollateralizeAddr(collateralizelog *pty.ReceiptCollateralize) (kvs []*types.KeyValue) {
	key := calcCollateralizeAddrKey(collateralizelog.AccountAddr, collateralizelog.Index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}
