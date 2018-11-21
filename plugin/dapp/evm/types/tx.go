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
}

// BindABI  绑定ABI的RPC请求结构
type BindABI struct {
	// Data 要绑定的ABI数据
	Data string `json:"data"`
	// Name 要绑定的EVM合约名称
	Name string `json:"name"`
	// Note 备注
	Note string `json:"note"`
}

// ABICall ABI方式调用的RPC请求结构
type ABICall struct {
	// Data ABI调用信息
	Data string `json:"data"`
	// Name 调用的合约名称
	Name string `json:"name"`
	// Amount 调用时传递的金额信息
	Amount uint64 `json:"amount"`
}
