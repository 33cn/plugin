// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"math"
	"math/big"
)

// 常用的大整数常量定义
var (
	// Big0 大数字0
	Big0 = big.NewInt(0)
	// Big1 大数字1
	Big1 = big.NewInt(1)
	// Big32 大数字32
	Big32 = big.NewInt(32)
	// Big256 大数字256
	Big256 = big.NewInt(256)
	// Big257 大数字257
	Big257 = big.NewInt(257)
)

// 2的各种常用取幂结果
var (
	// TT255 2的255次幂
	TT255   = BigPow(2, 255)
	tt256   = BigPow(2, 256)
	tt256m1 = new(big.Int).Sub(tt256, big.NewInt(1))
)

const (
	// WordBits 一个big.Word类型取值占用多少个位
	WordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// WordBytes 一个big.Word类型取值占用多少个字节
	WordBytes = WordBits / 8
)

// BigMax 返回两者之中的较大值
func BigMax(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 {
		return y
	}
	return x
}

// BigMin 返回两者之中的较小值
func BigMin(x, y *big.Int) *big.Int {
	if x.Cmp(y) > 0 {
		return y
	}
	return x
}

// BigPow 返回a的b次幂
func BigPow(a, b int64) *big.Int {
	r := big.NewInt(a)
	return r.Exp(r, big.NewInt(b), nil)
}

// U256 求补
func U256(x *big.Int) *big.Int {
	return x.And(x, tt256m1)
}

// S256 interprets x as a two's complement number.
// x must not exceed 256 bits (the result is undefined if it does) and is not modified.
//   S256(0)        = 0
//   S256(1)        = 1
//   S256(2**255)   = -2**255
//   S256(2**256-1) = -1
func S256(x *big.Int) *big.Int {
	if x.Cmp(TT255) < 0 {
		return x
	}
	return new(big.Int).Sub(x, tt256)
}

// Exp 指数函数，可以指定底数，结果被截断为256位长度
func Exp(base, exponent *big.Int) *big.Int {
	result := big.NewInt(1)

	for _, word := range exponent.Bits() {
		for i := 0; i < WordBits; i++ {
			if word&1 == 1 {
				U256(result.Mul(result, base))
			}
			U256(base.Mul(base, base))
			word >>= 1
		}
	}
	return result
}

// Byte big.Int以小端编码时，第n个位置的字节取值
// 例如: bigint '5', padlength 32, n=31 => 5
func Byte(bigint *big.Int, padlength, n int) byte {
	if n >= padlength {
		return byte(0)
	}
	return bigEndianByteAt(bigint, padlength-1-n)
}

// 将big.Int以大端方式编码，返回第n个位置的字节取值
func bigEndianByteAt(bigint *big.Int, n int) byte {
	words := bigint.Bits()

	// 确保n不会越界
	i := n / WordBytes
	if i >= len(words) {
		return byte(0)
	}

	// 先按字的长度获取
	word := words[i]

	// 要获取的字节在当前字中的偏移量
	shift := 8 * uint(n%WordBytes)

	return byte(word >> shift)
}

// ReadBits 以大端方式将big.Int编码为字节数组
func ReadBits(bigint *big.Int, buf []byte) {
	i := len(buf)
	for _, d := range bigint.Bits() {
		for j := 0; j < WordBytes && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
}

// SafeAdd 加法运算，返回是否溢出
func SafeAdd(x, y uint64) (uint64, bool) {
	return x + y, y > math.MaxUint64-x
}

// SafeMul 乘法运算，返回是否溢出
func SafeMul(x, y uint64) (uint64, bool) {
	if x == 0 || y == 0 {
		return 0, false
	}
	return x * y, y > math.MaxUint64/x
}

// Zero 检查数字是否为0
func Zero(value *big.Int) bool {
	if value == nil || value.Sign() == 0 {
		return true
	}
	return false
}
