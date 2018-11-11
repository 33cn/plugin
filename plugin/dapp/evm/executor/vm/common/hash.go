// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"math/big"

	"github.com/33cn/chain33/common"
)

const (
	HashLength = 32
)

type Hash common.Hash

func (h Hash) Str() string   { return string(h[:]) }
func (h Hash) Bytes() []byte { return h[:] }
func (h Hash) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }
func (h Hash) Hex() string   { return Bytes2Hex(h[:]) }

// 设置哈希中的字节值，如果字节数组长度超过哈希长度，则被截断，只保留后面的部分
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

func BigToHash(b *big.Int) Hash {
	return Hash(common.BigToHash(b))
}

// 将[]byte直接当做哈希处理
func BytesToHash(b []byte) Hash {
	return Hash(common.BytesToHash(b))
}

func EmptyHash(h Hash) bool {
	return h == Hash{}
}

// 将[]byte经过哈希计算后转化为哈希对象
func ToHash(data []byte) Hash {
	return BytesToHash(common.Sha256(data))
}
