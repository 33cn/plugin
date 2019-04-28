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
	nodeConfigKey := calcParaNodeGroupKey(Title)
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

	return
}

func checkGroupApplyReceipt(suite *NodeManageTestSuite, receipt *types.Receipt) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 1)
	assert.Len(suite.T(), receipt.Logs, 1)

	var stat pt.ParaNodeAddrStatus
	err := types.Decode(receipt.KV[0].Value, &stat)
	assert.Nil(suite.T(), err, "decode ParaNodeAddrStatus failed")
	//suite.T().Log("titleHeight", titleHeight)
	assert.Equal(suite.T(), int32(pt.TyLogParaNodeGroupApply), receipt.Logs[0].Ty)
	assert.Equal(suite.T(), int32(pt.ParacrossNodeGroupApply), stat.Status)

}

func checkGroupApproveReceipt(suite *NodeManageTestSuite, receipt *types.Receipt) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 6)
	assert.Len(suite.T(), receipt.Logs, 6)

	len := len(receipt.KV)

	var stat pt.ParaNodeAddrStatus
	err := types.Decode(receipt.KV[len-1].Value, &stat)
	assert.Nil(suite.T(), err, "decode ParaNodeAddrStatus failed")
	//suite.T().Log("titleHeight", titleHeight)
	assert.Equal(suite.T(), int32(pt.TyLogParaNodeGroupApprove), receipt.Logs[len-1].Ty)
	assert.Equal(suite.T(), int32(pt.ParacrossNodeGroupApprove), stat.Status)

}

func checkJoinReceipt(suite *NodeManageTestSuite, receipt *types.Receipt) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 1)
	assert.Len(suite.T(), receipt.Logs, 1)

	var stat pt.ParaNodeAddrStatus
	err := types.Decode(receipt.KV[0].Value, &stat)
	assert.Nil(suite.T(), err, "decode ParaNodeAddrStatus failed")
	//suite.T().Log("titleHeight", titleHeight)
	assert.Equal(suite.T(), int32(pt.TyLogParaNodeConfig), receipt.Logs[0].Ty)
	assert.Equal(suite.T(), int32(pt.ParacrossNodeAdding), stat.Status)
	assert.NotNil(suite.T(), stat.Votes)

}

func checkQuitReceipt(suite *NodeManageTestSuite, receipt *types.Receipt) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 1)
	assert.Len(suite.T(), receipt.Logs, 1)

	var stat pt.ParaNodeAddrStatus
	err := types.Decode(receipt.KV[0].Value, &stat)
	assert.Nil(suite.T(), err, "decode ParaNodeAddrStatus failed")
	//suite.T().Log("titleHeight", titleHeight)
	assert.Equal(suite.T(), int32(pt.TyLogParaNodeConfig), receipt.Logs[0].Ty)
	assert.Equal(suite.T(), int32(pt.ParacrossNodeQuiting), stat.Status)
	assert.NotNil(suite.T(), stat.Votes)

}

func checkVoteReceipt(suite *NodeManageTestSuite, receipt *types.Receipt, count int) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 1)
	assert.Len(suite.T(), receipt.Logs, 1)

	var stat pt.ParaNodeAddrStatus
	err := types.Decode(receipt.KV[0].Value, &stat)
	assert.Nil(suite.T(), err, "decode ParaNodeAddrStatus failed")
	assert.Len(suite.T(), stat.Votes.Votes, count)

}

func checkVoteDoneReceipt(suite *NodeManageTestSuite, receipt *types.Receipt, count int, join bool) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 2)
	assert.Len(suite.T(), receipt.Logs, 3)

	var stat pt.ParaNodeAddrStatus
	err := types.Decode(receipt.KV[0].Value, &stat)
	assert.Nil(suite.T(), err, "decode ParaNodeAddrStatus failed")
	assert.Len(suite.T(), stat.Votes.Votes, count)

	var item types.ConfigItem
	err = types.Decode(receipt.KV[1].Value, &item)
	assert.Nil(suite.T(), err, "decode ParaNodeAddrStatus failed")
	if join {
		suite.Contains(item.GetArr().Value, Account14K)
	} else {
		suite.NotContains(item.GetArr().Value, Account14K)
	}

}

func voteTest(suite *NodeManageTestSuite, addr string, join bool) {
	config := &pt.ParaNodeAddrConfig{
		Op:    pt.ParaNodeVote,
		Addr:  addr,
		Value: pt.ParaNodeVoteYes,
	}
	tx, err := pt.CreateRawNodeConfigTx(config)
	suite.Nil(err)

	receipt := nodeCommit(suite, PrivKeyA, tx)
	checkVoteReceipt(suite, receipt, 1)

	receipt = nodeCommit(suite, PrivKeyB, tx)
	checkVoteReceipt(suite, receipt, 2)

	receipt = nodeCommit(suite, PrivKeyC, tx)
	checkVoteDoneReceipt(suite, receipt, 3, join)
}

func (suite *NodeManageTestSuite) testNodeGroupConfigQuit() {
	config := &pt.ParaNodeGroupApply{
		Addrs: applyAddrs,
		Op:    pt.ParacrossNodeGroupApply,
	}
	tx, err := pt.CreateRawNodeGroupApplyTx(config)
	suite.Nil(err)

	receipt := nodeCommit(suite, PrivKeyB, tx)
	checkGroupApplyReceipt(suite, receipt)

	config = &pt.ParaNodeGroupApply{
		Addrs: applyAddrs,
		Op:    pt.ParacrossNodeGroupQuit,
	}
	tx, err = pt.CreateRawNodeGroupApplyTx(config)
	suite.Nil(err)

	nodeCommit(suite, PrivKeyB, tx)
	//checkGroupApproveReceipt(suite, receipt)

}

func (suite *NodeManageTestSuite) testNodeGroupConfig() {
	suite.testNodeGroupConfigQuit()

	config := &pt.ParaNodeGroupApply{
		Addrs: applyAddrs,
		Op:    pt.ParacrossNodeGroupApply,
	}
	tx, err := pt.CreateRawNodeGroupApplyTx(config)
	suite.Nil(err)

	receipt := nodeCommit(suite, PrivKeyB, tx)
	checkGroupApplyReceipt(suite, receipt)

	config = &pt.ParaNodeGroupApply{
		Addrs: applyAddrs,
		Op:    pt.ParacrossNodeGroupApprove,
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

	//vote test
	voteTest(suite, Account14K, true)

	//Quit test
	config = &pt.ParaNodeAddrConfig{
		Op:   pt.ParaNodeQuit,
		Addr: Account1MC,
	}
	tx, err = pt.CreateRawNodeConfigTx(config)
	suite.Nil(err)
	receipt = nodeCommit(suite, PrivKeyD, tx)
	checkQuitReceipt(suite, receipt)

	//vote test
	voteTest(suite, Account1MC, false)
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
