// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"strings"
	"testing"

	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	dbmock "github.com/33cn/chain33/common/db/mocks"
	"github.com/33cn/chain33/common/log"
	mty "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	_ "github.com/33cn/plugin/plugin/crypto/bls"
	"github.com/33cn/plugin/plugin/dapp/paracross/testnode"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	MainBlockHash10 = []byte("main block hash 10")
	MainBlockHeight = int64(10)
	CurHeight       = int64(10)
	Title           = "user.p.test."
	TitleHeight     = int64(10)
	PerBlock        = []byte("block-hash-9")
	CurBlock        = []byte("block-hash-10")

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

	PrivKeyE         = "0x3a35610ba6e1e72d7878f4c819e6a6768668cb5481f423ef04b6a11e0e16e44f" // 15HmJz2abkExxgcmSRt2Q5D4hZg6zJUD1h
	PrivKeyF         = "0xaffa90b6a897c798e63890312b2ec9fb5a3c156dac290479ccb67c25c78e9413" // 1JQjqDChawMYfG3yyxByrhJ467HorPVfFZ
	SupervisionNodes = [][]byte{
		[]byte("15HmJz2abkExxgcmSRt2Q5D4hZg6zJUD1h"),
		[]byte("1JQjqDChawMYfG3yyxByrhJ467HorPVfFZ"),
	}

	chain33TestCfg     = types.NewChain33Config(testnode.DefaultConfig)
	chain33TestMainCfg = types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"test\"", 1))
)

type CommitTestSuite struct {
	suite.Suite
	stateDB dbm.KV
	localDB *dbmock.KVDB
	api     *apimock.QueueProtocolAPI

	exec *Paracross
}

func makeNodeInfo(key, addr string, cnt int) *types.ConfigItem {
	var item types.ConfigItem
	item.Key = key
	item.Addr = addr
	item.Ty = mty.ConfigItemArrayConfig
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	for i, n := range Nodes {
		if i >= cnt {
			break
		}
		item.GetArr().Value = append(item.GetArr().Value, string(n))
	}
	return &item
}

func makeSupervisionNodeInfo(suite *CommitTestSuite) {
	SupervisionNodeKey := calcParaSupervisionNodeGroupAddrsKey(Title)
	var item types.ConfigItem
	item.Key = Title
	item.Addr = Title
	item.Ty = mty.ConfigItemArrayConfig
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	for _, n := range SupervisionNodes {
		item.GetArr().Value = append(item.GetArr().Value, string(n))
	}

	_ = suite.stateDB.Set(SupervisionNodeKey, types.Encode(&item))
	value, err := suite.stateDB.Get(SupervisionNodeKey)
	if err != nil {
		suite.T().Error("get setup title failed", err)
		return
	}
	assert.Equal(suite.T(), value, types.Encode(&item))
}

func init() {
	log.SetFileLog(nil)
	log.SetLogLevel("debug")
	Init(pt.ParaX, chain33TestCfg, nil)
}

func (suite *CommitTestSuite) SetupSuite() {
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
	enableParacrossTransfer = false

	// TODO, more fields
	// setup block
	blockDetail := &types.BlockDetail{
		Block: &types.Block{},
	}
	MainBlockHash10 = blockDetail.Block.Hash(chain33TestCfg)
	blockDetail.Block.MainHash = MainBlockHash10

	// setup title nodes : len = 4
	nodeConfigKey := calcManageConfigNodesKey(Title)
	nodeValue := makeNodeInfo(Title, Title, 4)
	_ = suite.stateDB.Set(nodeConfigKey, types.Encode(nodeValue))
	value, err := suite.stateDB.Get(nodeConfigKey)
	if err != nil {
		suite.T().Error("get setup title failed", err)
		return
	}
	assert.Equal(suite.T(), value, types.Encode(nodeValue))

	makeSupervisionNodeInfo(suite)

	stageKey := calcParaSelfConsStagesKey()
	stage := &pt.SelfConsensStage{StartHeight: 0, Enable: pt.ParaConfigYes}
	stages := &pt.SelfConsensStages{Items: []*pt.SelfConsensStage{stage}}
	_ = suite.stateDB.Set(stageKey, types.Encode(stages))
	value, err = suite.stateDB.Get(stageKey)
	if err != nil {
		suite.T().Error("get setup stages failed", err)
		return
	}
	assert.Equal(suite.T(), value, types.Encode(stages))

	// setup state title 'test' height is 9
	var titleStatus pt.ParacrossStatus
	titleStatus.Title = Title
	titleStatus.Height = CurHeight - 1
	titleStatus.BlockHash = PerBlock
	_ = saveTitle(suite.stateDB, calcTitleKey(Title), &titleStatus)

	// setup api
	hashes := &types.ReqHashes{Hashes: [][]byte{MainBlockHash10}}
	suite.api.On("GetBlockByHashes", hashes).Return(
		&types.BlockDetails{
			Items: []*types.BlockDetail{blockDetail},
		}, nil).Once()
	suite.api.On("GetBlocks", &types.ReqBlocks{Start: TitleHeight, End: TitleHeight}).Return(
		&types.BlockDetails{
			Items: []*types.BlockDetail{blockDetail},
		}, nil)
	suite.api.On("GetBlockHash", &types.ReqInt{Height: TitleHeight}).Return(
		&types.ReplyHash{Hash: CurBlock}, nil)
}

func (suite *CommitTestSuite) TestSetup() {
	nodeConfigKey := calcManageConfigNodesKey(Title)
	suite.T().Log(string(nodeConfigKey))
	_, err := suite.stateDB.Get(nodeConfigKey)
	if err != nil {
		suite.T().Error("get setup title failed", err)
		return
	}
}

func fillRawCommitTx(suite suite.Suite) (*types.Transaction, error) {
	st1 := &pt.ParacrossNodeStatus{
		MainBlockHash:   MainBlockHash10,
		MainBlockHeight: MainBlockHeight,
		Title:           Title,
		Height:          TitleHeight,
		PreBlockHash:    []byte("block-hash-9"),
		BlockHash:       []byte("block-hash-10"),
		PreStateHash:    []byte("state-hash-9"),
		StateHash:       []byte("state-hash-10"),
		TxCounts:        10,
		TxResult:        []byte("abc"),
		TxHashs:         [][]byte{},
		CrossTxResult:   []byte("abc"),
		CrossTxHashs:    [][]byte{},
	}
	act := &pt.ParacrossCommitAction{Status: st1}
	tx, err := pt.CreateRawCommitTx4MainChain(chain33TestCfg, act, pt.GetExecName(chain33TestCfg), 0)
	if err != nil {
		suite.T().Error("TestExec", "create tx failed", err)
	}
	return tx, err
}

func signTx(s suite.Suite, tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName("", signType), -1)
	if err != nil {
		s.T().Error("TestExec", "new crypto failed", err)
		return tx, err
	}

	bytesData, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		s.T().Error("TestExec", "Hex2Bytes privkey faiiled", err)
		return tx, err
	}

	privKey, err := c.PrivKeyFromBytes(bytesData)
	if err != nil {
		s.T().Error("TestExec", "PrivKeyFromBytes failed", err)
		return tx, err
	}

	tx.Sign(int32(signType), privKey)
	return tx, nil
}

func getPrivKey(s suite.Suite, hexPrivKey string) (crypto.PrivKey, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName("", signType), -1)
	if err != nil {
		s.T().Error("TestExec", "new crypto failed", err)
		return nil, err
	}

	bytesData, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		s.T().Error("TestExec", "Hex2Bytes privkey faiiled", err)
		return nil, err
	}

	privKey, err := c.PrivKeyFromBytes(bytesData)
	if err != nil {
		s.T().Error("TestExec", "PrivKeyFromBytes failed", err)
		return nil, err
	}

	return privKey, nil
}

func commitOnce(suite *CommitTestSuite, privkeyStr string) (receipt *types.Receipt) {
	return commitOnceImpl(suite.Suite, suite.exec, privkeyStr)
}

func commitOnceImpl(suite suite.Suite, exec *Paracross, privkeyStr string) (receipt *types.Receipt) {
	tx, _ := fillRawCommitTx(suite)
	tx, _ = signTx(suite, tx, privkeyStr)

	suite.T().Log(tx.From())
	receipt, err := exec.Exec(tx, 0)
	suite.T().Log(receipt)
	assert.NotNil(suite.T(), receipt)
	assert.Nil(suite.T(), err)

	return
}

func checkCommitReceipt(suite *CommitTestSuite, receipt *types.Receipt, commitCnt int, commitSupervisionCnt int) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 1)
	assert.Len(suite.T(), receipt.Logs, 1)

	key := calcTitleHeightKey(Title, TitleHeight)
	suite.T().Log("title height key", string(key))
	assert.Equal(suite.T(), key, receipt.KV[0].Key,
		"receipt not match", string(key), string(receipt.KV[0].Key))

	var titleHeight pt.ParacrossHeightStatus
	err := types.Decode(receipt.KV[0].Value, &titleHeight)
	assert.Nil(suite.T(), err, "decode titleHeight failed")
	suite.T().Log("titleHeight", titleHeight)
	assert.Equal(suite.T(), int32(pt.TyLogParacrossCommit), receipt.Logs[0].Ty)
	assert.Equal(suite.T(), int32(pt.ParacrossStatusCommiting), titleHeight.Status)
	assert.Equal(suite.T(), Title, titleHeight.Title)
	assert.Equal(suite.T(), commitCnt, len(titleHeight.Details.Addrs))
	if commitSupervisionCnt > 0 {
		assert.Equal(suite.T(), commitSupervisionCnt, len(titleHeight.SupervisionDetails.Addrs))
	}
}

func checkDoneReceipt(suite suite.Suite, receipt *types.Receipt, commitCnt int) {
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 8)
	assert.Len(suite.T(), receipt.Logs, 8)

	key := calcTitleHeightKey(Title, TitleHeight)
	suite.T().Log("title height key", string(key))
	assert.Equal(suite.T(), key, receipt.KV[0].Key,
		"receipt not match", string(key), string(receipt.KV[0].Key))

	var titleHeight pt.ParacrossHeightStatus
	err := types.Decode(receipt.KV[0].Value, &titleHeight)
	assert.Nil(suite.T(), err, "decode titleHeight failed")
	suite.T().Log("titleHeight", titleHeight)
	assert.Equal(suite.T(), int32(pt.TyLogParacrossCommit), receipt.Logs[0].Ty)
	assert.Equal(suite.T(), int32(pt.ParacrossStatusCommiting), titleHeight.Status)
	assert.Equal(suite.T(), Title, titleHeight.Title)
	assert.Equal(suite.T(), commitCnt, len(titleHeight.Details.Addrs))

	keyTitle := calcTitleKey(Title)
	suite.T().Log("title key", string(keyTitle), "receipt key", len(receipt.KV))
	assert.Equal(suite.T(), keyTitle, receipt.KV[1].Key,
		"receipt not match", string(keyTitle), string(receipt.KV[1].Key))

	var titleStat pt.ParacrossStatus
	err = types.Decode(receipt.KV[1].Value, &titleStat)
	assert.Nil(suite.T(), err, "decode title failed")
	suite.T().Log("title", titleStat)
	assert.Equal(suite.T(), int32(pt.TyLogParacrossCommitDone), receipt.Logs[1].Ty)
	assert.Equal(suite.T(), TitleHeight, titleStat.Height)
	assert.Equal(suite.T(), Title, titleStat.Title)
	assert.Equal(suite.T(), CurBlock, titleStat.BlockHash)
}

func checkRecordReceipt(suite *CommitTestSuite, receipt *types.Receipt, commitCnt int) {
	_ = commitCnt
	assert.Equal(suite.T(), receipt.Ty, int32(types.ExecOk))
	assert.Len(suite.T(), receipt.KV, 0)
	assert.Len(suite.T(), receipt.Logs, 1)

	var record pt.ReceiptParacrossRecord
	err := types.Decode(receipt.Logs[0].Log, &record)
	assert.Nil(suite.T(), err)
	suite.T().Log("record", record)
	assert.Equal(suite.T(), int32(pt.TyLogParacrossCommitRecord), receipt.Logs[0].Ty)
	assert.Equal(suite.T(), Title, record.Status.Title)
	assert.Equal(suite.T(), TitleHeight, record.Status.Height)
	assert.Equal(suite.T(), CurBlock, record.Status.BlockHash)
}

func (suite *CommitTestSuite) TestExec() {
	receipt := commitOnce(suite, PrivKeyA)
	checkCommitReceipt(suite, receipt, 1, 0)

	receipt = commitOnce(suite, PrivKeyA)
	checkCommitReceipt(suite, receipt, 1, 0)

	receipt = commitOnce(suite, PrivKeyB)
	checkCommitReceipt(suite, receipt, 2, 0)

	receipt = commitOnce(suite, PrivKeyA)
	checkCommitReceipt(suite, receipt, 2, 0)

	receipt = commitOnce(suite, PrivKeyE)
	checkCommitReceipt(suite, receipt, 2, 1)

	receipt = commitOnce(suite, PrivKeyF)
	checkCommitReceipt(suite, receipt, 2, 2)

	receipt = commitOnce(suite, PrivKeyC)
	checkDoneReceipt(suite.Suite, receipt, 3)

	receipt = commitOnce(suite, PrivKeyC)
	checkRecordReceipt(suite, receipt, 3)

	receipt = commitOnce(suite, PrivKeyD)
	checkRecordReceipt(suite, receipt, 4)
}

func TestCommitSuite(t *testing.T) {
	suite.Run(t, new(CommitTestSuite))
}

func TestGetTitle(t *testing.T) {
	exec := "p.user.guodun.token"
	titleExpect := []byte("p.user.guodun.")
	title, err := getTitleFrom([]byte(exec))
	if err != nil {
		t.Error("getTitleFrom", "failed", err)
		return
	}
	assert.Equal(t, titleExpect, title)
}

type VoteTestSuite struct {
	suite.Suite
	stateDB dbm.KV
	localDB *dbmock.KVDB

	exec *Paracross
}

func (s *VoteTestSuite) SetupSuite() {
	//para_init(Title)
	s.exec = newParacross().(*Paracross)
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(chain33TestCfg, nil)
	s.exec.SetAPI(api)

	s.stateDB, _ = dbm.NewGoMemDB("state", "state", 1024)
	// memdb 不支持KVDB接口， 等测试完Exec ， 再扩展 memdb
	//s.localDB, _ = dbm.NewGoMemDB("local", "local", 1024)
	s.localDB = new(dbmock.KVDB)

	s.exec.SetLocalDB(s.localDB)
	s.exec.SetStateDB(s.stateDB)
	s.exec.SetEnv(0, 0, 0)

	stageKey := calcParaSelfConsStagesKey()
	stage := &pt.SelfConsensStage{StartHeight: 0, Enable: pt.ParaConfigYes}
	stages := &pt.SelfConsensStages{Items: []*pt.SelfConsensStage{stage}}
	_ = s.stateDB.Set(stageKey, types.Encode(stages))
	value, err := s.stateDB.Get(stageKey)
	if err != nil {
		s.T().Error("get setup stages failed", err)
		return
	}
	assert.Equal(s.T(), value, types.Encode(stages))

}

func (s *VoteTestSuite) TestVoteTx() {
	status := &pt.ParacrossNodeStatus{
		MainBlockHash:   MainBlockHash10,
		MainBlockHeight: 0,
		PreBlockHash:    PerBlock,
		Height:          CurHeight,
		Title:           Title,
	}
	tx, err := s.createVoteTx(status, PrivKeyA)
	s.Nil(err)
	tx1, err := createAssetTransferTx(s.Suite, PrivKeyA, nil)
	s.Nil(err)
	tx2, err := createAssetTransferTx(s.Suite, PrivKeyB, nil)
	s.Nil(err)
	tx3, err := createParaNormalTx(s.Suite, PrivKeyB, nil)
	s.Nil(err)
	tx4, err := createCrossParaTx(s.Suite, []byte("toB"))
	s.Nil(err)
	tx34 := []*types.Transaction{tx3, tx4}
	txGroup34, err := createTxsGroup(s.Suite, tx34)
	s.Nil(err)

	tx5, err := createCrossParaTx(s.Suite, nil)
	s.Nil(err)
	tx6, err := createCrossParaTx(s.Suite, nil)
	s.Nil(err)
	tx56 := []*types.Transaction{tx5, tx6}
	txGroup56, err := createTxsGroup(s.Suite, tx56)
	s.Nil(err)

	tx7, err := createAssetTransferTx(s.Suite, PrivKeyC, nil)
	s.Nil(err)
	tx8, err := createCrossCommitTx(s.Suite)
	s.Nil(err)

	txs := []*types.Transaction{tx, tx1, tx2}
	txs = append(txs, txGroup34...)
	txs = append(txs, txGroup56...)
	txs = append(txs, tx7)
	txs = append(txs, tx8)
	s.exec.SetTxs(txs)

	//for i,tx := range txs{
	//	s.T().Log("tx exec name","i",i,"name",string(tx.Execer))
	//}

	receipt0, err := s.exec.Exec(tx, 0)
	s.Nil(err)
	recpt0 := &types.ReceiptData{Ty: receipt0.Ty, Logs: receipt0.Logs}
	recpt1 := &types.ReceiptData{Ty: types.ExecOk}
	recpt2 := &types.ReceiptData{Ty: types.ExecErr}
	recpt3 := &types.ReceiptData{Ty: types.ExecOk}
	recpt4 := &types.ReceiptData{Ty: types.ExecOk}
	recpt5 := &types.ReceiptData{Ty: types.ExecPack}
	recpt6 := &types.ReceiptData{Ty: types.ExecPack}
	recpt7 := &types.ReceiptData{Ty: types.ExecOk}
	recpt8 := &types.ReceiptData{Ty: types.ExecOk}
	receipts := []*types.ReceiptData{recpt0, recpt1, recpt2, recpt3, recpt4, recpt5, recpt6, recpt7, recpt8}
	s.exec.SetReceipt(receipts)
	set, err := s.exec.ExecLocal(tx, recpt0, 0)
	s.Nil(err)
	key := pt.CalcMinerHeightKey(status.Title, status.Height)
	for _, kv := range set.KV {
		//s.T().Log(string(kv.GetKey()))
		if bytes.Equal(key, kv.Key) {
			var rst pt.ParacrossNodeStatus
			_ = types.Decode(kv.GetValue(), &rst)
			s.Equal([]byte{0x4d}, rst.TxResult)
			s.Equal([]byte{0x25}, rst.CrossTxResult)
			s.Equal(7, len(rst.TxHashs))
			s.Equal(6, len(rst.CrossTxHashs))
			break
		}
	}
}

func (s *VoteTestSuite) TestVoteTxFork() {
	status := &pt.ParacrossNodeStatus{
		MainBlockHash:   MainBlockHash10,
		MainBlockHeight: MainBlockHeight,
		PreBlockHash:    PerBlock,
		Height:          CurHeight,
		Title:           Title,
	}

	tx1, err := createAssetTransferTx(s.Suite, PrivKeyA, nil)
	s.Nil(err)
	tx2, err := createParaNormalTx(s.Suite, PrivKeyB, nil)
	s.Nil(err)
	tx3, err := createParaNormalTx(s.Suite, PrivKeyA, []byte("toA"))
	s.Nil(err)
	tx4, err := createCrossParaTx(s.Suite, []byte("toB"))
	s.Nil(err)
	tx34 := []*types.Transaction{tx3, tx4}
	txGroup34, err := createTxsGroup(s.Suite, tx34)
	s.Nil(err)

	tx5, err := createCrossParaTx(s.Suite, nil)
	s.Nil(err)
	tx6, err := createCrossParaTx(s.Suite, nil)
	s.Nil(err)
	tx56 := []*types.Transaction{tx5, tx6}
	txGroup56, err := createTxsGroup(s.Suite, tx56)
	s.Nil(err)

	tx7, err := createAssetTransferTx(s.Suite, PrivKeyC, nil)
	s.Nil(err)

	tx8, err := createAssetTransferTx(s.Suite, PrivKeyA, nil)
	s.Nil(err)

	txs := []*types.Transaction{tx1, tx2}
	txs = append(txs, txGroup34...)
	txs = append(txs, txGroup56...)
	txs = append(txs, tx7)
	txs = append(txs, tx8)
	for _, tx := range txs {
		status.TxHashs = append(status.TxHashs, tx.Hash())
	}
	txHashs := FilterParaCrossTxHashes(txs)
	status.CrossTxHashs = append(status.CrossTxHashs, txHashs...)

	baseCheckTxHash := CalcTxHashsHash(status.TxHashs)
	baseCrossTxHash := CalcTxHashsHash(status.CrossTxHashs)

	tx, err := s.createVoteTx(status, PrivKeyA)
	s.Nil(err)

	txs2 := []*types.Transaction{tx}
	txs2 = append(txs2, txs...)

	s.exec.SetTxs(txs2)

	//for i,tx := range txs{
	//	s.T().Log("tx exec name","i",i,"name",string(tx.Execer))
	//}

	//types.S("config.consensus.sub.para.MainForkParacrossCommitTx", int64(1))
	//val,_:=types.G("config.consensus.sub.para.MainForkParacrossCommitTx")

	errlog := &types.ReceiptLog{Ty: types.TyLogErr, Log: []byte("")}
	feelog := &types.Receipt{}
	feelog.Logs = append(feelog.Logs, errlog)
	receipt0, err := s.exec.Exec(tx, 0)
	s.Nil(err)
	recpt0 := &types.ReceiptData{Ty: receipt0.Ty, Logs: receipt0.Logs}
	recpt1 := &types.ReceiptData{Ty: types.ExecErr}
	recpt2 := &types.ReceiptData{Ty: types.ExecOk}
	recpt3 := &types.ReceiptData{Ty: types.ExecOk}
	recpt4 := &types.ReceiptData{Ty: types.ExecOk}
	recpt5 := &types.ReceiptData{Ty: types.ExecPack, Logs: feelog.Logs}
	recpt6 := &types.ReceiptData{Ty: types.ExecPack}
	recpt7 := &types.ReceiptData{Ty: types.ExecPack, Logs: feelog.Logs}
	recpt8 := &types.ReceiptData{Ty: types.ExecOk}
	receipts := []*types.ReceiptData{recpt0, recpt1, recpt2, recpt3, recpt4, recpt5, recpt6, recpt7, recpt8}
	s.exec.SetReceipt(receipts)
	set, err := s.exec.ExecLocal(tx, recpt0, 0)
	s.Nil(err)
	key := pt.CalcMinerHeightKey(status.Title, status.Height)
	for _, kv := range set.KV {
		//s.T().Log(string(kv.GetKey()))
		if bytes.Equal(key, kv.Key) {
			var rst pt.ParacrossNodeStatus
			_ = types.Decode(kv.GetValue(), &rst)
			s.Equal([]byte("8e"), rst.TxResult)
			s.Equal([]byte("22"), rst.CrossTxResult)
			s.Equal(1, len(rst.TxHashs))
			s.Equal(1, len(rst.CrossTxHashs))

			s.Equal(baseCheckTxHash, rst.TxHashs[0])
			s.Equal(baseCrossTxHash, rst.CrossTxHashs[0])
			break
		}
	}
}

func (s *VoteTestSuite) createVoteTx(status *pt.ParacrossNodeStatus, privFrom string) (*types.Transaction, error) {
	tx, err := pt.CreateRawMinerTx(chain33TestCfg, &pt.ParacrossMinerAction{Status: status})
	assert.Nil(s.T(), err, "create asset transfer failed")
	if err != nil {
		return nil, err
	}

	tx, err = signTx(s.Suite, tx, privFrom)
	assert.Nil(s.T(), err, "sign asset transfer failed")
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func createCrossParaTx(s suite.Suite, to []byte) (*types.Transaction, error) {
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

	//tx, err = signTx(s, tx, privFrom)
	//assert.Nil(s.T(), err, "sign asset transfer failed")
	//if err != nil {
	//	return nil, err
	//}

	return tx, nil
}

func createCrossCommitTx(s suite.Suite) (*types.Transaction, error) {
	status := &pt.ParacrossNodeStatus{MainBlockHash: []byte("hash"), MainBlockHeight: 0, Title: Title}
	act := &pt.ParacrossCommitAction{Status: status}
	tx, err := pt.CreateRawCommitTx4MainChain(chain33TestCfg, act, Title+pt.ParaX, 0)
	assert.Nil(s.T(), err, "create asset transfer failed")
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func createTxsGroup(s suite.Suite, txs []*types.Transaction) ([]*types.Transaction, error) {
	group, err := types.CreateTxGroup(txs, chain33TestCfg.GetMinTxFeeRate())
	if err != nil {
		return nil, err
	}
	err = group.Check(chain33TestCfg, 0, chain33TestCfg.GetMinTxFeeRate(), chain33TestCfg.GetMaxTxFee())
	if err != nil {
		return nil, err
	}
	privKey, _ := getPrivKey(s, PrivKeyA)
	for i := range group.Txs {
		_ = group.SignN(i, int32(types.SECP256K1), privKey)
	}
	return group.Txs, nil
}

func TestVoteSuite(t *testing.T) {
	suite.Run(t, new(VoteTestSuite))
}

func createParaNormalTx(s suite.Suite, privFrom string, to []byte) (*types.Transaction, error) {
	param := types.CreateTx{
		To:          string(to),
		Amount:      Amount,
		Fee:         0,
		Note:        []byte("token"),
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: "",
		ExecName:    Title + "token",
	}
	tx := &types.Transaction{
		Execer:  []byte(param.GetExecName()),
		Payload: []byte{},
		To:      address.ExecAddress(param.GetExecName()),
		Fee:     param.Fee,
	}
	tx, err := types.FormatTx(chain33TestCfg, param.GetExecName(), tx)
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

func TestUpdateCommitBlockHashs(t *testing.T) {
	stat := &pt.ParacrossHeightStatus{}
	stat.BlockDetails = &pt.ParacrossStatusBlockDetails{}
	commit := &pt.ParacrossNodeStatus{
		MainBlockHash:   []byte("main"),
		MainBlockHeight: 1,
		BlockHash:       []byte("1122"),
		StateHash:       []byte("statehash"),
		TxResult:        []byte(""),
		TxHashs:         [][]byte{nil},
		CrossTxResult:   []byte(""),
		CrossTxHashs:    [][]byte{nil},
	}

	updateCommitBlockHashs(stat, commit)
	assert.Equal(t, 1, len(stat.BlockDetails.BlockHashs))
	assert.Equal(t, commit.BlockHash, stat.BlockDetails.BlockHashs[0])

	updateCommitBlockHashs(stat, commit)
	assert.Equal(t, 1, len(stat.BlockDetails.BlockHashs))
	assert.Equal(t, commit.BlockHash, stat.BlockDetails.BlockHashs[0])

	commit2 := &pt.ParacrossNodeStatus{
		MainBlockHash:   []byte("main"),
		MainBlockHeight: 1,
		BlockHash:       []byte("2233"),
		StateHash:       []byte("statehash"),
		TxResult:        []byte("11"),
		TxHashs:         [][]byte{[]byte("hash2")},
		CrossTxResult:   []byte("11"),
		CrossTxHashs:    [][]byte{[]byte("hash2")},
	}
	updateCommitBlockHashs(stat, commit2)
	assert.Equal(t, 2, len(stat.BlockDetails.BlockHashs))
	assert.Equal(t, commit2.BlockHash, stat.BlockDetails.BlockHashs[1])

}

func TestValidParaCrossExec(t *testing.T) {
	exec := []byte("paracross")
	valid := types.IsParaExecName(string(exec))
	assert.Equal(t, false, valid)

	exec = []byte("user.p.para.paracross")
	valid = types.IsParaExecName(string(exec))
	assert.Equal(t, true, valid)
}

func TestVerifyBlsSign(t *testing.T) {
	cryptoCli, err := crypto.Load("bls", -1)
	assert.NoError(t, err)

	status := &pt.ParacrossNodeStatus{}
	status.Height = 0
	status.Title = "user.p.para."
	msg := types.Encode(status)
	blsInfo := &pt.ParacrossCommitBlsInfo{}
	commit := &pt.ParacrossCommitAction{Status: status, Bls: blsInfo}

	priKSStr := "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
	p, err := common.FromHex(priKSStr)
	assert.NoError(t, err)
	priKS, err := cryptoCli.PrivKeyFromBytes(p)
	assert.NoError(t, err)

	priJRStr := "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
	p, err = common.FromHex(priJRStr)
	assert.NoError(t, err)
	priJR, err := cryptoCli.PrivKeyFromBytes(p)
	assert.NoError(t, err)

	signKs := priKS.Sign(msg)
	signJr := priJR.Sign(msg)
	pubKs := priKS.PubKey()
	pubJr := priJR.PubKey()

	agg, err := crypto.ToAggregate(cryptoCli)
	assert.NoError(t, err)
	aggSigns, err := agg.Aggregate([]crypto.Signature{signKs, signJr})
	assert.NoError(t, err)
	pubs := []crypto.PubKey{pubKs, pubJr}

	err = agg.VerifyAggregatedOne(pubs, msg, aggSigns)
	assert.NoError(t, err)

	blsInfo.Sign = aggSigns.Bytes()
	PubKS := "a3d97d4186c80268fe6d3689dd574599e25df2dffdcff03f7d8ef64a3bd483241b7d0985958990de2d373d5604caf805"
	PubJR := "81307df1fdde8f0e846ed1542c859c1e9daba2553e62e48db0877329c5c63fb86e70b9e2e83263da0eb7fcad275857f8"
	pubKeys := []string{PubJR, PubKS}
	err = verifyBlsSign(cryptoCli, pubKeys, commit)
	assert.Equal(t, nil, err)

	blsInfo.Sign = signKs.Bytes()
	pubKeys = []string{PubKS}
	err = verifyBlsSign(cryptoCli, pubKeys, commit)
	assert.Equal(t, nil, err)

}

func TestCheckIsIgnoreHeight(t *testing.T) {
	status1 := &pt.ParacrossNodeStatus{
		Title:  "user.p.MC.",
		Height: 1000,
	}
	strList := []string{"mcc.hit.260", "mc.hit.7.9.250", "mc.ignore.1-100.200-300"}
	isIn, err := checkIsIgnoreHeight(strList, status1)
	assert.Nil(t, err)
	assert.Equal(t, false, isIn)

	status1.Height = 9
	isIn, err = checkIsIgnoreHeight(strList, status1)
	assert.Nil(t, err)
	assert.Equal(t, false, isIn)
	status1.Height = 250
	isIn, err = checkIsIgnoreHeight(strList, status1)
	assert.Nil(t, err)
	assert.Equal(t, false, isIn)

	status1.Height = 1
	isIn, err = checkIsIgnoreHeight(strList, status1)
	assert.Nil(t, err)
	assert.Equal(t, true, isIn)

	status1.Height = 10
	isIn, err = checkIsIgnoreHeight(strList, status1)
	assert.Nil(t, err)
	assert.Equal(t, true, isIn)

	status1.Height = 251
	isIn, err = checkIsIgnoreHeight(strList, status1)
	assert.Nil(t, err)
	assert.Equal(t, true, isIn)

	//mc和mcc不能混淆
	status1.Height = 260
	isIn, err = checkIsIgnoreHeight(strList, status1)
	assert.Nil(t, err)
	assert.Equal(t, true, isIn)

}
