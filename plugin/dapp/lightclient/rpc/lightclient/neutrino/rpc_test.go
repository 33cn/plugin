// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package neutrino

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

// TestGetRgbxWithdrawAssetParamValidation tests parameter validation
func TestGetRgbxWithdrawAssetParamValidation(t *testing.T) {
	n := &neutrinoClient{}

	// Test with empty hash
	result, err := n.getRgbxWithdrawAsset([]byte{})
	assert.Error(t, err)
	assert.Equal(t, types.ErrInvalidParam, err)
	assert.Nil(t, result)
}

// TestGetRgbxPendingTxByHashParamValidation tests parameter validation
func TestGetRgbxPendingTxByHashParamValidation(t *testing.T) {
	n := &neutrinoClient{}

	// Test with empty hash
	result, err := n.getRgbxPendingTxByHash([]byte{})
	assert.Error(t, err)
	assert.Equal(t, types.ErrInvalidParam, err)
	assert.Nil(t, result)
}
