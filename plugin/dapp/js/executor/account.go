package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/robertkrimen/otto"
)

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
genesis_init_exec(obj, execer, addr string, amount int64)

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
*/
func (u *js) getAccount(args otto.Value) (*account.DB, error) {
	if !args.IsObject() {
		return nil, types.ErrInvalidParam
	}
	obj := args.Object()
	execer, err := getString(obj, "execer")
	if err != nil {
		return nil, err
	}
	symbol, err := getString(obj, "symbol")
	if err != nil {
		return nil, err
	}
	return account.NewAccountDB(execer, symbol, u.GetStateDB())
}

func (u *js) genesisInitExecFunc(vm *otto.Otto) {
	vm.Set("genesis_init_exec", func(call otto.FunctionCall) otto.Value {
		acc, err := u.getAccount(call.Argument(0))
		if err != nil {
			return errReturn(vm, err)
		}
		execer, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		addr, err := call.Argument(2).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		if err := address.CheckAddress(addr); err != nil {
			return errReturn(vm, err)
		}
		amount, err := call.Argument(3).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		receipt, err := acc.GenesisInitExec(addr, amount, address.ExecAddress(execer))
		if err != nil {
			return errReturn(vm, err)
		}
		return receiptReturn(vm, receipt)
	})
}

func (u *js) genesisInitFunc(vm *otto.Otto) {
	vm.Set("genesis_init", func(call otto.FunctionCall) otto.Value {
		acc, err := u.getAccount(call.Argument(0))
		if err != nil {
			return errReturn(vm, err)
		}
		addr, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		if err := address.CheckAddress(addr); err != nil {
			return errReturn(vm, err)
		}
		amount, err := call.Argument(2).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		receipt, err := acc.GenesisInit(addr, amount)
		if err != nil {
			return errReturn(vm, err)
		}
		return receiptReturn(vm, receipt)
	})
}
