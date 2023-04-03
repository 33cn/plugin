package runtime

import (
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
)

//TOKEN 预编译地址
const TokenPrecompileAddr = "0x0000000000000000000000000000000000200001"

//其他复杂合约下的预编译地址
const ComflixPrecompileAddr = "0x0000000000000000000000000000000000200002"

//TODO 添加新的 GO 合约地址

// CustomizePrecompiledContracts 存储自定义的预编译地址
var CustomizePrecompiledContracts = map[common.Hash160Address]StatefulPrecompiledContract{}

//StatefulPrecompiledContract precompile contract interface
type StatefulPrecompiledContract interface {
	// 计算当前合约执行需要消耗的Gas
	RequiredGas(input []byte) uint64
	Run(evm *EVM, caller ContractRef, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error)
}

//RunStateFulPrecompiledContract 调用自定义的预编译的合约逻辑并返回结果
func RunStateFulPrecompiledContract(evm *EVM, caller ContractRef, sp StatefulPrecompiledContract, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	ret, remainingGas, err = sp.Run(evm, caller, input, suppliedGas)
	return
}
