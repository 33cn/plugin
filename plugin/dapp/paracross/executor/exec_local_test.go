// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

func TestGetCrossAssetTxBitMap(t *testing.T) {
	var allTxs [][]byte
	for i := 0; i < 6; i++ {
		allTxs = append(allTxs, common.Sha256([]byte(fmt.Sprint(i))))
	}

	var crossTx [][]byte
	hit := []int{0, 1, 2, 3}
	for _, v := range hit {
		crossTx = append(crossTx, common.Sha256([]byte(fmt.Sprint(v))))
	}

	d0 := &types.ReceiptData{Ty: types.ExecOk}
	d1 := &types.ReceiptData{Ty: types.ExecPack}
	d2 := &types.ReceiptData{Ty: types.ExecOk}
	d3 := &types.ReceiptData{Ty: types.ExecErr}
	d4 := &types.ReceiptData{Ty: types.ExecOk}
	d5 := &types.ReceiptData{Ty: types.ExecOk}
	receipts := []*types.ReceiptData{d0, d1, d2, d3, d4, d5}

	rst := getCrossAssetTxBitMap(crossTx, allTxs, receipts)
	assert.Equal(t, "000110101", rst)

	//test nil
	rst = getCrossAssetTxBitMap(nil, allTxs, receipts)
	assert.Equal(t, "0001", rst)

	//only one tx
	cross2 := crossTx[:1]
	rst = getCrossAssetTxBitMap(cross2, allTxs, receipts)
	assert.Equal(t, "000111", rst)

	//all cross fail
	d0.Ty = types.ExecErr
	d2.Ty = types.ExecErr
	rst = getCrossAssetTxBitMap(crossTx, allTxs, receipts)
	assert.Equal(t, "000110000", rst)

	val := big.NewInt(0)
	fmt.Println("v", val.Text(2))
}
