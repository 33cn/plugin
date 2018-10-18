package executor

import (
	uf "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/types"
)

func (u *Unfreeze) execLocal(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	
	
	dbSet := &types.LocalDBSet{}
	return dbSet, nil
}

func (u *Unfreeze) ExecLocal_Create(payload *uf.UnfreezeCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execLocal(receiptData)
}

func (u *Unfreeze) ExecLocal_Withdraw(payload *uf.UnfreezeWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execLocal(receiptData)
}

func (u *Unfreeze) ExecLocal_Terminate(payload *uf.UnfreezeTerminate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return u.execLocal(receiptData)
}
