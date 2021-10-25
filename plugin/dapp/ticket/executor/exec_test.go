// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor_test

import (
	"testing"

	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	executor "github.com/33cn/plugin/plugin/dapp/ticket/executor"
	pty "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

type execEnv struct {
	blockTime   int64 // 1539918074
	blockHeight int64
	index       int
	difficulty  uint64

	txHash string
}

var (
	Symbol         = "TEST"
	SymbolA        = "TESTA"
	AssetExecToken = "token"
	AssetExecPara  = "paracross"

	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	Nodes    = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}
	//chain33TestCfg = types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
)

func Test_Exec_Bind_Unbind(t *testing.T) {
	chain33TestCfg := mock33.GetAPI().GetConfig()
	env := execEnv{
		1539918074,
		10000,
		2,
		1539918074,
		"hash",
	}

	_, ldb, kvdb := util.CreateTestDB()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(chain33TestCfg, nil)
	driver, err := dapp.LoadDriver("ticket", 1000)
	assert.Nil(t, err)
	driver.SetAPI(api)
	driver.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	driver.SetStateDB(kvdb)
	driver.SetLocalDB(kvdb)

	priv, err := FromPrivkey(PrivKeyA)
	assert.Nil(t, err)

	bindTx := createBindMiner(t, chain33TestCfg, string(Nodes[1]), string(Nodes[0]), priv)
	receipt, err := driver.Exec(bindTx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}
	assert.Equal(t, 1, len(receipt.KV))
	assert.Equal(t, executor.BindKey(string(Nodes[0])), receipt.KV[0].Key)
	var bindInfo pty.TicketBind
	err = types.Decode(receipt.KV[0].Value, &bindInfo)
	assert.Nil(t, err)
	assert.Equal(t, string(Nodes[1]), bindInfo.MinerAddress)
	assert.Equal(t, string(Nodes[0]), bindInfo.ReturnAddress)

	unbindTx := createBindMiner(t, chain33TestCfg, "", string(Nodes[0]), priv)
	receipt, err = driver.Exec(unbindTx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}
	assert.Equal(t, 1, len(receipt.KV))
	assert.Equal(t, executor.BindKey(string(Nodes[0])), receipt.KV[0].Key)
	var bindInfo2 pty.TicketBind
	err = types.Decode(receipt.KV[0].Value, &bindInfo2)
	assert.Nil(t, err)
	assert.Equal(t, "", bindInfo2.MinerAddress)
	assert.Equal(t, string(Nodes[0]), bindInfo2.ReturnAddress)

	ldb.Close()
}

func FromPrivkey(hexPrivKey string) (crypto.PrivKey, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName("ticket", signType), -1)
	if err != nil {
		return nil, err
	}

	bytes, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		return nil, err
	}

	return c.PrivKeyFromBytes(bytes)

}
