// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"math/big"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// Address 封装地址结构体，并提供各种常用操作封装
// 这里封装的操作主要是为了提供Address<->big.Int， Address<->[]byte 之间的互相转换
// 并且转换的核心是使用地址对象中的Hash160元素，因为在EVM中地址固定为[20]byte，超出此范围的地址无法正确解释执行
type Address struct {
	addr *address.Address
}

// String 字符串结构
func (a Address) String() string { return a.addr.String() }

// Bytes 字节数组
func (a Address) Bytes() []byte {
	return a.addr.Hash160[:]
}

// Big 大数字
func (a Address) Big() *big.Int {
	ret := new(big.Int).SetBytes(a.Bytes())
	return ret
}

// NewAddress xHash生成EVM合约地址
func NewAddress(txHash []byte) Address {
	execAddr := address.GetExecAddress(types.ExecName(evmtypes.EvmPrefix) + BytesToHash(txHash).Hex())
	return Address{addr: execAddr}
}

// ExecAddress 返回合约地址
func ExecAddress(execName string) Address {
	execAddr := address.GetExecAddress(execName)
	return Address{addr: execAddr}
}

// Hash 计算地址哈希
func (a Address) Hash() Hash { return ToHash(a.Bytes()) }

// BytesToAddress 字节向地址转换
func BytesToAddress(b []byte) Address {
	a := new(address.Address)
	a.Version = 0
	a.Hash160 = copyBytes(LeftPadBytes(b, 20))
	return Address{addr: a}
}

// StringToAddress 字符串转换为地址
func StringToAddress(s string) *Address {
	addr, err := address.NewAddrFromString(s)
	if err != nil {
		log15.Error("create address form string error", "string:", s)
		return nil
	}
	return &Address{addr: addr}
}

func copyBytes(data []byte) (out [20]byte) {
	copy(out[:], data)
	return
}

func bigBytes(b *big.Int) (out [20]byte) {
	copy(out[:], b.Bytes())
	return
}

// BigToAddress 大数字转换为地址
func BigToAddress(b *big.Int) Address {
	a := new(address.Address)
	a.Version = 0
	a.Hash160 = bigBytes(b)
	return Address{addr: a}
}

// EmptyAddress 返回空地址
func EmptyAddress() Address { return BytesToAddress([]byte{0}) }
