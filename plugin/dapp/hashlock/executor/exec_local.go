package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/hashlock/types"
)

func (h *Hashlock) ExecLocal_Hlock(hlock *pty.HashlockLock, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	info := pty.Hashlockquery{hlock.Time, hashlockLocked, hlock.Amount, h.GetBlockTime(), 0}
	clog.Error("ExecLocal", "info", info)
	kv, err := UpdateHashReciver(h.GetLocalDB(), hlock.Hash, info)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: []*types.KeyValue{kv}}, nil
}

func (h *Hashlock) ExecLocal_Hsend(hsend *pty.HashlockSend, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	info := pty.Hashlockquery{0, hashlockSent, 0, 0, 0}
	clog.Error("ExecLocal", "info", info)
	kv, err := UpdateHashReciver(h.GetLocalDB(), common.Sha256(hsend.Secret), info)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: []*types.KeyValue{kv}}, nil
}

func (h *Hashlock) ExecLocal_Hunlock(hunlock *pty.HashlockUnlock, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	info := pty.Hashlockquery{0, hashlockUnlocked, 0, 0, 0}
	clog.Error("ExecLocal", "info", info)
	kv, err := UpdateHashReciver(h.GetLocalDB(), common.Sha256(hunlock.Secret), info)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: []*types.KeyValue{kv}}, nil
}
