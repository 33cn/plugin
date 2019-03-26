// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"

	node "github.com/33cn/plugin/plugin/dapp/paracross/testnode"
	"github.com/stretchr/testify/assert"
)

func TestParaNode(t *testing.T) {
	authAcc := "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"

	para := node.NewParaNode(nil, nil)
	defer para.Close()
	//通过rpc 发生信息
	genesis := para.Main.GetGenesisAddress()
	genesisKey := para.Main.GetGenesisKey()
	block := para.Main.GetBlock(0)
	acc := para.Main.GetAccount(block.StateHash, genesis)
	assert.Equal(t, acc.Balance, 100000000*types.Coin)

	//super acc
	tx := util.CreateCoinsTx(genesisKey, para.Main.GetHotAddress(), 10*types.Coin)
	para.Main.SendTx(tx)
	para.Main.Wait()
	block = para.Main.GetLastBlock()
	acc = para.Main.GetAccount(block.StateHash, para.Main.GetHotAddress())
	assert.Equal(t, acc.Balance, 10*types.Coin)

	//auth acc
	tx = util.CreateCoinsTx(genesisKey, authAcc, 10*types.Coin)
	para.Main.SendTx(tx)
	para.Main.Wait()
	block = para.Main.GetLastBlock()
	acc = para.Main.GetAccount(block.StateHash, authAcc)
	assert.Equal(t, acc.Balance, 10*types.Coin)

	//create manage config
	tx = util.CreateManageTx(para.Main.GetHotKey(), "paracross-nodes-user.p.guodun.", "add", "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF")
	reply, err := para.Main.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	detail, err := para.Main.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
	assert.Equal(t, detail.Receipt.Ty, int32(types.ExecOk))

	for i := 0; i < 3; i++ {
		tx = util.CreateTxWithExecer(para.Para.GetGenesisKey(), "user.p.guodun.none")
		para.Para.SendTxRPC(tx)
		para.Para.WaitHeight(int64(i) + 1)
	}

}
