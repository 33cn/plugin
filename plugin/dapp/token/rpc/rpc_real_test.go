// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"testing"

	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	_ "github.com/33cn/plugin/plugin"
	tokenty "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/stretchr/testify/assert"
)

func TestRPCTokenPreCreate(t *testing.T) {
	// 启动RPCmocker
	mock33 := testnode.New("", nil)
	defer mock33.Close()
	mock33.Listen()
	//precreate
	err := mock33.SendHot()
	assert.Nil(t, err)
	block := mock33.GetLastBlock()
	acc := mock33.GetAccount(block.StateHash, mock33.GetGenesisAddress())
	assert.Equal(t, acc.Balance, int64(9998999999900000))
	acc = mock33.GetAccount(block.StateHash, mock33.GetHotAddress())
	assert.Equal(t, acc.Balance, 10000*types.Coin)

	tx := util.CreateManageTx(mock33.GetHotKey(), "token-blacklist", "add", "BTY")
	reply, err := mock33.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	detail, err := mock33.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
	assert.Equal(t, detail.Receipt.Ty, int32(types.ExecOk))
	//开始发行percreate
	param := tokenty.TokenPreCreate{
		Name:   "Test",
		Symbol: "TEST",
		Total:  10000 * types.Coin,
		Owner:  mock33.GetHotAddress(),
	}
	var txhex string
	err = mock33.GetJSONC().Call("token.CreateRawTokenPreCreateTx", param, &txhex)
	assert.Nil(t, err)
	hash, err := mock33.SendAndSign(mock33.GetHotKey(), txhex)
	assert.Nil(t, err)
	assert.NotNil(t, hash)
	detail, err = mock33.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, detail.Receipt.Ty, int32(types.ExecOk))
}
