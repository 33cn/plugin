// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"math/big"
	"time"

	"encoding/json"
	"io"

	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/mm"
)

// Tracer 接口用来在合约执行过程中收集跟踪数据。
// CaptureState 会在EVM解释每条指令时调用。
// 需要注意的是，传入的引用参数不允许修改，否则会影响EVM解释执行；如果需要使用其中的数据，请复制后使用。
type Tracer interface {
	// CaptureStart 开始记录
	CaptureStart(from common.Address, to common.Address, call bool, input []byte, gas uint64, value uint64) error
	// CaptureState 保存状态
	CaptureState(env *EVM, pc uint64, op OpCode, gas, cost uint64, memory *mm.Memory, stack *mm.Stack, contract *Contract, depth int, err error) error
	// CaptureFault 保存错误
	CaptureFault(env *EVM, pc uint64, op OpCode, gas, cost uint64, memory *mm.Memory, stack *mm.Stack, contract *Contract, depth int, err error) error
	// CaptureEnd 结束记录
	CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) error
}

// JSONLogger 使用json格式打印日志
type JSONLogger struct {
	encoder *json.Encoder
}

// StructLog 指令执行状态信息
type StructLog struct {
	// Pc pc指针
	Pc uint64 `json:"pc"`
	// Op 操作码
	Op string `json:"op"`
	// Gas gas
	Gas uint64 `json:"gas"`
	// GasCost 花费
	GasCost uint64 `json:"gasCost"`
	// Memory 内存对象
	Memory []string `json:"memory"`
	// MemorySize 内存大小
	MemorySize int `json:"memSize"`
	// Stack 栈对象
	Stack []string `json:"stack"`
	// Storage 存储对象
	Storage map[common.Hash]common.Hash `json:"-"`
	// Depth 调用深度
	Depth int `json:"depth"`
	// Err 错误信息
	Err error `json:"-"`
}

// NewJSONLogger 创建新的日志记录器
func NewJSONLogger(writer io.Writer) *JSONLogger {
	return &JSONLogger{json.NewEncoder(writer)}
}

// CaptureStart 开始记录
func (logger *JSONLogger) CaptureStart(from common.Address, to common.Address, create bool, input []byte, gas uint64, value uint64) error {
	return nil
}

// CaptureState 输出当前虚拟机状态
func (logger *JSONLogger) CaptureState(env *EVM, pc uint64, op OpCode, gas, cost uint64, memory *mm.Memory, stack *mm.Stack, contract *Contract, depth int, err error) error {
	log := StructLog{
		Pc:         pc,
		Op:         op.String(),
		Gas:        gas,
		GasCost:    cost,
		MemorySize: memory.Len(),
		Storage:    nil,
		Depth:      depth,
		Err:        err,
	}
	log.Memory = formatMemory(memory.Data())
	log.Stack = formatStack(stack.Data())
	return logger.encoder.Encode(log)
}

func formatStack(data []*big.Int) (res []string) {
	for _, v := range data {
		res = append(res, v.Text(16))
	}
	return
}

func formatMemory(data []byte) (res []string) {
	for idx := 0; idx < len(data); idx += 32 {
		res = append(res, common.Bytes2HexTrim(data[idx:idx+32]))
	}
	return
}

// CaptureFault 目前实现为空
func (logger *JSONLogger) CaptureFault(env *EVM, pc uint64, op OpCode, gas, cost uint64, memory *mm.Memory, stack *mm.Stack, contract *Contract, depth int, err error) error {
	return nil
}

// CaptureEnd 结束记录
func (logger *JSONLogger) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) error {
	type endLog struct {
		Output  string        `json:"output"`
		GasUsed int64         `json:"gasUsed"`
		Time    time.Duration `json:"time"`
		Err     string        `json:"error,omitempty"`
	}

	if err != nil {
		return logger.encoder.Encode(endLog{common.Bytes2Hex(output), int64(gasUsed), t, err.Error()})
	}
	return logger.encoder.Encode(endLog{common.Bytes2Hex(output), int64(gasUsed), t, ""})
}
