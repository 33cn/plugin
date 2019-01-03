package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) ExecDelLocal_Create(payload *jsproto.Create, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (c *js) ExecDelLocal_Call(payload *jsproto.Call, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	krollback := calcRollbackKey(tx.Hash())
	execer := types.ExecName("user." + ptypes.JsX + "." + payload.Name)
	c.prefix = calcLocalPrefix([]byte(execer))
	kvc := dapp.NewKVCreator(c.GetLocalDB(), c.prefix, krollback)
	kvs, err := kvc.GetRollbackKVList()
	if err != nil {
		return nil, err
	}
	for _, kv := range kvs {
		kvc.AddNoPrefix(kv.Key, kv.Value)
	}
	kvc.DelRollbackKV()
	r := &types.LocalDBSet{}
	r.KV = kvc.KVList()
	return r, nil
}
