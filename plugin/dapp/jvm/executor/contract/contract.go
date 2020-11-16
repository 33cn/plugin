package contract

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
)

// ContractRef 合约对象引用
type ContractRef interface {
	Address() address.Address
}

// AccountRef 账户对象引用 （实现了合约对象引用ContractRef接口）
// 因为在合约调用过程中，调用者有可能是外部账户，也有可能是合约账户，所以两者的结构是互通的
type AccountRef address.Address

// Address 将账户引用转换为普通地址对象
func (ar AccountRef) Address() address.Address { return (address.Address)(ar) }

// Contract 合约对象，它在内存中表示一个合约账户的代码即地址信息实体
// 每次合约调用都会创建一个新的合约对象
type Contract struct {
	// 调用者地址，应该为外部账户的地址
	// 如果是通过合约再调用合约时，会从上级合约中获取调用者地址进行赋值
	CallerAddress address.Address

	// 调用此合约的地址，有可能是外部地址（直接调用时），也有可能是合约地址（委托调用时）
	caller ContractRef

	// 一般情况下为合约自身地址
	// 但是，二般情况下（外部账户通过CallCode直接调用合约代码时，此地址会设置为外部账户的地址，就是和caller一样）
	self ContractRef

	// 合约代码和代码哈希
	Code     []byte
	CodeHash common.Hash

	// 合约地址以及输入参数
	CodeAddr address.Address
	Input    []byte

	// 合约调用的同时，如果包含转账逻辑，则此处为转账金额
	value uint64

	// 委托调用时，此属性会被设置为true
	DelegateCall bool
}

// NewContract 创建一个新的合约调用对象
// 不管合约是否存在，每次调用时都会新创建一个合约对象交给解释器执行，对象持有合约代码和合约地址
func NewContract(caller ContractRef, object ContractRef, value uint64) *Contract {
	c := &Contract{CallerAddress: caller.Address(), caller: caller, self: object}
	c.value = value

	return c
}

// AsDelegate 设置当前的合约对象为被委托调用
// 返回当前合约对象的指针，以便在多层调用链模式下使用
func (c *Contract) AsDelegate() *Contract {
	c.DelegateCall = true

	// 在委托调用模式下，调用者必定为合约对象，而非外部对象
	parent := c.caller.(*Contract)

	// 在一个多层的委托调用链中，调用者地址始终为最初发起合约调用的外部账户地址
	c.CallerAddress = parent.CallerAddress

	// 其它数据正常传递
	c.value = parent.value
	return c
}

// 获取合约代码中制定位置的操作码
//func (c *Contract) GetOp(n uint64) OpCode {
//	return OpCode(c.GetByte(n))
//}

// GetByte 获取合约代码中制定位置的字节值
func (c *Contract) GetByte(n uint64) byte {
	if n < uint64(len(c.Code)) {
		return c.Code[n]
	}

	return 0
}

// Caller 返回合约的调用者
// 如果当前合约为委托调用，则调用它的不是外部账户，而是合约账户，所以此时的caller为调用此合约的合约的caller
// 这个关系可以一直递归向上，直到定位到caller为外部账户地址
func (c *Contract) Caller() address.Address {
	return c.CallerAddress
}

// Address 返回上下文中合约自身的地址
// 注意，当合约通过CallCode调用时，这个地址并不是当前合约代码对应的地址，而是调用者的地址
func (c *Contract) Address() address.Address {
	return c.self.Address()
}

// Value 合约包含转账逻辑时，转账的金额
func (c *Contract) Value() uint64 {
	return c.value
}

// SetCode 设置合约代码内容
func (c *Contract) SetCode(hash common.Hash, code []byte) {
	c.Code = code
	c.CodeHash = hash
}

// SetCallCode 设置合约代码和代码哈希
func (c *Contract) SetCallCode(addr address.Address, hash common.Hash, code []byte) {
	c.Code = code
	c.CodeHash = hash
	c.CodeAddr = addr
}
