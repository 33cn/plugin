// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mm

import (
	"math/big"

	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
)

// 本文件中定义各种操作下计算内存大小的逻辑

type (
	// MemorySizeFunc 计算所需内存大小
	MemorySizeFunc func(*Stack) *big.Int
)

// MemorySha3 sha3计算所需内存大小
func MemorySha3(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(0), stack.Back(1))
}

//MemoryCallDataCopy callDataCopy所需内存大小
func MemoryCallDataCopy(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(0), stack.Back(2))
}

//MemoryReturnDataCopy returnDataCopy所需内存大小
func MemoryReturnDataCopy(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(0), stack.Back(2))
}

//MemoryCodeCopy codeCopy所需内存大小
func MemoryCodeCopy(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(0), stack.Back(2))
}

//MemoryExtCodeCopy extCodeCopy所需内存大小
func MemoryExtCodeCopy(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(1), stack.Back(3))
}

//MemoryMLoad mload所需内存大小
func MemoryMLoad(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(0), big.NewInt(32))
}

//MemoryMStore8 mstore8所需内存大小
func MemoryMStore8(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(0), big.NewInt(1))
}

//MemoryMStore mstore所需内存大小
func MemoryMStore(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(0), big.NewInt(32))
}

//MemoryCreate create所需内存大小
func MemoryCreate(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(1), stack.Back(2))
}

//MemoryCall call所需内存大小
func MemoryCall(stack *Stack) *big.Int {
	x := calcMemSize(stack.Back(5), stack.Back(6))
	y := calcMemSize(stack.Back(3), stack.Back(4))

	return common.BigMax(x, y)
}

//MemoryDelegateCall delegateCall所需内存大小
func MemoryDelegateCall(stack *Stack) *big.Int {
	x := calcMemSize(stack.Back(4), stack.Back(5))
	y := calcMemSize(stack.Back(2), stack.Back(3))

	return common.BigMax(x, y)
}

//MemoryStaticCall staticCall所需内存大小
func MemoryStaticCall(stack *Stack) *big.Int {
	x := calcMemSize(stack.Back(4), stack.Back(5))
	y := calcMemSize(stack.Back(2), stack.Back(3))

	return common.BigMax(x, y)
}

//MemoryReturn return所需内存大小
func MemoryReturn(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(0), stack.Back(1))
}

//MemoryRevert revert所需内存大小
func MemoryRevert(stack *Stack) *big.Int {
	return calcMemSize(stack.Back(0), stack.Back(1))
}

//MemoryLog log所需内存大小
func MemoryLog(stack *Stack) *big.Int {
	mSize, mStart := stack.Back(1), stack.Back(0)
	return calcMemSize(mStart, mSize)
}
