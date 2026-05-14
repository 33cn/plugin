// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package neutrino

import (
	"testing"

	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEncodeDecodeOutPoint tests the OutPoint encoding/decoding functions
func TestEncodeDecodeOutPoint(t *testing.T) {
	tests := []struct {
		name string
		op   wire.OutPoint
	}{
		{
			name: "normal outpoint",
			op: wire.OutPoint{
				Hash:  [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				Index: 1,
			},
		},
		{
			name: "zero hash",
			op: wire.OutPoint{
				Hash:  [32]byte{},
				Index: 0,
			},
		},
		{
			name: "max index",
			op: wire.OutPoint{
				Hash:  [32]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				Index: 0xffffffff,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded := encodeOutPoint(&tt.op)
			require.NotNil(t, encoded)
			assert.Equal(t, tt.op.String(), string(encoded))

			// Decode
			decoded, err := decodeOutPoint(encoded)
			require.NoError(t, err)
			assert.Equal(t, tt.op.Hash, decoded.Hash)
			assert.Equal(t, tt.op.Index, decoded.Index)
		})
	}
}

// TestDecodeOutPointError tests decodeOutPoint with invalid input
func TestDecodeOutPointError(t *testing.T) {
	invalidInputs := []string{
		"invalid",
		"",
		"too:many:colons",
		"badhash:999999999999999999999",
	}

	for _, input := range invalidInputs {
		_, err := decodeOutPoint([]byte(input))
		assert.Error(t, err, "should error for input: %s", input)
	}
}

// TestPending2WithdrawRequest tests the conversion function
func TestPending2WithdrawRequest(t *testing.T) {
	chain33Hash := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	amount := int64(1000000)
	feeRate := int64(10)
	toAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

	// Create a mock pending tx (simplified)
	pending := &rtypes.PendingTx{
		TxHash:        chain33Hash,
		Amount:        amount,
		FeeRate:       feeRate,
		TargetAddress: toAddress,
		ActionType:    rtypes.TyWithDrawAsset,
	}

	req := pending2WithdrawRequest(pending)

	assert.Equal(t, chain33Hash, req.chain33WithDrawHash)
	assert.Equal(t, btcutil.Amount(amount), req.amount)
	assert.Equal(t, btcutil.Amount(feeRate), req.feeRate)
	assert.Equal(t, toAddress, req.toAddress)
	assert.Nil(t, req.stickyUTXO)
}
