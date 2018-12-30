package executor

import (
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) Exec_Create(payload *jsproto.Create, tx *types.Transaction, index int) (*types.Receipt, error) {
	db := c.GetStateDB()
	_, err := db.Get(calcCodeKey(payload.Name))
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	if err == nil {
		return nil, ptypes.ErrDupName
	}
	r := &types.Receipt{Ty: types.ExecOk}
	//code must be utf-8 encoding
	r.KV = append(r.KV, &types.KeyValue{
		Key: calcCodeKey(payload.Name), Value: []byte(payload.Code)})
	return r, nil
}

func (c *js) Exec_Call(payload *jsproto.Call, tx *types.Transaction, index int) (*types.Receipt, error) {
	jsvalue, err := c.callVM("exec", payload, tx, index)
	if err != nil {
		return nil, err
	}
	r := &types.Receipt{Ty: types.ExecOk}
	kvs, err := parseKVS(jsvalue)
	if err != nil {
		return nil, err
	}
	r.KV = kvs
	return r, nil
}
