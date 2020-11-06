package executor

import (
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/stretchr/testify/assert"
)

// createRawSupervisionNodeConfigTx create raw tx for node config
func createRawSupervisionNodeConfigTx(config *pt.ParaNodeAddrConfig) *types.Transaction {
	action := &pt.ParacrossAction{
		Ty:    pt.ParacrossActionSupervisionNodeConfig,
		Value: &pt.ParacrossAction_SupervisionNodeConfig{SupervisionNodeConfig: config},
	}
	tx := &types.Transaction{
		Payload: types.Encode(action),
		Execer:  []byte(config.Title + pt.ParaX),
	}
	return tx
}

func (suite *NodeManageTestSuite) testSupervisionExec() {
	suite.testSupervisionNodeConfigCancel(Account14K, PrivKey14K)
	suite.testSupervisionNodeConfigApprove(Account14K, PrivKey14K)
	suite.testSupervisionNodeConfigApprove(Account1Ku, PrivKey1Ku)
	suite.testSupervisionNodeConfigApprove(Account1M3, PrivKey1M3)
	suite.testSupervisionNodeError()
	suite.testSupervisionQuery()
	suite.testSupervisionNodeQuit()
}

func (suite *NodeManageTestSuite) testSupervisionNodeConfigCancel(addr, privKey string) {
	// Apply
	config := &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParacrossSupervisionNodeApply,
		Addr:  addr,
	}
	tx := createRawSupervisionNodeConfigTx(config)
	receipt := nodeCommit(suite, privKey, tx)
	checkSupervisionGroupApplyReceipt(suite, receipt)

	var g pt.ReceiptParaNodeGroupConfig
	err := types.Decode(receipt.Logs[0].Log, &g)
	suite.Nil(err)

	// cancel
	config = &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParacrossSupervisionNodeCancel,
		Id:    getParaNodeIDSuffix(g.Current.Id),
	}
	tx = createRawSupervisionNodeConfigTx(config)
	receipt = nodeCommit(suite, privKey, tx)
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
}

func (suite *NodeManageTestSuite) testSupervisionNodeConfigApprove(addr, privKey string) {
	config := &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParacrossSupervisionNodeApply,
		Addr:  addr,
	}
	tx := createRawSupervisionNodeConfigTx(config)
	receipt := nodeCommit(suite, privKey, tx)
	checkSupervisionGroupApplyReceipt(suite, receipt)

	var g pt.ReceiptParaNodeGroupConfig
	err := types.Decode(receipt.Logs[0].Log, &g)
	suite.Nil(err)

	config = &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Id:    getParaNodeIDSuffix(g.Current.Id),
		Op:    pt.ParacrossSupervisionNodeApprove,
	}

	tx = createRawSupervisionNodeConfigTx(config)
	receipt = nodeCommit(suite, privKey, tx)
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
}

func (suite *NodeManageTestSuite) testSupervisionNodeError() {
	config := &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParacrossSupervisionNodeApply,
		Addr:  Account1M3,
	}
	tx := createRawSupervisionNodeConfigTx(config)
	tx, _ = signTx(suite.Suite, tx, PrivKey1M3)
	_, err := suite.exec.Exec(tx, 0)
	assert.Equal(suite.T(), err, pt.ErrParaSupervisionNodeAddrExisted)

	config = &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParacrossSupervisionNodeApply,
		Addr:  "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",
	}
	tx = createRawSupervisionNodeConfigTx(config)
	tx, _ = signTx(suite.Suite, tx, PrivKey1KS)
	_, err = suite.exec.Exec(tx, 0)
	assert.Equal(suite.T(), err, pt.ErrParaNodeAddrExisted)
}

func (suite *NodeManageTestSuite) testSupervisionNodeQuit() {
	config := &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParacrossSupervisionNodeQuit,
		Addr:  Account1Ku,
	}
	tx := createRawSupervisionNodeConfigTx(config)
	receipt := nodeCommit(suite, PrivKey1Ku, tx)
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 3)
	assert.Len(suite.T(), receipt.Logs, 3)
	assert.Equal(suite.T(), int32(pt.TyLogParaSupervisionNodeGroupAddrsUpdate), receipt.Logs[0].Ty)

	ret, err := suite.exec.Query_GetSupervisionNodeGroupAddrs(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle()})
	suite.Nil(err)
	resp, ok := ret.(*types.ReplyConfig)
	assert.Equal(suite.T(), ok, true)
	assert.Equal(suite.T(), resp.Value, "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt,1M3XCbWVxAPBH5AR8VmLky4ZtDdGgC6ugD")
}

func checkSupervisionGroupApplyReceipt(suite *NodeManageTestSuite, receipt *types.Receipt) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 1)
	assert.Len(suite.T(), receipt.Logs, 1)
	assert.Equal(suite.T(), int32(pt.TyLogParaSupervisionNodeConfig), receipt.Logs[0].Ty)
}

func (suite *NodeManageTestSuite) testSupervisionQuery() {
	ret, err := suite.exec.Query_GetSupervisionNodeGroupAddrs(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle()})
	suite.Nil(err)
	resp, ok := ret.(*types.ReplyConfig)
	assert.Equal(suite.T(), ok, true)
	assert.Equal(suite.T(), resp.Value, "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt,1KufZaLTKVAy37AsXNd9bsva5WZvP8w5uG,1M3XCbWVxAPBH5AR8VmLky4ZtDdGgC6ugD")

	ret, err = suite.exec.Query_GetNodeAddrInfo(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle(), Addr: Account14K})
	suite.Nil(err)
	resp2, ok := ret.(*pt.ParaNodeAddrIdStatus)
	assert.Equal(suite.T(), ok, true)
	assert.NotNil(suite.T(), resp2)
}
