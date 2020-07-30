// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package bls12381

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
)

// fe is base field element representation
type fe [6]uint64

// fe2 is element representation of 'fp2' which is quadratic extension of base field 'fp'
// Representation follows c[0] + c[1] * u encoding order.
type fe2 [2]fe

// fe6 is element representation of 'fp6' field which is cubic extension of 'fp2'
// Representation follows c[0] + c[1] * v + c[2] * v^2 encoding order.
type fe6 [3]fe2

// fe12 is element representation of 'fp12' field which is quadratic extension of 'fp6'
// Representation follows c[0] + c[1] * w encoding order.
type fe12 [2]fe6

func (fe1 *fe) setBytes(in []byte) *fe {
	size := 48
	l := len(in)
	if l >= size {
		l = size
	}
	padded := make([]byte, size)
	copy(padded[size-l:], in[:])
	var a int
	for i := 0; i < 6; i++ {
		a = size - i*8
		fe1[i] = uint64(padded[a-1]) | uint64(padded[a-2])<<8 |
			uint64(padded[a-3])<<16 | uint64(padded[a-4])<<24 |
			uint64(padded[a-5])<<32 | uint64(padded[a-6])<<40 |
			uint64(padded[a-7])<<48 | uint64(padded[a-8])<<56
	}
	return fe1
}

func (fe1 *fe) setBig(a *big.Int) *fe {
	return fe1.setBytes(a.Bytes())
}

func (fe1 *fe) setString(s string) (*fe, error) {
	if s[:2] == "0x" {
		s = s[2:]
	}
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return fe1.setBytes(bytes), nil
}

func (fe1 *fe) set(fe2 *fe) *fe {
	fe1[0] = fe2[0]
	fe1[1] = fe2[1]
	fe1[2] = fe2[2]
	fe1[3] = fe2[3]
	fe1[4] = fe2[4]
	fe1[5] = fe2[5]
	return fe1
}

func (fe1 *fe) bytes() []byte {
	out := make([]byte, 48)
	var a int
	for i := 0; i < 6; i++ {
		a = 48 - i*8
		out[a-1] = byte(fe1[i])
		out[a-2] = byte(fe1[i] >> 8)
		out[a-3] = byte(fe1[i] >> 16)
		out[a-4] = byte(fe1[i] >> 24)
		out[a-5] = byte(fe1[i] >> 32)
		out[a-6] = byte(fe1[i] >> 40)
		out[a-7] = byte(fe1[i] >> 48)
		out[a-8] = byte(fe1[i] >> 56)
	}
	return out
}

func (fe1 *fe) big() *big.Int {
	return new(big.Int).SetBytes(fe1.bytes())
}

func (fe1 *fe) string() (s string) {
	for i := 5; i >= 0; i-- {
		s = fmt.Sprintf("%s%16.16x", s, fe1[i])
	}
	return "0x" + s
}

func (fe1 *fe) zero() *fe {
	fe1[0] = 0
	fe1[1] = 0
	fe1[2] = 0
	fe1[3] = 0
	fe1[4] = 0
	fe1[5] = 0
	return fe1
}

func (fe1 *fe) one() *fe {
	return fe1.set(r1)
}

func (fe1 *fe) rand(r io.Reader) (*fe, error) {
	bi, err := rand.Int(r, modulus.big())
	if err != nil {
		return nil, err
	}
	return fe1.setBig(bi), nil
}

func (fe1 *fe) isValid() bool {
	return fe1.cmp(&modulus) < 0
}

func (fe1 *fe) isOdd() bool {
	var mask uint64 = 1
	return fe1[0]&mask != 0
}

func (fe1 *fe) isEven() bool {
	var mask uint64 = 1
	return fe1[0]&mask == 0
}

func (fe1 *fe) isZero() bool {
	return (fe1[5] | fe1[4] | fe1[3] | fe1[2] | fe1[1] | fe1[0]) == 0
}

func (fe1 *fe) isOne() bool {
	return fe1.equal(r1)
}

func (fe1 *fe) cmp(fe2 *fe) int {
	for i := 5; i >= 0; i-- {
		if fe1[i] > fe2[i] {
			return 1
		} else if fe1[i] < fe2[i] {
			return -1
		}
	}
	return 0
}

func (fe1 *fe) equal(fe2 *fe) bool {
	return fe2[0] == fe1[0] && fe2[1] == fe1[1] && fe2[2] == fe1[2] && fe2[3] == fe1[3] && fe2[4] == fe1[4] && fe2[5] == fe1[5]
}

func (fe1 *fe) sign() bool {
	r := new(fe)
	fromMont(r, fe1)
	return r[0]&1 == 0
}

func (fe1 *fe) div2(e uint64) {
	fe1[0] = fe1[0]>>1 | fe1[1]<<63
	fe1[1] = fe1[1]>>1 | fe1[2]<<63
	fe1[2] = fe1[2]>>1 | fe1[3]<<63
	fe1[3] = fe1[3]>>1 | fe1[4]<<63
	fe1[4] = fe1[4]>>1 | fe1[5]<<63
	fe1[5] = fe1[5]>>1 | e<<63
}

func (fe1 *fe) mul2() uint64 {
	e := fe1[5] >> 63
	fe1[5] = fe1[5]<<1 | fe1[4]>>63
	fe1[4] = fe1[4]<<1 | fe1[3]>>63
	fe1[3] = fe1[3]<<1 | fe1[2]>>63
	fe1[2] = fe1[2]<<1 | fe1[1]>>63
	fe1[1] = fe1[1]<<1 | fe1[0]>>63
	fe1[0] = fe1[0] << 1
	return e
}

func (e *fe2) zero() *fe2 {
	e[0].zero()
	e[1].zero()
	return e
}

func (e *fe2) one() *fe2 {
	e[0].one()
	e[1].zero()
	return e
}

func (e *fe2) set(e2 *fe2) *fe2 {
	e[0].set(&e2[0])
	e[1].set(&e2[1])
	return e
}

func (e *fe2) rand(r io.Reader) (*fe2, error) {
	a0, err := new(fe).rand(r)
	if err != nil {
		return nil, err
	}
	a1, err := new(fe).rand(r)
	if err != nil {
		return nil, err
	}
	return &fe2{*a0, *a1}, nil
}

func (e *fe2) isOne() bool {
	return e[0].isOne() && e[1].isZero()
}

func (e *fe2) isZero() bool {
	return e[0].isZero() && e[1].isZero()
}

func (e *fe2) equal(e2 *fe2) bool {
	return e[0].equal(&e2[0]) && e[1].equal(&e2[1])
}

func (e *fe2) sign() bool {
	r := new(fe)
	if !e[0].isZero() {
		fromMont(r, &e[0])
		return r[0]&1 == 0
	}
	fromMont(r, &e[1])
	return r[0]&1 == 0
}

func (e *fe6) zero() *fe6 {
	e[0].zero()
	e[1].zero()
	e[2].zero()
	return e
}

func (e *fe6) one() *fe6 {
	e[0].one()
	e[1].zero()
	e[2].zero()
	return e
}

func (e *fe6) set(e2 *fe6) *fe6 {
	e[0].set(&e2[0])
	e[1].set(&e2[1])
	e[2].set(&e2[2])
	return e
}

func (e *fe6) rand(r io.Reader) (*fe6, error) {
	a0, err := new(fe2).rand(r)
	if err != nil {
		return nil, err
	}
	a1, err := new(fe2).rand(r)
	if err != nil {
		return nil, err
	}
	a2, err := new(fe2).rand(r)
	if err != nil {
		return nil, err
	}
	return &fe6{*a0, *a1, *a2}, nil
}

func (e *fe6) isOne() bool {
	return e[0].isOne() && e[1].isZero() && e[2].isZero()
}

func (e *fe6) isZero() bool {
	return e[0].isZero() && e[1].isZero() && e[2].isZero()
}

func (e *fe6) equal(e2 *fe6) bool {
	return e[0].equal(&e2[0]) && e[1].equal(&e2[1]) && e[2].equal(&e2[2])
}

func (e *fe12) zero() *fe12 {
	e[0].zero()
	e[1].zero()
	return e
}

func (e *fe12) one() *fe12 {
	e[0].one()
	e[1].zero()
	return e
}

func (e *fe12) set(e2 *fe12) *fe12 {
	e[0].set(&e2[0])
	e[1].set(&e2[1])
	return e
}

func (e *fe12) rand(r io.Reader) (*fe12, error) {
	a0, err := new(fe6).rand(r)
	if err != nil {
		return nil, err
	}
	a1, err := new(fe6).rand(r)
	if err != nil {
		return nil, err
	}
	return &fe12{*a0, *a1}, nil
}

func (e *fe12) isOne() bool {
	return e[0].isOne() && e[1].isZero()
}

func (e *fe12) isZero() bool {
	return e[0].isZero() && e[1].isZero()
}

func (e *fe12) equal(e2 *fe12) bool {
	return e[0].equal(&e2[0]) && e[1].equal(&e2[1])
}
