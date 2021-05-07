// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
)

var clog = log.New("module", "execs.qbftNode")
var driverName = "qbftNode"

// Init method
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	clog.Debug("register qbftNode execer")
	drivers.Register(cfg, GetName(), newQbftNode, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

//InitExecType ...
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&QbftNode{}))
}

// GetName method
func GetName() string {
	return newQbftNode().GetName()
}

// QbftNode strucyt
type QbftNode struct {
	drivers.DriverBase
}

func newQbftNode() drivers.Driver {
	n := &QbftNode{}
	n.SetChild(n)
	n.SetIsFree(true)
	n.SetExecutorType(types.LoadExecutorType(driverName))
	return n
}

// GetDriverName method
func (qbft *QbftNode) GetDriverName() string {
	return driverName
}

// CheckTx method
func (qbft *QbftNode) CheckTx(tx *types.Transaction, index int) error {
	return nil
}

// CalcQbftNodeUpdateHeightIndexKey method
func CalcQbftNodeUpdateHeightIndexKey(height int64, index int) []byte {
	return []byte(fmt.Sprintf("LODB-qbftNode-Update:%18d:%18d", height, int64(index)))
}

// CalcQbftNodeUpdateHeightKey method
func CalcQbftNodeUpdateHeightKey(height int64) []byte {
	return []byte(fmt.Sprintf("LODB-qbftNode-Update:%18d:", height))
}

// CalcQbftNodeBlockInfoHeightKey method
func CalcQbftNodeBlockInfoHeightKey(height int64) []byte {
	return []byte(fmt.Sprintf("LODB-qbftNode-BlockInfo:%18d:", height))
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (qbft *QbftNode) CheckReceiptExecOk() bool {
	return true
}
