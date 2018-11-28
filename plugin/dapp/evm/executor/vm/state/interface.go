// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package state

import (
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
)

// EVMStateDB 状态数据库封装，面向EVM业务执行逻辑；
// 生命周期为一个区块，在同一个区块内多个交易执行时使用的是同一个StateDB实例；
// StateDB包含区块的状态和交易的状态（当前上下文），所以不支持并发操作，区块内的多个交易只能按顺序单线程执行；
// StateDB除了查询状态数据，还会保留在交易执行时对数据的变更信息，每个交易完成之后会返回变更影响的数据给执行器；
type EVMStateDB interface {
	// CreateAccount 创建新的合约对象
	CreateAccount(string, string, string, string)

	// SubBalance 从指定地址扣除金额
	SubBalance(string, string, uint64)
	// AddBalance 向指定地址增加金额
	AddBalance(string, string, uint64)
	// GetBalance 获取指定地址的余额
	GetBalance(string) uint64

	// GetNonce 获取nonce值（只有合约对象有，外部对象为0）
	GetNonce(string) uint64
	// SetNonce 设置nonce值（只有合约对象有，外部对象为0）
	SetNonce(string, uint64)

	// GetCodeHash 获取指定地址合约的代码哈希
	GetCodeHash(string) common.Hash
	// GetCode 获取指定地址合约代码
	GetCode(string) []byte
	// SetCode 设置指定地址合约代码
	SetCode(string, []byte)
	// GetCodeSize 获取指定地址合约代码大小
	GetCodeSize(string) int
	// SetAbi 设置ABI内容
	SetAbi(addr, abi string)
	// GetAbi 获取ABI
	GetAbi(addr string) string

	// AddRefund 合约Gas奖励回馈
	AddRefund(uint64)
	// GetRefund 获取合约Gas奖励
	GetRefund() uint64

	// GetState 获取合约状态数据
	GetState(string, common.Hash) common.Hash
	// SetState 设置合约状态数据
	SetState(string, common.Hash, common.Hash)

	// Suicide 合约自销毁
	Suicide(string) bool
	// HasSuicided 合约是否已经销毁
	HasSuicided(string) bool

	// Exist 判断一个合约地址是否存在（已经销毁的合约地址对象依然存在）
	Exist(string) bool
	// Empty 判断一个合约地址是否为空（不包含任何代码、也没有余额的合约为空）
	Empty(string) bool

	// RevertToSnapshot 回滚到制定版本（从当前版本到回滚版本之间的数据变更全部撤销）
	RevertToSnapshot(int)
	// Snapshot 生成一个新版本号（递增操作）
	Snapshot() int
	// TransferStateData 转换合约状态数据存储
	TransferStateData(addr string)

	// AddLog 添加新的日志信息
	AddLog(*model.ContractLog)
	// AddPreimage 添加sha3记录
	AddPreimage(common.Hash, []byte)

	// CanTransfer 当前账户余额是否足够转账
	CanTransfer(sender, recipient string, amount uint64) bool
	// Transfer 转账交易
	Transfer(sender, recipient string, amount uint64) bool

	// GetBlockHeight 返回当前区块高度
	GetBlockHeight() int64
}
