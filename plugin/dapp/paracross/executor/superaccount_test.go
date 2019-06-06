// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	//"github.com/stretchr/testify/mock"
	"testing"

	apimock "github.com/33cn/chain33/client/mocks"
	dbm "github.com/33cn/chain33/common/db"
	dbmock "github.com/33cn/chain33/common/db/mocks"
	"github.com/33cn/chain33/types"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

var (
	PrivKey14K = "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" // 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
	Account14K = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
	Account1MC = "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
	applyAddrs = "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4, 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR, 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k,1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"

	Account12Q = "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
	PrivKey12Q = "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
)

type NodeManageTestSuite struct {
	suite.Suite
	stateDB dbm.KV
	localDB *dbmock.KVDB
	api     *apimock.QueueProtocolAPI

	exec *Paracross
	//title string
}

func (suite *NodeManageTestSuite) SetupSuite() {

	suite.stateDB, _ = dbm.NewGoMemDB("state", "state", 1024)
	// memdb 不支持KVDB接口， 等测试完Exec ， 再扩展 memdb
	//suite.localDB, _ = dbm.NewGoMemDB("local", "local", 1024)
	suite.localDB = new(dbmock.KVDB)
	suite.api = new(apimock.QueueProtocolAPI)

	suite.exec = newParacross().(*Paracross)
	suite.exec.SetLocalDB(suite.localDB)
	suite.exec.SetStateDB(suite.stateDB)
	suite.exec.SetEnv(0, 0, 0)
	suite.exec.SetBlockInfo([]byte(""), []byte(""), 3)
	suite.exec.SetAPI(suite.api)
	enableParacrossTransfer = false

	//forkHeight := types.GetDappFork(pt.ParaX, pt.ForkCommitTx)
	//if forkHeight == types.MaxHeight {
	//	types.ReplaceDappFork(MainTitle, pt.ParaX, pt.ForkCommitTx, 0)
	//
	//}

	types.S("config.consensus.sub.para.MainForkParacrossCommitTx", int64(1))
	types.S("config.exec.sub.manage.superManager", []interface{}{Account12Q})

	// TODO, more fields
	// setup block
	blockDetail := &types.BlockDetail{
		Block: &types.Block{},
	}
	MainBlockHash10 = blockDetail.Block.Hash()

}

func (suite *NodeManageTestSuite) TestSetup() {
	nodeConfigKey := calcParaNodeGroupAddrsKey(Title)
	suite.T().Log(string(nodeConfigKey))
	_, err := suite.stateDB.Get(nodeConfigKey)
	if err != nil {
		suite.T().Error("get setup title failed", err)
		return
	}
}

func nodeCommit(suite *NodeManageTestSuite, privkeyStr string, tx *types.Transaction) (receipt *types.Receipt) {
	return nodeCommitImpl(suite.Suite, suite.exec, privkeyStr, tx)
}

func nodeCommitImpl(suite suite.Suite, exec *Paracross, privkeyStr string, tx *types.Transaction) (receipt *types.Receipt) {
	tx, _ = signTx(suite, tx, privkeyStr)

	suite.T().Log(tx.From())
	receipt, err := exec.Exec(tx, 0)
	suite.T().Log(receipt)
	assert.NotNil(suite.T(), receipt)
	assert.Nil(suite.T(), err)

	for _, v := range receipt.KV {
		if err := exec.GetStateDB().Set(v.Key, v.Value); err != nil {
			panic(err)
		}
	}
	return
}

func checkGroupApplyReceipt(suite *NodeManageTestSuite, receipt *types.Receipt) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 1)
	assert.Len(suite.T(), receipt.Logs, 1)

	assert.Equal(suite.T(), int32(pt.TyLogParaNodeGroupConfig), receipt.Logs[0].Ty)

}

func checkGroupApproveReceipt(suite *NodeManageTestSuite, receipt *types.Receipt) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
}

func checkJoinReceipt(suite *NodeManageTestSuite, receipt *types.Receipt) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 1)
	assert.Len(suite.T(), receipt.Logs, 1)

	var stat pt.ParaNodeIdStatus
	err := types.Decode(receipt.KV[0].Value, &stat)
	assert.Nil(suite.T(), err, "decode ParaNodeAddrStatus failed")
	//suite.T().Log("titleHeight", titleHeight)
	assert.Equal(suite.T(), int32(pt.TyLogParaNodeConfig), receipt.Logs[0].Ty)
	assert.Equal(suite.T(), int32(pt.ParacrossNodeJoining), stat.Status)
	assert.NotNil(suite.T(), stat.Votes)

}

func checkQuitReceipt(suite *NodeManageTestSuite, receipt *types.Receipt) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 1)
	assert.Len(suite.T(), receipt.Logs, 1)

	var stat pt.ParaNodeIdStatus
	err := types.Decode(receipt.KV[0].Value, &stat)
	assert.Nil(suite.T(), err, "decode ParaNodeAddrStatus failed")
	//suite.T().Log("titleHeight", titleHeight)
	assert.Equal(suite.T(), int32(pt.TyLogParaNodeConfig), receipt.Logs[0].Ty)
	assert.Equal(suite.T(), int32(pt.ParacrossNodeQuiting), stat.Status)
	assert.NotNil(suite.T(), stat.Votes)

}

func checkVoteReceipt(suite *NodeManageTestSuite, receipt *types.Receipt, count int) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))

	var stat pt.ParaNodeIdStatus
	err := types.Decode(receipt.KV[0].Value, &stat)
	assert.Nil(suite.T(), err, "decode ParaNodeAddrStatus failed")
	assert.Len(suite.T(), stat.Votes.Votes, count)

}

func checkVoteDoneReceipt(suite *NodeManageTestSuite, receipt *types.Receipt, count int, join bool) {
	suite.NotNil(receipt)
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))

	suite.T().Log("checkVoteDoneReceipt", "kvlen", len(receipt.KV))

	_, arry, err := getParacrossNodes(suite.stateDB, Title)
	suite.Suite.Nil(err)
	if join {
		suite.Contains(arry, Account14K)
	} else {
		suite.NotContains(arry, Account14K)
	}
}

func voteTest(suite *NodeManageTestSuite, id string, join bool) {
	var count int
	config := &pt.ParaNodeAddrConfig{
		Op:    pt.ParaNodeVote,
		Id:    id,
		Value: pt.ParaNodeVoteYes,
	}
	tx, err := pt.CreateRawNodeConfigTx(config)
	suite.Nil(err)

	count++
	receipt := nodeCommit(suite, PrivKeyA, tx)
	checkVoteReceipt(suite, receipt, count)
	count++

	receipt = nodeCommit(suite, PrivKeyB, tx)
	checkVoteReceipt(suite, receipt, count)
	count++

	if !join {
		receipt = nodeCommit(suite, PrivKey14K, tx)
		checkVoteReceipt(suite, receipt, count)
		count++
	}

	receipt = nodeCommit(suite, PrivKeyC, tx)
	checkVoteDoneReceipt(suite, receipt, count, join)
}

func (suite *NodeManageTestSuite) testNodeGroupConfigQuit() {
	config := &pt.ParaNodeGroupConfig{
		Addrs: applyAddrs,
		Op:    pt.ParacrossNodeGroupApply,
	}
	tx, err := pt.CreateRawNodeGroupApplyTx(config)
	suite.Nil(err)

	receipt := nodeCommit(suite, PrivKeyB, tx)
	checkGroupApplyReceipt(suite, receipt)

	suite.Equal(int32(pt.TyLogParaNodeGroupConfig), receipt.Logs[0].Ty)
	var g pt.ReceiptParaNodeGroupConfig
	err = types.Decode(receipt.Logs[0].Log, &g)
	suite.Nil(err)

	config = &pt.ParaNodeGroupConfig{
		Id: g.Current.Id,
		Op: pt.ParacrossNodeGroupQuit,
	}
	tx, err = pt.CreateRawNodeGroupApplyTx(config)
	suite.Nil(err)

	nodeCommit(suite, PrivKeyB, tx)
	//checkGroupApproveReceipt(suite, receipt)

}

func (suite *NodeManageTestSuite) testNodeGroupConfig() {
	suite.testNodeGroupConfigQuit()

	config := &pt.ParaNodeGroupConfig{
		Addrs: applyAddrs,
		Op:    pt.ParacrossNodeGroupApply,
	}
	tx, err := pt.CreateRawNodeGroupApplyTx(config)
	suite.Nil(err)

	receipt := nodeCommit(suite, PrivKeyB, tx)
	checkGroupApplyReceipt(suite, receipt)

	suite.Equal(int32(pt.TyLogParaNodeGroupConfig), receipt.Logs[0].Ty)
	var g pt.ReceiptParaNodeGroupConfig
	err = types.Decode(receipt.Logs[0].Log, &g)
	suite.Nil(err)

	config = &pt.ParaNodeGroupConfig{
		Id: g.Current.Id,
		Op: pt.ParacrossNodeGroupApprove,
	}
	tx, err = pt.CreateRawNodeGroupApplyTx(config)
	suite.Nil(err)

	receipt = nodeCommit(suite, PrivKey12Q, tx)
	checkGroupApproveReceipt(suite, receipt)

}

func (suite *NodeManageTestSuite) testNodeConfig() {
	//Join test
	config := &pt.ParaNodeAddrConfig{
		Op:   pt.ParaNodeJoin,
		Addr: Account14K,
	}
	tx, err := pt.CreateRawNodeConfigTx(config)
	suite.Nil(err)

	receipt := nodeCommit(suite, PrivKey14K, tx)
	checkJoinReceipt(suite, receipt)

	suite.Equal(int32(pt.TyLogParaNodeConfig), receipt.Logs[0].Ty)
	var g pt.ReceiptParaNodeConfig
	err = types.Decode(receipt.Logs[0].Log, &g)
	suite.Nil(err)

	//vote test
	voteTest(suite, g.Current.Id, true)

	//Quit test
	config = &pt.ParaNodeAddrConfig{
		Op:   pt.ParaNodeQuit,
		Addr: Account14K,
	}
	tx, err = pt.CreateRawNodeConfigTx(config)
	suite.Nil(err)
	receipt = nodeCommit(suite, PrivKeyD, tx)
	checkQuitReceipt(suite, receipt)

	suite.Equal(int32(pt.TyLogParaNodeConfig), receipt.Logs[0].Ty)
	err = types.Decode(receipt.Logs[0].Log, &g)
	suite.Nil(err)

	//vote test
	voteTest(suite, g.Current.Id, false)
}

func (suite *NodeManageTestSuite) TestExec() {
	suite.testNodeGroupConfig()
	suite.testNodeConfig()

}

func TestNodeManageSuite(t *testing.T) {
	tempTitle = types.GetTitle()
	types.SetTitleOnlyForTest(Title)

	suite.Run(t, new(NodeManageTestSuite))

	types.SetTitleOnlyForTest(tempTitle)
}

func (suite *NodeManageTestSuite) TearDownSuite() {

}

func TestGetAddrGroup(t *testing.T) {
	addrs := " 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4,    1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR, 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k, ,,,  1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs ,   "

	retAddrs := getConfigAddrs(addrs)
	expectAddrs := []string{"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4", "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR", "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k", "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"}
	assert.Equal(t, expectAddrs, retAddrs)

	addrs = " 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 , ,   "
	retAddrs = getConfigAddrs(addrs)
	expectAddrs = []string{"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"}
	assert.Equal(t, expectAddrs, retAddrs)

	addrs = " , "
	ret := getConfigAddrs(addrs)
	assert.Equal(t, []string(nil), ret)
	assert.Equal(t, 0, len(ret))

	addrs = " "
	ret = getConfigAddrs(addrs)
	assert.Equal(t, []string(nil), ret)
	assert.Equal(t, 0, len(ret))

}

func TestUpdateVotes(t *testing.T) {
	stat := &pt.ParaNodeIdStatus{}
	votes := &pt.ParaNodeVoteDetail{
		Addrs: []string{"AA", "BB", "CC"},
		Votes: []string{"yes", "no", "no"}}
	stat.Votes = votes
	nodes := make(map[string]struct{})
	nodes["BB"] = struct{}{}
	nodes["CC"] = struct{}{}

	updateVotes(stat, nodes)
	assert.Equal(t, []string{"BB", "CC"}, stat.Votes.Addrs)
	assert.Equal(t, []string{"no", "no"}, stat.Votes.Votes)
}
