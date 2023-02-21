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
	Amount = 1 * types.DefaultCoinPrecision
	Title  = string("user.p.para.")
	Title2 = string("user.p.test.")
)

func TestFilterTxsForPara(t *testing.T) {
	cfg := types.NewChain33Config(types.ReadFile("../../../plugin/dapp/paracross/cmd/build/chain33.para.test.toml"))

	detail, filterTxs, _ := createTestTxs(cfg, t)
	rst := paraexec.FilterTxsForPara(cfg, detail.FilterParaTxsByTitle(cfg, Title))

	assert.Equal(t, filterTxs, rst)

}

func createCrossParaTx(cfg *types.Chain33Config, to string, amount int64) (*types.Transaction, error) {
	param := types.CreateTx{
		To:          to,
		Amount:      amount,
		Fee:         0,
		Note:        []byte("test asset transfer"),
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: "",
		ExecName:    cfg.ExecName(pt.ParaX),
	}
	tx, err := pt.CreateRawAssetTransferTx(cfg, &param)

	return tx, err
}

func createCrossParaTempTx(cfg *types.Chain33Config, to string, amount int64) (*types.Transaction, error) {
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
	tx, err := pt.CreateRawAssetTransferTx(cfg, &param)

	return tx, err
}

func createTxsGroup(cfg *types.Chain33Config, txs []*types.Transaction) ([]*types.Transaction, error) {

	group, err := types.CreateTxGroup(txs, cfg.GetMinTxFeeRate())
	if err != nil {
		return nil, err
	}
	err = group.Check(cfg, 0, cfg.GetMinTxFeeRate(), cfg.GetMaxTxFee())
	if err != nil {
		return nil, err
	}
	return group.Txs, nil
}

func createTestTxs(cfg *types.Chain33Config, t *testing.T) (*types.BlockDetail, []*types.Transaction, []*types.Transaction) {
	//all para tx group
	tx5, err := createCrossParaTx(cfg, "toB", 5)
	assert.Nil(t, err)
	tx6, err := createCrossParaTx(cfg, "toB", 6)
	assert.Nil(t, err)
	tx56 := []*types.Transaction{tx5, tx6}
	txGroup56, err := createTxsGroup(cfg, tx56)
	assert.Nil(t, err)

	//para cross tx group fail
	tx7, _ := createCrossParaTx(cfg, "toA", 1)
	tx8, err := createCrossParaTx(cfg, "toB", 8)
	assert.Nil(t, err)
	tx78 := []*types.Transaction{tx7, tx8}
	txGroup78, err := createTxsGroup(cfg, tx78)
	assert.Nil(t, err)

	//all para tx group
	txB, err := createCrossParaTx(cfg, "toB", 11)
	assert.Nil(t, err)
	txC, err := createCrossParaTx(cfg, "toB", 12)
	assert.Nil(t, err)
	txBC := []*types.Transaction{txB, txC}
	txGroupBC, err := createTxsGroup(cfg, txBC)
	assert.Nil(t, err)

	//single para tx
	txD, err := createCrossParaTempTx(cfg, "toB", 10)
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
	cfg := types.NewChain33Config(types.ReadFile("../../../plugin/dapp/paracross/cmd/build/chain33.para.test.toml"))
	pk, err := hex.DecodeString(minerPrivateKey)
	assert.Nil(t, err)

	secp, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	assert.Nil(t, err)

	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)

	block := &types.Block{}

	_, filterTxs, _ := createTestTxs(cfg, t)
	localBlock := &pt.ParaLocalDbBlock{
		Height:     1,
		MainHeight: 10,
		MainHash:   []byte("mainhash"),
		Txs:        filterTxs}

	api := new(apimocks.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	para := &client{BaseClient: &drivers.BaseClient{}}
	para.SetAPI(api)
	para.subCfg = new(subConfig)
	para.minerPrivateKey = priKey
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
	qClient.On("GetConfig").Return(&types.Chain33Config{})
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
	int1 := &emptyBlockInterval{startHeight: 0, interval: 1}
	int2 := &emptyBlockInterval{startHeight: 10, interval: 10}
	int3 := &emptyBlockInterval{startHeight: 15, interval: 15}

	ints := []*emptyBlockInterval{int1, int2, int3}
	para := new(client)
	para.dldCfg = &downloadClient{}
	para.dldCfg.emptyInterval = append(para.dldCfg.emptyInterval, ints...)

	lastBlock := &pt.ParaLocalDbBlock{Height: 1}
	ret := para.getEmptyInterval(lastBlock)
	assert.Equal(t, int1.interval, ret)

	lastBlock = &pt.ParaLocalDbBlock{Height: 10}
	ret = para.getEmptyInterval(lastBlock)
	assert.Equal(t, int2.interval, ret)

	lastBlock = &pt.ParaLocalDbBlock{Height: 11}
	ret = para.getEmptyInterval(lastBlock)
	assert.Equal(t, int2.interval, ret)

	lastBlock = &pt.ParaLocalDbBlock{Height: 16}
	ret = para.getEmptyInterval(lastBlock)
	assert.Equal(t, int3.interval, ret)

}

func TestCheckEmptyInterval(t *testing.T) {
	int1 := &emptyBlockInterval{startHeight: 0, interval: 1}
	int2 := &emptyBlockInterval{startHeight: 10, interval: 10}
	int3 := &emptyBlockInterval{startHeight: 15, interval: 15}

	int1.startHeight = 5
	ints := []*emptyBlockInterval{int1, int2, int3}
	err := checkEmptyBlockInterval(ints)
	assert.Equal(t, types.ErrInvalidParam, err)
	int1.startHeight = 0

	int3.startHeight = 5
	ints = []*emptyBlockInterval{int1, int2, int3}
	err = checkEmptyBlockInterval(ints)
	assert.Equal(t, types.ErrInvalidParam, err)

	int3.startHeight = 10
	ints = []*emptyBlockInterval{int1, int2, int3}
	err = checkEmptyBlockInterval(ints)
	assert.Equal(t, types.ErrInvalidParam, err)
	int3.startHeight = 15

	int2.interval = 0
	ints = []*emptyBlockInterval{int1, int2, int3}
	err = checkEmptyBlockInterval(ints)
	assert.Equal(t, types.ErrInvalidParam, err)

	int2.interval = 2
	ints = []*emptyBlockInterval{int1, int2, int3}
	err = checkEmptyBlockInterval(ints)
	assert.Equal(t, nil, err)

}

func TestParseEmptyBlockInterval(t *testing.T) {
	cfg := []string{}
	ret, err := parseEmptyBlockInterval(cfg)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), ret[0].startHeight)
	assert.Equal(t, defaultEmptyBlockInterval, ret[0].interval)

	cfg = []string{"0:50"}
	ret, err = parseEmptyBlockInterval(cfg)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), ret[0].startHeight)
	assert.Equal(t, defaultEmptyBlockInterval, ret[0].interval)

	cfg = []string{"0:50", "100:20"}
	ret, err = parseEmptyBlockInterval(cfg)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), ret[0].startHeight)
	assert.Equal(t, defaultEmptyBlockInterval, ret[0].interval)
	assert.Equal(t, int64(100), ret[1].startHeight)
	assert.Equal(t, int64(20), ret[1].interval)

	cfg = []string{"10:50"}
	ret, err = parseEmptyBlockInterval(cfg)
	assert.Nil(t, err)
	assert.Equal(t, int64(10), ret[0].startHeight)
	assert.Equal(t, defaultEmptyBlockInterval, ret[0].interval)

	cfg = []string{"10:50", "20-30"}
	_, err = parseEmptyBlockInterval(cfg)
	assert.NotNil(t, err)

	cfg = []string{"10:50", "20:"}
	_, err = parseEmptyBlockInterval(cfg)
	assert.NotNil(t, err)

	cfg = []string{"10:50", ":20"}
	_, err = parseEmptyBlockInterval(cfg)
	assert.NotNil(t, err)

	//mess sequence
	cfg = []string{"100:30", "0:50", "30:20", "200:10"}
	ret, err = parseEmptyBlockInterval(cfg)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), ret[0].startHeight)
	assert.Equal(t, defaultEmptyBlockInterval, ret[0].interval)
	assert.Equal(t, int64(30), ret[1].startHeight)
	assert.Equal(t, int64(20), ret[1].interval)
	assert.Equal(t, int64(100), ret[2].startHeight)
	assert.Equal(t, int64(30), ret[2].interval)
	assert.Equal(t, int64(200), ret[3].startHeight)
	assert.Equal(t, int64(10), ret[3].interval)
}
