package runtime

//import (
//"fmt"
//"sort"
//
////"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/params"
//)
////
////var activators = map[int]func(*JumpTable){
////	2929: enable2929,
////	2200: enable2200,
////	1884: enable1884,
////	//1344: enable1344,
////}
//
//// EnableEIP enables the given EIP on the config.
//// This operation writes in-place, and callers need to ensure that the globally
//// defined jump tables are not polluted.
//func EnableEIP(eipNum int, jt *JumpTable) error {
//	enablerFn, ok := activators[eipNum]
//	if !ok {
//		return fmt.Errorf("undefined eip %d", eipNum)
//	}
//	enablerFn(jt)
//	return nil
//}
//
//func ValidEip(eipNum int) bool {
//	_, ok := activators[eipNum]
//	return ok
//}
//func ActivateableEips() []string {
//	var nums []string
//	for k := range activators {
//		nums = append(nums, fmt.Sprintf("%d", k))
//	}
//	sort.Strings(nums)
//	return nums
//}

// enable1884 applies EIP-1884 to the given jump table:
// - Increase cost of BALANCE to 700
// - Increase cost of EXTCODEHASH to 700
// - Increase cost of SLOAD to 800
// - Define SELFBALANCE, with cost GasFastStep (5)
//func enable1884(jt *JumpTable) {
//	// Gas cost changes
//	jt[SLOAD].constantGas = params.SloadGasEIP1884
//	jt[BALANCE].constantGas = params.BalanceGasEIP1884
//	jt[EXTCODEHASH].constantGas = params.ExtcodeHashGasEIP1884
//
//	// New opcode
//	jt[SELFBALANCE] = &operation{
//		execute:     opSelfBalance,
//		constantGas: GasFastStep,
//		minStack:    minStack(0, 1),
//		maxStack:    maxStack(0, 1),
//	}
//}

// enable1344 applies EIP-1344 (ChainID Opcode)
// - Adds an opcode that returns the current chain’s EIP-155 unique identifier
//func enable1344(jt *JumpTable) {
//	// New opcode
//	jt[CHAINID] = &operation{
//		execute:     opChainID,
//		constantGas: GasQuickStep,
//		minStack:    minStack(0, 1),
//		maxStack:    maxStack(0, 1),
//	}
//}

//enable2200 applies EIP-2200 (Rebalance net-metered SSTORE)
//func enable2200(jt *JumpTable) {
//	jt[SLOAD].constantGas = params.SloadGasEIP2200
//	//jt[SSTORE].dynamicGas = gasSStoreEIP2200
//}

//enable2929 enables "EIP-2929: Gas cost increases for state access opcodes"
//https://eips.ethereum.org/EIPS/eip-2929
//func enable2929(jt *JumpTable) {
//	//jt[SSTORE].dynamicGas = gasSStoreEIP2929
//	//
//	//jt[SLOAD].constantGas = 0
//	//jt[SLOAD].dynamicGas = gasSLoadEIP2929
//	//
//	//jt[EXTCODECOPY].constantGas = WarmStorageReadCostEIP2929
//	//jt[EXTCODECOPY].dynamicGas = gasExtCodeCopyEIP2929
//	//
//	//jt[EXTCODESIZE].constantGas = WarmStorageReadCostEIP2929
//	//jt[EXTCODESIZE].dynamicGas = gasEip2929AccountCheck
//	//
//	//jt[EXTCODEHASH].constantGas = WarmStorageReadCostEIP2929
//	//jt[EXTCODEHASH].dynamicGas = gasEip2929AccountCheck
//	//
//	//jt[BALANCE].constantGas = WarmStorageReadCostEIP2929
//	//jt[BALANCE].dynamicGas = gasEip2929AccountCheck
//	//
//	//jt[CALL].constantGas = WarmStorageReadCostEIP2929
//	//jt[CALL].dynamicGas = gasCallEIP2929
//	//
//	//jt[CALLCODE].constantGas = WarmStorageReadCostEIP2929
//	//jt[CALLCODE].dynamicGas = gasCallCodeEIP2929
//	//
//	//jt[STATICCALL].constantGas = WarmStorageReadCostEIP2929
//	//jt[STATICCALL].dynamicGas = gasStaticCallEIP2929
//	//
//	//jt[DELEGATECALL].constantGas = WarmStorageReadCostEIP2929
//	//jt[DELEGATECALL].dynamicGas = gasDelegateCallEIP2929
//
//	// This was previously part of the dynamic cost, but we're using it as a constantGas
//	// factor here
//	jt[SELFDESTRUCT].constantGas = params.SelfdestructGasEIP150
//	//jt[SELFDESTRUCT].dynamicGas = gasSelfdestructEIP2929
//}
