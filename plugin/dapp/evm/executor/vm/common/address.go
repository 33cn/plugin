// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"encoding/hex"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto/sha3"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/holiman/uint256"
)

// Address 封装地址结构体，并提供各种常用操作封装
// 这里封装的操作主要是为了提供Address<->big.Int， Address<->[]byte 之间的互相转换
// 并且转换的核心是使用地址对象中的Hash160元素，因为在EVM中地址固定为[20]byte，超出此范围的地址无法正确解释执行
type Address struct {
	Addr *address.Address
}

// Hash160Address EVM中使用的地址格式
type Hash160Address [Hash160Length]byte

// String 字符串结构
func (a Address) String() string { return a.Addr.String() }

// Bytes 字节数组
func (a Address) Bytes() []byte {
	return a.Addr.Hash160[:]
}

// Big 大数字
func (a Address) Big() *big.Int {
	ret := new(big.Int).SetBytes(a.Bytes())
	return ret
}

// Hash 计算地址哈希
func (a Address) Hash() Hash { return ToHash(a.Bytes()) }

// ToHash160 返回EVM类型地址
func (a Address) ToHash160() Hash160Address {
	var h Hash160Address
	h.SetBytes(a.Bytes())
	return h
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (h *Hash160Address) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-Hash160Length:]
	}
	copy(h[Hash160Length-len(b):], b)
}

// String implements fmt.Stringer.
func (h Hash160Address) String() string {
	return h.Hex()
}

// Hex returns an EIP55-compliant hex string representation of the address.
func (h Hash160Address) Hex() string {
	unchecksummed := hex.EncodeToString(h[:])
	sha := sha3.NewLegacyKeccak256()
	sha.Write([]byte(unchecksummed))
	hash := sha.Sum(nil)

	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return "0x" + string(result)
}

// ToAddress 返回Chain33格式的地址
func (h Hash160Address) ToAddress() Address {
	return BytesToAddress(h[:])
}

// NewAddress xHash生成EVM合约地址
func NewAddress(cfg *types.Chain33Config, txHash []byte) Address {
	execAddr := address.GetExecAddress(cfg.ExecName("user.evm.") + BytesToHash(txHash).Hex())
	return Address{Addr: execAddr}
}

func NewContractAddress(b Address, txHash []byte) Address {
	execAddr := address.GetExecAddress(b.String() + common.Bytes2Hex(txHash))
	return Address{Addr: execAddr}
}

// ExecAddress 返回合约地址
func ExecAddress(execName string) Address {
	execAddr := address.GetExecAddress(execName)
	return Address{Addr: execAddr}
}

// BytesToAddress 字节向地址转换
func BytesToAddress(b []byte) Address {
	a := new(address.Address)
	a.Version = 0
	a.SetBytes(copyBytes(LeftPadBytes(b, 20)))
	return Address{Addr: a}
}

// BytesToHash160Address 字节向地址转换
func BytesToHash160Address(b []byte) Hash160Address {
	var h Hash160Address
	h.SetBytes(b)
	return h
}

// StringToAddress 字符串转换为地址
func StringToAddress(s string) *Address {
	addr, err := address.NewAddrFromString(s)
	if err != nil {
		//检查是否是十六进制地址数据
		hbytes, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
		if err == nil {
			if len(hbytes) == 20 {
				var addr address.Address
				addr.SetBytes(hbytes)
				return &Address{Addr: &addr}
			}
		}
		log15.Error("create address form string error", "string:", s)
		return nil
	}
	return &Address{Addr: addr}
}

func copyBytes(data []byte) (out []byte) {
	out = make([]byte, 20)
	copy(out[:], data)
	return
}

func bigBytes(b *big.Int) (out []byte) {
	out = make([]byte, 20)
	copy(out[:], b.Bytes())
	return
}

// BigToAddress 大数字转换为地址
func BigToAddress(b *big.Int) Address {
	a := new(address.Address)
	a.Version = 0
	a.SetBytes(bigBytes(b))
	return Address{Addr: a}
}

// EmptyAddress 返回空地址
func EmptyAddress() Address { return BytesToAddress([]byte{0}) }

// HexToAddress returns Address with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
func HexToAddress(s string) Hash160Address { return BytesToHash160Address(FromHex(s)) }

// Uint256ToAddress 大数字转换为地址
func Uint256ToAddress(b *uint256.Int) Address {
	a := new(address.Address)
	a.Version = 0
	out := make([]byte, 20)

	copy(out[:], b.Bytes())
	a.SetBytes(out)
	return Address{Addr: a}
}

// HexToAddr 十六进制转换为虚拟机中的地址
func HexToAddr(s string) Address {
	a := new(address.Address)
	a.Version = 0
	out := make([]byte, 20)
	copy(out[:], FromHex(s))
	a.SetBytes(out)
	return Address{Addr: a}
}
