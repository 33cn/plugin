package executor

import (
	"encoding/json"

	"github.com/33cn/chain33/common"
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
	"github.com/robertkrimen/otto"
)

var (
	ptylog = log.New("module", "execs.js")
)

var driverName = ptypes.JsX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&js{}))
}

//Init 插件初始化
func Init(name string, sub []byte) {
	drivers.Register(GetName(), newjs, 0)
}

type js struct {
	drivers.DriverBase
	prefix      []byte
	localprefix []byte
}

func newjs() drivers.Driver {
	t := &js{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

//GetName 获取名字
func GetName() string {
	return newjs().GetName()
}

//GetDriverName 获取插件的名字
func (u *js) GetDriverName() string {
	return driverName
}

func (u *js) callVM(prefix string, payload *jsproto.Call, tx *types.Transaction,
	index int, receiptData *types.ReceiptData) (*otto.Object, error) {
	vm, err := u.createVM(payload.Name, tx, index)
	if err != nil {
		return nil, err
	}
	db := u.GetStateDB()
	code, err := db.Get(calcCodeKey(payload.Name))
	if err != nil {
		return nil, err
	}
	loglist, err := jslogs(receiptData)
	if err != nil {
		return nil, err
	}
	vm.Set("loglist", loglist)
	vm.Set("code", code)
	if prefix == "init" {
		vm.Set("f", "init")
	} else {
		vm.Set("f", prefix+"_"+payload.Funcname)
	}
	vm.Set("args", payload.Args)
	callfunc := "callcode(context, f, args, loglist)"
	jsvalue, err := vm.Run(callcode + string(code) + "\n" + callfunc)
	if err != nil {
		return nil, err
	}
	if prefix == "query" {
		s, err := jsvalue.ToString()
		if err != nil {
			return nil, err
		}
		return newObject(vm).setValue("result", s).object(), nil
	}
	if !jsvalue.IsObject() {
		return nil, ptypes.ErrJsReturnNotObject
	}
	return jsvalue.Object(), nil
}

func jslogs(receiptData *types.ReceiptData) ([]string, error) {
	data := make([]string, 0)
	if receiptData == nil {
		return data, nil
	}
	for i := 0; i < len(receiptData.Logs); i++ {
		logitem := receiptData.Logs[i]
		if logitem.Ty != ptypes.TyLogJs {
			continue
		}
		var jslog jsproto.JsLog
		err := types.Decode(logitem.Log, &jslog)
		if err != nil {
			return nil, err
		}
		data = append(data, jslog.Data)
	}
	return data, nil
}

func (u *js) getContext(tx *types.Transaction, index int64) *blockContext {
	var hash [32]byte
	if tx != nil {
		copy(hash[:], tx.Hash())
	}
	return &blockContext{
		Height:     u.GetHeight(),
		Name:       u.GetName(),
		Blocktime:  u.GetBlockTime(),
		Curname:    u.GetCurrentExecName(),
		DriverName: u.GetDriverName(),
		Difficulty: u.GetDifficulty(),
		TxHash:     common.ToHex(hash[:]),
		Index:      index,
	}
}

func (u *js) getstatedbFunc(vm *otto.Otto, name string) {
	prefix, _ := calcAllPrefix(name)
	vm.Set("getstatedb", func(call otto.FunctionCall) otto.Value {
		key, err := call.Argument(0).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		v, err := u.getstatedb(string(prefix) + key)
		if err != nil {
			return errReturn(vm, err)
		}
		return okReturn(vm, v)
	})
}

func (u *js) getlocaldbFunc(vm *otto.Otto, name string) {
	_, prefix := calcAllPrefix(name)
	vm.Set("getlocaldb", func(call otto.FunctionCall) otto.Value {
		key, err := call.Argument(0).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		v, err := u.getlocaldb(string(prefix) + key)
		if err != nil {
			return errReturn(vm, err)
		}
		return okReturn(vm, v)
	})
}

func (u *js) listdbFunc(vm *otto.Otto, name string) {
	//List(prefix, key []byte, count, direction int32) ([][]byte, error)
	_, plocal := calcAllPrefix(name)
	vm.Set("listdb", func(call otto.FunctionCall) otto.Value {
		prefix, err := call.Argument(0).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		key, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		count, err := call.Argument(2).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		direction, err := call.Argument(3).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		v, err := u.listdb(string(plocal)+prefix, key, int32(count), int32(direction))
		if err != nil {
			return errReturn(vm, err)
		}
		return listReturn(vm, v)
	})
}

func (u *js) createVM(name string, tx *types.Transaction, index int) (*otto.Otto, error) {
	data, err := json.Marshal(u.getContext(tx, int64(index)))
	if err != nil {
		return nil, err
	}
	vm := otto.New()
	vm.Set("context", string(data))
	u.getstatedbFunc(vm, name)
	u.getlocaldbFunc(vm, name)
	u.listdbFunc(vm, name)
	return vm, nil
}

func errReturn(vm *otto.Otto, err error) otto.Value {
	return newObject(vm).setErr(err).value()
}

func okReturn(vm *otto.Otto, value string) otto.Value {
	return newObject(vm).setValue("value", value).value()
}

func listReturn(vm *otto.Otto, value []string) otto.Value {
	return newObject(vm).setValue("value", value).value()
}

type object struct {
	vm  *otto.Otto
	obj *otto.Object
}

func newObject(vm *otto.Otto) *object {
	obj, err := vm.Object("({})")
	if err != nil {
		panic(err)
	}
	return &object{vm: vm, obj: obj}
}

func (o *object) setErr(err error) *object {
	if err != nil {
		o.obj.Set("err", err.Error())
	}
	return o
}

func (o *object) setValue(key string, value interface{}) *object {
	o.obj.Set(key, value)
	return o
}

func (o *object) object() *otto.Object {
	return o.obj
}

func (o *object) value() otto.Value {
	v, err := otto.ToValue(o.obj)
	if err != nil {
		panic(err)
	}
	return v
}

func (u *js) getstatedb(key string) (value string, err error) {
	s, err := u.GetStateDB().Get([]byte(key))
	value = string(s)
	return value, err
}

func (u *js) getlocaldb(key string) (value string, err error) {
	s, err := u.GetLocalDB().Get([]byte(key))
	value = string(s)
	return value, err
}

func (u *js) listdb(prefix, key string, count, direction int32) (value []string, err error) {
	values, err := u.GetLocalDB().List([]byte(prefix), []byte(key), count, direction)
	for _, v := range values {
		value = append(value, string(v))
	}
	return value, err
}
