package executor

import (
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) ExecDelLocal_Create(payload *jsproto.Create, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (c *js) ExecDelLocal_Call(payload *jsproto.Call, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	execer := types.ExecName("user." + ptypes.JsX + "." + payload.Name)
	r := &types.LocalDBSet{}
	c.prefix = types.CalcLocalPrefix([]byte(execer))
	kvs, err := c.DelRollbackKV(tx, []byte(execer))
	if err != nil {
		return nil, err
	}
	r.KV = kvs
	return r, nil
}
