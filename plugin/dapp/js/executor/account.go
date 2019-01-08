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
get_balance(obj, execer, addr) Account

transfer(obj, from, to, amount) Receipt
transfer_to_exec(obj, execer, addr, amount) Receipt
withdraw(obj, execer, addr, amount) Receipt

exec_frozen(obj, execer, addr, amount) Receipt
exec_active(obj, execer, addr, amount) Receipt
exec_deposit(obj, execer, addr, amount) Receipt
exec_withdraw(obj, execer, addr, amount) Receipt
exec_transfer(obj, execer, from, to, amount) Receipt
*/
func (u *js) registerAccountFunc(vm *otto.Otto) {
	u.genesisInitExecFunc(vm)
	u.genesisInitFunc(vm)
	u.loadAccountFunc(vm)
	u.getBalanceFunc(vm)
	u.transferFunc(vm)
	u.transferToExecFunc(vm)
	u.withdrawFunc(vm)
	u.execActiveFunc(vm)
	u.execDepositFunc(vm)
	u.execFrozenFunc(vm)
	u.execTransferFunc(vm)
	u.execWithdrawFunc(vm)
}

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

func (u *js) loadAccountFunc(vm *otto.Otto) {
	vm.Set("load_account", func(call otto.FunctionCall) otto.Value {
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
		account := acc.LoadAccount(addr)
		return accountReturn(vm, account)
	})
}

func (u *js) getBalanceFunc(vm *otto.Otto) {
	vm.Set("get_balance", func(call otto.FunctionCall) otto.Value {
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
		account, err := acc.GetBalance(u.GetAPI(), &types.ReqBalance{
			Addresses: []string{addr},
			Execer:    execer,
		})
		if err != nil {
			return errReturn(vm, err)
		}
		if len(account) == 0 {
			return accountReturn(vm, &types.Account{})
		}
		return accountReturn(vm, account[0])
	})
}

func (u *js) transferFunc(vm *otto.Otto) {
	vm.Set("transfer", func(call otto.FunctionCall) otto.Value {
		acc, err := u.getAccount(call.Argument(0))
		if err != nil {
			return errReturn(vm, err)
		}
		from, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		if err := address.CheckAddress(from); err != nil {
			return errReturn(vm, err)
		}
		to, err := call.Argument(2).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		if err := address.CheckAddress(to); err != nil {
			return errReturn(vm, err)
		}
		amount, err := call.Argument(3).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		receipt, err := acc.Transfer(from, to, amount)
		if err != nil {
			return errReturn(vm, err)
		}
		return receiptReturn(vm, receipt)
	})
}

func (u *js) transferToExecFunc(vm *otto.Otto) {
	vm.Set("transfer_to_exec", func(call otto.FunctionCall) otto.Value {
		acc, err := u.getAccount(call.Argument(0))
		if err != nil {
			return errReturn(vm, err)
		}
		execer, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		from, err := call.Argument(2).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		if err := address.CheckAddress(from); err != nil {
			return errReturn(vm, err)
		}
		amount, err := call.Argument(3).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		receipt, err := acc.TransferToExec(from, address.ExecAddress(execer), amount)
		if err != nil {
			return errReturn(vm, err)
		}
		return receiptReturn(vm, receipt)
	})
}

func (u *js) withdrawFunc(vm *otto.Otto) {
	vm.Set("withdraw", func(call otto.FunctionCall) otto.Value {
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
		receipt, err := acc.TransferWithdraw(address.ExecAddress(execer), addr, amount)
		if err != nil {
			return errReturn(vm, err)
		}
		return receiptReturn(vm, receipt)
	})
}

func (u *js) execFrozenFunc(vm *otto.Otto) {
	vm.Set("exec_frozen", func(call otto.FunctionCall) otto.Value {
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
		receipt, err := acc.ExecFrozen(addr, address.ExecAddress(execer), amount)
		if err != nil {
			return errReturn(vm, err)
		}
		return receiptReturn(vm, receipt)
	})
}

func (u *js) execActiveFunc(vm *otto.Otto) {
	vm.Set("exec_active", func(call otto.FunctionCall) otto.Value {
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
		receipt, err := acc.ExecActive(addr, address.ExecAddress(execer), amount)
		if err != nil {
			return errReturn(vm, err)
		}
		return receiptReturn(vm, receipt)
	})
}

func (u *js) execDepositFunc(vm *otto.Otto) {
	vm.Set("exec_deposit", func(call otto.FunctionCall) otto.Value {
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
		receipt, err := acc.ExecDeposit(addr, address.ExecAddress(execer), amount)
		if err != nil {
			return errReturn(vm, err)
		}
		return receiptReturn(vm, receipt)
	})
}

func (u *js) execWithdrawFunc(vm *otto.Otto) {
	vm.Set("exec_withdraw", func(call otto.FunctionCall) otto.Value {
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
		receipt, err := acc.ExecWithdraw(address.ExecAddress(execer), addr, amount)
		if err != nil {
			return errReturn(vm, err)
		}
		return receiptReturn(vm, receipt)
	})
}

func (u *js) execTransferFunc(vm *otto.Otto) {
	vm.Set("exec_transfer", func(call otto.FunctionCall) otto.Value {
		acc, err := u.getAccount(call.Argument(0))
		if err != nil {
			return errReturn(vm, err)
		}
		execer, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		from, err := call.Argument(2).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		if err := address.CheckAddress(from); err != nil {
			return errReturn(vm, err)
		}
		to, err := call.Argument(3).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		if err := address.CheckAddress(to); err != nil {
			return errReturn(vm, err)
		}
		amount, err := call.Argument(4).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		receipt, err := acc.ExecTransfer(from, to, address.ExecAddress(execer), amount)
		if err != nil {
			return errReturn(vm, err)
		}
		return receiptReturn(vm, receipt)
	})
}
