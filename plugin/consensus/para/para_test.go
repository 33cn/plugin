// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"encoding/hex"

	"github.com/stretchr/testify/assert"

	"testing"

	apimocks "github.com/33cn/chain33/client/mocks"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/queue"
	qmocks "github.com/33cn/chain33/queue/mocks"
	drivers "github.com/33cn/chain33/system/consensus"
	"github.com/33cn/chain33/types"
	typesmocks "github.com/33cn/chain33/types/mocks"
	paraexec "github.com/33cn/plugin/plugin/dapp/paracross/executor"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/stretchr/testify/mock"
)

var (
	Amount = 1 * types.Coin
	Title  = string("user.p.para.")
	Title2 = string("user.p.test.")
)

func TestFilterTxsForPara(t *testing.T) {
	cfg, _ := types.InitCfg("../../../plugin/dapp/paracross/cmd/build/chain33.para.test.toml")
	types.Init(Title, cfg)

	detail, filterTxs, _ := createTestTxs(t)
	rst := paraexec.FilterTxsForPara(detail.FilterParaTxsByTitle(Title))

	assert.Equal(t, filterTxs, rst)

}

func createCrossParaTx(to string, amount int64) (*types.Transaction, error) {
	param := types.CreateTx{
		To:          to,
		Amount:      amount,
		Fee:         0,
		Note:        []byte("test asset transfer"),
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: "",
		ExecName:    types.ExecName(pt.ParaX),
	}
	tx, err := pt.CreateRawAssetTransferTx(&param)

	return tx, err
}

func createCrossParaTempTx(to string, amount int64) (*types.Transaction, error) {
	param := types.CreateTx{
		To:          to,
		Amount:      amount,
		Fee:         0,
		Note:        []byte("test asset transfer"),
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: "",
		ExecName:    Title2 + pt.ParaX,
	}
	tx, err := pt.CreateRawAssetTransferTx(&param)

	return tx, err
}

func createTxsGroup(txs []*types.Transaction) ([]*types.Transaction, error) {

	group, err := types.CreateTxGroup(txs, types.GInt("MinFee"))
	if err != nil {
		return nil, err
	}
	err = group.Check(0, types.GInt("MinFee"), types.GInt("MaxFee"))
	if err != nil {
		return nil, err
	}
	return group.Txs, nil
}

func createTestTxs(t *testing.T) (*types.BlockDetail, []*types.Transaction, []*types.Transaction) {
	//all para tx group
	tx5, err := createCrossParaTx("toB", 5)
	assert.Nil(t, err)
	tx6, err := createCrossParaTx("toB", 6)
	assert.Nil(t, err)
	tx56 := []*types.Transaction{tx5, tx6}
	txGroup56, err := createTxsGroup(tx56)
	assert.Nil(t, err)

	//para cross tx group fail
	tx7, _ := createCrossParaTx("toA", 1)
	tx8, err := createCrossParaTx("toB", 8)
	assert.Nil(t, err)
	tx78 := []*types.Transaction{tx7, tx8}
	txGroup78, err := createTxsGroup(tx78)
	assert.Nil(t, err)

	//all para tx group
	txB, err := createCrossParaTx("toB", 11)
	assert.Nil(t, err)
	txC, err := createCrossParaTx("toB", 12)
	assert.Nil(t, err)
	txBC := []*types.Transaction{txB, txC}
	txGroupBC, err := createTxsGroup(txBC)
	assert.Nil(t, err)

	//single para tx
	txD, err := createCrossParaTempTx("toB", 10)
	assert.Nil(t, err)

	txs := []*types.Transaction{}
	txs = append(txs, txGroup56...)
	txs = append(txs, txGroup78...)
	txs = append(txs, txGroupBC...)
	txs = append(txs, txD)

	//for i, tx := range txs {
	//	t.Log("tx exec name", "i", i, "name", string(tx.Execer))
	//}

	recpt5 := &types.ReceiptData{Ty: types.ExecOk}
	recpt6 := &types.ReceiptData{Ty: types.ExecOk}

	log7 := &types.ReceiptLog{Ty: types.TyLogErr}
	logs := []*types.ReceiptLog{log7}
	recpt7 := &types.ReceiptData{Ty: types.ExecPack, Logs: logs}
	recpt8 := &types.ReceiptData{Ty: types.ExecPack}

	recptB := &types.ReceiptData{Ty: types.ExecPack}
	recptC := &types.ReceiptData{Ty: types.ExecPack}
	recptD := &types.ReceiptData{Ty: types.ExecPack}
	receipts := []*types.ReceiptData{recpt5, recpt6, recpt7, recpt8, recptB, recptC, recptD}

	block := &types.Block{Height: 10, Txs: txs}
	detail := &types.BlockDetail{
		Block:    block,
		Receipts: receipts,
	}

	filterTxs := []*types.Transaction{tx5, tx6, txB, txC}
	return detail, filterTxs, txs

}

func TestAddMinerTx(t *testing.T) {
	pk, err := hex.DecodeString(minerPrivateKey)
	assert.Nil(t, err)

	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	assert.Nil(t, err)

	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)

	block := &types.Block{}

	_, filterTxs, _ := createTestTxs(t)
	localBlock := &pt.ParaLocalDbBlock{
		Height:     1,
		MainHeight: 10,
		MainHash:   []byte("mainhash"),
		Txs:        filterTxs}
	para := new(client)
	para.subCfg = new(subConfig)
	para.subCfg.SelfConsensusEnable = append(para.subCfg.SelfConsensusEnable, &paraSelfConsEnable{Enable: false})
	para.privateKey = priKey
	para.commitMsgClient = new(commitMsgClient)
	para.commitMsgClient.paraClient = para

	para.blockSyncClient = new(blockSyncClient)
	para.blockSyncClient.paraClient = para
	para.blockSyncClient.addMinerTx(nil, block, localBlock)
	assert.Equal(t, 1, len(block.Txs))

}

func initBlock() {
	println("initblock")
}

func getMockLastBlock(para *client, returnBlock *types.Block) {
	baseCli := drivers.NewBaseClient(&types.Consensus{Name: "name"})
	para.BaseClient = baseCli

	qClient := &qmocks.Client{}
	para.InitClient(qClient, initBlock)

	msg := queue.NewMessage(0, "", 1, returnBlock)

	qClient.On("NewMessage", "blockchain", int64(types.EventGetLastBlock), mock.Anything).Return(msg)
	qClient.On("Send", mock.Anything, mock.Anything).Return(nil)

	qClient.On("Wait", mock.Anything).Return(msg, nil)
}

func TestGetLastBlockInfo(t *testing.T) {
	para := new(client)
	grpcClient := &typesmocks.Chain33Client{}
	para.grpcClient = grpcClient

	block := &types.Block{Height: 0}
	getMockLastBlock(para, block)

	api := &apimocks.QueueProtocolAPI{}
	para.SetAPI(api)

	grpcClient.On("GetSequenceByHash", mock.Anything, mock.Anything).Return(&types.Int64{Data: int64(10)}, nil)

	mainSeq, lastBlock, err := para.getLastBlockMainInfo()
	assert.NoError(t, err)
	assert.Equal(t, int64(10), mainSeq)
	assert.Equal(t, lastBlock.Height, block.Height)
}

func TestGetEmptyInterval(t *testing.T) {
	int1 := &emptyBlockInterval{BlockHeight: 0, Interval: 1}
	int2 := &emptyBlockInterval{BlockHeight: 10, Interval: 10}
	int3 := &emptyBlockInterval{BlockHeight: 15, Interval: 15}

	ints := []*emptyBlockInterval{int1, int2, int3}
	para := new(client)
	para.subCfg = &subConfig{EmptyBlockInterval: ints}

	lastBlock := &pt.ParaLocalDbBlock{Height: 1}
	ret := para.getEmptyInterval(lastBlock)
	assert.Equal(t, int1.Interval, ret)

	lastBlock = &pt.ParaLocalDbBlock{Height: 10}
	ret = para.getEmptyInterval(lastBlock)
	assert.Equal(t, int2.Interval, ret)

	lastBlock = &pt.ParaLocalDbBlock{Height: 11}
	ret = para.getEmptyInterval(lastBlock)
	assert.Equal(t, int2.Interval, ret)

	lastBlock = &pt.ParaLocalDbBlock{Height: 16}
	ret = para.getEmptyInterval(lastBlock)
	assert.Equal(t, int3.Interval, ret)

}

func TestCheckEmptyInterval(t *testing.T) {
	int1 := &emptyBlockInterval{BlockHeight: 0, Interval: 1}
	int2 := &emptyBlockInterval{BlockHeight: 10, Interval: 10}
	int3 := &emptyBlockInterval{BlockHeight: 15, Interval: 15}

	int1.BlockHeight = 5
	ints := []*emptyBlockInterval{int1, int2, int3}
	err := checkEmptyBlockInterval(ints)
	assert.Equal(t, types.ErrInvalidParam, err)
	int1.BlockHeight = 0

	int3.BlockHeight = 5
	ints = []*emptyBlockInterval{int1, int2, int3}
	err = checkEmptyBlockInterval(ints)
	assert.Equal(t, types.ErrInvalidParam, err)

	int3.BlockHeight = 10
	ints = []*emptyBlockInterval{int1, int2, int3}
	err = checkEmptyBlockInterval(ints)
	assert.Equal(t, types.ErrInvalidParam, err)
	int3.BlockHeight = 15

	int2.Interval = 0
	ints = []*emptyBlockInterval{int1, int2, int3}
	err = checkEmptyBlockInterval(ints)
	assert.Equal(t, types.ErrInvalidParam, err)

	int2.Interval = 2
	ints = []*emptyBlockInterval{int1, int2, int3}
	err = checkEmptyBlockInterval(ints)
	assert.Equal(t, nil, err)

}

func TestCheckSelfConsensEnable(t *testing.T) {
	int1 := &paraSelfConsEnable{BlockHeight: 0, Enable: false}
	int2 := &paraSelfConsEnable{BlockHeight: 10, Enable: true}
	int3 := &paraSelfConsEnable{BlockHeight: 5, Enable: false}

	ints := []*paraSelfConsEnable{int1, int2, int3}
	err := checkSelfConsensEnable(ints)
	assert.Equal(t, types.ErrInvalidParam, err)

	int3.BlockHeight = 15
	err = checkSelfConsensEnable(ints)
	assert.Nil(t, err)
}

func TestGetSelfConsEnableStatus(t *testing.T) {
	int1 := &paraSelfConsEnable{BlockHeight: 0, Enable: false}
	int2 := &paraSelfConsEnable{BlockHeight: 10, Enable: true}
	int3 := &paraSelfConsEnable{BlockHeight: 15, Enable: false}

	para := new(client)
	para.subCfg = new(subConfig)
	para.subCfg.SelfConsensusEnable = []*paraSelfConsEnable{int1, int2, int3}
	selfConf := para.getSelfConsEnableStatus(5)
	assert.Equal(t, int64(0), selfConf.BlockHeight)
	assert.Equal(t, false, selfConf.Enable)

	selfConf = para.getSelfConsEnableStatus(10)
	assert.Equal(t, int64(10), selfConf.BlockHeight)
	assert.Equal(t, true, selfConf.Enable)

	selfConf = para.getSelfConsEnableStatus(16)
	assert.Equal(t, int64(15), selfConf.BlockHeight)
	assert.Equal(t, false, selfConf.Enable)
}
