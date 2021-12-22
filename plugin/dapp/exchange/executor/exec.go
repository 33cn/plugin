package executor

import (
	"github.com/33cn/chain33/types"
	exchangetypes "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (e *exchange) Exec_LimitOrder(payload *exchangetypes.LimitOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.LimitOrder(payload)
}

func (e *exchange) Exec_MarketOrder(payload *exchangetypes.MarketOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	//TODO marketOrder
	return nil, types.ErrActionNotSupport
}

func (e *exchange) Exec_RevokeOrder(payload *exchangetypes.RevokeOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.RevokeOrder(payload)
}

func (e *exchange) Exec_Deposit(payload *exchangetypes.Deposit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.Deposit(payload)
}

func (e *exchange) Exec_Withdraw(payload *exchangetypes.Withdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.Withdraw(payload)
}

func (e *exchange) Exec_Transfer(payload *exchangetypes.Transfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.Transfer(payload)
}

func (e *exchange) Exec_TransferToNew(payload *exchangetypes.TransferToNew, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.TransferToNew(payload)
}

func (e *exchange) Exec_ForceQuit(payload *exchangetypes.ForceQuit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.ForceQuit(payload)
}


