// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	//"github.com/stretchr/testify/mock"
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
	suite.api.On("GetConfig", mock.Anything).Return(chain33TestCfg, nil)

	suite.exec = newParacross().(*Paracross)
	suite.exec.SetAPI(suite.api)
	suite.exec.SetLocalDB(suite.localDB)
	suite.exec.SetStateDB(suite.stateDB)
	suite.exec.SetEnv(0, 0, 0)
	suite.exec.SetBlockInfo([]byte(""), []byte(""), 3)
	enableParacrossTransfer = false

	//forkHeight := types.GetDappFork(pt.ParaX, pt.ForkCommitTx)
	//if forkHeight == types.MaxHeight {
	//	types.ReplaceDappFork(MainTitle, pt.ParaX, pt.ForkCommitTx, 0)
	//
	//}

	chain33TestCfg.S("config.consensus.sub.para.MainForkParacrossCommitTx", int64(1))
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
	//suite.T().Log("titleHeight", titleHeight)
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
		Title: chain33TestCfg.GetTitle(),
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
		Title: chain33TestCfg.GetTitle(),
		Id:    g.Current.Id,
		Op:    pt.ParacrossNodeGroupQuit,
	}
	tx, err = pt.CreateRawNodeGroupApplyTx(config)
	suite.Nil(err)

	nodeCommit(suite, PrivKeyB, tx)
	//checkGroupApproveReceipt(suite, receipt)

}

func (suite *NodeManageTestSuite) testNodeGroupConfig() {
	suite.testNodeGroupConfigQuit()

	config := &pt.ParaNodeGroupConfig{
		Title: chain33TestCfg.GetTitle(),
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
		Title: chain33TestCfg.GetTitle(),
		Id:    g.Current.Id,
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
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParaOpNewApply,
		Addr:  Account14K,
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
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParaOpQuit,
		Addr:  Account14K,
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

func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func BenchmarkDeepCody(b *testing.B) {
	hexBlock := "0aad071220d29ccba4a90178614c8036f962201fdf56e77b419ca570371778e129a0e0e2841a20dd690056e7719d5b5fd1ad3e76f3b4366db9f3278ca29fbe4fc523bbf756fa8a222064cdbc89fe14ae7b851a71e41bd36a82db8cd3b69f4434f6420b279fea4fd25028e90330d386c4ea053a99040a067469636b657412ed02501022e80208ffff8bf8011080cab5ee011a70313271796f6361794e46374c7636433971573461767873324537553431664b5366763a3078386434663130653666313762616533636538663764303239616263623461393839616563333333386238386662333537656165663035613265326465373930343a30303030303035363533224230783537656632356366363036613734393462626164326432666233373734316137636332346663663066653064303637363638636564306235653961363735336632206b9836b2d295ca16634ea83359342db3ab1ab5f15200993bf6a09024089b7b693a810136f5a704a427a0562653345659193a0e292f16200ce67d6a8f8af631149a904fb80f0c12f525f7ce208fbf2183f2052af6252a108bb2614db0ccf8d91d4e910a04c472f5113275fe937f68ed6b2b0b522cc5fc2594b9fec60c0b22524b828333aaa982be125ec0f69645c86de3d331b26aa84c29a06824e977ce51f76f34f0629c1a6d080112210320bbac09528e19c55b0f89cb37ab265e7e856b1a8c388780322dbbfd194b52ba1a46304402200971163de32cb6a17e925eb3fcec8a0ccc193493635ecbf52357f5365edc2c82022039d84aa7078bc51ef83536038b7fd83087bf4deb965370f211a5589d4add551720a08d063097c5abcc9db1e4a92b3a22313668747663424e53454137665a6841644c4a706844775152514a614870794854703af4010a05746f6b656e124438070a400a0879696e6865626962120541424344452080d0dbc3f4022864322231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b38011a6d08011221021afd97c3d9a47f7ead3ca34fe5a73789714318df2c608cf2c7962378dc858ecb1a4630440220316a241f19b392e685ef940dee48dc53f90fc5c8be108ceeef1d3c65f530ee5f02204c89708cc7dac408a88e9fa6ad8e723d93fc18fe114d4a416e280f5a15656b0920a08d0628ca87c4ea0530dd91be81ae9ed1882d3a22313268704a4248796268316d537943696a51324d514a506b377a376b5a376a6e516158ffff8bf80162200b97166ad507aea57a4bb6e6b9295ec082cdc670b8468a83b559dbd900ffb83068e90312e80508021a5e0802125a0a2b1080978dcaef192222313271796f6361794e46374c7636433971573461767873324537553431664b536676122b10e08987caef192222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a9f010871129a010a70313271796f6361794e46374c7636433971573461767873324537553431664b5366763a3078386434663130653666313762616533636538663764303239616263623461393839616563333333386238386662333537656165663035613265326465373930343a30303030303035363533100218012222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a620805125e0a2d1080e0cd90f6a6ca342222313668747663424e53454137665a6841644c4a706844775152514a61487079485470122d1080aa83fff7a6ca342222313668747663424e53454137665a6841644c4a706844775152514a614870794854701a870108081282010a22313668747663424e53454137665a6841644c4a706844775152514a61487079485470122d1880f0cae386e68611222231344b454b6259744b4b516d34774d7468534b394a344c61346e41696964476f7a741a2d1880ba80d288e68611222231344b454b6259744b4b516d34774d7468534b394a344c61346e41696964476f7a741a620805125e0a2d1080aa83fff7a6ca342222313668747663424e53454137665a6841644c4a706844775152514a61487079485470122d1080f0898ef9a6ca342222313668747663424e53454137665a6841644c4a706844775152514a614870794854701a8f010808128a010a22313668747663424e53454137665a6841644c4a706844775152514a6148707948547012311080e0ba84bf03188090c0ac622222314251585336547861595947356d41446157696a344178685a5a55547077393561351a311080e0ba84bf031880d6c6bb632222314251585336547861595947356d41446157696a344178685a5a555470773935613512970208021a5c080212580a2a10c099b8c321222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b122a10a08cb2c321222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b1a82010809127e0a22313268704a4248796268316d537943696a51324d514a506b377a376b5a376a6e5161122a108094ebdc03222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b1a2c109c93ebdc031864222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b1a3008d301122b0a054142434445122231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b"
	bs, err := hex.DecodeString(hexBlock)
	if err != nil {
		return
	}
	var block types.BlockDetail
	err = types.Decode(bs, &block)
	if err != nil {
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var b types.BlockDetail
		err = deepCopy(b, block)
	}

}

func BenchmarkProtoClone(b *testing.B) {
	hexBlock := "0aad071220d29ccba4a90178614c8036f962201fdf56e77b419ca570371778e129a0e0e2841a20dd690056e7719d5b5fd1ad3e76f3b4366db9f3278ca29fbe4fc523bbf756fa8a222064cdbc89fe14ae7b851a71e41bd36a82db8cd3b69f4434f6420b279fea4fd25028e90330d386c4ea053a99040a067469636b657412ed02501022e80208ffff8bf8011080cab5ee011a70313271796f6361794e46374c7636433971573461767873324537553431664b5366763a3078386434663130653666313762616533636538663764303239616263623461393839616563333333386238386662333537656165663035613265326465373930343a30303030303035363533224230783537656632356366363036613734393462626164326432666233373734316137636332346663663066653064303637363638636564306235653961363735336632206b9836b2d295ca16634ea83359342db3ab1ab5f15200993bf6a09024089b7b693a810136f5a704a427a0562653345659193a0e292f16200ce67d6a8f8af631149a904fb80f0c12f525f7ce208fbf2183f2052af6252a108bb2614db0ccf8d91d4e910a04c472f5113275fe937f68ed6b2b0b522cc5fc2594b9fec60c0b22524b828333aaa982be125ec0f69645c86de3d331b26aa84c29a06824e977ce51f76f34f0629c1a6d080112210320bbac09528e19c55b0f89cb37ab265e7e856b1a8c388780322dbbfd194b52ba1a46304402200971163de32cb6a17e925eb3fcec8a0ccc193493635ecbf52357f5365edc2c82022039d84aa7078bc51ef83536038b7fd83087bf4deb965370f211a5589d4add551720a08d063097c5abcc9db1e4a92b3a22313668747663424e53454137665a6841644c4a706844775152514a614870794854703af4010a05746f6b656e124438070a400a0879696e6865626962120541424344452080d0dbc3f4022864322231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b38011a6d08011221021afd97c3d9a47f7ead3ca34fe5a73789714318df2c608cf2c7962378dc858ecb1a4630440220316a241f19b392e685ef940dee48dc53f90fc5c8be108ceeef1d3c65f530ee5f02204c89708cc7dac408a88e9fa6ad8e723d93fc18fe114d4a416e280f5a15656b0920a08d0628ca87c4ea0530dd91be81ae9ed1882d3a22313268704a4248796268316d537943696a51324d514a506b377a376b5a376a6e516158ffff8bf80162200b97166ad507aea57a4bb6e6b9295ec082cdc670b8468a83b559dbd900ffb83068e90312e80508021a5e0802125a0a2b1080978dcaef192222313271796f6361794e46374c7636433971573461767873324537553431664b536676122b10e08987caef192222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a9f010871129a010a70313271796f6361794e46374c7636433971573461767873324537553431664b5366763a3078386434663130653666313762616533636538663764303239616263623461393839616563333333386238386662333537656165663035613265326465373930343a30303030303035363533100218012222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a620805125e0a2d1080e0cd90f6a6ca342222313668747663424e53454137665a6841644c4a706844775152514a61487079485470122d1080aa83fff7a6ca342222313668747663424e53454137665a6841644c4a706844775152514a614870794854701a870108081282010a22313668747663424e53454137665a6841644c4a706844775152514a61487079485470122d1880f0cae386e68611222231344b454b6259744b4b516d34774d7468534b394a344c61346e41696964476f7a741a2d1880ba80d288e68611222231344b454b6259744b4b516d34774d7468534b394a344c61346e41696964476f7a741a620805125e0a2d1080aa83fff7a6ca342222313668747663424e53454137665a6841644c4a706844775152514a61487079485470122d1080f0898ef9a6ca342222313668747663424e53454137665a6841644c4a706844775152514a614870794854701a8f010808128a010a22313668747663424e53454137665a6841644c4a706844775152514a6148707948547012311080e0ba84bf03188090c0ac622222314251585336547861595947356d41446157696a344178685a5a55547077393561351a311080e0ba84bf031880d6c6bb632222314251585336547861595947356d41446157696a344178685a5a555470773935613512970208021a5c080212580a2a10c099b8c321222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b122a10a08cb2c321222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b1a82010809127e0a22313268704a4248796268316d537943696a51324d514a506b377a376b5a376a6e5161122a108094ebdc03222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b1a2c109c93ebdc031864222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b1a3008d301122b0a054142434445122231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b"
	bs, err := hex.DecodeString(hexBlock)
	if err != nil {
		return
	}
	var block types.BlockDetail
	err = types.Decode(bs, &block)
	if err != nil {
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = proto.Clone(&block).(*types.BlockDetail)
	}
}

func BenchmarkProtoMarshal(b *testing.B) {
	hexBlock := "0aad071220d29ccba4a90178614c8036f962201fdf56e77b419ca570371778e129a0e0e2841a20dd690056e7719d5b5fd1ad3e76f3b4366db9f3278ca29fbe4fc523bbf756fa8a222064cdbc89fe14ae7b851a71e41bd36a82db8cd3b69f4434f6420b279fea4fd25028e90330d386c4ea053a99040a067469636b657412ed02501022e80208ffff8bf8011080cab5ee011a70313271796f6361794e46374c7636433971573461767873324537553431664b5366763a3078386434663130653666313762616533636538663764303239616263623461393839616563333333386238386662333537656165663035613265326465373930343a30303030303035363533224230783537656632356366363036613734393462626164326432666233373734316137636332346663663066653064303637363638636564306235653961363735336632206b9836b2d295ca16634ea83359342db3ab1ab5f15200993bf6a09024089b7b693a810136f5a704a427a0562653345659193a0e292f16200ce67d6a8f8af631149a904fb80f0c12f525f7ce208fbf2183f2052af6252a108bb2614db0ccf8d91d4e910a04c472f5113275fe937f68ed6b2b0b522cc5fc2594b9fec60c0b22524b828333aaa982be125ec0f69645c86de3d331b26aa84c29a06824e977ce51f76f34f0629c1a6d080112210320bbac09528e19c55b0f89cb37ab265e7e856b1a8c388780322dbbfd194b52ba1a46304402200971163de32cb6a17e925eb3fcec8a0ccc193493635ecbf52357f5365edc2c82022039d84aa7078bc51ef83536038b7fd83087bf4deb965370f211a5589d4add551720a08d063097c5abcc9db1e4a92b3a22313668747663424e53454137665a6841644c4a706844775152514a614870794854703af4010a05746f6b656e124438070a400a0879696e6865626962120541424344452080d0dbc3f4022864322231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b38011a6d08011221021afd97c3d9a47f7ead3ca34fe5a73789714318df2c608cf2c7962378dc858ecb1a4630440220316a241f19b392e685ef940dee48dc53f90fc5c8be108ceeef1d3c65f530ee5f02204c89708cc7dac408a88e9fa6ad8e723d93fc18fe114d4a416e280f5a15656b0920a08d0628ca87c4ea0530dd91be81ae9ed1882d3a22313268704a4248796268316d537943696a51324d514a506b377a376b5a376a6e516158ffff8bf80162200b97166ad507aea57a4bb6e6b9295ec082cdc670b8468a83b559dbd900ffb83068e90312e80508021a5e0802125a0a2b1080978dcaef192222313271796f6361794e46374c7636433971573461767873324537553431664b536676122b10e08987caef192222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a9f010871129a010a70313271796f6361794e46374c7636433971573461767873324537553431664b5366763a3078386434663130653666313762616533636538663764303239616263623461393839616563333333386238386662333537656165663035613265326465373930343a30303030303035363533100218012222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a620805125e0a2d1080e0cd90f6a6ca342222313668747663424e53454137665a6841644c4a706844775152514a61487079485470122d1080aa83fff7a6ca342222313668747663424e53454137665a6841644c4a706844775152514a614870794854701a870108081282010a22313668747663424e53454137665a6841644c4a706844775152514a61487079485470122d1880f0cae386e68611222231344b454b6259744b4b516d34774d7468534b394a344c61346e41696964476f7a741a2d1880ba80d288e68611222231344b454b6259744b4b516d34774d7468534b394a344c61346e41696964476f7a741a620805125e0a2d1080aa83fff7a6ca342222313668747663424e53454137665a6841644c4a706844775152514a61487079485470122d1080f0898ef9a6ca342222313668747663424e53454137665a6841644c4a706844775152514a614870794854701a8f010808128a010a22313668747663424e53454137665a6841644c4a706844775152514a6148707948547012311080e0ba84bf03188090c0ac622222314251585336547861595947356d41446157696a344178685a5a55547077393561351a311080e0ba84bf031880d6c6bb632222314251585336547861595947356d41446157696a344178685a5a555470773935613512970208021a5c080212580a2a10c099b8c321222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b122a10a08cb2c321222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b1a82010809127e0a22313268704a4248796268316d537943696a51324d514a506b377a376b5a376a6e5161122a108094ebdc03222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b1a2c109c93ebdc031864222231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b1a3008d301122b0a054142434445122231513868474c666f47653633656665576138664a34506e756b686b6e677436706f4b"
	bs, err := hex.DecodeString(hexBlock)
	if err != nil {
		return
	}
	var block types.BlockDetail
	err = types.Decode(bs, &block)
	if err != nil {
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x, _ := proto.Marshal(&block)
		var b types.BlockDetail
		proto.Unmarshal(x, &b)
	}
}
