// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// CreateCallTx 创建或调用合约交易结构
type CreateCallTx struct {
	// Amount 金额
	Amount uint64 `json:"amount"`
	// Code 合约代码
	Code string `json:"code"`
	// GasLimit gas限制
	GasLimit uint64 `json:"gasLimit"`
	// GasPrice gas定价
	GasPrice uint32 `json:"gasPrice"`
	// Note 备注
	Note string `json:"note"`
	// Alias 合约别名
	Alias string `json:"alias"`
	// Fee 交易手续费
	Fee int64 `json:"fee"`
	// Name 交易名称
	Name string `json:"name"`
	// IsCreate 是否创建合约
	IsCreate bool `json:"isCreate"`
	// 调用参数
	Para string `json:"para"`
}

// UpdateTx 更新合约交易结构
type UpdateTx struct {
	// Address 合约地址
	Addr string `json:"addr"`
	// Amount 金额
	Amount uint64 `json:"amount"`
	// Code 合约代码
	Code string `json:"code"`
	// GasLimit gas限制
	GasLimit uint64 `json:"gasLimit"`
	// GasPrice gas定价
	GasPrice uint32 `json:"gasPrice"`
	// Note 备注
	Note string `json:"note"`
	// Alias 合约别名
	Alias string `json:"alias"`
	// Fee 交易手续费
	Fee int64 `json:"fee"`
}

// DestroyTx 销毁合约
type DestroyTx struct {
	// Addr 合约地址
	Addr string `json:"Addr"`
	// Fee 交易手续费
	Fee int64 `json:"fee"`
}

// FreezeTx 冻结合约
type FreezeTx struct {
	// Addr 合约地址
	Addr string `json:"Addr"`
	// Fee 交易手续费
	Fee int64 `json:"fee"`
}

// ReleaseTx 解冻合约
type ReleaseTx struct {
	// Addr 合约地址
	Addr string `json:"Addr"`
	// Fee 交易手续费
	Fee int64 `json:"fee"`
}
