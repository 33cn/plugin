// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package sousChef

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// AddressABI is the input ABI used to generate the binding from.
const AddressABI = "[]"

// AddressBin is the compiled bytecode used for deploying new contracts.
var AddressBin = "0x60566023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea26469706673582212202b00d1886e0a10edb0744d9ee92b7e2394da080a59f42c16070c3484ccad4e2564736f6c634300060c0033"

// DeployAddress deploys a new Ethereum contract, binding an instance of Address to it.
func DeployAddress(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Address, error) {
	parsed, err := abi.JSON(strings.NewReader(AddressABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(AddressBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Address{AddressCaller: AddressCaller{contract: contract}, AddressTransactor: AddressTransactor{contract: contract}, AddressFilterer: AddressFilterer{contract: contract}}, nil
}

// Address is an auto generated Go binding around an Ethereum contract.
type Address struct {
	AddressCaller     // Read-only binding to the contract
	AddressTransactor // Write-only binding to the contract
	AddressFilterer   // Log filterer for contract events
}

// AddressCaller is an auto generated read-only Go binding around an Ethereum contract.
type AddressCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddressTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AddressTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddressFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AddressFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddressSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AddressSession struct {
	Contract     *Address          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AddressCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AddressCallerSession struct {
	Contract *AddressCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// AddressTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AddressTransactorSession struct {
	Contract     *AddressTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// AddressRaw is an auto generated low-level Go binding around an Ethereum contract.
type AddressRaw struct {
	Contract *Address // Generic contract binding to access the raw methods on
}

// AddressCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AddressCallerRaw struct {
	Contract *AddressCaller // Generic read-only contract binding to access the raw methods on
}

// AddressTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AddressTransactorRaw struct {
	Contract *AddressTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAddress creates a new instance of Address, bound to a specific deployed contract.
func NewAddress(address common.Address, backend bind.ContractBackend) (*Address, error) {
	contract, err := bindAddress(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Address{AddressCaller: AddressCaller{contract: contract}, AddressTransactor: AddressTransactor{contract: contract}, AddressFilterer: AddressFilterer{contract: contract}}, nil
}

// NewAddressCaller creates a new read-only instance of Address, bound to a specific deployed contract.
func NewAddressCaller(address common.Address, caller bind.ContractCaller) (*AddressCaller, error) {
	contract, err := bindAddress(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AddressCaller{contract: contract}, nil
}

// NewAddressTransactor creates a new write-only instance of Address, bound to a specific deployed contract.
func NewAddressTransactor(address common.Address, transactor bind.ContractTransactor) (*AddressTransactor, error) {
	contract, err := bindAddress(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AddressTransactor{contract: contract}, nil
}

// NewAddressFilterer creates a new log filterer instance of Address, bound to a specific deployed contract.
func NewAddressFilterer(address common.Address, filterer bind.ContractFilterer) (*AddressFilterer, error) {
	contract, err := bindAddress(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AddressFilterer{contract: contract}, nil
}

// bindAddress binds a generic wrapper to an already deployed contract.
func bindAddress(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AddressABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Address *AddressRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Address.Contract.AddressCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Address *AddressRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Address.Contract.AddressTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Address *AddressRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Address.Contract.AddressTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Address *AddressCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Address.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Address *AddressTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Address.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Address *AddressTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Address.Contract.contract.Transact(opts, method, params...)
}

// IBEP20ABI is the input ABI used to generate the binding from.
const IBEP20ABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// IBEP20FuncSigs maps the 4-byte function signature to its string representation.
var IBEP20FuncSigs = map[string]string{
	"dd62ed3e": "allowance(address,address)",
	"095ea7b3": "approve(address,uint256)",
	"70a08231": "balanceOf(address)",
	"313ce567": "decimals()",
	"893d20e8": "getOwner()",
	"06fdde03": "name()",
	"95d89b41": "symbol()",
	"18160ddd": "totalSupply()",
	"a9059cbb": "transfer(address,uint256)",
	"23b872dd": "transferFrom(address,address,uint256)",
}

// IBEP20 is an auto generated Go binding around an Ethereum contract.
type IBEP20 struct {
	IBEP20Caller     // Read-only binding to the contract
	IBEP20Transactor // Write-only binding to the contract
	IBEP20Filterer   // Log filterer for contract events
}

// IBEP20Caller is an auto generated read-only Go binding around an Ethereum contract.
type IBEP20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBEP20Transactor is an auto generated write-only Go binding around an Ethereum contract.
type IBEP20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBEP20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IBEP20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBEP20Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IBEP20Session struct {
	Contract     *IBEP20           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IBEP20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IBEP20CallerSession struct {
	Contract *IBEP20Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// IBEP20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IBEP20TransactorSession struct {
	Contract     *IBEP20Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IBEP20Raw is an auto generated low-level Go binding around an Ethereum contract.
type IBEP20Raw struct {
	Contract *IBEP20 // Generic contract binding to access the raw methods on
}

// IBEP20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IBEP20CallerRaw struct {
	Contract *IBEP20Caller // Generic read-only contract binding to access the raw methods on
}

// IBEP20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IBEP20TransactorRaw struct {
	Contract *IBEP20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewIBEP20 creates a new instance of IBEP20, bound to a specific deployed contract.
func NewIBEP20(address common.Address, backend bind.ContractBackend) (*IBEP20, error) {
	contract, err := bindIBEP20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IBEP20{IBEP20Caller: IBEP20Caller{contract: contract}, IBEP20Transactor: IBEP20Transactor{contract: contract}, IBEP20Filterer: IBEP20Filterer{contract: contract}}, nil
}

// NewIBEP20Caller creates a new read-only instance of IBEP20, bound to a specific deployed contract.
func NewIBEP20Caller(address common.Address, caller bind.ContractCaller) (*IBEP20Caller, error) {
	contract, err := bindIBEP20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IBEP20Caller{contract: contract}, nil
}

// NewIBEP20Transactor creates a new write-only instance of IBEP20, bound to a specific deployed contract.
func NewIBEP20Transactor(address common.Address, transactor bind.ContractTransactor) (*IBEP20Transactor, error) {
	contract, err := bindIBEP20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IBEP20Transactor{contract: contract}, nil
}

// NewIBEP20Filterer creates a new log filterer instance of IBEP20, bound to a specific deployed contract.
func NewIBEP20Filterer(address common.Address, filterer bind.ContractFilterer) (*IBEP20Filterer, error) {
	contract, err := bindIBEP20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IBEP20Filterer{contract: contract}, nil
}

// bindIBEP20 binds a generic wrapper to an already deployed contract.
func bindIBEP20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IBEP20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IBEP20 *IBEP20Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IBEP20.Contract.IBEP20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IBEP20 *IBEP20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBEP20.Contract.IBEP20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IBEP20 *IBEP20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IBEP20.Contract.IBEP20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IBEP20 *IBEP20CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IBEP20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IBEP20 *IBEP20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBEP20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IBEP20 *IBEP20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IBEP20.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address spender) view returns(uint256)
func (_IBEP20 *IBEP20Caller) Allowance(opts *bind.CallOpts, _owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IBEP20.contract.Call(opts, &out, "allowance", _owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address spender) view returns(uint256)
func (_IBEP20 *IBEP20Session) Allowance(_owner common.Address, spender common.Address) (*big.Int, error) {
	return _IBEP20.Contract.Allowance(&_IBEP20.CallOpts, _owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address spender) view returns(uint256)
func (_IBEP20 *IBEP20CallerSession) Allowance(_owner common.Address, spender common.Address) (*big.Int, error) {
	return _IBEP20.Contract.Allowance(&_IBEP20.CallOpts, _owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_IBEP20 *IBEP20Caller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IBEP20.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_IBEP20 *IBEP20Session) BalanceOf(account common.Address) (*big.Int, error) {
	return _IBEP20.Contract.BalanceOf(&_IBEP20.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_IBEP20 *IBEP20CallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _IBEP20.Contract.BalanceOf(&_IBEP20.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_IBEP20 *IBEP20Caller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _IBEP20.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_IBEP20 *IBEP20Session) Decimals() (uint8, error) {
	return _IBEP20.Contract.Decimals(&_IBEP20.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_IBEP20 *IBEP20CallerSession) Decimals() (uint8, error) {
	return _IBEP20.Contract.Decimals(&_IBEP20.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_IBEP20 *IBEP20Caller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _IBEP20.contract.Call(opts, &out, "getOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_IBEP20 *IBEP20Session) GetOwner() (common.Address, error) {
	return _IBEP20.Contract.GetOwner(&_IBEP20.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_IBEP20 *IBEP20CallerSession) GetOwner() (common.Address, error) {
	return _IBEP20.Contract.GetOwner(&_IBEP20.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_IBEP20 *IBEP20Caller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _IBEP20.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_IBEP20 *IBEP20Session) Name() (string, error) {
	return _IBEP20.Contract.Name(&_IBEP20.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_IBEP20 *IBEP20CallerSession) Name() (string, error) {
	return _IBEP20.Contract.Name(&_IBEP20.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_IBEP20 *IBEP20Caller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _IBEP20.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_IBEP20 *IBEP20Session) Symbol() (string, error) {
	return _IBEP20.Contract.Symbol(&_IBEP20.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_IBEP20 *IBEP20CallerSession) Symbol() (string, error) {
	return _IBEP20.Contract.Symbol(&_IBEP20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_IBEP20 *IBEP20Caller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IBEP20.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_IBEP20 *IBEP20Session) TotalSupply() (*big.Int, error) {
	return _IBEP20.Contract.TotalSupply(&_IBEP20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_IBEP20 *IBEP20CallerSession) TotalSupply() (*big.Int, error) {
	return _IBEP20.Contract.TotalSupply(&_IBEP20.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_IBEP20 *IBEP20Transactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IBEP20.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_IBEP20 *IBEP20Session) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IBEP20.Contract.Approve(&_IBEP20.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_IBEP20 *IBEP20TransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IBEP20.Contract.Approve(&_IBEP20.TransactOpts, spender, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_IBEP20 *IBEP20Transactor) Transfer(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IBEP20.contract.Transact(opts, "transfer", recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_IBEP20 *IBEP20Session) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IBEP20.Contract.Transfer(&_IBEP20.TransactOpts, recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_IBEP20 *IBEP20TransactorSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IBEP20.Contract.Transfer(&_IBEP20.TransactOpts, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_IBEP20 *IBEP20Transactor) TransferFrom(opts *bind.TransactOpts, sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IBEP20.contract.Transact(opts, "transferFrom", sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_IBEP20 *IBEP20Session) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IBEP20.Contract.TransferFrom(&_IBEP20.TransactOpts, sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_IBEP20 *IBEP20TransactorSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IBEP20.Contract.TransferFrom(&_IBEP20.TransactOpts, sender, recipient, amount)
}

// IBEP20ApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the IBEP20 contract.
type IBEP20ApprovalIterator struct {
	Event *IBEP20Approval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IBEP20ApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IBEP20Approval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IBEP20Approval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IBEP20ApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IBEP20ApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IBEP20Approval represents a Approval event raised by the IBEP20 contract.
type IBEP20Approval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_IBEP20 *IBEP20Filterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*IBEP20ApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _IBEP20.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &IBEP20ApprovalIterator{contract: _IBEP20.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_IBEP20 *IBEP20Filterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *IBEP20Approval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _IBEP20.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IBEP20Approval)
				if err := _IBEP20.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_IBEP20 *IBEP20Filterer) ParseApproval(log types.Log) (*IBEP20Approval, error) {
	event := new(IBEP20Approval)
	if err := _IBEP20.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IBEP20TransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the IBEP20 contract.
type IBEP20TransferIterator struct {
	Event *IBEP20Transfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IBEP20TransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IBEP20Transfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IBEP20Transfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IBEP20TransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IBEP20TransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IBEP20Transfer represents a Transfer event raised by the IBEP20 contract.
type IBEP20Transfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_IBEP20 *IBEP20Filterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*IBEP20TransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IBEP20.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &IBEP20TransferIterator{contract: _IBEP20.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_IBEP20 *IBEP20Filterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *IBEP20Transfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IBEP20.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IBEP20Transfer)
				if err := _IBEP20.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_IBEP20 *IBEP20Filterer) ParseTransfer(log types.Log) (*IBEP20Transfer, error) {
	event := new(IBEP20Transfer)
	if err := _IBEP20.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeBEP20ABI is the input ABI used to generate the binding from.
const SafeBEP20ABI = "[]"

// SafeBEP20Bin is the compiled bytecode used for deploying new contracts.
var SafeBEP20Bin = "0x60566023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea26469706673582212201fc1adb7565c79eab91107fb316e27a66f5b6b82d647fce911b13d702f685d8564736f6c634300060c0033"

// DeploySafeBEP20 deploys a new Ethereum contract, binding an instance of SafeBEP20 to it.
func DeploySafeBEP20(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SafeBEP20, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeBEP20ABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SafeBEP20Bin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SafeBEP20{SafeBEP20Caller: SafeBEP20Caller{contract: contract}, SafeBEP20Transactor: SafeBEP20Transactor{contract: contract}, SafeBEP20Filterer: SafeBEP20Filterer{contract: contract}}, nil
}

// SafeBEP20 is an auto generated Go binding around an Ethereum contract.
type SafeBEP20 struct {
	SafeBEP20Caller     // Read-only binding to the contract
	SafeBEP20Transactor // Write-only binding to the contract
	SafeBEP20Filterer   // Log filterer for contract events
}

// SafeBEP20Caller is an auto generated read-only Go binding around an Ethereum contract.
type SafeBEP20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeBEP20Transactor is an auto generated write-only Go binding around an Ethereum contract.
type SafeBEP20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeBEP20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SafeBEP20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeBEP20Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SafeBEP20Session struct {
	Contract     *SafeBEP20        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SafeBEP20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SafeBEP20CallerSession struct {
	Contract *SafeBEP20Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// SafeBEP20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SafeBEP20TransactorSession struct {
	Contract     *SafeBEP20Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// SafeBEP20Raw is an auto generated low-level Go binding around an Ethereum contract.
type SafeBEP20Raw struct {
	Contract *SafeBEP20 // Generic contract binding to access the raw methods on
}

// SafeBEP20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SafeBEP20CallerRaw struct {
	Contract *SafeBEP20Caller // Generic read-only contract binding to access the raw methods on
}

// SafeBEP20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SafeBEP20TransactorRaw struct {
	Contract *SafeBEP20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewSafeBEP20 creates a new instance of SafeBEP20, bound to a specific deployed contract.
func NewSafeBEP20(address common.Address, backend bind.ContractBackend) (*SafeBEP20, error) {
	contract, err := bindSafeBEP20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SafeBEP20{SafeBEP20Caller: SafeBEP20Caller{contract: contract}, SafeBEP20Transactor: SafeBEP20Transactor{contract: contract}, SafeBEP20Filterer: SafeBEP20Filterer{contract: contract}}, nil
}

// NewSafeBEP20Caller creates a new read-only instance of SafeBEP20, bound to a specific deployed contract.
func NewSafeBEP20Caller(address common.Address, caller bind.ContractCaller) (*SafeBEP20Caller, error) {
	contract, err := bindSafeBEP20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SafeBEP20Caller{contract: contract}, nil
}

// NewSafeBEP20Transactor creates a new write-only instance of SafeBEP20, bound to a specific deployed contract.
func NewSafeBEP20Transactor(address common.Address, transactor bind.ContractTransactor) (*SafeBEP20Transactor, error) {
	contract, err := bindSafeBEP20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SafeBEP20Transactor{contract: contract}, nil
}

// NewSafeBEP20Filterer creates a new log filterer instance of SafeBEP20, bound to a specific deployed contract.
func NewSafeBEP20Filterer(address common.Address, filterer bind.ContractFilterer) (*SafeBEP20Filterer, error) {
	contract, err := bindSafeBEP20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SafeBEP20Filterer{contract: contract}, nil
}

// bindSafeBEP20 binds a generic wrapper to an already deployed contract.
func bindSafeBEP20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeBEP20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SafeBEP20 *SafeBEP20Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SafeBEP20.Contract.SafeBEP20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SafeBEP20 *SafeBEP20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SafeBEP20.Contract.SafeBEP20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SafeBEP20 *SafeBEP20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SafeBEP20.Contract.SafeBEP20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SafeBEP20 *SafeBEP20CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SafeBEP20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SafeBEP20 *SafeBEP20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SafeBEP20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SafeBEP20 *SafeBEP20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SafeBEP20.Contract.contract.Transact(opts, method, params...)
}

// SafeMathABI is the input ABI used to generate the binding from.
const SafeMathABI = "[]"

// SafeMathBin is the compiled bytecode used for deploying new contracts.
var SafeMathBin = "0x60566023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea26469706673582212204a516442da9f9cb47c0961cb6feee9120401d44e69866291508bed825ec56d9e64736f6c634300060c0033"

// DeploySafeMath deploys a new Ethereum contract, binding an instance of SafeMath to it.
func DeploySafeMath(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SafeMath, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeMathABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SafeMathBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SafeMath{SafeMathCaller: SafeMathCaller{contract: contract}, SafeMathTransactor: SafeMathTransactor{contract: contract}, SafeMathFilterer: SafeMathFilterer{contract: contract}}, nil
}

// SafeMath is an auto generated Go binding around an Ethereum contract.
type SafeMath struct {
	SafeMathCaller     // Read-only binding to the contract
	SafeMathTransactor // Write-only binding to the contract
	SafeMathFilterer   // Log filterer for contract events
}

// SafeMathCaller is an auto generated read-only Go binding around an Ethereum contract.
type SafeMathCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SafeMathTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SafeMathFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SafeMathSession struct {
	Contract     *SafeMath         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SafeMathCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SafeMathCallerSession struct {
	Contract *SafeMathCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// SafeMathTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SafeMathTransactorSession struct {
	Contract     *SafeMathTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// SafeMathRaw is an auto generated low-level Go binding around an Ethereum contract.
type SafeMathRaw struct {
	Contract *SafeMath // Generic contract binding to access the raw methods on
}

// SafeMathCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SafeMathCallerRaw struct {
	Contract *SafeMathCaller // Generic read-only contract binding to access the raw methods on
}

// SafeMathTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SafeMathTransactorRaw struct {
	Contract *SafeMathTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSafeMath creates a new instance of SafeMath, bound to a specific deployed contract.
func NewSafeMath(address common.Address, backend bind.ContractBackend) (*SafeMath, error) {
	contract, err := bindSafeMath(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SafeMath{SafeMathCaller: SafeMathCaller{contract: contract}, SafeMathTransactor: SafeMathTransactor{contract: contract}, SafeMathFilterer: SafeMathFilterer{contract: contract}}, nil
}

// NewSafeMathCaller creates a new read-only instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathCaller(address common.Address, caller bind.ContractCaller) (*SafeMathCaller, error) {
	contract, err := bindSafeMath(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SafeMathCaller{contract: contract}, nil
}

// NewSafeMathTransactor creates a new write-only instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathTransactor(address common.Address, transactor bind.ContractTransactor) (*SafeMathTransactor, error) {
	contract, err := bindSafeMath(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SafeMathTransactor{contract: contract}, nil
}

// NewSafeMathFilterer creates a new log filterer instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathFilterer(address common.Address, filterer bind.ContractFilterer) (*SafeMathFilterer, error) {
	contract, err := bindSafeMath(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SafeMathFilterer{contract: contract}, nil
}

// bindSafeMath binds a generic wrapper to an already deployed contract.
func bindSafeMath(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeMathABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SafeMath *SafeMathRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SafeMath.Contract.SafeMathCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SafeMath *SafeMathRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SafeMath.Contract.SafeMathTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SafeMath *SafeMathRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SafeMath.Contract.SafeMathTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SafeMath *SafeMathCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SafeMath.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SafeMath *SafeMathTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SafeMath.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SafeMath *SafeMathTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SafeMath.Contract.contract.Transact(opts, method, params...)
}

// SousChefABI is the input ABI used to generate the binding from.
const SousChefABI = "[{\"inputs\":[{\"internalType\":\"contractIBEP20\",\"name\":\"_syrup\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_rewardPerBlock\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_startBlock\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_endBlock\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"EmergencyWithdraw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"addressLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"addressList\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bonusEndBlock\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"emergencyWithdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"pendingReward\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"poolInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"lastRewardBlock\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"accRewardPerShare\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardPerBlock\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"startBlock\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"syrup\",\"outputs\":[{\"internalType\":\"contractIBEP20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"updatePool\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"userInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDebt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardPending\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// SousChefFuncSigs maps the 4-byte function signature to its string representation.
var SousChefFuncSigs = map[string]string{
	"dc881888": "addressLength()",
	"b810fb43": "addressList(uint256)",
	"1aed6553": "bonusEndBlock()",
	"b6b55f25": "deposit(uint256)",
	"db2e21bc": "emergencyWithdraw()",
	"f40f0f52": "pendingReward(address)",
	"5a2f3d09": "poolInfo()",
	"8ae39cac": "rewardPerBlock()",
	"48cd4cb1": "startBlock()",
	"86a952c4": "syrup()",
	"e3161ddd": "updatePool()",
	"1959a002": "userInfo(address)",
	"2e1a7d4d": "withdraw(uint256)",
}

// SousChefBin is the compiled bytecode used for deploying new contracts.
var SousChefBin = "0x608060405234801561001057600080fd5b50604051610ecc380380610ecc8339818101604052608081101561003357600080fd5b508051602080830151604080850151606090950151600080546001600160a01b039096166001600160a01b03199096169590951785556001929092556006859055600791909155805180820190915283815201819052600291909155600355610e2b806100a16000396000f3fe608060405234801561001057600080fd5b50600436106100cf5760003560e01c80638ae39cac1161008c578063db2e21bc11610066578063db2e21bc146101e0578063dc881888146101e8578063e3161ddd146101f0578063f40f0f52146101f8576100cf565b80638ae39cac1461019e578063b6b55f25146101a6578063b810fb43146101c3576100cf565b80631959a002146100d45780631aed6553146101185780632e1a7d4d1461013257806348cd4cb1146101515780635a2f3d091461015957806386a952c41461017a575b600080fd5b6100fa600480360360208110156100ea57600080fd5b50356001600160a01b031661021e565b60408051938452602084019290925282820152519081900360600190f35b61012061023f565b60408051918252519081900360200190f35b61014f6004803603602081101561014857600080fd5b5035610245565b005b6101206103b2565b6101616103b8565b6040805192835260208301919091528051918290030190f35b6101826103c1565b604080516001600160a01b039092168252519081900360200190f35b6101206103d0565b61014f600480360360208110156101bc57600080fd5b50356103d6565b610182600480360360208110156101d957600080fd5b503561054c565b61014f610573565b6101206105e8565b61014f6105ee565b6101206004803603602081101561020e57600080fd5b50356001600160a01b03166106dd565b60046020526000908152604090208054600182015460029092015490919083565b60075481565b60008111610285576040805162461bcd60e51b81526020600482015260086024820152670616d6f756e7420360c41b604482015290519081900360640190fd5b33600090815260046020526040902080548211156102e1576040805162461bcd60e51b81526020600482015260146024820152730eed2e8d0c8e4c2ee7440dcdee840cadcdeeaced60631b604482015290519081900360640190fd5b6102e96105ee565b600054610300906001600160a01b0316338461080f565b6103478160020154610341836001015461033b64e8d4a51000610335600260010154886000015461086690919063ffffffff16565b906108c8565b9061090a565b9061094c565b60028201558054610358908361090a565b8082556003546103739164e8d4a51000916103359190610866565b600182015560408051838152905133917f884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364919081900360200190a25050565b60065481565b60025460035482565b6000546001600160a01b031681565b60015481565b60008111610416576040805162461bcd60e51b81526020600482015260086024820152670616d6f756e7420360c41b604482015290519081900360640190fd5b33600090815260046020526040902061042d6105ee565b600054610445906001600160a01b03163330856109a6565b805415801561045657506002810154155b801561046457506001810154155b156104ac57600580546001810182556000919091527f036b6384b5eca791c62761152d0c79bb0604c104a5fb6f4eb0703f3154bb3db00180546001600160a01b031916331790555b6104e18160020154610341836001015461033b64e8d4a51000610335600260010154886000015461086690919063ffffffff16565b600282015580546104f2908361094c565b80825560035461050d9164e8d4a51000916103359190610866565b600182015560408051838152905133917fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c919081900360200190a25050565b6005818154811061055957fe5b6000918252602090912001546001600160a01b0316905081565b33600081815260046020526040812080549154909261059d926001600160a01b039092169161080f565b8054604080519182525133917f5fafa99d0643513820be26656b45130b01e1c03062e1266bf36f88cbd3bd9695919081900360200190a2600080825560018201819055600290910155565b60055490565b60025443116105fc576106db565b60008054604080516370a0823160e01b815230600482015290516001600160a01b03909216916370a0823191602480820192602092909190829003018186803b15801561064857600080fd5b505afa15801561065c573d6000803e3d6000fd5b505050506040513d602081101561067257600080fd5b50519050806106855750436002556106db565b600061069660026000015443610a06565b905060006106af6001548361086690919063ffffffff16565b90506106d06106c7846103358464e8d4a51000610866565b6003549061094c565b600355505043600255505b565b6001600160a01b038082166000908152600460208181526040808420600354855483516370a0823160e01b8152309681019690965292519596600296929591948894909116926370a08231926024808201939291829003018186803b15801561074557600080fd5b505afa158015610759573d6000803e3d6000fd5b505050506040513d602081101561076f57600080fd5b505184549091504311801561078357508015155b156107d5576000610798856000015443610a06565b905060006107b16001548361086690919063ffffffff16565b90506107d06107c9846103358464e8d4a51000610866565b859061094c565b935050505b6108058360020154610341856001015461033b64e8d4a51000610335888a6000015461086690919063ffffffff16565b9695505050505050565b604080516001600160a01b038416602482015260448082018490528251808303909101815260649091019091526020810180516001600160e01b031663a9059cbb60e01b179052610861908490610a40565b505050565b600082610875575060006108c2565b8282028284828161088257fe5b04146108bf5760405162461bcd60e51b8152600401808060200182810382526021815260200180610dd56021913960400191505060405180910390fd5b90505b92915050565b60006108bf83836040518060400160405280601a81526020017f536166654d6174683a206469766973696f6e206279207a65726f000000000000815250610af1565b60006108bf83836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f770000815250610b93565b6000828201838110156108bf576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b604080516001600160a01b0380861660248301528416604482015260648082018490528251808303909101815260849091019091526020810180516001600160e01b03166323b872dd60e01b179052610a00908590610a40565b50505050565b60006007548211610a2257610a1b828461090a565b90506108c2565b6007548310610a33575060006108c2565b600754610a1b908461090a565b6060610a95826040518060400160405280602081526020017f5361666542455032303a206c6f772d6c6576656c2063616c6c206661696c6564815250856001600160a01b0316610bed9092919063ffffffff16565b80519091501561086157808060200190516020811015610ab457600080fd5b50516108615760405162461bcd60e51b815260040180806020018281038252602a815260200180610dab602a913960400191505060405180910390fd5b60008183610b7d5760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b83811015610b42578181015183820152602001610b2a565b50505050905090810190601f168015610b6f5780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b506000838581610b8957fe5b0495945050505050565b60008184841115610be55760405162461bcd60e51b8152602060048201818152835160248401528351909283926044909101919085019080838360008315610b42578181015183820152602001610b2a565b505050900390565b6060610bfc8484600085610c04565b949350505050565b6060610c0f85610d71565b610c60576040805162461bcd60e51b815260206004820152601d60248201527f416464726573733a2063616c6c20746f206e6f6e2d636f6e7472616374000000604482015290519081900360640190fd5b60006060866001600160a01b031685876040518082805190602001908083835b60208310610c9f5780518252601f199092019160209182019101610c80565b6001836020036101000a03801982511681845116808217855250505050505090500191505060006040518083038185875af1925050503d8060008114610d01576040519150601f19603f3d011682016040523d82523d6000602084013e610d06565b606091505b50915091508115610d1a579150610bfc9050565b805115610d2a5780518082602001fd5b60405162461bcd60e51b8152602060048201818152865160248401528651879391928392604401919085019080838360008315610b42578181015183820152602001610b2a565b6000813f7fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470818114801590610bfc57505015159291505056fe5361666542455032303a204245503230206f7065726174696f6e20646964206e6f742073756363656564536166654d6174683a206d756c7469706c69636174696f6e206f766572666c6f77a264697066735822122072a442dc45618e105eed117aace951c399e934f1c0912f3c886d0d30ea501a6864736f6c634300060c0033"

// DeploySousChef deploys a new Ethereum contract, binding an instance of SousChef to it.
func DeploySousChef(auth *bind.TransactOpts, backend bind.ContractBackend, _syrup common.Address, _rewardPerBlock *big.Int, _startBlock *big.Int, _endBlock *big.Int) (common.Address, *types.Transaction, *SousChef, error) {
	parsed, err := abi.JSON(strings.NewReader(SousChefABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SousChefBin), backend, _syrup, _rewardPerBlock, _startBlock, _endBlock)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SousChef{SousChefCaller: SousChefCaller{contract: contract}, SousChefTransactor: SousChefTransactor{contract: contract}, SousChefFilterer: SousChefFilterer{contract: contract}}, nil
}

// SousChef is an auto generated Go binding around an Ethereum contract.
type SousChef struct {
	SousChefCaller     // Read-only binding to the contract
	SousChefTransactor // Write-only binding to the contract
	SousChefFilterer   // Log filterer for contract events
}

// SousChefCaller is an auto generated read-only Go binding around an Ethereum contract.
type SousChefCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SousChefTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SousChefTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SousChefFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SousChefFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SousChefSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SousChefSession struct {
	Contract     *SousChef         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SousChefCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SousChefCallerSession struct {
	Contract *SousChefCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// SousChefTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SousChefTransactorSession struct {
	Contract     *SousChefTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// SousChefRaw is an auto generated low-level Go binding around an Ethereum contract.
type SousChefRaw struct {
	Contract *SousChef // Generic contract binding to access the raw methods on
}

// SousChefCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SousChefCallerRaw struct {
	Contract *SousChefCaller // Generic read-only contract binding to access the raw methods on
}

// SousChefTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SousChefTransactorRaw struct {
	Contract *SousChefTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSousChef creates a new instance of SousChef, bound to a specific deployed contract.
func NewSousChef(address common.Address, backend bind.ContractBackend) (*SousChef, error) {
	contract, err := bindSousChef(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SousChef{SousChefCaller: SousChefCaller{contract: contract}, SousChefTransactor: SousChefTransactor{contract: contract}, SousChefFilterer: SousChefFilterer{contract: contract}}, nil
}

// NewSousChefCaller creates a new read-only instance of SousChef, bound to a specific deployed contract.
func NewSousChefCaller(address common.Address, caller bind.ContractCaller) (*SousChefCaller, error) {
	contract, err := bindSousChef(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SousChefCaller{contract: contract}, nil
}

// NewSousChefTransactor creates a new write-only instance of SousChef, bound to a specific deployed contract.
func NewSousChefTransactor(address common.Address, transactor bind.ContractTransactor) (*SousChefTransactor, error) {
	contract, err := bindSousChef(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SousChefTransactor{contract: contract}, nil
}

// NewSousChefFilterer creates a new log filterer instance of SousChef, bound to a specific deployed contract.
func NewSousChefFilterer(address common.Address, filterer bind.ContractFilterer) (*SousChefFilterer, error) {
	contract, err := bindSousChef(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SousChefFilterer{contract: contract}, nil
}

// bindSousChef binds a generic wrapper to an already deployed contract.
func bindSousChef(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SousChefABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SousChef *SousChefRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SousChef.Contract.SousChefCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SousChef *SousChefRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SousChef.Contract.SousChefTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SousChef *SousChefRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SousChef.Contract.SousChefTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SousChef *SousChefCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SousChef.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SousChef *SousChefTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SousChef.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SousChef *SousChefTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SousChef.Contract.contract.Transact(opts, method, params...)
}

// AddressLength is a free data retrieval call binding the contract method 0xdc881888.
//
// Solidity: function addressLength() view returns(uint256)
func (_SousChef *SousChefCaller) AddressLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SousChef.contract.Call(opts, &out, "addressLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AddressLength is a free data retrieval call binding the contract method 0xdc881888.
//
// Solidity: function addressLength() view returns(uint256)
func (_SousChef *SousChefSession) AddressLength() (*big.Int, error) {
	return _SousChef.Contract.AddressLength(&_SousChef.CallOpts)
}

// AddressLength is a free data retrieval call binding the contract method 0xdc881888.
//
// Solidity: function addressLength() view returns(uint256)
func (_SousChef *SousChefCallerSession) AddressLength() (*big.Int, error) {
	return _SousChef.Contract.AddressLength(&_SousChef.CallOpts)
}

// AddressList is a free data retrieval call binding the contract method 0xb810fb43.
//
// Solidity: function addressList(uint256 ) view returns(address)
func (_SousChef *SousChefCaller) AddressList(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _SousChef.contract.Call(opts, &out, "addressList", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AddressList is a free data retrieval call binding the contract method 0xb810fb43.
//
// Solidity: function addressList(uint256 ) view returns(address)
func (_SousChef *SousChefSession) AddressList(arg0 *big.Int) (common.Address, error) {
	return _SousChef.Contract.AddressList(&_SousChef.CallOpts, arg0)
}

// AddressList is a free data retrieval call binding the contract method 0xb810fb43.
//
// Solidity: function addressList(uint256 ) view returns(address)
func (_SousChef *SousChefCallerSession) AddressList(arg0 *big.Int) (common.Address, error) {
	return _SousChef.Contract.AddressList(&_SousChef.CallOpts, arg0)
}

// BonusEndBlock is a free data retrieval call binding the contract method 0x1aed6553.
//
// Solidity: function bonusEndBlock() view returns(uint256)
func (_SousChef *SousChefCaller) BonusEndBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SousChef.contract.Call(opts, &out, "bonusEndBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BonusEndBlock is a free data retrieval call binding the contract method 0x1aed6553.
//
// Solidity: function bonusEndBlock() view returns(uint256)
func (_SousChef *SousChefSession) BonusEndBlock() (*big.Int, error) {
	return _SousChef.Contract.BonusEndBlock(&_SousChef.CallOpts)
}

// BonusEndBlock is a free data retrieval call binding the contract method 0x1aed6553.
//
// Solidity: function bonusEndBlock() view returns(uint256)
func (_SousChef *SousChefCallerSession) BonusEndBlock() (*big.Int, error) {
	return _SousChef.Contract.BonusEndBlock(&_SousChef.CallOpts)
}

// PendingReward is a free data retrieval call binding the contract method 0xf40f0f52.
//
// Solidity: function pendingReward(address _user) view returns(uint256)
func (_SousChef *SousChefCaller) PendingReward(opts *bind.CallOpts, _user common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SousChef.contract.Call(opts, &out, "pendingReward", _user)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PendingReward is a free data retrieval call binding the contract method 0xf40f0f52.
//
// Solidity: function pendingReward(address _user) view returns(uint256)
func (_SousChef *SousChefSession) PendingReward(_user common.Address) (*big.Int, error) {
	return _SousChef.Contract.PendingReward(&_SousChef.CallOpts, _user)
}

// PendingReward is a free data retrieval call binding the contract method 0xf40f0f52.
//
// Solidity: function pendingReward(address _user) view returns(uint256)
func (_SousChef *SousChefCallerSession) PendingReward(_user common.Address) (*big.Int, error) {
	return _SousChef.Contract.PendingReward(&_SousChef.CallOpts, _user)
}

// PoolInfo is a free data retrieval call binding the contract method 0x5a2f3d09.
//
// Solidity: function poolInfo() view returns(uint256 lastRewardBlock, uint256 accRewardPerShare)
func (_SousChef *SousChefCaller) PoolInfo(opts *bind.CallOpts) (struct {
	LastRewardBlock   *big.Int
	AccRewardPerShare *big.Int
}, error) {
	var out []interface{}
	err := _SousChef.contract.Call(opts, &out, "poolInfo")

	outstruct := new(struct {
		LastRewardBlock   *big.Int
		AccRewardPerShare *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.LastRewardBlock = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.AccRewardPerShare = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// PoolInfo is a free data retrieval call binding the contract method 0x5a2f3d09.
//
// Solidity: function poolInfo() view returns(uint256 lastRewardBlock, uint256 accRewardPerShare)
func (_SousChef *SousChefSession) PoolInfo() (struct {
	LastRewardBlock   *big.Int
	AccRewardPerShare *big.Int
}, error) {
	return _SousChef.Contract.PoolInfo(&_SousChef.CallOpts)
}

// PoolInfo is a free data retrieval call binding the contract method 0x5a2f3d09.
//
// Solidity: function poolInfo() view returns(uint256 lastRewardBlock, uint256 accRewardPerShare)
func (_SousChef *SousChefCallerSession) PoolInfo() (struct {
	LastRewardBlock   *big.Int
	AccRewardPerShare *big.Int
}, error) {
	return _SousChef.Contract.PoolInfo(&_SousChef.CallOpts)
}

// RewardPerBlock is a free data retrieval call binding the contract method 0x8ae39cac.
//
// Solidity: function rewardPerBlock() view returns(uint256)
func (_SousChef *SousChefCaller) RewardPerBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SousChef.contract.Call(opts, &out, "rewardPerBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RewardPerBlock is a free data retrieval call binding the contract method 0x8ae39cac.
//
// Solidity: function rewardPerBlock() view returns(uint256)
func (_SousChef *SousChefSession) RewardPerBlock() (*big.Int, error) {
	return _SousChef.Contract.RewardPerBlock(&_SousChef.CallOpts)
}

// RewardPerBlock is a free data retrieval call binding the contract method 0x8ae39cac.
//
// Solidity: function rewardPerBlock() view returns(uint256)
func (_SousChef *SousChefCallerSession) RewardPerBlock() (*big.Int, error) {
	return _SousChef.Contract.RewardPerBlock(&_SousChef.CallOpts)
}

// StartBlock is a free data retrieval call binding the contract method 0x48cd4cb1.
//
// Solidity: function startBlock() view returns(uint256)
func (_SousChef *SousChefCaller) StartBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SousChef.contract.Call(opts, &out, "startBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StartBlock is a free data retrieval call binding the contract method 0x48cd4cb1.
//
// Solidity: function startBlock() view returns(uint256)
func (_SousChef *SousChefSession) StartBlock() (*big.Int, error) {
	return _SousChef.Contract.StartBlock(&_SousChef.CallOpts)
}

// StartBlock is a free data retrieval call binding the contract method 0x48cd4cb1.
//
// Solidity: function startBlock() view returns(uint256)
func (_SousChef *SousChefCallerSession) StartBlock() (*big.Int, error) {
	return _SousChef.Contract.StartBlock(&_SousChef.CallOpts)
}

// Syrup is a free data retrieval call binding the contract method 0x86a952c4.
//
// Solidity: function syrup() view returns(address)
func (_SousChef *SousChefCaller) Syrup(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SousChef.contract.Call(opts, &out, "syrup")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Syrup is a free data retrieval call binding the contract method 0x86a952c4.
//
// Solidity: function syrup() view returns(address)
func (_SousChef *SousChefSession) Syrup() (common.Address, error) {
	return _SousChef.Contract.Syrup(&_SousChef.CallOpts)
}

// Syrup is a free data retrieval call binding the contract method 0x86a952c4.
//
// Solidity: function syrup() view returns(address)
func (_SousChef *SousChefCallerSession) Syrup() (common.Address, error) {
	return _SousChef.Contract.Syrup(&_SousChef.CallOpts)
}

// UserInfo is a free data retrieval call binding the contract method 0x1959a002.
//
// Solidity: function userInfo(address ) view returns(uint256 amount, uint256 rewardDebt, uint256 rewardPending)
func (_SousChef *SousChefCaller) UserInfo(opts *bind.CallOpts, arg0 common.Address) (struct {
	Amount        *big.Int
	RewardDebt    *big.Int
	RewardPending *big.Int
}, error) {
	var out []interface{}
	err := _SousChef.contract.Call(opts, &out, "userInfo", arg0)

	outstruct := new(struct {
		Amount        *big.Int
		RewardDebt    *big.Int
		RewardPending *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Amount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.RewardDebt = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.RewardPending = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// UserInfo is a free data retrieval call binding the contract method 0x1959a002.
//
// Solidity: function userInfo(address ) view returns(uint256 amount, uint256 rewardDebt, uint256 rewardPending)
func (_SousChef *SousChefSession) UserInfo(arg0 common.Address) (struct {
	Amount        *big.Int
	RewardDebt    *big.Int
	RewardPending *big.Int
}, error) {
	return _SousChef.Contract.UserInfo(&_SousChef.CallOpts, arg0)
}

// UserInfo is a free data retrieval call binding the contract method 0x1959a002.
//
// Solidity: function userInfo(address ) view returns(uint256 amount, uint256 rewardDebt, uint256 rewardPending)
func (_SousChef *SousChefCallerSession) UserInfo(arg0 common.Address) (struct {
	Amount        *big.Int
	RewardDebt    *big.Int
	RewardPending *big.Int
}, error) {
	return _SousChef.Contract.UserInfo(&_SousChef.CallOpts, arg0)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 _amount) returns()
func (_SousChef *SousChefTransactor) Deposit(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _SousChef.contract.Transact(opts, "deposit", _amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 _amount) returns()
func (_SousChef *SousChefSession) Deposit(_amount *big.Int) (*types.Transaction, error) {
	return _SousChef.Contract.Deposit(&_SousChef.TransactOpts, _amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 _amount) returns()
func (_SousChef *SousChefTransactorSession) Deposit(_amount *big.Int) (*types.Transaction, error) {
	return _SousChef.Contract.Deposit(&_SousChef.TransactOpts, _amount)
}

// EmergencyWithdraw is a paid mutator transaction binding the contract method 0xdb2e21bc.
//
// Solidity: function emergencyWithdraw() returns()
func (_SousChef *SousChefTransactor) EmergencyWithdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SousChef.contract.Transact(opts, "emergencyWithdraw")
}

// EmergencyWithdraw is a paid mutator transaction binding the contract method 0xdb2e21bc.
//
// Solidity: function emergencyWithdraw() returns()
func (_SousChef *SousChefSession) EmergencyWithdraw() (*types.Transaction, error) {
	return _SousChef.Contract.EmergencyWithdraw(&_SousChef.TransactOpts)
}

// EmergencyWithdraw is a paid mutator transaction binding the contract method 0xdb2e21bc.
//
// Solidity: function emergencyWithdraw() returns()
func (_SousChef *SousChefTransactorSession) EmergencyWithdraw() (*types.Transaction, error) {
	return _SousChef.Contract.EmergencyWithdraw(&_SousChef.TransactOpts)
}

// UpdatePool is a paid mutator transaction binding the contract method 0xe3161ddd.
//
// Solidity: function updatePool() returns()
func (_SousChef *SousChefTransactor) UpdatePool(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SousChef.contract.Transact(opts, "updatePool")
}

// UpdatePool is a paid mutator transaction binding the contract method 0xe3161ddd.
//
// Solidity: function updatePool() returns()
func (_SousChef *SousChefSession) UpdatePool() (*types.Transaction, error) {
	return _SousChef.Contract.UpdatePool(&_SousChef.TransactOpts)
}

// UpdatePool is a paid mutator transaction binding the contract method 0xe3161ddd.
//
// Solidity: function updatePool() returns()
func (_SousChef *SousChefTransactorSession) UpdatePool() (*types.Transaction, error) {
	return _SousChef.Contract.UpdatePool(&_SousChef.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _amount) returns()
func (_SousChef *SousChefTransactor) Withdraw(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _SousChef.contract.Transact(opts, "withdraw", _amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _amount) returns()
func (_SousChef *SousChefSession) Withdraw(_amount *big.Int) (*types.Transaction, error) {
	return _SousChef.Contract.Withdraw(&_SousChef.TransactOpts, _amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _amount) returns()
func (_SousChef *SousChefTransactorSession) Withdraw(_amount *big.Int) (*types.Transaction, error) {
	return _SousChef.Contract.Withdraw(&_SousChef.TransactOpts, _amount)
}

// SousChefDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the SousChef contract.
type SousChefDepositIterator struct {
	Event *SousChefDeposit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SousChefDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SousChefDeposit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SousChefDeposit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SousChefDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SousChefDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SousChefDeposit represents a Deposit event raised by the SousChef contract.
type SousChefDeposit struct {
	User   common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address indexed user, uint256 amount)
func (_SousChef *SousChefFilterer) FilterDeposit(opts *bind.FilterOpts, user []common.Address) (*SousChefDepositIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _SousChef.contract.FilterLogs(opts, "Deposit", userRule)
	if err != nil {
		return nil, err
	}
	return &SousChefDepositIterator{contract: _SousChef.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address indexed user, uint256 amount)
func (_SousChef *SousChefFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *SousChefDeposit, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _SousChef.contract.WatchLogs(opts, "Deposit", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SousChefDeposit)
				if err := _SousChef.contract.UnpackLog(event, "Deposit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeposit is a log parse operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address indexed user, uint256 amount)
func (_SousChef *SousChefFilterer) ParseDeposit(log types.Log) (*SousChefDeposit, error) {
	event := new(SousChefDeposit)
	if err := _SousChef.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SousChefEmergencyWithdrawIterator is returned from FilterEmergencyWithdraw and is used to iterate over the raw logs and unpacked data for EmergencyWithdraw events raised by the SousChef contract.
type SousChefEmergencyWithdrawIterator struct {
	Event *SousChefEmergencyWithdraw // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SousChefEmergencyWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SousChefEmergencyWithdraw)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SousChefEmergencyWithdraw)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SousChefEmergencyWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SousChefEmergencyWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SousChefEmergencyWithdraw represents a EmergencyWithdraw event raised by the SousChef contract.
type SousChefEmergencyWithdraw struct {
	User   common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterEmergencyWithdraw is a free log retrieval operation binding the contract event 0x5fafa99d0643513820be26656b45130b01e1c03062e1266bf36f88cbd3bd9695.
//
// Solidity: event EmergencyWithdraw(address indexed user, uint256 amount)
func (_SousChef *SousChefFilterer) FilterEmergencyWithdraw(opts *bind.FilterOpts, user []common.Address) (*SousChefEmergencyWithdrawIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _SousChef.contract.FilterLogs(opts, "EmergencyWithdraw", userRule)
	if err != nil {
		return nil, err
	}
	return &SousChefEmergencyWithdrawIterator{contract: _SousChef.contract, event: "EmergencyWithdraw", logs: logs, sub: sub}, nil
}

// WatchEmergencyWithdraw is a free log subscription operation binding the contract event 0x5fafa99d0643513820be26656b45130b01e1c03062e1266bf36f88cbd3bd9695.
//
// Solidity: event EmergencyWithdraw(address indexed user, uint256 amount)
func (_SousChef *SousChefFilterer) WatchEmergencyWithdraw(opts *bind.WatchOpts, sink chan<- *SousChefEmergencyWithdraw, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _SousChef.contract.WatchLogs(opts, "EmergencyWithdraw", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SousChefEmergencyWithdraw)
				if err := _SousChef.contract.UnpackLog(event, "EmergencyWithdraw", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEmergencyWithdraw is a log parse operation binding the contract event 0x5fafa99d0643513820be26656b45130b01e1c03062e1266bf36f88cbd3bd9695.
//
// Solidity: event EmergencyWithdraw(address indexed user, uint256 amount)
func (_SousChef *SousChefFilterer) ParseEmergencyWithdraw(log types.Log) (*SousChefEmergencyWithdraw, error) {
	event := new(SousChefEmergencyWithdraw)
	if err := _SousChef.contract.UnpackLog(event, "EmergencyWithdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SousChefWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the SousChef contract.
type SousChefWithdrawIterator struct {
	Event *SousChefWithdraw // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SousChefWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SousChefWithdraw)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SousChefWithdraw)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SousChefWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SousChefWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SousChefWithdraw represents a Withdraw event raised by the SousChef contract.
type SousChefWithdraw struct {
	User   common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed user, uint256 amount)
func (_SousChef *SousChefFilterer) FilterWithdraw(opts *bind.FilterOpts, user []common.Address) (*SousChefWithdrawIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _SousChef.contract.FilterLogs(opts, "Withdraw", userRule)
	if err != nil {
		return nil, err
	}
	return &SousChefWithdrawIterator{contract: _SousChef.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed user, uint256 amount)
func (_SousChef *SousChefFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *SousChefWithdraw, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _SousChef.contract.WatchLogs(opts, "Withdraw", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SousChefWithdraw)
				if err := _SousChef.contract.UnpackLog(event, "Withdraw", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdraw is a log parse operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed user, uint256 amount)
func (_SousChef *SousChefFilterer) ParseWithdraw(log types.Log) (*SousChefWithdraw, error) {
	event := new(SousChefWithdraw)
	if err := _SousChef.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
