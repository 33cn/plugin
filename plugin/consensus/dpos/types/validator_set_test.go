package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pubkey1 = "027848E7FA630B759DB406940B5506B666A344B1060794BBF314EB459D40881BB3"
	pubkey2 = "03F4AB6659E61E8512C9A24AC385CC1AC4D52B87D10ADBDF060086EA82BE62CDDE"
	pubkey3 = "03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"

	pubkey11 = "03541AB9887951C038273648545072E5B6A46A639BFF535F3957E8150CBE2A70D7"
	pubkey12 = "03F2A7AFFA090763C42B370C6F33CC3E9B6140228ABAF0591240F3B88E8792F890"
)

var (
	val1 *Validator
	val2 *Validator
	val3 *Validator

	val11 *Validator
	val12 *Validator
)

func init() {
	//为了使用VRF，需要使用SECP256K1体系的公私钥
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		panic("init ConsensusCrypto failed.")
	}

	ConsensusCrypto = cr

	pkbytes, _ := hex.DecodeString(pubkey1)

	pk1, _ := ConsensusCrypto.PubKeyFromBytes(pkbytes)

	pkbytes, _ = hex.DecodeString(pubkey2)
	pk2, _ := ConsensusCrypto.PubKeyFromBytes(pkbytes)

	pkbytes, _ = hex.DecodeString(pubkey3)
	pk3, _ := ConsensusCrypto.PubKeyFromBytes(pkbytes)

	val1 = NewValidator(pk1)
	val2 = NewValidator(pk2)
	val3 = NewValidator(pk3)

	pkbytes, _ = hex.DecodeString(pubkey11)
	pk11, _ := ConsensusCrypto.PubKeyFromBytes(pkbytes)
	val11 = NewValidator(pk11)

	pkbytes, _ = hex.DecodeString(pubkey12)
	pk12, _ := ConsensusCrypto.PubKeyFromBytes(pkbytes)
	val12 = NewValidator(pk12)

}
func TestValidator(t *testing.T) {
	cval1 := val1.Copy()
	assert.True(t, bytes.Equal(val1.PubKey, cval1.PubKey))
	assert.True(t, bytes.Equal(val1.Address, cval1.Address))

	assert.True(t, strings.HasPrefix(val2.String(), "Validator{"))
	assert.True(t, len(val3.Hash()) > 0)
}

func match(index int, val *Validator) bool {
	return bytes.Equal(val.Address, val1.Address)
}

func TestValidatorSet(t *testing.T) {
	var vals []*Validator

	vals = append(vals, val1)
	vals = append(vals, val2)
	vals = append(vals, val3)

	valset := NewValidatorSet(vals)

	//03f4ab6659e61e8512c9a24ac385cc1ac4d52b87d10adbdf060086ea82be62cdde
	//027848e7fa630b759db406940b5506b666a344b1060794bbf314eb459d40881bb3
	//03ef0e1d3112cf571743a3318125ede2e52a4eb904bcbaa4b1f75020c2846a7eb4

	for _, v := range valset.Validators {
		fmt.Println(hex.EncodeToString(v.PubKey))
	}

	assert.True(t, bytes.Equal(valset.Validators[0].PubKey, val2.PubKey))
	assert.True(t, bytes.Equal(valset.Validators[1].PubKey, val1.PubKey))
	assert.True(t, bytes.Equal(valset.Validators[2].PubKey, val3.PubKey))

	assert.True(t, valset.HasAddress(val1.Address))
	assert.True(t, valset.HasAddress(val2.Address))
	assert.True(t, valset.HasAddress(val3.Address))
	inx, val := valset.GetByAddress(val1.Address)
	assert.True(t, inx == 1 && bytes.Equal(val.PubKey, val1.PubKey))

	inx, val = valset.GetByAddress(val2.Address)
	assert.True(t, inx == 0 && bytes.Equal(val.PubKey, val2.PubKey))

	inx, val = valset.GetByAddress(val3.Address)
	assert.True(t, inx == 2 && bytes.Equal(val.PubKey, val3.PubKey))

	addr, val := valset.GetByIndex(1)
	assert.True(t, bytes.Equal(val.PubKey, val1.PubKey))
	assert.True(t, bytes.Equal(addr, val1.Address))

	assert.True(t, 3 == valset.Size())
	assert.True(t, 0 < len(valset.Hash()))
	assert.True(t, valset.Add(val1) == false)
	assert.True(t, valset.Size() == 3)

	assert.True(t, valset.Add(val11) == true)
	assert.True(t, valset.Size() == 4)

	assert.True(t, valset.Update(val11) == true)
	assert.True(t, valset.Size() == 4)

	assert.True(t, valset.Update(val12) == false)
	assert.True(t, valset.Size() == 4)

	val, flag := valset.Remove(val11.Address)

	assert.True(t, bytes.Equal(val.PubKey, val11.PubKey))
	assert.True(t, flag == true)

	val, flag = valset.Remove(val12.Address)
	assert.True(t, flag == false)
	require.Nil(t, val)

	assert.True(t, valset.HasAddress(val1.Address) == true)

	//fmt.Println(valset.String())
	//fmt.Println(valset.StringIndented("	"))

	valset.Iterate(match)
}

func TestValidatorsByAddress(t *testing.T) {
	var arr ValidatorsByAddress
	arr = append(arr, val1)
	arr = append(arr, val2)
	arr = append(arr, val3)
	assert.True(t, arr.Len() == 3)

	assert.True(t, arr.Less(0, 1) == false)
	assert.True(t, arr.Less(0, 2) == true)

	arr.Swap(0, 1)
	assert.True(t, bytes.Equal(arr[0].PubKey, val2.PubKey))

}

func TestValidatorSetException(t *testing.T) {
	var vals []*Validator
	valset := NewValidatorSet(vals)
	assert.True(t, len(valset.Validators) == 0)

}
