package executor

import (
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func (z *zksync) Exec_Deposit(payload *zt.ZkDeposit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.Deposit(payload)
}

func (z *zksync) Exec_Withdraw(payload *zt.ZkWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.Withdraw(payload)
}

func (z *zksync) Exec_ContractToLeaf(payload *zt.ZkContractToLeaf, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.ContractToLeaf(payload)
}

func (z *zksync) Exec_LeafToContract(payload *zt.ZkLeafToContract, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.LeafToContract(payload)
}

func (z *zksync) Exec_Transfer(payload *zt.ZkTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.Transfer(payload)
}

func (z *zksync) Exec_TransferToNew(payload *zt.ZkTransferToNew, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.TransferToNew(payload)
}

func (z *zksync) Exec_ForceExit(payload *zt.ZkForceExit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.ForceExit(payload)
}

func (z *zksync) Exec_SetPubKey(payload *zt.ZkSetPubKey, tx *types.Transaction, index int) (*types.Receipt, error) {
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
