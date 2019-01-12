// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
)

const retryNum = 10

// GetMainHeightByTxHash get Block height
func (action *Action) GetMainHeightByTxHash(txHash []byte) (int64, error) {
	param := &types.ReqHash{Hash: txHash}
	txDetail, err := action.lottery.GetExecutorAPI().QueryTx(param)
	if err != nil {
		return -1, err
	}
	return txDetail.GetHeight(), nil
}
