package runtime

import (
	"errors"
	"math/big"
	"strings"

	log "github.com/33cn/chain33/common/log/log15"
	ticket "github.com/33cn/plugin/plugin/dapp/evm/contracts/ticket/generated"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
)

/*
* tickets 预编译合约： evm合约 映射相关对go-ticket 合约,
* 功能：绑定挖矿，解除挖矿,向挖矿执行器转账
 */
const (
	getTicketCount      = "21c63a47"
	bindMiner           = ""
	transferToTickeExec = "6850d595"
)

const ticketExecer = "ticket"

//TicketContract ticket 合约
type TicketContract struct {
	SuperManager []string `json:"superManager,omitempty"`
}

type ticketPrecompile struct {
	precomileAddress string

	abi    evmAbi.ABI
	manage []string
}

//NewTicketPrecompile ...
func NewTicketPrecompile(tokeninfo *TokenContract) StatefulPrecompiledContract {
	call := &ticketPrecompile{}

	var err error
	call.abi, err = evmAbi.JSON(strings.NewReader(ticket.TicketMetaData.ABI))
	if err != nil {
		panic(err)
	}
	return call
}

func (t ticketPrecompile) RequiredGas(input []byte) uint64 {
	//免扣gas
	return 0
}
func (t ticketPrecompile) encode(k string, v interface{}) ([]byte, error) {
	return t.abi.Methods[k].Outputs.Pack(v)
}

func (t ticketPrecompile) Run(evm *EVM, caller ContractRef, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	if !t.checkCreator(evm, caller) {
		err = errors.New("ticket contract not authorized")
		ret = []byte(err.Error())
		return
	}

	remainingGas = suppliedGas
	//获取方法哈希
	action := common.Bytes2Hex(input[:4])[2:]
	switch action {
	case bindMiner:
		fromAddress := common.BytesToAddress(input[4:36])
		bindAddress := common.BytesToAddress(input[36 : 36+32])
		amount := big.NewInt(1).SetBytes(input[36+32:])
		var ok bool
		ok, err = t.callBindTicket(evm, fromAddress, bindAddress, caller.Address(), amount.Int64())
		if err != nil {
			log.Error("ticket.Precompiled Run", "callBindTicket", err, "input:", common.Bytes2Hex(input))
			ret = []byte(err.Error())
			return
		}
		ret, err = t.encode("createBindMiner", ok)

	case transferToTickeExec:
		fromAddress := common.BytesToAddress(input[4:36])
		amount := big.NewInt(1).SetBytes(input[36:])
		var ok bool
		ok, err = t.callTransfer2TicketExec(evm, fromAddress, amount.Int64())
		if err != nil {
			log.Error("ticket.Precompiled Run", "callTransfer2TicketExec", err, "input:", common.Bytes2Hex(input))
			ret = []byte(err.Error())
			return
		}
		ret, err = t.encode("transferToTickeExec", ok)

	}

	return
}

func (t ticketPrecompile) callBindTicket(evm *EVM, caller, bind, contract common.Address, amount int64) (ok bool, err error) {
	//check caller balance
	if !evm.CanTransfer(evm.StateDB, caller, uint64(amount)) {
		return false, errors.New("insufficient balance")
	}

	return evm.StateDB.CreateBindMiner(caller, bind, amount)

}

func (t ticketPrecompile) checkCreator(evm *EVM, caller ContractRef) bool {
	//检查创建合约的地址是否是管理员地址
	account := evm.StateDB.GetAccount(caller.Address().String())
	for _, mange := range t.manage {
		//要求合约创建者必须是管理员
		if strings.ToLower(account.GetCreator()) == strings.ToLower(mange) {
			return true
		}
	}

	return false
}

func (t ticketPrecompile) callTransfer2TicketExec(evm *EVM, from common.Address, amount int64) (ok bool, err error) {
	//check caller balance
	if !evm.CanTransfer(evm.StateDB, from, uint64(amount)) {
		return false, errors.New("insufficient balance")
	}

	return evm.StateDB.Transfer(from.String(), common.ExecAddress(ticketExecer).String(), uint64(amount)), nil
}

func (t ticketPrecompile) getTicketCount() (int, error) {
	return 0, nil
}
