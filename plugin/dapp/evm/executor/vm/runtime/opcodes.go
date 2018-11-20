// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// OpCode EVM操作码定义，本质上就是一个字节，所以操作码最多只支持256个
type OpCode byte

// IsPush 是否为压栈操作
func (op OpCode) IsPush() bool {
	if op >= PUSH1 && op <= PUSH32 {
		return true
	}
	return false
}

// IsStaticJump 是否为跳转操作
func (op OpCode) IsStaticJump() bool {
	return op == JUMP
}

var opmap map[OpCode]string

// String 打印字符串形式
func (op OpCode) String() string {
	if opmap == nil {
		initMap()
	}
	return opmap[op]
}

func initMap() {
	opmap = map[OpCode]string{
		// 0x0 range - arithmetic ops
		STOP:       "STOP",
		ADD:        "ADD",
		MUL:        "MUL",
		SUB:        "SUB",
		DIV:        "DIV",
		SDIV:       "SDIV",
		MOD:        "MOD",
		SMOD:       "SMOD",
		EXP:        "EXP",
		NOT:        "NOT",
		LT:         "LT",
		GT:         "GT",
		SLT:        "SLT",
		SGT:        "SGT",
		EQ:         "EQ",
		ISZERO:     "ISZERO",
		SIGNEXTEND: "SIGNEXTEND",

		// 0x10 range - bit ops
		AND:    "AND",
		OR:     "OR",
		XOR:    "XOR",
		BYTE:   "BYTE",
		SHL:    "SHL",
		SHR:    "SHR",
		SAR:    "SAR",
		ADDMOD: "ADDMOD",
		MULMOD: "MULMOD",

		// 0x20 range - crypto
		SHA3: "SHA3",

		// 0x30 range - closure state
		ADDRESS:        "ADDRESS",
		BALANCE:        "BALANCE",
		ORIGIN:         "ORIGIN",
		CALLER:         "CALLER",
		CALLVALUE:      "CALLVALUE",
		CALLDATALOAD:   "CALLDATALOAD",
		CALLDATASIZE:   "CALLDATASIZE",
		CALLDATACOPY:   "CALLDATACOPY",
		CODESIZE:       "CODESIZE",
		CODECOPY:       "CODECOPY",
		GASPRICE:       "GASPRICE",
		EXTCODESIZE:    "EXTCODESIZE",
		EXTCODECOPY:    "EXTCODECOPY",
		RETURNDATASIZE: "RETURNDATASIZE",
		RETURNDATACOPY: "RETURNDATACOPY",

		// 0x40 range - block operations
		BLOCKHASH:  "BLOCKHASH",
		COINBASE:   "COINBASE",
		TIMESTAMP:  "TIMESTAMP",
		NUMBER:     "NUMBER",
		DIFFICULTY: "DIFFICULTY",
		GASLIMIT:   "GASLIMIT",

		// 0x50 range - 'storage' and execution
		POP: "POP",
		//DUP:     "DUP",
		//SWAP:    "SWAP",
		MLOAD:    "MLOAD",
		MSTORE:   "MSTORE",
		MSTORE8:  "MSTORE8",
		SLOAD:    "SLOAD",
		SSTORE:   "SSTORE",
		JUMP:     "JUMP",
		JUMPI:    "JUMPI",
		PC:       "PC",
		MSIZE:    "MSIZE",
		GAS:      "GAS",
		JUMPDEST: "JUMPDEST",

		// 0x60 range - push
		PUSH1:  "PUSH1",
		PUSH2:  "PUSH2",
		PUSH3:  "PUSH3",
		PUSH4:  "PUSH4",
		PUSH5:  "PUSH5",
		PUSH6:  "PUSH6",
		PUSH7:  "PUSH7",
		PUSH8:  "PUSH8",
		PUSH9:  "PUSH9",
		PUSH10: "PUSH10",
		PUSH11: "PUSH11",
		PUSH12: "PUSH12",
		PUSH13: "PUSH13",
		PUSH14: "PUSH14",
		PUSH15: "PUSH15",
		PUSH16: "PUSH16",
		PUSH17: "PUSH17",
		PUSH18: "PUSH18",
		PUSH19: "PUSH19",
		PUSH20: "PUSH20",
		PUSH21: "PUSH21",
		PUSH22: "PUSH22",
		PUSH23: "PUSH23",
		PUSH24: "PUSH24",
		PUSH25: "PUSH25",
		PUSH26: "PUSH26",
		PUSH27: "PUSH27",
		PUSH28: "PUSH28",
		PUSH29: "PUSH29",
		PUSH30: "PUSH30",
		PUSH31: "PUSH31",
		PUSH32: "PUSH32",

		DUP1:  "DUP1",
		DUP2:  "DUP2",
		DUP3:  "DUP3",
		DUP4:  "DUP4",
		DUP5:  "DUP5",
		DUP6:  "DUP6",
		DUP7:  "DUP7",
		DUP8:  "DUP8",
		DUP9:  "DUP9",
		DUP10: "DUP10",
		DUP11: "DUP11",
		DUP12: "DUP12",
		DUP13: "DUP13",
		DUP14: "DUP14",
		DUP15: "DUP15",
		DUP16: "DUP16",

		SWAP1:  "SWAP1",
		SWAP2:  "SWAP2",
		SWAP3:  "SWAP3",
		SWAP4:  "SWAP4",
		SWAP5:  "SWAP5",
		SWAP6:  "SWAP6",
		SWAP7:  "SWAP7",
		SWAP8:  "SWAP8",
		SWAP9:  "SWAP9",
		SWAP10: "SWAP10",
		SWAP11: "SWAP11",
		SWAP12: "SWAP12",
		SWAP13: "SWAP13",
		SWAP14: "SWAP14",
		SWAP15: "SWAP15",
		SWAP16: "SWAP16",
		LOG0:   "LOG0",
		LOG1:   "LOG1",
		LOG2:   "LOG2",
		LOG3:   "LOG3",
		LOG4:   "LOG4",

		// 0xf0 range
		CREATE:       "CREATE",
		CALL:         "CALL",
		RETURN:       "RETURN",
		CALLCODE:     "CALLCODE",
		DELEGATECALL: "DELEGATECALL",
		STATICCALL:   "STATICCALL",
		REVERT:       "REVERT",
		SELFDESTRUCT: "SELFDESTRUCT",

		PUSH: "PUSH",
		DUP:  "DUP",
		SWAP: "SWAP",
	}
}

// unofficial opcodes used for parsing
const (
	// PUSH 压栈操作
	PUSH OpCode = 0xb0 + iota
	// DUP 操作
	DUP
	// SWAP 操作
	SWAP
)

const (
	// STOP 0x0 算术操作
	STOP OpCode = iota
	// ADD 操作
	ADD
	// MUL op
	MUL
	// SUB op
	SUB
	// DIV op
	DIV
	// SDIV op
	SDIV
	// MOD op
	MOD
	// SMOD op
	SMOD
	// ADDMOD op
	ADDMOD
	// MULMOD op
	MULMOD
	// EXP op
	EXP
	// SIGNEXTEND op
	SIGNEXTEND
)

const (
	// LT 比较、位操作
	LT OpCode = iota + 0x10
	// GT op
	GT
	// SLT op
	SLT
	// SGT op
	SGT
	// EQ op
	EQ
	// ISZERO op
	ISZERO
	// AND op
	AND
	// OR op
	OR
	// XOR op
	XOR
	// NOT op
	NOT
	// BYTE op
	BYTE
	// SHL op
	SHL
	// SHR op
	SHR
	// SAR op
	SAR

	// SHA3 op
	SHA3 = 0x20
)

const (
	// ADDRESS 0x30 合约数据操作
	ADDRESS OpCode = 0x30 + iota
	// BALANCE op
	BALANCE
	// ORIGIN op
	ORIGIN
	// CALLER op
	CALLER
	// CALLVALUE op
	CALLVALUE
	// CALLDATALOAD op
	CALLDATALOAD
	// CALLDATASIZE op
	CALLDATASIZE
	// CALLDATACOPY op
	CALLDATACOPY
	// CODESIZE op
	CODESIZE
	// CODECOPY op
	CODECOPY
	// GASPRICE op
	GASPRICE
	// EXTCODESIZE op
	EXTCODESIZE
	// EXTCODECOPY op
	EXTCODECOPY
	// RETURNDATASIZE op
	RETURNDATASIZE
	// RETURNDATACOPY op
	RETURNDATACOPY
)

const (
	// BLOCKHASH 0x40 区块相关操作
	BLOCKHASH OpCode = 0x40 + iota
	// COINBASE op
	COINBASE
	// TIMESTAMP op
	TIMESTAMP
	// NUMBER op
	NUMBER
	// DIFFICULTY op
	DIFFICULTY
	// GASLIMIT op
	GASLIMIT
)

const (
	// POP 0x50 存储相关操作
	POP OpCode = 0x50 + iota
	// MLOAD op
	MLOAD
	// MSTORE op
	MSTORE
	// MSTORE8 op
	MSTORE8
	// SLOAD op
	SLOAD
	// SSTORE op
	SSTORE
	// JUMP op
	JUMP
	// JUMPI op
	JUMPI
	// PC op
	PC
	// MSIZE op
	MSIZE
	// GAS op
	GAS
	// JUMPDEST op
	JUMPDEST
)

const (
	// PUSH1 0x60 栈操作
	PUSH1 OpCode = 0x60 + iota
	// PUSH2 op
	PUSH2
	// PUSH3 op
	PUSH3
	// PUSH4 op
	PUSH4
	// PUSH5 op
	PUSH5
	// PUSH6 op
	PUSH6
	// PUSH7 op
	PUSH7
	// PUSH8 op
	PUSH8
	// PUSH9 op
	PUSH9
	// PUSH10 op
	PUSH10
	// PUSH11 op
	PUSH11
	// PUSH12 op
	PUSH12
	// PUSH13 op
	PUSH13
	// PUSH14 op
	PUSH14
	// PUSH15 op
	PUSH15
	// PUSH16 op
	PUSH16
	// PUSH17 op
	PUSH17
	// PUSH18 op
	PUSH18
	// PUSH19 op
	PUSH19
	// PUSH20 op
	PUSH20
	// PUSH21 op
	PUSH21
	// PUSH22 op
	PUSH22
	// PUSH23 op
	PUSH23
	// PUSH24 op
	PUSH24
	// PUSH25 op
	PUSH25
	// PUSH26 op
	PUSH26
	// PUSH27 op
	PUSH27
	// PUSH28 op
	PUSH28
	// PUSH29 op
	PUSH29
	// PUSH30 op
	PUSH30
	// PUSH31 op
	PUSH31
	// PUSH32 op
	PUSH32
	// DUP1 op
	DUP1
	// DUP2 op
	DUP2
	// DUP3 op
	DUP3
	// DUP4 op
	DUP4
	// DUP5 op
	DUP5
	// DUP6 op
	DUP6
	// DUP7 op
	DUP7
	// DUP8 op
	DUP8
	// DUP9 op
	DUP9
	// DUP10 op
	DUP10
	// DUP11 op
	DUP11
	// DUP12 op
	DUP12
	// DUP13 op
	DUP13
	// DUP14 op
	DUP14
	// DUP15 op
	DUP15
	// DUP16 op
	DUP16
	// SWAP1 op
	SWAP1
	// SWAP2 op
	SWAP2
	// SWAP3 op
	SWAP3
	// SWAP4 op
	SWAP4
	// SWAP5 op
	SWAP5
	// SWAP6 op
	SWAP6
	// SWAP7 op
	SWAP7
	// SWAP8 op
	SWAP8
	// SWAP9 op
	SWAP9
	// SWAP10 op
	SWAP10
	// SWAP11 op
	SWAP11
	// SWAP12 op
	SWAP12
	// SWAP13 op
	SWAP13
	// SWAP14 op
	SWAP14
	// SWAP15 op
	SWAP15
	// SWAP16 op
	SWAP16
)

const (
	// LOG0 生成日志
	LOG0 OpCode = 0xa0 + iota
	// LOG1 op
	LOG1
	// LOG2 op
	LOG2
	// LOG3 op
	LOG3
	// LOG4 op
	LOG4
)

const (
	// CREATE 过程调用
	CREATE OpCode = 0xf0 + iota
	// CALL op
	CALL
	// CALLCODE op
	CALLCODE
	// RETURN op
	RETURN
	// DELEGATECALL op
	DELEGATECALL
	// STATICCALL  op
	STATICCALL = 0xfa

	// REVERT op
	REVERT = 0xfd
	// SELFDESTRUCT  op
	SELFDESTRUCT = 0xff
)
