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
//chain33 相关的账户操作函数 （操作某个execer 下面的 symbol）
function Account(execer, symbol) {
	this.execer = execer
	this.symbol = symbol
}

func Receipt(kvs, logs) {
	this.kvs = kvs
	this.logs = logs
}

var obj = new Account(execer, symbol)
//init 函数才能使用的两个函数(或者增发)
genesis_init(obj, addr string, amount int64)
genesis_exec_init(obj, execer, addr string, amount int64)

load_account(obj, addr) Account
get_balance(obj, addr, execer) Account
transfer(obj, from, to, amount) Receipt
transfer_to_exec(obj, execer, addr, amount) Receipt
withdraw(obj, execer, addr, amount) Receipt
exec_frozen(obj, addr) Receipt
exec_active(obj, addr) Receipt
exec_transfer(obj, from, to, amount) Receipt
exec_deposit(obj, addr, amount) Receipt
exec_withdraw(obj, addr, amount) Receipt

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
