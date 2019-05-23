package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) userExecName(name string, local bool) string {
	execer := "user." + ptypes.JsX + "." + name
	if local {
		execer = types.ExecName(execer)
	}
	return execer
}

// execName 在create 时为 jsvm
// 在 call 时为 user.jsvm.game
func (c *js) checkTxExec(txExec string, execName string) bool {
	return txExec == types.ExecName(execName)
}

func (c *js) Exec_Create(payload *jsproto.Create, tx *types.Transaction, index int) (*types.Receipt, error) {
	err := checkPriv(tx.From(), ptypes.JsCreator, c.GetStateDB())
	if err != nil {
		return nil, err
	}

	execer := c.userExecName(payload.Name, false)
	if !c.checkTxExec(string(tx.Execer), ptypes.JsX) {
		return nil, types.ErrExecNameNotMatch
	}
	c.prefix = types.CalcStatePrefix([]byte(execer))
	kvc := dapp.NewKVCreator(c.GetStateDB(), c.prefix, nil)
	_, err = kvc.GetNoPrefix(calcCodeKey(payload.Name))
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
	kvs, logs, err := parseJsReturn(c.prefix, jsvalue)
	if err != nil {
		return nil, err
	}
	kvc.AddListNoPrefix(kvs)
	r := &types.Receipt{Ty: types.ExecOk, KV: kvc.KVList(), Logs: logs}
	return r, nil
}

func (c *js) Exec_Call(payload *jsproto.Call, tx *types.Transaction, index int) (*types.Receipt, error) {
	execer := c.userExecName(payload.Name, false)
	if !c.checkTxExec(string(tx.Execer), execer) {
		return nil, types.ErrExecNameNotMatch
	}
	c.prefix = types.CalcStatePrefix([]byte(execer))
	kvc := dapp.NewKVCreator(c.GetStateDB(), c.prefix, nil)
	jsvalue, err := c.callVM("exec", payload, tx, index, nil)
	if err != nil {
		return nil, err
	}
	kvs, logs, err := parseJsReturn(c.prefix, jsvalue)
	if err != nil {
		return nil, err
	}
	kvc.AddListNoPrefix(kvs)
	r := &types.Receipt{Ty: types.ExecOk, KV: kvc.KVList(), Logs: logs}
	return r, nil
}
