// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mm

import (
	"fmt"

	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/params"
)

type (
	// StackValidationFunc 校验栈中数据是否满足计算要求
	StackValidationFunc func(*Stack) error
)

// MakeStackFunc 栈校验的通用逻辑封装（主要就是检查栈的深度和空间是否够用）
func MakeStackFunc(pop, push int) StackValidationFunc {
	return func(stack *Stack) error {
		if err := stack.Require(pop); err != nil {
			return err
		}

		if stack.Len()+push-pop > int(params.StackLimit) {
			return fmt.Errorf("stack limit reached %d (%d)", stack.Len(), params.StackLimit)
		}
		return nil
	}
}

// MakeDupStackFunc 创建栈大小计算方法对象
func MakeDupStackFunc(n int) StackValidationFunc {
	return MakeStackFunc(n, n+1)
}

// MakeSwapStackFunc 创建栈大小计算方法对象
func MakeSwapStackFunc(n int) StackValidationFunc {
	return MakeStackFunc(n, n)
}
