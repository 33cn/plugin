// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/33cn/chain33/system/address/btc"

	"encoding/hex"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto/sha3"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/holiman/uint256"
)

var (
	// 地址驱动, 适配chain33和evm间的地址格式差异
	evmAddressDriver address.Driver
	lock             sync.RWMutex
)

// 设置默认值, btc地址格式
func init() {
	evmAddressDriver, _ = address.LoadDriver(btc.NormalAddressID, -1)
}

// InitEvmAddressDriver 初始化地址类型
func InitEvmAddressDriver(driver address.Driver) {
	lock.Lock()
	defer lock.Unlock()
	evmAddressDriver = driver
}

// GetEvmAddressDriver get driver
func GetEvmAddressDriver() address.Driver {
	lock.RLock()
	defer lock.RUnlock()
	return evmAddressDriver
}

// Address 封装evm内部地址对象
// raw为地址原始数据
// formatAddr为chain33框架中格式化地址, 相关转换格式由默认地址插件指定
// chain33 => evm, 即将formatAddr转换为raw数据, [20]byte
// evm => chain33, 即将原始数据raw格式化为formatAddr
type Address struct {
	raw        [AddressLength]byte
	formatAddr string
}

// Hash160Address EVM中使用的地址格式
type Hash160Address [Hash160Length]byte

// SetBytes sets the address to the value of b.
// If b is larger than len(a), b will be cropped from the left.
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a.raw) {
		b = b[len(b)-AddressLength:]
	}
	copy(a.raw[AddressLength-len(b):], b)
}

// String 字符串结构
func (a Address) String() string {
	if a.formatAddr == "" {
		a.formatAddr = evmAddressDriver.ToString(a.raw[:])
	}
	return a.formatAddr
}

// Bytes 字节数组
func (a Address) Bytes() []byte {
	return a.raw[:]
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
	execPub := address.ExecPubKey(ToHash(append(txHash, []byte(cfg.ExecName("user.evm."))...)).Str())
	return PubKey2Address(execPub)
}

func NewContractAddress(b Address, txHash []byte) Address {
	execPub := address.ExecPubKey(ToHash(append(txHash, b.Bytes()...)).Str())
	return PubKey2Address(execPub)
}

//NewEvmContractAddress  通过nonce创建合约地址
func NewEvmContractAddress(b Address, nonce uint64) Address {
	execAddr := ethcrypto.CreateAddress(common.BytesToAddress(b.Bytes()), nonce)
	var a Address
	a.formatAddr = evmAddressDriver.FormatAddr(execAddr.String())
	raw, _ := evmAddressDriver.FromString(a.formatAddr)
	a.SetBytes(raw)
	return a
}

// PubKey2Address pub key to address
func PubKey2Address(pub []byte) Address {

	execAddr := evmAddressDriver.PubKeyToAddr(pub)
	var a Address
	a.formatAddr = execAddr
	raw, _ := evmAddressDriver.FromString(execAddr)
	a.SetBytes(raw)
	return a
}

// ExecAddress 返回合约地址
func ExecAddress(execName string) Address {
	execPub := address.ExecPubKey(execName)
	return PubKey2Address(execPub)
}

// BytesToAddress 字节向地址转换
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// BytesToHash160Address 字节向地址转换
func BytesToHash160Address(b []byte) Hash160Address {
	var h Hash160Address
	h.SetBytes(b)
	return h
}

// StringToAddressLegacy 字符串转换为地址
// Deprecated
func StringToAddressLegacy(s string) *Address {
	raw, err := evmAddressDriver.FromString(s)
	if err != nil {
		//检查是否是十六进制地址数据
		raw, err = hex.DecodeString(strings.TrimPrefix(s, "0x"))
		if err != nil {
			log15.Error("create address form string error", "string:", s)
			return nil
		}
	}
	a := &Address{}
	a.SetBytes(raw)
	return a
}

// StringToAddress try convert string to Address
func StringToAddress(addr string) *Address {

	a := &Address{}
	// 以太坊地址类型直接解析
	if address.IsEthAddress(addr) {
		a.SetBytes(common.HexToAddress(addr).Bytes())
		return a
	}
	// 其他地址类型,尝试用框架地址驱动解析
	raw, err := address.GetDefaultAddressDriver().FromString(addr)
	if err != nil {
		log15.Error("decode address from string error", "addr:", addr)
		return nil
	}
	a.SetBytes(raw)
	return a
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
	var a Address
	a.SetBytes(bigBytes(b))
	return a
}

// EmptyAddress 返回空地址
func EmptyAddress() Address { return BytesToAddress([]byte{0}) }

// HexToAddress returns Address with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
func HexToAddress(s string) Hash160Address { return BytesToHash160Address(FromHex(s)) }

// Uint256ToAddress 大数字转换为地址
func Uint256ToAddress(b *uint256.Int) Address {
	var a Address
	raw := b.Bytes20()
	a.SetBytes(raw[:])
	return a
}

// HexToAddr 十六进制转换为虚拟机中的地址
func HexToAddr(s string) Address {
	var a Address
	out := make([]byte, 20)
	copy(out[:], FromHex(s))
	a.SetBytes(out)
	return a
}
