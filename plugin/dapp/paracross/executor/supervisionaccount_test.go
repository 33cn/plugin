package executor

import (
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/stretchr/testify/assert"
)

// createRawSupervisionNodeConfigTx create raw tx for node config
func createRawSupervisionNodeConfigTx(config *pt.ParaNodeAddrConfig) (*types.Transaction, error) {
	action := &pt.ParacrossAction{
		Ty:    pt.ParacrossActionSupervisionNodeGroupConfig,
		Value: &pt.ParacrossAction_SupervisionNodeGroupConfig{SupervisionNodeGroupConfig: config},
	}
	tx := &types.Transaction{
		Payload: types.Encode(action),
		Execer:  []byte(config.Title + pt.ParaX),
	}
	return tx, nil
}

func (suite *NodeManageTestSuite) testSupervisionExec() {
	suite.testSupervisionNodeConfigQuit()
	suite.testSupervisionNodeConfigApprove()
	suite.testSupervisionQuery()
}

func (suite *NodeManageTestSuite) testSupervisionNodeConfigQuit() {
	// Apply
	config := &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParacrossSupervisionNodeApply,
		Addr:  Account14K,
	}
	tx, err := createRawSupervisionNodeConfigTx(config)
	suite.Nil(err)

	receipt := nodeCommit(suite, PrivKey14K, tx)
	checkSupervisionGroupApplyReceipt(suite, receipt)

	var g pt.ReceiptParaNodeGroupConfig
	err = types.Decode(receipt.Logs[0].Log, &g)
	suite.Nil(err)

	// Quit
	config = &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParacrossSupervisionNodeQuit,
		Id:    g.Current.Id,
	}
	tx, err = createRawSupervisionNodeConfigTx(config)
	suite.Nil(err)

	receipt = nodeCommit(suite, PrivKey14K, tx)
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
}

func (suite *NodeManageTestSuite) testSupervisionNodeConfigApprove() {
	config := &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Op:    pt.ParacrossSupervisionNodeApply,
		Addr:  Account14K,
	}
	tx, err := createRawSupervisionNodeConfigTx(config)
	suite.Nil(err)

	receipt := nodeCommit(suite, PrivKey14K, tx)
	checkSupervisionGroupApplyReceipt(suite, receipt)

	var g pt.ReceiptParaNodeGroupConfig
	err = types.Decode(receipt.Logs[0].Log, &g)
	suite.Nil(err)

	config = &pt.ParaNodeAddrConfig{
		Title: chain33TestCfg.GetTitle(),
		Id:    g.Current.Id,
		Op:    pt.ParacrossSupervisionNodeApprove,
	}
	tx, err = createRawSupervisionNodeConfigTx(config)
	suite.Nil(err)

	receipt = nodeCommit(suite, PrivKey14K, tx)
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
}

func checkSupervisionGroupApplyReceipt(suite *NodeManageTestSuite, receipt *types.Receipt) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 1)
	assert.Len(suite.T(), receipt.Logs, 1)
	assert.Equal(suite.T(), int32(pt.TyLogParaSupervisionNodeGroupConfig), receipt.Logs[0].Ty)
}

func (suite *NodeManageTestSuite) testSupervisionQuery() {
	ret, err := suite.exec.Query_GetSupervisionNodeGroupAddrs(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle()})
	suite.Nil(err)
	resp, ok := ret.(*types.ReplyConfig)
	assert.Equal(suite.T(), ok, true)
	assert.Equal(suite.T(), resp.Value, Account14K)

	ret, err = suite.exec.Query_GetSupervisionNodeAddrInfo(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle(), Addr: Account14K})
	suite.Nil(err)
	resp2, ok := ret.(*pt.ParaNodeAddrIdStatus)
	assert.Equal(suite.T(), ok, true)
	assert.NotNil(suite.T(), resp2)

	ret, err = suite.exec.Query_GetSupervisionNodeGroupStatus(&pt.ReqParacrossNodeInfo{Title: chain33TestCfg.GetTitle()})
	suite.Nil(err)
	resp3, ok := ret.(*pt.ParaNodeGroupStatus)
	assert.Equal(suite.T(), ok, true)
	assert.Equal(suite.T(), resp3.Status, int32(pt.ParacrossSupervisionNodeApprove))
}
