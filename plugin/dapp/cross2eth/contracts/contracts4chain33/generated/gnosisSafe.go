// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package generated

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

// EnumABI is the input ABI used to generate the binding from.
const EnumABI = "[]"

// EnumBin is the compiled bytecode used for deploying new contracts.
var EnumBin = "0x6080604052348015600f57600080fd5b50603f80601d6000396000f3fe6080604052600080fdfea2646970667358221220173d047b75ec0bb7cf2dd65e9a9c094370c02928543ca74ea922e8e4d1d9bba864736f6c63430007000033"

// DeployEnum deploys a new Ethereum contract, binding an instance of Enum to it.
func DeployEnum(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Enum, error) {
	parsed, err := abi.JSON(strings.NewReader(EnumABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(EnumBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Enum{EnumCaller: EnumCaller{contract: contract}, EnumTransactor: EnumTransactor{contract: contract}, EnumFilterer: EnumFilterer{contract: contract}}, nil
}

// Enum is an auto generated Go binding around an Ethereum contract.
type Enum struct {
	EnumCaller     // Read-only binding to the contract
	EnumTransactor // Write-only binding to the contract
	EnumFilterer   // Log filterer for contract events
}

// EnumCaller is an auto generated read-only Go binding around an Ethereum contract.
type EnumCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EnumTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EnumTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EnumFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EnumFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EnumSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EnumSession struct {
	Contract     *Enum             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EnumCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EnumCallerSession struct {
	Contract *EnumCaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// EnumTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EnumTransactorSession struct {
	Contract     *EnumTransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EnumRaw is an auto generated low-level Go binding around an Ethereum contract.
type EnumRaw struct {
	Contract *Enum // Generic contract binding to access the raw methods on
}

// EnumCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EnumCallerRaw struct {
	Contract *EnumCaller // Generic read-only contract binding to access the raw methods on
}

// EnumTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EnumTransactorRaw struct {
	Contract *EnumTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEnum creates a new instance of Enum, bound to a specific deployed contract.
func NewEnum(address common.Address, backend bind.ContractBackend) (*Enum, error) {
	contract, err := bindEnum(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Enum{EnumCaller: EnumCaller{contract: contract}, EnumTransactor: EnumTransactor{contract: contract}, EnumFilterer: EnumFilterer{contract: contract}}, nil
}

// NewEnumCaller creates a new read-only instance of Enum, bound to a specific deployed contract.
func NewEnumCaller(address common.Address, caller bind.ContractCaller) (*EnumCaller, error) {
	contract, err := bindEnum(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EnumCaller{contract: contract}, nil
}

// NewEnumTransactor creates a new write-only instance of Enum, bound to a specific deployed contract.
func NewEnumTransactor(address common.Address, transactor bind.ContractTransactor) (*EnumTransactor, error) {
	contract, err := bindEnum(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EnumTransactor{contract: contract}, nil
}

// NewEnumFilterer creates a new log filterer instance of Enum, bound to a specific deployed contract.
func NewEnumFilterer(address common.Address, filterer bind.ContractFilterer) (*EnumFilterer, error) {
	contract, err := bindEnum(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EnumFilterer{contract: contract}, nil
}

// bindEnum binds a generic wrapper to an already deployed contract.
func bindEnum(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EnumABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Enum *EnumRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Enum.Contract.EnumCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Enum *EnumRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Enum.Contract.EnumTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Enum *EnumRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Enum.Contract.EnumTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Enum *EnumCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Enum.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Enum *EnumTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Enum.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Enum *EnumTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Enum.Contract.contract.Transact(opts, method, params...)
}

// EtherPaymentFallbackABI is the input ABI used to generate the binding from.
const EtherPaymentFallbackABI = "[{\"stateMutability\":\"payable\",\"type\":\"receive\"}]"

// EtherPaymentFallbackBin is the compiled bytecode used for deploying new contracts.
var EtherPaymentFallbackBin = "0x6080604052348015600f57600080fd5b50604580601d6000396000f3fe608060405236600a57005b600080fdfea264697066735822122082df967331ddbf7eeaac3fac431d838250411e9dc164eb757e2765c222045bc064736f6c63430007000033"

// DeployEtherPaymentFallback deploys a new Ethereum contract, binding an instance of EtherPaymentFallback to it.
func DeployEtherPaymentFallback(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *EtherPaymentFallback, error) {
	parsed, err := abi.JSON(strings.NewReader(EtherPaymentFallbackABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(EtherPaymentFallbackBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &EtherPaymentFallback{EtherPaymentFallbackCaller: EtherPaymentFallbackCaller{contract: contract}, EtherPaymentFallbackTransactor: EtherPaymentFallbackTransactor{contract: contract}, EtherPaymentFallbackFilterer: EtherPaymentFallbackFilterer{contract: contract}}, nil
}

// EtherPaymentFallback is an auto generated Go binding around an Ethereum contract.
type EtherPaymentFallback struct {
	EtherPaymentFallbackCaller     // Read-only binding to the contract
	EtherPaymentFallbackTransactor // Write-only binding to the contract
	EtherPaymentFallbackFilterer   // Log filterer for contract events
}

// EtherPaymentFallbackCaller is an auto generated read-only Go binding around an Ethereum contract.
type EtherPaymentFallbackCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EtherPaymentFallbackTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EtherPaymentFallbackTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EtherPaymentFallbackFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EtherPaymentFallbackFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EtherPaymentFallbackSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EtherPaymentFallbackSession struct {
	Contract     *EtherPaymentFallback // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// EtherPaymentFallbackCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EtherPaymentFallbackCallerSession struct {
	Contract *EtherPaymentFallbackCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// EtherPaymentFallbackTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EtherPaymentFallbackTransactorSession struct {
	Contract     *EtherPaymentFallbackTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// EtherPaymentFallbackRaw is an auto generated low-level Go binding around an Ethereum contract.
type EtherPaymentFallbackRaw struct {
	Contract *EtherPaymentFallback // Generic contract binding to access the raw methods on
}

// EtherPaymentFallbackCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EtherPaymentFallbackCallerRaw struct {
	Contract *EtherPaymentFallbackCaller // Generic read-only contract binding to access the raw methods on
}

// EtherPaymentFallbackTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EtherPaymentFallbackTransactorRaw struct {
	Contract *EtherPaymentFallbackTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEtherPaymentFallback creates a new instance of EtherPaymentFallback, bound to a specific deployed contract.
func NewEtherPaymentFallback(address common.Address, backend bind.ContractBackend) (*EtherPaymentFallback, error) {
	contract, err := bindEtherPaymentFallback(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EtherPaymentFallback{EtherPaymentFallbackCaller: EtherPaymentFallbackCaller{contract: contract}, EtherPaymentFallbackTransactor: EtherPaymentFallbackTransactor{contract: contract}, EtherPaymentFallbackFilterer: EtherPaymentFallbackFilterer{contract: contract}}, nil
}

// NewEtherPaymentFallbackCaller creates a new read-only instance of EtherPaymentFallback, bound to a specific deployed contract.
func NewEtherPaymentFallbackCaller(address common.Address, caller bind.ContractCaller) (*EtherPaymentFallbackCaller, error) {
	contract, err := bindEtherPaymentFallback(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EtherPaymentFallbackCaller{contract: contract}, nil
}

// NewEtherPaymentFallbackTransactor creates a new write-only instance of EtherPaymentFallback, bound to a specific deployed contract.
func NewEtherPaymentFallbackTransactor(address common.Address, transactor bind.ContractTransactor) (*EtherPaymentFallbackTransactor, error) {
	contract, err := bindEtherPaymentFallback(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EtherPaymentFallbackTransactor{contract: contract}, nil
}

// NewEtherPaymentFallbackFilterer creates a new log filterer instance of EtherPaymentFallback, bound to a specific deployed contract.
func NewEtherPaymentFallbackFilterer(address common.Address, filterer bind.ContractFilterer) (*EtherPaymentFallbackFilterer, error) {
	contract, err := bindEtherPaymentFallback(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EtherPaymentFallbackFilterer{contract: contract}, nil
}

// bindEtherPaymentFallback binds a generic wrapper to an already deployed contract.
func bindEtherPaymentFallback(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EtherPaymentFallbackABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EtherPaymentFallback *EtherPaymentFallbackRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _EtherPaymentFallback.Contract.EtherPaymentFallbackCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EtherPaymentFallback *EtherPaymentFallbackRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EtherPaymentFallback.Contract.EtherPaymentFallbackTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EtherPaymentFallback *EtherPaymentFallbackRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EtherPaymentFallback.Contract.EtherPaymentFallbackTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EtherPaymentFallback *EtherPaymentFallbackCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _EtherPaymentFallback.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EtherPaymentFallback *EtherPaymentFallbackTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EtherPaymentFallback.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EtherPaymentFallback *EtherPaymentFallbackTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EtherPaymentFallback.Contract.contract.Transact(opts, method, params...)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_EtherPaymentFallback *EtherPaymentFallbackTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EtherPaymentFallback.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_EtherPaymentFallback *EtherPaymentFallbackSession) Receive() (*types.Transaction, error) {
	return _EtherPaymentFallback.Contract.Receive(&_EtherPaymentFallback.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_EtherPaymentFallback *EtherPaymentFallbackTransactorSession) Receive() (*types.Transaction, error) {
	return _EtherPaymentFallback.Contract.Receive(&_EtherPaymentFallback.TransactOpts)
}

// ExecutorABI is the input ABI used to generate the binding from.
const ExecutorABI = "[]"

// ExecutorBin is the compiled bytecode used for deploying new contracts.
var ExecutorBin = "0x6080604052348015600f57600080fd5b50603f80601d6000396000f3fe6080604052600080fdfea2646970667358221220b6fe8d495cd46e16bcdbc5fea6c1d020bb8a836ea0a5df1ebc3cbe99a6928ec464736f6c63430007000033"

// DeployExecutor deploys a new Ethereum contract, binding an instance of Executor to it.
func DeployExecutor(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Executor, error) {
	parsed, err := abi.JSON(strings.NewReader(ExecutorABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ExecutorBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Executor{ExecutorCaller: ExecutorCaller{contract: contract}, ExecutorTransactor: ExecutorTransactor{contract: contract}, ExecutorFilterer: ExecutorFilterer{contract: contract}}, nil
}

// Executor is an auto generated Go binding around an Ethereum contract.
type Executor struct {
	ExecutorCaller     // Read-only binding to the contract
	ExecutorTransactor // Write-only binding to the contract
	ExecutorFilterer   // Log filterer for contract events
}

// ExecutorCaller is an auto generated read-only Go binding around an Ethereum contract.
type ExecutorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExecutorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ExecutorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExecutorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ExecutorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExecutorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ExecutorSession struct {
	Contract     *Executor         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ExecutorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ExecutorCallerSession struct {
	Contract *ExecutorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ExecutorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ExecutorTransactorSession struct {
	Contract     *ExecutorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ExecutorRaw is an auto generated low-level Go binding around an Ethereum contract.
type ExecutorRaw struct {
	Contract *Executor // Generic contract binding to access the raw methods on
}

// ExecutorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ExecutorCallerRaw struct {
	Contract *ExecutorCaller // Generic read-only contract binding to access the raw methods on
}

// ExecutorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ExecutorTransactorRaw struct {
	Contract *ExecutorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewExecutor creates a new instance of Executor, bound to a specific deployed contract.
func NewExecutor(address common.Address, backend bind.ContractBackend) (*Executor, error) {
	contract, err := bindExecutor(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Executor{ExecutorCaller: ExecutorCaller{contract: contract}, ExecutorTransactor: ExecutorTransactor{contract: contract}, ExecutorFilterer: ExecutorFilterer{contract: contract}}, nil
}

// NewExecutorCaller creates a new read-only instance of Executor, bound to a specific deployed contract.
func NewExecutorCaller(address common.Address, caller bind.ContractCaller) (*ExecutorCaller, error) {
	contract, err := bindExecutor(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ExecutorCaller{contract: contract}, nil
}

// NewExecutorTransactor creates a new write-only instance of Executor, bound to a specific deployed contract.
func NewExecutorTransactor(address common.Address, transactor bind.ContractTransactor) (*ExecutorTransactor, error) {
	contract, err := bindExecutor(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ExecutorTransactor{contract: contract}, nil
}

// NewExecutorFilterer creates a new log filterer instance of Executor, bound to a specific deployed contract.
func NewExecutorFilterer(address common.Address, filterer bind.ContractFilterer) (*ExecutorFilterer, error) {
	contract, err := bindExecutor(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ExecutorFilterer{contract: contract}, nil
}

// bindExecutor binds a generic wrapper to an already deployed contract.
func bindExecutor(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ExecutorABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Executor *ExecutorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Executor.Contract.ExecutorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Executor *ExecutorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Executor.Contract.ExecutorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Executor *ExecutorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Executor.Contract.ExecutorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Executor *ExecutorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Executor.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Executor *ExecutorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Executor.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Executor *ExecutorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Executor.Contract.contract.Transact(opts, method, params...)
}

// FallbackManagerABI is the input ABI used to generate the binding from.
const FallbackManagerABI = "[{\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"handler\",\"type\":\"address\"}],\"name\":\"setFallbackHandler\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// FallbackManagerFuncSigs maps the 4-byte function signature to its string representation.
var FallbackManagerFuncSigs = map[string]string{
	"f08a0323": "setFallbackHandler(address)",
}

// FallbackManagerBin is the compiled bytecode used for deploying new contracts.
var FallbackManagerBin = "0x608060405234801561001057600080fd5b50610186806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c8063f08a032314610084575b7f6c9a6c4a39284e37ed1cf53d337577d14212a4870fb976a4366c693b939918d580548061005557005b36600080373360601b365260008060143601600080855af190503d6000803e8061007e573d6000fd5b503d6000f35b6100aa6004803603602081101561009a57600080fd5b50356001600160a01b03166100ac565b005b6100b46100c0565b6100bd81610100565b50565b3330146100fe5760405162461bcd60e51b815260040180806020018281038252602c815260200180610125602c913960400191505060405180910390fd5b565b7f6c9a6c4a39284e37ed1cf53d337577d14212a4870fb976a4366c693b939918d55556fe4d6574686f642063616e206f6e6c792062652063616c6c65642066726f6d207468697320636f6e7472616374a2646970667358221220262a744e1e1d5c103a624b623b13705e4b4924a1431590603aa991ca2af6b26664736f6c63430007000033"

// DeployFallbackManager deploys a new Ethereum contract, binding an instance of FallbackManager to it.
func DeployFallbackManager(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *FallbackManager, error) {
	parsed, err := abi.JSON(strings.NewReader(FallbackManagerABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(FallbackManagerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &FallbackManager{FallbackManagerCaller: FallbackManagerCaller{contract: contract}, FallbackManagerTransactor: FallbackManagerTransactor{contract: contract}, FallbackManagerFilterer: FallbackManagerFilterer{contract: contract}}, nil
}

// FallbackManager is an auto generated Go binding around an Ethereum contract.
type FallbackManager struct {
	FallbackManagerCaller     // Read-only binding to the contract
	FallbackManagerTransactor // Write-only binding to the contract
	FallbackManagerFilterer   // Log filterer for contract events
}

// FallbackManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type FallbackManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FallbackManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FallbackManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FallbackManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FallbackManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FallbackManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FallbackManagerSession struct {
	Contract     *FallbackManager  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FallbackManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FallbackManagerCallerSession struct {
	Contract *FallbackManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// FallbackManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FallbackManagerTransactorSession struct {
	Contract     *FallbackManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// FallbackManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type FallbackManagerRaw struct {
	Contract *FallbackManager // Generic contract binding to access the raw methods on
}

// FallbackManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FallbackManagerCallerRaw struct {
	Contract *FallbackManagerCaller // Generic read-only contract binding to access the raw methods on
}

// FallbackManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FallbackManagerTransactorRaw struct {
	Contract *FallbackManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFallbackManager creates a new instance of FallbackManager, bound to a specific deployed contract.
func NewFallbackManager(address common.Address, backend bind.ContractBackend) (*FallbackManager, error) {
	contract, err := bindFallbackManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FallbackManager{FallbackManagerCaller: FallbackManagerCaller{contract: contract}, FallbackManagerTransactor: FallbackManagerTransactor{contract: contract}, FallbackManagerFilterer: FallbackManagerFilterer{contract: contract}}, nil
}

// NewFallbackManagerCaller creates a new read-only instance of FallbackManager, bound to a specific deployed contract.
func NewFallbackManagerCaller(address common.Address, caller bind.ContractCaller) (*FallbackManagerCaller, error) {
	contract, err := bindFallbackManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FallbackManagerCaller{contract: contract}, nil
}

// NewFallbackManagerTransactor creates a new write-only instance of FallbackManager, bound to a specific deployed contract.
func NewFallbackManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*FallbackManagerTransactor, error) {
	contract, err := bindFallbackManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FallbackManagerTransactor{contract: contract}, nil
}

// NewFallbackManagerFilterer creates a new log filterer instance of FallbackManager, bound to a specific deployed contract.
func NewFallbackManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*FallbackManagerFilterer, error) {
	contract, err := bindFallbackManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FallbackManagerFilterer{contract: contract}, nil
}

// bindFallbackManager binds a generic wrapper to an already deployed contract.
func bindFallbackManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(FallbackManagerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FallbackManager *FallbackManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FallbackManager.Contract.FallbackManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FallbackManager *FallbackManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FallbackManager.Contract.FallbackManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FallbackManager *FallbackManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FallbackManager.Contract.FallbackManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FallbackManager *FallbackManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FallbackManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FallbackManager *FallbackManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FallbackManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FallbackManager *FallbackManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FallbackManager.Contract.contract.Transact(opts, method, params...)
}

// SetFallbackHandler is a paid mutator transaction binding the contract method 0xf08a0323.
//
// Solidity: function setFallbackHandler(address handler) returns()
func (_FallbackManager *FallbackManagerTransactor) SetFallbackHandler(opts *bind.TransactOpts, handler common.Address) (*types.Transaction, error) {
	return _FallbackManager.contract.Transact(opts, "setFallbackHandler", handler)
}

// SetFallbackHandler is a paid mutator transaction binding the contract method 0xf08a0323.
//
// Solidity: function setFallbackHandler(address handler) returns()
func (_FallbackManager *FallbackManagerSession) SetFallbackHandler(handler common.Address) (*types.Transaction, error) {
	return _FallbackManager.Contract.SetFallbackHandler(&_FallbackManager.TransactOpts, handler)
}

// SetFallbackHandler is a paid mutator transaction binding the contract method 0xf08a0323.
//
// Solidity: function setFallbackHandler(address handler) returns()
func (_FallbackManager *FallbackManagerTransactorSession) SetFallbackHandler(handler common.Address) (*types.Transaction, error) {
	return _FallbackManager.Contract.SetFallbackHandler(&_FallbackManager.TransactOpts, handler)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_FallbackManager *FallbackManagerTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _FallbackManager.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_FallbackManager *FallbackManagerSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _FallbackManager.Contract.Fallback(&_FallbackManager.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_FallbackManager *FallbackManagerTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _FallbackManager.Contract.Fallback(&_FallbackManager.TransactOpts, calldata)
}

// GnosisSafeABI is the input ABI used to generate the binding from.
const GnosisSafeABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AddedOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"approvedHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ApproveHash\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"name\":\"ChangedThreshold\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"DisabledModule\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"EnabledModule\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"txHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"payment\",\"type\":\"uint256\"}],\"name\":\"ExecutionFailure\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"ExecutionFromModuleFailure\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"ExecutionFromModuleSuccess\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"result\",\"type\":\"bool\"}],\"name\":\"ExecutionResult\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"txHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"payment\",\"type\":\"uint256\"}],\"name\":\"ExecutionSuccess\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"RemovedOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"msgHash\",\"type\":\"bytes32\"}],\"name\":\"SignMsg\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"SignatureRecover\",\"type\":\"event\"},{\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"inputs\":[],\"name\":\"NAME\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"}],\"name\":\"addOwnerWithThreshold\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hashToApprove\",\"type\":\"bytes32\"}],\"name\":\"approveHash\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"approvedHashes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"}],\"name\":\"changeThreshold\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"checkSignatures\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"prevModule\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"disableModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"domainSeparator\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"enableModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"safeTxGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"baseGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPrice\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"gasToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"refundReceiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"encodeTransactionData\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"safeTxGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"baseGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPrice\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"gasToken\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"refundReceiver\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"execTransaction\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"}],\"name\":\"execTransactionFromModule\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"}],\"name\":\"execTransactionFromModuleReturnData\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"returnData\",\"type\":\"bytes\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChainId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"getMessageHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"start\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"pageSize\",\"type\":\"uint256\"}],\"name\":\"getModulesPaginated\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"array\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"next\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwners\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSelfBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"getStorageAt\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"safeTxGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"baseGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPrice\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"gasToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"refundReceiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"getTransactionHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"isModuleEnabled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"isOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"prevOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"}],\"name\":\"removeOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"}],\"name\":\"requiredTxGas\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"handler\",\"type\":\"address\"}],\"name\":\"setFallbackHandler\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_owners\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"fallbackHandler\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"paymentToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"payment\",\"type\":\"uint256\"},{\"internalType\":\"addresspayable\",\"name\":\"paymentReceiver\",\"type\":\"address\"}],\"name\":\"setup\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"signMessage\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"signedMessages\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"targetContract\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"calldataPayload\",\"type\":\"bytes\"}],\"name\":\"simulateDelegatecall\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"targetContract\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"calldataPayload\",\"type\":\"bytes\"}],\"name\":\"simulateDelegatecallInternal\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"prevOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"oldOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"swapOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]"

// GnosisSafeFuncSigs maps the 4-byte function signature to its string representation.
var GnosisSafeFuncSigs = map[string]string{
	"a3f4df7e": "NAME()",
	"ffa1ad74": "VERSION()",
	"0d582f13": "addOwnerWithThreshold(address,uint256)",
	"d4d9bdcd": "approveHash(bytes32)",
	"7d832974": "approvedHashes(address,bytes32)",
	"694e80c3": "changeThreshold(uint256)",
	"934f3a11": "checkSignatures(bytes32,bytes,bytes)",
	"e009cfde": "disableModule(address,address)",
	"f698da25": "domainSeparator()",
	"610b5925": "enableModule(address)",
	"e86637db": "encodeTransactionData(address,uint256,bytes,uint8,uint256,uint256,uint256,address,address,uint256)",
	"6a761202": "execTransaction(address,uint256,bytes,uint8,uint256,uint256,uint256,address,address,bytes)",
	"468721a7": "execTransactionFromModule(address,uint256,bytes,uint8)",
	"5229073f": "execTransactionFromModuleReturnData(address,uint256,bytes,uint8)",
	"3408e470": "getChainId()",
	"0a1028c4": "getMessageHash(bytes)",
	"cc2f8452": "getModulesPaginated(address,uint256)",
	"a0e67e2b": "getOwners()",
	"048a5fed": "getSelfBalance()",
	"5624b25b": "getStorageAt(uint256,uint256)",
	"e75235b8": "getThreshold()",
	"d8d11f78": "getTransactionHash(address,uint256,bytes,uint8,uint256,uint256,uint256,address,address,uint256)",
	"2d9ad53d": "isModuleEnabled(address)",
	"2f54bf6e": "isOwner(address)",
	"affed0e0": "nonce()",
	"f8dc5dd9": "removeOwner(address,address,uint256)",
	"c4ca3a9c": "requiredTxGas(address,uint256,bytes,uint8)",
	"f08a0323": "setFallbackHandler(address)",
	"b63e800d": "setup(address[],uint256,address,bytes,address,address,uint256,address)",
	"85a5affe": "signMessage(bytes)",
	"5ae6bd37": "signedMessages(bytes32)",
	"f84436bd": "simulateDelegatecall(address,bytes)",
	"43218e19": "simulateDelegatecallInternal(address,bytes)",
	"e318b52b": "swapOwner(address,address,address)",
}

// GnosisSafeBin is the compiled bytecode used for deploying new contracts.
var GnosisSafeBin = "0x608060405234801561001057600080fd5b50613b09806100206000396000f3fe6080604052600436106101fd5760003560e01c8063a0e67e2b1161010d578063e009cfde116100a0578063f08a03231161006f578063f08a0323146110e3578063f698da2514611116578063f84436bd1461112b578063f8dc5dd9146111ec578063ffa1ad741461122f57610204565b8063e009cfde14610f54578063e318b52b14610f8f578063e75235b814610fd4578063e86637db14610fe957610204565b8063c4ca3a9c116100dc578063c4ca3a9c14610d00578063cc2f845214610d93578063d4d9bdcd14610e30578063d8d11f7814610e5a57610204565b8063a0e67e2b14610b71578063a3f4df7e14610bd6578063affed0e014610beb578063b63e800d14610c0057610204565b80635229073f11610190578063694e80c31161015f578063694e80c3146107e65780636a761202146108105780637d8329741461098057806385a5affe146109b9578063934f3a1114610a3457610204565b80635229073f1461060d5780635624b25b146107595780635ae6bd3714610789578063610b5925146107b357610204565b80632f54bf6e116101cc5780632f54bf6e146103c45780633408e470146103f757806343218e191461040c578063468721a71461054257610204565b8063048a5fed1461026a5780630a1028c4146102915780630d582f13146103425780632d9ad53d1461037d57610204565b3661020457005b34801561021057600080fd5b507f6c9a6c4a39284e37ed1cf53d337577d14212a4870fb976a4366c693b939918d580548061023b57005b36600080373360601b365260008060143601600080855af190503d6000803e80610264573d6000fd5b503d6000f35b34801561027657600080fd5b5061027f611244565b60408051918252519081900360200190f35b34801561029d57600080fd5b5061027f600480360360208110156102b457600080fd5b810190602081018135600160201b8111156102ce57600080fd5b8201836020820111156102e057600080fd5b803590602001918460018302840111600160201b8311171561030157600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550611248945050505050565b34801561034e57600080fd5b5061037b6004803603604081101561036557600080fd5b506001600160a01b038135169060200135611305565b005b34801561038957600080fd5b506103b0600480360360208110156103a057600080fd5b50356001600160a01b03166114a5565b604080519115158252519081900360200190f35b3480156103d057600080fd5b506103b0600480360360208110156103e757600080fd5b50356001600160a01b03166114e0565b34801561040357600080fd5b5061027f611518565b34801561041857600080fd5b506104cd6004803603604081101561042f57600080fd5b6001600160a01b038235169190810190604081016020820135600160201b81111561045957600080fd5b82018360208201111561046b57600080fd5b803590602001918460018302840111600160201b8311171561048c57600080fd5b91908080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525092955061151c945050505050565b6040805160208082528351818301528351919283929083019185019080838360005b838110156105075781810151838201526020016104ef565b50505050905090810190601f1680156105345780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561054e57600080fd5b506103b06004803603608081101561056557600080fd5b6001600160a01b0382351691602081013591810190606081016040820135600160201b81111561059457600080fd5b8201836020820111156105a657600080fd5b803590602001918460018302840111600160201b831117156105c757600080fd5b91908080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509295505050903560ff16915061164a9050565b34801561061957600080fd5b506106d86004803603608081101561063057600080fd5b6001600160a01b0382351691602081013591810190606081016040820135600160201b81111561065f57600080fd5b82018360208201111561067157600080fd5b803590602001918460018302840111600160201b8311171561069257600080fd5b91908080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509295505050903560ff1691506117289050565b60405180831515815260200180602001828103825283818151815260200191508051906020019080838360005b8381101561071d578181015183820152602001610705565b50505050905090810190601f16801561074a5780820380516001836020036101000a031916815260200191505b50935050505060405180910390f35b34801561076557600080fd5b506104cd6004803603604081101561077c57600080fd5b508035906020013561175e565b34801561079557600080fd5b5061027f600480360360208110156107ac57600080fd5b50356117d1565b3480156107bf57600080fd5b5061037b600480360360208110156107d657600080fd5b50356001600160a01b03166117e3565b3480156107f257600080fd5b5061037b6004803603602081101561080957600080fd5b503561195f565b6103b0600480360361014081101561082757600080fd5b6001600160a01b0382351691602081013591810190606081016040820135600160201b81111561085657600080fd5b82018360208201111561086857600080fd5b803590602001918460018302840111600160201b8311171561088957600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929560ff8535169560208601359560408101359550606081013594506001600160a01b0360808201358116945060a08201351692919060e081019060c00135600160201b81111561090c57600080fd5b82018360208201111561091e57600080fd5b803590602001918460018302840111600160201b8311171561093f57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550611a23945050505050565b34801561098c57600080fd5b5061027f600480360360408110156109a357600080fd5b506001600160a01b038135169060200135611cfa565b3480156109c557600080fd5b5061037b600480360360208110156109dc57600080fd5b810190602081018135600160201b8111156109f657600080fd5b820183602082011115610a0857600080fd5b803590602001918460018302840111600160201b83111715610a2957600080fd5b509092509050611d17565b348015610a4057600080fd5b5061037b60048036036060811015610a5757600080fd5b81359190810190604081016020820135600160201b811115610a7857600080fd5b820183602082011115610a8a57600080fd5b803590602001918460018302840111600160201b83111715610aab57600080fd5b91908080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509295949360208101935035915050600160201b811115610afd57600080fd5b820183602082011115610b0f57600080fd5b803590602001918460018302840111600160201b83111715610b3057600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550611da2945050505050565b348015610b7d57600080fd5b50610b866123cc565b60408051602080825283518183015283519192839290830191858101910280838360005b83811015610bc2578181015183820152602001610baa565b505050509050019250505060405180910390f35b348015610be257600080fd5b506104cd6124af565b348015610bf757600080fd5b5061027f6124d6565b348015610c0c57600080fd5b5061037b6004803603610100811015610c2457600080fd5b810190602081018135600160201b811115610c3e57600080fd5b820183602082011115610c5057600080fd5b803590602001918460208302840111600160201b83111715610c7157600080fd5b919390928235926001600160a01b03602082013516929190606081019060400135600160201b811115610ca357600080fd5b820183602082011115610cb557600080fd5b803590602001918460018302840111600160201b83111715610cd657600080fd5b91935091506001600160a01b038135811691602081013582169160408201359160600135166124dc565b348015610d0c57600080fd5b5061027f60048036036080811015610d2357600080fd5b6001600160a01b0382351691602081013591810190606081016040820135600160201b811115610d5257600080fd5b820183602082011115610d6457600080fd5b803590602001918460018302840111600160201b83111715610d8557600080fd5b91935091503560ff16612594565b348015610d9f57600080fd5b50610dcc60048036036040811015610db657600080fd5b506001600160a01b038135169060200135612691565b6040518080602001836001600160a01b03168152602001828103825284818151815260200191508051906020019060200280838360005b83811015610e1b578181015183820152602001610e03565b50505050905001935050505060405180910390f35b348015610e3c57600080fd5b5061037b60048036036020811015610e5357600080fd5b503561277c565b348015610e6657600080fd5b5061027f6004803603610140811015610e7e57600080fd5b6001600160a01b0382351691602081013591810190606081016040820135600160201b811115610ead57600080fd5b820183602082011115610ebf57600080fd5b803590602001918460018302840111600160201b83111715610ee057600080fd5b91908080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509295505060ff833516935050506020810135906040810135906060810135906001600160a01b03608082013581169160a08101359091169060c0013561282e565b348015610f6057600080fd5b5061037b60048036036040811015610f7757600080fd5b506001600160a01b0381358116916020013516612859565b348015610f9b57600080fd5b5061037b60048036036060811015610fb257600080fd5b506001600160a01b0381358116916020810135821691604090910135166129ad565b348015610fe057600080fd5b5061027f612c20565b348015610ff557600080fd5b506104cd600480360361014081101561100d57600080fd5b6001600160a01b0382351691602081013591810190606081016040820135600160201b81111561103c57600080fd5b82018360208201111561104e57600080fd5b803590602001918460018302840111600160201b8311171561106f57600080fd5b91908080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509295505060ff833516935050506020810135906040810135906060810135906001600160a01b03608082013581169160a08101359091169060c00135612c26565b3480156110ef57600080fd5b5061037b6004803603602081101561110657600080fd5b50356001600160a01b0316612d4f565b34801561112257600080fd5b5061027f612d63565b34801561113757600080fd5b506104cd6004803603604081101561114e57600080fd5b6001600160a01b038235169190810190604081016020820135600160201b81111561117857600080fd5b82018360208201111561118a57600080fd5b803590602001918460018302840111600160201b831117156111ab57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550612dd1945050505050565b3480156111f857600080fd5b5061037b6004803603606081101561120f57600080fd5b506001600160a01b03813581169160208101359091169060400135612fa8565b34801561123b57600080fd5b506104cd61314b565b4790565b8051602080830191909120604080517f60b3cbf8b4a223d68d641b3b6ddf9a298e7f33710cf3d3a9d1146b5a6150fbca8185015280820192909252805180830382018152606090920190528051910120600090601960f81b600160f81b6112ad612d63565b8360405160200180856001600160f81b0319168152600101846001600160f81b031916815260010183815260200182815260200194505050505060405160208183030381529060405280519060200120915050919050565b61130d61316c565b6001600160a01b0382161580159061132f57506001600160a01b038216600114155b801561134457506001600160a01b0382163014155b611383576040805162461bcd60e51b815260206004820152601e60248201526000805160206137e3833981519152604482015290519081900360640190fd5b6001600160a01b0382811660009081526002602052604090205416156113f0576040805162461bcd60e51b815260206004820152601b60248201527f4164647265737320697320616c726561647920616e206f776e65720000000000604482015290519081900360640190fd5b600260209081527fe90b7bceb6e7df5418fb78d8ee546e97c83a08bbccc01a0644d599ccd2a7c2e080546001600160a01b03858116600081815260408082208054949095166001600160a01b031994851617909455600190819052845490921681179093556003805490910190558051918252517f9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea26929181900390910190a180600454146114a1576114a18161195f565b5050565b600060016001600160a01b038316148015906114da57506001600160a01b038281166000908152600160205260409020541615155b92915050565b60006001600160a01b0382166001148015906114da5750506001600160a01b0390811660009081526002602052604090205416151590565b4690565b606060006060846001600160a01b0316846040518082805190602001908083835b6020831061155c5780518252601f19909201916020918201910161153d565b6001836020036101000a038019825116818451168082178552505050505050905001915050600060405180830381855af49150503d80600081146115bc576040519150601f19603f3d011682016040523d82523d6000602084013e6115c1565b606091505b509150915061164281836040516020018083805190602001908083835b602083106115fd5780518252601f1990920191602091820191016115de565b6001836020036101000a03801982511681845116808217855250505050505090500182151560f81b8152600101925050506040516020818303038152906040526131ac565b505092915050565b6000336001148015906116745750336000908152600160205260409020546001600160a01b031615155b6116af5760405162461bcd60e51b8152600401808060200182810382526030815260200180613a216030913960400191505060405180910390fd5b6116bc858585855a6131b4565b905080156116f45760405133907f6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb890600090a2611720565b60405133907facd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd37590600090a25b949350505050565b600060606117388686868661164a565b915060405160203d0181016040523d81523d6000602083013e8091505094509492505050565b6060808260200267ffffffffffffffff8111801561177b57600080fd5b506040519080825280601f01601f1916602001820160405280156117a6576020820181803683370190505b50905060005b838110156117c957848101546020808302840101526001016117ac565b509392505050565b60076020526000908152604090205481565b6117eb61316c565b6001600160a01b0381161580159061180d57506001600160a01b038116600114155b61185e576040805162461bcd60e51b815260206004820152601f60248201527f496e76616c6964206d6f64756c6520616464726573732070726f766964656400604482015290519081900360640190fd5b6001600160a01b0381811660009081526001602052604090205416156118cb576040805162461bcd60e51b815260206004820152601d60248201527f4d6f64756c652068617320616c7265616479206265656e206164646564000000604482015290519081900360640190fd5b600160208181527fcc69885fda6bcc1a4ace058b4a62bf5e179ea78fd58a1ccd71c22cc9b688792f80546001600160a01b03858116600081815260408082208054949095166001600160a01b0319948516179094559590955282541684179091558051928352517fecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f84409281900390910190a150565b61196761316c565b6003548111156119a85760405162461bcd60e51b81526004018080602001828103825260238152602001806138736023913960400191505060405180910390fd5b60018110156119e85760405162461bcd60e51b815260040180806020018281038252602481526020018061399a6024913960400191505060405180910390fd5b60048190556040805182815290517f610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c939181900360200190a150565b6000806060611a3c8d8d8d8d8d8d8d8d8d600554612c26565b6005805460010190558051602082012092509050611a5b828286611da2565b50611a70603f60408a02046109c48a016131f6565b6101f4015a1015611ab25760405162461bcd60e51b815260040180806020018281038252602a815260200180613aaa602a913960400191505060405180910390fd5b8a15611bca578a4711611af65760405162461bcd60e51b81526004018080602001828103825260288152602001806139136028913960400191505060405180910390fd5b6040516001600160a01b038d16908c156108fc02908d906000818181858888f19350505050158015611b2c573d6000803e3d6000fd5b506040805160016020820152818152600f818301526e115e1958dd5d1a5bdb94995cdd5b1d608a1b606082015290517f36bd3cb3e572bed2e31aa120b605e9d3cb596f0703790070410c5f0b0ac5e34e9181900360800190a160408051828152602081018d905281517f442e715f626346e8c54381002da614f62bee8d27386535b2521ec8540898556e929181900390910190a16001915050611cec565b60005a9050611bed8d8d8d8d8b15611be2578d611be8565b6109c45a035b6131b4565b9250611bfa5a829061320f565b905060008715611c1457611c11828a8a8a8a613224565b90505b604080518515156020820152818152600f818301526e115e1958dd5d1a5bdb94995cdd5b1d608a1b606082015290517f36bd3cb3e572bed2e31aa120b605e9d3cb596f0703790070410c5f0b0ac5e34e9181900360800190a183611ca95760405162461bcd60e51b815260040180806020018281038252602981526020018061393b6029913960400191505060405180910390fd5b604080518481526020810183905281517f442e715f626346e8c54381002da614f62bee8d27386535b2521ec8540898556e929181900390910190a1600193505050505b9a9950505050505050505050565b600860209081526000928352604080842090915290825290205481565b611d1f61316c565b6000611d6083838080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061124892505050565b600081815260076020526040808220600190555191925082917fe7f4675038f4f6034dfcbbb24c4dc08e4ebf10eb9d257d3d02c0f38d122ac6e49190a2505050565b60045480611df7576040805162461bcd60e51b815260206004820152601e60248201527f5468726573686f6c64206e6565647320746f20626520646566696e6564210000604482015290519081900360640190fd5b611e02816041613338565b82511015611e57576040805162461bcd60e51b815260206004820152601960248201527f5369676e617475726573206461746120746f6f2073686f727400000000000000604482015290519081900360640190fd5b6000808060008060005b868110156123c057611e73888261335f565b9195509350915060ff841661211c579193508391611e92876041613338565b821015611ed05760405162461bcd60e51b81526004018080602001828103825260378152602001806139be6037913960400191505060405180910390fd5b8751611edd83602061337d565b1115611f1a5760405162461bcd60e51b8152600401808060200182810382526037815260200180613a516037913960400191505060405180910390fd5b602082890181015189519091611f3d908390611f3790879061337d565b9061337d565b1115611f7a5760405162461bcd60e51b81526004018080602001828103825260368152602001806139646036913960400191505060405180910390fd5b60606020848b010190506320c13b0b60e01b6001600160e01b031916876001600160a01b03166320c13b0b8d846040518363ffffffff1660e01b8152600401808060200180602001838103835285818151815260200191508051906020019080838360005b83811015611ff7578181015183820152602001611fdf565b50505050905090810190601f1680156120245780820380516001836020036101000a031916815260200191505b50838103825284518152845160209182019186019080838360005b8381101561205757818101518382015260200161203f565b50505050905090810190601f1680156120845780820380516001836020036101000a031916815260200191505b5094505050505060206040518083038186803b1580156120a357600080fd5b505afa1580156120b7573d6000803e3d6000fd5b505050506040513d60208110156120cd57600080fd5b50516001600160e01b031916146121155760405162461bcd60e51b81526004018080602001828103825260238152602001806138286023913960400191505060405180910390fd5b50506122eb565b8360ff16600114156121bc579193508391336001600160a01b038416148061216657506001600160a01b03851660009081526008602090815260408083208d845290915290205415155b6121b7576040805162461bcd60e51b815260206004820152601a60248201527f4861736820686173206e6f74206265656e20617070726f766564000000000000604482015290519081900360640190fd5b6122eb565b601e8460ff1611156122845760018a60405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c018281526020019150506040516020818303038152906040528051906020012060048603858560405160008152602001604052604051808581526020018460ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa158015612273573d6000803e3d6000fd5b5050506020604051035194506122eb565b60018a85858560405160008152602001604052604051808581526020018460ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa1580156122de573d6000803e3d6000fd5b5050506020604051035194505b6040805160ff861681526001600160a01b03871660208201528082018c905290517f5a06eb92afc8e6aed01ca739e4a5f3d7b76155d973f4428c2f7316095d98abd39181900360600190a16001600160a01b03858116600090815260026020526040902054161580159061236957506001600160a01b038516600114155b6123b3576040805162461bcd60e51b8152602060048201526016602482015275125b9d985b1a59081bdddb995c881c1c9bdd9a59195960521b604482015290519081900360640190fd5b9394508493600101611e61565b50505050505050505050565b60608060035467ffffffffffffffff811180156123e857600080fd5b50604051908082528060200260200182016040528015612412578160200160208202803683370190505b506001600090815260026020527fe90b7bceb6e7df5418fb78d8ee546e97c83a08bbccc01a0644d599ccd2a7c2e054919250906001600160a01b03165b6001600160a01b0381166001146124a7578083838151811061246d57fe5b6001600160a01b0392831660209182029290920181019190915291811660009081526002909252604090912054600192909201911661244f565b509091505090565b6040518060400160405280600b81526020016a476e6f736973205361666560a81b81525081565b60055481565b61251a8a8a808060200260200160405190810160405280939291908181526020018383602002808284376000920191909152508c925061338f915050565b6001600160a01b038416156125325761253284613600565b6125728787878080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061362492505050565b81156123c05761258782600060018685613224565b5050505050505050505050565b6000805a90506125dd878787878080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525089925050505a6131b4565b6125e657600080fd5b60005a82039050806040516020018082815260200191505060405160208183030381529060405260405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b8381101561265657818101518382015260200161263e565b50505050905090810190601f1680156126835780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b606060008267ffffffffffffffff811180156126ac57600080fd5b506040519080825280602002602001820160405280156126d6578160200160208202803683370190505b506001600160a01b0380861660009081526001602052604081205492945091165b6001600160a01b0381161580159061271957506001600160a01b038116600114155b801561272457508482105b1561276e578084838151811061273657fe5b6001600160a01b03928316602091820292909201810191909152918116600090815260019283905260409020549290910191166126f7565b908352919491935090915050565b336000908152600260205260409020546001600160a01b03166127e6576040805162461bcd60e51b815260206004820152601e60248201527f4f6e6c79206f776e6572732063616e20617070726f7665206120686173680000604482015290519081900360640190fd5b336000818152600860209081526040808320858452909152808220600190555183917ff2a0eb156472d1440255b0d7c1e19cc07115d1051fe605b0dce69acfec884d9c91a350565b60006128428b8b8b8b8b8b8b8b8b8b612c26565b8051906020012090509a9950505050505050505050565b61286161316c565b6001600160a01b0381161580159061288357506001600160a01b038116600114155b6128d4576040805162461bcd60e51b815260206004820152601f60248201527f496e76616c6964206d6f64756c6520616464726573732070726f766964656400604482015290519081900360640190fd5b6001600160a01b0382811660009081526001602052604090205481169082161461292f5760405162461bcd60e51b815260040180806020018281038252602881526020018061384b6028913960400191505060405180910390fd5b6001600160a01b038181166000818152600160209081526040808320805488871685528285208054919097166001600160a01b031991821617909655928490528254909416909155825191825291517faab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276929181900390910190a15050565b6129b561316c565b6001600160a01b038116158015906129d757506001600160a01b038116600114155b80156129ec57506001600160a01b0381163014155b612a2b576040805162461bcd60e51b815260206004820152601e60248201526000805160206137e3833981519152604482015290519081900360640190fd5b6001600160a01b038181166000908152600260205260409020541615612a98576040805162461bcd60e51b815260206004820152601b60248201527f4164647265737320697320616c726561647920616e206f776e65720000000000604482015290519081900360640190fd5b6001600160a01b03821615801590612aba57506001600160a01b038216600114155b612af9576040805162461bcd60e51b815260206004820152601e60248201526000805160206137e3833981519152604482015290519081900360640190fd5b6001600160a01b03838116600090815260026020526040902054811690831614612b545760405162461bcd60e51b81526004018080602001828103825260268152602001806138ed6026913960400191505060405180910390fd5b6001600160a01b038281166000818152600260209081526040808320805487871680865283862080549289166001600160a01b0319938416179055968a16855282852080548216909717909655928490528254909416909155825191825291517ff8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf929181900390910190a1604080516001600160a01b038316815290517f9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea269181900360200190a1505050565b60045490565b606060007fbb8310d486368db6bd6f849402fdd73ad53d316b5a4b2644ad6efe0f941286d860001b8c8c8c805190602001208c8c8c8c8c8c8c604051602001808c81526020018b6001600160a01b031681526020018a8152602001898152602001886001811115612c9357fe5b8152602001878152602001868152602001858152602001846001600160a01b03168152602001836001600160a01b031681526020018281526020019b505050505050505050505050604051602081830303815290604052805190602001209050601960f81b600160f81b612d05612d63565b604080516001600160f81b0319948516602082015292909316602183015260228201526042808201939093528151808203909301835260620190529b9a5050505050505050505050565b612d5761316c565b612d6081613600565b50565b60007f47e79534a245952e8b16893a336b85a3d9ea9fa8c573f3d803afb92a79469218612d8e611518565b3060405160200180848152602001838152602001826001600160a01b03168152602001935050505060405160208183030381529060405280519060200120905090565b6060807f43218e198a5f5c70ca65adf1973b6285a79c4d29a39cc2a8bb67b912f447dc64848460405160240180836001600160a01b0316815260200180602001828103825283818151815260200191508051906020019080838360005b83811015612e46578181015183820152602001612e2e565b50505050905090810190601f168015612e735780820380516001836020036101000a031916815260200191505b5060408051601f198184030181529181526020820180516001600160e01b03166001600160e01b0319909816979097178752518151919750606096309650889550909350839250908083835b60208310612ede5780518252601f199092019160209182019101612ebf565b6001836020036101000a0380198251168184511680821785525050505050509050019150506000604051808303816000865af19150503d8060008114612f40576040519150601f19603f3d011682016040523d82523d6000602084013e612f45565b606091505b50915050600081600183510381518110612f5b57fe5b602001015160f81c60f81b6001600160f81b031916600160f81b149050612f86826001845103613742565b8015612f96575091506114da9050565b612f9f826131ac565b50505092915050565b612fb061316c565b806001600354031015612ff45760405162461bcd60e51b81526004018080602001828103825260358152602001806138966035913960400191505060405180910390fd5b6001600160a01b0382161580159061301657506001600160a01b038216600114155b613055576040805162461bcd60e51b815260206004820152601e60248201526000805160206137e3833981519152604482015290519081900360640190fd5b6001600160a01b038381166000908152600260205260409020548116908316146130b05760405162461bcd60e51b81526004018080602001828103825260268152602001806138ed6026913960400191505060405180910390fd5b6001600160a01b038281166000818152600260209081526040808320805489871685528285208054919097166001600160a01b03199182161790965592849052825490941690915560038054600019019055825191825291517ff8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf929181900390910190a18060045414613146576131468161195f565b505050565b604051806040016040528060058152602001640312e322e360dc1b81525081565b3330146131aa5760405162461bcd60e51b815260040180806020018281038252602c8152602001806139f5602c913960400191505060405180910390fd5b565b805160208201fd5b600060018360018111156131c457fe5b14156131dd576000808551602087018986f490506131ed565b600080855160208701888a87f190505b95945050505050565b6000818310156132065781613208565b825b9392505050565b60008282111561321e57600080fd5b50900390565b6000806001600160a01b0383161561323c578261323e565b325b90506001600160a01b0384166132d8576132703a861061325e573a613260565b855b61326a898961337d565b90613338565b6040519092506001600160a01b0382169083156108fc029084906000818181858888f193505050506132d35760405162461bcd60e51b8152600401808060200182810382526022815260200180613a886022913960400191505060405180910390fd5b61332e565b6132e68561326a898961337d565b91506132f3848284613746565b61332e5760405162461bcd60e51b81526004018080602001828103825260228152602001806138cb6022913960400191505060405180910390fd5b5095945050505050565b600082613347575060006114da565b8282028284828161335457fe5b041461320857600080fd5b60419081029190910160208101516040820151919092015160ff1692565b60008282018381101561320857600080fd5b600454156133e4576040805162461bcd60e51b815260206004820152601e60248201527f4f776e657273206861766520616c7265616479206265656e2073657475700000604482015290519081900360640190fd5b81518111156134245760405162461bcd60e51b81526004018080602001828103825260238152602001806138736023913960400191505060405180910390fd5b60018110156134645760405162461bcd60e51b815260040180806020018281038252602481526020018061399a6024913960400191505060405180910390fd5b600160005b83518110156135cd57600084828151811061348057fe5b6020026020010151905060006001600160a01b0316816001600160a01b0316141580156134b757506001600160a01b038116600114155b80156134cc57506001600160a01b0381163014155b80156134ea5750806001600160a01b0316836001600160a01b031614155b613529576040805162461bcd60e51b815260206004820152601e60248201526000805160206137e3833981519152604482015290519081900360640190fd5b6001600160a01b038181166000908152600260205260409020541615613596576040805162461bcd60e51b815260206004820181905260248201527f4475706c6963617465206f776e657220616464726573732070726f7669646564604482015290519081900360640190fd5b6001600160a01b03928316600090815260026020526040902080546001600160a01b03191693821693909317909255600101613469565b506001600160a01b0316600090815260026020526040902080546001600160a01b03191660011790559051600355600455565b7f6c9a6c4a39284e37ed1cf53d337577d14212a4870fb976a4366c693b939918d555565b600160008190526020527fcc69885fda6bcc1a4ace058b4a62bf5e179ea78fd58a1ccd71c22cc9b688792f546001600160a01b0316156136955760405162461bcd60e51b81526004018080602001828103825260258152602001806138036025913960400191505060405180910390fd5b6001600081905260208190527fcc69885fda6bcc1a4ace058b4a62bf5e179ea78fd58a1ccd71c22cc9b688792f80546001600160a01b03191690911790556001600160a01b038216156114a1576136f18260008360015a6131b4565b6114a1576040805162461bcd60e51b815260206004820152601f60248201527f436f756c64206e6f742066696e69736820696e697469616c697a6174696f6e00604482015290519081900360640190fd5b9052565b604080516001600160a01b038416602482015260448082018490528251808303909101815260649091019091526020810180516001600160e01b031663a9059cbb60e01b1781528151600092918391829182896127105a03f16040513d81016040523d6000823e3d80156137c557602081146137cd57600094506137d7565b8294506137d7565b8151158315171594505b50505050939250505056fe496e76616c6964206f776e657220616464726573732070726f766964656400004d6f64756c6573206861766520616c7265616479206265656e20696e697469616c697a6564496e76616c696420636f6e7472616374207369676e61747572652070726f7669646564496e76616c696420707265764d6f64756c652c206d6f64756c6520706169722070726f76696465645468726573686f6c642063616e6e6f7420657863656564206f776e657220636f756e744e6577206f776e657220636f756e74206e6565647320746f206265206c6172676572207468616e206e6577207468726573686f6c64436f756c64206e6f74207061792067617320636f737473207769746820746f6b656e496e76616c696420707265764f776e65722c206f776e657220706169722070726f7669646564546865206f66666c696e652077616c6c65742062616c616e6365206973206e6f7420656e6f756768657865635472616e73616374696f6e206e6f74206578656375746564207375636365737366756c6c79496e76616c696420636f6e7472616374207369676e6174757265206c6f636174696f6e3a2064617461206e6f7420636f6d706c6574655468726573686f6c64206e6565647320746f2062652067726561746572207468616e2030496e76616c696420636f6e7472616374207369676e6174757265206c6f636174696f6e3a20696e736964652073746174696320706172744d6574686f642063616e206f6e6c792062652063616c6c65642066726f6d207468697320636f6e74726163744d6574686f642063616e206f6e6c792062652063616c6c65642066726f6d20616e20656e61626c6564206d6f64756c65496e76616c696420636f6e7472616374207369676e6174757265206c6f636174696f6e3a206c656e677468206e6f742070726573656e74436f756c64206e6f74207061792067617320636f73747320776974682065746865724e6f7420656e6f7567682067617320746f20657865637574652073616665207472616e73616374696f6ea2646970667358221220abca3168d763577d347724d72f0efa9d6bf174b564dc93caa1f512c640d2567564736f6c63430007000033"

// DeployGnosisSafe deploys a new Ethereum contract, binding an instance of GnosisSafe to it.
func DeployGnosisSafe(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GnosisSafe, error) {
	parsed, err := abi.JSON(strings.NewReader(GnosisSafeABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(GnosisSafeBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GnosisSafe{GnosisSafeCaller: GnosisSafeCaller{contract: contract}, GnosisSafeTransactor: GnosisSafeTransactor{contract: contract}, GnosisSafeFilterer: GnosisSafeFilterer{contract: contract}}, nil
}

// GnosisSafe is an auto generated Go binding around an Ethereum contract.
type GnosisSafe struct {
	GnosisSafeCaller     // Read-only binding to the contract
	GnosisSafeTransactor // Write-only binding to the contract
	GnosisSafeFilterer   // Log filterer for contract events
}

// GnosisSafeCaller is an auto generated read-only Go binding around an Ethereum contract.
type GnosisSafeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GnosisSafeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GnosisSafeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GnosisSafeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GnosisSafeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GnosisSafeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GnosisSafeSession struct {
	Contract     *GnosisSafe       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GnosisSafeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GnosisSafeCallerSession struct {
	Contract *GnosisSafeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// GnosisSafeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GnosisSafeTransactorSession struct {
	Contract     *GnosisSafeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// GnosisSafeRaw is an auto generated low-level Go binding around an Ethereum contract.
type GnosisSafeRaw struct {
	Contract *GnosisSafe // Generic contract binding to access the raw methods on
}

// GnosisSafeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GnosisSafeCallerRaw struct {
	Contract *GnosisSafeCaller // Generic read-only contract binding to access the raw methods on
}

// GnosisSafeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GnosisSafeTransactorRaw struct {
	Contract *GnosisSafeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGnosisSafe creates a new instance of GnosisSafe, bound to a specific deployed contract.
func NewGnosisSafe(address common.Address, backend bind.ContractBackend) (*GnosisSafe, error) {
	contract, err := bindGnosisSafe(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GnosisSafe{GnosisSafeCaller: GnosisSafeCaller{contract: contract}, GnosisSafeTransactor: GnosisSafeTransactor{contract: contract}, GnosisSafeFilterer: GnosisSafeFilterer{contract: contract}}, nil
}

// NewGnosisSafeCaller creates a new read-only instance of GnosisSafe, bound to a specific deployed contract.
func NewGnosisSafeCaller(address common.Address, caller bind.ContractCaller) (*GnosisSafeCaller, error) {
	contract, err := bindGnosisSafe(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeCaller{contract: contract}, nil
}

// NewGnosisSafeTransactor creates a new write-only instance of GnosisSafe, bound to a specific deployed contract.
func NewGnosisSafeTransactor(address common.Address, transactor bind.ContractTransactor) (*GnosisSafeTransactor, error) {
	contract, err := bindGnosisSafe(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeTransactor{contract: contract}, nil
}

// NewGnosisSafeFilterer creates a new log filterer instance of GnosisSafe, bound to a specific deployed contract.
func NewGnosisSafeFilterer(address common.Address, filterer bind.ContractFilterer) (*GnosisSafeFilterer, error) {
	contract, err := bindGnosisSafe(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeFilterer{contract: contract}, nil
}

// bindGnosisSafe binds a generic wrapper to an already deployed contract.
func bindGnosisSafe(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GnosisSafeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GnosisSafe *GnosisSafeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GnosisSafe.Contract.GnosisSafeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GnosisSafe *GnosisSafeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GnosisSafe.Contract.GnosisSafeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GnosisSafe *GnosisSafeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GnosisSafe.Contract.GnosisSafeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GnosisSafe *GnosisSafeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GnosisSafe.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GnosisSafe *GnosisSafeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GnosisSafe.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GnosisSafe *GnosisSafeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GnosisSafe.Contract.contract.Transact(opts, method, params...)
}

// NAME is a free data retrieval call binding the contract method 0xa3f4df7e.
//
// Solidity: function NAME() view returns(string)
func (_GnosisSafe *GnosisSafeCaller) NAME(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "NAME")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// NAME is a free data retrieval call binding the contract method 0xa3f4df7e.
//
// Solidity: function NAME() view returns(string)
func (_GnosisSafe *GnosisSafeSession) NAME() (string, error) {
	return _GnosisSafe.Contract.NAME(&_GnosisSafe.CallOpts)
}

// NAME is a free data retrieval call binding the contract method 0xa3f4df7e.
//
// Solidity: function NAME() view returns(string)
func (_GnosisSafe *GnosisSafeCallerSession) NAME() (string, error) {
	return _GnosisSafe.Contract.NAME(&_GnosisSafe.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_GnosisSafe *GnosisSafeCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_GnosisSafe *GnosisSafeSession) VERSION() (string, error) {
	return _GnosisSafe.Contract.VERSION(&_GnosisSafe.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_GnosisSafe *GnosisSafeCallerSession) VERSION() (string, error) {
	return _GnosisSafe.Contract.VERSION(&_GnosisSafe.CallOpts)
}

// ApprovedHashes is a free data retrieval call binding the contract method 0x7d832974.
//
// Solidity: function approvedHashes(address , bytes32 ) view returns(uint256)
func (_GnosisSafe *GnosisSafeCaller) ApprovedHashes(opts *bind.CallOpts, arg0 common.Address, arg1 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "approvedHashes", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ApprovedHashes is a free data retrieval call binding the contract method 0x7d832974.
//
// Solidity: function approvedHashes(address , bytes32 ) view returns(uint256)
func (_GnosisSafe *GnosisSafeSession) ApprovedHashes(arg0 common.Address, arg1 [32]byte) (*big.Int, error) {
	return _GnosisSafe.Contract.ApprovedHashes(&_GnosisSafe.CallOpts, arg0, arg1)
}

// ApprovedHashes is a free data retrieval call binding the contract method 0x7d832974.
//
// Solidity: function approvedHashes(address , bytes32 ) view returns(uint256)
func (_GnosisSafe *GnosisSafeCallerSession) ApprovedHashes(arg0 common.Address, arg1 [32]byte) (*big.Int, error) {
	return _GnosisSafe.Contract.ApprovedHashes(&_GnosisSafe.CallOpts, arg0, arg1)
}

// DomainSeparator is a free data retrieval call binding the contract method 0xf698da25.
//
// Solidity: function domainSeparator() view returns(bytes32)
func (_GnosisSafe *GnosisSafeCaller) DomainSeparator(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "domainSeparator")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DomainSeparator is a free data retrieval call binding the contract method 0xf698da25.
//
// Solidity: function domainSeparator() view returns(bytes32)
func (_GnosisSafe *GnosisSafeSession) DomainSeparator() ([32]byte, error) {
	return _GnosisSafe.Contract.DomainSeparator(&_GnosisSafe.CallOpts)
}

// DomainSeparator is a free data retrieval call binding the contract method 0xf698da25.
//
// Solidity: function domainSeparator() view returns(bytes32)
func (_GnosisSafe *GnosisSafeCallerSession) DomainSeparator() ([32]byte, error) {
	return _GnosisSafe.Contract.DomainSeparator(&_GnosisSafe.CallOpts)
}

// EncodeTransactionData is a free data retrieval call binding the contract method 0xe86637db.
//
// Solidity: function encodeTransactionData(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, uint256 _nonce) view returns(bytes)
func (_GnosisSafe *GnosisSafeCaller) EncodeTransactionData(opts *bind.CallOpts, to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, _nonce *big.Int) ([]byte, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "encodeTransactionData", to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, _nonce)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// EncodeTransactionData is a free data retrieval call binding the contract method 0xe86637db.
//
// Solidity: function encodeTransactionData(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, uint256 _nonce) view returns(bytes)
func (_GnosisSafe *GnosisSafeSession) EncodeTransactionData(to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, _nonce *big.Int) ([]byte, error) {
	return _GnosisSafe.Contract.EncodeTransactionData(&_GnosisSafe.CallOpts, to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, _nonce)
}

// EncodeTransactionData is a free data retrieval call binding the contract method 0xe86637db.
//
// Solidity: function encodeTransactionData(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, uint256 _nonce) view returns(bytes)
func (_GnosisSafe *GnosisSafeCallerSession) EncodeTransactionData(to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, _nonce *big.Int) ([]byte, error) {
	return _GnosisSafe.Contract.EncodeTransactionData(&_GnosisSafe.CallOpts, to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, _nonce)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256)
func (_GnosisSafe *GnosisSafeCaller) GetChainId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "getChainId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256)
func (_GnosisSafe *GnosisSafeSession) GetChainId() (*big.Int, error) {
	return _GnosisSafe.Contract.GetChainId(&_GnosisSafe.CallOpts)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256)
func (_GnosisSafe *GnosisSafeCallerSession) GetChainId() (*big.Int, error) {
	return _GnosisSafe.Contract.GetChainId(&_GnosisSafe.CallOpts)
}

// GetMessageHash is a free data retrieval call binding the contract method 0x0a1028c4.
//
// Solidity: function getMessageHash(bytes message) view returns(bytes32)
func (_GnosisSafe *GnosisSafeCaller) GetMessageHash(opts *bind.CallOpts, message []byte) ([32]byte, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "getMessageHash", message)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetMessageHash is a free data retrieval call binding the contract method 0x0a1028c4.
//
// Solidity: function getMessageHash(bytes message) view returns(bytes32)
func (_GnosisSafe *GnosisSafeSession) GetMessageHash(message []byte) ([32]byte, error) {
	return _GnosisSafe.Contract.GetMessageHash(&_GnosisSafe.CallOpts, message)
}

// GetMessageHash is a free data retrieval call binding the contract method 0x0a1028c4.
//
// Solidity: function getMessageHash(bytes message) view returns(bytes32)
func (_GnosisSafe *GnosisSafeCallerSession) GetMessageHash(message []byte) ([32]byte, error) {
	return _GnosisSafe.Contract.GetMessageHash(&_GnosisSafe.CallOpts, message)
}

// GetModulesPaginated is a free data retrieval call binding the contract method 0xcc2f8452.
//
// Solidity: function getModulesPaginated(address start, uint256 pageSize) view returns(address[] array, address next)
func (_GnosisSafe *GnosisSafeCaller) GetModulesPaginated(opts *bind.CallOpts, start common.Address, pageSize *big.Int) (struct {
	Array []common.Address
	Next  common.Address
}, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "getModulesPaginated", start, pageSize)

	outstruct := new(struct {
		Array []common.Address
		Next  common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Array = *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	outstruct.Next = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// GetModulesPaginated is a free data retrieval call binding the contract method 0xcc2f8452.
//
// Solidity: function getModulesPaginated(address start, uint256 pageSize) view returns(address[] array, address next)
func (_GnosisSafe *GnosisSafeSession) GetModulesPaginated(start common.Address, pageSize *big.Int) (struct {
	Array []common.Address
	Next  common.Address
}, error) {
	return _GnosisSafe.Contract.GetModulesPaginated(&_GnosisSafe.CallOpts, start, pageSize)
}

// GetModulesPaginated is a free data retrieval call binding the contract method 0xcc2f8452.
//
// Solidity: function getModulesPaginated(address start, uint256 pageSize) view returns(address[] array, address next)
func (_GnosisSafe *GnosisSafeCallerSession) GetModulesPaginated(start common.Address, pageSize *big.Int) (struct {
	Array []common.Address
	Next  common.Address
}, error) {
	return _GnosisSafe.Contract.GetModulesPaginated(&_GnosisSafe.CallOpts, start, pageSize)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_GnosisSafe *GnosisSafeCaller) GetOwners(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "getOwners")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_GnosisSafe *GnosisSafeSession) GetOwners() ([]common.Address, error) {
	return _GnosisSafe.Contract.GetOwners(&_GnosisSafe.CallOpts)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_GnosisSafe *GnosisSafeCallerSession) GetOwners() ([]common.Address, error) {
	return _GnosisSafe.Contract.GetOwners(&_GnosisSafe.CallOpts)
}

// GetSelfBalance is a free data retrieval call binding the contract method 0x048a5fed.
//
// Solidity: function getSelfBalance() view returns(uint256)
func (_GnosisSafe *GnosisSafeCaller) GetSelfBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "getSelfBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSelfBalance is a free data retrieval call binding the contract method 0x048a5fed.
//
// Solidity: function getSelfBalance() view returns(uint256)
func (_GnosisSafe *GnosisSafeSession) GetSelfBalance() (*big.Int, error) {
	return _GnosisSafe.Contract.GetSelfBalance(&_GnosisSafe.CallOpts)
}

// GetSelfBalance is a free data retrieval call binding the contract method 0x048a5fed.
//
// Solidity: function getSelfBalance() view returns(uint256)
func (_GnosisSafe *GnosisSafeCallerSession) GetSelfBalance() (*big.Int, error) {
	return _GnosisSafe.Contract.GetSelfBalance(&_GnosisSafe.CallOpts)
}

// GetStorageAt is a free data retrieval call binding the contract method 0x5624b25b.
//
// Solidity: function getStorageAt(uint256 offset, uint256 length) view returns(bytes)
func (_GnosisSafe *GnosisSafeCaller) GetStorageAt(opts *bind.CallOpts, offset *big.Int, length *big.Int) ([]byte, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "getStorageAt", offset, length)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetStorageAt is a free data retrieval call binding the contract method 0x5624b25b.
//
// Solidity: function getStorageAt(uint256 offset, uint256 length) view returns(bytes)
func (_GnosisSafe *GnosisSafeSession) GetStorageAt(offset *big.Int, length *big.Int) ([]byte, error) {
	return _GnosisSafe.Contract.GetStorageAt(&_GnosisSafe.CallOpts, offset, length)
}

// GetStorageAt is a free data retrieval call binding the contract method 0x5624b25b.
//
// Solidity: function getStorageAt(uint256 offset, uint256 length) view returns(bytes)
func (_GnosisSafe *GnosisSafeCallerSession) GetStorageAt(offset *big.Int, length *big.Int) ([]byte, error) {
	return _GnosisSafe.Contract.GetStorageAt(&_GnosisSafe.CallOpts, offset, length)
}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint256)
func (_GnosisSafe *GnosisSafeCaller) GetThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "getThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint256)
func (_GnosisSafe *GnosisSafeSession) GetThreshold() (*big.Int, error) {
	return _GnosisSafe.Contract.GetThreshold(&_GnosisSafe.CallOpts)
}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint256)
func (_GnosisSafe *GnosisSafeCallerSession) GetThreshold() (*big.Int, error) {
	return _GnosisSafe.Contract.GetThreshold(&_GnosisSafe.CallOpts)
}

// GetTransactionHash is a free data retrieval call binding the contract method 0xd8d11f78.
//
// Solidity: function getTransactionHash(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, uint256 _nonce) view returns(bytes32)
func (_GnosisSafe *GnosisSafeCaller) GetTransactionHash(opts *bind.CallOpts, to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, _nonce *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "getTransactionHash", to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, _nonce)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetTransactionHash is a free data retrieval call binding the contract method 0xd8d11f78.
//
// Solidity: function getTransactionHash(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, uint256 _nonce) view returns(bytes32)
func (_GnosisSafe *GnosisSafeSession) GetTransactionHash(to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, _nonce *big.Int) ([32]byte, error) {
	return _GnosisSafe.Contract.GetTransactionHash(&_GnosisSafe.CallOpts, to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, _nonce)
}

// GetTransactionHash is a free data retrieval call binding the contract method 0xd8d11f78.
//
// Solidity: function getTransactionHash(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, uint256 _nonce) view returns(bytes32)
func (_GnosisSafe *GnosisSafeCallerSession) GetTransactionHash(to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, _nonce *big.Int) ([32]byte, error) {
	return _GnosisSafe.Contract.GetTransactionHash(&_GnosisSafe.CallOpts, to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, _nonce)
}

// IsModuleEnabled is a free data retrieval call binding the contract method 0x2d9ad53d.
//
// Solidity: function isModuleEnabled(address module) view returns(bool)
func (_GnosisSafe *GnosisSafeCaller) IsModuleEnabled(opts *bind.CallOpts, module common.Address) (bool, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "isModuleEnabled", module)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsModuleEnabled is a free data retrieval call binding the contract method 0x2d9ad53d.
//
// Solidity: function isModuleEnabled(address module) view returns(bool)
func (_GnosisSafe *GnosisSafeSession) IsModuleEnabled(module common.Address) (bool, error) {
	return _GnosisSafe.Contract.IsModuleEnabled(&_GnosisSafe.CallOpts, module)
}

// IsModuleEnabled is a free data retrieval call binding the contract method 0x2d9ad53d.
//
// Solidity: function isModuleEnabled(address module) view returns(bool)
func (_GnosisSafe *GnosisSafeCallerSession) IsModuleEnabled(module common.Address) (bool, error) {
	return _GnosisSafe.Contract.IsModuleEnabled(&_GnosisSafe.CallOpts, module)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address owner) view returns(bool)
func (_GnosisSafe *GnosisSafeCaller) IsOwner(opts *bind.CallOpts, owner common.Address) (bool, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "isOwner", owner)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address owner) view returns(bool)
func (_GnosisSafe *GnosisSafeSession) IsOwner(owner common.Address) (bool, error) {
	return _GnosisSafe.Contract.IsOwner(&_GnosisSafe.CallOpts, owner)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address owner) view returns(bool)
func (_GnosisSafe *GnosisSafeCallerSession) IsOwner(owner common.Address) (bool, error) {
	return _GnosisSafe.Contract.IsOwner(&_GnosisSafe.CallOpts, owner)
}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_GnosisSafe *GnosisSafeCaller) Nonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "nonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_GnosisSafe *GnosisSafeSession) Nonce() (*big.Int, error) {
	return _GnosisSafe.Contract.Nonce(&_GnosisSafe.CallOpts)
}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_GnosisSafe *GnosisSafeCallerSession) Nonce() (*big.Int, error) {
	return _GnosisSafe.Contract.Nonce(&_GnosisSafe.CallOpts)
}

// SignedMessages is a free data retrieval call binding the contract method 0x5ae6bd37.
//
// Solidity: function signedMessages(bytes32 ) view returns(uint256)
func (_GnosisSafe *GnosisSafeCaller) SignedMessages(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _GnosisSafe.contract.Call(opts, &out, "signedMessages", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SignedMessages is a free data retrieval call binding the contract method 0x5ae6bd37.
//
// Solidity: function signedMessages(bytes32 ) view returns(uint256)
func (_GnosisSafe *GnosisSafeSession) SignedMessages(arg0 [32]byte) (*big.Int, error) {
	return _GnosisSafe.Contract.SignedMessages(&_GnosisSafe.CallOpts, arg0)
}

// SignedMessages is a free data retrieval call binding the contract method 0x5ae6bd37.
//
// Solidity: function signedMessages(bytes32 ) view returns(uint256)
func (_GnosisSafe *GnosisSafeCallerSession) SignedMessages(arg0 [32]byte) (*big.Int, error) {
	return _GnosisSafe.Contract.SignedMessages(&_GnosisSafe.CallOpts, arg0)
}

// AddOwnerWithThreshold is a paid mutator transaction binding the contract method 0x0d582f13.
//
// Solidity: function addOwnerWithThreshold(address owner, uint256 _threshold) returns()
func (_GnosisSafe *GnosisSafeTransactor) AddOwnerWithThreshold(opts *bind.TransactOpts, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "addOwnerWithThreshold", owner, _threshold)
}

// AddOwnerWithThreshold is a paid mutator transaction binding the contract method 0x0d582f13.
//
// Solidity: function addOwnerWithThreshold(address owner, uint256 _threshold) returns()
func (_GnosisSafe *GnosisSafeSession) AddOwnerWithThreshold(owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _GnosisSafe.Contract.AddOwnerWithThreshold(&_GnosisSafe.TransactOpts, owner, _threshold)
}

// AddOwnerWithThreshold is a paid mutator transaction binding the contract method 0x0d582f13.
//
// Solidity: function addOwnerWithThreshold(address owner, uint256 _threshold) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) AddOwnerWithThreshold(owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _GnosisSafe.Contract.AddOwnerWithThreshold(&_GnosisSafe.TransactOpts, owner, _threshold)
}

// ApproveHash is a paid mutator transaction binding the contract method 0xd4d9bdcd.
//
// Solidity: function approveHash(bytes32 hashToApprove) returns()
func (_GnosisSafe *GnosisSafeTransactor) ApproveHash(opts *bind.TransactOpts, hashToApprove [32]byte) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "approveHash", hashToApprove)
}

// ApproveHash is a paid mutator transaction binding the contract method 0xd4d9bdcd.
//
// Solidity: function approveHash(bytes32 hashToApprove) returns()
func (_GnosisSafe *GnosisSafeSession) ApproveHash(hashToApprove [32]byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.ApproveHash(&_GnosisSafe.TransactOpts, hashToApprove)
}

// ApproveHash is a paid mutator transaction binding the contract method 0xd4d9bdcd.
//
// Solidity: function approveHash(bytes32 hashToApprove) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) ApproveHash(hashToApprove [32]byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.ApproveHash(&_GnosisSafe.TransactOpts, hashToApprove)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 _threshold) returns()
func (_GnosisSafe *GnosisSafeTransactor) ChangeThreshold(opts *bind.TransactOpts, _threshold *big.Int) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "changeThreshold", _threshold)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 _threshold) returns()
func (_GnosisSafe *GnosisSafeSession) ChangeThreshold(_threshold *big.Int) (*types.Transaction, error) {
	return _GnosisSafe.Contract.ChangeThreshold(&_GnosisSafe.TransactOpts, _threshold)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 _threshold) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) ChangeThreshold(_threshold *big.Int) (*types.Transaction, error) {
	return _GnosisSafe.Contract.ChangeThreshold(&_GnosisSafe.TransactOpts, _threshold)
}

// CheckSignatures is a paid mutator transaction binding the contract method 0x934f3a11.
//
// Solidity: function checkSignatures(bytes32 dataHash, bytes data, bytes signatures) returns()
func (_GnosisSafe *GnosisSafeTransactor) CheckSignatures(opts *bind.TransactOpts, dataHash [32]byte, data []byte, signatures []byte) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "checkSignatures", dataHash, data, signatures)
}

// CheckSignatures is a paid mutator transaction binding the contract method 0x934f3a11.
//
// Solidity: function checkSignatures(bytes32 dataHash, bytes data, bytes signatures) returns()
func (_GnosisSafe *GnosisSafeSession) CheckSignatures(dataHash [32]byte, data []byte, signatures []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.CheckSignatures(&_GnosisSafe.TransactOpts, dataHash, data, signatures)
}

// CheckSignatures is a paid mutator transaction binding the contract method 0x934f3a11.
//
// Solidity: function checkSignatures(bytes32 dataHash, bytes data, bytes signatures) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) CheckSignatures(dataHash [32]byte, data []byte, signatures []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.CheckSignatures(&_GnosisSafe.TransactOpts, dataHash, data, signatures)
}

// DisableModule is a paid mutator transaction binding the contract method 0xe009cfde.
//
// Solidity: function disableModule(address prevModule, address module) returns()
func (_GnosisSafe *GnosisSafeTransactor) DisableModule(opts *bind.TransactOpts, prevModule common.Address, module common.Address) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "disableModule", prevModule, module)
}

// DisableModule is a paid mutator transaction binding the contract method 0xe009cfde.
//
// Solidity: function disableModule(address prevModule, address module) returns()
func (_GnosisSafe *GnosisSafeSession) DisableModule(prevModule common.Address, module common.Address) (*types.Transaction, error) {
	return _GnosisSafe.Contract.DisableModule(&_GnosisSafe.TransactOpts, prevModule, module)
}

// DisableModule is a paid mutator transaction binding the contract method 0xe009cfde.
//
// Solidity: function disableModule(address prevModule, address module) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) DisableModule(prevModule common.Address, module common.Address) (*types.Transaction, error) {
	return _GnosisSafe.Contract.DisableModule(&_GnosisSafe.TransactOpts, prevModule, module)
}

// EnableModule is a paid mutator transaction binding the contract method 0x610b5925.
//
// Solidity: function enableModule(address module) returns()
func (_GnosisSafe *GnosisSafeTransactor) EnableModule(opts *bind.TransactOpts, module common.Address) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "enableModule", module)
}

// EnableModule is a paid mutator transaction binding the contract method 0x610b5925.
//
// Solidity: function enableModule(address module) returns()
func (_GnosisSafe *GnosisSafeSession) EnableModule(module common.Address) (*types.Transaction, error) {
	return _GnosisSafe.Contract.EnableModule(&_GnosisSafe.TransactOpts, module)
}

// EnableModule is a paid mutator transaction binding the contract method 0x610b5925.
//
// Solidity: function enableModule(address module) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) EnableModule(module common.Address) (*types.Transaction, error) {
	return _GnosisSafe.Contract.EnableModule(&_GnosisSafe.TransactOpts, module)
}

// ExecTransaction is a paid mutator transaction binding the contract method 0x6a761202.
//
// Solidity: function execTransaction(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, bytes signatures) payable returns(bool success)
func (_GnosisSafe *GnosisSafeTransactor) ExecTransaction(opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, signatures []byte) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "execTransaction", to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, signatures)
}

// ExecTransaction is a paid mutator transaction binding the contract method 0x6a761202.
//
// Solidity: function execTransaction(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, bytes signatures) payable returns(bool success)
func (_GnosisSafe *GnosisSafeSession) ExecTransaction(to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, signatures []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.ExecTransaction(&_GnosisSafe.TransactOpts, to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, signatures)
}

// ExecTransaction is a paid mutator transaction binding the contract method 0x6a761202.
//
// Solidity: function execTransaction(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, bytes signatures) payable returns(bool success)
func (_GnosisSafe *GnosisSafeTransactorSession) ExecTransaction(to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, signatures []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.ExecTransaction(&_GnosisSafe.TransactOpts, to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, signatures)
}

// ExecTransactionFromModule is a paid mutator transaction binding the contract method 0x468721a7.
//
// Solidity: function execTransactionFromModule(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_GnosisSafe *GnosisSafeTransactor) ExecTransactionFromModule(opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "execTransactionFromModule", to, value, data, operation)
}

// ExecTransactionFromModule is a paid mutator transaction binding the contract method 0x468721a7.
//
// Solidity: function execTransactionFromModule(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_GnosisSafe *GnosisSafeSession) ExecTransactionFromModule(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _GnosisSafe.Contract.ExecTransactionFromModule(&_GnosisSafe.TransactOpts, to, value, data, operation)
}

// ExecTransactionFromModule is a paid mutator transaction binding the contract method 0x468721a7.
//
// Solidity: function execTransactionFromModule(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_GnosisSafe *GnosisSafeTransactorSession) ExecTransactionFromModule(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _GnosisSafe.Contract.ExecTransactionFromModule(&_GnosisSafe.TransactOpts, to, value, data, operation)
}

// ExecTransactionFromModuleReturnData is a paid mutator transaction binding the contract method 0x5229073f.
//
// Solidity: function execTransactionFromModuleReturnData(address to, uint256 value, bytes data, uint8 operation) returns(bool success, bytes returnData)
func (_GnosisSafe *GnosisSafeTransactor) ExecTransactionFromModuleReturnData(opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "execTransactionFromModuleReturnData", to, value, data, operation)
}

// ExecTransactionFromModuleReturnData is a paid mutator transaction binding the contract method 0x5229073f.
//
// Solidity: function execTransactionFromModuleReturnData(address to, uint256 value, bytes data, uint8 operation) returns(bool success, bytes returnData)
func (_GnosisSafe *GnosisSafeSession) ExecTransactionFromModuleReturnData(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _GnosisSafe.Contract.ExecTransactionFromModuleReturnData(&_GnosisSafe.TransactOpts, to, value, data, operation)
}

// ExecTransactionFromModuleReturnData is a paid mutator transaction binding the contract method 0x5229073f.
//
// Solidity: function execTransactionFromModuleReturnData(address to, uint256 value, bytes data, uint8 operation) returns(bool success, bytes returnData)
func (_GnosisSafe *GnosisSafeTransactorSession) ExecTransactionFromModuleReturnData(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _GnosisSafe.Contract.ExecTransactionFromModuleReturnData(&_GnosisSafe.TransactOpts, to, value, data, operation)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0xf8dc5dd9.
//
// Solidity: function removeOwner(address prevOwner, address owner, uint256 _threshold) returns()
func (_GnosisSafe *GnosisSafeTransactor) RemoveOwner(opts *bind.TransactOpts, prevOwner common.Address, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "removeOwner", prevOwner, owner, _threshold)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0xf8dc5dd9.
//
// Solidity: function removeOwner(address prevOwner, address owner, uint256 _threshold) returns()
func (_GnosisSafe *GnosisSafeSession) RemoveOwner(prevOwner common.Address, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _GnosisSafe.Contract.RemoveOwner(&_GnosisSafe.TransactOpts, prevOwner, owner, _threshold)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0xf8dc5dd9.
//
// Solidity: function removeOwner(address prevOwner, address owner, uint256 _threshold) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) RemoveOwner(prevOwner common.Address, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _GnosisSafe.Contract.RemoveOwner(&_GnosisSafe.TransactOpts, prevOwner, owner, _threshold)
}

// RequiredTxGas is a paid mutator transaction binding the contract method 0xc4ca3a9c.
//
// Solidity: function requiredTxGas(address to, uint256 value, bytes data, uint8 operation) returns(uint256)
func (_GnosisSafe *GnosisSafeTransactor) RequiredTxGas(opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "requiredTxGas", to, value, data, operation)
}

// RequiredTxGas is a paid mutator transaction binding the contract method 0xc4ca3a9c.
//
// Solidity: function requiredTxGas(address to, uint256 value, bytes data, uint8 operation) returns(uint256)
func (_GnosisSafe *GnosisSafeSession) RequiredTxGas(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _GnosisSafe.Contract.RequiredTxGas(&_GnosisSafe.TransactOpts, to, value, data, operation)
}

// RequiredTxGas is a paid mutator transaction binding the contract method 0xc4ca3a9c.
//
// Solidity: function requiredTxGas(address to, uint256 value, bytes data, uint8 operation) returns(uint256)
func (_GnosisSafe *GnosisSafeTransactorSession) RequiredTxGas(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _GnosisSafe.Contract.RequiredTxGas(&_GnosisSafe.TransactOpts, to, value, data, operation)
}

// SetFallbackHandler is a paid mutator transaction binding the contract method 0xf08a0323.
//
// Solidity: function setFallbackHandler(address handler) returns()
func (_GnosisSafe *GnosisSafeTransactor) SetFallbackHandler(opts *bind.TransactOpts, handler common.Address) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "setFallbackHandler", handler)
}

// SetFallbackHandler is a paid mutator transaction binding the contract method 0xf08a0323.
//
// Solidity: function setFallbackHandler(address handler) returns()
func (_GnosisSafe *GnosisSafeSession) SetFallbackHandler(handler common.Address) (*types.Transaction, error) {
	return _GnosisSafe.Contract.SetFallbackHandler(&_GnosisSafe.TransactOpts, handler)
}

// SetFallbackHandler is a paid mutator transaction binding the contract method 0xf08a0323.
//
// Solidity: function setFallbackHandler(address handler) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) SetFallbackHandler(handler common.Address) (*types.Transaction, error) {
	return _GnosisSafe.Contract.SetFallbackHandler(&_GnosisSafe.TransactOpts, handler)
}

// Setup is a paid mutator transaction binding the contract method 0xb63e800d.
//
// Solidity: function setup(address[] _owners, uint256 _threshold, address to, bytes data, address fallbackHandler, address paymentToken, uint256 payment, address paymentReceiver) returns()
func (_GnosisSafe *GnosisSafeTransactor) Setup(opts *bind.TransactOpts, _owners []common.Address, _threshold *big.Int, to common.Address, data []byte, fallbackHandler common.Address, paymentToken common.Address, payment *big.Int, paymentReceiver common.Address) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "setup", _owners, _threshold, to, data, fallbackHandler, paymentToken, payment, paymentReceiver)
}

// Setup is a paid mutator transaction binding the contract method 0xb63e800d.
//
// Solidity: function setup(address[] _owners, uint256 _threshold, address to, bytes data, address fallbackHandler, address paymentToken, uint256 payment, address paymentReceiver) returns()
func (_GnosisSafe *GnosisSafeSession) Setup(_owners []common.Address, _threshold *big.Int, to common.Address, data []byte, fallbackHandler common.Address, paymentToken common.Address, payment *big.Int, paymentReceiver common.Address) (*types.Transaction, error) {
	return _GnosisSafe.Contract.Setup(&_GnosisSafe.TransactOpts, _owners, _threshold, to, data, fallbackHandler, paymentToken, payment, paymentReceiver)
}

// Setup is a paid mutator transaction binding the contract method 0xb63e800d.
//
// Solidity: function setup(address[] _owners, uint256 _threshold, address to, bytes data, address fallbackHandler, address paymentToken, uint256 payment, address paymentReceiver) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) Setup(_owners []common.Address, _threshold *big.Int, to common.Address, data []byte, fallbackHandler common.Address, paymentToken common.Address, payment *big.Int, paymentReceiver common.Address) (*types.Transaction, error) {
	return _GnosisSafe.Contract.Setup(&_GnosisSafe.TransactOpts, _owners, _threshold, to, data, fallbackHandler, paymentToken, payment, paymentReceiver)
}

// SignMessage is a paid mutator transaction binding the contract method 0x85a5affe.
//
// Solidity: function signMessage(bytes _data) returns()
func (_GnosisSafe *GnosisSafeTransactor) SignMessage(opts *bind.TransactOpts, _data []byte) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "signMessage", _data)
}

// SignMessage is a paid mutator transaction binding the contract method 0x85a5affe.
//
// Solidity: function signMessage(bytes _data) returns()
func (_GnosisSafe *GnosisSafeSession) SignMessage(_data []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.SignMessage(&_GnosisSafe.TransactOpts, _data)
}

// SignMessage is a paid mutator transaction binding the contract method 0x85a5affe.
//
// Solidity: function signMessage(bytes _data) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) SignMessage(_data []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.SignMessage(&_GnosisSafe.TransactOpts, _data)
}

// SimulateDelegatecall is a paid mutator transaction binding the contract method 0xf84436bd.
//
// Solidity: function simulateDelegatecall(address targetContract, bytes calldataPayload) returns(bytes)
func (_GnosisSafe *GnosisSafeTransactor) SimulateDelegatecall(opts *bind.TransactOpts, targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "simulateDelegatecall", targetContract, calldataPayload)
}

// SimulateDelegatecall is a paid mutator transaction binding the contract method 0xf84436bd.
//
// Solidity: function simulateDelegatecall(address targetContract, bytes calldataPayload) returns(bytes)
func (_GnosisSafe *GnosisSafeSession) SimulateDelegatecall(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.SimulateDelegatecall(&_GnosisSafe.TransactOpts, targetContract, calldataPayload)
}

// SimulateDelegatecall is a paid mutator transaction binding the contract method 0xf84436bd.
//
// Solidity: function simulateDelegatecall(address targetContract, bytes calldataPayload) returns(bytes)
func (_GnosisSafe *GnosisSafeTransactorSession) SimulateDelegatecall(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.SimulateDelegatecall(&_GnosisSafe.TransactOpts, targetContract, calldataPayload)
}

// SimulateDelegatecallInternal is a paid mutator transaction binding the contract method 0x43218e19.
//
// Solidity: function simulateDelegatecallInternal(address targetContract, bytes calldataPayload) returns(bytes)
func (_GnosisSafe *GnosisSafeTransactor) SimulateDelegatecallInternal(opts *bind.TransactOpts, targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "simulateDelegatecallInternal", targetContract, calldataPayload)
}

// SimulateDelegatecallInternal is a paid mutator transaction binding the contract method 0x43218e19.
//
// Solidity: function simulateDelegatecallInternal(address targetContract, bytes calldataPayload) returns(bytes)
func (_GnosisSafe *GnosisSafeSession) SimulateDelegatecallInternal(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.SimulateDelegatecallInternal(&_GnosisSafe.TransactOpts, targetContract, calldataPayload)
}

// SimulateDelegatecallInternal is a paid mutator transaction binding the contract method 0x43218e19.
//
// Solidity: function simulateDelegatecallInternal(address targetContract, bytes calldataPayload) returns(bytes)
func (_GnosisSafe *GnosisSafeTransactorSession) SimulateDelegatecallInternal(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.SimulateDelegatecallInternal(&_GnosisSafe.TransactOpts, targetContract, calldataPayload)
}

// SwapOwner is a paid mutator transaction binding the contract method 0xe318b52b.
//
// Solidity: function swapOwner(address prevOwner, address oldOwner, address newOwner) returns()
func (_GnosisSafe *GnosisSafeTransactor) SwapOwner(opts *bind.TransactOpts, prevOwner common.Address, oldOwner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _GnosisSafe.contract.Transact(opts, "swapOwner", prevOwner, oldOwner, newOwner)
}

// SwapOwner is a paid mutator transaction binding the contract method 0xe318b52b.
//
// Solidity: function swapOwner(address prevOwner, address oldOwner, address newOwner) returns()
func (_GnosisSafe *GnosisSafeSession) SwapOwner(prevOwner common.Address, oldOwner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _GnosisSafe.Contract.SwapOwner(&_GnosisSafe.TransactOpts, prevOwner, oldOwner, newOwner)
}

// SwapOwner is a paid mutator transaction binding the contract method 0xe318b52b.
//
// Solidity: function swapOwner(address prevOwner, address oldOwner, address newOwner) returns()
func (_GnosisSafe *GnosisSafeTransactorSession) SwapOwner(prevOwner common.Address, oldOwner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _GnosisSafe.Contract.SwapOwner(&_GnosisSafe.TransactOpts, prevOwner, oldOwner, newOwner)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_GnosisSafe *GnosisSafeTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _GnosisSafe.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_GnosisSafe *GnosisSafeSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.Fallback(&_GnosisSafe.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_GnosisSafe *GnosisSafeTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _GnosisSafe.Contract.Fallback(&_GnosisSafe.TransactOpts, calldata)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_GnosisSafe *GnosisSafeTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GnosisSafe.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_GnosisSafe *GnosisSafeSession) Receive() (*types.Transaction, error) {
	return _GnosisSafe.Contract.Receive(&_GnosisSafe.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_GnosisSafe *GnosisSafeTransactorSession) Receive() (*types.Transaction, error) {
	return _GnosisSafe.Contract.Receive(&_GnosisSafe.TransactOpts)
}

// GnosisSafeAddedOwnerIterator is returned from FilterAddedOwner and is used to iterate over the raw logs and unpacked data for AddedOwner events raised by the GnosisSafe contract.
type GnosisSafeAddedOwnerIterator struct {
	Event *GnosisSafeAddedOwner // Event containing the contract specifics and raw log

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
func (it *GnosisSafeAddedOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeAddedOwner)
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
		it.Event = new(GnosisSafeAddedOwner)
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
func (it *GnosisSafeAddedOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeAddedOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeAddedOwner represents a AddedOwner event raised by the GnosisSafe contract.
type GnosisSafeAddedOwner struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterAddedOwner is a free log retrieval operation binding the contract event 0x9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea26.
//
// Solidity: event AddedOwner(address owner)
func (_GnosisSafe *GnosisSafeFilterer) FilterAddedOwner(opts *bind.FilterOpts) (*GnosisSafeAddedOwnerIterator, error) {

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "AddedOwner")
	if err != nil {
		return nil, err
	}
	return &GnosisSafeAddedOwnerIterator{contract: _GnosisSafe.contract, event: "AddedOwner", logs: logs, sub: sub}, nil
}

// WatchAddedOwner is a free log subscription operation binding the contract event 0x9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea26.
//
// Solidity: event AddedOwner(address owner)
func (_GnosisSafe *GnosisSafeFilterer) WatchAddedOwner(opts *bind.WatchOpts, sink chan<- *GnosisSafeAddedOwner) (event.Subscription, error) {

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "AddedOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeAddedOwner)
				if err := _GnosisSafe.contract.UnpackLog(event, "AddedOwner", log); err != nil {
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

// ParseAddedOwner is a log parse operation binding the contract event 0x9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea26.
//
// Solidity: event AddedOwner(address owner)
func (_GnosisSafe *GnosisSafeFilterer) ParseAddedOwner(log types.Log) (*GnosisSafeAddedOwner, error) {
	event := new(GnosisSafeAddedOwner)
	if err := _GnosisSafe.contract.UnpackLog(event, "AddedOwner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeApproveHashIterator is returned from FilterApproveHash and is used to iterate over the raw logs and unpacked data for ApproveHash events raised by the GnosisSafe contract.
type GnosisSafeApproveHashIterator struct {
	Event *GnosisSafeApproveHash // Event containing the contract specifics and raw log

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
func (it *GnosisSafeApproveHashIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeApproveHash)
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
		it.Event = new(GnosisSafeApproveHash)
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
func (it *GnosisSafeApproveHashIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeApproveHashIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeApproveHash represents a ApproveHash event raised by the GnosisSafe contract.
type GnosisSafeApproveHash struct {
	ApprovedHash [32]byte
	Owner        common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterApproveHash is a free log retrieval operation binding the contract event 0xf2a0eb156472d1440255b0d7c1e19cc07115d1051fe605b0dce69acfec884d9c.
//
// Solidity: event ApproveHash(bytes32 indexed approvedHash, address indexed owner)
func (_GnosisSafe *GnosisSafeFilterer) FilterApproveHash(opts *bind.FilterOpts, approvedHash [][32]byte, owner []common.Address) (*GnosisSafeApproveHashIterator, error) {

	var approvedHashRule []interface{}
	for _, approvedHashItem := range approvedHash {
		approvedHashRule = append(approvedHashRule, approvedHashItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "ApproveHash", approvedHashRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeApproveHashIterator{contract: _GnosisSafe.contract, event: "ApproveHash", logs: logs, sub: sub}, nil
}

// WatchApproveHash is a free log subscription operation binding the contract event 0xf2a0eb156472d1440255b0d7c1e19cc07115d1051fe605b0dce69acfec884d9c.
//
// Solidity: event ApproveHash(bytes32 indexed approvedHash, address indexed owner)
func (_GnosisSafe *GnosisSafeFilterer) WatchApproveHash(opts *bind.WatchOpts, sink chan<- *GnosisSafeApproveHash, approvedHash [][32]byte, owner []common.Address) (event.Subscription, error) {

	var approvedHashRule []interface{}
	for _, approvedHashItem := range approvedHash {
		approvedHashRule = append(approvedHashRule, approvedHashItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "ApproveHash", approvedHashRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeApproveHash)
				if err := _GnosisSafe.contract.UnpackLog(event, "ApproveHash", log); err != nil {
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

// ParseApproveHash is a log parse operation binding the contract event 0xf2a0eb156472d1440255b0d7c1e19cc07115d1051fe605b0dce69acfec884d9c.
//
// Solidity: event ApproveHash(bytes32 indexed approvedHash, address indexed owner)
func (_GnosisSafe *GnosisSafeFilterer) ParseApproveHash(log types.Log) (*GnosisSafeApproveHash, error) {
	event := new(GnosisSafeApproveHash)
	if err := _GnosisSafe.contract.UnpackLog(event, "ApproveHash", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeChangedThresholdIterator is returned from FilterChangedThreshold and is used to iterate over the raw logs and unpacked data for ChangedThreshold events raised by the GnosisSafe contract.
type GnosisSafeChangedThresholdIterator struct {
	Event *GnosisSafeChangedThreshold // Event containing the contract specifics and raw log

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
func (it *GnosisSafeChangedThresholdIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeChangedThreshold)
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
		it.Event = new(GnosisSafeChangedThreshold)
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
func (it *GnosisSafeChangedThresholdIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeChangedThresholdIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeChangedThreshold represents a ChangedThreshold event raised by the GnosisSafe contract.
type GnosisSafeChangedThreshold struct {
	Threshold *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterChangedThreshold is a free log retrieval operation binding the contract event 0x610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c93.
//
// Solidity: event ChangedThreshold(uint256 threshold)
func (_GnosisSafe *GnosisSafeFilterer) FilterChangedThreshold(opts *bind.FilterOpts) (*GnosisSafeChangedThresholdIterator, error) {

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "ChangedThreshold")
	if err != nil {
		return nil, err
	}
	return &GnosisSafeChangedThresholdIterator{contract: _GnosisSafe.contract, event: "ChangedThreshold", logs: logs, sub: sub}, nil
}

// WatchChangedThreshold is a free log subscription operation binding the contract event 0x610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c93.
//
// Solidity: event ChangedThreshold(uint256 threshold)
func (_GnosisSafe *GnosisSafeFilterer) WatchChangedThreshold(opts *bind.WatchOpts, sink chan<- *GnosisSafeChangedThreshold) (event.Subscription, error) {

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "ChangedThreshold")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeChangedThreshold)
				if err := _GnosisSafe.contract.UnpackLog(event, "ChangedThreshold", log); err != nil {
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

// ParseChangedThreshold is a log parse operation binding the contract event 0x610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c93.
//
// Solidity: event ChangedThreshold(uint256 threshold)
func (_GnosisSafe *GnosisSafeFilterer) ParseChangedThreshold(log types.Log) (*GnosisSafeChangedThreshold, error) {
	event := new(GnosisSafeChangedThreshold)
	if err := _GnosisSafe.contract.UnpackLog(event, "ChangedThreshold", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeDisabledModuleIterator is returned from FilterDisabledModule and is used to iterate over the raw logs and unpacked data for DisabledModule events raised by the GnosisSafe contract.
type GnosisSafeDisabledModuleIterator struct {
	Event *GnosisSafeDisabledModule // Event containing the contract specifics and raw log

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
func (it *GnosisSafeDisabledModuleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeDisabledModule)
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
		it.Event = new(GnosisSafeDisabledModule)
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
func (it *GnosisSafeDisabledModuleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeDisabledModuleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeDisabledModule represents a DisabledModule event raised by the GnosisSafe contract.
type GnosisSafeDisabledModule struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDisabledModule is a free log retrieval operation binding the contract event 0xaab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276.
//
// Solidity: event DisabledModule(address module)
func (_GnosisSafe *GnosisSafeFilterer) FilterDisabledModule(opts *bind.FilterOpts) (*GnosisSafeDisabledModuleIterator, error) {

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "DisabledModule")
	if err != nil {
		return nil, err
	}
	return &GnosisSafeDisabledModuleIterator{contract: _GnosisSafe.contract, event: "DisabledModule", logs: logs, sub: sub}, nil
}

// WatchDisabledModule is a free log subscription operation binding the contract event 0xaab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276.
//
// Solidity: event DisabledModule(address module)
func (_GnosisSafe *GnosisSafeFilterer) WatchDisabledModule(opts *bind.WatchOpts, sink chan<- *GnosisSafeDisabledModule) (event.Subscription, error) {

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "DisabledModule")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeDisabledModule)
				if err := _GnosisSafe.contract.UnpackLog(event, "DisabledModule", log); err != nil {
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

// ParseDisabledModule is a log parse operation binding the contract event 0xaab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276.
//
// Solidity: event DisabledModule(address module)
func (_GnosisSafe *GnosisSafeFilterer) ParseDisabledModule(log types.Log) (*GnosisSafeDisabledModule, error) {
	event := new(GnosisSafeDisabledModule)
	if err := _GnosisSafe.contract.UnpackLog(event, "DisabledModule", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeEnabledModuleIterator is returned from FilterEnabledModule and is used to iterate over the raw logs and unpacked data for EnabledModule events raised by the GnosisSafe contract.
type GnosisSafeEnabledModuleIterator struct {
	Event *GnosisSafeEnabledModule // Event containing the contract specifics and raw log

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
func (it *GnosisSafeEnabledModuleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeEnabledModule)
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
		it.Event = new(GnosisSafeEnabledModule)
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
func (it *GnosisSafeEnabledModuleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeEnabledModuleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeEnabledModule represents a EnabledModule event raised by the GnosisSafe contract.
type GnosisSafeEnabledModule struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterEnabledModule is a free log retrieval operation binding the contract event 0xecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f8440.
//
// Solidity: event EnabledModule(address module)
func (_GnosisSafe *GnosisSafeFilterer) FilterEnabledModule(opts *bind.FilterOpts) (*GnosisSafeEnabledModuleIterator, error) {

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "EnabledModule")
	if err != nil {
		return nil, err
	}
	return &GnosisSafeEnabledModuleIterator{contract: _GnosisSafe.contract, event: "EnabledModule", logs: logs, sub: sub}, nil
}

// WatchEnabledModule is a free log subscription operation binding the contract event 0xecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f8440.
//
// Solidity: event EnabledModule(address module)
func (_GnosisSafe *GnosisSafeFilterer) WatchEnabledModule(opts *bind.WatchOpts, sink chan<- *GnosisSafeEnabledModule) (event.Subscription, error) {

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "EnabledModule")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeEnabledModule)
				if err := _GnosisSafe.contract.UnpackLog(event, "EnabledModule", log); err != nil {
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

// ParseEnabledModule is a log parse operation binding the contract event 0xecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f8440.
//
// Solidity: event EnabledModule(address module)
func (_GnosisSafe *GnosisSafeFilterer) ParseEnabledModule(log types.Log) (*GnosisSafeEnabledModule, error) {
	event := new(GnosisSafeEnabledModule)
	if err := _GnosisSafe.contract.UnpackLog(event, "EnabledModule", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeExecutionFailureIterator is returned from FilterExecutionFailure and is used to iterate over the raw logs and unpacked data for ExecutionFailure events raised by the GnosisSafe contract.
type GnosisSafeExecutionFailureIterator struct {
	Event *GnosisSafeExecutionFailure // Event containing the contract specifics and raw log

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
func (it *GnosisSafeExecutionFailureIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeExecutionFailure)
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
		it.Event = new(GnosisSafeExecutionFailure)
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
func (it *GnosisSafeExecutionFailureIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeExecutionFailureIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeExecutionFailure represents a ExecutionFailure event raised by the GnosisSafe contract.
type GnosisSafeExecutionFailure struct {
	TxHash  [32]byte
	Payment *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterExecutionFailure is a free log retrieval operation binding the contract event 0x23428b18acfb3ea64b08dc0c1d296ea9c09702c09083ca5272e64d115b687d23.
//
// Solidity: event ExecutionFailure(bytes32 txHash, uint256 payment)
func (_GnosisSafe *GnosisSafeFilterer) FilterExecutionFailure(opts *bind.FilterOpts) (*GnosisSafeExecutionFailureIterator, error) {

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "ExecutionFailure")
	if err != nil {
		return nil, err
	}
	return &GnosisSafeExecutionFailureIterator{contract: _GnosisSafe.contract, event: "ExecutionFailure", logs: logs, sub: sub}, nil
}

// WatchExecutionFailure is a free log subscription operation binding the contract event 0x23428b18acfb3ea64b08dc0c1d296ea9c09702c09083ca5272e64d115b687d23.
//
// Solidity: event ExecutionFailure(bytes32 txHash, uint256 payment)
func (_GnosisSafe *GnosisSafeFilterer) WatchExecutionFailure(opts *bind.WatchOpts, sink chan<- *GnosisSafeExecutionFailure) (event.Subscription, error) {

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "ExecutionFailure")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeExecutionFailure)
				if err := _GnosisSafe.contract.UnpackLog(event, "ExecutionFailure", log); err != nil {
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

// ParseExecutionFailure is a log parse operation binding the contract event 0x23428b18acfb3ea64b08dc0c1d296ea9c09702c09083ca5272e64d115b687d23.
//
// Solidity: event ExecutionFailure(bytes32 txHash, uint256 payment)
func (_GnosisSafe *GnosisSafeFilterer) ParseExecutionFailure(log types.Log) (*GnosisSafeExecutionFailure, error) {
	event := new(GnosisSafeExecutionFailure)
	if err := _GnosisSafe.contract.UnpackLog(event, "ExecutionFailure", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeExecutionFromModuleFailureIterator is returned from FilterExecutionFromModuleFailure and is used to iterate over the raw logs and unpacked data for ExecutionFromModuleFailure events raised by the GnosisSafe contract.
type GnosisSafeExecutionFromModuleFailureIterator struct {
	Event *GnosisSafeExecutionFromModuleFailure // Event containing the contract specifics and raw log

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
func (it *GnosisSafeExecutionFromModuleFailureIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeExecutionFromModuleFailure)
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
		it.Event = new(GnosisSafeExecutionFromModuleFailure)
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
func (it *GnosisSafeExecutionFromModuleFailureIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeExecutionFromModuleFailureIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeExecutionFromModuleFailure represents a ExecutionFromModuleFailure event raised by the GnosisSafe contract.
type GnosisSafeExecutionFromModuleFailure struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterExecutionFromModuleFailure is a free log retrieval operation binding the contract event 0xacd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd375.
//
// Solidity: event ExecutionFromModuleFailure(address indexed module)
func (_GnosisSafe *GnosisSafeFilterer) FilterExecutionFromModuleFailure(opts *bind.FilterOpts, module []common.Address) (*GnosisSafeExecutionFromModuleFailureIterator, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "ExecutionFromModuleFailure", moduleRule)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeExecutionFromModuleFailureIterator{contract: _GnosisSafe.contract, event: "ExecutionFromModuleFailure", logs: logs, sub: sub}, nil
}

// WatchExecutionFromModuleFailure is a free log subscription operation binding the contract event 0xacd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd375.
//
// Solidity: event ExecutionFromModuleFailure(address indexed module)
func (_GnosisSafe *GnosisSafeFilterer) WatchExecutionFromModuleFailure(opts *bind.WatchOpts, sink chan<- *GnosisSafeExecutionFromModuleFailure, module []common.Address) (event.Subscription, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "ExecutionFromModuleFailure", moduleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeExecutionFromModuleFailure)
				if err := _GnosisSafe.contract.UnpackLog(event, "ExecutionFromModuleFailure", log); err != nil {
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

// ParseExecutionFromModuleFailure is a log parse operation binding the contract event 0xacd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd375.
//
// Solidity: event ExecutionFromModuleFailure(address indexed module)
func (_GnosisSafe *GnosisSafeFilterer) ParseExecutionFromModuleFailure(log types.Log) (*GnosisSafeExecutionFromModuleFailure, error) {
	event := new(GnosisSafeExecutionFromModuleFailure)
	if err := _GnosisSafe.contract.UnpackLog(event, "ExecutionFromModuleFailure", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeExecutionFromModuleSuccessIterator is returned from FilterExecutionFromModuleSuccess and is used to iterate over the raw logs and unpacked data for ExecutionFromModuleSuccess events raised by the GnosisSafe contract.
type GnosisSafeExecutionFromModuleSuccessIterator struct {
	Event *GnosisSafeExecutionFromModuleSuccess // Event containing the contract specifics and raw log

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
func (it *GnosisSafeExecutionFromModuleSuccessIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeExecutionFromModuleSuccess)
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
		it.Event = new(GnosisSafeExecutionFromModuleSuccess)
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
func (it *GnosisSafeExecutionFromModuleSuccessIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeExecutionFromModuleSuccessIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeExecutionFromModuleSuccess represents a ExecutionFromModuleSuccess event raised by the GnosisSafe contract.
type GnosisSafeExecutionFromModuleSuccess struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterExecutionFromModuleSuccess is a free log retrieval operation binding the contract event 0x6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb8.
//
// Solidity: event ExecutionFromModuleSuccess(address indexed module)
func (_GnosisSafe *GnosisSafeFilterer) FilterExecutionFromModuleSuccess(opts *bind.FilterOpts, module []common.Address) (*GnosisSafeExecutionFromModuleSuccessIterator, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "ExecutionFromModuleSuccess", moduleRule)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeExecutionFromModuleSuccessIterator{contract: _GnosisSafe.contract, event: "ExecutionFromModuleSuccess", logs: logs, sub: sub}, nil
}

// WatchExecutionFromModuleSuccess is a free log subscription operation binding the contract event 0x6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb8.
//
// Solidity: event ExecutionFromModuleSuccess(address indexed module)
func (_GnosisSafe *GnosisSafeFilterer) WatchExecutionFromModuleSuccess(opts *bind.WatchOpts, sink chan<- *GnosisSafeExecutionFromModuleSuccess, module []common.Address) (event.Subscription, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "ExecutionFromModuleSuccess", moduleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeExecutionFromModuleSuccess)
				if err := _GnosisSafe.contract.UnpackLog(event, "ExecutionFromModuleSuccess", log); err != nil {
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

// ParseExecutionFromModuleSuccess is a log parse operation binding the contract event 0x6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb8.
//
// Solidity: event ExecutionFromModuleSuccess(address indexed module)
func (_GnosisSafe *GnosisSafeFilterer) ParseExecutionFromModuleSuccess(log types.Log) (*GnosisSafeExecutionFromModuleSuccess, error) {
	event := new(GnosisSafeExecutionFromModuleSuccess)
	if err := _GnosisSafe.contract.UnpackLog(event, "ExecutionFromModuleSuccess", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeExecutionResultIterator is returned from FilterExecutionResult and is used to iterate over the raw logs and unpacked data for ExecutionResult events raised by the GnosisSafe contract.
type GnosisSafeExecutionResultIterator struct {
	Event *GnosisSafeExecutionResult // Event containing the contract specifics and raw log

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
func (it *GnosisSafeExecutionResultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeExecutionResult)
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
		it.Event = new(GnosisSafeExecutionResult)
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
func (it *GnosisSafeExecutionResultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeExecutionResultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeExecutionResult represents a ExecutionResult event raised by the GnosisSafe contract.
type GnosisSafeExecutionResult struct {
	Name   string
	Result bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterExecutionResult is a free log retrieval operation binding the contract event 0x36bd3cb3e572bed2e31aa120b605e9d3cb596f0703790070410c5f0b0ac5e34e.
//
// Solidity: event ExecutionResult(string name, bool result)
func (_GnosisSafe *GnosisSafeFilterer) FilterExecutionResult(opts *bind.FilterOpts) (*GnosisSafeExecutionResultIterator, error) {

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "ExecutionResult")
	if err != nil {
		return nil, err
	}
	return &GnosisSafeExecutionResultIterator{contract: _GnosisSafe.contract, event: "ExecutionResult", logs: logs, sub: sub}, nil
}

// WatchExecutionResult is a free log subscription operation binding the contract event 0x36bd3cb3e572bed2e31aa120b605e9d3cb596f0703790070410c5f0b0ac5e34e.
//
// Solidity: event ExecutionResult(string name, bool result)
func (_GnosisSafe *GnosisSafeFilterer) WatchExecutionResult(opts *bind.WatchOpts, sink chan<- *GnosisSafeExecutionResult) (event.Subscription, error) {

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "ExecutionResult")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeExecutionResult)
				if err := _GnosisSafe.contract.UnpackLog(event, "ExecutionResult", log); err != nil {
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

// ParseExecutionResult is a log parse operation binding the contract event 0x36bd3cb3e572bed2e31aa120b605e9d3cb596f0703790070410c5f0b0ac5e34e.
//
// Solidity: event ExecutionResult(string name, bool result)
func (_GnosisSafe *GnosisSafeFilterer) ParseExecutionResult(log types.Log) (*GnosisSafeExecutionResult, error) {
	event := new(GnosisSafeExecutionResult)
	if err := _GnosisSafe.contract.UnpackLog(event, "ExecutionResult", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeExecutionSuccessIterator is returned from FilterExecutionSuccess and is used to iterate over the raw logs and unpacked data for ExecutionSuccess events raised by the GnosisSafe contract.
type GnosisSafeExecutionSuccessIterator struct {
	Event *GnosisSafeExecutionSuccess // Event containing the contract specifics and raw log

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
func (it *GnosisSafeExecutionSuccessIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeExecutionSuccess)
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
		it.Event = new(GnosisSafeExecutionSuccess)
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
func (it *GnosisSafeExecutionSuccessIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeExecutionSuccessIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeExecutionSuccess represents a ExecutionSuccess event raised by the GnosisSafe contract.
type GnosisSafeExecutionSuccess struct {
	TxHash  [32]byte
	Payment *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterExecutionSuccess is a free log retrieval operation binding the contract event 0x442e715f626346e8c54381002da614f62bee8d27386535b2521ec8540898556e.
//
// Solidity: event ExecutionSuccess(bytes32 txHash, uint256 payment)
func (_GnosisSafe *GnosisSafeFilterer) FilterExecutionSuccess(opts *bind.FilterOpts) (*GnosisSafeExecutionSuccessIterator, error) {

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "ExecutionSuccess")
	if err != nil {
		return nil, err
	}
	return &GnosisSafeExecutionSuccessIterator{contract: _GnosisSafe.contract, event: "ExecutionSuccess", logs: logs, sub: sub}, nil
}

// WatchExecutionSuccess is a free log subscription operation binding the contract event 0x442e715f626346e8c54381002da614f62bee8d27386535b2521ec8540898556e.
//
// Solidity: event ExecutionSuccess(bytes32 txHash, uint256 payment)
func (_GnosisSafe *GnosisSafeFilterer) WatchExecutionSuccess(opts *bind.WatchOpts, sink chan<- *GnosisSafeExecutionSuccess) (event.Subscription, error) {

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "ExecutionSuccess")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeExecutionSuccess)
				if err := _GnosisSafe.contract.UnpackLog(event, "ExecutionSuccess", log); err != nil {
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

// ParseExecutionSuccess is a log parse operation binding the contract event 0x442e715f626346e8c54381002da614f62bee8d27386535b2521ec8540898556e.
//
// Solidity: event ExecutionSuccess(bytes32 txHash, uint256 payment)
func (_GnosisSafe *GnosisSafeFilterer) ParseExecutionSuccess(log types.Log) (*GnosisSafeExecutionSuccess, error) {
	event := new(GnosisSafeExecutionSuccess)
	if err := _GnosisSafe.contract.UnpackLog(event, "ExecutionSuccess", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeRemovedOwnerIterator is returned from FilterRemovedOwner and is used to iterate over the raw logs and unpacked data for RemovedOwner events raised by the GnosisSafe contract.
type GnosisSafeRemovedOwnerIterator struct {
	Event *GnosisSafeRemovedOwner // Event containing the contract specifics and raw log

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
func (it *GnosisSafeRemovedOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeRemovedOwner)
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
		it.Event = new(GnosisSafeRemovedOwner)
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
func (it *GnosisSafeRemovedOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeRemovedOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeRemovedOwner represents a RemovedOwner event raised by the GnosisSafe contract.
type GnosisSafeRemovedOwner struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRemovedOwner is a free log retrieval operation binding the contract event 0xf8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf.
//
// Solidity: event RemovedOwner(address owner)
func (_GnosisSafe *GnosisSafeFilterer) FilterRemovedOwner(opts *bind.FilterOpts) (*GnosisSafeRemovedOwnerIterator, error) {

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "RemovedOwner")
	if err != nil {
		return nil, err
	}
	return &GnosisSafeRemovedOwnerIterator{contract: _GnosisSafe.contract, event: "RemovedOwner", logs: logs, sub: sub}, nil
}

// WatchRemovedOwner is a free log subscription operation binding the contract event 0xf8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf.
//
// Solidity: event RemovedOwner(address owner)
func (_GnosisSafe *GnosisSafeFilterer) WatchRemovedOwner(opts *bind.WatchOpts, sink chan<- *GnosisSafeRemovedOwner) (event.Subscription, error) {

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "RemovedOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeRemovedOwner)
				if err := _GnosisSafe.contract.UnpackLog(event, "RemovedOwner", log); err != nil {
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

// ParseRemovedOwner is a log parse operation binding the contract event 0xf8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf.
//
// Solidity: event RemovedOwner(address owner)
func (_GnosisSafe *GnosisSafeFilterer) ParseRemovedOwner(log types.Log) (*GnosisSafeRemovedOwner, error) {
	event := new(GnosisSafeRemovedOwner)
	if err := _GnosisSafe.contract.UnpackLog(event, "RemovedOwner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeSignMsgIterator is returned from FilterSignMsg and is used to iterate over the raw logs and unpacked data for SignMsg events raised by the GnosisSafe contract.
type GnosisSafeSignMsgIterator struct {
	Event *GnosisSafeSignMsg // Event containing the contract specifics and raw log

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
func (it *GnosisSafeSignMsgIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeSignMsg)
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
		it.Event = new(GnosisSafeSignMsg)
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
func (it *GnosisSafeSignMsgIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeSignMsgIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeSignMsg represents a SignMsg event raised by the GnosisSafe contract.
type GnosisSafeSignMsg struct {
	MsgHash [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterSignMsg is a free log retrieval operation binding the contract event 0xe7f4675038f4f6034dfcbbb24c4dc08e4ebf10eb9d257d3d02c0f38d122ac6e4.
//
// Solidity: event SignMsg(bytes32 indexed msgHash)
func (_GnosisSafe *GnosisSafeFilterer) FilterSignMsg(opts *bind.FilterOpts, msgHash [][32]byte) (*GnosisSafeSignMsgIterator, error) {

	var msgHashRule []interface{}
	for _, msgHashItem := range msgHash {
		msgHashRule = append(msgHashRule, msgHashItem)
	}

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "SignMsg", msgHashRule)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeSignMsgIterator{contract: _GnosisSafe.contract, event: "SignMsg", logs: logs, sub: sub}, nil
}

// WatchSignMsg is a free log subscription operation binding the contract event 0xe7f4675038f4f6034dfcbbb24c4dc08e4ebf10eb9d257d3d02c0f38d122ac6e4.
//
// Solidity: event SignMsg(bytes32 indexed msgHash)
func (_GnosisSafe *GnosisSafeFilterer) WatchSignMsg(opts *bind.WatchOpts, sink chan<- *GnosisSafeSignMsg, msgHash [][32]byte) (event.Subscription, error) {

	var msgHashRule []interface{}
	for _, msgHashItem := range msgHash {
		msgHashRule = append(msgHashRule, msgHashItem)
	}

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "SignMsg", msgHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeSignMsg)
				if err := _GnosisSafe.contract.UnpackLog(event, "SignMsg", log); err != nil {
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

// ParseSignMsg is a log parse operation binding the contract event 0xe7f4675038f4f6034dfcbbb24c4dc08e4ebf10eb9d257d3d02c0f38d122ac6e4.
//
// Solidity: event SignMsg(bytes32 indexed msgHash)
func (_GnosisSafe *GnosisSafeFilterer) ParseSignMsg(log types.Log) (*GnosisSafeSignMsg, error) {
	event := new(GnosisSafeSignMsg)
	if err := _GnosisSafe.contract.UnpackLog(event, "SignMsg", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeSignatureRecoverIterator is returned from FilterSignatureRecover and is used to iterate over the raw logs and unpacked data for SignatureRecover events raised by the GnosisSafe contract.
type GnosisSafeSignatureRecoverIterator struct {
	Event *GnosisSafeSignatureRecover // Event containing the contract specifics and raw log

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
func (it *GnosisSafeSignatureRecoverIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GnosisSafeSignatureRecover)
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
		it.Event = new(GnosisSafeSignatureRecover)
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
func (it *GnosisSafeSignatureRecoverIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GnosisSafeSignatureRecoverIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GnosisSafeSignatureRecover represents a SignatureRecover event raised by the GnosisSafe contract.
type GnosisSafeSignatureRecover struct {
	Index *big.Int
	Owner common.Address
	Hash  [32]byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSignatureRecover is a free log retrieval operation binding the contract event 0x5a06eb92afc8e6aed01ca739e4a5f3d7b76155d973f4428c2f7316095d98abd3.
//
// Solidity: event SignatureRecover(uint256 index, address owner, bytes32 hash)
func (_GnosisSafe *GnosisSafeFilterer) FilterSignatureRecover(opts *bind.FilterOpts) (*GnosisSafeSignatureRecoverIterator, error) {

	logs, sub, err := _GnosisSafe.contract.FilterLogs(opts, "SignatureRecover")
	if err != nil {
		return nil, err
	}
	return &GnosisSafeSignatureRecoverIterator{contract: _GnosisSafe.contract, event: "SignatureRecover", logs: logs, sub: sub}, nil
}

// WatchSignatureRecover is a free log subscription operation binding the contract event 0x5a06eb92afc8e6aed01ca739e4a5f3d7b76155d973f4428c2f7316095d98abd3.
//
// Solidity: event SignatureRecover(uint256 index, address owner, bytes32 hash)
func (_GnosisSafe *GnosisSafeFilterer) WatchSignatureRecover(opts *bind.WatchOpts, sink chan<- *GnosisSafeSignatureRecover) (event.Subscription, error) {

	logs, sub, err := _GnosisSafe.contract.WatchLogs(opts, "SignatureRecover")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GnosisSafeSignatureRecover)
				if err := _GnosisSafe.contract.UnpackLog(event, "SignatureRecover", log); err != nil {
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

// ParseSignatureRecover is a log parse operation binding the contract event 0x5a06eb92afc8e6aed01ca739e4a5f3d7b76155d973f4428c2f7316095d98abd3.
//
// Solidity: event SignatureRecover(uint256 index, address owner, bytes32 hash)
func (_GnosisSafe *GnosisSafeFilterer) ParseSignatureRecover(log types.Log) (*GnosisSafeSignatureRecover, error) {
	event := new(GnosisSafeSignatureRecover)
	if err := _GnosisSafe.contract.UnpackLog(event, "SignatureRecover", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GnosisSafeMathABI is the input ABI used to generate the binding from.
const GnosisSafeMathABI = "[]"

// GnosisSafeMathBin is the compiled bytecode used for deploying new contracts.
var GnosisSafeMathBin = "0x60566023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea264697066735822122027e9a0611330dd98ea1f853f5351f15f5b01a2ba2188d51eb63539cde54b53f964736f6c63430007000033"

// DeployGnosisSafeMath deploys a new Ethereum contract, binding an instance of GnosisSafeMath to it.
func DeployGnosisSafeMath(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GnosisSafeMath, error) {
	parsed, err := abi.JSON(strings.NewReader(GnosisSafeMathABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(GnosisSafeMathBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GnosisSafeMath{GnosisSafeMathCaller: GnosisSafeMathCaller{contract: contract}, GnosisSafeMathTransactor: GnosisSafeMathTransactor{contract: contract}, GnosisSafeMathFilterer: GnosisSafeMathFilterer{contract: contract}}, nil
}

// GnosisSafeMath is an auto generated Go binding around an Ethereum contract.
type GnosisSafeMath struct {
	GnosisSafeMathCaller     // Read-only binding to the contract
	GnosisSafeMathTransactor // Write-only binding to the contract
	GnosisSafeMathFilterer   // Log filterer for contract events
}

// GnosisSafeMathCaller is an auto generated read-only Go binding around an Ethereum contract.
type GnosisSafeMathCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GnosisSafeMathTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GnosisSafeMathTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GnosisSafeMathFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GnosisSafeMathFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GnosisSafeMathSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GnosisSafeMathSession struct {
	Contract     *GnosisSafeMath   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GnosisSafeMathCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GnosisSafeMathCallerSession struct {
	Contract *GnosisSafeMathCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// GnosisSafeMathTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GnosisSafeMathTransactorSession struct {
	Contract     *GnosisSafeMathTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// GnosisSafeMathRaw is an auto generated low-level Go binding around an Ethereum contract.
type GnosisSafeMathRaw struct {
	Contract *GnosisSafeMath // Generic contract binding to access the raw methods on
}

// GnosisSafeMathCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GnosisSafeMathCallerRaw struct {
	Contract *GnosisSafeMathCaller // Generic read-only contract binding to access the raw methods on
}

// GnosisSafeMathTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GnosisSafeMathTransactorRaw struct {
	Contract *GnosisSafeMathTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGnosisSafeMath creates a new instance of GnosisSafeMath, bound to a specific deployed contract.
func NewGnosisSafeMath(address common.Address, backend bind.ContractBackend) (*GnosisSafeMath, error) {
	contract, err := bindGnosisSafeMath(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeMath{GnosisSafeMathCaller: GnosisSafeMathCaller{contract: contract}, GnosisSafeMathTransactor: GnosisSafeMathTransactor{contract: contract}, GnosisSafeMathFilterer: GnosisSafeMathFilterer{contract: contract}}, nil
}

// NewGnosisSafeMathCaller creates a new read-only instance of GnosisSafeMath, bound to a specific deployed contract.
func NewGnosisSafeMathCaller(address common.Address, caller bind.ContractCaller) (*GnosisSafeMathCaller, error) {
	contract, err := bindGnosisSafeMath(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeMathCaller{contract: contract}, nil
}

// NewGnosisSafeMathTransactor creates a new write-only instance of GnosisSafeMath, bound to a specific deployed contract.
func NewGnosisSafeMathTransactor(address common.Address, transactor bind.ContractTransactor) (*GnosisSafeMathTransactor, error) {
	contract, err := bindGnosisSafeMath(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeMathTransactor{contract: contract}, nil
}

// NewGnosisSafeMathFilterer creates a new log filterer instance of GnosisSafeMath, bound to a specific deployed contract.
func NewGnosisSafeMathFilterer(address common.Address, filterer bind.ContractFilterer) (*GnosisSafeMathFilterer, error) {
	contract, err := bindGnosisSafeMath(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GnosisSafeMathFilterer{contract: contract}, nil
}

// bindGnosisSafeMath binds a generic wrapper to an already deployed contract.
func bindGnosisSafeMath(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GnosisSafeMathABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GnosisSafeMath *GnosisSafeMathRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GnosisSafeMath.Contract.GnosisSafeMathCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GnosisSafeMath *GnosisSafeMathRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GnosisSafeMath.Contract.GnosisSafeMathTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GnosisSafeMath *GnosisSafeMathRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GnosisSafeMath.Contract.GnosisSafeMathTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GnosisSafeMath *GnosisSafeMathCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GnosisSafeMath.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GnosisSafeMath *GnosisSafeMathTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GnosisSafeMath.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GnosisSafeMath *GnosisSafeMathTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GnosisSafeMath.Contract.contract.Transact(opts, method, params...)
}

// ISignatureValidatorABI is the input ABI used to generate the binding from.
const ISignatureValidatorABI = "[{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"}],\"name\":\"isValidSignature\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ISignatureValidatorFuncSigs maps the 4-byte function signature to its string representation.
var ISignatureValidatorFuncSigs = map[string]string{
	"20c13b0b": "isValidSignature(bytes,bytes)",
}

// ISignatureValidator is an auto generated Go binding around an Ethereum contract.
type ISignatureValidator struct {
	ISignatureValidatorCaller     // Read-only binding to the contract
	ISignatureValidatorTransactor // Write-only binding to the contract
	ISignatureValidatorFilterer   // Log filterer for contract events
}

// ISignatureValidatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type ISignatureValidatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISignatureValidatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ISignatureValidatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISignatureValidatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ISignatureValidatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISignatureValidatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ISignatureValidatorSession struct {
	Contract     *ISignatureValidator // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ISignatureValidatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ISignatureValidatorCallerSession struct {
	Contract *ISignatureValidatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// ISignatureValidatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ISignatureValidatorTransactorSession struct {
	Contract     *ISignatureValidatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// ISignatureValidatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type ISignatureValidatorRaw struct {
	Contract *ISignatureValidator // Generic contract binding to access the raw methods on
}

// ISignatureValidatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ISignatureValidatorCallerRaw struct {
	Contract *ISignatureValidatorCaller // Generic read-only contract binding to access the raw methods on
}

// ISignatureValidatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ISignatureValidatorTransactorRaw struct {
	Contract *ISignatureValidatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewISignatureValidator creates a new instance of ISignatureValidator, bound to a specific deployed contract.
func NewISignatureValidator(address common.Address, backend bind.ContractBackend) (*ISignatureValidator, error) {
	contract, err := bindISignatureValidator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ISignatureValidator{ISignatureValidatorCaller: ISignatureValidatorCaller{contract: contract}, ISignatureValidatorTransactor: ISignatureValidatorTransactor{contract: contract}, ISignatureValidatorFilterer: ISignatureValidatorFilterer{contract: contract}}, nil
}

// NewISignatureValidatorCaller creates a new read-only instance of ISignatureValidator, bound to a specific deployed contract.
func NewISignatureValidatorCaller(address common.Address, caller bind.ContractCaller) (*ISignatureValidatorCaller, error) {
	contract, err := bindISignatureValidator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ISignatureValidatorCaller{contract: contract}, nil
}

// NewISignatureValidatorTransactor creates a new write-only instance of ISignatureValidator, bound to a specific deployed contract.
func NewISignatureValidatorTransactor(address common.Address, transactor bind.ContractTransactor) (*ISignatureValidatorTransactor, error) {
	contract, err := bindISignatureValidator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ISignatureValidatorTransactor{contract: contract}, nil
}

// NewISignatureValidatorFilterer creates a new log filterer instance of ISignatureValidator, bound to a specific deployed contract.
func NewISignatureValidatorFilterer(address common.Address, filterer bind.ContractFilterer) (*ISignatureValidatorFilterer, error) {
	contract, err := bindISignatureValidator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ISignatureValidatorFilterer{contract: contract}, nil
}

// bindISignatureValidator binds a generic wrapper to an already deployed contract.
func bindISignatureValidator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ISignatureValidatorABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ISignatureValidator *ISignatureValidatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ISignatureValidator.Contract.ISignatureValidatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ISignatureValidator *ISignatureValidatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ISignatureValidator.Contract.ISignatureValidatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ISignatureValidator *ISignatureValidatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ISignatureValidator.Contract.ISignatureValidatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ISignatureValidator *ISignatureValidatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ISignatureValidator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ISignatureValidator *ISignatureValidatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ISignatureValidator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ISignatureValidator *ISignatureValidatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ISignatureValidator.Contract.contract.Transact(opts, method, params...)
}

// IsValidSignature is a free data retrieval call binding the contract method 0x20c13b0b.
//
// Solidity: function isValidSignature(bytes _data, bytes _signature) view returns(bytes4)
func (_ISignatureValidator *ISignatureValidatorCaller) IsValidSignature(opts *bind.CallOpts, _data []byte, _signature []byte) ([4]byte, error) {
	var out []interface{}
	err := _ISignatureValidator.contract.Call(opts, &out, "isValidSignature", _data, _signature)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// IsValidSignature is a free data retrieval call binding the contract method 0x20c13b0b.
//
// Solidity: function isValidSignature(bytes _data, bytes _signature) view returns(bytes4)
func (_ISignatureValidator *ISignatureValidatorSession) IsValidSignature(_data []byte, _signature []byte) ([4]byte, error) {
	return _ISignatureValidator.Contract.IsValidSignature(&_ISignatureValidator.CallOpts, _data, _signature)
}

// IsValidSignature is a free data retrieval call binding the contract method 0x20c13b0b.
//
// Solidity: function isValidSignature(bytes _data, bytes _signature) view returns(bytes4)
func (_ISignatureValidator *ISignatureValidatorCallerSession) IsValidSignature(_data []byte, _signature []byte) ([4]byte, error) {
	return _ISignatureValidator.Contract.IsValidSignature(&_ISignatureValidator.CallOpts, _data, _signature)
}

// ISignatureValidatorConstantsABI is the input ABI used to generate the binding from.
const ISignatureValidatorConstantsABI = "[]"

// ISignatureValidatorConstantsBin is the compiled bytecode used for deploying new contracts.
var ISignatureValidatorConstantsBin = "0x6080604052348015600f57600080fd5b50603f80601d6000396000f3fe6080604052600080fdfea2646970667358221220b33e806db0db6ae03ffd5df26733aa553e714771c18d1b60d77b7dd01ea4b87164736f6c63430007000033"

// DeployISignatureValidatorConstants deploys a new Ethereum contract, binding an instance of ISignatureValidatorConstants to it.
func DeployISignatureValidatorConstants(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ISignatureValidatorConstants, error) {
	parsed, err := abi.JSON(strings.NewReader(ISignatureValidatorConstantsABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ISignatureValidatorConstantsBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ISignatureValidatorConstants{ISignatureValidatorConstantsCaller: ISignatureValidatorConstantsCaller{contract: contract}, ISignatureValidatorConstantsTransactor: ISignatureValidatorConstantsTransactor{contract: contract}, ISignatureValidatorConstantsFilterer: ISignatureValidatorConstantsFilterer{contract: contract}}, nil
}

// ISignatureValidatorConstants is an auto generated Go binding around an Ethereum contract.
type ISignatureValidatorConstants struct {
	ISignatureValidatorConstantsCaller     // Read-only binding to the contract
	ISignatureValidatorConstantsTransactor // Write-only binding to the contract
	ISignatureValidatorConstantsFilterer   // Log filterer for contract events
}

// ISignatureValidatorConstantsCaller is an auto generated read-only Go binding around an Ethereum contract.
type ISignatureValidatorConstantsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISignatureValidatorConstantsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ISignatureValidatorConstantsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISignatureValidatorConstantsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ISignatureValidatorConstantsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISignatureValidatorConstantsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ISignatureValidatorConstantsSession struct {
	Contract     *ISignatureValidatorConstants // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                 // Call options to use throughout this session
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// ISignatureValidatorConstantsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ISignatureValidatorConstantsCallerSession struct {
	Contract *ISignatureValidatorConstantsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                       // Call options to use throughout this session
}

// ISignatureValidatorConstantsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ISignatureValidatorConstantsTransactorSession struct {
	Contract     *ISignatureValidatorConstantsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                       // Transaction auth options to use throughout this session
}

// ISignatureValidatorConstantsRaw is an auto generated low-level Go binding around an Ethereum contract.
type ISignatureValidatorConstantsRaw struct {
	Contract *ISignatureValidatorConstants // Generic contract binding to access the raw methods on
}

// ISignatureValidatorConstantsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ISignatureValidatorConstantsCallerRaw struct {
	Contract *ISignatureValidatorConstantsCaller // Generic read-only contract binding to access the raw methods on
}

// ISignatureValidatorConstantsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ISignatureValidatorConstantsTransactorRaw struct {
	Contract *ISignatureValidatorConstantsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewISignatureValidatorConstants creates a new instance of ISignatureValidatorConstants, bound to a specific deployed contract.
func NewISignatureValidatorConstants(address common.Address, backend bind.ContractBackend) (*ISignatureValidatorConstants, error) {
	contract, err := bindISignatureValidatorConstants(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ISignatureValidatorConstants{ISignatureValidatorConstantsCaller: ISignatureValidatorConstantsCaller{contract: contract}, ISignatureValidatorConstantsTransactor: ISignatureValidatorConstantsTransactor{contract: contract}, ISignatureValidatorConstantsFilterer: ISignatureValidatorConstantsFilterer{contract: contract}}, nil
}

// NewISignatureValidatorConstantsCaller creates a new read-only instance of ISignatureValidatorConstants, bound to a specific deployed contract.
func NewISignatureValidatorConstantsCaller(address common.Address, caller bind.ContractCaller) (*ISignatureValidatorConstantsCaller, error) {
	contract, err := bindISignatureValidatorConstants(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ISignatureValidatorConstantsCaller{contract: contract}, nil
}

// NewISignatureValidatorConstantsTransactor creates a new write-only instance of ISignatureValidatorConstants, bound to a specific deployed contract.
func NewISignatureValidatorConstantsTransactor(address common.Address, transactor bind.ContractTransactor) (*ISignatureValidatorConstantsTransactor, error) {
	contract, err := bindISignatureValidatorConstants(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ISignatureValidatorConstantsTransactor{contract: contract}, nil
}

// NewISignatureValidatorConstantsFilterer creates a new log filterer instance of ISignatureValidatorConstants, bound to a specific deployed contract.
func NewISignatureValidatorConstantsFilterer(address common.Address, filterer bind.ContractFilterer) (*ISignatureValidatorConstantsFilterer, error) {
	contract, err := bindISignatureValidatorConstants(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ISignatureValidatorConstantsFilterer{contract: contract}, nil
}

// bindISignatureValidatorConstants binds a generic wrapper to an already deployed contract.
func bindISignatureValidatorConstants(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ISignatureValidatorConstantsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ISignatureValidatorConstants *ISignatureValidatorConstantsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ISignatureValidatorConstants.Contract.ISignatureValidatorConstantsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ISignatureValidatorConstants *ISignatureValidatorConstantsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ISignatureValidatorConstants.Contract.ISignatureValidatorConstantsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ISignatureValidatorConstants *ISignatureValidatorConstantsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ISignatureValidatorConstants.Contract.ISignatureValidatorConstantsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ISignatureValidatorConstants *ISignatureValidatorConstantsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ISignatureValidatorConstants.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ISignatureValidatorConstants *ISignatureValidatorConstantsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ISignatureValidatorConstants.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ISignatureValidatorConstants *ISignatureValidatorConstantsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ISignatureValidatorConstants.Contract.contract.Transact(opts, method, params...)
}

// ModuleManagerABI is the input ABI used to generate the binding from.
const ModuleManagerABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"DisabledModule\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"EnabledModule\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"ExecutionFromModuleFailure\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"ExecutionFromModuleSuccess\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"prevModule\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"disableModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"enableModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"}],\"name\":\"execTransactionFromModule\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"}],\"name\":\"execTransactionFromModuleReturnData\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"returnData\",\"type\":\"bytes\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"start\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"pageSize\",\"type\":\"uint256\"}],\"name\":\"getModulesPaginated\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"array\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"next\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"isModuleEnabled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ModuleManagerFuncSigs maps the 4-byte function signature to its string representation.
var ModuleManagerFuncSigs = map[string]string{
	"e009cfde": "disableModule(address,address)",
	"610b5925": "enableModule(address)",
	"468721a7": "execTransactionFromModule(address,uint256,bytes,uint8)",
	"5229073f": "execTransactionFromModuleReturnData(address,uint256,bytes,uint8)",
	"cc2f8452": "getModulesPaginated(address,uint256)",
	"2d9ad53d": "isModuleEnabled(address)",
}

// ModuleManagerBin is the compiled bytecode used for deploying new contracts.
var ModuleManagerBin = "0x608060405234801561001057600080fd5b506109cc806100206000396000f3fe608060405234801561001057600080fd5b50600436106100625760003560e01c80632d9ad53d14610067578063468721a7146100a15780635229073f14610161578063610b5925146102a2578063cc2f8452146102ca578063e009cfde1461035a575b600080fd5b61008d6004803603602081101561007d57600080fd5b50356001600160a01b0316610388565b604080519115158252519081900360200190f35b61008d600480360360808110156100b757600080fd5b6001600160a01b03823516916020810135918101906060810160408201356401000000008111156100e757600080fd5b8201836020820111156100f957600080fd5b8035906020019184600183028401116401000000008311171561011b57600080fd5b91908080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509295505050903560ff1691506103c39050565b6102216004803603608081101561017757600080fd5b6001600160a01b03823516916020810135918101906060810160408201356401000000008111156101a757600080fd5b8201836020820111156101b957600080fd5b803590602001918460018302840111640100000000831117156101db57600080fd5b91908080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509295505050903560ff1691506104a19050565b60405180831515815260200180602001828103825283818151815260200191508051906020019080838360005b8381101561026657818101518382015260200161024e565b50505050905090810190601f1680156102935780820380516001836020036101000a031916815260200191505b50935050505060405180910390f35b6102c8600480360360208110156102b857600080fd5b50356001600160a01b03166104d7565b005b6102f6600480360360408110156102e057600080fd5b506001600160a01b038135169060200135610652565b6040518080602001836001600160a01b03168152602001828103825284818151815260200191508051906020019060200280838360005b8381101561034557818101518382015260200161032d565b50505050905001935050505060405180910390f35b6102c86004803603604081101561037057600080fd5b506001600160a01b038135811691602001351661073e565b600060016001600160a01b038316148015906103bd57506001600160a01b038281166000908152602081905260409020541615155b92915050565b6000336001148015906103ed5750336000908152602081905260409020546001600160a01b031615155b6104285760405162461bcd60e51b81526004018080602001828103825260308152602001806109676030913960400191505060405180910390fd5b610435858585855a610890565b9050801561046d5760405133907f6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb890600090a2610499565b60405133907facd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd37590600090a25b949350505050565b600060606104b1868686866103c3565b915060405160203d0181016040523d81523d6000602083013e8091505094509492505050565b6104df6108d2565b6001600160a01b0381161580159061050157506001600160a01b038116600114155b610552576040805162461bcd60e51b815260206004820152601f60248201527f496e76616c6964206d6f64756c6520616464726573732070726f766964656400604482015290519081900360640190fd5b6001600160a01b0381811660009081526020819052604090205416156105bf576040805162461bcd60e51b815260206004820152601d60248201527f4d6f64756c652068617320616c7265616479206265656e206164646564000000604482015290519081900360640190fd5b600060208181527fada5013122d395ba3c54772283fb069b10426056ef8ca54750cb9bb552a59e7d80546001600160a01b0385811680865260408087208054939094166001600160a01b031993841617909355600190955282541684179091558051928352517fecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f84409281900390910190a150565b606060008267ffffffffffffffff8111801561066d57600080fd5b50604051908082528060200260200182016040528015610697578160200160208202803683370190505b506001600160a01b0380861660009081526020819052604081205492945091165b6001600160a01b038116158015906106da57506001600160a01b038116600114155b80156106e557508482105b1561073057808483815181106106f757fe5b6001600160a01b0392831660209182029290920181019190915291811660009081529182905260409091205460019290920191166106b8565b908352919491935090915050565b6107466108d2565b6001600160a01b0381161580159061076857506001600160a01b038116600114155b6107b9576040805162461bcd60e51b815260206004820152601f60248201527f496e76616c6964206d6f64756c6520616464726573732070726f766964656400604482015290519081900360640190fd5b6001600160a01b038281166000908152602081905260409020548116908216146108145760405162461bcd60e51b81526004018080602001828103825260288152602001806109136028913960400191505060405180910390fd5b6001600160a01b03818116600081815260208181526040808320805488871685528285208054919097166001600160a01b031991821617909655928490528254909416909155825191825291517faab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276929181900390910190a15050565b600060018360018111156108a057fe5b14156108b9576000808551602087018986f490506108c9565b600080855160208701888a87f190505b95945050505050565b3330146109105760405162461bcd60e51b815260040180806020018281038252602c81526020018061093b602c913960400191505060405180910390fd5b56fe496e76616c696420707265764d6f64756c652c206d6f64756c6520706169722070726f76696465644d6574686f642063616e206f6e6c792062652063616c6c65642066726f6d207468697320636f6e74726163744d6574686f642063616e206f6e6c792062652063616c6c65642066726f6d20616e20656e61626c6564206d6f64756c65a2646970667358221220678c61a09e1f14c61341f9ca8d4d7f45e9a81407eef6980a166a723b98e75d1a64736f6c63430007000033"

// DeployModuleManager deploys a new Ethereum contract, binding an instance of ModuleManager to it.
func DeployModuleManager(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ModuleManager, error) {
	parsed, err := abi.JSON(strings.NewReader(ModuleManagerABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ModuleManagerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ModuleManager{ModuleManagerCaller: ModuleManagerCaller{contract: contract}, ModuleManagerTransactor: ModuleManagerTransactor{contract: contract}, ModuleManagerFilterer: ModuleManagerFilterer{contract: contract}}, nil
}

// ModuleManager is an auto generated Go binding around an Ethereum contract.
type ModuleManager struct {
	ModuleManagerCaller     // Read-only binding to the contract
	ModuleManagerTransactor // Write-only binding to the contract
	ModuleManagerFilterer   // Log filterer for contract events
}

// ModuleManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type ModuleManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ModuleManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ModuleManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ModuleManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ModuleManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ModuleManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ModuleManagerSession struct {
	Contract     *ModuleManager    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ModuleManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ModuleManagerCallerSession struct {
	Contract *ModuleManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ModuleManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ModuleManagerTransactorSession struct {
	Contract     *ModuleManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ModuleManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type ModuleManagerRaw struct {
	Contract *ModuleManager // Generic contract binding to access the raw methods on
}

// ModuleManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ModuleManagerCallerRaw struct {
	Contract *ModuleManagerCaller // Generic read-only contract binding to access the raw methods on
}

// ModuleManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ModuleManagerTransactorRaw struct {
	Contract *ModuleManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewModuleManager creates a new instance of ModuleManager, bound to a specific deployed contract.
func NewModuleManager(address common.Address, backend bind.ContractBackend) (*ModuleManager, error) {
	contract, err := bindModuleManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ModuleManager{ModuleManagerCaller: ModuleManagerCaller{contract: contract}, ModuleManagerTransactor: ModuleManagerTransactor{contract: contract}, ModuleManagerFilterer: ModuleManagerFilterer{contract: contract}}, nil
}

// NewModuleManagerCaller creates a new read-only instance of ModuleManager, bound to a specific deployed contract.
func NewModuleManagerCaller(address common.Address, caller bind.ContractCaller) (*ModuleManagerCaller, error) {
	contract, err := bindModuleManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ModuleManagerCaller{contract: contract}, nil
}

// NewModuleManagerTransactor creates a new write-only instance of ModuleManager, bound to a specific deployed contract.
func NewModuleManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*ModuleManagerTransactor, error) {
	contract, err := bindModuleManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ModuleManagerTransactor{contract: contract}, nil
}

// NewModuleManagerFilterer creates a new log filterer instance of ModuleManager, bound to a specific deployed contract.
func NewModuleManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*ModuleManagerFilterer, error) {
	contract, err := bindModuleManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ModuleManagerFilterer{contract: contract}, nil
}

// bindModuleManager binds a generic wrapper to an already deployed contract.
func bindModuleManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ModuleManagerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ModuleManager *ModuleManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ModuleManager.Contract.ModuleManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ModuleManager *ModuleManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ModuleManager.Contract.ModuleManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ModuleManager *ModuleManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ModuleManager.Contract.ModuleManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ModuleManager *ModuleManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ModuleManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ModuleManager *ModuleManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ModuleManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ModuleManager *ModuleManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ModuleManager.Contract.contract.Transact(opts, method, params...)
}

// GetModulesPaginated is a free data retrieval call binding the contract method 0xcc2f8452.
//
// Solidity: function getModulesPaginated(address start, uint256 pageSize) view returns(address[] array, address next)
func (_ModuleManager *ModuleManagerCaller) GetModulesPaginated(opts *bind.CallOpts, start common.Address, pageSize *big.Int) (struct {
	Array []common.Address
	Next  common.Address
}, error) {
	var out []interface{}
	err := _ModuleManager.contract.Call(opts, &out, "getModulesPaginated", start, pageSize)

	outstruct := new(struct {
		Array []common.Address
		Next  common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Array = *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	outstruct.Next = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// GetModulesPaginated is a free data retrieval call binding the contract method 0xcc2f8452.
//
// Solidity: function getModulesPaginated(address start, uint256 pageSize) view returns(address[] array, address next)
func (_ModuleManager *ModuleManagerSession) GetModulesPaginated(start common.Address, pageSize *big.Int) (struct {
	Array []common.Address
	Next  common.Address
}, error) {
	return _ModuleManager.Contract.GetModulesPaginated(&_ModuleManager.CallOpts, start, pageSize)
}

// GetModulesPaginated is a free data retrieval call binding the contract method 0xcc2f8452.
//
// Solidity: function getModulesPaginated(address start, uint256 pageSize) view returns(address[] array, address next)
func (_ModuleManager *ModuleManagerCallerSession) GetModulesPaginated(start common.Address, pageSize *big.Int) (struct {
	Array []common.Address
	Next  common.Address
}, error) {
	return _ModuleManager.Contract.GetModulesPaginated(&_ModuleManager.CallOpts, start, pageSize)
}

// IsModuleEnabled is a free data retrieval call binding the contract method 0x2d9ad53d.
//
// Solidity: function isModuleEnabled(address module) view returns(bool)
func (_ModuleManager *ModuleManagerCaller) IsModuleEnabled(opts *bind.CallOpts, module common.Address) (bool, error) {
	var out []interface{}
	err := _ModuleManager.contract.Call(opts, &out, "isModuleEnabled", module)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsModuleEnabled is a free data retrieval call binding the contract method 0x2d9ad53d.
//
// Solidity: function isModuleEnabled(address module) view returns(bool)
func (_ModuleManager *ModuleManagerSession) IsModuleEnabled(module common.Address) (bool, error) {
	return _ModuleManager.Contract.IsModuleEnabled(&_ModuleManager.CallOpts, module)
}

// IsModuleEnabled is a free data retrieval call binding the contract method 0x2d9ad53d.
//
// Solidity: function isModuleEnabled(address module) view returns(bool)
func (_ModuleManager *ModuleManagerCallerSession) IsModuleEnabled(module common.Address) (bool, error) {
	return _ModuleManager.Contract.IsModuleEnabled(&_ModuleManager.CallOpts, module)
}

// DisableModule is a paid mutator transaction binding the contract method 0xe009cfde.
//
// Solidity: function disableModule(address prevModule, address module) returns()
func (_ModuleManager *ModuleManagerTransactor) DisableModule(opts *bind.TransactOpts, prevModule common.Address, module common.Address) (*types.Transaction, error) {
	return _ModuleManager.contract.Transact(opts, "disableModule", prevModule, module)
}

// DisableModule is a paid mutator transaction binding the contract method 0xe009cfde.
//
// Solidity: function disableModule(address prevModule, address module) returns()
func (_ModuleManager *ModuleManagerSession) DisableModule(prevModule common.Address, module common.Address) (*types.Transaction, error) {
	return _ModuleManager.Contract.DisableModule(&_ModuleManager.TransactOpts, prevModule, module)
}

// DisableModule is a paid mutator transaction binding the contract method 0xe009cfde.
//
// Solidity: function disableModule(address prevModule, address module) returns()
func (_ModuleManager *ModuleManagerTransactorSession) DisableModule(prevModule common.Address, module common.Address) (*types.Transaction, error) {
	return _ModuleManager.Contract.DisableModule(&_ModuleManager.TransactOpts, prevModule, module)
}

// EnableModule is a paid mutator transaction binding the contract method 0x610b5925.
//
// Solidity: function enableModule(address module) returns()
func (_ModuleManager *ModuleManagerTransactor) EnableModule(opts *bind.TransactOpts, module common.Address) (*types.Transaction, error) {
	return _ModuleManager.contract.Transact(opts, "enableModule", module)
}

// EnableModule is a paid mutator transaction binding the contract method 0x610b5925.
//
// Solidity: function enableModule(address module) returns()
func (_ModuleManager *ModuleManagerSession) EnableModule(module common.Address) (*types.Transaction, error) {
	return _ModuleManager.Contract.EnableModule(&_ModuleManager.TransactOpts, module)
}

// EnableModule is a paid mutator transaction binding the contract method 0x610b5925.
//
// Solidity: function enableModule(address module) returns()
func (_ModuleManager *ModuleManagerTransactorSession) EnableModule(module common.Address) (*types.Transaction, error) {
	return _ModuleManager.Contract.EnableModule(&_ModuleManager.TransactOpts, module)
}

// ExecTransactionFromModule is a paid mutator transaction binding the contract method 0x468721a7.
//
// Solidity: function execTransactionFromModule(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_ModuleManager *ModuleManagerTransactor) ExecTransactionFromModule(opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _ModuleManager.contract.Transact(opts, "execTransactionFromModule", to, value, data, operation)
}

// ExecTransactionFromModule is a paid mutator transaction binding the contract method 0x468721a7.
//
// Solidity: function execTransactionFromModule(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_ModuleManager *ModuleManagerSession) ExecTransactionFromModule(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _ModuleManager.Contract.ExecTransactionFromModule(&_ModuleManager.TransactOpts, to, value, data, operation)
}

// ExecTransactionFromModule is a paid mutator transaction binding the contract method 0x468721a7.
//
// Solidity: function execTransactionFromModule(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_ModuleManager *ModuleManagerTransactorSession) ExecTransactionFromModule(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _ModuleManager.Contract.ExecTransactionFromModule(&_ModuleManager.TransactOpts, to, value, data, operation)
}

// ExecTransactionFromModuleReturnData is a paid mutator transaction binding the contract method 0x5229073f.
//
// Solidity: function execTransactionFromModuleReturnData(address to, uint256 value, bytes data, uint8 operation) returns(bool success, bytes returnData)
func (_ModuleManager *ModuleManagerTransactor) ExecTransactionFromModuleReturnData(opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _ModuleManager.contract.Transact(opts, "execTransactionFromModuleReturnData", to, value, data, operation)
}

// ExecTransactionFromModuleReturnData is a paid mutator transaction binding the contract method 0x5229073f.
//
// Solidity: function execTransactionFromModuleReturnData(address to, uint256 value, bytes data, uint8 operation) returns(bool success, bytes returnData)
func (_ModuleManager *ModuleManagerSession) ExecTransactionFromModuleReturnData(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _ModuleManager.Contract.ExecTransactionFromModuleReturnData(&_ModuleManager.TransactOpts, to, value, data, operation)
}

// ExecTransactionFromModuleReturnData is a paid mutator transaction binding the contract method 0x5229073f.
//
// Solidity: function execTransactionFromModuleReturnData(address to, uint256 value, bytes data, uint8 operation) returns(bool success, bytes returnData)
func (_ModuleManager *ModuleManagerTransactorSession) ExecTransactionFromModuleReturnData(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _ModuleManager.Contract.ExecTransactionFromModuleReturnData(&_ModuleManager.TransactOpts, to, value, data, operation)
}

// ModuleManagerDisabledModuleIterator is returned from FilterDisabledModule and is used to iterate over the raw logs and unpacked data for DisabledModule events raised by the ModuleManager contract.
type ModuleManagerDisabledModuleIterator struct {
	Event *ModuleManagerDisabledModule // Event containing the contract specifics and raw log

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
func (it *ModuleManagerDisabledModuleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ModuleManagerDisabledModule)
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
		it.Event = new(ModuleManagerDisabledModule)
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
func (it *ModuleManagerDisabledModuleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ModuleManagerDisabledModuleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ModuleManagerDisabledModule represents a DisabledModule event raised by the ModuleManager contract.
type ModuleManagerDisabledModule struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDisabledModule is a free log retrieval operation binding the contract event 0xaab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276.
//
// Solidity: event DisabledModule(address module)
func (_ModuleManager *ModuleManagerFilterer) FilterDisabledModule(opts *bind.FilterOpts) (*ModuleManagerDisabledModuleIterator, error) {

	logs, sub, err := _ModuleManager.contract.FilterLogs(opts, "DisabledModule")
	if err != nil {
		return nil, err
	}
	return &ModuleManagerDisabledModuleIterator{contract: _ModuleManager.contract, event: "DisabledModule", logs: logs, sub: sub}, nil
}

// WatchDisabledModule is a free log subscription operation binding the contract event 0xaab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276.
//
// Solidity: event DisabledModule(address module)
func (_ModuleManager *ModuleManagerFilterer) WatchDisabledModule(opts *bind.WatchOpts, sink chan<- *ModuleManagerDisabledModule) (event.Subscription, error) {

	logs, sub, err := _ModuleManager.contract.WatchLogs(opts, "DisabledModule")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ModuleManagerDisabledModule)
				if err := _ModuleManager.contract.UnpackLog(event, "DisabledModule", log); err != nil {
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

// ParseDisabledModule is a log parse operation binding the contract event 0xaab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276.
//
// Solidity: event DisabledModule(address module)
func (_ModuleManager *ModuleManagerFilterer) ParseDisabledModule(log types.Log) (*ModuleManagerDisabledModule, error) {
	event := new(ModuleManagerDisabledModule)
	if err := _ModuleManager.contract.UnpackLog(event, "DisabledModule", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ModuleManagerEnabledModuleIterator is returned from FilterEnabledModule and is used to iterate over the raw logs and unpacked data for EnabledModule events raised by the ModuleManager contract.
type ModuleManagerEnabledModuleIterator struct {
	Event *ModuleManagerEnabledModule // Event containing the contract specifics and raw log

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
func (it *ModuleManagerEnabledModuleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ModuleManagerEnabledModule)
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
		it.Event = new(ModuleManagerEnabledModule)
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
func (it *ModuleManagerEnabledModuleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ModuleManagerEnabledModuleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ModuleManagerEnabledModule represents a EnabledModule event raised by the ModuleManager contract.
type ModuleManagerEnabledModule struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterEnabledModule is a free log retrieval operation binding the contract event 0xecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f8440.
//
// Solidity: event EnabledModule(address module)
func (_ModuleManager *ModuleManagerFilterer) FilterEnabledModule(opts *bind.FilterOpts) (*ModuleManagerEnabledModuleIterator, error) {

	logs, sub, err := _ModuleManager.contract.FilterLogs(opts, "EnabledModule")
	if err != nil {
		return nil, err
	}
	return &ModuleManagerEnabledModuleIterator{contract: _ModuleManager.contract, event: "EnabledModule", logs: logs, sub: sub}, nil
}

// WatchEnabledModule is a free log subscription operation binding the contract event 0xecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f8440.
//
// Solidity: event EnabledModule(address module)
func (_ModuleManager *ModuleManagerFilterer) WatchEnabledModule(opts *bind.WatchOpts, sink chan<- *ModuleManagerEnabledModule) (event.Subscription, error) {

	logs, sub, err := _ModuleManager.contract.WatchLogs(opts, "EnabledModule")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ModuleManagerEnabledModule)
				if err := _ModuleManager.contract.UnpackLog(event, "EnabledModule", log); err != nil {
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

// ParseEnabledModule is a log parse operation binding the contract event 0xecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f8440.
//
// Solidity: event EnabledModule(address module)
func (_ModuleManager *ModuleManagerFilterer) ParseEnabledModule(log types.Log) (*ModuleManagerEnabledModule, error) {
	event := new(ModuleManagerEnabledModule)
	if err := _ModuleManager.contract.UnpackLog(event, "EnabledModule", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ModuleManagerExecutionFromModuleFailureIterator is returned from FilterExecutionFromModuleFailure and is used to iterate over the raw logs and unpacked data for ExecutionFromModuleFailure events raised by the ModuleManager contract.
type ModuleManagerExecutionFromModuleFailureIterator struct {
	Event *ModuleManagerExecutionFromModuleFailure // Event containing the contract specifics and raw log

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
func (it *ModuleManagerExecutionFromModuleFailureIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ModuleManagerExecutionFromModuleFailure)
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
		it.Event = new(ModuleManagerExecutionFromModuleFailure)
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
func (it *ModuleManagerExecutionFromModuleFailureIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ModuleManagerExecutionFromModuleFailureIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ModuleManagerExecutionFromModuleFailure represents a ExecutionFromModuleFailure event raised by the ModuleManager contract.
type ModuleManagerExecutionFromModuleFailure struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterExecutionFromModuleFailure is a free log retrieval operation binding the contract event 0xacd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd375.
//
// Solidity: event ExecutionFromModuleFailure(address indexed module)
func (_ModuleManager *ModuleManagerFilterer) FilterExecutionFromModuleFailure(opts *bind.FilterOpts, module []common.Address) (*ModuleManagerExecutionFromModuleFailureIterator, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _ModuleManager.contract.FilterLogs(opts, "ExecutionFromModuleFailure", moduleRule)
	if err != nil {
		return nil, err
	}
	return &ModuleManagerExecutionFromModuleFailureIterator{contract: _ModuleManager.contract, event: "ExecutionFromModuleFailure", logs: logs, sub: sub}, nil
}

// WatchExecutionFromModuleFailure is a free log subscription operation binding the contract event 0xacd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd375.
//
// Solidity: event ExecutionFromModuleFailure(address indexed module)
func (_ModuleManager *ModuleManagerFilterer) WatchExecutionFromModuleFailure(opts *bind.WatchOpts, sink chan<- *ModuleManagerExecutionFromModuleFailure, module []common.Address) (event.Subscription, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _ModuleManager.contract.WatchLogs(opts, "ExecutionFromModuleFailure", moduleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ModuleManagerExecutionFromModuleFailure)
				if err := _ModuleManager.contract.UnpackLog(event, "ExecutionFromModuleFailure", log); err != nil {
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

// ParseExecutionFromModuleFailure is a log parse operation binding the contract event 0xacd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd375.
//
// Solidity: event ExecutionFromModuleFailure(address indexed module)
func (_ModuleManager *ModuleManagerFilterer) ParseExecutionFromModuleFailure(log types.Log) (*ModuleManagerExecutionFromModuleFailure, error) {
	event := new(ModuleManagerExecutionFromModuleFailure)
	if err := _ModuleManager.contract.UnpackLog(event, "ExecutionFromModuleFailure", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ModuleManagerExecutionFromModuleSuccessIterator is returned from FilterExecutionFromModuleSuccess and is used to iterate over the raw logs and unpacked data for ExecutionFromModuleSuccess events raised by the ModuleManager contract.
type ModuleManagerExecutionFromModuleSuccessIterator struct {
	Event *ModuleManagerExecutionFromModuleSuccess // Event containing the contract specifics and raw log

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
func (it *ModuleManagerExecutionFromModuleSuccessIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ModuleManagerExecutionFromModuleSuccess)
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
		it.Event = new(ModuleManagerExecutionFromModuleSuccess)
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
func (it *ModuleManagerExecutionFromModuleSuccessIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ModuleManagerExecutionFromModuleSuccessIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ModuleManagerExecutionFromModuleSuccess represents a ExecutionFromModuleSuccess event raised by the ModuleManager contract.
type ModuleManagerExecutionFromModuleSuccess struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterExecutionFromModuleSuccess is a free log retrieval operation binding the contract event 0x6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb8.
//
// Solidity: event ExecutionFromModuleSuccess(address indexed module)
func (_ModuleManager *ModuleManagerFilterer) FilterExecutionFromModuleSuccess(opts *bind.FilterOpts, module []common.Address) (*ModuleManagerExecutionFromModuleSuccessIterator, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _ModuleManager.contract.FilterLogs(opts, "ExecutionFromModuleSuccess", moduleRule)
	if err != nil {
		return nil, err
	}
	return &ModuleManagerExecutionFromModuleSuccessIterator{contract: _ModuleManager.contract, event: "ExecutionFromModuleSuccess", logs: logs, sub: sub}, nil
}

// WatchExecutionFromModuleSuccess is a free log subscription operation binding the contract event 0x6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb8.
//
// Solidity: event ExecutionFromModuleSuccess(address indexed module)
func (_ModuleManager *ModuleManagerFilterer) WatchExecutionFromModuleSuccess(opts *bind.WatchOpts, sink chan<- *ModuleManagerExecutionFromModuleSuccess, module []common.Address) (event.Subscription, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _ModuleManager.contract.WatchLogs(opts, "ExecutionFromModuleSuccess", moduleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ModuleManagerExecutionFromModuleSuccess)
				if err := _ModuleManager.contract.UnpackLog(event, "ExecutionFromModuleSuccess", log); err != nil {
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

// ParseExecutionFromModuleSuccess is a log parse operation binding the contract event 0x6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb8.
//
// Solidity: event ExecutionFromModuleSuccess(address indexed module)
func (_ModuleManager *ModuleManagerFilterer) ParseExecutionFromModuleSuccess(log types.Log) (*ModuleManagerExecutionFromModuleSuccess, error) {
	event := new(ModuleManagerExecutionFromModuleSuccess)
	if err := _ModuleManager.contract.UnpackLog(event, "ExecutionFromModuleSuccess", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OwnerManagerABI is the input ABI used to generate the binding from.
const OwnerManagerABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AddedOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"name\":\"ChangedThreshold\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"RemovedOwner\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"}],\"name\":\"addOwnerWithThreshold\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"}],\"name\":\"changeThreshold\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwners\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"isOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"prevOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"}],\"name\":\"removeOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"prevOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"oldOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"swapOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// OwnerManagerFuncSigs maps the 4-byte function signature to its string representation.
var OwnerManagerFuncSigs = map[string]string{
	"0d582f13": "addOwnerWithThreshold(address,uint256)",
	"694e80c3": "changeThreshold(uint256)",
	"a0e67e2b": "getOwners()",
	"e75235b8": "getThreshold()",
	"2f54bf6e": "isOwner(address)",
	"f8dc5dd9": "removeOwner(address,address,uint256)",
	"e318b52b": "swapOwner(address,address,address)",
}

// OwnerManagerBin is the compiled bytecode used for deploying new contracts.
var OwnerManagerBin = "0x608060405234801561001057600080fd5b50610adf806100206000396000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c8063a0e67e2b1161005b578063a0e67e2b14610107578063e318b52b1461015f578063e75235b814610197578063f8dc5dd9146101b15761007d565b80630d582f13146100825780632f54bf6e146100b0578063694e80c3146100ea575b600080fd5b6100ae6004803603604081101561009857600080fd5b506001600160a01b0381351690602001356101e7565b005b6100d6600480360360208110156100c657600080fd5b50356001600160a01b0316610383565b604080519115158252519081900360200190f35b6100ae6004803603602081101561010057600080fd5b50356103be565b61010f610482565b60408051602080825283518183015283519192839290830191858101910280838360005b8381101561014b578181015183820152602001610133565b505050509050019250505060405180910390f35b6100ae6004803603606081101561017557600080fd5b506001600160a01b038135811691602081013582169160409091013516610564565b61019f6107d5565b60408051918252519081900360200190f35b6100ae600480360360608110156101c757600080fd5b506001600160a01b038135811691602081013590911690604001356107db565b6101ef61097b565b6001600160a01b0382161580159061021157506001600160a01b038216600114155b801561022657506001600160a01b0382163014155b610265576040805162461bcd60e51b815260206004820152601e60248201526000805160206109bc833981519152604482015290519081900360640190fd5b6001600160a01b0382811660009081526020819052604090205416156102d2576040805162461bcd60e51b815260206004820152601b60248201527f4164647265737320697320616c726561647920616e206f776e65720000000000604482015290519081900360640190fd5b600060208181527fada5013122d395ba3c54772283fb069b10426056ef8ca54750cb9bb552a59e7d80546001600160a01b0386811680865260408087208054939094166001600160a01b0319938416179093556001958690528354909116811790925583548401909355825190815291517f9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea269281900390910190a1806002541461037f5761037f816103be565b5050565b60006001600160a01b0382166001148015906103b857506001600160a01b038281166000908152602081905260409020541615155b92915050565b6103c661097b565b6001548111156104075760405162461bcd60e51b81526004018080602001828103825260238152602001806109dc6023913960400191505060405180910390fd5b60018110156104475760405162461bcd60e51b8152600401808060200182810382526024815260200180610a5a6024913960400191505060405180910390fd5b60028190556040805182815290517f610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c939181900360200190a150565b60608060015467ffffffffffffffff8111801561049e57600080fd5b506040519080825280602002602001820160405280156104c8578160200160208202803683370190505b506001600090815260208190527fada5013122d395ba3c54772283fb069b10426056ef8ca54750cb9bb552a59e7d54919250906001600160a01b03165b6001600160a01b03811660011461055c578083838151811061052357fe5b6001600160a01b039283166020918202929092018101919091529181166000908152918290526040909120546001929092019116610505565b509091505090565b61056c61097b565b6001600160a01b0381161580159061058e57506001600160a01b038116600114155b80156105a357506001600160a01b0381163014155b6105e2576040805162461bcd60e51b815260206004820152601e60248201526000805160206109bc833981519152604482015290519081900360640190fd5b6001600160a01b03818116600090815260208190526040902054161561064f576040805162461bcd60e51b815260206004820152601b60248201527f4164647265737320697320616c726561647920616e206f776e65720000000000604482015290519081900360640190fd5b6001600160a01b0382161580159061067157506001600160a01b038216600114155b6106b0576040805162461bcd60e51b815260206004820152601e60248201526000805160206109bc833981519152604482015290519081900360640190fd5b6001600160a01b0383811660009081526020819052604090205481169083161461070b5760405162461bcd60e51b8152600401808060200182810382526026815260200180610a346026913960400191505060405180910390fd5b6001600160a01b03828116600081815260208181526040808320805487871680865283862080549289166001600160a01b0319938416179055968a16855282852080548216909717909655928490528254909416909155825191825291517ff8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf929181900390910190a1604080516001600160a01b038316815290517f9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea269181900360200190a1505050565b60025490565b6107e361097b565b80600180540310156108265760405162461bcd60e51b81526004018080602001828103825260358152602001806109ff6035913960400191505060405180910390fd5b6001600160a01b0382161580159061084857506001600160a01b038216600114155b610887576040805162461bcd60e51b815260206004820152601e60248201526000805160206109bc833981519152604482015290519081900360640190fd5b6001600160a01b038381166000908152602081905260409020548116908316146108e25760405162461bcd60e51b8152600401808060200182810382526026815260200180610a346026913960400191505060405180910390fd5b6001600160a01b03828116600081815260208181526040808320805489871685528285208054919097166001600160a01b03199182161790965592849052825490941690915560018054600019019055825191825291517ff8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf929181900390910190a1806002541461097657610976816103be565b505050565b3330146109b95760405162461bcd60e51b815260040180806020018281038252602c815260200180610a7e602c913960400191505060405180910390fd5b56fe496e76616c6964206f776e657220616464726573732070726f766964656400005468726573686f6c642063616e6e6f7420657863656564206f776e657220636f756e744e6577206f776e657220636f756e74206e6565647320746f206265206c6172676572207468616e206e6577207468726573686f6c64496e76616c696420707265764f776e65722c206f776e657220706169722070726f76696465645468726573686f6c64206e6565647320746f2062652067726561746572207468616e20304d6574686f642063616e206f6e6c792062652063616c6c65642066726f6d207468697320636f6e7472616374a264697066735822122076459d7829fa821e915fa6b8e533c822fd80926e139e4c7696b4a592b9b4891764736f6c63430007000033"

// DeployOwnerManager deploys a new Ethereum contract, binding an instance of OwnerManager to it.
func DeployOwnerManager(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *OwnerManager, error) {
	parsed, err := abi.JSON(strings.NewReader(OwnerManagerABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(OwnerManagerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &OwnerManager{OwnerManagerCaller: OwnerManagerCaller{contract: contract}, OwnerManagerTransactor: OwnerManagerTransactor{contract: contract}, OwnerManagerFilterer: OwnerManagerFilterer{contract: contract}}, nil
}

// OwnerManager is an auto generated Go binding around an Ethereum contract.
type OwnerManager struct {
	OwnerManagerCaller     // Read-only binding to the contract
	OwnerManagerTransactor // Write-only binding to the contract
	OwnerManagerFilterer   // Log filterer for contract events
}

// OwnerManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type OwnerManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnerManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OwnerManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnerManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OwnerManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnerManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OwnerManagerSession struct {
	Contract     *OwnerManager     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OwnerManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OwnerManagerCallerSession struct {
	Contract *OwnerManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// OwnerManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OwnerManagerTransactorSession struct {
	Contract     *OwnerManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// OwnerManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type OwnerManagerRaw struct {
	Contract *OwnerManager // Generic contract binding to access the raw methods on
}

// OwnerManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OwnerManagerCallerRaw struct {
	Contract *OwnerManagerCaller // Generic read-only contract binding to access the raw methods on
}

// OwnerManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OwnerManagerTransactorRaw struct {
	Contract *OwnerManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOwnerManager creates a new instance of OwnerManager, bound to a specific deployed contract.
func NewOwnerManager(address common.Address, backend bind.ContractBackend) (*OwnerManager, error) {
	contract, err := bindOwnerManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OwnerManager{OwnerManagerCaller: OwnerManagerCaller{contract: contract}, OwnerManagerTransactor: OwnerManagerTransactor{contract: contract}, OwnerManagerFilterer: OwnerManagerFilterer{contract: contract}}, nil
}

// NewOwnerManagerCaller creates a new read-only instance of OwnerManager, bound to a specific deployed contract.
func NewOwnerManagerCaller(address common.Address, caller bind.ContractCaller) (*OwnerManagerCaller, error) {
	contract, err := bindOwnerManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OwnerManagerCaller{contract: contract}, nil
}

// NewOwnerManagerTransactor creates a new write-only instance of OwnerManager, bound to a specific deployed contract.
func NewOwnerManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*OwnerManagerTransactor, error) {
	contract, err := bindOwnerManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OwnerManagerTransactor{contract: contract}, nil
}

// NewOwnerManagerFilterer creates a new log filterer instance of OwnerManager, bound to a specific deployed contract.
func NewOwnerManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*OwnerManagerFilterer, error) {
	contract, err := bindOwnerManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OwnerManagerFilterer{contract: contract}, nil
}

// bindOwnerManager binds a generic wrapper to an already deployed contract.
func bindOwnerManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OwnerManagerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OwnerManager *OwnerManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OwnerManager.Contract.OwnerManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OwnerManager *OwnerManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OwnerManager.Contract.OwnerManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OwnerManager *OwnerManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OwnerManager.Contract.OwnerManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OwnerManager *OwnerManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OwnerManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OwnerManager *OwnerManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OwnerManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OwnerManager *OwnerManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OwnerManager.Contract.contract.Transact(opts, method, params...)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_OwnerManager *OwnerManagerCaller) GetOwners(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _OwnerManager.contract.Call(opts, &out, "getOwners")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_OwnerManager *OwnerManagerSession) GetOwners() ([]common.Address, error) {
	return _OwnerManager.Contract.GetOwners(&_OwnerManager.CallOpts)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_OwnerManager *OwnerManagerCallerSession) GetOwners() ([]common.Address, error) {
	return _OwnerManager.Contract.GetOwners(&_OwnerManager.CallOpts)
}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint256)
func (_OwnerManager *OwnerManagerCaller) GetThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OwnerManager.contract.Call(opts, &out, "getThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint256)
func (_OwnerManager *OwnerManagerSession) GetThreshold() (*big.Int, error) {
	return _OwnerManager.Contract.GetThreshold(&_OwnerManager.CallOpts)
}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint256)
func (_OwnerManager *OwnerManagerCallerSession) GetThreshold() (*big.Int, error) {
	return _OwnerManager.Contract.GetThreshold(&_OwnerManager.CallOpts)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address owner) view returns(bool)
func (_OwnerManager *OwnerManagerCaller) IsOwner(opts *bind.CallOpts, owner common.Address) (bool, error) {
	var out []interface{}
	err := _OwnerManager.contract.Call(opts, &out, "isOwner", owner)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address owner) view returns(bool)
func (_OwnerManager *OwnerManagerSession) IsOwner(owner common.Address) (bool, error) {
	return _OwnerManager.Contract.IsOwner(&_OwnerManager.CallOpts, owner)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address owner) view returns(bool)
func (_OwnerManager *OwnerManagerCallerSession) IsOwner(owner common.Address) (bool, error) {
	return _OwnerManager.Contract.IsOwner(&_OwnerManager.CallOpts, owner)
}

// AddOwnerWithThreshold is a paid mutator transaction binding the contract method 0x0d582f13.
//
// Solidity: function addOwnerWithThreshold(address owner, uint256 _threshold) returns()
func (_OwnerManager *OwnerManagerTransactor) AddOwnerWithThreshold(opts *bind.TransactOpts, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _OwnerManager.contract.Transact(opts, "addOwnerWithThreshold", owner, _threshold)
}

// AddOwnerWithThreshold is a paid mutator transaction binding the contract method 0x0d582f13.
//
// Solidity: function addOwnerWithThreshold(address owner, uint256 _threshold) returns()
func (_OwnerManager *OwnerManagerSession) AddOwnerWithThreshold(owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _OwnerManager.Contract.AddOwnerWithThreshold(&_OwnerManager.TransactOpts, owner, _threshold)
}

// AddOwnerWithThreshold is a paid mutator transaction binding the contract method 0x0d582f13.
//
// Solidity: function addOwnerWithThreshold(address owner, uint256 _threshold) returns()
func (_OwnerManager *OwnerManagerTransactorSession) AddOwnerWithThreshold(owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _OwnerManager.Contract.AddOwnerWithThreshold(&_OwnerManager.TransactOpts, owner, _threshold)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 _threshold) returns()
func (_OwnerManager *OwnerManagerTransactor) ChangeThreshold(opts *bind.TransactOpts, _threshold *big.Int) (*types.Transaction, error) {
	return _OwnerManager.contract.Transact(opts, "changeThreshold", _threshold)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 _threshold) returns()
func (_OwnerManager *OwnerManagerSession) ChangeThreshold(_threshold *big.Int) (*types.Transaction, error) {
	return _OwnerManager.Contract.ChangeThreshold(&_OwnerManager.TransactOpts, _threshold)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 _threshold) returns()
func (_OwnerManager *OwnerManagerTransactorSession) ChangeThreshold(_threshold *big.Int) (*types.Transaction, error) {
	return _OwnerManager.Contract.ChangeThreshold(&_OwnerManager.TransactOpts, _threshold)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0xf8dc5dd9.
//
// Solidity: function removeOwner(address prevOwner, address owner, uint256 _threshold) returns()
func (_OwnerManager *OwnerManagerTransactor) RemoveOwner(opts *bind.TransactOpts, prevOwner common.Address, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _OwnerManager.contract.Transact(opts, "removeOwner", prevOwner, owner, _threshold)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0xf8dc5dd9.
//
// Solidity: function removeOwner(address prevOwner, address owner, uint256 _threshold) returns()
func (_OwnerManager *OwnerManagerSession) RemoveOwner(prevOwner common.Address, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _OwnerManager.Contract.RemoveOwner(&_OwnerManager.TransactOpts, prevOwner, owner, _threshold)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0xf8dc5dd9.
//
// Solidity: function removeOwner(address prevOwner, address owner, uint256 _threshold) returns()
func (_OwnerManager *OwnerManagerTransactorSession) RemoveOwner(prevOwner common.Address, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _OwnerManager.Contract.RemoveOwner(&_OwnerManager.TransactOpts, prevOwner, owner, _threshold)
}

// SwapOwner is a paid mutator transaction binding the contract method 0xe318b52b.
//
// Solidity: function swapOwner(address prevOwner, address oldOwner, address newOwner) returns()
func (_OwnerManager *OwnerManagerTransactor) SwapOwner(opts *bind.TransactOpts, prevOwner common.Address, oldOwner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _OwnerManager.contract.Transact(opts, "swapOwner", prevOwner, oldOwner, newOwner)
}

// SwapOwner is a paid mutator transaction binding the contract method 0xe318b52b.
//
// Solidity: function swapOwner(address prevOwner, address oldOwner, address newOwner) returns()
func (_OwnerManager *OwnerManagerSession) SwapOwner(prevOwner common.Address, oldOwner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _OwnerManager.Contract.SwapOwner(&_OwnerManager.TransactOpts, prevOwner, oldOwner, newOwner)
}

// SwapOwner is a paid mutator transaction binding the contract method 0xe318b52b.
//
// Solidity: function swapOwner(address prevOwner, address oldOwner, address newOwner) returns()
func (_OwnerManager *OwnerManagerTransactorSession) SwapOwner(prevOwner common.Address, oldOwner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _OwnerManager.Contract.SwapOwner(&_OwnerManager.TransactOpts, prevOwner, oldOwner, newOwner)
}

// OwnerManagerAddedOwnerIterator is returned from FilterAddedOwner and is used to iterate over the raw logs and unpacked data for AddedOwner events raised by the OwnerManager contract.
type OwnerManagerAddedOwnerIterator struct {
	Event *OwnerManagerAddedOwner // Event containing the contract specifics and raw log

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
func (it *OwnerManagerAddedOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OwnerManagerAddedOwner)
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
		it.Event = new(OwnerManagerAddedOwner)
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
func (it *OwnerManagerAddedOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OwnerManagerAddedOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OwnerManagerAddedOwner represents a AddedOwner event raised by the OwnerManager contract.
type OwnerManagerAddedOwner struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterAddedOwner is a free log retrieval operation binding the contract event 0x9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea26.
//
// Solidity: event AddedOwner(address owner)
func (_OwnerManager *OwnerManagerFilterer) FilterAddedOwner(opts *bind.FilterOpts) (*OwnerManagerAddedOwnerIterator, error) {

	logs, sub, err := _OwnerManager.contract.FilterLogs(opts, "AddedOwner")
	if err != nil {
		return nil, err
	}
	return &OwnerManagerAddedOwnerIterator{contract: _OwnerManager.contract, event: "AddedOwner", logs: logs, sub: sub}, nil
}

// WatchAddedOwner is a free log subscription operation binding the contract event 0x9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea26.
//
// Solidity: event AddedOwner(address owner)
func (_OwnerManager *OwnerManagerFilterer) WatchAddedOwner(opts *bind.WatchOpts, sink chan<- *OwnerManagerAddedOwner) (event.Subscription, error) {

	logs, sub, err := _OwnerManager.contract.WatchLogs(opts, "AddedOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OwnerManagerAddedOwner)
				if err := _OwnerManager.contract.UnpackLog(event, "AddedOwner", log); err != nil {
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

// ParseAddedOwner is a log parse operation binding the contract event 0x9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea26.
//
// Solidity: event AddedOwner(address owner)
func (_OwnerManager *OwnerManagerFilterer) ParseAddedOwner(log types.Log) (*OwnerManagerAddedOwner, error) {
	event := new(OwnerManagerAddedOwner)
	if err := _OwnerManager.contract.UnpackLog(event, "AddedOwner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OwnerManagerChangedThresholdIterator is returned from FilterChangedThreshold and is used to iterate over the raw logs and unpacked data for ChangedThreshold events raised by the OwnerManager contract.
type OwnerManagerChangedThresholdIterator struct {
	Event *OwnerManagerChangedThreshold // Event containing the contract specifics and raw log

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
func (it *OwnerManagerChangedThresholdIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OwnerManagerChangedThreshold)
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
		it.Event = new(OwnerManagerChangedThreshold)
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
func (it *OwnerManagerChangedThresholdIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OwnerManagerChangedThresholdIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OwnerManagerChangedThreshold represents a ChangedThreshold event raised by the OwnerManager contract.
type OwnerManagerChangedThreshold struct {
	Threshold *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterChangedThreshold is a free log retrieval operation binding the contract event 0x610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c93.
//
// Solidity: event ChangedThreshold(uint256 threshold)
func (_OwnerManager *OwnerManagerFilterer) FilterChangedThreshold(opts *bind.FilterOpts) (*OwnerManagerChangedThresholdIterator, error) {

	logs, sub, err := _OwnerManager.contract.FilterLogs(opts, "ChangedThreshold")
	if err != nil {
		return nil, err
	}
	return &OwnerManagerChangedThresholdIterator{contract: _OwnerManager.contract, event: "ChangedThreshold", logs: logs, sub: sub}, nil
}

// WatchChangedThreshold is a free log subscription operation binding the contract event 0x610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c93.
//
// Solidity: event ChangedThreshold(uint256 threshold)
func (_OwnerManager *OwnerManagerFilterer) WatchChangedThreshold(opts *bind.WatchOpts, sink chan<- *OwnerManagerChangedThreshold) (event.Subscription, error) {

	logs, sub, err := _OwnerManager.contract.WatchLogs(opts, "ChangedThreshold")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OwnerManagerChangedThreshold)
				if err := _OwnerManager.contract.UnpackLog(event, "ChangedThreshold", log); err != nil {
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

// ParseChangedThreshold is a log parse operation binding the contract event 0x610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c93.
//
// Solidity: event ChangedThreshold(uint256 threshold)
func (_OwnerManager *OwnerManagerFilterer) ParseChangedThreshold(log types.Log) (*OwnerManagerChangedThreshold, error) {
	event := new(OwnerManagerChangedThreshold)
	if err := _OwnerManager.contract.UnpackLog(event, "ChangedThreshold", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OwnerManagerRemovedOwnerIterator is returned from FilterRemovedOwner and is used to iterate over the raw logs and unpacked data for RemovedOwner events raised by the OwnerManager contract.
type OwnerManagerRemovedOwnerIterator struct {
	Event *OwnerManagerRemovedOwner // Event containing the contract specifics and raw log

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
func (it *OwnerManagerRemovedOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OwnerManagerRemovedOwner)
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
		it.Event = new(OwnerManagerRemovedOwner)
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
func (it *OwnerManagerRemovedOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OwnerManagerRemovedOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OwnerManagerRemovedOwner represents a RemovedOwner event raised by the OwnerManager contract.
type OwnerManagerRemovedOwner struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRemovedOwner is a free log retrieval operation binding the contract event 0xf8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf.
//
// Solidity: event RemovedOwner(address owner)
func (_OwnerManager *OwnerManagerFilterer) FilterRemovedOwner(opts *bind.FilterOpts) (*OwnerManagerRemovedOwnerIterator, error) {

	logs, sub, err := _OwnerManager.contract.FilterLogs(opts, "RemovedOwner")
	if err != nil {
		return nil, err
	}
	return &OwnerManagerRemovedOwnerIterator{contract: _OwnerManager.contract, event: "RemovedOwner", logs: logs, sub: sub}, nil
}

// WatchRemovedOwner is a free log subscription operation binding the contract event 0xf8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf.
//
// Solidity: event RemovedOwner(address owner)
func (_OwnerManager *OwnerManagerFilterer) WatchRemovedOwner(opts *bind.WatchOpts, sink chan<- *OwnerManagerRemovedOwner) (event.Subscription, error) {

	logs, sub, err := _OwnerManager.contract.WatchLogs(opts, "RemovedOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OwnerManagerRemovedOwner)
				if err := _OwnerManager.contract.UnpackLog(event, "RemovedOwner", log); err != nil {
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

// ParseRemovedOwner is a log parse operation binding the contract event 0xf8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf.
//
// Solidity: event RemovedOwner(address owner)
func (_OwnerManager *OwnerManagerFilterer) ParseRemovedOwner(log types.Log) (*OwnerManagerRemovedOwner, error) {
	event := new(OwnerManagerRemovedOwner)
	if err := _OwnerManager.contract.UnpackLog(event, "RemovedOwner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SecuredTokenTransferABI is the input ABI used to generate the binding from.
const SecuredTokenTransferABI = "[]"

// SecuredTokenTransferBin is the compiled bytecode used for deploying new contracts.
var SecuredTokenTransferBin = "0x6080604052348015600f57600080fd5b50603f80601d6000396000f3fe6080604052600080fdfea26469706673582212204e88d66db1f76911c3732ca9301dd45ffd885694eefef4b5095b8ded5b09343964736f6c63430007000033"

// DeploySecuredTokenTransfer deploys a new Ethereum contract, binding an instance of SecuredTokenTransfer to it.
func DeploySecuredTokenTransfer(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SecuredTokenTransfer, error) {
	parsed, err := abi.JSON(strings.NewReader(SecuredTokenTransferABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SecuredTokenTransferBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SecuredTokenTransfer{SecuredTokenTransferCaller: SecuredTokenTransferCaller{contract: contract}, SecuredTokenTransferTransactor: SecuredTokenTransferTransactor{contract: contract}, SecuredTokenTransferFilterer: SecuredTokenTransferFilterer{contract: contract}}, nil
}

// SecuredTokenTransfer is an auto generated Go binding around an Ethereum contract.
type SecuredTokenTransfer struct {
	SecuredTokenTransferCaller     // Read-only binding to the contract
	SecuredTokenTransferTransactor // Write-only binding to the contract
	SecuredTokenTransferFilterer   // Log filterer for contract events
}

// SecuredTokenTransferCaller is an auto generated read-only Go binding around an Ethereum contract.
type SecuredTokenTransferCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SecuredTokenTransferTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SecuredTokenTransferTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SecuredTokenTransferFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SecuredTokenTransferFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SecuredTokenTransferSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SecuredTokenTransferSession struct {
	Contract     *SecuredTokenTransfer // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// SecuredTokenTransferCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SecuredTokenTransferCallerSession struct {
	Contract *SecuredTokenTransferCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// SecuredTokenTransferTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SecuredTokenTransferTransactorSession struct {
	Contract     *SecuredTokenTransferTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// SecuredTokenTransferRaw is an auto generated low-level Go binding around an Ethereum contract.
type SecuredTokenTransferRaw struct {
	Contract *SecuredTokenTransfer // Generic contract binding to access the raw methods on
}

// SecuredTokenTransferCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SecuredTokenTransferCallerRaw struct {
	Contract *SecuredTokenTransferCaller // Generic read-only contract binding to access the raw methods on
}

// SecuredTokenTransferTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SecuredTokenTransferTransactorRaw struct {
	Contract *SecuredTokenTransferTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSecuredTokenTransfer creates a new instance of SecuredTokenTransfer, bound to a specific deployed contract.
func NewSecuredTokenTransfer(address common.Address, backend bind.ContractBackend) (*SecuredTokenTransfer, error) {
	contract, err := bindSecuredTokenTransfer(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SecuredTokenTransfer{SecuredTokenTransferCaller: SecuredTokenTransferCaller{contract: contract}, SecuredTokenTransferTransactor: SecuredTokenTransferTransactor{contract: contract}, SecuredTokenTransferFilterer: SecuredTokenTransferFilterer{contract: contract}}, nil
}

// NewSecuredTokenTransferCaller creates a new read-only instance of SecuredTokenTransfer, bound to a specific deployed contract.
func NewSecuredTokenTransferCaller(address common.Address, caller bind.ContractCaller) (*SecuredTokenTransferCaller, error) {
	contract, err := bindSecuredTokenTransfer(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SecuredTokenTransferCaller{contract: contract}, nil
}

// NewSecuredTokenTransferTransactor creates a new write-only instance of SecuredTokenTransfer, bound to a specific deployed contract.
func NewSecuredTokenTransferTransactor(address common.Address, transactor bind.ContractTransactor) (*SecuredTokenTransferTransactor, error) {
	contract, err := bindSecuredTokenTransfer(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SecuredTokenTransferTransactor{contract: contract}, nil
}

// NewSecuredTokenTransferFilterer creates a new log filterer instance of SecuredTokenTransfer, bound to a specific deployed contract.
func NewSecuredTokenTransferFilterer(address common.Address, filterer bind.ContractFilterer) (*SecuredTokenTransferFilterer, error) {
	contract, err := bindSecuredTokenTransfer(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SecuredTokenTransferFilterer{contract: contract}, nil
}

// bindSecuredTokenTransfer binds a generic wrapper to an already deployed contract.
func bindSecuredTokenTransfer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SecuredTokenTransferABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SecuredTokenTransfer *SecuredTokenTransferRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SecuredTokenTransfer.Contract.SecuredTokenTransferCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SecuredTokenTransfer *SecuredTokenTransferRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SecuredTokenTransfer.Contract.SecuredTokenTransferTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SecuredTokenTransfer *SecuredTokenTransferRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SecuredTokenTransfer.Contract.SecuredTokenTransferTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SecuredTokenTransfer *SecuredTokenTransferCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SecuredTokenTransfer.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SecuredTokenTransfer *SecuredTokenTransferTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SecuredTokenTransfer.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SecuredTokenTransfer *SecuredTokenTransferTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SecuredTokenTransfer.Contract.contract.Transact(opts, method, params...)
}

// SelfAuthorizedABI is the input ABI used to generate the binding from.
const SelfAuthorizedABI = "[]"

// SelfAuthorizedBin is the compiled bytecode used for deploying new contracts.
var SelfAuthorizedBin = "0x6080604052348015600f57600080fd5b50603f80601d6000396000f3fe6080604052600080fdfea264697066735822122057caa95aa4179b68fb7e99acbf02181b57c0f54af3052683c5b7f5cd0e8f738f64736f6c63430007000033"

// DeploySelfAuthorized deploys a new Ethereum contract, binding an instance of SelfAuthorized to it.
func DeploySelfAuthorized(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SelfAuthorized, error) {
	parsed, err := abi.JSON(strings.NewReader(SelfAuthorizedABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SelfAuthorizedBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SelfAuthorized{SelfAuthorizedCaller: SelfAuthorizedCaller{contract: contract}, SelfAuthorizedTransactor: SelfAuthorizedTransactor{contract: contract}, SelfAuthorizedFilterer: SelfAuthorizedFilterer{contract: contract}}, nil
}

// SelfAuthorized is an auto generated Go binding around an Ethereum contract.
type SelfAuthorized struct {
	SelfAuthorizedCaller     // Read-only binding to the contract
	SelfAuthorizedTransactor // Write-only binding to the contract
	SelfAuthorizedFilterer   // Log filterer for contract events
}

// SelfAuthorizedCaller is an auto generated read-only Go binding around an Ethereum contract.
type SelfAuthorizedCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SelfAuthorizedTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SelfAuthorizedTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SelfAuthorizedFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SelfAuthorizedFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SelfAuthorizedSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SelfAuthorizedSession struct {
	Contract     *SelfAuthorized   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SelfAuthorizedCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SelfAuthorizedCallerSession struct {
	Contract *SelfAuthorizedCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// SelfAuthorizedTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SelfAuthorizedTransactorSession struct {
	Contract     *SelfAuthorizedTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// SelfAuthorizedRaw is an auto generated low-level Go binding around an Ethereum contract.
type SelfAuthorizedRaw struct {
	Contract *SelfAuthorized // Generic contract binding to access the raw methods on
}

// SelfAuthorizedCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SelfAuthorizedCallerRaw struct {
	Contract *SelfAuthorizedCaller // Generic read-only contract binding to access the raw methods on
}

// SelfAuthorizedTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SelfAuthorizedTransactorRaw struct {
	Contract *SelfAuthorizedTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSelfAuthorized creates a new instance of SelfAuthorized, bound to a specific deployed contract.
func NewSelfAuthorized(address common.Address, backend bind.ContractBackend) (*SelfAuthorized, error) {
	contract, err := bindSelfAuthorized(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SelfAuthorized{SelfAuthorizedCaller: SelfAuthorizedCaller{contract: contract}, SelfAuthorizedTransactor: SelfAuthorizedTransactor{contract: contract}, SelfAuthorizedFilterer: SelfAuthorizedFilterer{contract: contract}}, nil
}

// NewSelfAuthorizedCaller creates a new read-only instance of SelfAuthorized, bound to a specific deployed contract.
func NewSelfAuthorizedCaller(address common.Address, caller bind.ContractCaller) (*SelfAuthorizedCaller, error) {
	contract, err := bindSelfAuthorized(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SelfAuthorizedCaller{contract: contract}, nil
}

// NewSelfAuthorizedTransactor creates a new write-only instance of SelfAuthorized, bound to a specific deployed contract.
func NewSelfAuthorizedTransactor(address common.Address, transactor bind.ContractTransactor) (*SelfAuthorizedTransactor, error) {
	contract, err := bindSelfAuthorized(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SelfAuthorizedTransactor{contract: contract}, nil
}

// NewSelfAuthorizedFilterer creates a new log filterer instance of SelfAuthorized, bound to a specific deployed contract.
func NewSelfAuthorizedFilterer(address common.Address, filterer bind.ContractFilterer) (*SelfAuthorizedFilterer, error) {
	contract, err := bindSelfAuthorized(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SelfAuthorizedFilterer{contract: contract}, nil
}

// bindSelfAuthorized binds a generic wrapper to an already deployed contract.
func bindSelfAuthorized(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SelfAuthorizedABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SelfAuthorized *SelfAuthorizedRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SelfAuthorized.Contract.SelfAuthorizedCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SelfAuthorized *SelfAuthorizedRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SelfAuthorized.Contract.SelfAuthorizedTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SelfAuthorized *SelfAuthorizedRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SelfAuthorized.Contract.SelfAuthorizedTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SelfAuthorized *SelfAuthorizedCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SelfAuthorized.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SelfAuthorized *SelfAuthorizedTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SelfAuthorized.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SelfAuthorized *SelfAuthorizedTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SelfAuthorized.Contract.contract.Transact(opts, method, params...)
}

// SignatureDecoderABI is the input ABI used to generate the binding from.
const SignatureDecoderABI = "[]"

// SignatureDecoderBin is the compiled bytecode used for deploying new contracts.
var SignatureDecoderBin = "0x6080604052348015600f57600080fd5b50603f80601d6000396000f3fe6080604052600080fdfea2646970667358221220806432e0f23b00d958914da0f81d40f89fada574f0058099516d5fd58d05b47064736f6c63430007000033"

// DeploySignatureDecoder deploys a new Ethereum contract, binding an instance of SignatureDecoder to it.
func DeploySignatureDecoder(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SignatureDecoder, error) {
	parsed, err := abi.JSON(strings.NewReader(SignatureDecoderABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SignatureDecoderBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SignatureDecoder{SignatureDecoderCaller: SignatureDecoderCaller{contract: contract}, SignatureDecoderTransactor: SignatureDecoderTransactor{contract: contract}, SignatureDecoderFilterer: SignatureDecoderFilterer{contract: contract}}, nil
}

// SignatureDecoder is an auto generated Go binding around an Ethereum contract.
type SignatureDecoder struct {
	SignatureDecoderCaller     // Read-only binding to the contract
	SignatureDecoderTransactor // Write-only binding to the contract
	SignatureDecoderFilterer   // Log filterer for contract events
}

// SignatureDecoderCaller is an auto generated read-only Go binding around an Ethereum contract.
type SignatureDecoderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SignatureDecoderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SignatureDecoderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SignatureDecoderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SignatureDecoderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SignatureDecoderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SignatureDecoderSession struct {
	Contract     *SignatureDecoder // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SignatureDecoderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SignatureDecoderCallerSession struct {
	Contract *SignatureDecoderCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// SignatureDecoderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SignatureDecoderTransactorSession struct {
	Contract     *SignatureDecoderTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// SignatureDecoderRaw is an auto generated low-level Go binding around an Ethereum contract.
type SignatureDecoderRaw struct {
	Contract *SignatureDecoder // Generic contract binding to access the raw methods on
}

// SignatureDecoderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SignatureDecoderCallerRaw struct {
	Contract *SignatureDecoderCaller // Generic read-only contract binding to access the raw methods on
}

// SignatureDecoderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SignatureDecoderTransactorRaw struct {
	Contract *SignatureDecoderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSignatureDecoder creates a new instance of SignatureDecoder, bound to a specific deployed contract.
func NewSignatureDecoder(address common.Address, backend bind.ContractBackend) (*SignatureDecoder, error) {
	contract, err := bindSignatureDecoder(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SignatureDecoder{SignatureDecoderCaller: SignatureDecoderCaller{contract: contract}, SignatureDecoderTransactor: SignatureDecoderTransactor{contract: contract}, SignatureDecoderFilterer: SignatureDecoderFilterer{contract: contract}}, nil
}

// NewSignatureDecoderCaller creates a new read-only instance of SignatureDecoder, bound to a specific deployed contract.
func NewSignatureDecoderCaller(address common.Address, caller bind.ContractCaller) (*SignatureDecoderCaller, error) {
	contract, err := bindSignatureDecoder(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SignatureDecoderCaller{contract: contract}, nil
}

// NewSignatureDecoderTransactor creates a new write-only instance of SignatureDecoder, bound to a specific deployed contract.
func NewSignatureDecoderTransactor(address common.Address, transactor bind.ContractTransactor) (*SignatureDecoderTransactor, error) {
	contract, err := bindSignatureDecoder(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SignatureDecoderTransactor{contract: contract}, nil
}

// NewSignatureDecoderFilterer creates a new log filterer instance of SignatureDecoder, bound to a specific deployed contract.
func NewSignatureDecoderFilterer(address common.Address, filterer bind.ContractFilterer) (*SignatureDecoderFilterer, error) {
	contract, err := bindSignatureDecoder(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SignatureDecoderFilterer{contract: contract}, nil
}

// bindSignatureDecoder binds a generic wrapper to an already deployed contract.
func bindSignatureDecoder(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SignatureDecoderABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SignatureDecoder *SignatureDecoderRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SignatureDecoder.Contract.SignatureDecoderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SignatureDecoder *SignatureDecoderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SignatureDecoder.Contract.SignatureDecoderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SignatureDecoder *SignatureDecoderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SignatureDecoder.Contract.SignatureDecoderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SignatureDecoder *SignatureDecoderCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SignatureDecoder.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SignatureDecoder *SignatureDecoderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SignatureDecoder.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SignatureDecoder *SignatureDecoderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SignatureDecoder.Contract.contract.Transact(opts, method, params...)
}

// SingletonABI is the input ABI used to generate the binding from.
const SingletonABI = "[]"

// SingletonBin is the compiled bytecode used for deploying new contracts.
var SingletonBin = "0x6080604052348015600f57600080fd5b50603f80601d6000396000f3fe6080604052600080fdfea2646970667358221220dbd0d3e059f56cd44c9561f0174e50a020aa0000f6f82ad4561a08186470a3ef64736f6c63430007000033"

// DeploySingleton deploys a new Ethereum contract, binding an instance of Singleton to it.
func DeploySingleton(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Singleton, error) {
	parsed, err := abi.JSON(strings.NewReader(SingletonABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SingletonBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Singleton{SingletonCaller: SingletonCaller{contract: contract}, SingletonTransactor: SingletonTransactor{contract: contract}, SingletonFilterer: SingletonFilterer{contract: contract}}, nil
}

// Singleton is an auto generated Go binding around an Ethereum contract.
type Singleton struct {
	SingletonCaller     // Read-only binding to the contract
	SingletonTransactor // Write-only binding to the contract
	SingletonFilterer   // Log filterer for contract events
}

// SingletonCaller is an auto generated read-only Go binding around an Ethereum contract.
type SingletonCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SingletonTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SingletonTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SingletonFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SingletonFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SingletonSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SingletonSession struct {
	Contract     *Singleton        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SingletonCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SingletonCallerSession struct {
	Contract *SingletonCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// SingletonTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SingletonTransactorSession struct {
	Contract     *SingletonTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// SingletonRaw is an auto generated low-level Go binding around an Ethereum contract.
type SingletonRaw struct {
	Contract *Singleton // Generic contract binding to access the raw methods on
}

// SingletonCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SingletonCallerRaw struct {
	Contract *SingletonCaller // Generic read-only contract binding to access the raw methods on
}

// SingletonTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SingletonTransactorRaw struct {
	Contract *SingletonTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSingleton creates a new instance of Singleton, bound to a specific deployed contract.
func NewSingleton(address common.Address, backend bind.ContractBackend) (*Singleton, error) {
	contract, err := bindSingleton(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Singleton{SingletonCaller: SingletonCaller{contract: contract}, SingletonTransactor: SingletonTransactor{contract: contract}, SingletonFilterer: SingletonFilterer{contract: contract}}, nil
}

// NewSingletonCaller creates a new read-only instance of Singleton, bound to a specific deployed contract.
func NewSingletonCaller(address common.Address, caller bind.ContractCaller) (*SingletonCaller, error) {
	contract, err := bindSingleton(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SingletonCaller{contract: contract}, nil
}

// NewSingletonTransactor creates a new write-only instance of Singleton, bound to a specific deployed contract.
func NewSingletonTransactor(address common.Address, transactor bind.ContractTransactor) (*SingletonTransactor, error) {
	contract, err := bindSingleton(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SingletonTransactor{contract: contract}, nil
}

// NewSingletonFilterer creates a new log filterer instance of Singleton, bound to a specific deployed contract.
func NewSingletonFilterer(address common.Address, filterer bind.ContractFilterer) (*SingletonFilterer, error) {
	contract, err := bindSingleton(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SingletonFilterer{contract: contract}, nil
}

// bindSingleton binds a generic wrapper to an already deployed contract.
func bindSingleton(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SingletonABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Singleton *SingletonRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Singleton.Contract.SingletonCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Singleton *SingletonRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Singleton.Contract.SingletonTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Singleton *SingletonRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Singleton.Contract.SingletonTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Singleton *SingletonCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Singleton.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Singleton *SingletonTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Singleton.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Singleton *SingletonTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Singleton.Contract.contract.Transact(opts, method, params...)
}

// StorageAccessibleABI is the input ABI used to generate the binding from.
const StorageAccessibleABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"getStorageAt\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"targetContract\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"calldataPayload\",\"type\":\"bytes\"}],\"name\":\"simulateDelegatecall\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"targetContract\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"calldataPayload\",\"type\":\"bytes\"}],\"name\":\"simulateDelegatecallInternal\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// StorageAccessibleFuncSigs maps the 4-byte function signature to its string representation.
var StorageAccessibleFuncSigs = map[string]string{
	"5624b25b": "getStorageAt(uint256,uint256)",
	"f84436bd": "simulateDelegatecall(address,bytes)",
	"43218e19": "simulateDelegatecallInternal(address,bytes)",
}

// StorageAccessibleBin is the compiled bytecode used for deploying new contracts.
var StorageAccessibleBin = "0x608060405234801561001057600080fd5b50610606806100206000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c806343218e19146100465780635624b25b14610171578063f84436bd14610194575b600080fd5b6100fc6004803603604081101561005c57600080fd5b6001600160a01b03823516919081019060408101602082013564010000000081111561008757600080fd5b82018360208201111561009957600080fd5b803590602001918460018302840111640100000000831117156100bb57600080fd5b91908080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525092955061024a945050505050565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561013657818101518382015260200161011e565b50505050905090810190601f1680156101635780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b6100fc6004803603604081101561018757600080fd5b5080359060200135610378565b6100fc600480360360408110156101aa57600080fd5b6001600160a01b0382351691908101906040810160208201356401000000008111156101d557600080fd5b8201836020820111156101e757600080fd5b8035906020019184600183028401116401000000008311171561020957600080fd5b91908080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509295506103ed945050505050565b606060006060846001600160a01b0316846040518082805190602001908083835b6020831061028a5780518252601f19909201916020918201910161026b565b6001836020036101000a038019825116818451168082178552505050505050905001915050600060405180830381855af49150503d80600081146102ea576040519150601f19603f3d011682016040523d82523d6000602084013e6102ef565b606091505b509150915061037081836040516020018083805190602001908083835b6020831061032b5780518252601f19909201916020918201910161030c565b6001836020036101000a03801982511681845116808217855250505050505090500182151560f81b8152600101925050506040516020818303038152906040526105c4565b505092915050565b6060808260200267ffffffffffffffff8111801561039557600080fd5b506040519080825280601f01601f1916602001820160405280156103c0576020820181803683370190505b50905060005b838110156103e357848101546020808302840101526001016103c6565b5090505b92915050565b6060807f43218e198a5f5c70ca65adf1973b6285a79c4d29a39cc2a8bb67b912f447dc64848460405160240180836001600160a01b0316815260200180602001828103825283818151815260200191508051906020019080838360005b8381101561046257818101518382015260200161044a565b50505050905090810190601f16801561048f5780820380516001836020036101000a031916815260200191505b5060408051601f198184030181529181526020820180516001600160e01b03166001600160e01b0319909816979097178752518151919750606096309650889550909350839250908083835b602083106104fa5780518252601f1990920191602091820191016104db565b6001836020036101000a0380198251168184511680821785525050505050509050019150506000604051808303816000865af19150503d806000811461055c576040519150601f19603f3d011682016040523d82523d6000602084013e610561565b606091505b5091505060008160018351038151811061057757fe5b602001015160f81c60f81b6001600160f81b031916600160f81b1490506105a28260018451036105cc565b80156105b2575091506103e79050565b6105bb826105c4565b50505092915050565b805160208201fd5b905256fea264697066735822122004c2230ae7be3e9db1251adca83e6244f264a48dcb771dc72b6d6bb82466865a64736f6c63430007000033"

// DeployStorageAccessible deploys a new Ethereum contract, binding an instance of StorageAccessible to it.
func DeployStorageAccessible(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *StorageAccessible, error) {
	parsed, err := abi.JSON(strings.NewReader(StorageAccessibleABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(StorageAccessibleBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &StorageAccessible{StorageAccessibleCaller: StorageAccessibleCaller{contract: contract}, StorageAccessibleTransactor: StorageAccessibleTransactor{contract: contract}, StorageAccessibleFilterer: StorageAccessibleFilterer{contract: contract}}, nil
}

// StorageAccessible is an auto generated Go binding around an Ethereum contract.
type StorageAccessible struct {
	StorageAccessibleCaller     // Read-only binding to the contract
	StorageAccessibleTransactor // Write-only binding to the contract
	StorageAccessibleFilterer   // Log filterer for contract events
}

// StorageAccessibleCaller is an auto generated read-only Go binding around an Ethereum contract.
type StorageAccessibleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StorageAccessibleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type StorageAccessibleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StorageAccessibleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type StorageAccessibleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StorageAccessibleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type StorageAccessibleSession struct {
	Contract     *StorageAccessible // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// StorageAccessibleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type StorageAccessibleCallerSession struct {
	Contract *StorageAccessibleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// StorageAccessibleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type StorageAccessibleTransactorSession struct {
	Contract     *StorageAccessibleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// StorageAccessibleRaw is an auto generated low-level Go binding around an Ethereum contract.
type StorageAccessibleRaw struct {
	Contract *StorageAccessible // Generic contract binding to access the raw methods on
}

// StorageAccessibleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type StorageAccessibleCallerRaw struct {
	Contract *StorageAccessibleCaller // Generic read-only contract binding to access the raw methods on
}

// StorageAccessibleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type StorageAccessibleTransactorRaw struct {
	Contract *StorageAccessibleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStorageAccessible creates a new instance of StorageAccessible, bound to a specific deployed contract.
func NewStorageAccessible(address common.Address, backend bind.ContractBackend) (*StorageAccessible, error) {
	contract, err := bindStorageAccessible(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &StorageAccessible{StorageAccessibleCaller: StorageAccessibleCaller{contract: contract}, StorageAccessibleTransactor: StorageAccessibleTransactor{contract: contract}, StorageAccessibleFilterer: StorageAccessibleFilterer{contract: contract}}, nil
}

// NewStorageAccessibleCaller creates a new read-only instance of StorageAccessible, bound to a specific deployed contract.
func NewStorageAccessibleCaller(address common.Address, caller bind.ContractCaller) (*StorageAccessibleCaller, error) {
	contract, err := bindStorageAccessible(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StorageAccessibleCaller{contract: contract}, nil
}

// NewStorageAccessibleTransactor creates a new write-only instance of StorageAccessible, bound to a specific deployed contract.
func NewStorageAccessibleTransactor(address common.Address, transactor bind.ContractTransactor) (*StorageAccessibleTransactor, error) {
	contract, err := bindStorageAccessible(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StorageAccessibleTransactor{contract: contract}, nil
}

// NewStorageAccessibleFilterer creates a new log filterer instance of StorageAccessible, bound to a specific deployed contract.
func NewStorageAccessibleFilterer(address common.Address, filterer bind.ContractFilterer) (*StorageAccessibleFilterer, error) {
	contract, err := bindStorageAccessible(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StorageAccessibleFilterer{contract: contract}, nil
}

// bindStorageAccessible binds a generic wrapper to an already deployed contract.
func bindStorageAccessible(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(StorageAccessibleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StorageAccessible *StorageAccessibleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StorageAccessible.Contract.StorageAccessibleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StorageAccessible *StorageAccessibleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StorageAccessible.Contract.StorageAccessibleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StorageAccessible *StorageAccessibleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StorageAccessible.Contract.StorageAccessibleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StorageAccessible *StorageAccessibleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StorageAccessible.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StorageAccessible *StorageAccessibleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StorageAccessible.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StorageAccessible *StorageAccessibleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StorageAccessible.Contract.contract.Transact(opts, method, params...)
}

// GetStorageAt is a free data retrieval call binding the contract method 0x5624b25b.
//
// Solidity: function getStorageAt(uint256 offset, uint256 length) view returns(bytes)
func (_StorageAccessible *StorageAccessibleCaller) GetStorageAt(opts *bind.CallOpts, offset *big.Int, length *big.Int) ([]byte, error) {
	var out []interface{}
	err := _StorageAccessible.contract.Call(opts, &out, "getStorageAt", offset, length)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetStorageAt is a free data retrieval call binding the contract method 0x5624b25b.
//
// Solidity: function getStorageAt(uint256 offset, uint256 length) view returns(bytes)
func (_StorageAccessible *StorageAccessibleSession) GetStorageAt(offset *big.Int, length *big.Int) ([]byte, error) {
	return _StorageAccessible.Contract.GetStorageAt(&_StorageAccessible.CallOpts, offset, length)
}

// GetStorageAt is a free data retrieval call binding the contract method 0x5624b25b.
//
// Solidity: function getStorageAt(uint256 offset, uint256 length) view returns(bytes)
func (_StorageAccessible *StorageAccessibleCallerSession) GetStorageAt(offset *big.Int, length *big.Int) ([]byte, error) {
	return _StorageAccessible.Contract.GetStorageAt(&_StorageAccessible.CallOpts, offset, length)
}

// SimulateDelegatecall is a paid mutator transaction binding the contract method 0xf84436bd.
//
// Solidity: function simulateDelegatecall(address targetContract, bytes calldataPayload) returns(bytes)
func (_StorageAccessible *StorageAccessibleTransactor) SimulateDelegatecall(opts *bind.TransactOpts, targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _StorageAccessible.contract.Transact(opts, "simulateDelegatecall", targetContract, calldataPayload)
}

// SimulateDelegatecall is a paid mutator transaction binding the contract method 0xf84436bd.
//
// Solidity: function simulateDelegatecall(address targetContract, bytes calldataPayload) returns(bytes)
func (_StorageAccessible *StorageAccessibleSession) SimulateDelegatecall(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _StorageAccessible.Contract.SimulateDelegatecall(&_StorageAccessible.TransactOpts, targetContract, calldataPayload)
}

// SimulateDelegatecall is a paid mutator transaction binding the contract method 0xf84436bd.
//
// Solidity: function simulateDelegatecall(address targetContract, bytes calldataPayload) returns(bytes)
func (_StorageAccessible *StorageAccessibleTransactorSession) SimulateDelegatecall(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _StorageAccessible.Contract.SimulateDelegatecall(&_StorageAccessible.TransactOpts, targetContract, calldataPayload)
}

// SimulateDelegatecallInternal is a paid mutator transaction binding the contract method 0x43218e19.
//
// Solidity: function simulateDelegatecallInternal(address targetContract, bytes calldataPayload) returns(bytes)
func (_StorageAccessible *StorageAccessibleTransactor) SimulateDelegatecallInternal(opts *bind.TransactOpts, targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _StorageAccessible.contract.Transact(opts, "simulateDelegatecallInternal", targetContract, calldataPayload)
}

// SimulateDelegatecallInternal is a paid mutator transaction binding the contract method 0x43218e19.
//
// Solidity: function simulateDelegatecallInternal(address targetContract, bytes calldataPayload) returns(bytes)
func (_StorageAccessible *StorageAccessibleSession) SimulateDelegatecallInternal(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _StorageAccessible.Contract.SimulateDelegatecallInternal(&_StorageAccessible.TransactOpts, targetContract, calldataPayload)
}

// SimulateDelegatecallInternal is a paid mutator transaction binding the contract method 0x43218e19.
//
// Solidity: function simulateDelegatecallInternal(address targetContract, bytes calldataPayload) returns(bytes)
func (_StorageAccessible *StorageAccessibleTransactorSession) SimulateDelegatecallInternal(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _StorageAccessible.Contract.SimulateDelegatecallInternal(&_StorageAccessible.TransactOpts, targetContract, calldataPayload)
}
