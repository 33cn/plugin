package executor

import (
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func (z *zksync) Exec_Deposit(payload *zt.Deposit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.Deposit(payload)
}

func (z *zksync) Exec_Withdraw(payload *zt.Withdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.Withdraw(payload)
}

func (z *zksync) Exec_ContractToLeaf(payload *zt.ContractToLeaf, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.ContractToLeaf(payload)
}

func (z *zksync) Exec_LeafToContract(payload *zt.LeafToContract, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.LeafToContract(payload)
}

func (z *zksync) Exec_Transfer(payload *zt.Transfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.Transfer(payload)
}

func (z *zksync) Exec_TransferToNew(payload *zt.TransferToNew, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.TransferToNew(payload)
}

func (z *zksync) Exec_ForceQuit(payload *zt.ForceQuit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.ForceQuit(payload)
}

func (z *zksync) Exec_SetPubKey(payload *zt.SetPubKey, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.SetPubKey(payload)
}

func (z *zksync) Exec_SetVerifyKey(payload *zt.VerifyKey, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.setVerifyKey(payload)
}

func (z *zksync) Exec_CommitProof(payload *zt.CommitProof, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.commitProof(payload)
}
