package executor

import (
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func (e *zksync) Exec_Deposit(payload *zt.Deposit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.Deposit(payload)
}

func (e *zksync) Exec_Withdraw(payload *zt.Withdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.Withdraw(payload)
}

func (e *zksync) Exec_Contract_To_Leaf(payload *zt.ContractToLeaf, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.Withdraw(payload)
}

func (e *zksync) Exec_Leaf_To_Leaf(payload *zt.LeafToContract, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.Withdraw(payload)
}

func (e *zksync) Exec_Transfer(payload *zt.Transfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.Transfer(payload)
}

func (e *zksync) Exec_TransferToNew(payload *zt.TransferToNew, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.TransferToNew(payload)
}

func (e *zksync) Exec_ForceQuit(payload *zt.ForceQuit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.ForceQuit(payload)
}
