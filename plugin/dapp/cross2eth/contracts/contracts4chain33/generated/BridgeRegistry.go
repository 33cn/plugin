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

// BridgeRegistryABI is the input ABI used to generate the binding from.
const BridgeRegistryABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_ethereumBridge\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_bridgeBank\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_oracle\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_valset\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_ethereumBridge\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_bridgeBank\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_oracle\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_valset\",\"type\":\"address\"}],\"name\":\"LogContractsRegistered\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[],\"name\":\"bridgeBank\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"deployHeight\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ethereumBridge\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"oracle\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"valset\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// BridgeRegistryFuncSigs maps the 4-byte function signature to its string representation.
var BridgeRegistryFuncSigs = map[string]string{
	"0e41f373": "bridgeBank()",
	"53d953b5": "deployHeight()",
	"1cf86a63": "ethereumBridge()",
	"7dc0d1d0": "oracle()",
	"7f54af0c": "valset()",
}

// BridgeRegistryBin is the compiled bytecode used for deploying new contracts.
var BridgeRegistryBin = "0x608060405234801561001057600080fd5b5060405161021b38038061021b8339818101604052608081101561003357600080fd5b508051602080830151604080850151606095860151600080546001600160a01b038089166001600160a01b031992831617928390556001805482891690841617908190556002805483881690851617908190556003805484881695169490941793849055436004558751948316855290821698840198909852968716828601529095169685019690965290519394919390927f039b733f31259b106f1d278c726870d5b28c7db22957d63df8dbaa70bd3a032a919081900360800190a150505050610118806101036000396000f3fe6080604052348015600f57600080fd5b506004361060505760003560e01c80630e41f3731460555780631cf86a6314607757806353d953b514607d5780637dc0d1d01460955780637f54af0c14609b575b600080fd5b605b60a1565b604080516001600160a01b039092168252519081900360200190f35b605b60b0565b608360bf565b60408051918252519081900360200190f35b605b60c5565b605b60d4565b6001546001600160a01b031681565b6000546001600160a01b031681565b60045481565b6002546001600160a01b031681565b6003546001600160a01b03168156fea265627a7a72315820de7778aa9d4b720d9ffcd64c45b73b8a23dfcdb9d7a658c8aba48310cd17ab7164736f6c63430005100032"

// DeployBridgeRegistry deploys a new Ethereum contract, binding an instance of BridgeRegistry to it.
func DeployBridgeRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, _ethereumBridge common.Address, _bridgeBank common.Address, _oracle common.Address, _valset common.Address) (common.Address, *types.Transaction, *BridgeRegistry, error) {
	parsed, err := abi.JSON(strings.NewReader(BridgeRegistryABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(BridgeRegistryBin), backend, _ethereumBridge, _bridgeBank, _oracle, _valset)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &BridgeRegistry{BridgeRegistryCaller: BridgeRegistryCaller{contract: contract}, BridgeRegistryTransactor: BridgeRegistryTransactor{contract: contract}, BridgeRegistryFilterer: BridgeRegistryFilterer{contract: contract}}, nil
}

// BridgeRegistry is an auto generated Go binding around an Ethereum contract.
type BridgeRegistry struct {
	BridgeRegistryCaller     // Read-only binding to the contract
	BridgeRegistryTransactor // Write-only binding to the contract
	BridgeRegistryFilterer   // Log filterer for contract events
}

// BridgeRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type BridgeRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BridgeRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BridgeRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BridgeRegistrySession struct {
	Contract     *BridgeRegistry   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BridgeRegistryCallerSession struct {
	Contract *BridgeRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// BridgeRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BridgeRegistryTransactorSession struct {
	Contract     *BridgeRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// BridgeRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type BridgeRegistryRaw struct {
	Contract *BridgeRegistry // Generic contract binding to access the raw methods on
}

// BridgeRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BridgeRegistryCallerRaw struct {
	Contract *BridgeRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// BridgeRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BridgeRegistryTransactorRaw struct {
	Contract *BridgeRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBridgeRegistry creates a new instance of BridgeRegistry, bound to a specific deployed contract.
func NewBridgeRegistry(address common.Address, backend bind.ContractBackend) (*BridgeRegistry, error) {
	contract, err := bindBridgeRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BridgeRegistry{BridgeRegistryCaller: BridgeRegistryCaller{contract: contract}, BridgeRegistryTransactor: BridgeRegistryTransactor{contract: contract}, BridgeRegistryFilterer: BridgeRegistryFilterer{contract: contract}}, nil
}

// NewBridgeRegistryCaller creates a new read-only instance of BridgeRegistry, bound to a specific deployed contract.
func NewBridgeRegistryCaller(address common.Address, caller bind.ContractCaller) (*BridgeRegistryCaller, error) {
	contract, err := bindBridgeRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeRegistryCaller{contract: contract}, nil
}

// NewBridgeRegistryTransactor creates a new write-only instance of BridgeRegistry, bound to a specific deployed contract.
func NewBridgeRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*BridgeRegistryTransactor, error) {
	contract, err := bindBridgeRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeRegistryTransactor{contract: contract}, nil
}

// NewBridgeRegistryFilterer creates a new log filterer instance of BridgeRegistry, bound to a specific deployed contract.
func NewBridgeRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*BridgeRegistryFilterer, error) {
	contract, err := bindBridgeRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BridgeRegistryFilterer{contract: contract}, nil
}

// bindBridgeRegistry binds a generic wrapper to an already deployed contract.
func bindBridgeRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BridgeRegistryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BridgeRegistry *BridgeRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BridgeRegistry.Contract.BridgeRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BridgeRegistry *BridgeRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BridgeRegistry.Contract.BridgeRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BridgeRegistry *BridgeRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BridgeRegistry.Contract.BridgeRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BridgeRegistry *BridgeRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BridgeRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BridgeRegistry *BridgeRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BridgeRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BridgeRegistry *BridgeRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BridgeRegistry.Contract.contract.Transact(opts, method, params...)
}

// BridgeBank is a free data retrieval call binding the contract method 0x0e41f373.
//
// Solidity: function bridgeBank() view returns(address)
func (_BridgeRegistry *BridgeRegistryCaller) BridgeBank(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BridgeRegistry.contract.Call(opts, &out, "bridgeBank")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BridgeBank is a free data retrieval call binding the contract method 0x0e41f373.
//
// Solidity: function bridgeBank() view returns(address)
func (_BridgeRegistry *BridgeRegistrySession) BridgeBank() (common.Address, error) {
	return _BridgeRegistry.Contract.BridgeBank(&_BridgeRegistry.CallOpts)
}

// BridgeBank is a free data retrieval call binding the contract method 0x0e41f373.
//
// Solidity: function bridgeBank() view returns(address)
func (_BridgeRegistry *BridgeRegistryCallerSession) BridgeBank() (common.Address, error) {
	return _BridgeRegistry.Contract.BridgeBank(&_BridgeRegistry.CallOpts)
}

// DeployHeight is a free data retrieval call binding the contract method 0x53d953b5.
//
// Solidity: function deployHeight() view returns(uint256)
func (_BridgeRegistry *BridgeRegistryCaller) DeployHeight(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BridgeRegistry.contract.Call(opts, &out, "deployHeight")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DeployHeight is a free data retrieval call binding the contract method 0x53d953b5.
//
// Solidity: function deployHeight() view returns(uint256)
func (_BridgeRegistry *BridgeRegistrySession) DeployHeight() (*big.Int, error) {
	return _BridgeRegistry.Contract.DeployHeight(&_BridgeRegistry.CallOpts)
}

// DeployHeight is a free data retrieval call binding the contract method 0x53d953b5.
//
// Solidity: function deployHeight() view returns(uint256)
func (_BridgeRegistry *BridgeRegistryCallerSession) DeployHeight() (*big.Int, error) {
	return _BridgeRegistry.Contract.DeployHeight(&_BridgeRegistry.CallOpts)
}

// EthereumBridge is a free data retrieval call binding the contract method 0x1cf86a63.
//
// Solidity: function ethereumBridge() view returns(address)
func (_BridgeRegistry *BridgeRegistryCaller) EthereumBridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BridgeRegistry.contract.Call(opts, &out, "ethereumBridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EthereumBridge is a free data retrieval call binding the contract method 0x1cf86a63.
//
// Solidity: function ethereumBridge() view returns(address)
func (_BridgeRegistry *BridgeRegistrySession) EthereumBridge() (common.Address, error) {
	return _BridgeRegistry.Contract.EthereumBridge(&_BridgeRegistry.CallOpts)
}

// EthereumBridge is a free data retrieval call binding the contract method 0x1cf86a63.
//
// Solidity: function ethereumBridge() view returns(address)
func (_BridgeRegistry *BridgeRegistryCallerSession) EthereumBridge() (common.Address, error) {
	return _BridgeRegistry.Contract.EthereumBridge(&_BridgeRegistry.CallOpts)
}

// Oracle is a free data retrieval call binding the contract method 0x7dc0d1d0.
//
// Solidity: function oracle() view returns(address)
func (_BridgeRegistry *BridgeRegistryCaller) Oracle(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BridgeRegistry.contract.Call(opts, &out, "oracle")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Oracle is a free data retrieval call binding the contract method 0x7dc0d1d0.
//
// Solidity: function oracle() view returns(address)
func (_BridgeRegistry *BridgeRegistrySession) Oracle() (common.Address, error) {
	return _BridgeRegistry.Contract.Oracle(&_BridgeRegistry.CallOpts)
}

// Oracle is a free data retrieval call binding the contract method 0x7dc0d1d0.
//
// Solidity: function oracle() view returns(address)
func (_BridgeRegistry *BridgeRegistryCallerSession) Oracle() (common.Address, error) {
	return _BridgeRegistry.Contract.Oracle(&_BridgeRegistry.CallOpts)
}

// Valset is a free data retrieval call binding the contract method 0x7f54af0c.
//
// Solidity: function valset() view returns(address)
func (_BridgeRegistry *BridgeRegistryCaller) Valset(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BridgeRegistry.contract.Call(opts, &out, "valset")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Valset is a free data retrieval call binding the contract method 0x7f54af0c.
//
// Solidity: function valset() view returns(address)
func (_BridgeRegistry *BridgeRegistrySession) Valset() (common.Address, error) {
	return _BridgeRegistry.Contract.Valset(&_BridgeRegistry.CallOpts)
}

// Valset is a free data retrieval call binding the contract method 0x7f54af0c.
//
// Solidity: function valset() view returns(address)
func (_BridgeRegistry *BridgeRegistryCallerSession) Valset() (common.Address, error) {
	return _BridgeRegistry.Contract.Valset(&_BridgeRegistry.CallOpts)
}

// BridgeRegistryLogContractsRegisteredIterator is returned from FilterLogContractsRegistered and is used to iterate over the raw logs and unpacked data for LogContractsRegistered events raised by the BridgeRegistry contract.
type BridgeRegistryLogContractsRegisteredIterator struct {
	Event *BridgeRegistryLogContractsRegistered // Event containing the contract specifics and raw log

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
func (it *BridgeRegistryLogContractsRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeRegistryLogContractsRegistered)
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
		it.Event = new(BridgeRegistryLogContractsRegistered)
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
func (it *BridgeRegistryLogContractsRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeRegistryLogContractsRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeRegistryLogContractsRegistered represents a LogContractsRegistered event raised by the BridgeRegistry contract.
type BridgeRegistryLogContractsRegistered struct {
	EthereumBridge common.Address
	BridgeBank     common.Address
	Oracle         common.Address
	Valset         common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterLogContractsRegistered is a free log retrieval operation binding the contract event 0x039b733f31259b106f1d278c726870d5b28c7db22957d63df8dbaa70bd3a032a.
//
// Solidity: event LogContractsRegistered(address _ethereumBridge, address _bridgeBank, address _oracle, address _valset)
func (_BridgeRegistry *BridgeRegistryFilterer) FilterLogContractsRegistered(opts *bind.FilterOpts) (*BridgeRegistryLogContractsRegisteredIterator, error) {

	logs, sub, err := _BridgeRegistry.contract.FilterLogs(opts, "LogContractsRegistered")
	if err != nil {
		return nil, err
	}
	return &BridgeRegistryLogContractsRegisteredIterator{contract: _BridgeRegistry.contract, event: "LogContractsRegistered", logs: logs, sub: sub}, nil
}

// WatchLogContractsRegistered is a free log subscription operation binding the contract event 0x039b733f31259b106f1d278c726870d5b28c7db22957d63df8dbaa70bd3a032a.
//
// Solidity: event LogContractsRegistered(address _ethereumBridge, address _bridgeBank, address _oracle, address _valset)
func (_BridgeRegistry *BridgeRegistryFilterer) WatchLogContractsRegistered(opts *bind.WatchOpts, sink chan<- *BridgeRegistryLogContractsRegistered) (event.Subscription, error) {

	logs, sub, err := _BridgeRegistry.contract.WatchLogs(opts, "LogContractsRegistered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeRegistryLogContractsRegistered)
				if err := _BridgeRegistry.contract.UnpackLog(event, "LogContractsRegistered", log); err != nil {
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

// ParseLogContractsRegistered is a log parse operation binding the contract event 0x039b733f31259b106f1d278c726870d5b28c7db22957d63df8dbaa70bd3a032a.
//
// Solidity: event LogContractsRegistered(address _ethereumBridge, address _bridgeBank, address _oracle, address _valset)
func (_BridgeRegistry *BridgeRegistryFilterer) ParseLogContractsRegistered(log types.Log) (*BridgeRegistryLogContractsRegistered, error) {
	event := new(BridgeRegistryLogContractsRegistered)
	if err := _BridgeRegistry.contract.UnpackLog(event, "LogContractsRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
