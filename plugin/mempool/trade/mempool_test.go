// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trade

import (
	"encoding/json"
	"testing"

	"github.com/33cn/chain33/system/mempool"
	"github.com/33cn/chain33/types"
)

func TestNewMempool(t *testing.T) {
	sub, _ := json.Marshal(&subConfig{
		PoolCacheSize:      2,
		MinTxFee:           100000,
		MaxTxNumPerAccount: 10000,
		TimeParam:          1,
		PriceConstant:      1,
		PricePower:         1,
	})
	module := New(&types.Mempool{}, sub)
	mem := module.(*mempool.Mempool)
	mem.Close()
}
