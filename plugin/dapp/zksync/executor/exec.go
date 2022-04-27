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

func (z *zksync) Exec_ContractToTree(payload *zt.ZkContractToTree, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.ContractToTree(payload)
}

func (z *zksync) Exec_TreeToContract(payload *zt.ZkTreeToContract, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.TreeToContract(payload)
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

func (z *zksync) Exec_FullExit(payload *zt.ZkFullExit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.FullExit(payload)
}

func (z *zksync) Exec_Swap(payload *zt.ZkSwap, tx *types.Transaction, index int) (*types.Receipt, error) {
	//todo swap stub
	return nil, nil
}

func (z *zksync) Exec_SetVerifyKey(payload *zt.ZkVerifyKey, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.setVerifyKey(payload)
}

func (z *zksync) Exec_CommitProof(payload *zt.ZkCommitProof, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.commitProof(payload)
}

func (z *zksync) Exec_SetVerifier(payload *zt.ZkVerifier, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.setVerifier(payload)
}

func (z *zksync) Exec_SetFee(payload *zt.ZkSetFee, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.setFee(payload)
}

func (z *zksync) Exec_MintNFT(payload *zt.ZkMintNFT, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.MintNFT(payload)
}

func (z *zksync) Exec_WithdrawNFT(payload *zt.ZkWithdrawNFT, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.withdrawNFT(payload)
}

func (z *zksync) Exec_TransferNFT(payload *zt.ZkTransferNFT, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.transferNFT(payload)
}
