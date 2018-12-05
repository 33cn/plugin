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

var clog = log.New("module", "execs.valnode")
var driverName = "valnode"

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&ValNode{}))
}

// Init method
func Init(name string, sub []byte) {
	clog.Debug("register valnode execer")
	drivers.Register(GetName(), newValNode, 0)
}

// GetName method
func GetName() string {
	return newValNode().GetName()
}

// ValNode strucyt
type ValNode struct {
	drivers.DriverBase
}

func newValNode() drivers.Driver {
	n := &ValNode{}
	n.SetChild(n)
	n.SetIsFree(true)
	n.SetExecutorType(types.LoadExecutorType(driverName))
	return n
}

// GetDriverName method
func (val *ValNode) GetDriverName() string {
	return driverName
}

// CheckTx method
func (val *ValNode) CheckTx(tx *types.Transaction, index int) error {
	return nil
}

// CalcValNodeUpdateHeightIndexKey method
func CalcValNodeUpdateHeightIndexKey(height int64, index int) []byte {
	return []byte(fmt.Sprintf("LODB-valnode-Update:%18d:%18d", height, int64(index)))
}

// CalcValNodeUpdateHeightKey method
func CalcValNodeUpdateHeightKey(height int64) []byte {
	return []byte(fmt.Sprintf("LODB-valnode-Update:%18d:", height))
}

// CalcValNodeBlockInfoHeightKey method
func CalcValNodeBlockInfoHeightKey(height int64) []byte {
	return []byte(fmt.Sprintf("LODB-valnode-BlockInfo:%18d:", height))
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (val *ValNode) CheckReceiptExecOk() bool {
	return true
}
