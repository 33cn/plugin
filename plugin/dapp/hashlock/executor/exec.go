// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common/address"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/hashlock/types"
)

// Exec_Hlock Action
func (h *Hashlock) Exec_Hlock(hlock *pty.HashlockLock, tx *types.Transaction, index int) (*types.Receipt, error) {
	clog.Debug("hashlocklock action")
	if hlock.Amount <= 0 {
		clog.Warn("hashlock amount <=0")
		return nil, pty.ErrHashlockAmount
	}
	if err := address.CheckAddress(hlock.ToAddress); err != nil {
		clog.Warn("hashlock checkaddress")
		return nil, err
	}
	if err := address.CheckAddress(hlock.ReturnAddress); err != nil {
		clog.Warn("hashlock checkaddress")
		return nil, err
	}
	if hlock.ReturnAddress != tx.From() {
		clog.Warn("hashlock return address")
		return nil, pty.ErrHashlockReturnAddrss
	}

	if hlock.Time <= minLockTime {
		clog.Warn("exec hashlock time not enough")
		return nil, pty.ErrHashlockTime
	}
	actiondb := NewAction(h, tx, drivers.ExecAddress(string(tx.Execer)))
	return actiondb.Hashlocklock(hlock)
}

// Exec_Hsend Action
func (h *Hashlock) Exec_Hsend(transfer *pty.HashlockSend, tx *types.Transaction, index int) (*types.Receipt, error) {
	//unlock 有两个条件： 1. 时间已经过期 2. 密码是对的，返回原来的账户
	clog.Debug("hashlockunlock action")
	actiondb := NewAction(h, tx, drivers.ExecAddress(string(tx.Execer)))
	return actiondb.Hashlocksend(transfer)
}

// Exec_Hunlock Action
func (h *Hashlock) Exec_Hunlock(transfer *pty.HashlockUnlock, tx *types.Transaction, index int) (*types.Receipt, error) {
	//send 有两个条件：1. 时间没有过期 2. 密码是对的，币转移到 ToAddress
	clog.Debug("hashlocksend action")
	actiondb := NewAction(h, tx, drivers.ExecAddress(string(tx.Execer)))
	return actiondb.Hashlockunlock(transfer)
}
