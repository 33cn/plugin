package runtime

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	log "github.com/33cn/chain33/common/log/log15"
	token "github.com/33cn/plugin/plugin/dapp/evm/contracts/token/generated"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
)

const (
	balanceOf   = "70a08231"
	decimals    = "313ce567"
	transfer    = "beabacc8"
	totalSupply = "18160ddd"
	//transfer(address,address,uint256)
)

const tokenExecer = "token"

//TokenContract token 合约
type TokenContract struct {
	SuperManager []string `json:"superManager,omitempty"`
}

//NewTokenPrecompile ...
func NewTokenPrecompile(tokeninfo *TokenContract) StatefulPrecompiledContract {
	call := &tokenPrecompile{}
	call.contractInfo = make(map[string]string)
	call.decimals = 8
	call.manage = tokeninfo.SuperManager
	var err error
	call.abi, err = evmAbi.JSON(strings.NewReader(token.TokenMetaData.ABI))
	if err != nil {
		panic(err)
	}
	return call
}

type tokenPrecompile struct {
	precomileAddress string
	decimals         int
	abi              evmAbi.ABI
	manage           []string
	cacheLock        sync.Mutex
	//缓存合约地址与tokenName 的对应关系
	contractInfo map[string]string
}

func (t *tokenPrecompile) RequiredGas(input []byte) uint64 {
	return 0
}

func (t *tokenPrecompile) checkCreator(evm *EVM, caller ContractRef) bool {
	//TODO 先在缓存中校验
	account := evm.StateDB.GetAccount(caller.Address().String())
	for _, mange := range t.manage {
		//要求合约创建者必须是管理员
		if strings.ToLower(account.GetCreator()) == strings.ToLower(mange) {
			return true
		}
	}
	return false
}

//setTokenSymbol 把token下币种的名称缓存起来
func (t *tokenPrecompile) setTokenSymbol(evm *EVM, caller ContractRef) {
	t.cacheLock.Lock()
	defer t.cacheLock.Unlock()
	if _, ok := t.contractInfo[strings.ToLower(caller.Address().String())]; ok {
		return
	}
	abidata, err := t.abi.Pack("name")
	if err != nil {
		panic(err)
	}
	contractAddr := caller.Address()
	contract := NewContract(caller, caller, new(big.Int), 21000)
	contract.SetCallCode(&contractAddr, evm.StateDB.GetCodeHash(contractAddr.String()), evm.StateDB.GetCode(contractAddr.String()))
	ret, err := run(evm, contract, abidata, true)
	if err == nil {
		var tokenName string
		err = t.abi.Unpack(&tokenName, "name", ret)
		if err != nil {
			log.Error("token.Precompiled setTokenSymbol", "tokenName:", err.Error())
			return
		}
		log.Info("token.Precompiled setTokenSymbol", "tokenName:", tokenName)
		t.contractInfo[strings.ToLower(caller.Address().String())] = tokenName
		return
	}

	log.Error("token.Precompiled setTokenSymbol", "err:", err)
}

//Run ...
func (t *tokenPrecompile) Run(evm *EVM, caller ContractRef, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	log.Info("token.Precompiled", "Run.Caller", caller.Address().String(), "inputSize:", len(input))
	if !t.checkCreator(evm, caller) {
		err = errors.New("contract not authorized")
		ret = []byte(err.Error())
		return
	}
	remainingGas = suppliedGas
	//获取方法哈希
	action := common.Bytes2Hex(input[:4])[2:]
	t.setTokenSymbol(evm, caller)
	switch action {
	case transfer:
		if len(input) < 68 {
			err = errors.New("input size too low")
			ret = []byte(err.Error())
			return
		}

		from := common.BytesToAddress(input[4:36])
		to := common.BytesToAddress(input[36 : 36+32])
		amount := big.NewInt(1).SetBytes(input[36+32:])
		var ok bool
		ok, err = t.callTransfer(evm, from, to, caller.Address(), amount.Int64())
		if err != nil {
			log.Error("token.Precompiled Run", "callTransfer", err, "input:", common.Bytes2Hex(input))
			ret = []byte(err.Error())
			return
		}
		ret, err = t.encode("transfer", ok)
		return

	case balanceOf:
		accountAddr := common.BytesToAddress(input[4:])
		var balance int64
		balance, err = t.callBalanceOf(evm, accountAddr, caller.Address())
		if err != nil {
			return
		}
		ret, err = t.encode("balanceOf", big.NewInt(balance))
		return

	case decimals:
		ret, err = t.encode("decimals", uint8(t.decimals))
		return

	case totalSupply:
		var total int64
		total, err = t.callTotalSupply(evm, caller.Address())
		if err != nil {
			return
		}
		ret, err = t.encode("totalSupply", big.NewInt(total))
		return
	}

	err = fmt.Errorf("token.Precompiled not support method:%v", action)
	return

}

func (t *tokenPrecompile) encode(k string, v interface{}) ([]byte, error) {
	return t.abi.Methods[k].Outputs.Pack(v)
}

func (t *tokenPrecompile) callTransfer(evm *EVM, caller, to, contract common.Address, amount int64) (ok bool, err error) {
	t.cacheLock.Lock()
	defer t.cacheLock.Unlock()
	if amount == 0 {
		ok = true
		return
	}
	tokenName := t.contractInfo[strings.ToLower(contract.String())]
	ok, err = evm.StateDB.TransferToToken(caller.String(), to.String(), tokenName, amount)
	return
}

func (t *tokenPrecompile) callBalanceOf(evm *EVM, caller, contract common.Address) (int64, error) {
	t.cacheLock.Lock()
	defer t.cacheLock.Unlock()
	tokenName := t.contractInfo[strings.ToLower(contract.String())]
	return evm.StateDB.TokenBalance(caller, tokenExecer, tokenName)

}

func (f *tokenPrecompile) callTotalSupply(evm *EVM, contract common.Address) (int64, error) {
	//total
	tokenName := f.contractInfo[strings.ToLower(contract.String())]
	return evm.StateDB.TokenSupply(tokenName)
}
