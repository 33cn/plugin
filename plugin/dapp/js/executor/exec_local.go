package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) ExecLocal_Create(payload *jsproto.Create, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (c *js) ExecLocal_Call(payload *jsproto.Call, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	k := calcRollbackKey(tx.Hash())
	execer := types.ExecName("user.js." + payload.Name)
	c.prefix = calcLocalPrefix([]byte(execer))
	kvc := dapp.NewKVCreator(c.GetLocalDB(), c.prefix, k)
	jsvalue, err := c.callVM("execlocal", payload, tx, index, receiptData)
	if err != nil {
		return nil, err
	}
	kvs, _, err := parseJsReturn(jsvalue)
	if err != nil {
		return nil, err
	}
	kvc.AddList(kvs)
	kvc.AddRollbackKV()
	r := &types.LocalDBSet{}
	r.KV = kvc.KVList()
	return r, nil
}
