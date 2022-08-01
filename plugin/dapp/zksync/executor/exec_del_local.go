package executor

import (
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func (z *zksync) ExecDelLocal_Deposit(payload *zt.ZkDeposit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_ZkWithdraw(payload *zt.ZkWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_ContractToTree(payload *zt.ZkContractToTree, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_TreeToContract(payload *zt.ZkTreeToContract, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_ZkTransfer(payload *zt.ZkTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_TransferToNew(payload *zt.ZkTransferToNew, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_ProxyExit(payload *zt.ZkProxyExit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_SetPubKey(payload *zt.ZkSetPubKey, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_FullExit(payload *zt.ZkFullExit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_MintNFT(payload *zt.ZkMintNFT, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_WithdrawNFT(payload *zt.ZkWithdrawNFT, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_TransferNFT(payload *zt.ZkTransferNFT, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}
func (z *zksync) ExecDelLocal_TransferToExec(payload *types.AssetsTransferToExec, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}
func (z *zksync) ExecDelLocal_Withdraw(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_CommitProof(payload *zt.ZkCommitProof, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}
