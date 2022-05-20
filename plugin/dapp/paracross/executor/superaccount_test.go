// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	//"github.com/stretchr/testify/mock"
	"strings"
	"testing"

	apimock "github.com/33cn/chain33/client/mocks"
	dbm "github.com/33cn/chain33/common/db"
	dbmock "github.com/33cn/chain33/common/db/mocks"
	"github.com/33cn/chain33/types"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/stretchr/testify/mock"
)

var (
	PrivKey14K = "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" // 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
	Account14K = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
	Bls14K     = "80e713aae96a44607ba6e0f1acfe88641ac72b789e81696cb646b1e1ae5335bd92011593eee303f9e909fd752c762db3"
	applyAddrs = "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4,1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR,1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k,1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
	PrivKey1KS = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
	Account12Q = "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
	PrivKey12Q = "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	PrivKey1Ku = "0xb1474387fe5beda4d6a9d999226bab561c708e07da15d74d575dad890e24cef0"
	Account1Ku = "1KufZaLTKVAy37AsXNd9bsva5WZvP8w5uG"
	PrivKey1M3 = "0xa4ebb30bd017f3d8e60532dd4f7059704d7897071fc07482aa0681c57c88f874"
	Account1M3 = "1M3XCbWVxAPBH5AR8VmLky4ZtDdGgC6ugD"
)

// createRawNodeConfigTx create raw tx for node config
func createRawNodeConfigTx(config *pt.ParaNodeAddrConfig) (*types.Transaction, error) {
	action := &pt.ParacrossAction{
		Ty:    pt.ParacrossActionNodeConfig,
		Value: &pt.ParacrossAction_NodeConfig{NodeConfig: config},
	}
	tx := &types.Transaction{
		Payload: types.Encode(action),
		Execer:  []byte(config.Title + pt.ParaX),
	}
	return tx, nil
}

//createRawNodeGroupApplyTx create raw tx for node group
func createRawNodeGroupApplyTx(apply *pt.ParaNodeGroupConfig) (*types.Transaction, error) {
	apply.Id = strings.Trim(apply.Id, " ")

	action := &pt.ParacrossAction{
		Ty:    pt.ParacrossActionNodeGroupApply,
		Value: &pt.ParacrossAction_NodeGroupConfig{NodeGroupConfig: apply},
	}

	tx := &types.Transaction{
		Payload: types.Encode(action),
		Execer:  []byte(apply.Title + pt.ParaX),
	}

	return tx, nil
}

type NodeManageTestSuite struct {
	suite.Suite
	stateDB dbm.KV
	localDB *dbmock.KVDB
	api     *apimock.QueueProtocolAPI
	exec    *Paracross
}

func (suite *NodeManageTestSuite) SetupSuite() {
	suite.stateDB, _ = dbm.NewGoMemDB("state", "state", 1024)
	// memdb 不支持KVDB接口， 等测试完Exec ， 再扩展 memdb
	//suite.localDB, _ = dbm.NewGoMemDB("local", "local", 1024)
	suite.localDB = new(dbmock.KVDB)
	suite.api = new(apimock.QueueProtocolAPI)
	suite.api.On("GetConfig", mock.Anything).Return(chain33TestCfg, nil)

	block := &types.Block{
		Height:     1,
		MainHeight: 10,
	}
	detail := &types.BlockDetail{Block: block}
	details := &types.BlockDetails{Items: []*types.BlockDetail{detail}}
	suite.api.On("GetBlocks", mock.Anything).Return(details, nil)

	suite.exec = newParacross().(*Paracross)
	suite.exec.SetAPI(suite.api)
	suite.exec.SetLocalDB(suite.localDB)
	suite.exec.SetStateDB(suite.stateDB)
	suite.exec.SetEnv(0, 0, 0)
	suite.exec.SetBlockInfo([]byte(""), []byte(""), 3)
	enableParacrossTransfer = false

	chain33TestCfg.S("config.consensus.sub.para.MainForkParacrossCommitTx", int64(1))
	chain33TestCfg.S("config.consensus.sub.para.MainLoopCheckCommitTxDoneForkHeight", int64(1))
	chain33TestCfg.S("config.exec.sub.manage.superManager", []interface{}{Account12Q})

	// TODO, more fields
	// setup block
	blockDetail := &types.BlockDetail{
		Block: &types.Block{},
	}
	MainBlockHash10 = blockDetail.Block.Hash(chain33TestCfg)
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
	assert.Equal(suite.T(), int32(pt.TyLogParaNodeConfig), receipt.Logs[0].Ty)
	assert.Equal(suite.T(), int32(pt.ParaApplyJoining), stat.Status)
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
	assert.Equal(suite.T(), int32(pt.ParaApplyQuiting), stat.Status)
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
	_ = count
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
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParaOpVote,
		Id:    id,
		Value: pt.ParaVoteYes,
	}
	tx, err := createRawNodeConfigTx(config)
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
		Title: chain33TestCfg.GetTitle(),
		Addrs: applyAddrs,
		Op:    pt.ParacrossNodeGroupApply,
	}
	tx, err := createRawNodeGroupApplyTx(config)
	suite.Nil(err)

	receipt := nodeCommit(suite, PrivKeyB, tx)
	checkGroupApplyReceipt(suite, receipt)

	var g pt.ReceiptParaNodeGroupConfig
	err = types.Decode(receipt.Logs[0].Log, &g)
	suite.Nil(err)

	config = &pt.ParaNodeGroupConfig{
		Title: chain33TestCfg.GetTitle(),
		Id:    g.Current.Id,
		Op:    pt.ParacrossNodeGroupQuit,
	}
	tx, err = createRawNodeGroupApplyTx(config)
	suite.Nil(err)

	nodeCommit(suite, PrivKeyB, tx)
}

func (suite *NodeManageTestSuite) testNodeGroupConfig() {
	suite.testNodeGroupConfigQuit()

	config := &pt.ParaNodeGroupConfig{
		Title: chain33TestCfg.GetTitle(),
		Addrs: applyAddrs,
		Op:    pt.ParacrossNodeGroupApply,
	}
	tx, err := createRawNodeGroupApplyTx(config)
	suite.Nil(err)

	receipt := nodeCommit(suite, PrivKeyB, tx)
	checkGroupApplyReceipt(suite, receipt)

	var g pt.ReceiptParaNodeGroupConfig
	err = types.Decode(receipt.Logs[0].Log, &g)
	suite.Nil(err)

	config = &pt.ParaNodeGroupConfig{
		Title: chain33TestCfg.GetTitle(),
		Id:    g.Current.Id,
		Op:    pt.ParacrossNodeGroupApprove,
	}
	tx, err = createRawNodeGroupApplyTx(config)
	suite.Nil(err)

	receipt = nodeCommit(suite, PrivKey12Q, tx)
	checkGroupApproveReceipt(suite, receipt)
}

func (suite *NodeManageTestSuite) testNodeConfig() {
	//Join test
	config := &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParaOpNewApply,
		Addr:  Account14K,
	}
	tx, err := createRawNodeConfigTx(config)
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
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParaOpQuit,
		Addr:  Account14K,
	}
	tx, err = createRawNodeConfigTx(config)
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
	suite.testSuperExec()
	suite.testSupervisionExec()
}

func (suite *NodeManageTestSuite) testSuperExec() {
	suite.testNodeGroupConfig()
	suite.testNodeConfig()
	suite.testSuperQuery()
}

func TestNodeManageSuite(t *testing.T) {
	suite.Run(t, new(NodeManageTestSuite))
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

	stat.Votes = updateVotes(stat.Votes, nodes)
	assert.Equal(t, []string{"BB", "CC"}, stat.Votes.Addrs)
	assert.Equal(t, []string{"no", "no"}, stat.Votes.Votes)
}

func TestGetNodeIdSuffix(t *testing.T) {
	txID := "0xb6cd0274aa5f839fa2291ecfbfc626b494aacac7587a61e444e9f848a4c02d7b"
	id := "mavl-paracross-title-nodegroupid-user.p.para.-0xb6cd0274aa5f839fa2291ecfbfc626b494aacac7587a61e444e9f848a4c02d7b"
	rtID := getParaNodeIDSuffix(id)
	assert.Equal(t, txID, rtID)

	txID = "0xb6cd0274aa5f839fa2291ecfbfc626b494aacac7587a61e444e9f848a4c02d7b-1"
	id = "mavl-paracross-title-nodegroupid-user.p.para.-0xb6cd0274aa5f839fa2291ecfbfc626b494aacac7587a61e444e9f848a4c02d7b-1"
	rtID = getParaNodeIDSuffix(id)
	assert.Equal(t, txID, rtID)
}

func (suite *NodeManageTestSuite) testSuperQuery() {
	ret, err := suite.exec.Query_GetNodeGroupAddrs(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle()})
	suite.Nil(err)
	resp, ok := ret.(*types.ReplyConfig)
	assert.Equal(suite.T(), ok, true)
	assert.Equal(suite.T(), resp.Value, applyAddrs)

	ret, err = suite.exec.Query_GetNodeAddrInfo(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle(), Addr: "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"})
	suite.Nil(err)
	resp2, ok := ret.(*pt.ParaNodeAddrIdStatus)
	assert.Equal(suite.T(), ok, true)
	assert.NotNil(suite.T(), resp2)

	_, err = suite.exec.Query_GetNodeAddrInfo(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle(), Addr: "1FbS6G4CRYAYeSEPGg7uKP9MukUo6crEE5"})
	suite.NotNil(err)

	ret, err = suite.exec.Query_GetNodeGroupStatus(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle()})
	suite.Nil(err)
	resp3, ok := ret.(*pt.ParaNodeGroupStatus)
	assert.Equal(suite.T(), ok, true)
	assert.Equal(suite.T(), resp3.Status, int32(pt.ParacrossNodeGroupApprove))

	_, err = suite.exec.Query_GetNodeIDInfo(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle(), Id: "mavl-paracross-title-nodeid-user.p.test.-0x8cf0e600667b8e6cf66516369acd4e1b5f6c93b3ae1c0b5edf458dfbe01f1607"})
	suite.Nil(err)
}
