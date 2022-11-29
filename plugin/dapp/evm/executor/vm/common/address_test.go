// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/33cn/chain33/system/address/btc"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/address/eth"

	"github.com/33cn/chain33/common"

	"github.com/33cn/chain33/types"

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

func TestNewContractAddress(t *testing.T) {

	tx := types.Transaction{}
	hash := tx.Hash()
	hexHash := hex.EncodeToString(hash)
	addr := BytesToAddress([]byte{1})

	addr1 := NewContractAddress(addr, hash)
	assert.Equal(t, hexHash, hex.EncodeToString(hash))
	assert.Equal(t, "1DgSnASfaE2J4xcD9ghDs6zptXPq68SwAf", addr1.String())

	cfg := types.NewChain33Config(types.GetDefaultCfgstring())

	addr2 := NewAddress(cfg, hash)
	assert.Equal(t, hexHash, hex.EncodeToString(hash))
	assert.Equal(t, "17ZDZhQrFRnwQgBxZdLWvqh8dqfLXBRzyj", addr2.String())

	ethdriver, err := address.LoadDriver(eth.ID, -1)
	assert.Equal(t, nil, err)
	InitEvmAddressDriver(ethdriver)
	ab, _ := common.FromHex("0xd83b69C56834E85e023B1738E69BFA2F0dd52905")
	addr = BytesToAddress(ab)
	nonce := 0x13

	addr3 := NewEvmContractAddress(addr, uint64(nonce))
	t.Log("addr3", addr3)
	assert.Equal(t, "0xdcadbe74054cdb2733ba95875339afad7af9fdf4", addr3.String())
}

func TestStringToAddress(t *testing.T) {

	lock.Lock()
	defer lock.Unlock()
	driver := evmAddressDriver
	evmAddressDriver, _ = address.LoadDriver(btc.NormalAddressID, -1)
	// evm为比特币地址类型时
	btcAddr := "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
	ethAddr := "0xdcadbe74054CDB2733ba95875339afad7af9fdf4"
	a := StringToAddress(btcAddr)
	assert.NotNil(t, a)
	assert.Equal(t, btcAddr, a.String())

	a = StringToAddress(ethAddr)
	assert.NotNil(t, a)
	assert.Equal(t, "1M7qo9PADK9cJqB5sM2bzdpvn4gXCchTo7", a.String())
	a = StringToAddressLegacy(ethAddr)
	assert.NotNil(t, a)
	assert.Equal(t, "1M7qo9PADK9cJqB5sM2bzdpvn4gXCchTo7", a.String())

	// evm调整为 eth地址类型
	evmAddressDriver, _ = address.LoadDriver(eth.ID, -1)
	a = StringToAddress(btcAddr)
	assert.NotNil(t, a)
	assert.Equal(t, "0x245afbf176934ccdd7ca291a8dddaa13c8184822", a.String())

	a = StringToAddress(ethAddr)
	assert.NotNil(t, a)
	assert.Equal(t, strings.ToLower(ethAddr), a.String())
	evmAddressDriver = driver
}
