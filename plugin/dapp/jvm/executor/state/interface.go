package state

import (
	"github.com/33cn/chain33/common"
)

// jvmStateDB 状态数据库封装，面向EVM业务执行逻辑；
// 生命周期为一个区块，在同一个区块内多个交易执行时使用的是同一个StateDB实例；
// StateDB包含区块的状态和交易的状态（当前上下文），所以不支持并发操作，区块内的多个交易只能按顺序单线程执行；
// StateDB除了查询状态数据，还会保留在交易执行时对数据的变更信息，每个交易完成之后会返回变更影响的数据给执行器；
type JvmStateDB interface {
	// 创建新的合约对象
	CreateAccount(string, string, string, string)

	// 从从指定地址扣除金额
	SubBalance(string, string, uint64)
	// 向指定地址增加金额
	AddBalance(string, string, uint64)
	// 获取指定地址的余额
	GetBalance(string) uint64

	// 获取指定地址合约的代码哈希
	GetCodeHash(string) common.Hash
	// 获取指定地址合约代码
	GetCode(string) []byte
	// 获取指定地址合约ABI
	GetAbi(string) []byte
	// 设置指定地址合约代码
	SetCodeAndAbi(string, []byte, []byte)
	// 获取指定地址合约代码大小
	GetCodeSize(string) int

	// 获取合约状态数据
	GetState(string, common.Hash) common.Hash
	// 设置合约状态数据
	SetState(string, common.Hash, common.Hash)

	// 判断一个合约地址是否存在（已经销毁的合约地址对象依然存在）
	Exist(string) bool
	// 判断一个合约地址是否为空（不包含任何代码、也没有余额的合约为空）
	Empty(string) bool

	// 生成一个新版本号（递增操作）
	Snapshot() int
}
