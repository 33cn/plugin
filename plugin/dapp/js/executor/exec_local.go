package executor

import (
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) ExecLocal_Create(payload *jsproto.Create, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (c *js) ExecLocal_Call(payload *jsproto.Call, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	execer := types.ExecName("user." + ptypes.JsX + "." + payload.Name)
	c.prefix = types.CalcLocalPrefix([]byte(execer))
	jsvalue, err := c.callVM("execlocal", payload, tx, index, receiptData)
	if err != nil {
		return nil, err
	}
	kvs, _, err := parseJsReturn(c.prefix, jsvalue)
	if err != nil {
		return nil, err
	}
	r := &types.LocalDBSet{}
	r.KV = c.AddRollbackKV(tx, []byte(execer), kvs)
	return r, nil
}
