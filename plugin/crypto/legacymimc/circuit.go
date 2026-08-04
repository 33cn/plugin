// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package legacymimc

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
)

// CircuitMiMC 电路内 MiMC 实现，使用旧版 round constants（sha3.Sum256 推导）。
// 参考 gnark v0.9.0 std/hash/mimc，但允许注入自定义 params 以保持链上协议兼容。
type CircuitMiMC struct {
	params []big.Int         // slice containing constants for the encryption rounds
	h      frontend.Variable // current vector in the Miyaguchi–Preneel scheme
	data   []frontend.Variable
	api    frontend.API
}

// NewCircuitMiMC returns a CircuitMiMC instance with given seed's old constants
func NewCircuitMiMC(api frontend.API, seed string) (CircuitMiMC, error) {
	params := NewParams(seed)
	res := CircuitMiMC{}
	res.params = make([]big.Int, len(params))
	for i := range params {
		params[i].BigInt(&res.params[i])
	}
	res.h = 0
	res.api = api
	return res, nil
}

// Write adds more data to the running hash.
func (h *CircuitMiMC) Write(data ...frontend.Variable) {
	h.data = append(h.data, data...)
}

// Reset resets the Hash to its initial state.
func (h *CircuitMiMC) Reset() {
	h.data = nil
	h.h = 0
}

// Sum hash (in r1cs form) using Miyaguchi–Preneel:
// https://en.wikipedia.org/wiki/One-way_compression_function
// 与 gnark v0.5.2 电路实现保持一致：h = E(stream, h.h) + stream
func (h *CircuitMiMC) Sum() frontend.Variable {
	for _, stream := range h.data {
		h.h = h.encrypt(stream, h.h)
		h.h = h.api.Add(h.h, stream)
	}
	h.data = nil
	return h.h
}

// encrypt a mimc run expressed as r1cs
// message: m, key: k, 对应 v0.5.2 的 encrypt(message, key)
func (h *CircuitMiMC) encrypt(message, key frontend.Variable) frontend.Variable {
	x := message
	for i := 0; i < len(h.params); i++ {
		x = h.pow5(h.api, h.api.Add(x, key, h.params[i]))
	}
	return h.api.Add(x, key)
}

func (h *CircuitMiMC) pow5(api frontend.API, x frontend.Variable) frontend.Variable {
	x2 := api.Mul(x, x)
	x3 := api.Mul(x2, x2)
	x5 := api.Mul(x3, x)
	return x5
}
