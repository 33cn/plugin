package executor

import (
	"fmt"

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

func (suite *NodeManageTestSuite) TestSupervisionExec() {
	suite.testSupervisionNodeConfigQuit()
	suite.testSupervisionNodeConfigApprove()
}

func (suite *NodeManageTestSuite) testSupervisionNodeConfigQuit() {
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
		Op:    pt.ParacrossSupervisionNodeQuit,
		Id:    g.Current.Id,
	}
	tx, err = createRawSupervisionNodeConfigTx(config)
	suite.Nil(err)

	receipt = nodeCommit(suite, PrivKey14K, tx)
	fmt.Println("***", receipt)
	fmt.Println("***", receipt.Logs[0].Ty)
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
