package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) ExecDelLocal_Create(payload *jsproto.Create, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (c *js) ExecDelLocal_Call(payload *jsproto.Call, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	krollback := calcRollbackKey(tx.Hash())
	kvc := dapp.NewKVCreator(c.GetLocalDB(), calcLocalPrefix(tx.Execer), krollback)
	kvs, err := kvc.GetRollbackKVList()
	if err != nil {
		return nil, err
	}
	kvc.AddKVListOnly(kvs)
	kvc.DelRollbackKV()
	r := &types.LocalDBSet{}
	r.KV = kvc.KVList()
	return r, nil
}
