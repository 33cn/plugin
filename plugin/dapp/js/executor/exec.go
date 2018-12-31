package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) Exec_Create(payload *jsproto.Create, tx *types.Transaction, index int) (*types.Receipt, error) {
	execer := types.ExecName("user.js." + payload.Name)
	kvc := dapp.NewKVCreator(c.GetStateDB(), calcStatePrefix([]byte(execer)), nil)
	_, err := kvc.GetNoPrefix(calcCodeKey(payload.Name))
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	if err == nil {
		return nil, ptypes.ErrDupName
	}
	kvc.AddNoPrefix(calcCodeKey(payload.Name), []byte(payload.Code))
	jsvalue, err := c.callVM("init", &jsproto.Call{Name: payload.Name}, tx, index, nil)
	if err != nil {
		return nil, err
	}
	kvs, logs, err := parseJsReturn(jsvalue)
	if err != nil {
		return nil, err
	}
	kvc.AddList(kvs)
	r := &types.Receipt{Ty: types.ExecOk, KV: kvc.KVList(), Logs: logs}
	return r, nil
}

func (c *js) Exec_Call(payload *jsproto.Call, tx *types.Transaction, index int) (*types.Receipt, error) {
	execer := types.ExecName("user.js." + payload.Name)
	kvc := dapp.NewKVCreator(c.GetStateDB(), calcStatePrefix([]byte(execer)), nil)
	jsvalue, err := c.callVM("exec", payload, tx, index, nil)
	if err != nil {
		return nil, err
	}
	kvs, logs, err := parseJsReturn(jsvalue)
	if err != nil {
		return nil, err
	}
	kvc.AddList(kvs)
	r := &types.Receipt{Ty: types.ExecOk, KV: kvc.KVList(), Logs: logs}
	return r, nil
}
