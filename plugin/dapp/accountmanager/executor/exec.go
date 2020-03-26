package executor

import (
	"github.com/33cn/chain33/types"
	aty "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (a *Accountmanager) Exec_Register(payload *aty.Register, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(a, tx, index)
	return action.Register(payload)
}

func (a *Accountmanager) Exec_ResetKey(payload *aty.ResetKey, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(a, tx, index)
	return action.Reset(payload)
}

func (a *Accountmanager) Exec_Transfer(payload *aty.Transfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(a, tx, index)
	return action.Transfer(payload)
}

func (a *Accountmanager) Exec_Supervise(payload *aty.Supervise, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(a, tx, index)
	return action.Supervise(payload)
}

func (a *Accountmanager) Exec_Apply(payload *aty.Apply, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(a, tx, index)
	return action.Apply(payload)
}
