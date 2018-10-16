package executor

import (
	uf "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/types"
)

func (u *Unfreeze) Exec_Create(payload *uf.GameCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(u, tx, index)
	return action.UnfreezeCreate(payload)
}

func (u *Unfreeze) Exec_Cancel(payload *uf.GameCancel, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(u, tx, index)
	return action.UnfreezeWithdraw(payload)
}

func (u *Unfreeze) Exec_Terminate(payload *uf.GameClose, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(u, tx, index)
	return action.UnfreezeTerminate(payload)
}
