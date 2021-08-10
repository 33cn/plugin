// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package model //nolint

const (
	// WordBitSize 内存中存储的字，占用多少位
	WordBitSize = 256
	// WordByteSize 内存中存储的字，占用多少字节
	WordByteSize = WordBitSize / 8

	// StatisticEVMError evm内部错误
	StatisticEVMError = "evm"
	// StatisticExecError 执行器错误
	StatisticExecError = "exec"
	// StatisticGasError gas错误
	StatisticGasError = "gas"
)
