package executor

import (
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)


//ExecLocal_Deposit asset withdraw local db process
func (z *zksync) ExecLocal_Deposit(payload *zt.ZkDeposit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

//ExecLocal_Withdraw asset withdraw local db process
func (z *zksync) ExecLocal_Withdraw(payload *zt.ZkWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

// ExecLocal_Transfer asset transfer local db process
func (z *zksync) ExecLocal_ContractToLeaf(payload *zt.ZkContractToLeaf, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

//ExecLocal_Authorize asset withdraw local db process
func (z *zksync) ExecLocal_LeafToContract(payload *zt.ZkLeafToContract, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

func (z *zksync) ExecLocal_Transfer(payload *zt.ZkTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

func (z *zksync) ExecLocal_TransferToNew(payload *zt.ZkTransferToNew, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

func (z *zksync) ExecLocal_ForceExit(payload *zt.ZkForceExit, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

func (z *zksync) ExecLocal_SetPubKey(payload *zt.ZkSetPubKey, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return z.execAutoLocalZksync(tx, receiptData, index)
}

