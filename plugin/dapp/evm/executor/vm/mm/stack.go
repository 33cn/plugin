// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mm

import (
	"fmt"
	"sync"

	"github.com/holiman/uint256"
)

var stackPool = sync.Pool{
	New: func() interface{} {
		return &Stack{data: make([]uint256.Int, 0, 16)}
	},
}

// Stack 栈对象封装，提供常用的栈操作
type Stack struct {
	data []uint256.Int
}

// NewStack 新创建栈对象
func NewStack() *Stack {
	return stackPool.Get().(*Stack)
}

// Returnstack 把用完的stack还给stackpool
func Returnstack(s *Stack) {
	s.data = s.data[:0]
	stackPool.Put(s)
}

// Data 返回栈中的所有底层数据
func (st *Stack) Data() []uint256.Int {
	return st.data
}

// Push 数据入栈
func (st *Stack) Push(d *uint256.Int) {
	// NOTE push limit (1024) is checked in baseCheck
	st.data = append(st.data, *d)
}

// PushN 同时压栈多个数据
func (st *Stack) PushN(ds ...uint256.Int) {
	// FIXME: Is there a way to pass args by pointers.
	st.data = append(st.data, ds...)
}

// Pop 弹出栈顶数据
func (st *Stack) Pop() (ret uint256.Int) {
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

// Len 栈长度
func (st *Stack) Len() int {
	return len(st.data)
}

// Swap 将栈顶数据和栈中指定位置的数据互换位置
func (st *Stack) Swap(n int) {
	st.data[st.Len()-n], st.data[st.Len()-1] = st.data[st.Len()-1], st.data[st.Len()-n]
}

// Dup 复制栈中指定位置的数据的栈顶
func (st *Stack) Dup(n int) {
	st.Push(&st.data[st.Len()-n])
}

// Peek 返回顶端数据
func (st *Stack) Peek() *uint256.Int {
	return &st.data[st.Len()-1]
}

// Back 返回第n个取值
func (st *Stack) Back(n int) *uint256.Int {
	return &st.data[st.Len()-n-1]
}

// Require 检查栈是否满足长度要求
func (st *Stack) Require(n int) error {
	if st.Len() < n {
		return fmt.Errorf("stack underflow (%d <=> %d)", len(st.data), n)
	}
	return nil
}

// Print 印栈对象（调试用）
func (st *Stack) Print() {
	fmt.Println("### stack ###")
	if len(st.data) > 0 {
		for i, val := range st.data {
			fmt.Printf("%-3d  %v\n", i, val)
		}
	} else {
		fmt.Println("-- empty --")
	}
	fmt.Println("#############")
}

var rStackPool = sync.Pool{
	New: func() interface{} {
		return &ReturnStack{data: make([]uint32, 0, 10)}
	},
}

// ReturnStack 返回栈对象
type ReturnStack struct {
	data []uint32
}

// NewReturnStack 返回栈对象封装，提供常用的返回栈操作
func NewReturnStack() *ReturnStack {
	return rStackPool.Get().(*ReturnStack)
}

// ReturnRStack 将returnStack还给rStackPool
func ReturnRStack(rs *ReturnStack) {
	rs.data = rs.data[:0]
	rStackPool.Put(rs)
}

// Push 压栈
func (st *ReturnStack) Push(d uint32) {
	st.data = append(st.data, d)
}

// Pop  A uint32 is sufficient as for code below 4.2G
func (st *ReturnStack) Pop() (ret uint32) {
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

// Len ReturnStack大小
func (st *ReturnStack) Len() int {
	return len(st.data)
}

// Data 返回栈中的所有底层数据
func (st *ReturnStack) Data() []uint32 {
	return st.data
}
