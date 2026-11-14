// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/hashlock/types"
)

var clog = log.New("module", "execs.hashlock")

const minLockTime = 60

var driverName = "hashlock"

// Init hashlock
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, GetName(), newHashlock, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

//InitExecType ...
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Hashlock{}))
}

// GetName for hashlock
func GetName() string {
	return newHashlock().GetName()
}

// Hashlock driver
type Hashlock struct {
	drivers.DriverBase
}

func newHashlock() drivers.Driver {
	h := &Hashlock{}
	h.SetChild(h)
	h.SetExecutorType(types.LoadExecutorType(driverName))
	return h
}

// GetDriverName driverName
func (h *Hashlock) GetDriverName() string {
	return driverName
}

// CheckTx rejects a lock with an empty hash so the mempool does not pack a
// transaction that cannot be saved. Payloads that fail to decode are left to
// Exec, which returns the same decode error.
func (h *Hashlock) CheckTx(tx *types.Transaction, index int) error {
	var action pty.HashlockAction
	if err := types.Decode(tx.GetPayload(), &action); err != nil {
		return nil
	}
	if action.GetTy() == pty.HashlockActionLock && len(action.GetHlock().GetHash()) == 0 {
		return pty.ErrHashlockHash
	}
	return nil
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (h *Hashlock) CheckReceiptExecOk() bool {
	return true
}
