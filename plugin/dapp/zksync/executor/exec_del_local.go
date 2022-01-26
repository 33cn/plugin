package executor

import (
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func (z *zksync) ExecDelLocal_Deposit(payload *zt.ZkDeposit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_Withdraw(payload *zt.ZkWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_ContractToLeaf(payload *zt.ZkContractToLeaf, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_LeafToContract(payload *zt.ZkLeafToContract, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_Transfer(payload *zt.ZkTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_TransferToNew(payload *zt.ZkTransferToNew, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_ForceExit(payload *zt.ZkForceExit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_SetPubKey(payload *zt.ZkSetPubKey, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}
