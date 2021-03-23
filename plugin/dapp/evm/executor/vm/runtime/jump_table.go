// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/params"
)




//type (
//	// ExecutionFunc 指令执行函数，每个操作指令对应一个实现，它实现了指令的具体操作逻辑
//	ExecutionFunc func(pc *uint64, evm *EVM, callContext *callCtx) ([]byte, error)
//)


type (
	//executionFunc func(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx) ([]byte, error)
	executionFunc func(pc *uint64, evm *EVM, callContext *callCtx) ([]byte, error)

	gasFunc       func(*EVM, *Contract, *Stack, *Memory, uint64) (uint64, error) // last parameter is the requested memory size as a uint64
	// memorySizeFunc returns the required size, and whether the operation overflowed a uint64
	memorySizeFunc func(*Stack) (size uint64, overflow bool)
)

type operation struct {
	// execute is the operation function
	execute     executionFunc
	constantGas uint64
	dynamicGas  gasFunc
	// minStack tells how many stack items are required
	minStack int
	// maxStack specifies the max length the stack can have for this operation
	// to not overflow the stack.
	maxStack int

	// memorySize returns the memory size required for the operation
	memorySize memorySizeFunc

	halts   bool // indicates whether the operation should halt further execution
	jumps   bool // indicates whether the program counter should not increment
	writes  bool // determines whether this a state modifying operation
	reverts bool // determines whether the operation reverts state (implicitly halts)
	returns bool // determines whether the operations sets the return data content
}


var (
	frontierInstructionSet         = newFrontierInstructionSet()
	homesteadInstructionSet        = newHomesteadInstructionSet()
	tangerineWhistleInstructionSet = newTangerineWhistleInstructionSet()
	spuriousDragonInstructionSet   = newSpuriousDragonInstructionSet()
	byzantiumInstructionSet        = newByzantiumInstructionSet()
	constantinopleInstructionSet   = newConstantinopleInstructionSet()
	istanbulInstructionSet         = newIstanbulInstructionSet()
	berlinInstructionSet           = newBerlinInstructionSet()
)

// JumpTable contains the EVM opcodes supported at a given fork.
type JumpTable [256]*operation

// newBerlinInstructionSet returns the frontier, homestead, byzantium,
// contantinople, istanbul, petersburg and berlin instructions.
func newBerlinInstructionSet() JumpTable {
	instructionSet := newIstanbulInstructionSet()
	return instructionSet
}

// newIstanbulInstructionSet returns the frontier, homestead, byzantium,
// contantinople, istanbul and petersburg instructions.
func newIstanbulInstructionSet() JumpTable {
	instructionSet := newConstantinopleInstructionSet()

	//enable1884(&instructionSet) // Reprice reader opcodes - https://eips.ethereum.org/EIPS/eip-1884
	//enable2200(&instructionSet) // Net metered SSTORE - https://eips.ethereum.org/EIPS/eip-2200
	// New opcode
	instructionSet[CHAINID] = &operation{
		execute:     opChainID,
		constantGas: GasQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}

	// New opcode
	instructionSet[SELFBALANCE] = &operation{
		execute:     opSelfBalance,
		constantGas: GasFastStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
	return instructionSet
}

// newConstantinopleInstructionSet returns the frontier, homestead,
// byzantium and contantinople instructions.
func newConstantinopleInstructionSet() JumpTable {
	instructionSet := newByzantiumInstructionSet()
	instructionSet[SHL] = &operation{
		execute:     opSHL,
		constantGas: GasFastestStep,
		minStack:    minStack(2, 1),
		maxStack:    maxStack(2, 1),
	}
	instructionSet[SHR] = &operation{
		execute:     opSHR,
		constantGas: GasFastestStep,
		minStack:    minStack(2, 1),
		maxStack:    maxStack(2, 1),
	}
	instructionSet[SAR] = &operation{
		execute:     opSAR,
		constantGas: GasFastestStep,
		minStack:    minStack(2, 1),
		maxStack:    maxStack(2, 1),
	}
	instructionSet[EXTCODEHASH] = &operation{
		execute:     opExtCodeHash,
		constantGas: params.ExtcodeHashGasConstantinople,
		minStack:    minStack(1, 1),
		maxStack:    maxStack(1, 1),
	}
	instructionSet[CREATE2] = &operation{
		execute:     opCreate2,
		constantGas: params.Create2Gas,
		dynamicGas:  gasCreate2,
		minStack:    minStack(4, 1),
		maxStack:    maxStack(4, 1),
		memorySize:  memoryCreate2,
		writes:      true,
		returns:     true,
	}
	return instructionSet
}

// newByzantiumInstructionSet returns the frontier, homestead and
// byzantium instructions.
func newByzantiumInstructionSet() JumpTable {
	instructionSet := newSpuriousDragonInstructionSet()
	instructionSet[STATICCALL] = &operation{
		execute:     opStaticCall,
		constantGas: params.CallGasEIP150,
		dynamicGas:  gasStaticCall,
		minStack:    minStack(6, 1),
		maxStack:    maxStack(6, 1),
		memorySize:  memoryStaticCall,
		returns:     true,
	}
	instructionSet[RETURNDATASIZE] = &operation{
		execute:     opReturnDataSize,
		constantGas: GasQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
	instructionSet[RETURNDATACOPY] = &operation{
		execute:     opReturnDataCopy,
		constantGas: GasFastestStep,
		dynamicGas:  gasReturnDataCopy,
		minStack:    minStack(3, 0),
		maxStack:    maxStack(3, 0),
		memorySize:  memoryReturnDataCopy,
	}
	instructionSet[REVERT] = &operation{
		execute:    opRevert,
		dynamicGas: gasRevert,
		minStack:   minStack(2, 0),
		maxStack:   maxStack(2, 0),
		memorySize: memoryRevert,
		reverts:    true,
		returns:    true,
	}
	return instructionSet
}

// EIP 158 a.k.a Spurious Dragon
func newSpuriousDragonInstructionSet() JumpTable {
	instructionSet := newTangerineWhistleInstructionSet()
	instructionSet[EXP].dynamicGas = gasExpEIP158
	return instructionSet

}

// EIP 150 a.k.a Tangerine Whistle
func newTangerineWhistleInstructionSet() JumpTable {
	instructionSet := newHomesteadInstructionSet()
	instructionSet[BALANCE].constantGas = params.BalanceGasEIP150
	instructionSet[EXTCODESIZE].constantGas = params.ExtcodeSizeGasEIP150
	instructionSet[SLOAD].constantGas = params.SloadGasEIP150
	instructionSet[EXTCODECOPY].constantGas = params.ExtcodeCopyBaseEIP150
	instructionSet[CALL].constantGas = params.CallGasEIP150
	instructionSet[CALLCODE].constantGas = params.CallGasEIP150
	instructionSet[DELEGATECALL].constantGas = params.CallGasEIP150
	return instructionSet
}

// newHomesteadInstructionSet returns the frontier and homestead
// instructions that can be executed during the homestead phase.
func newHomesteadInstructionSet() JumpTable {
	instructionSet := newFrontierInstructionSet()
	instructionSet[DELEGATECALL] = &operation{
		execute:     opDelegateCall,
		dynamicGas:  gasDelegateCall,
		constantGas: params.CallGasFrontier,
		minStack:    minStack(6, 1),
		maxStack:    maxStack(6, 1),
		memorySize:  memoryDelegateCall,
		returns:     true,
	}
	return instructionSet
}

// newFrontierInstructionSet returns the frontier instructions
// that can be executed during the frontier phase.
func newFrontierInstructionSet() JumpTable {
	return JumpTable{
		STOP: {
			execute:     opStop,
			constantGas: 0,
			minStack:    minStack(0, 0),
			maxStack:    maxStack(0, 0),
			halts:       true,
		},
		ADD: {
			execute:     opAdd,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		MUL: {
			execute:     opMul,
			constantGas: GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		SUB: {
			execute:     opSub,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		DIV: {
			execute:     opDiv,
			constantGas: GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		SDIV: {
			execute:     opSdiv,
			constantGas: GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		MOD: {
			execute:     opMod,
			constantGas: GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		SMOD: {
			execute:     opSmod,
			constantGas: GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		ADDMOD: {
			execute:     opAddmod,
			constantGas: GasMidStep,
			minStack:    minStack(3, 1),
			maxStack:    maxStack(3, 1),
		},
		MULMOD: {
			execute:     opMulmod,
			constantGas: GasMidStep,
			minStack:    minStack(3, 1),
			maxStack:    maxStack(3, 1),
		},
		EXP: {
			execute:    opExp,
			dynamicGas: gasExpFrontier,
			minStack:   minStack(2, 1),
			maxStack:   maxStack(2, 1),
		},
		SIGNEXTEND: {
			execute:     opSignExtend,
			constantGas: GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		LT: {
			execute:     opLt,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		GT: {
			execute:     opGt,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		SLT: {
			execute:     opSlt,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		SGT: {
			execute:     opSgt,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		EQ: {
			execute:     opEq,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		ISZERO: {
			execute:     opIszero,
			constantGas: GasFastestStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		AND: {
			execute:     opAnd,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		XOR: {
			execute:     opXor,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		OR: {
			execute:     opOr,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		NOT: {
			execute:     opNot,
			constantGas: GasFastestStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		BYTE: {
			execute:     opByte,
			constantGas: GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		SHA3: {
			execute:     opSha3,
			constantGas: params.Sha3Gas,
			dynamicGas:  gasSha3,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			memorySize:  memorySha3,
		},
		ADDRESS: {
			execute:     opAddress,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		BALANCE: {
			execute:     opBalance,
			constantGas: params.BalanceGasFrontier,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		ORIGIN: {
			execute:     opOrigin,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		CALLER: {
			execute:     opCaller,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		CALLVALUE: {
			execute:     opCallValue,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		CALLDATALOAD: {
			execute:     opCallDataLoad,
			constantGas: GasFastestStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		CALLDATASIZE: {
			execute:     opCallDataSize,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		CALLDATACOPY: {
			execute:     opCallDataCopy,
			constantGas: GasFastestStep,
			dynamicGas:  gasCallDataCopy,
			minStack:    minStack(3, 0),
			maxStack:    maxStack(3, 0),
			memorySize:  memoryCallDataCopy,
		},
		CODESIZE: {
			execute:     opCodeSize,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		CODECOPY: {
			execute:     opCodeCopy,
			constantGas: GasFastestStep,
			dynamicGas:  gasCodeCopy,
			minStack:    minStack(3, 0),
			maxStack:    maxStack(3, 0),
			memorySize:  memoryCodeCopy,
		},
		GASPRICE: {
			execute:     opGasprice,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		EXTCODESIZE: {
			execute:     opExtCodeSize,
			constantGas: params.ExtcodeSizeGasFrontier,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		EXTCODECOPY: {
			execute:     opExtCodeCopy,
			constantGas: params.ExtcodeCopyBaseFrontier,
			dynamicGas:  gasExtCodeCopy,
			minStack:    minStack(4, 0),
			maxStack:    maxStack(4, 0),
			memorySize:  memoryExtCodeCopy,
		},
		BLOCKHASH: {
			execute:     opBlockhash,
			constantGas: GasExtStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		COINBASE: {
			execute:     opCoinbase,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		TIMESTAMP: {
			execute:     opTimestamp,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		NUMBER: {
			execute:     opNumber,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		DIFFICULTY: {
			execute:     opDifficulty,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		GASLIMIT: {
			execute:     opGasLimit,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		POP: {
			execute:     opPop,
			constantGas: GasQuickStep,
			minStack:    minStack(1, 0),
			maxStack:    maxStack(1, 0),
		},
		MLOAD: {
			execute:     opMload,
			constantGas: GasFastestStep,
			dynamicGas:  gasMLoad,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
			memorySize:  memoryMLoad,
		},
		MSTORE: {
			execute:     opMstore,
			constantGas: GasFastestStep,
			dynamicGas:  gasMStore,
			minStack:    minStack(2, 0),
			maxStack:    maxStack(2, 0),
			memorySize:  memoryMStore,
		},
		MSTORE8: {
			execute:     opMstore8,
			constantGas: GasFastestStep,
			dynamicGas:  gasMStore8,
			memorySize:  memoryMStore8,
			minStack:    minStack(2, 0),
			maxStack:    maxStack(2, 0),
		},
		SLOAD: {
			execute:     opSload,
			constantGas: params.SloadGasFrontier,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		SSTORE: {
			execute:    opSstore,
			dynamicGas: gasSStore,
			minStack:   minStack(2, 0),
			maxStack:   maxStack(2, 0),
			writes:     true,
		},
		JUMP: {
			execute:     opJump,
			constantGas: GasMidStep,
			minStack:    minStack(1, 0),
			maxStack:    maxStack(1, 0),
			jumps:       true,
		},
		JUMPI: {
			execute:     opJumpi,
			constantGas: GasSlowStep,
			minStack:    minStack(2, 0),
			maxStack:    maxStack(2, 0),
			jumps:       true,
		},
		PC: {
			execute:     opPc,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		MSIZE: {
			execute:     opMsize,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		GAS: {
			execute:     opGas,
			constantGas: GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		JUMPDEST: {
			execute:     opJumpdest,
			constantGas: params.JumpdestGas,
			minStack:    minStack(0, 0),
			maxStack:    maxStack(0, 0),
		},
		PUSH1: {
			execute:     opPush1,
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH2: {
			execute:     makePush(2, 2),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH3: {
			execute:     makePush(3, 3),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH4: {
			execute:     makePush(4, 4),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH5: {
			execute:     makePush(5, 5),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH6: {
			execute:     makePush(6, 6),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH7: {
			execute:     makePush(7, 7),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH8: {
			execute:     makePush(8, 8),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH9: {
			execute:     makePush(9, 9),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH10: {
			execute:     makePush(10, 10),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH11: {
			execute:     makePush(11, 11),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH12: {
			execute:     makePush(12, 12),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH13: {
			execute:     makePush(13, 13),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH14: {
			execute:     makePush(14, 14),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH15: {
			execute:     makePush(15, 15),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH16: {
			execute:     makePush(16, 16),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH17: {
			execute:     makePush(17, 17),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH18: {
			execute:     makePush(18, 18),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH19: {
			execute:     makePush(19, 19),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH20: {
			execute:     makePush(20, 20),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH21: {
			execute:     makePush(21, 21),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH22: {
			execute:     makePush(22, 22),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH23: {
			execute:     makePush(23, 23),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH24: {
			execute:     makePush(24, 24),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH25: {
			execute:     makePush(25, 25),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH26: {
			execute:     makePush(26, 26),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH27: {
			execute:     makePush(27, 27),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH28: {
			execute:     makePush(28, 28),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH29: {
			execute:     makePush(29, 29),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH30: {
			execute:     makePush(30, 30),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH31: {
			execute:     makePush(31, 31),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		PUSH32: {
			execute:     makePush(32, 32),
			constantGas: GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		DUP1: {
			execute:     makeDup(1),
			constantGas: GasFastestStep,
			minStack:    minDupStack(1),
			maxStack:    maxDupStack(1),
		},
		DUP2: {
			execute:     makeDup(2),
			constantGas: GasFastestStep,
			minStack:    minDupStack(2),
			maxStack:    maxDupStack(2),
		},
		DUP3: {
			execute:     makeDup(3),
			constantGas: GasFastestStep,
			minStack:    minDupStack(3),
			maxStack:    maxDupStack(3),
		},
		DUP4: {
			execute:     makeDup(4),
			constantGas: GasFastestStep,
			minStack:    minDupStack(4),
			maxStack:    maxDupStack(4),
		},
		DUP5: {
			execute:     makeDup(5),
			constantGas: GasFastestStep,
			minStack:    minDupStack(5),
			maxStack:    maxDupStack(5),
		},
		DUP6: {
			execute:     makeDup(6),
			constantGas: GasFastestStep,
			minStack:    minDupStack(6),
			maxStack:    maxDupStack(6),
		},
		DUP7: {
			execute:     makeDup(7),
			constantGas: GasFastestStep,
			minStack:    minDupStack(7),
			maxStack:    maxDupStack(7),
		},
		DUP8: {
			execute:     makeDup(8),
			constantGas: GasFastestStep,
			minStack:    minDupStack(8),
			maxStack:    maxDupStack(8),
		},
		DUP9: {
			execute:     makeDup(9),
			constantGas: GasFastestStep,
			minStack:    minDupStack(9),
			maxStack:    maxDupStack(9),
		},
		DUP10: {
			execute:     makeDup(10),
			constantGas: GasFastestStep,
			minStack:    minDupStack(10),
			maxStack:    maxDupStack(10),
		},
		DUP11: {
			execute:     makeDup(11),
			constantGas: GasFastestStep,
			minStack:    minDupStack(11),
			maxStack:    maxDupStack(11),
		},
		DUP12: {
			execute:     makeDup(12),
			constantGas: GasFastestStep,
			minStack:    minDupStack(12),
			maxStack:    maxDupStack(12),
		},
		DUP13: {
			execute:     makeDup(13),
			constantGas: GasFastestStep,
			minStack:    minDupStack(13),
			maxStack:    maxDupStack(13),
		},
		DUP14: {
			execute:     makeDup(14),
			constantGas: GasFastestStep,
			minStack:    minDupStack(14),
			maxStack:    maxDupStack(14),
		},
		DUP15: {
			execute:     makeDup(15),
			constantGas: GasFastestStep,
			minStack:    minDupStack(15),
			maxStack:    maxDupStack(15),
		},
		DUP16: {
			execute:     makeDup(16),
			constantGas: GasFastestStep,
			minStack:    minDupStack(16),
			maxStack:    maxDupStack(16),
		},
		SWAP1: {
			execute:     makeSwap(1),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(2),
			maxStack:    maxSwapStack(2),
		},
		SWAP2: {
			execute:     makeSwap(2),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(3),
			maxStack:    maxSwapStack(3),
		},
		SWAP3: {
			execute:     makeSwap(3),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(4),
			maxStack:    maxSwapStack(4),
		},
		SWAP4: {
			execute:     makeSwap(4),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(5),
			maxStack:    maxSwapStack(5),
		},
		SWAP5: {
			execute:     makeSwap(5),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(6),
			maxStack:    maxSwapStack(6),
		},
		SWAP6: {
			execute:     makeSwap(6),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(7),
			maxStack:    maxSwapStack(7),
		},
		SWAP7: {
			execute:     makeSwap(7),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(8),
			maxStack:    maxSwapStack(8),
		},
		SWAP8: {
			execute:     makeSwap(8),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(9),
			maxStack:    maxSwapStack(9),
		},
		SWAP9: {
			execute:     makeSwap(9),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(10),
			maxStack:    maxSwapStack(10),
		},
		SWAP10: {
			execute:     makeSwap(10),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(11),
			maxStack:    maxSwapStack(11),
		},
		SWAP11: {
			execute:     makeSwap(11),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(12),
			maxStack:    maxSwapStack(12),
		},
		SWAP12: {
			execute:     makeSwap(12),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(13),
			maxStack:    maxSwapStack(13),
		},
		SWAP13: {
			execute:     makeSwap(13),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(14),
			maxStack:    maxSwapStack(14),
		},
		SWAP14: {
			execute:     makeSwap(14),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(15),
			maxStack:    maxSwapStack(15),
		},
		SWAP15: {
			execute:     makeSwap(15),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(16),
			maxStack:    maxSwapStack(16),
		},
		SWAP16: {
			execute:     makeSwap(16),
			constantGas: GasFastestStep,
			minStack:    minSwapStack(17),
			maxStack:    maxSwapStack(17),
		},
		LOG0: {
			execute:    makeLog(0),
			dynamicGas: makeGasLog(0),
			minStack:   minStack(2, 0),
			maxStack:   maxStack(2, 0),
			memorySize: memoryLog,
			writes:     true,
		},
		LOG1: {
			execute:    makeLog(1),
			dynamicGas: makeGasLog(1),
			minStack:   minStack(3, 0),
			maxStack:   maxStack(3, 0),
			memorySize: memoryLog,
			writes:     true,
		},
		LOG2: {
			execute:    makeLog(2),
			dynamicGas: makeGasLog(2),
			minStack:   minStack(4, 0),
			maxStack:   maxStack(4, 0),
			memorySize: memoryLog,
			writes:     true,
		},
		LOG3: {
			execute:    makeLog(3),
			dynamicGas: makeGasLog(3),
			minStack:   minStack(5, 0),
			maxStack:   maxStack(5, 0),
			memorySize: memoryLog,
			writes:     true,
		},
		LOG4: {
			execute:    makeLog(4),
			dynamicGas: makeGasLog(4),
			minStack:   minStack(6, 0),
			maxStack:   maxStack(6, 0),
			memorySize: memoryLog,
			writes:     true,
		},
		CREATE: {
			execute:     opCreate,
			constantGas: params.CreateGas,
			dynamicGas:  gasCreate,
			minStack:    minStack(3, 1),
			maxStack:    maxStack(3, 1),
			memorySize:  memoryCreate,
			writes:      true,
			returns:     true,
		},
		CALL: {
			execute:     opCall,
			constantGas: params.CallGasFrontier,
			dynamicGas:  gasCall,
			minStack:    minStack(7, 1),
			maxStack:    maxStack(7, 1),
			memorySize:  memoryCall,
			returns:     true,
		},
		CALLCODE: {
			execute:     opCallCode,
			constantGas: params.CallGasFrontier,
			dynamicGas:  gasCallCode,
			minStack:    minStack(7, 1),
			maxStack:    maxStack(7, 1),
			memorySize:  memoryCall,
			returns:     true,
		},
		RETURN: {
			execute:    opReturn,
			dynamicGas: gasReturn,
			minStack:   minStack(2, 0),
			maxStack:   maxStack(2, 0),
			memorySize: memoryReturn,
			halts:      true,
		},
		SELFDESTRUCT: {
			execute:    opSuicide,
			dynamicGas: gasSelfdestruct,
			minStack:   minStack(1, 0),
			maxStack:   maxStack(1, 0),
			halts:      true,
			writes:     true,
		},
	}
}




////Operation 定义指令操作的结构提
//type Operation struct {
//	// Execute 指令的具体操作逻辑
//	Execute ExecutionFunc
//
//	// GasCost 计算当前指令执行所需消耗的Gas
//	GasCost gas.CalcGasFunc
//
//	// ValidateStack 检查内存栈中的数据是否满足本操作执行的要求
//	ValidateStack mm.StackValidationFunc
//
//	// MemorySize 计算本次操作所需要的内存大小
//	MemorySize mm.MemorySizeFunc
//
//	// Halts   是否需要暂停（将会结束本合约后面操作的执行）
//	Halts bool
//	// Jumps   是否需要执行跳转（此种情况下PC不递增）
//	Jumps bool
//	// Writes  是否涉及到修改状态操作（在合约委托调用的情况下，此操作非法，将会抛异常）
//	Writes bool
//	// Valid   是否为有效操作
//	Valid bool
//	// Reverts 是否恢复原始状态（强制暂停，将会结束本合约后面操作的执行）
//	Reverts bool
//	// Returns 是否返回
//	Returns bool
//}
//
//var (
//	// ConstantinopleInstructionSet 对应EVM不同版本的指令集，从上往下，从旧版本到新版本，
//	// 新版本包含旧版本的指令集（目前直接使用康士坦丁堡指令集）
//	ConstantinopleInstructionSet = NewConstantinopleInstructionSet()
//	// YoloV1InstructionSet 黄皮书指令集
//	YoloV1InstructionSet = NewYoloV1InstructionSet()
//)
//
//// JumpTable contains the EVM opcodes supported at a given fork.
//type JumpTable [256]Operation
//
//// NewYoloV1InstructionSet 黄皮书指令集
//func NewYoloV1InstructionSet() JumpTable {
//	instructionSet := NewConstantinopleInstructionSet()
//	// New opcode
//	instructionSet[BEGINSUB] = Operation{
//		Execute:       opBeginSub,
//		GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//		ValidateStack: mm.MakeStackFunc(0, 0),
//		Valid:         true,
//	}
//	// New opcode
//	instructionSet[JUMPSUB] = Operation{
//		Execute:       opJumpSub,
//		GasCost:       gas.ConstGasFunc(gas.GasSlowStep),
//		ValidateStack: mm.MakeStackFunc(1, 0),
//		Jumps:         true,
//		Valid:         true,
//	}
//	// New opcode
//	instructionSet[RETURNSUB] = Operation{
//		Execute:       opReturnSub,
//		GasCost:       gas.ConstGasFunc(gas.GasFastStep),
//		ValidateStack: mm.MakeStackFunc(0, 0),
//		Jumps:         true,
//		Valid:         true,
//	}
//	// New opcode
//	instructionSet[SELFBALANCE] = Operation{
//		Execute:       opSelfBalance,
//		GasCost:       gas.ConstGasFunc(gas.GasFastStep),
//		ValidateStack: mm.MakeStackFunc(0, 1),
//		Valid:         true,
//	}
//	// New opcode
//	instructionSet[EXTCODEHASH] = Operation{
//		Execute:       opExtCodeHash,
//		GasCost:       gas.ConstGasFunc(params.ExtcodeHashGasConstantinople),
//		ValidateStack: mm.MakeStackFunc(1, 1),
//		Valid:         true,
//	}
//	// create2 不支持
//	// chainID 不支持
//	// New opcode
//	instructionSet[CREATE2] = Operation{
//		Execute:       opCreate2,
//		GasCost:       gas.ConstGasFunc(params.CreateGas),
//		ValidateStack: mm.MakeStackFunc(4, 1),
//		Valid:         true,
//	}
//	instructionSet[CHAINID] = Operation{
//		Execute:       opChainID,
//		GasCost:       gas.ConstGasFunc(params.GasQuickStep),
//		ValidateStack: mm.MakeStackFunc(0, 1),
//		Valid:         true,
//	}
//
//	//PUSH1 指令变更
//	instructionSet[PUSH1] = Operation{
//		Execute:       opPush1,
//		GasCost:       gas.Push,
//		ValidateStack: mm.MakeStackFunc(0, 1),
//		Valid:         true,
//	}
//	return instructionSet
//}
//
//// NewConstantinopleInstructionSet 康士坦丁堡 版本支持的指令集
//func NewConstantinopleInstructionSet() JumpTable {
//	instructionSet := NewByzantiumInstructionSet()
//	instructionSet[SHL] = Operation{
//		Execute:       opSHL,
//		GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//		ValidateStack: mm.MakeStackFunc(2, 1),
//		Valid:         true,
//	}
//	instructionSet[SHR] = Operation{
//		Execute:       opSHR,
//		GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//		ValidateStack: mm.MakeStackFunc(2, 1),
//		Valid:         true,
//	}
//	instructionSet[SAR] = Operation{
//		Execute:       opSAR,
//		GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//		ValidateStack: mm.MakeStackFunc(2, 1),
//		Valid:         true,
//	}
//	return instructionSet
//}
//
//// NewByzantiumInstructionSet 拜占庭 版本支持的指令集
//func NewByzantiumInstructionSet() [256]Operation {
//	instructionSet := NewHomesteadInstructionSet()
//	instructionSet[STATICCALL] = Operation{
//		Execute:       opStaticCall,
//		GasCost:       gas.StaticCall,
//		ValidateStack: mm.MakeStackFunc(6, 1),
//		MemorySize:    mm.MemoryStaticCall,
//		Valid:         true,
//		Returns:       true,
//	}
//	instructionSet[RETURNDATASIZE] = Operation{
//		Execute:       opReturnDataSize,
//		GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//		ValidateStack: mm.MakeStackFunc(0, 1),
//		Valid:         true,
//	}
//	instructionSet[RETURNDATACOPY] = Operation{
//		Execute:       opReturnDataCopy,
//		GasCost:       gas.ReturnDataCopy,
//		ValidateStack: mm.MakeStackFunc(3, 0),
//		MemorySize:    mm.MemoryReturnDataCopy,
//		Valid:         true,
//	}
//	instructionSet[REVERT] = Operation{
//		Execute:       opRevert,
//		GasCost:       gas.Revert,
//		ValidateStack: mm.MakeStackFunc(2, 0),
//		MemorySize:    mm.MemoryRevert,
//		Valid:         true,
//		Reverts:       true,
//		Returns:       true,
//	}
//	return instructionSet
//}
//
//// NewHomesteadInstructionSet 家园 版本支持的指令集
//func NewHomesteadInstructionSet() [256]Operation {
//	instructionSet := NewFrontierInstructionSet()
//	instructionSet[DELEGATECALL] = Operation{
//		Execute:       opDelegateCall,
//		GasCost:       gas.DelegateCall,
//		ValidateStack: mm.MakeStackFunc(6, 1),
//		MemorySize:    mm.MemoryDelegateCall,
//		Valid:         true,
//		Returns:       true,
//	}
//	return instructionSet
//}
//
//// NewFrontierInstructionSet 边境 版本支持的指令集
//func NewFrontierInstructionSet() [256]Operation {
//	return [256]Operation{
//		STOP: {
//			Execute:       opStop,
//			GasCost:       gas.ConstGasFunc(0),
//			ValidateStack: mm.MakeStackFunc(0, 0),
//			Halts:         true,
//			Valid:         true,
//		},
//		ADD: {
//			Execute:       opAdd,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		MUL: {
//			Execute:       opMul,
//			GasCost:       gas.ConstGasFunc(gas.GasFastStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		SUB: {
//			Execute:       opSub,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		DIV: {
//			Execute:       opDiv,
//			GasCost:       gas.ConstGasFunc(gas.GasFastStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		SDIV: {
//			Execute:       opSdiv,
//			GasCost:       gas.ConstGasFunc(gas.GasFastStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		MOD: {
//			Execute:       opMod,
//			GasCost:       gas.ConstGasFunc(gas.GasFastStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		SMOD: {
//			Execute:       opSmod,
//			GasCost:       gas.ConstGasFunc(gas.GasFastStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		ADDMOD: {
//			Execute:       opAddmod,
//			GasCost:       gas.ConstGasFunc(gas.GasMidStep),
//			ValidateStack: mm.MakeStackFunc(3, 1),
//			Valid:         true,
//		},
//		MULMOD: {
//			Execute:       opMulmod,
//			GasCost:       gas.ConstGasFunc(gas.GasMidStep),
//			ValidateStack: mm.MakeStackFunc(3, 1),
//			Valid:         true,
//		},
//		EXP: {
//			Execute:       opExp,
//			GasCost:       gas.Exp,
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		SIGNEXTEND: {
//			Execute:       opSignExtend,
//			GasCost:       gas.ConstGasFunc(gas.GasFastStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		LT: {
//			Execute:       opLt,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		GT: {
//			Execute:       opGt,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		SLT: {
//			Execute:       opSlt,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		SGT: {
//			Execute:       opSgt,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		EQ: {
//			Execute:       opEq,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		ISZERO: {
//			Execute:       opIszero,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(1, 1),
//			Valid:         true,
//		},
//		AND: {
//			Execute:       opAnd,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		XOR: {
//			Execute:       opXor,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		OR: {
//			Execute:       opOr,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		NOT: {
//			Execute:       opNot,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(1, 1),
//			Valid:         true,
//		},
//		BYTE: {
//			Execute:       opByte,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			Valid:         true,
//		},
//		SHA3: {
//			Execute:       opSha3,
//			GasCost:       gas.Sha3,
//			ValidateStack: mm.MakeStackFunc(2, 1),
//			MemorySize:    mm.MemorySha3,
//			Valid:         true,
//		},
//		ADDRESS: {
//			Execute:       opAddress,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		BALANCE: {
//			Execute:       opBalance,
//			GasCost:       gas.Balance,
//			ValidateStack: mm.MakeStackFunc(1, 1),
//			Valid:         true,
//		},
//		ORIGIN: {
//			Execute:       opOrigin,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		CALLER: {
//			Execute:       opCaller,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		CALLVALUE: {
//			Execute:       opCallValue,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		CALLDATALOAD: {
//			Execute:       opCallDataLoad,
//			GasCost:       gas.ConstGasFunc(gas.GasFastestStep),
//			ValidateStack: mm.MakeStackFunc(1, 1),
//			Valid:         true,
//		},
//		CALLDATASIZE: {
//			Execute:       opCallDataSize,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		CALLDATACOPY: {
//			Execute:       opCallDataCopy,
//			GasCost:       gas.CallDataCopy,
//			ValidateStack: mm.MakeStackFunc(3, 0),
//			MemorySize:    mm.MemoryCallDataCopy,
//			Valid:         true,
//		},
//		CODESIZE: {
//			Execute:       opCodeSize,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		CODECOPY: {
//			Execute:       opCodeCopy,
//			GasCost:       gas.CodeCopy,
//			ValidateStack: mm.MakeStackFunc(3, 0),
//			MemorySize:    mm.MemoryCodeCopy,
//			Valid:         true,
//		},
//		GASPRICE: {
//			Execute:       opGasprice,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		EXTCODESIZE: {
//			Execute:       opExtCodeSize,
//			GasCost:       gas.ExtCodeSize,
//			ValidateStack: mm.MakeStackFunc(1, 1),
//			Valid:         true,
//		},
//		EXTCODECOPY: {
//			Execute:       opExtCodeCopy,
//			GasCost:       gas.ExtCodeCopy,
//			ValidateStack: mm.MakeStackFunc(4, 0),
//			MemorySize:    mm.MemoryExtCodeCopy,
//			Valid:         true,
//		},
//		BLOCKHASH: {
//			Execute:       opBlockhash,
//			GasCost:       gas.ConstGasFunc(gas.GasExtStep),
//			ValidateStack: mm.MakeStackFunc(1, 1),
//			Valid:         true,
//		},
//		COINBASE: {
//			Execute:       opCoinbase,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		TIMESTAMP: {
//			Execute:       opTimestamp,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		NUMBER: {
//			Execute:       opNumber,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		DIFFICULTY: {
//			Execute:       opDifficulty,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		GASLIMIT: {
//			Execute:       opGasLimit,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		POP: {
//			Execute:       opPop,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(1, 0),
//			Valid:         true,
//		},
//		MLOAD: {
//			Execute:       opMload,
//			GasCost:       gas.MLoad,
//			ValidateStack: mm.MakeStackFunc(1, 1),
//			MemorySize:    mm.MemoryMLoad,
//			Valid:         true,
//		},
//		MSTORE: {
//			Execute:       opMstore,
//			GasCost:       gas.MStore,
//			ValidateStack: mm.MakeStackFunc(2, 0),
//			MemorySize:    mm.MemoryMStore,
//			Valid:         true,
//		},
//		MSTORE8: {
//			Execute:       opMstore8,
//			GasCost:       gas.MStore8,
//			MemorySize:    mm.MemoryMStore8,
//			ValidateStack: mm.MakeStackFunc(2, 0),
//
//			Valid: true,
//		},
//		SLOAD: {
//			Execute:       opSload,
//			GasCost:       gas.SLoad,
//			ValidateStack: mm.MakeStackFunc(1, 1),
//			Valid:         true,
//		},
//		SSTORE: {
//			Execute:       opSstore,
//			GasCost:       gas.SStore,
//			ValidateStack: mm.MakeStackFunc(2, 0),
//			Valid:         true,
//			Writes:        true,
//		},
//		JUMP: {
//			Execute:       opJump,
//			GasCost:       gas.ConstGasFunc(gas.GasMidStep),
//			ValidateStack: mm.MakeStackFunc(1, 0),
//			Jumps:         true,
//			Valid:         true,
//		},
//		JUMPI: {
//			Execute:       opJumpi,
//			GasCost:       gas.ConstGasFunc(gas.GasSlowStep),
//			ValidateStack: mm.MakeStackFunc(2, 0),
//			Jumps:         true,
//			Valid:         true,
//		},
//		PC: {
//			Execute:       opPc,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		MSIZE: {
//			Execute:       opMsize,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		GAS: {
//			Execute:       opGas,
//			GasCost:       gas.ConstGasFunc(gas.GasQuickStep),
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		JUMPDEST: {
//			Execute:       opJumpdest,
//			GasCost:       gas.ConstGasFunc(params.JumpdestGas),
//			ValidateStack: mm.MakeStackFunc(0, 0),
//			Valid:         true,
//		},
//		PUSH1: {
//			Execute:       makePush(1, 1),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH2: {
//			Execute:       makePush(2, 2),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH3: {
//			Execute:       makePush(3, 3),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH4: {
//			Execute:       makePush(4, 4),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH5: {
//			Execute:       makePush(5, 5),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH6: {
//			Execute:       makePush(6, 6),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH7: {
//			Execute:       makePush(7, 7),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH8: {
//			Execute:       makePush(8, 8),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH9: {
//			Execute:       makePush(9, 9),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH10: {
//			Execute:       makePush(10, 10),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH11: {
//			Execute:       makePush(11, 11),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH12: {
//			Execute:       makePush(12, 12),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH13: {
//			Execute:       makePush(13, 13),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH14: {
//			Execute:       makePush(14, 14),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH15: {
//			Execute:       makePush(15, 15),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH16: {
//			Execute:       makePush(16, 16),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH17: {
//			Execute:       makePush(17, 17),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH18: {
//			Execute:       makePush(18, 18),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH19: {
//			Execute:       makePush(19, 19),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH20: {
//			Execute:       makePush(20, 20),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH21: {
//			Execute:       makePush(21, 21),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH22: {
//			Execute:       makePush(22, 22),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH23: {
//			Execute:       makePush(23, 23),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH24: {
//			Execute:       makePush(24, 24),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH25: {
//			Execute:       makePush(25, 25),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH26: {
//			Execute:       makePush(26, 26),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH27: {
//			Execute:       makePush(27, 27),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH28: {
//			Execute:       makePush(28, 28),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH29: {
//			Execute:       makePush(29, 29),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH30: {
//			Execute:       makePush(30, 30),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH31: {
//			Execute:       makePush(31, 31),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		PUSH32: {
//			Execute:       makePush(32, 32),
//			GasCost:       gas.Push,
//			ValidateStack: mm.MakeStackFunc(0, 1),
//			Valid:         true,
//		},
//		DUP1: {
//			Execute:       makeDup(1),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(1),
//			Valid:         true,
//		},
//		DUP2: {
//			Execute:       makeDup(2),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(2),
//			Valid:         true,
//		},
//		DUP3: {
//			Execute:       makeDup(3),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(3),
//			Valid:         true,
//		},
//		DUP4: {
//			Execute:       makeDup(4),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(4),
//			Valid:         true,
//		},
//		DUP5: {
//			Execute:       makeDup(5),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(5),
//			Valid:         true,
//		},
//		DUP6: {
//			Execute:       makeDup(6),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(6),
//			Valid:         true,
//		},
//		DUP7: {
//			Execute:       makeDup(7),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(7),
//			Valid:         true,
//		},
//		DUP8: {
//			Execute:       makeDup(8),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(8),
//			Valid:         true,
//		},
//		DUP9: {
//			Execute:       makeDup(9),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(9),
//			Valid:         true,
//		},
//		DUP10: {
//			Execute:       makeDup(10),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(10),
//			Valid:         true,
//		},
//		DUP11: {
//			Execute:       makeDup(11),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(11),
//			Valid:         true,
//		},
//		DUP12: {
//			Execute:       makeDup(12),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(12),
//			Valid:         true,
//		},
//		DUP13: {
//			Execute:       makeDup(13),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(13),
//			Valid:         true,
//		},
//		DUP14: {
//			Execute:       makeDup(14),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(14),
//			Valid:         true,
//		},
//		DUP15: {
//			Execute:       makeDup(15),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(15),
//			Valid:         true,
//		},
//		DUP16: {
//			Execute:       makeDup(16),
//			GasCost:       gas.Dup,
//			ValidateStack: mm.MakeDupStackFunc(16),
//			Valid:         true,
//		},
//		SWAP1: {
//			Execute:       makeSwap(1),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(2),
//			Valid:         true,
//		},
//		SWAP2: {
//			Execute:       makeSwap(2),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(3),
//			Valid:         true,
//		},
//		SWAP3: {
//			Execute:       makeSwap(3),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(4),
//			Valid:         true,
//		},
//		SWAP4: {
//			Execute:       makeSwap(4),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(5),
//			Valid:         true,
//		},
//		SWAP5: {
//			Execute:       makeSwap(5),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(6),
//			Valid:         true,
//		},
//		SWAP6: {
//			Execute:       makeSwap(6),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(7),
//			Valid:         true,
//		},
//		SWAP7: {
//			Execute:       makeSwap(7),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(8),
//			Valid:         true,
//		},
//		SWAP8: {
//			Execute:       makeSwap(8),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(9),
//			Valid:         true,
//		},
//		SWAP9: {
//			Execute:       makeSwap(9),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(10),
//			Valid:         true,
//		},
//		SWAP10: {
//			Execute:       makeSwap(10),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(11),
//			Valid:         true,
//		},
//		SWAP11: {
//			Execute:       makeSwap(11),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(12),
//			Valid:         true,
//		},
//		SWAP12: {
//			Execute:       makeSwap(12),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(13),
//			Valid:         true,
//		},
//		SWAP13: {
//			Execute:       makeSwap(13),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(14),
//			Valid:         true,
//		},
//		SWAP14: {
//			Execute:       makeSwap(14),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(15),
//			Valid:         true,
//		},
//		SWAP15: {
//			Execute:       makeSwap(15),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(16),
//			Valid:         true,
//		},
//		SWAP16: {
//			Execute:       makeSwap(16),
//			GasCost:       gas.Swap,
//			ValidateStack: mm.MakeSwapStackFunc(17),
//			Valid:         true,
//		},
//		LOG0: {
//			Execute:       makeLog(0),
//			GasCost:       gas.MakeGasLog(0),
//			ValidateStack: mm.MakeStackFunc(2, 0),
//			MemorySize:    mm.MemoryLog,
//			Valid:         true,
//			Writes:        true,
//		},
//		LOG1: {
//			Execute:       makeLog(1),
//			GasCost:       gas.MakeGasLog(1),
//			ValidateStack: mm.MakeStackFunc(3, 0),
//			MemorySize:    mm.MemoryLog,
//			Valid:         true,
//			Writes:        true,
//		},
//		LOG2: {
//			Execute:       makeLog(2),
//			GasCost:       gas.MakeGasLog(2),
//			ValidateStack: mm.MakeStackFunc(4, 0),
//			MemorySize:    mm.MemoryLog,
//			Valid:         true,
//			Writes:        true,
//		},
//		LOG3: {
//			Execute:       makeLog(3),
//			GasCost:       gas.MakeGasLog(3),
//			ValidateStack: mm.MakeStackFunc(5, 0),
//			MemorySize:    mm.MemoryLog,
//			Valid:         true,
//			Writes:        true,
//		},
//		LOG4: {
//			Execute:       makeLog(4),
//			GasCost:       gas.MakeGasLog(4),
//			ValidateStack: mm.MakeStackFunc(6, 0),
//			MemorySize:    mm.MemoryLog,
//			Valid:         true,
//			Writes:        true,
//		},
//		CREATE: {
//			Execute:       opCreate,
//			GasCost:       gas.Create,
//			ValidateStack: mm.MakeStackFunc(3, 1),
//			MemorySize:    mm.MemoryCreate,
//			Valid:         true,
//			Writes:        true,
//			Returns:       true,
//		},
//		CALL: {
//			Execute:       opCall,
//			GasCost:       gas.Call,
//			ValidateStack: mm.MakeStackFunc(7, 1),
//			MemorySize:    mm.MemoryCall,
//			Valid:         true,
//			Returns:       true,
//		},
//		CALLCODE: {
//			Execute:       opCallCode,
//			GasCost:       gas.CallCode,
//			ValidateStack: mm.MakeStackFunc(7, 1),
//			MemorySize:    mm.MemoryCall,
//			Valid:         true,
//			Returns:       true,
//		},
//		RETURN: {
//			Execute:       opReturn,
//			GasCost:       gas.Return,
//			ValidateStack: mm.MakeStackFunc(2, 0),
//			MemorySize:    mm.MemoryReturn,
//			Halts:         true,
//			Valid:         true,
//		},
//		SELFDESTRUCT: {
//			Execute:       opSuicide,
//			GasCost:       gas.Suicide,
//			ValidateStack: mm.MakeStackFunc(1, 0),
//			Halts:         true,
//			Valid:         true,
//			Writes:        true,
//		},
//	}
//}
