package executor

import (
	"github.com/33cn/chain33/types"
	aty "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (a *accountmanager) Exec_Register(payload *aty.Register, tx *types.Transaction, index int) (*types.Receipt, error) {
	var receipt *types.Receipt
	//implement code
	return receipt, nil
}

func (a *accountmanager) Exec_Reset(payload *aty.Reset, tx *types.Transaction, index int) (*types.Receipt, error) {
	var receipt *types.Receipt
	//implement code
	return receipt, nil
}


func (a *accountmanager) Exec_Transfer(payload *aty.Transfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	var receipt *types.Receipt
	//implement code
	return receipt, nil
}

func (a *accountmanager) Exec_Supervise(payload *aty.Supervise, tx *types.Transaction, index int) (*types.Receipt, error) {
	var receipt *types.Receipt
	//implement code
	return receipt, nil
}

func (a *accountmanager) ExecApply(payload *aty.Apply, tx *types.Transaction, index int) (*types.Receipt, error) {
	var receipt *types.Receipt
	//implement code
	return receipt, nil
}