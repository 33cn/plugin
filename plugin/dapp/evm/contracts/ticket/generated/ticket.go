// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ticket

import (
	"errors"
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
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// PreTicketMetaData contains all meta data concerning the PreTicket contract.
var PreTicketMetaData = &bind.MetaData{
	ABI: "[]",
	Bin: "0x60566037600b82828239805160001a607314602a57634e487b7160e01b600052600060045260246000fd5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea2646970667358221220c5b9bf7bc1becaf40d51f6f1c0334bcb66f244437a5e61e72c1dfa0f6b9b87d264736f6c634300080a0033",
}

// PreTicketABI is the input ABI used to generate the binding from.
// Deprecated: Use PreTicketMetaData.ABI instead.
var PreTicketABI = PreTicketMetaData.ABI

// PreTicketBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use PreTicketMetaData.Bin instead.
var PreTicketBin = PreTicketMetaData.Bin

// DeployPreTicket deploys a new Ethereum contract, binding an instance of PreTicket to it.
func DeployPreTicket(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *PreTicket, error) {
	parsed, err := PreTicketMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(PreTicketBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PreTicket{PreTicketCaller: PreTicketCaller{contract: contract}, PreTicketTransactor: PreTicketTransactor{contract: contract}, PreTicketFilterer: PreTicketFilterer{contract: contract}}, nil
}

// PreTicket is an auto generated Go binding around an Ethereum contract.
type PreTicket struct {
	PreTicketCaller     // Read-only binding to the contract
	PreTicketTransactor // Write-only binding to the contract
	PreTicketFilterer   // Log filterer for contract events
}

// PreTicketCaller is an auto generated read-only Go binding around an Ethereum contract.
type PreTicketCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PreTicketTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PreTicketTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PreTicketFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PreTicketFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PreTicketSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PreTicketSession struct {
	Contract     *PreTicket        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PreTicketCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PreTicketCallerSession struct {
	Contract *PreTicketCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// PreTicketTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PreTicketTransactorSession struct {
	Contract     *PreTicketTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// PreTicketRaw is an auto generated low-level Go binding around an Ethereum contract.
type PreTicketRaw struct {
	Contract *PreTicket // Generic contract binding to access the raw methods on
}

// PreTicketCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PreTicketCallerRaw struct {
	Contract *PreTicketCaller // Generic read-only contract binding to access the raw methods on
}

// PreTicketTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PreTicketTransactorRaw struct {
	Contract *PreTicketTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPreTicket creates a new instance of PreTicket, bound to a specific deployed contract.
func NewPreTicket(address common.Address, backend bind.ContractBackend) (*PreTicket, error) {
	contract, err := bindPreTicket(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PreTicket{PreTicketCaller: PreTicketCaller{contract: contract}, PreTicketTransactor: PreTicketTransactor{contract: contract}, PreTicketFilterer: PreTicketFilterer{contract: contract}}, nil
}

// NewPreTicketCaller creates a new read-only instance of PreTicket, bound to a specific deployed contract.
func NewPreTicketCaller(address common.Address, caller bind.ContractCaller) (*PreTicketCaller, error) {
	contract, err := bindPreTicket(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PreTicketCaller{contract: contract}, nil
}

// NewPreTicketTransactor creates a new write-only instance of PreTicket, bound to a specific deployed contract.
func NewPreTicketTransactor(address common.Address, transactor bind.ContractTransactor) (*PreTicketTransactor, error) {
	contract, err := bindPreTicket(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PreTicketTransactor{contract: contract}, nil
}

// NewPreTicketFilterer creates a new log filterer instance of PreTicket, bound to a specific deployed contract.
func NewPreTicketFilterer(address common.Address, filterer bind.ContractFilterer) (*PreTicketFilterer, error) {
	contract, err := bindPreTicket(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PreTicketFilterer{contract: contract}, nil
}

// bindPreTicket binds a generic wrapper to an already deployed contract.
func bindPreTicket(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PreTicketABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PreTicket *PreTicketRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PreTicket.Contract.PreTicketCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PreTicket *PreTicketRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PreTicket.Contract.PreTicketTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PreTicket *PreTicketRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PreTicket.Contract.PreTicketTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PreTicket *PreTicketCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PreTicket.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PreTicket *PreTicketTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PreTicket.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PreTicket *PreTicketTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PreTicket.Contract.contract.Transact(opts, method, params...)
}

// TicketMetaData contains all meta data concerning the Ticket contract.
var TicketMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"origin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"bind\",\"type\":\"address\"}],\"name\":\"CloseBindMiner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"origin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"bind\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"CreateBindMiner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"TransferToTicketExec\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bind\",\"type\":\"address\"}],\"name\":\"closeBindMiner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bind\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"createBindMiner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTicketCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferToTickeExec\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"57385d1b": "closeBindMiner(address)",
		"5ca4b8fb": "createBindMiner(address,uint256)",
		"21c63a47": "getTicketCount()",
		"6850d595": "transferToTickeExec(uint256)",
	},
	Bin: "0x608060405234801561001057600080fd5b506106a5806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c806321c63a471461005157806357385d1b1461006c5780635ca4b8fb1461008f5780636850d595146100a2575b600080fd5b6100596100b5565b6040519081526020015b60405180910390f35b61007f61007a366004610568565b6100c4565b6040519015158152602001610063565b61007f61009d36600461058a565b610148565b61007f6100b03660046105b4565b61022d565b60006100bf6102e6565b905090565b60006001600160a01b0382166100f55760405162461bcd60e51b81526004016100ec906105cd565b60405180910390fd5b6100ff8233610396565b604080513381526001600160a01b03841660208201527f1e1e210b9e04215b745006f4b006a0b6f420f87563005fb9baa0e7407b8f5da9910160405180910390a1506001919050565b60006001600160a01b0383166101705760405162461bcd60e51b81526004016100ec906105cd565b816101cc5760405162461bcd60e51b815260206004820152602660248201527f5469636b65743a2063726561746542696e644d696e657220616d6f756e74206960448201526573207a65726f60d01b60648201526084016100ec565b336101d8818585610430565b604080516001600160a01b038084168252861660208201529081018490527ff68d6e786351c6914a35e97ea2605fb495e868c5f8e81f930394d7acbe8f7f3b9060600160405180910390a15060019392505050565b6000816102905760405162461bcd60e51b815260206004820152602b60248201527f5469636b65743a207472616e73666572546f5469636b65744578656320616d6f60448201526a756e74206973207a65726f60a81b60648201526084016100ec565b3361029b81846104f1565b604080516001600160a01b0383168152602081018590527fa96959215ce46afdbce738a9b441a353e38c3c54c71c2894bfe6c6d4ccf10f18910160405180910390a150600192915050565b60408051600481526024810182526020810180516001600160e01b03166321c63a4760e01b1790529051600091829182916220000291610326919061061b565b600060405180830381855afa9150503d8060008114610361576040519150601f19603f3d011682016040523d82523d6000602084013e610366565b606091505b5091509150600082141561037b573d60208201fd5b8080602001905181019061038f9190610656565b9250505090565b60408051600481526024810182526020810180516001600160e01b031663474f72d360e01b1790529051600091829162200002916103d39161061b565b6000604051808303816000865af19150503d8060008114610410576040519150601f19603f3d011682016040523d82523d6000602084013e610415565b606091505b5091509150600082141561042a573d60208201fd5b50505050565b6040516001600160a01b03848116602483015283166044820152606481018290526000908190622000029060840160408051601f198184030181529181526020820180516001600160e01b0316632dd8308f60e11b17905251610493919061061b565b6000604051808303816000865af19150503d80600081146104d0576040519150601f19603f3d011682016040523d82523d6000602084013e6104d5565b606091505b509150915060008214156104ea573d60208201fd5b5050505050565b6040516001600160a01b0383166024820152604481018290526000908190622000029060640160408051601f198184030181529181526020820180516001600160e01b0316632a80b0ab60e01b179052516103d3919061061b565b80356001600160a01b038116811461056357600080fd5b919050565b60006020828403121561057a57600080fd5b6105838261054c565b9392505050565b6000806040838503121561059d57600080fd5b6105a68361054c565b946020939093013593505050565b6000602082840312156105c657600080fd5b5035919050565b6020808252602e908201527f5469636b65743a2063726561746542696e644d696e65722066726f6d2074686560408201526d20207a65726f206164647265737360901b606082015260800190565b6000825160005b8181101561063c5760208186018101518583015201610622565b8181111561064b576000828501525b509190910192915050565b60006020828403121561066857600080fd5b505191905056fea264697066735822122006fc5fc9c646adb73af5389e73c1a83fcfeb105d17d2fab72665d71091b0c18c64736f6c634300080a0033",
}

// TicketABI is the input ABI used to generate the binding from.
// Deprecated: Use TicketMetaData.ABI instead.
var TicketABI = TicketMetaData.ABI

// Deprecated: Use TicketMetaData.Sigs instead.
// TicketFuncSigs maps the 4-byte function signature to its string representation.
var TicketFuncSigs = TicketMetaData.Sigs

// TicketBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use TicketMetaData.Bin instead.
var TicketBin = TicketMetaData.Bin

// DeployTicket deploys a new Ethereum contract, binding an instance of Ticket to it.
func DeployTicket(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Ticket, error) {
	parsed, err := TicketMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(TicketBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Ticket{TicketCaller: TicketCaller{contract: contract}, TicketTransactor: TicketTransactor{contract: contract}, TicketFilterer: TicketFilterer{contract: contract}}, nil
}

// Ticket is an auto generated Go binding around an Ethereum contract.
type Ticket struct {
	TicketCaller     // Read-only binding to the contract
	TicketTransactor // Write-only binding to the contract
	TicketFilterer   // Log filterer for contract events
}

// TicketCaller is an auto generated read-only Go binding around an Ethereum contract.
type TicketCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TicketTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TicketTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TicketFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TicketFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TicketSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TicketSession struct {
	Contract     *Ticket           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TicketCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TicketCallerSession struct {
	Contract *TicketCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// TicketTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TicketTransactorSession struct {
	Contract     *TicketTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TicketRaw is an auto generated low-level Go binding around an Ethereum contract.
type TicketRaw struct {
	Contract *Ticket // Generic contract binding to access the raw methods on
}

// TicketCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TicketCallerRaw struct {
	Contract *TicketCaller // Generic read-only contract binding to access the raw methods on
}

// TicketTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TicketTransactorRaw struct {
	Contract *TicketTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTicket creates a new instance of Ticket, bound to a specific deployed contract.
func NewTicket(address common.Address, backend bind.ContractBackend) (*Ticket, error) {
	contract, err := bindTicket(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Ticket{TicketCaller: TicketCaller{contract: contract}, TicketTransactor: TicketTransactor{contract: contract}, TicketFilterer: TicketFilterer{contract: contract}}, nil
}

// NewTicketCaller creates a new read-only instance of Ticket, bound to a specific deployed contract.
func NewTicketCaller(address common.Address, caller bind.ContractCaller) (*TicketCaller, error) {
	contract, err := bindTicket(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TicketCaller{contract: contract}, nil
}

// NewTicketTransactor creates a new write-only instance of Ticket, bound to a specific deployed contract.
func NewTicketTransactor(address common.Address, transactor bind.ContractTransactor) (*TicketTransactor, error) {
	contract, err := bindTicket(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TicketTransactor{contract: contract}, nil
}

// NewTicketFilterer creates a new log filterer instance of Ticket, bound to a specific deployed contract.
func NewTicketFilterer(address common.Address, filterer bind.ContractFilterer) (*TicketFilterer, error) {
	contract, err := bindTicket(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TicketFilterer{contract: contract}, nil
}

// bindTicket binds a generic wrapper to an already deployed contract.
func bindTicket(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TicketABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ticket *TicketRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ticket.Contract.TicketCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ticket *TicketRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ticket.Contract.TicketTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ticket *TicketRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ticket.Contract.TicketTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ticket *TicketCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ticket.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ticket *TicketTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ticket.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ticket *TicketTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ticket.Contract.contract.Transact(opts, method, params...)
}

// CloseBindMiner is a paid mutator transaction binding the contract method 0x57385d1b.
//
// Solidity: function closeBindMiner(address bind) returns(bool)
func (_Ticket *TicketTransactor) CloseBindMiner(opts *bind.TransactOpts, bind common.Address) (*types.Transaction, error) {
	return _Ticket.contract.Transact(opts, "closeBindMiner", bind)
}

// CloseBindMiner is a paid mutator transaction binding the contract method 0x57385d1b.
//
// Solidity: function closeBindMiner(address bind) returns(bool)
func (_Ticket *TicketSession) CloseBindMiner(bind common.Address) (*types.Transaction, error) {
	return _Ticket.Contract.CloseBindMiner(&_Ticket.TransactOpts, bind)
}

// CloseBindMiner is a paid mutator transaction binding the contract method 0x57385d1b.
//
// Solidity: function closeBindMiner(address bind) returns(bool)
func (_Ticket *TicketTransactorSession) CloseBindMiner(bind common.Address) (*types.Transaction, error) {
	return _Ticket.Contract.CloseBindMiner(&_Ticket.TransactOpts, bind)
}

// CreateBindMiner is a paid mutator transaction binding the contract method 0x5ca4b8fb.
//
// Solidity: function createBindMiner(address bind, uint256 amount) returns(bool)
func (_Ticket *TicketTransactor) CreateBindMiner(opts *bind.TransactOpts, bind common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Ticket.contract.Transact(opts, "createBindMiner", bind, amount)
}

// CreateBindMiner is a paid mutator transaction binding the contract method 0x5ca4b8fb.
//
// Solidity: function createBindMiner(address bind, uint256 amount) returns(bool)
func (_Ticket *TicketSession) CreateBindMiner(bind common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Ticket.Contract.CreateBindMiner(&_Ticket.TransactOpts, bind, amount)
}

// CreateBindMiner is a paid mutator transaction binding the contract method 0x5ca4b8fb.
//
// Solidity: function createBindMiner(address bind, uint256 amount) returns(bool)
func (_Ticket *TicketTransactorSession) CreateBindMiner(bind common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Ticket.Contract.CreateBindMiner(&_Ticket.TransactOpts, bind, amount)
}

// GetTicketCount is a paid mutator transaction binding the contract method 0x21c63a47.
//
// Solidity: function getTicketCount() returns(uint256)
func (_Ticket *TicketTransactor) GetTicketCount(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ticket.contract.Transact(opts, "getTicketCount")
}

// GetTicketCount is a paid mutator transaction binding the contract method 0x21c63a47.
//
// Solidity: function getTicketCount() returns(uint256)
func (_Ticket *TicketSession) GetTicketCount() (*types.Transaction, error) {
	return _Ticket.Contract.GetTicketCount(&_Ticket.TransactOpts)
}

// GetTicketCount is a paid mutator transaction binding the contract method 0x21c63a47.
//
// Solidity: function getTicketCount() returns(uint256)
func (_Ticket *TicketTransactorSession) GetTicketCount() (*types.Transaction, error) {
	return _Ticket.Contract.GetTicketCount(&_Ticket.TransactOpts)
}

// TransferToTickeExec is a paid mutator transaction binding the contract method 0x6850d595.
//
// Solidity: function transferToTickeExec(uint256 amount) returns(bool)
func (_Ticket *TicketTransactor) TransferToTickeExec(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Ticket.contract.Transact(opts, "transferToTickeExec", amount)
}

// TransferToTickeExec is a paid mutator transaction binding the contract method 0x6850d595.
//
// Solidity: function transferToTickeExec(uint256 amount) returns(bool)
func (_Ticket *TicketSession) TransferToTickeExec(amount *big.Int) (*types.Transaction, error) {
	return _Ticket.Contract.TransferToTickeExec(&_Ticket.TransactOpts, amount)
}

// TransferToTickeExec is a paid mutator transaction binding the contract method 0x6850d595.
//
// Solidity: function transferToTickeExec(uint256 amount) returns(bool)
func (_Ticket *TicketTransactorSession) TransferToTickeExec(amount *big.Int) (*types.Transaction, error) {
	return _Ticket.Contract.TransferToTickeExec(&_Ticket.TransactOpts, amount)
}

// TicketCloseBindMinerIterator is returned from FilterCloseBindMiner and is used to iterate over the raw logs and unpacked data for CloseBindMiner events raised by the Ticket contract.
type TicketCloseBindMinerIterator struct {
	Event *TicketCloseBindMiner // Event containing the contract specifics and raw log

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
func (it *TicketCloseBindMinerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TicketCloseBindMiner)
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
		it.Event = new(TicketCloseBindMiner)
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
func (it *TicketCloseBindMinerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TicketCloseBindMinerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TicketCloseBindMiner represents a CloseBindMiner event raised by the Ticket contract.
type TicketCloseBindMiner struct {
	Origin common.Address
	Bind   common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterCloseBindMiner is a free log retrieval operation binding the contract event 0x1e1e210b9e04215b745006f4b006a0b6f420f87563005fb9baa0e7407b8f5da9.
//
// Solidity: event CloseBindMiner(address origin, address bind)
func (_Ticket *TicketFilterer) FilterCloseBindMiner(opts *bind.FilterOpts) (*TicketCloseBindMinerIterator, error) {

	logs, sub, err := _Ticket.contract.FilterLogs(opts, "CloseBindMiner")
	if err != nil {
		return nil, err
	}
	return &TicketCloseBindMinerIterator{contract: _Ticket.contract, event: "CloseBindMiner", logs: logs, sub: sub}, nil
}

// WatchCloseBindMiner is a free log subscription operation binding the contract event 0x1e1e210b9e04215b745006f4b006a0b6f420f87563005fb9baa0e7407b8f5da9.
//
// Solidity: event CloseBindMiner(address origin, address bind)
func (_Ticket *TicketFilterer) WatchCloseBindMiner(opts *bind.WatchOpts, sink chan<- *TicketCloseBindMiner) (event.Subscription, error) {

	logs, sub, err := _Ticket.contract.WatchLogs(opts, "CloseBindMiner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TicketCloseBindMiner)
				if err := _Ticket.contract.UnpackLog(event, "CloseBindMiner", log); err != nil {
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

// ParseCloseBindMiner is a log parse operation binding the contract event 0x1e1e210b9e04215b745006f4b006a0b6f420f87563005fb9baa0e7407b8f5da9.
//
// Solidity: event CloseBindMiner(address origin, address bind)
func (_Ticket *TicketFilterer) ParseCloseBindMiner(log types.Log) (*TicketCloseBindMiner, error) {
	event := new(TicketCloseBindMiner)
	if err := _Ticket.contract.UnpackLog(event, "CloseBindMiner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TicketCreateBindMinerIterator is returned from FilterCreateBindMiner and is used to iterate over the raw logs and unpacked data for CreateBindMiner events raised by the Ticket contract.
type TicketCreateBindMinerIterator struct {
	Event *TicketCreateBindMiner // Event containing the contract specifics and raw log

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
func (it *TicketCreateBindMinerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TicketCreateBindMiner)
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
		it.Event = new(TicketCreateBindMiner)
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
func (it *TicketCreateBindMinerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TicketCreateBindMinerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TicketCreateBindMiner represents a CreateBindMiner event raised by the Ticket contract.
type TicketCreateBindMiner struct {
	Origin common.Address
	Bind   common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterCreateBindMiner is a free log retrieval operation binding the contract event 0xf68d6e786351c6914a35e97ea2605fb495e868c5f8e81f930394d7acbe8f7f3b.
//
// Solidity: event CreateBindMiner(address origin, address bind, uint256 amount)
func (_Ticket *TicketFilterer) FilterCreateBindMiner(opts *bind.FilterOpts) (*TicketCreateBindMinerIterator, error) {

	logs, sub, err := _Ticket.contract.FilterLogs(opts, "CreateBindMiner")
	if err != nil {
		return nil, err
	}
	return &TicketCreateBindMinerIterator{contract: _Ticket.contract, event: "CreateBindMiner", logs: logs, sub: sub}, nil
}

// WatchCreateBindMiner is a free log subscription operation binding the contract event 0xf68d6e786351c6914a35e97ea2605fb495e868c5f8e81f930394d7acbe8f7f3b.
//
// Solidity: event CreateBindMiner(address origin, address bind, uint256 amount)
func (_Ticket *TicketFilterer) WatchCreateBindMiner(opts *bind.WatchOpts, sink chan<- *TicketCreateBindMiner) (event.Subscription, error) {

	logs, sub, err := _Ticket.contract.WatchLogs(opts, "CreateBindMiner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TicketCreateBindMiner)
				if err := _Ticket.contract.UnpackLog(event, "CreateBindMiner", log); err != nil {
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

// ParseCreateBindMiner is a log parse operation binding the contract event 0xf68d6e786351c6914a35e97ea2605fb495e868c5f8e81f930394d7acbe8f7f3b.
//
// Solidity: event CreateBindMiner(address origin, address bind, uint256 amount)
func (_Ticket *TicketFilterer) ParseCreateBindMiner(log types.Log) (*TicketCreateBindMiner, error) {
	event := new(TicketCreateBindMiner)
	if err := _Ticket.contract.UnpackLog(event, "CreateBindMiner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TicketTransferToTicketExecIterator is returned from FilterTransferToTicketExec and is used to iterate over the raw logs and unpacked data for TransferToTicketExec events raised by the Ticket contract.
type TicketTransferToTicketExecIterator struct {
	Event *TicketTransferToTicketExec // Event containing the contract specifics and raw log

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
func (it *TicketTransferToTicketExecIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TicketTransferToTicketExec)
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
		it.Event = new(TicketTransferToTicketExec)
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
func (it *TicketTransferToTicketExecIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TicketTransferToTicketExecIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TicketTransferToTicketExec represents a TransferToTicketExec event raised by the Ticket contract.
type TicketTransferToTicketExec struct {
	From   common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTransferToTicketExec is a free log retrieval operation binding the contract event 0xa96959215ce46afdbce738a9b441a353e38c3c54c71c2894bfe6c6d4ccf10f18.
//
// Solidity: event TransferToTicketExec(address from, uint256 amount)
func (_Ticket *TicketFilterer) FilterTransferToTicketExec(opts *bind.FilterOpts) (*TicketTransferToTicketExecIterator, error) {

	logs, sub, err := _Ticket.contract.FilterLogs(opts, "TransferToTicketExec")
	if err != nil {
		return nil, err
	}
	return &TicketTransferToTicketExecIterator{contract: _Ticket.contract, event: "TransferToTicketExec", logs: logs, sub: sub}, nil
}

// WatchTransferToTicketExec is a free log subscription operation binding the contract event 0xa96959215ce46afdbce738a9b441a353e38c3c54c71c2894bfe6c6d4ccf10f18.
//
// Solidity: event TransferToTicketExec(address from, uint256 amount)
func (_Ticket *TicketFilterer) WatchTransferToTicketExec(opts *bind.WatchOpts, sink chan<- *TicketTransferToTicketExec) (event.Subscription, error) {

	logs, sub, err := _Ticket.contract.WatchLogs(opts, "TransferToTicketExec")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TicketTransferToTicketExec)
				if err := _Ticket.contract.UnpackLog(event, "TransferToTicketExec", log); err != nil {
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

// ParseTransferToTicketExec is a log parse operation binding the contract event 0xa96959215ce46afdbce738a9b441a353e38c3c54c71c2894bfe6c6d4ccf10f18.
//
// Solidity: event TransferToTicketExec(address from, uint256 amount)
func (_Ticket *TicketFilterer) ParseTransferToTicketExec(log types.Log) (*TicketTransferToTicketExec, error) {
	event := new(TicketTransferToTicketExec)
	if err := _Ticket.contract.UnpackLog(event, "TransferToTicketExec", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
