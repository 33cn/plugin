// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mm

import (
	"math/big"
	"sync"
)

// 整数池允许的最大长度
const poolLimit = 256

// IntPool big.Int组成的内存池
type IntPool struct {
	pool *Stack
}

// NewIntPool 创建新的内存池
func NewIntPool() *IntPool {
	return &IntPool{pool: NewStack()}
}

// Get 取数据
func (p *IntPool) Get() *big.Int {
	if p.pool.Len() > 0 {
		return p.pool.Pop()
	}
	return new(big.Int)
}

// Put 存数据
func (p *IntPool) Put(is ...*big.Int) {
	if len(p.pool.Items) > poolLimit {
		return
	}

	for _, i := range is {
		p.pool.Push(i)
	}
}

// GetZero 返回一个零值的big.Int
func (p *IntPool) GetZero() *big.Int {
	if p.pool.Len() > 0 {
		return p.pool.Pop().SetUint64(0)
	}
	return new(big.Int)
}

// 默认容量
const poolDefaultCap = 25

// IntPoolPool 用于管理IntPool的Pool
type IntPoolPool struct {
	pools []*IntPool
	lock  sync.Mutex
}

// PoolOfIntPools 内存缓冲池
var PoolOfIntPools = &IntPoolPool{
	pools: make([]*IntPool, 0, poolDefaultCap),
}

// Get 返回一个可用的内存池
func (ipp *IntPoolPool) Get() *IntPool {
	ipp.lock.Lock()
	defer ipp.lock.Unlock()

	if len(PoolOfIntPools.pools) > 0 {
		ip := ipp.pools[len(ipp.pools)-1]
		ipp.pools = ipp.pools[:len(ipp.pools)-1]
		return ip
	}
	return NewIntPool()
}

// Put 放入一个初始化过的内存池
func (ipp *IntPoolPool) Put(ip *IntPool) {
	ipp.lock.Lock()
	defer ipp.lock.Unlock()

	if len(ipp.pools) < cap(ipp.pools) {
		ipp.pools = append(ipp.pools, ip)
	}
}
