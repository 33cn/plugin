package runtime

import (
	"fmt"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/params"
)

//因为common.Address结构体中定义的是指针，map中的key值不能使用address作为key值来使用，于是使用Hash160Address作为key来进行索引
// PrecompiledContractsBerlin contains the default set of pre-compiled Ethereum
// contracts used in the Berlin release.
var CustomizePrecompiledContractsBinjiang = map[common.Hash160Address]PrecompiledContract{
	//0x01-0x08 是evm 系统自带的预编译合约接口，自定义预编译合约接口从0x1000 开始
	common.BytesToHash160Address([]byte{129}): &helloword{},
}

type helloword struct{}

func (h *helloword) RequiredGas(input []byte) uint64 {
	return uint64(len(input)+31)/32*params.IdentityPerWordGas + params.IdentityBaseGas
}

func (h *helloword) Run(input []byte) ([]byte, error) {
	fmt.Println("hellowordddddddddd,input", string(input))
	return []byte("hello,world-bj"), nil
}
