package runtime

import (
	"errors"
	"fmt"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/params"
	"math/big"
)

// StatefulPrecompiledContract is the interface for executing a precompiled contract
// This wraps the PrecompiledContracts native to Ethereum and allows adding in stateful
// precompiled contracts to support native chain33 asset transfers.
type StatefulPrecompiledContract interface {
	RequiredGas(input []byte) uint64
	Run(evm *EVM, caller ContractRef, input []byte) (ret []byte, err error)
}

//因为common.Address结构体中定义的是指针，map中的key值不能使用address作为key值来使用，于是使用Hash160Address作为key来进行索引
var CustomizePrecompiledContractsBinjiang = map[common.Hash160Address]StatefulPrecompiledContract{
	//0x01-0x08 是evm 系统自带的预编译合约接口，自定义预编译合约接口从0x1000 开始
	common.BytesToHash160Address([]byte{129}): &evm2Exchange{},
}

type evm2Exchange struct {
}

func (e *evm2Exchange) RequiredGas(input []byte) uint64 {
	return uint64(len(input)+31)/32*params.IdentityPerWordGas + params.IdentityBaseGas
}

func (e *evm2Exchange) Run(evm *EVM, caller ContractRef, input []byte) ([]byte, error) {

	// step 1 对input 数据进行解析 transferOf
	if len(input) < 20 {
		return nil, errors.New("input size to low")
	}
	outParam, err := abi.Unpack(input[20:], "transfer", innerErc20AbiData)
	if err != nil {
		return nil, err
	}

	assertSymbol := string(input[:20])
	//input: assertContractaddress(20 bytes)|packdata
	//evm合约地址，把币转移到某个地址下托管
	recipient := outParam[0].Value.(common.Address)
	amount, _ := outParam[1].Value.(*big.Int)
	// step2 把evm的币转移到一个托管地址下
	// 在此处打印下自定义合约的错误信息
	if !evm.StateDB.Exist(recipient.String()) {
		return nil, model.ErrContractNotExist
	}
	contract := NewContract(caller, AccountRef(recipient), amount.Uint64(), e.RequiredGas(input[20:]))
	contract.SetCallCode(&recipient, evm.StateDB.GetCodeHash(recipient.String()), evm.StateDB.GetCode(recipient.String()))
	//lock evm assert
	ret, err := evm.Interpreter.Run(contract, input[20:], false) //lock/tuoguan
	if err != nil {
		return ret, err
	}

	// step3  给exchange account 账户 增加对应的余额
	receipt, err := evm.evmexecutor.GetMStateDB().TransferToExchange(recipient.String(), assertSymbol, amount.Int64())
	//(evm, caller.Address().String(), assertSymbol, amount.Int64())
	if err != nil {
		return ret, err
	}
	fmt.Println("receipt", receipt)
	//sep4 transfertoexec exchange
	return ret, nil

}

/*
func (e *evm2Exchange) xgoTransfer(evm *EVM, recipient, symbol string, amount int64) (*types.Receipt, error) { //([]*types.KeyValue, []*types.ReceiptLog, error) {
	evmxgoAccount, err := account.NewAccountDB(evm.cfg, "evmxgo", symbol, evm.evmexecutor.GetMStateDB().StateDB)
	if err != nil {
		return nil, err
	}

	execName := evm.cfg.ExecName("exchange")
	execaddress := address.ExecAddress(execName)
	//导出账户地址
	acc, err := evmxgoAccount.LoadExecAccountQueue(evm.evmexecutor.GetAPI(), recipient, execaddress)
	if err != nil {
		return nil, err
	}

	newbalance, err := safeAdd(acc.Balance, amount)
	if err != nil {
		return nil, err
	}
	copyAcc := types.CloneAccount(acc)
	acc.Balance = newbalance
	receipt := &types.ReceiptAccountMint{
		Prev:    copyAcc,
		Current: acc,
	}
	kvset := evmxgoAccount.GetKVSet(acc)
	evmxgoAccount.SaveKVSet(kvset)
	ty := int32(types.TyLogExecTransfer)
	log1 := &types.ReceiptLog{
		Ty:  ty,
		Log: types.Encode(receipt),
	}

	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kvset,
		Logs: []*types.ReceiptLog{log1},
	}, nil

}

func safeAdd(balance, amount int64) (int64, error) {
	if balance+amount < amount || balance+amount > types.MaxTokenBalance {
		return balance, types.ErrAmount
	}
	return balance + amount, nil
}*/

const innerErc20AbiData = `[
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "recipient",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "transfer",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	
]`
