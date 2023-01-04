// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"sync/atomic"

	"github.com/33cn/chain33/common/log/log15"

	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"

	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common/math"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/params"
)

// Config 解释器的配置模型
type Config struct {
	// Debug 调试开关
	Debug int32
	// Tracer 记录操作日志
	Tracer Tracer
	// NoRecursion 不允许使用Call, CallCode, DelegateCall
	NoRecursion bool
	// EnablePreimageRecording SHA3/keccak 操作时是否保存数据
	EnablePreimageRecording bool
	// JumpTable 指令跳转表
	JumpTable [256]*operation
}

// Interpreter  解释器接结构定义
type Interpreter struct {
	evm *EVM
	cfg Config
	// 是否允许修改数据
	readOnly bool
	// 合约执行返回的结果数据
	returnData []byte
}

const (
	EVMDebugOn  = int32(1)
	EVMDebugOff = int32(0)
)

// NewInterpreter 新创建一个解释器
func NewInterpreter(evm *EVM, cfg Config) *Interpreter {
	// 使用是否包含第一个STOP指令判断jump table是否完成初始化
	// 需要注意，后继如果新增指令，需要在这里判断硬分叉，指定不同的指令集
	if cfg.JumpTable[STOP] == nil {
		cfg.JumpTable = berlinInstructionSet
		if evm.cfg.IsDappFork(evm.StateDB.GetBlockHeight(), "evm", evmtypes.ForkEVMYoloV1) {
			//这里需要替换为最新得指令集
			cfg.JumpTable = berlinInstructionSet
		}
	}

	return &Interpreter{
		evm: evm,
		cfg: cfg,
	}
}

func (in *Interpreter) enforceRestrictions(op OpCode, operation *operation, stack *Stack) error {
	if in.readOnly {
		// 在只读状态下如果包含了写操作，
		// 也不允许进行转账操作（通过第二个条件可以判断）
		if operation.writes || (op == CALL && stack.Back(2).BitLen() > 0) {
			return model.ErrWriteProtection
		}
	}
	return nil
}

// Run 合约代码的解释执行主逻辑
// 需要注意的是，如果返回执行出错，依然会扣除剩余的Gas
// （除非返回的是ErrExecutionReverted，这种情况下会保留剩余的Gas）
func (in *Interpreter) Run(contract *Contract, input []byte, readOnly bool) (ret []byte, err error) {
	//TODO 切换为最新的管理方式,合约涉及转账报错？
	// 每次递归调用，深度加1
	in.evm.depth++
	defer func() { in.evm.depth-- }()
	// Make sure the readOnly is only set if we aren't in readOnly yet.
	// This makes also sure that the readOnly flag isn't removed for child calls.
	if readOnly && !in.readOnly {
		in.readOnly = true
		defer func() { in.readOnly = false }()
	}

	// 执行前讲返回数据置空
	in.returnData = nil

	// 无合约代码直接返回
	if len(contract.Code) == 0 {
		return nil, nil
	}

	var (
		// 当前操作指令码
		op OpCode
		// 内存空间
		mem = NewMemory()
		// 本地栈空间
		stack = newstack()
		// 本地返回的栈
		//returns     = mm.NewReturnStack() // local returns stack
		callContext = &callCtx{
			memory:   mem,
			stack:    stack,
			contract: contract,
		}
		// 指令计数器
		pc = uint64(0)
		// 操作消耗的Gas
		cost uint64
		// 在使用tracer打印调试日志时，复制一份下面的数据进行操作
		pcCopy  uint64
		gasCopy uint64
		logged  bool
		// 操作码执行函数的结果
		res []byte
	)
	contract.Input = input

	// 执行结束后，返还堆栈
	defer func() {
		returnStack(stack)
	}()

	if EVMDebugOn == in.cfg.Debug {
		defer func() {
			if err != nil {
				if !logged {
					in.cfg.Tracer.CaptureState(in.evm, pcCopy, op, gasCopy, cost, mem, stack, in.returnData, contract, in.evm.depth, err)
				} else {
					in.cfg.Tracer.CaptureFault(in.evm, pcCopy, op, gasCopy, cost, mem, stack, contract, in.evm.depth, err)
				}
			}
		}()
	}
	// 遍历合约代码中的指令执行，直到遇到特殊指令（停止、自毁、暂停、恢复、返回）
	steps := 0
	for {
		steps++
		if steps%1000 == 0 && atomic.LoadInt32(&in.evm.abort) != 0 {
			break
		}
		if EVMDebugOn == in.cfg.Debug {
			// 记录当前指令执行前的状态数据
			logged, pcCopy, gasCopy = false, pc, contract.Gas
		}

		// 从合约代码中获取具体操作指令
		op = contract.GetOp(pc)

		operation := in.cfg.JumpTable[op]
		if operation == nil {
			log15.Error("can't found operation:%s", op)
			return nil, &ErrInvalidOpCode{opcode: op}
		}

		// Validate stack
		if sLen := stack.len(); sLen < operation.minStack {
			return nil, &ErrStackUnderflow{stackLen: sLen, required: operation.minStack}
		} else if sLen > operation.maxStack {
			return nil, &ErrStackOverflow{stackLen: sLen, limit: operation.maxStack}
		}
		// 检查写约束
		if err := in.enforceRestrictions(op, operation, stack); err != nil {
			return nil, err
		}

		// Static portion of gas
		cost = operation.constantGas // For tracing
		if !contract.UseGas(operation.constantGas) {
			log15.Error("Run:outOfGas", "op=", op.String(), "contract addr=", contract.self.Address().String(),
				"CallerAddress=", contract.CallerAddress.String(),
				"caller=", contract.caller.Address().String())
			return nil, ErrOutOfGas
		}

		var memorySize uint64
		// 计算需要开辟的内存空间
		// Memory check needs to be done prior to evaluating the dynamic gas portion,
		// to detect calculation overflows
		if operation.memorySize != nil {
			memSize, overflow := operation.memorySize(stack)
			if overflow {
				return nil, ErrGasUintOverflow
			}
			// memory is expanded in words of 32 bytes. Gas
			// is also calculated in words.
			if memorySize, overflow = math.SafeMul(toWordSize(memSize), 32); overflow {
				return nil, ErrGasUintOverflow
			}
		}
		// Dynamic portion of gas
		// consume the gas and return an error if not enough gas is available.
		// cost is explicitly set so that the capture state defer method can get the proper cost
		// 计算本操作具体需要消耗的Gas
		if operation.dynamicGas != nil {
			var dynamicCost uint64
			dynamicCost, err = operation.dynamicGas(in.evm, contract, stack, mem, memorySize)
			cost += dynamicCost // total cost, for debug tracing
			if err != nil || !contract.UseGas(dynamicCost) {

				log15.Error("Run:outOfGas", "op=", op.String(), "contract addr=", contract.self.Address().String(),
					"CallerAddress=", contract.CallerAddress.String(),
					"caller=", contract.caller.Address().String(), "dynamicCost", dynamicCost)
				return nil, ErrOutOfGas
			}
		}
		if memorySize > 0 {
			mem.Resize(memorySize)
		}

		if EVMDebugOn == in.cfg.Debug {
			in.cfg.Tracer.CaptureState(in.evm, pc, op, gasCopy, cost, mem, stack, in.returnData, contract, in.evm.depth, err)
			logged = true
		}

		// 执行具体的指令操作逻辑（合约执行的核心）
		res, err = operation.execute(&pc, in.evm, callContext)
		// 如果本操作需要返回，则讲操作返回的结果最为合约执行的结果
		if operation.returns {
			in.returnData = common.CopyBytes(res)
		}

		switch {
		case err != nil:
			return nil, err
		case operation.reverts:
			return res, model.ErrExecutionReverted
		case operation.halts:
			return res, nil
		case !operation.jumps:
			pc++
		}
	}
	return nil, nil
}

// CanRun tells if the contract, passed as an argument, can be
// run by the current interpreter.
func (in *Interpreter) CanRun(code []byte) bool {
	return true
}

// 从Contract构造参数传递给GasFunc逻辑使用
// 目前只按需构造必要的参数，理论上GasFun进行Gas计算时可以使用Contract中的所有参数
// 后继视需要修改GasParam结构
func buildGasParam(contract *Contract) *params.GasParam {
	return &params.GasParam{Gas: contract.Gas, Address: contract.Address()}
}

// 从EVM构造参数传递给GasFunc逻辑使用
// 目前只按需构造必要的参数，理论上GasFun进行Gas计算时可以使用EVM中的所有参数
// 后继视需要修改EVMParam结构
func buildEVMParam(evm *EVM) *params.EVMParam {
	return &params.EVMParam{
		StateDB:     evm.StateDB,
		CallGasTemp: evm.callGasTemp,
		BlockNumber: evm.BlockNumber,
	}
}

// 使用操作结果反向填充EVM中的参数
// 之所以只设置CallGasTemp，是因为其它参数均为指针引用，参数中可以直接修改EVM中的状态
func fillEVM(param *params.EVMParam, evm *EVM) {
	evm.callGasTemp = param.CallGasTemp
}

// callCtx contains the things that are per-call, such as stack and memory,
// but not transients like pc and gas
type callCtx struct {
	memory *Memory
	stack  *Stack
	//rstack   *ReturnStack
	contract *Contract
}
