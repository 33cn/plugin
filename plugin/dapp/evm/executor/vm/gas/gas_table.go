// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gas

import (
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/mm"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/params"
)

// 本文件中定义各种操作中需要花费的Gas逻辑

type (
	// CalcGasFunc 计算Gas的方法定义
	CalcGasFunc func(Table, *params.EVMParam, *params.GasParam, *mm.Stack, *mm.Memory, uint64) (uint64, error) // last parameter is the requested memory size as a uint64
)

// Table 此文件中定义各种指令操作花费的Gas计算
// Gas定价表结构
type Table struct {
	// ExtcodeSize 扩展代码大小计价
	ExtcodeSize uint64
	// ExtcodeCopy 代码复制价格
	ExtcodeCopy uint64
	// Balance 账户计价
	Balance uint64
	// SLoad 加载数据计价
	SLoad uint64
	// Calls 调用方法计价
	Calls uint64
	// Suicide 自杀计价
	Suicide uint64
	// ExpByte 额外数据计价
	ExpByte uint64
}

var (
	// TableHomestead 定义各种操作的Gas定价
	TableHomestead = Table{
		ExtcodeSize: 20,
		ExtcodeCopy: 20,
		Balance:     20,
		SLoad:       50,
		Calls:       40,
		Suicide:     0,
		ExpByte:     10,
	}
)

// 计算新开辟内存空间需要使用多少Gas
func memoryGasCost(mem *mm.Memory, newMemSize uint64) (uint64, error) {
	if newMemSize == 0 {
		return 0, nil
	}

	// 如果超过最大值，则溢出
	if newMemSize > MaxNewMemSize {
		return 0, model.ErrGasUintOverflow
	}

	newMemSizeWords := common.ToWordSize(newMemSize)
	// 这里之所以要再算一遍，是因为内存开辟是按字长，
	// 第一次计算出来的自己长度不一定是字长的整数倍
	newMemSize = newMemSizeWords * 32

	if newMemSize > uint64(mem.Len()) {
		square := newMemSizeWords * newMemSizeWords
		linCoef := newMemSizeWords * params.MemoryGas
		quadCoef := square / params.QuadCoeffDiv
		newTotalFee := linCoef + quadCoef

		// 本次逻辑只返回新增的内存空间需要的Gas，所以需要减去上次已经花费的Gas
		fee := newTotalFee - mem.LastGasCost
		mem.LastGasCost = newTotalFee

		return fee, nil
	}
	return 0, nil
}

// ConstGasFunc Gas计算逻辑封装，返回固定值的Gas计算都可以使用此方法
func ConstGasFunc(gas uint64) CalcGasFunc {
	return func(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
		return gas, nil
	}
}

// CallDataCopy 计算数据复制需要花费的Gas
func CallDataCopy(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = common.SafeAdd(gas, GasFastestStep); overflow {
		return 0, model.ErrGasUintOverflow
	}

	words, overflow := common.BigUint64(stack.Back(2))
	if overflow {
		return 0, model.ErrGasUintOverflow
	}

	if words, overflow = common.SafeMul(common.ToWordSize(words), params.CopyGas); overflow {
		return 0, model.ErrGasUintOverflow
	}

	if gas, overflow = common.SafeAdd(gas, words); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// ReturnDataCopy 计算数据复制的价格
func ReturnDataCopy(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = common.SafeAdd(gas, GasFastestStep); overflow {
		return 0, model.ErrGasUintOverflow
	}

	words, overflow := common.BigUint64(stack.Back(2))
	if overflow {
		return 0, model.ErrGasUintOverflow
	}

	if words, overflow = common.SafeMul(common.ToWordSize(words), params.CopyGas); overflow {
		return 0, model.ErrGasUintOverflow
	}

	if gas, overflow = common.SafeAdd(gas, words); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// SStore 计算数据存储的价格
func SStore(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	var (
		y, x = stack.Back(1), stack.Back(0)
		val  = evm.StateDB.GetState(contractGas.Address.String(), common.BigToHash(x))
	)

	// 三种场景消耗的Gas是不一样的
	if val == (common.Hash{}) && y.Sign() != 0 {
		// 从零值地址到非零值地址存储， 赋值的情况
		// 0 => non 0
		return params.SstoreSetGas, nil
	} else if val != (common.Hash{}) && y.Sign() == 0 {
		// 从非零值地址到零值地址存储， 删除值的情况
		// non 0 => 0
		evm.StateDB.AddRefund(params.SstoreRefundGas)
		return params.SstoreClearGas, nil
	} else {
		// 从非零值地址到非零值地址存储， 变更值的情况
		// non 0 => non 0
		return params.SstoreResetGas, nil
	}
}

// MakeGasLog 生成Gas计算方法
func MakeGasLog(n uint64) CalcGasFunc {
	return func(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
		requestedSize, overflow := common.BigUint64(stack.Back(1))
		if overflow {
			return 0, model.ErrGasUintOverflow
		}

		gas, err := memoryGasCost(mem, memorySize)
		if err != nil {
			return 0, err
		}

		if gas, overflow = common.SafeAdd(gas, params.LogGas); overflow {
			return 0, model.ErrGasUintOverflow
		}
		if gas, overflow = common.SafeAdd(gas, n*params.LogTopicGas); overflow {
			return 0, model.ErrGasUintOverflow
		}

		var memorySizeGas uint64
		if memorySizeGas, overflow = common.SafeMul(requestedSize, params.LogDataGas); overflow {
			return 0, model.ErrGasUintOverflow
		}
		if gas, overflow = common.SafeAdd(gas, memorySizeGas); overflow {
			return 0, model.ErrGasUintOverflow
		}
		return gas, nil
	}
}

// Sha3 sha3计费
func Sha3(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	if gas, overflow = common.SafeAdd(gas, params.Sha3Gas); overflow {
		return 0, model.ErrGasUintOverflow
	}

	wordGas, overflow := common.BigUint64(stack.Back(1))
	if overflow {
		return 0, model.ErrGasUintOverflow
	}
	if wordGas, overflow = common.SafeMul(common.ToWordSize(wordGas), params.Sha3WordGas); overflow {
		return 0, model.ErrGasUintOverflow
	}
	if gas, overflow = common.SafeAdd(gas, wordGas); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// CodeCopy 代码复制计费
func CodeCopy(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = common.SafeAdd(gas, GasFastestStep); overflow {
		return 0, model.ErrGasUintOverflow
	}

	wordGas, overflow := common.BigUint64(stack.Back(2))
	if overflow {
		return 0, model.ErrGasUintOverflow
	}
	if wordGas, overflow = common.SafeMul(common.ToWordSize(wordGas), params.CopyGas); overflow {
		return 0, model.ErrGasUintOverflow
	}
	if gas, overflow = common.SafeAdd(gas, wordGas); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// ExtCodeCopy 扩展代码复制计费
func ExtCodeCopy(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = common.SafeAdd(gas, gt.ExtcodeCopy); overflow {
		return 0, model.ErrGasUintOverflow
	}

	wordGas, overflow := common.BigUint64(stack.Back(3))
	if overflow {
		return 0, model.ErrGasUintOverflow
	}

	if wordGas, overflow = common.SafeMul(common.ToWordSize(wordGas), params.CopyGas); overflow {
		return 0, model.ErrGasUintOverflow
	}

	if gas, overflow = common.SafeAdd(gas, wordGas); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// MLoad 内存加载计费
func MLoad(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, model.ErrGasUintOverflow
	}
	if gas, overflow = common.SafeAdd(gas, GasFastestStep); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// MStore8 内存存储计费
func MStore8(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, model.ErrGasUintOverflow
	}
	if gas, overflow = common.SafeAdd(gas, GasFastestStep); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// MStore 内存存储计费
func MStore(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, model.ErrGasUintOverflow
	}
	if gas, overflow = common.SafeAdd(gas, GasFastestStep); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// Create 开辟内存计费
func Create(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	if gas, overflow = common.SafeAdd(gas, params.CreateGas); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// Balance 获取余额计费
func Balance(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	return gt.Balance, nil
}

// ExtCodeSize 获取代码大小计费
func ExtCodeSize(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	return gt.ExtcodeSize, nil
}

// SLoad 加载存储计费
func SLoad(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	return gt.SLoad, nil
}

// Exp exp运算计费
func Exp(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	expByteLen := uint64((stack.Items[stack.Len()-2].BitLen() + 7) / 8)

	var (
		gas      = expByteLen * gt.ExpByte // no overflow check required. Max is 256 * ExpByte gas
		overflow bool
	)
	if gas, overflow = common.SafeAdd(gas, GasSlowStep); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// Call 调用合约计费
func Call(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	var (
		gas            = gt.Calls
		transfersValue = stack.Back(2).Sign() != 0
		address        = common.BigToAddress(stack.Back(1))
	)
	if !evm.StateDB.Exist(address.String()) {
		gas += params.CallNewAccountGas
	}
	if transfersValue {
		gas += params.CallValueTransferGas
	}
	memoryGas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = common.SafeAdd(gas, memoryGas); overflow {
		return 0, model.ErrGasUintOverflow
	}

	evm.CallGasTemp, err = callGas(gt, contractGas.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = common.SafeAdd(gas, evm.CallGasTemp); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// CallCode 调用合约代码计费
func CallCode(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	gas := gt.Calls
	if stack.Back(2).Sign() != 0 {
		gas += params.CallValueTransferGas
	}
	memoryGas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = common.SafeAdd(gas, memoryGas); overflow {
		return 0, model.ErrGasUintOverflow
	}

	evm.CallGasTemp, err = callGas(gt, contractGas.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = common.SafeAdd(gas, evm.CallGasTemp); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// Return 返回操作计费
func Return(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	return memoryGasCost(mem, memorySize)
}

// Revert revert操作计费
func Revert(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	return memoryGasCost(mem, memorySize)
}

// Suicide 自杀操作计费
func Suicide(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	var gas uint64
	if !evm.StateDB.HasSuicided(contractGas.Address.String()) {
		evm.StateDB.AddRefund(params.SuicideRefundGas)
	}
	return gas, nil
}

// DelegateCall 委托调用计费
func DelegateCall(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = common.SafeAdd(gas, gt.Calls); overflow {
		return 0, model.ErrGasUintOverflow
	}

	evm.CallGasTemp, err = callGas(gt, contractGas.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = common.SafeAdd(gas, evm.CallGasTemp); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// StaticCall 静态调用计费
func StaticCall(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = common.SafeAdd(gas, gt.Calls); overflow {
		return 0, model.ErrGasUintOverflow
	}

	evm.CallGasTemp, err = callGas(gt, contractGas.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = common.SafeAdd(gas, evm.CallGasTemp); overflow {
		return 0, model.ErrGasUintOverflow
	}
	return gas, nil
}

// Push 压栈计费
func Push(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}

// Swap 交换计费
func Swap(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}

// Dup dup操作计费
func Dup(gt Table, evm *params.EVMParam, contractGas *params.GasParam, stack *mm.Stack, mem *mm.Memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}
