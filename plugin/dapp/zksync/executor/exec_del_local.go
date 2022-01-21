package executor

import (
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)


func (z *zksync) ExecDelLocal_Deposit(payload *zt.Deposit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_Withdraw(payload *zt.Withdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_ContractToLeaf(payload *zt.ContractToLeaf, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_LeafToContract(payload *zt.LeafToContract, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_Transfer(payload *zt.Transfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_TransferToNew(payload *zt.TransferToNew, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_ForceQuit(payload *zt.ForceQuit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

func (z *zksync) ExecDelLocal_SetPubKey(payload *zt.SetPubKey, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoDelLocal(tx, receiptData)
}

