// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	//"github.com/stretchr/testify/mock"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	dbmock "github.com/33cn/chain33/common/db/mocks"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/paracross/testnode"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

// para-exec addr on main 1HPkPopVe3ERfvaAgedDtJQ792taZFEHCe
// para-exec addr on para 16zsMh7mvNDKPG6E9NVrPhw6zL93gWsTpR

var (
	Amount = types.DefaultCoinPrecision
)

//func para_init(title string) {
//	cfg, _ := types.InitCfgString(testnode.DefaultConfig)
//	types.Init(title, cfg)
//}

// 构建跨链交易, 用1个节点即可， 不测试共识
//    assetTransfer
//	 分别测试在主链和平行链的情况

type AssetTransferTestSuite struct {
	suite.Suite
	stateDB dbm.KV
	localDB *dbmock.KVDB
	api     *apimock.QueueProtocolAPI

	exec *Paracross
}

func TestAssetTransfer(t *testing.T) {
	suite.Run(t, new(AssetTransferTestSuite))
}

func (suite *AssetTransferTestSuite) SetupTest() {
	suite.stateDB, _ = dbm.NewGoMemDB("state", "state", 1024)
	// memdb 不支持KVDB接口， 等测试完Exec ， 再扩展 memdb
	//suite.localDB, _ = dbm.NewGoMemDB("local", "local", 1024)
	suite.localDB = new(dbmock.KVDB)

	suite.exec = newParacross().(*Paracross)
	suite.api = new(apimock.QueueProtocolAPI)
	suite.api.On("GetConfig", mock.Anything).Return(chain33TestCfg, nil)
	suite.exec.SetAPI(suite.api)
	suite.exec.SetLocalDB(suite.localDB)
	suite.exec.SetStateDB(suite.stateDB)
	suite.exec.SetEnv(0, 0, 0)
	enableParacrossTransfer = true

	// setup block
	blockDetail := &types.BlockDetail{
		Block: &types.Block{},
	}
	MainBlockHash10 = blockDetail.Block.Hash(chain33TestCfg)

	// setup title nodes : len = 1
	nodeConfigKey := calcManageConfigNodesKey(Title)
	nodeValue := makeNodeInfo(Title, Title, 1)
	suite.stateDB.Set(nodeConfigKey, types.Encode(nodeValue))
	value, err := suite.stateDB.Get(nodeConfigKey)
	if err != nil {
		suite.T().Error("get setup title failed", err)
		return
	}
	assert.Equal(suite.T(), value, types.Encode(nodeValue))

	// setup state title 'test' height is 9
	var titleStatus pt.ParacrossStatus
	titleStatus.Title = Title
	titleStatus.Height = CurHeight - 1
	titleStatus.BlockHash = PerBlock
	saveTitle(suite.stateDB, calcTitleKey(Title), &titleStatus)

	// setup api
	hashes := &types.ReqHashes{Hashes: [][]byte{MainBlockHash10}}
	suite.api.On("GetBlockByHashes", hashes).Return(
		&types.BlockDetails{
			Items: []*types.BlockDetail{blockDetail},
		}, nil)
	suite.api.On("GetBlockHash", &types.ReqInt{Height: MainBlockHeight}).Return(
		&types.ReplyHash{Hash: MainBlockHash10}, nil)
}

func (suite *AssetTransferTestSuite) TestExecTransferNobalance() {
	//types.Init("test", nil)
	suite.api = new(apimock.QueueProtocolAPI)
	suite.api.On("GetConfig", mock.Anything).Return(chain33TestMainCfg, nil)
	suite.exec.SetAPI(suite.api)

	toB := Nodes[1]
	tx, err := createAssetTransferTx(suite.Suite, PrivKeyD, toB)
	if err != nil {
		suite.T().Error("TestExecTransfer", "createTxGroup", err)
		return
	}

	_, err = suite.exec.Exec(tx, 1)
	if errors.Cause(err) != types.ErrNoBalance {
		suite.T().Error("Exec Transfer", err)
		return
	}
}

func (suite *AssetTransferTestSuite) TestExecTransfer() {
	//types.Init("test", nil)
	suite.api = new(apimock.QueueProtocolAPI)
	suite.api.On("GetConfig", mock.Anything).Return(chain33TestMainCfg, nil)
	suite.exec.SetAPI(suite.api)

	toB := Nodes[1]

	total := 1000 * types.DefaultCoinPrecision
	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[0]),
	}
	acc := account.NewCoinsAccount(chain33TestCfg)
	acc.SetDB(suite.stateDB)
	addrMain := address.ExecAddress(pt.ParaX)
	addrPara := address.ExecAddress(Title + pt.ParaX)

	acc.SaveExecAccount(addrMain, &accountA)

	tx, err := createAssetTransferTx(suite.Suite, PrivKeyA, toB)
	if err != nil {
		suite.T().Error("TestExecTransfer", "createTxGroup", err)
		return
	}
	suite.T().Log(string(tx.Execer))
	receipt, err := suite.exec.Exec(tx, 1)
	if err != nil {
		suite.T().Error("Exec Transfer", err)
		return
	}
	for _, kv := range receipt.KV {
		var v types.Account
		err = types.Decode(kv.Value, &v)
		if err != nil {
			// skip, only check frozen
			continue
		}
		suite.T().Log(string(kv.Key), v)
	}
	suite.T().Log("para-exec addr on main", addrMain)
	suite.T().Log("para-exec addr on para", addrPara)
	suite.T().Log("para-exec addr for A account", accountA.Addr)
	accTest := acc.LoadExecAccount(addrPara, addrMain)
	assert.Equal(suite.T(), Amount, accTest.Balance)

	resultA := acc.LoadExecAccount(string(Nodes[0]), addrMain)
	assert.Equal(suite.T(), total-Amount, resultA.Balance)
}

func (suite *AssetTransferTestSuite) TestExecTransferInPara() {
	chain33TestCfg = types.NewChain33Config(testnode.DefaultConfig)
	//para_init(Title)
	toB := Nodes[1]

	tx, err := createAssetTransferTx(suite.Suite, PrivKeyA, toB)
	if err != nil {
		suite.T().Error("TestExecTransfer", "createTxGroup", err)
		return
	}

	receipt, err := suite.exec.Exec(tx, 1)
	if err != nil {
		suite.T().Error("Exec Transfer", err)
		return
	}
	for _, kv := range receipt.KV {
		var v types.Account
		err = types.Decode(kv.Value, &v)
		if err != nil {
			// skip, only check frozen
			continue
		}
		suite.T().Log(string(kv.Key), v)
	}

	acc, _ := NewParaAccount(chain33TestCfg, Title, "coins", "bty", suite.stateDB)
	resultB := acc.LoadAccount(string(toB))
	assert.Equal(suite.T(), Amount, resultB.Balance)
}

func createAssetTransferTx(s suite.Suite, privFrom string, to []byte) (*types.Transaction, error) {
	param := types.CreateTx{
		To:          string(to),
		Amount:      Amount,
		Fee:         0,
		Note:        []byte("test asset transfer"),
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: "",
		ExecName:    Title + pt.ParaX,
	}
	tx, err := pt.CreateRawAssetTransferTx(chain33TestCfg, &param)
	assert.Nil(s.T(), err, "create asset transfer failed")
	if err != nil {
		return nil, err
	}

	tx, err = signTx(s, tx, privFrom)
	assert.Nil(s.T(), err, "sign asset transfer failed")
	if err != nil {
		return nil, err
	}

	return tx, nil
}

const TestSymbol = "TEST"

func (suite *AssetTransferTestSuite) TestExecTransferToken() {
	//types.Init("test", nil)
	suite.api = new(apimock.QueueProtocolAPI)
	suite.api.On("GetConfig", mock.Anything).Return(chain33TestMainCfg, nil)
	suite.exec.SetAPI(suite.api)

	toB := Nodes[1]

	total := 1000 * types.DefaultCoinPrecision
	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[0]),
	}
	acc, _ := account.NewAccountDB(chain33TestMainCfg, "token", TestSymbol, suite.stateDB)
	addrMain := address.ExecAddress(pt.ParaX)
	addrPara := address.ExecAddress(Title + pt.ParaX)

	acc.SaveExecAccount(addrMain, &accountA)

	tx, err := createAssetTransferTokenTx(suite.Suite, PrivKeyA, toB)
	if err != nil {
		suite.T().Error("TestExecTransfer", "createTxGroup", err)
		return
	}
	suite.T().Log(string(tx.Execer))
	receipt, err := suite.exec.Exec(tx, 1)
	if err != nil {
		suite.T().Error("Exec Transfer", err)
		return
	}
	for _, kv := range receipt.KV {
		var v types.Account
		err = types.Decode(kv.Value, &v)
		if err != nil {
			// skip, only check frozen
			continue
		}
		suite.T().Log(string(kv.Key), v)
	}
	suite.T().Log("para-exec addr on main", addrMain)
	suite.T().Log("para-exec addr on para", addrPara)
	suite.T().Log("para-exec addr for A account", accountA.Addr)
	accTest := acc.LoadExecAccount(addrPara, addrMain)
	assert.Equal(suite.T(), Amount, accTest.Balance)

	resultA := acc.LoadExecAccount(string(Nodes[0]), addrMain)
	assert.Equal(suite.T(), total-Amount, resultA.Balance)
}

func (suite *AssetTransferTestSuite) TestExecTransferTokenInPara() {
	chain33TestCfg = types.NewChain33Config(testnode.DefaultConfig)
	// para_init(Title)
	toB := Nodes[1]

	tx, err := createAssetTransferTokenTx(suite.Suite, PrivKeyA, toB)
	if err != nil {
		suite.T().Error("TestExecTransfer", "createTxGroup", err)
		return
	}

	receipt, err := suite.exec.Exec(tx, 1)
	if err != nil {
		suite.T().Error("Exec Transfer", err)
		return
	}
	for _, kv := range receipt.KV {
		var v types.Account
		err = types.Decode(kv.Value, &v)
		if err != nil {
			// skip, only check frozen
			continue
		}
		suite.T().Log(string(kv.Key), v)
	}

	acc, _ := NewParaAccount(chain33TestCfg, Title, "token", TestSymbol, suite.stateDB)
	resultB := acc.LoadAccount(string(toB))
	assert.Equal(suite.T(), Amount, resultB.Balance)
}

func createAssetTransferTokenTx(s suite.Suite, privFrom string, to []byte) (*types.Transaction, error) {
	param := types.CreateTx{
		To:          string(to),
		Amount:      Amount,
		Fee:         0,
		Note:        []byte("test asset transfer"),
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: TestSymbol,
		ExecName:    Title + pt.ParaX,
	}
	tx, err := pt.CreateRawAssetTransferTx(chain33TestCfg, &param)
	assert.Nil(s.T(), err, "create asset transfer failed")
	if err != nil {
		return nil, err
	}

	tx, err = signTx(s, tx, privFrom)
	assert.Nil(s.T(), err, "sign asset transfer failed")
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func TestGetCrossAction(t *testing.T) {
	txExec := "paracross"
	transfer := &pt.CrossAssetTransfer{AssetExec: "coins", AssetSymbol: "bty"}
	action, err := getCrossAction(transfer, txExec)
	assert.NotNil(t, err)
	assert.Equal(t, int64(pt.ParacrossNoneTransfer), action)

	txExec = "user.p.para.paracross."
	transfer = &pt.CrossAssetTransfer{AssetExec: "user.p.test.coins", AssetSymbol: "bty"}
	action, err = getCrossAction(transfer, txExec)
	t.Log("ParacrossNoneTransfer e=", err)
	assert.NotNil(t, err)
	assert.Equal(t, int64(pt.ParacrossNoneTransfer), action)

	transfer = &pt.CrossAssetTransfer{AssetExec: "coins", AssetSymbol: "bty"}
	action, err = getCrossAction(transfer, txExec)
	assert.Nil(t, err)
	assert.Equal(t, int64(pt.ParacrossMainAssetTransfer), action)

	transfer = &pt.CrossAssetTransfer{AssetExec: "paracross", AssetSymbol: "user.p.para.coins.cbt"}
	action, err = getCrossAction(transfer, txExec)
	assert.Nil(t, err)
	assert.Equal(t, int64(pt.ParacrossParaAssetWithdraw), action)

	transfer = &pt.CrossAssetTransfer{AssetExec: "paracross", AssetSymbol: "user.p.test.coins.cbt"}
	action, err = getCrossAction(transfer, txExec)
	assert.Nil(t, err)
	assert.Equal(t, int64(pt.ParacrossMainAssetTransfer), action)

	transfer = &pt.CrossAssetTransfer{AssetExec: "user.p.para.coins", AssetSymbol: "bty"}
	action, err = getCrossAction(transfer, txExec)
	assert.Nil(t, err)
	assert.Equal(t, int64(pt.ParacrossParaAssetTransfer), action)

	transfer = &pt.CrossAssetTransfer{AssetExec: "user.p.para.paracross", AssetSymbol: "coin.bty"}
	action, err = getCrossAction(transfer, txExec)
	assert.Nil(t, err)
	assert.Equal(t, int64(pt.ParacrossMainAssetWithdraw), action)

	transfer = &pt.CrossAssetTransfer{AssetExec: "user.p.para.paracross", AssetSymbol: "paracross.user.p.test.coin.bty"}
	action, err = getCrossAction(transfer, txExec)
	assert.Nil(t, err)
	assert.Equal(t, int64(pt.ParacrossMainAssetWithdraw), action)

}

func TestAmendTransferParam(t *testing.T) {
	act := int64(pt.ParacrossMainAssetTransfer)
	transfer := &pt.CrossAssetTransfer{AssetExec: "coins", AssetSymbol: "bty"}
	rst, err := amendTransferParam(transfer, act)
	assert.Nil(t, err)
	assert.Equal(t, transfer.AssetExec, rst.AssetExec)
	assert.Equal(t, transfer.AssetSymbol, rst.AssetSymbol)

	transfer = &pt.CrossAssetTransfer{AssetExec: "paracross", AssetSymbol: "user.p.para.coins.bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.Nil(t, err)
	assert.Equal(t, transfer.AssetExec, rst.AssetExec)
	assert.Equal(t, transfer.AssetSymbol, rst.AssetSymbol)

	//
	act = int64(pt.ParacrossMainAssetTransfer)
	transfer = &pt.CrossAssetTransfer{AssetExec: "token", AssetSymbol: "coins.bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.NotNil(t, err)
	t.Log("token.coins.bty,err=", err)

	transfer = &pt.CrossAssetTransfer{AssetExec: "paracross", AssetSymbol: "bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.NotNil(t, err)
	t.Log("paracross.bty,err=", err)

	transfer = &pt.CrossAssetTransfer{AssetExec: "user.p.para.coins", AssetSymbol: "coins.bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.NotNil(t, err)
	t.Log("user.p.para.coins.coins.bty,err=", err)

	transfer = &pt.CrossAssetTransfer{AssetExec: "user.p.para.paracross", AssetSymbol: "bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.NotNil(t, err)
	t.Log("user.p.para.paracross.bty,err=", err)

	//
	act = int64(pt.ParacrossMainAssetWithdraw)
	transfer = &pt.CrossAssetTransfer{AssetExec: "user.p.test.paracross", AssetSymbol: "coins.bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.Nil(t, err)
	assert.Equal(t, "coins", rst.AssetExec)
	assert.Equal(t, "bty", rst.AssetSymbol)

	act = int64(pt.ParacrossMainAssetWithdraw)
	transfer = &pt.CrossAssetTransfer{AssetExec: "user.p.test2.paracross", AssetSymbol: "paracross.user.p.test.coins.bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.Nil(t, err)
	assert.Equal(t, "paracross", rst.AssetExec)
	assert.Equal(t, "user.p.test.coins.bty", rst.AssetSymbol)

	act = int64(pt.ParacrossMainAssetWithdraw)
	transfer = &pt.CrossAssetTransfer{AssetExec: "user.p.test.paracross", AssetSymbol: "bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.NotNil(t, err)

	//
	act = int64(pt.ParacrossParaAssetTransfer)
	transfer = &pt.CrossAssetTransfer{AssetExec: "user.p.test.coins", AssetSymbol: "bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.Nil(t, err)
	assert.Equal(t, "coins", rst.AssetExec)
	assert.Equal(t, "bty", rst.AssetSymbol)

	//
	act = int64(pt.ParacrossParaAssetWithdraw)
	transfer = &pt.CrossAssetTransfer{AssetExec: "paracross", AssetSymbol: "user.p.test.coins.bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.Nil(t, err)
	assert.Equal(t, "coins", rst.AssetExec)
	assert.Equal(t, "bty", rst.AssetSymbol)

	act = int64(pt.ParacrossParaAssetWithdraw)
	transfer = &pt.CrossAssetTransfer{AssetExec: "paracross", AssetSymbol: "bty"}
	rst, err = amendTransferParam(transfer, act)
	assert.NotNil(t, err)

}
