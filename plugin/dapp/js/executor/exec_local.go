package executor

import (
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) ExecLocal_Create(payload *jsproto.Create, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (c *js) ExecLocal_Call(payload *jsproto.Call, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	jsvalue, err := c.callVM("execlocal", payload, tx, index)
	if err != nil {
		return nil, err
	}
	r := &types.LocalDBSet{}
	kvs, err := parseKVS(jsvalue)
	if err != nil {
		return nil, err
	}
	r.KV = kvs
	return r, nil
}
