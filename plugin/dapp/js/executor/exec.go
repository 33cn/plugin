package executor

import (
	"strings"

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

func (c *js) checkJsName(name string) bool {
	if types.IsPara() {
		return name == types.GetTitle()+ptypes.JsX
	}
	return name == ptypes.JsX
}

func (c *js) checkJsPrefix(name string) bool {
	if types.IsPara() {
		return strings.HasPrefix(name, types.GetTitle()+"user."+ptypes.JsX)
	}
	return strings.HasPrefix(name, "user."+ptypes.JsX)
}

func (c *js) Exec_Create(payload *jsproto.Create, tx *types.Transaction, index int) (*types.Receipt, error) {
	err := checkPriv(tx.From(), ptypes.JsCreator, c.GetStateDB())
	if err != nil {
		return nil, err
	}

	execer := c.userExecName(payload.Name, false)
	if !c.checkJsName(string(tx.Execer)) {
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
	if !c.checkJsPrefix(string(tx.Execer)) {
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
