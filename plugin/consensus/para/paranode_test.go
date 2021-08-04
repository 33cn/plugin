// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"

	_ "github.com/33cn/plugin/plugin/dapp/init" //dapp init
	node "github.com/33cn/plugin/plugin/dapp/paracross/testnode"
	"github.com/stretchr/testify/assert"
)

func TestParaNode(t *testing.T) {
	authAcc := "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"

	para := node.NewParaNode(nil, nil)
	defer para.Close()
	cfg := para.Para.GetClient().GetConfig()
	//通过rpc 发生信息
	genesis := para.Main.GetGenesisAddress()
	genesisKey := para.Main.GetGenesisKey()
	block := para.Main.GetBlock(0)
	acc := para.Main.GetAccount(block.StateHash, genesis)
	assert.Equal(t, acc.Balance, 100000000*types.DefaultCoinPrecision)

	//super acc
	tx := util.CreateCoinsTx(cfg, genesisKey, para.Main.GetHotAddress(), 10*types.DefaultCoinPrecision)
	para.Main.SendTx(tx)
	para.Main.Wait()
	block = para.Main.GetLastBlock()
	acc = para.Main.GetAccount(block.StateHash, para.Main.GetHotAddress())
	assert.Equal(t, acc.Balance, 10*types.DefaultCoinPrecision)

	//auth acc
	tx = util.CreateCoinsTx(cfg, genesisKey, authAcc, 10*types.DefaultCoinPrecision)
	para.Main.SendTx(tx)
	para.Main.Wait()
	block = para.Main.GetLastBlock()
	acc = para.Main.GetAccount(block.StateHash, authAcc)
	assert.Equal(t, acc.Balance, 10*types.DefaultCoinPrecision)

	//create manage config
	tx = util.CreateManageTx(cfg, para.Main.GetHotKey(), "paracross-nodes-user.p.guodun.", "add", "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF")
	reply, err := para.Main.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	detail, err := para.Main.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
	assert.Equal(t, detail.Receipt.Ty, int32(types.ExecOk))

	testParaQuery(para)

	for i := 0; i < 2; i++ {
		tx = util.CreateTxWithExecer(cfg, para.Para.GetGenesisKey(), "user.p.guodun.none")
		para.Para.SendTxRPC(tx)
		para.Para.WaitHeight(int64(i) + 1)
	}

}

func testParaQuery(para *node.ParaNode) {
	var acc types.Account
	acc.Addr = "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"
	para.Para.GetAPI().Notify(
		"consensus", types.EventConsensusQuery, &types.ChainExecutor{
			Driver:   "para",
			FuncName: "CreateNewAccount",
			Param:    types.Encode(&acc),
		})

	var walletsatus types.WalletStatus
	walletsatus.IsWalletLock = true
	para.Para.GetAPI().Notify(
		"consensus", types.EventConsensusQuery, &types.ChainExecutor{
			Driver:   "para",
			FuncName: "WalletStatus",
			Param:    types.Encode(&walletsatus),
		})

	walletsatus.IsWalletLock = false
	para.Para.GetAPI().Notify(
		"consensus", types.EventConsensusQuery, &types.ChainExecutor{
			Driver:   "para",
			FuncName: "WalletStatus",
			Param:    types.Encode(&walletsatus),
		})
}
