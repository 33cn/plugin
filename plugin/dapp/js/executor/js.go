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

func (u *js) callVM(prefix string, payload *jsproto.Call, tx *types.Transaction, index int) (*otto.Object, error) {
	vm, err := u.createVM(tx, index)
	if err != nil {
		return nil, err
	}
	db := u.GetStateDB()
	code, err := db.Get(calcCodeKey(payload.Name))
	if err != nil {
		return nil, err
	}
	vm.Set("code", code)
	vm.Set("f", prefix+"_"+payload.Funcname)
	vm.Set("args", payload.Args)
	callfunc := "callcode(context, f, args)"
	jsvalue, err := vm.Run(callcode + string(code) + "\n" + callfunc)
	if err != nil {
		return nil, err
	}
	if !jsvalue.IsObject() {
		return nil, ptypes.ErrJsReturnNotObject
	}
	return jsvalue.Object(), nil
}

func (u *js) getContext(tx *types.Transaction, index int64) *blockContext {
	return &blockContext{
		Height:     u.GetHeight(),
		Name:       u.GetName(),
		Blocktime:  u.GetBlockTime(),
		Curname:    u.GetCurrentExecName(),
		DriverName: u.GetDriverName(),
		Difficulty: u.GetDifficulty(),
		TxHash:     common.ToHex(tx.Hash()),
		Index:      index,
	}
}

func (u *js) createVM(tx *types.Transaction, index int) (*otto.Otto, error) {
	data, err := json.Marshal(u.getContext(tx, int64(index)))
	if err != nil {
		return nil, err
	}
	vm := otto.New()
	vm.Set("context", string(data))
	vm.Set("getstatedb", func(call otto.FunctionCall) otto.Value {
		key, err := call.Argument(0).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		v, err := u.getstatedb(key)
		if err != nil {
			return errReturn(vm, err)
		}
		return okReturn(vm, v)
	})
	vm.Set("getlocaldb", func(call otto.FunctionCall) otto.Value {
		key, err := call.Argument(0).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		v, err := u.getlocaldb(key)
		if err != nil {
			return errReturn(vm, err)
		}
		return okReturn(vm, v)
	})
	//List(prefix, key []byte, count, direction int32) ([][]byte, error)
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
		v, err := u.listdb(prefix, key, int32(count), int32(direction))
		if err != nil {
			return errReturn(vm, err)
		}
		return listReturn(vm, v)
	})
	return vm, nil
}

func errReturn(vm *otto.Otto, err error) otto.Value {
	v, _ := vm.ToValue(&dbReturn{Err: err.Error()})
	return v
}

func okReturn(vm *otto.Otto, value string) otto.Value {
	v, _ := vm.ToValue(&dbReturn{Value: value})
	return v
}

func listReturn(vm *otto.Otto, value []string) otto.Value {
	v, _ := vm.ToValue(&listdbReturn{Value: value})
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
