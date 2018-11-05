package executor

import (
	"gitlab.33.cn/chain33/chain33/account"
	pty "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/system/dapp"
	"gitlab.33.cn/chain33/chain33/types"
)

func (u *Unfreeze) Exec_Create(payload *pty.UnfreezeCreate, tx *types.Transaction, index int32) (*types.Receipt, error) {
	if payload.AssetSymbol == "" || payload.AssetSymbol == "" || payload.TotalCount <= 0 || payload.Means == "" {
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

	return MergeReceipt(receipt, receipt1)
}

func (u *Unfreeze) Exec_Withdraw(payload *pty.UnfreezeWithdraw, tx *types.Transaction, index int32) (*types.Receipt, error) {
	value, err := u.GetStateDB().Get([]byte(payload.UnfreezeID))
	if err != nil {
		uflog.Error("unfreeze withdraw get", "id", payload.UnfreezeID, "err", err)
		return nil, err
	}
	var unfreeze pty.Unfreeze
	err = types.Decode(value, &unfreeze)
	if err != nil {
		uflog.Error("unfreeze withdraw decode", "err", err)
		return nil, err
	}
	if unfreeze.Beneficiary != tx.From() {
		uflog.Error("unfreeze withdraw no privilege", "beneficiary", unfreeze.Beneficiary, "txFrom", tx.From())
		return nil, pty.ErrNoPrivilege
	}

	amount, receipt1, err := u.withdraw(&unfreeze)
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
		uflog.Error("unfreeze withdraw transfer", "execaddr", execAddr, "err", err)
		return nil, err
	}

	return MergeReceipt(receipt, receipt1)
}

func (u *Unfreeze) Exec_Terminate(payload *pty.UnfreezeTerminate, tx *types.Transaction, index int32) (*types.Receipt, error) {
	value, err := u.GetStateDB().Get([]byte(payload.UnfreezeID))
	if err != nil {
		uflog.Error("unfreeze terminate get", "id", payload.UnfreezeID, "err", err)
		return nil, err
	}
	var unfreeze pty.Unfreeze
	err = types.Decode(value, &unfreeze)
	if err != nil {
		uflog.Error("unfreeze terminate decode", "err", err)
		return nil, err
	}
	if tx.From() != unfreeze.Initiator {
		uflog.Error("unfreeze terminate no privilege", "err", pty.ErrUnfreezeID, "initiator",
			unfreeze.Initiator, "from", tx.From())
		return nil, pty.ErrUnfreezeID
	}

	amount, receipt1, err := u.terminator(&unfreeze)
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
	return MergeReceipt(receipt, receipt1)
}

func (u *Unfreeze) newEntity(payload *pty.UnfreezeCreate, tx *types.Transaction) (*pty.Unfreeze, error) {
	id := unfreezeID(string(tx.Hash()))
	unfreeze := &pty.Unfreeze{
		UnfreezeID:  string(id),
		StartTime:   payload.StartTime,
		AssetExec:   payload.AssetExec,
		AssetSymbol: payload.AssetSymbol,
		TotalCount:  payload.TotalCount,
		Initiator:   tx.From(),
		Beneficiary: payload.Beneficiary,
		Means:       payload.Means,
	}
	means, err := newMeans(payload.Means)
	if err != nil {
		return nil, err

	}
	unfreeze, err = means.setOpt(unfreeze, payload)
	if err != nil {
		return nil, err
	}
	return unfreeze, nil
}

// 创建解冻交易
func (u *Unfreeze) create(unfreeze *pty.Unfreeze) (*types.Receipt, error) {
	k := []byte(unfreeze.UnfreezeID)
	v := types.Encode(unfreeze)
	err := u.GetStateDB().Set(k, v)
	if err != nil {
		return nil, err
	}

	receiptLog := getUnfreezeLog(nil, unfreeze)
	return &types.Receipt{KV: []*types.KeyValue{{k, v}}, Logs: []*types.ReceiptLog{receiptLog}}, nil
}

func MergeReceipt(r1 *types.Receipt, r2 *types.Receipt) (*types.Receipt, error) {
	r1.Logs = append(r1.Logs, r2.Logs...)
	r1.KV = append(r1.KV, r2.KV...)
	r1.Ty = types.ExecOk
	return r1, nil
}

func getUnfreezeLog(prev, cur *pty.Unfreeze) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogCreateUnfreeze
	r := &pty.ReceiptUnfreeze{Prev: prev, Cur: cur}
	log.Log = types.Encode(r)
	return log
}

// 提取解冻币
func (u *Unfreeze) withdraw(unfreeze *pty.Unfreeze) (int64, *types.Receipt, error) {
	means, err := newMeans(unfreeze.Means)
	if err != nil {
		return 0, nil, err

	}
	frozen, err := means.calcFrozen(unfreeze, u.GetBlockTime())
	if err != nil {
		return 0, nil, err
	}
	unfreezeOld := *unfreeze
	unfreeze, amount := withdraw(unfreeze, frozen)
	receiptLog := getUnfreezeLog(&unfreezeOld, unfreeze)

	k := []byte(unfreeze.UnfreezeID)
	v := types.Encode(unfreeze)
	err = u.GetStateDB().Set(k, v)
	if err != nil {
		return 0, nil, err
	}

	return amount, &types.Receipt{Ty: types.ExecOk, KV: []*types.KeyValue{{k, v}},
		Logs: []*types.ReceiptLog{receiptLog}}, nil
}

// 中止定期解冻
func (u *Unfreeze) terminator(unfreeze *pty.Unfreeze) (int64, *types.Receipt, error) {
	if unfreeze.Remaining <= 0 {
		return 0, nil, pty.ErrUnfreezeEmptied
	}

	unfreezeOld := *unfreeze
	amount := unfreeze.Remaining
	unfreeze.Remaining = 0
	receiptLog := getUnfreezeLog(&unfreezeOld, unfreeze)

	k := []byte(unfreeze.UnfreezeID)
	v := types.Encode(unfreeze)
	err := u.GetStateDB().Set(k, v)
	if err != nil {
		return 0, nil, err
	}

	return amount, &types.Receipt{Ty: types.ExecOk, KV: []*types.KeyValue{{k, v}},
		Logs: []*types.ReceiptLog{receiptLog}}, nil

}
