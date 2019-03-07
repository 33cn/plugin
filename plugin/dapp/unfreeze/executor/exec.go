// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/account"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

// Exec_Create 执行创建冻结合约
func (u *Unfreeze) Exec_Create(payload *pty.UnfreezeCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload.AssetExec == "" || payload.AssetSymbol == "" || payload.TotalCount <= 0 || payload.Means == "" {
		return nil, types.ErrInvalidParam
	}

	unfreeze, err := u.newEntity(payload, tx)
	if err != nil {
		uflog.Error("unfreeze create entity", "addr", tx.From(), "payload", payload)
		return nil, err
	}

	receipt1, err := u.create(unfreeze)
	if err != nil {
		uflog.Error("unfreeze create order", "addr", tx.From(), "unfreeze", unfreeze)
		return nil, err
	}

	acc, err := account.NewAccountDB(payload.AssetExec, payload.AssetSymbol, u.GetStateDB())
	if err != nil {
		uflog.Error("unfreeze create new account", "addr", tx.From(), "execAddr",
			dapp.ExecAddress(string(tx.Execer)), "exec", payload.AssetExec, "symbol", payload.AssetSymbol)
		return nil, err
	}
	receipt, err := acc.ExecFrozen(unfreeze.Initiator, dapp.ExecAddress(string(tx.Execer)), payload.TotalCount)
	if err != nil {
		uflog.Error("unfreeze create exec frozen", "addr", tx.From(), "execAddr", dapp.ExecAddress(string(tx.Execer)),
			"ExecFrozen amount", payload.TotalCount, "exec", payload.AssetExec, "symbol", payload.AssetSymbol)
		return nil, err
	}

	return mergeReceipt(receipt, receipt1)
}

// Exec_Withdraw 执行冻结合约中提币
func (u *Unfreeze) Exec_Withdraw(payload *pty.UnfreezeWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	if types.IsDappFork(u.GetHeight(), pty.UnfreezeX, pty.ForkUnfreezeIDX) {
		payload.UnfreezeID = unfreezeIDFromHex(payload.UnfreezeID)
	}
	unfreeze, err := loadUnfreeze(payload.UnfreezeID, u.GetStateDB())
	if err != nil {
		return nil, err
	}
	if unfreeze.Beneficiary != tx.From() {
		uflog.Error("unfreeze withdraw no privilege", "beneficiary", unfreeze.Beneficiary, "txFrom", tx.From())
		return nil, pty.ErrNoPrivilege
	}
	if unfreeze.Remaining <= 0 {
		uflog.Error("unfreeze withdraw no asset")
		return nil, pty.ErrUnfreezeEmptied
	}

	amount, receipt1, err := u.withdraw(unfreeze)
	if err != nil {
		uflog.Error("unfreeze withdraw withdraw", "err", err, "unfreeze", unfreeze)
		return nil, err
	}

	acc, err := account.NewAccountDB(unfreeze.AssetExec, unfreeze.AssetSymbol, u.GetStateDB())
	if err != nil {
		return nil, err
	}
	execAddr := dapp.ExecAddress(string(tx.Execer))
	receipt, err := acc.ExecTransferFrozen(unfreeze.Initiator, tx.From(), execAddr, amount)
	if err != nil {
		uflog.Error("unfreeze withdraw transfer", "execaddr", execAddr, "err", err, "from", unfreeze.Initiator,
			"remain", unfreeze.Remaining, "withdraw", amount)
		return nil, err
	}

	return mergeReceipt(receipt, receipt1)
}

// Exec_Terminate 执行终止冻结合约
func (u *Unfreeze) Exec_Terminate(payload *pty.UnfreezeTerminate, tx *types.Transaction, index int) (*types.Receipt, error) {
	if types.IsDappFork(u.GetHeight(), pty.UnfreezeX, pty.ForkUnfreezeIDX) {
		payload.UnfreezeID = unfreezeIDFromHex(payload.UnfreezeID)
	}
	unfreeze, err := loadUnfreeze(payload.UnfreezeID, u.GetStateDB())
	if err != nil {
		return nil, err
	}
	if tx.From() != unfreeze.Initiator {
		uflog.Error("unfreeze terminate no privilege", "err", pty.ErrUnfreezeID, "initiator",
			unfreeze.Initiator, "from", tx.From())
		return nil, pty.ErrNoPrivilege
	}

	amount, receipt1, err := u.terminator(unfreeze)
	if err != nil {
		uflog.Error("unfreeze terminate ", "err", err, "unfreeze", unfreeze)
		return nil, err
	}

	acc, err := account.NewAccountDB(unfreeze.AssetExec, unfreeze.AssetSymbol, u.GetStateDB())
	if err != nil {
		return nil, err
	}
	execAddr := dapp.ExecAddress(string(tx.Execer))
	receipt, err := acc.ExecActive(unfreeze.Initiator, execAddr, amount)
	if err != nil {
		uflog.Error("unfreeze terminate ", "addr", unfreeze.Initiator, "execaddr", execAddr, "err", err)
		return nil, err
	}
	return mergeReceipt(receipt, receipt1)
}

func (u *Unfreeze) newEntity(payload *pty.UnfreezeCreate, tx *types.Transaction) (*pty.Unfreeze, error) {
	id := unfreezeID(tx.Hash())
	unfreeze := &pty.Unfreeze{
		UnfreezeID:  string(id),
		StartTime:   payload.StartTime,
		AssetExec:   payload.AssetExec,
		AssetSymbol: payload.AssetSymbol,
		TotalCount:  payload.TotalCount,
		Remaining:   payload.TotalCount,
		Initiator:   tx.From(),
		Beneficiary: payload.Beneficiary,
		Means:       payload.Means,
	}
	if unfreeze.StartTime == 0 {
		unfreeze.StartTime = u.GetBlockTime()
	}
	means, err := newMeans(payload.Means, u.GetHeight())
	if err != nil {
		return nil, err

	}
	unfreeze, err = means.setOpt(unfreeze, payload)
	if err != nil {
		return nil, err
	}
	return unfreeze, nil
}

// 创建解冻状态
func (u *Unfreeze) create(unfreeze *pty.Unfreeze) (*types.Receipt, error) {
	k := []byte(unfreeze.UnfreezeID)
	v := types.Encode(unfreeze)
	err := u.GetStateDB().Set(k, v)
	if err != nil {
		return nil, err
	}

	receiptLog := getUnfreezeLog(nil, unfreeze, pty.TyLogCreateUnfreeze)
	return &types.Receipt{Ty: types.ExecOk,
		KV: []*types.KeyValue{{Key: k, Value: v}}, Logs: []*types.ReceiptLog{receiptLog}}, nil
}

func mergeReceipt(r1 *types.Receipt, r2 *types.Receipt) (*types.Receipt, error) {
	r1.Logs = append(r1.Logs, r2.Logs...)
	r1.KV = append(r1.KV, r2.KV...)
	r1.Ty = types.ExecOk
	return r1, nil
}

func getUnfreezeLog(prev, cur *pty.Unfreeze, ty int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = ty
	r := &pty.ReceiptUnfreeze{Prev: prev, Current: cur}
	log.Log = types.Encode(r)
	return log
}

// 提取解冻币
func (u *Unfreeze) withdraw(unfreeze *pty.Unfreeze) (int64, *types.Receipt, error) {
	means, err := newMeans(unfreeze.Means, u.GetHeight())
	if err != nil {
		return 0, nil, err

	}
	frozen, err := means.calcFrozen(unfreeze, u.GetBlockTime())
	if err != nil {
		return 0, nil, err
	}
	unfreezeOld := *unfreeze
	unfreeze, amount := withdraw(unfreeze, frozen)
	receiptLog := getUnfreezeLog(&unfreezeOld, unfreeze, pty.TyLogWithdrawUnfreeze)

	k := []byte(unfreeze.UnfreezeID)
	v := types.Encode(unfreeze)
	err = u.GetStateDB().Set(k, v)
	if err != nil {
		return 0, nil, err
	}

	return amount, &types.Receipt{Ty: types.ExecOk, KV: []*types.KeyValue{{Key: k, Value: v}},
		Logs: []*types.ReceiptLog{receiptLog}}, nil
}

// 中止定期解冻
func (u *Unfreeze) terminator(unfreeze *pty.Unfreeze) (int64, *types.Receipt, error) {
	if unfreeze.Remaining <= 0 {
		return 0, nil, pty.ErrUnfreezeEmptied
	}

	unfreezeOld := *unfreeze
	var amount int64
	if types.IsDappFork(u.GetHeight(), pty.UnfreezeX, "ForkTerminatePart") {
		if unfreeze.Terminated {
			return 0, nil, pty.ErrTerminated
		}
		m, err := newMeans(unfreeze.Means, u.GetHeight())
		if err != nil {
			return 0, nil, err
		}
		frozen, err := m.calcFrozen(unfreeze, u.GetBlockTime())
		if err != nil {
			return 0, nil, err
		}
		amount = frozen
		unfreeze.Remaining = unfreeze.Remaining - amount
		unfreeze.Terminated = true
	} else {
		amount = unfreeze.Remaining
		unfreeze.Remaining = 0
	}
	receiptLog := getUnfreezeLog(&unfreezeOld, unfreeze, pty.TyLogTerminateUnfreeze)

	k := []byte(unfreeze.UnfreezeID)
	v := types.Encode(unfreeze)
	err := u.GetStateDB().Set(k, v)
	if err != nil {
		return 0, nil, err
	}

	return amount, &types.Receipt{Ty: types.ExecOk, KV: []*types.KeyValue{{Key: k, Value: v}},
		Logs: []*types.ReceiptLog{receiptLog}}, nil

}

func loadUnfreeze(id string, db dbm.KV) (*pty.Unfreeze, error) {
	value, err := db.Get([]byte(id))
	if err != nil {
		uflog.Error("unfreeze terminate get", "id", id, "err", err)
		return nil, err
	}
	var unfreeze pty.Unfreeze
	err = types.Decode(value, &unfreeze)
	if err != nil {
		uflog.Error("unfreeze terminate decode", "err", err)
		return nil, err
	}
	return &unfreeze, nil
}
