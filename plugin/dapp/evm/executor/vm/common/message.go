// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

// Message 合约交易消息模型
// 在EVM执行器中传递此消息，由外部Tx等价构造
type Message struct {
	to       *Address
	from     Address
	alias    string
	nonce    int64
	amount   uint64
	gasLimit uint64
	gasPrice uint32
	data     []byte
	para     []byte
}

// NewMessage 新建消息结构
func NewMessage(from Address, to *Address, nonce int64, amount uint64, gasLimit uint64, gasPrice uint32, data, para []byte, alias string) *Message {
	return &Message{
		from:     from,
		to:       to,
		nonce:    nonce,
		amount:   amount,
		gasLimit: gasLimit,
		gasPrice: gasPrice,
		data:     data,
		alias:    alias,
		para:     para,
	}
}

// From 来源
func (m Message) From() Address { return m.from }

// To 目的地址
func (m Message) To() *Address { return m.to }

// GasPrice Gas价格
func (m Message) GasPrice() uint32 { return m.gasPrice }

// Value 转账金额
func (m Message) Value() uint64 { return m.amount }

// Nonce  nonce值
func (m Message) Nonce() int64 { return m.nonce }

// Data 附带数据
func (m Message) Data() []byte { return m.data }

// GasLimit Gas限制
func (m Message) GasLimit() uint64 { return m.gasLimit }

// Alias 合约别名
func (m Message) Alias() string { return m.alias }

// Para 合约参数
func (m Message) Para() []byte { return m.para }
