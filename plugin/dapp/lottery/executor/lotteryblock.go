// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"context"
	"time"

	"github.com/33cn/chain33/types"
)

const retryNum = 10

// GetMainHeightByTxHash get Block height
func (action *Action) GetMainHeightByTxHash(txHash []byte) int64 {
	for i := 0; i < retryNum; i++ {
		req := &types.ReqHash{Hash: txHash}
		txDetail, err := action.grpcClient.QueryTransaction(context.Background(), req)
		if err != nil {
			time.Sleep(time.Second)
		} else {
			return txDetail.GetHeight()
		}
	}

	return -1
}

// GetMainBlockHashByHeight get Hash
func (action *Action) GetMainBlockHashByHeight(height int64) ([]byte, error) {
	for i := 0; i < retryNum; i++ {
		req := &types.ReqInt{Height: height}
		replyHash, err := action.grpcClient.GetBlockHash(context.Background(), req)
		if err != nil {
			time.Sleep(time.Second)
		} else {
			return replyHash.Hash, nil
		}
	}

	return nil, types.ErrBlockNotFound
}
