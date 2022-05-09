// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"testing"

	"github.com/holiman/uint256"

	"github.com/stretchr/testify/assert"
)

func TestAddressBig(t *testing.T) {
	saddr := "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
	addr := StringToAddress(saddr)
	baddr := addr.Big()
	naddr := BigToAddress(baddr)
	if saddr != naddr.String() {
		t.Fail()
	}
}

func TestAddressBytes(t *testing.T) {
	addr := BytesToAddress([]byte{1})
	assert.Equal(t, addr.String(), "11111111111111111111BZbvjr")
}

func TestEvmPrecompileAddress(t *testing.T) {

	b := make([]byte, 1)
	var z uint256.Int
	for i := 0; i < 200; i++ {
		b[0] = byte(i)
		addr1 := BytesToHash160Address(b).String()
		z.SetBytes(b)
		addr2 := Uint256ToAddress(&z).ToHash160().String()
		assert.Equal(t, addr1, addr2)
	}
}
