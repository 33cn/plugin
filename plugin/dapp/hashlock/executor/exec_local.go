// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/hashlock/types"
)

// ExecLocal_Hlock Action
func (h *Hashlock) ExecLocal_Hlock(hlock *pty.HashlockLock, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	info := pty.Hashlockquery{Time: hlock.Time, Status: hashlockLocked, Amount: hlock.Amount, CreateTime: h.GetBlockTime(), CurrentTime: 0}
	clog.Error("ExecLocal", "info", info)
	kv, err := UpdateHashReciver(h.GetLocalDB(), hlock.Hash, info)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: []*types.KeyValue{kv}}, nil
}

// ExecLocal_Hsend Action
func (h *Hashlock) ExecLocal_Hsend(hsend *pty.HashlockSend, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	info := pty.Hashlockquery{Time: 0, Status: hashlockSent, Amount: 0, CreateTime: 0, CurrentTime: 0}
	clog.Error("ExecLocal", "info", info)
	kv, err := UpdateHashReciver(h.GetLocalDB(), common.Sha256(hsend.Secret), info)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: []*types.KeyValue{kv}}, nil
}

// ExecLocal_Hunlock Action
func (h *Hashlock) ExecLocal_Hunlock(hunlock *pty.HashlockUnlock, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	info := pty.Hashlockquery{Time: 0, Status: hashlockUnlocked, Amount: 0, CreateTime: 0, CurrentTime: 0}
	clog.Error("ExecLocal", "info", info)
	kv, err := UpdateHashReciver(h.GetLocalDB(), common.Sha256(hunlock.Secret), info)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: []*types.KeyValue{kv}}, nil
}
