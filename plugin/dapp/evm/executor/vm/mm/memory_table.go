// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mm

// 本文件中定义各种操作下计算内存大小的逻辑

type (
	// MemorySizeFunc 计算所需内存大小
	MemorySizeFunc func(*Stack) (size uint64, overflow bool)
)

//MemorySha3 sha3计算所需内存大小
func MemorySha3(stack *Stack) (uint64, bool) {
	return calcMemSize64(stack.Back(0), stack.Back(1))
}

//MemoryCallDataCopy callDataCopy所需内存大小
func MemoryCallDataCopy(stack *Stack) (uint64, bool) {
	return calcMemSize64(stack.Back(0), stack.Back(2))
}

//MemoryReturnDataCopy returnDataCopy所需内存大小
func MemoryReturnDataCopy(stack *Stack) (uint64, bool) {
	return calcMemSize64(stack.Back(0), stack.Back(2))
}

//MemoryCodeCopy codeCopy所需内存大小
func MemoryCodeCopy(stack *Stack) (uint64, bool) {
	return calcMemSize64(stack.Back(0), stack.Back(2))
}

//MemoryExtCodeCopy extCodeCopy所需内存大小
func MemoryExtCodeCopy(stack *Stack) (uint64, bool) {
	return calcMemSize64(stack.Back(1), stack.Back(3))
}

//MemoryMLoad mload所需内存大小
func MemoryMLoad(stack *Stack) (uint64, bool) {
	return calcMemSize64WithUint(stack.Back(0), 32)
}

//MemoryMStore8 mstore8所需内存大小
func MemoryMStore8(stack *Stack) (uint64, bool) {
	return calcMemSize64WithUint(stack.Back(0), 1)
}

//MemoryMStore mstore所需内存大小
func MemoryMStore(stack *Stack) (uint64, bool) {
	return calcMemSize64WithUint(stack.Back(0), 32)
}

//MemoryCreate create所需内存大小
func MemoryCreate(stack *Stack) (uint64, bool) {
	return calcMemSize64(stack.Back(1), stack.Back(2))
}

//func memoryCreate2(stack *Stack) (uint64, bool) {
//	return calcMemSize64(stack.Back(1), stack.Back(2))
//}

//MemoryCall call所需内存大小
func MemoryCall(stack *Stack) (uint64, bool) {
	x, overflow := calcMemSize64(stack.Back(5), stack.Back(6))
	if overflow {
		return 0, true
	}
	y, overflow := calcMemSize64(stack.Back(3), stack.Back(4))
	if overflow {
		return 0, true
	}
	if x > y {
		return x, false
	}
	return y, false
}

//MemoryDelegateCall delegateCall所需内存大小
func MemoryDelegateCall(stack *Stack) (uint64, bool) {
	x, overflow := calcMemSize64(stack.Back(4), stack.Back(5))
	if overflow {
		return 0, true
	}
	y, overflow := calcMemSize64(stack.Back(2), stack.Back(3))
	if overflow {
		return 0, true
	}
	if x > y {
		return x, false
	}
	return y, false
}

//MemoryStaticCall staticCall所需内存大小
func MemoryStaticCall(stack *Stack) (uint64, bool) {
	x, overflow := calcMemSize64(stack.Back(4), stack.Back(5))
	if overflow {
		return 0, true
	}
	y, overflow := calcMemSize64(stack.Back(2), stack.Back(3))
	if overflow {
		return 0, true
	}
	if x > y {
		return x, false
	}
	return y, false
}

//MemoryReturn return所需内存大小
func MemoryReturn(stack *Stack) (uint64, bool) {
	return calcMemSize64(stack.Back(0), stack.Back(1))
}

//MemoryRevert revert所需内存大小
func MemoryRevert(stack *Stack) (uint64, bool) {
	return calcMemSize64(stack.Back(0), stack.Back(1))
}

//MemoryLog log所需内存大小
func MemoryLog(stack *Stack) (uint64, bool) {
	return calcMemSize64(stack.Back(0), stack.Back(1))
}
