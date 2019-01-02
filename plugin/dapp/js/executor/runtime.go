package executor

import (
	"github.com/33cn/chain33/common/address"
	"github.com/robertkrimen/otto"
)

//让 js 具有访问 区块链 的一些能力
func execaddressFunc(vm *otto.Otto) {
	vm.Set("execaddress", func(call otto.FunctionCall) otto.Value {
		key, err := call.Argument(0).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		addr := address.ExecAddress(key)
		return okReturn(vm, addr)
	})
}

/*
//table
//要开发一个适合json的table, row 就是一个 js object
//handle := new_table(config)
//table_add(handle, row)
//table_replace(handle, row)
//table_del(handle, row)
//table_savekvs(handle)
//table_close(handle)

//join table 的操作(接口完全相同)
//handle3 := new_table(newcofifg{config1, config2})

//获取系统随机数的接口
//randnum

//获取前一个区块hash的接口
//prev_blockhash()
*/
