package executor

import (
	uf "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/types"
)

func (u *Unfreeze) Query_QueryUnfreezeWithdraw(in *uf.QueryUnfreezeWithdraw) (types.Message, error) {
	return QueryUnfreezeWithdraw(u.GetStateDB(), in)
}
