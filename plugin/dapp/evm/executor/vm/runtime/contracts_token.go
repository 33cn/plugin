package runtime

import (
	"errors"
	"fmt"
	"github.com/33cn/chain33/account"
	log "github.com/33cn/chain33/common/log/log15"
	token "github.com/33cn/plugin/plugin/dapp/evm/contracts/token/generated"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"math/big"
	"strings"
)

//StatefulPrecompiledContract precompile contract interface
type StatefulPrecompiledContract interface {
	Run(evm *EVM, caller ContractRef, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error)
}

// CustomizePrecompiledContracts 存储自定义的预编译地址
var CustomizePrecompiledContracts = map[common.Hash160Address]StatefulPrecompiledContract{}

const (
	//transfer  = "a9059cbb"
	balanceOf   = "70a08231"
	decimals    = "313ce567"
	symbol      = "95d89b41"
	name        = "06fdde03"
	totalSupply = "18160ddd"
	transfer    = "beabacc8"
	//transfer(address,address,uint256)
)

const tokenExecer = "token"

//TokenContract token 合约
type TokenContract struct {
	TotalSupply       int64  `json:"totalSupply,omitempty"`
	Symbol            string `json:"symbol,omitempty"`
	Decimals          int    `json:"decimals,omitempty"`
	PreCompileAddress string `json:"preCompileAddress,omitempty"`
}

//RunStateFulPrecompiledContract 调用自定义的预编译的合约逻辑并返回结果
func RunStateFulPrecompiledContract(evm *EVM, caller ContractRef, sp StatefulPrecompiledContract, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	ret, remainingGas, err = sp.Run(evm, caller, input, suppliedGas)
	return
}

//NewTokenCall ...
func NewTokenCall(tokeninfo *TokenContract) StatefulPrecompiledContract {
	call := &tokenCall{}
	call.precomileAddress = tokeninfo.PreCompileAddress
	call.tokenName = tokeninfo.Symbol
	call.totalSupply = tokeninfo.TotalSupply
	call.decimals = 8
	var err error
	call.abi, err = evmAbi.JSON(strings.NewReader(token.TokenMetaData.ABI))
	if err != nil {
		panic(err)
	}
	return call
}

type tokenCall struct {
	precomileAddress string
	tokenName        string
	decimals         int
	totalSupply      int64
	abi              evmAbi.ABI
}

//Run ...
func (t *tokenCall) Run(evm *EVM, caller ContractRef, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	log.Info("tokenCall", "Run.Caller", caller.Address().String(), "inputSize:", len(input))
	remainingGas = suppliedGas
	action := common.Bytes2Hex(input[:4])[2:]
	switch action {
	case transfer:
		if len(input) < 100 {
			err = errors.New("input size too low")
			return
		}
		from := common.BytesToAddress(input[4:36])
		to := common.BytesToAddress(input[36 : 36+32])
		amount := big.NewInt(1).SetBytes(input[36+32:])
		var ok bool
		ok, err = t.callTransfer(evm, from, to, amount.Int64())
		if err != nil {
			log.Error("tokenCall.Run", "callTransfer", err, "input:", common.Bytes2Hex(input))
			return
		}
		ret, err = t.encode("transfer", ok)
		return

	case balanceOf:
		accountAddr := common.BytesToAddress(input[4:])
		var balance int64
		balance, err = t.callBalanceOf(evm, accountAddr)
		if err != nil {
			return
		}
		ret, err = t.encode("balanceOf", big.NewInt(balance))
		return

	case decimals:
		ret, err = t.encode("decimals", uint8(t.decimals))
		return

	case symbol:
		ret, err = t.encode("symbol", t.tokenName)
		return

	case name:
		ret, err = t.encode("name", t.tokenName)
		return

	case totalSupply:
		ret, err = t.encode("totalSupply", t.totalSupply)
		return
	}

	err = fmt.Errorf("no support method:%v", action)
	return

}

func (t *tokenCall) encode(k string, v interface{}) ([]byte, error) {
	return t.abi.Methods[k].Outputs.Pack(v)
}

func (t *tokenCall) callTransfer(evm *EVM, caller, to common.Address, amount int64) (ok bool, err error) {
	if amount == 0 {
		ok = true
		return
	}
	ok, err = evm.MStateDB.TransferToToken(caller.String(), to.String(), t.tokenName, amount)
	return
}

func (t *tokenCall) callBalanceOf(evm *EVM, caller common.Address) (int64, error) {
	cfg := evm.MStateDB.GetConfig()
	tokenAccount, err := account.NewAccountDB(cfg, tokenExecer, t.tokenName, evm.MStateDB.StateDB)
	if err != nil {
		return 0, err
	}
	acc := tokenAccount.LoadAccount(caller.String())
	if acc == nil {
		return 0, nil
	}
	return acc.Balance, nil
}
