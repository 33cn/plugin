package executor

import (
	uf "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/types"
)

func (u *Unfreeze) Exec_Create(payload *uf.UnfreezeCreate, tx *types.Transaction, index int32) (*types.Receipt, error) {
	action := newAction(u, tx, index)
	return action.UnfreezeCreate(payload)
}

func (u *Unfreeze) Exec_Withdraw(payload *uf.UnfreezeWithdraw, tx *types.Transaction, index int32) (*types.Receipt, error) {
	action := newAction(u, tx, index)
	return action.UnfreezeWithdraw(payload)
}

func (u *Unfreeze) Exec_Terminate(payload *uf.UnfreezeTerminate, tx *types.Transaction, index int32) (*types.Receipt, error) {
	action := newAction(u, tx, index)
	return action.UnfreezeTerminate(payload)
}
