package executor

import (
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)


//ExecLocal_Deposit asset withdraw local db process
func (z *zksync) ExecLocal_Deposit(payload *zt.Deposit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

//ExecLocal_Withdraw asset withdraw local db process
func (z *zksync) ExecLocal_Withdraw(payload *zt.Withdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

// ExecLocal_Transfer asset transfer local db process
func (z *zksync) ExecLocal_ContractToLeaf(payload *zt.ContractToLeaf, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

//ExecLocal_Authorize asset withdraw local db process
func (z *zksync) ExecLocal_LeafToContract(payload *zt.LeafToContract, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

func (z *zksync) ExecLocal_Transfer(payload *zt.Transfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

func (z *zksync) ExecLocal_TransferToNew(payload *zt.TransferToNew, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

func (z *zksync) ExecLocal_ForceQuit(payload *zt.ForceQuit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

func (z *zksync) ExecLocal_SetPubKey(payload *zt.SetPubKey, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

