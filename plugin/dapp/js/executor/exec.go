package executor

import (
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) Exec_Create(payload *jsproto.Create, tx *types.Transaction, index int) (*types.Receipt, error) {
	return &types.Receipt{}, nil
}

func (c *js) Exec_Call(payload *jsproto.Call, tx *types.Transaction, index int) (*types.Receipt, error) {
	return &types.Receipt{}, nil
}
