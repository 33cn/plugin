// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"errors"
	"testing"

	"github.com/33cn/chain33/blockchain"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/mempool"
	"github.com/33cn/chain33/p2p"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/store"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	typesmocks "github.com/33cn/chain33/types/mocks"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	//types.Init("user.p.para.", nil)
	log.SetLogLevel("error")
}

type suiteParaClient struct {
	// Include our basic suite logic.
	suite.Suite
	para          *client
	grpcCli       *typesmocks.Chain33Client
	q             queue.Queue
	block         *blockchain.BlockChain
	exec          *executor.Executor
	store         queue.Module
	mem           queue.Module
	network       *p2p.P2p
	mainBlockList []*types.Block
}

func (s *suiteParaClient) initEnv(cfg *types.Config, sub *types.ConfigSubModule) {
	s.createTempBlock()
	q := queue.New("channel")
	s.q = q
	//api, _ = client.New(q.Client(), nil)
	var subcfg subConfig
	if sub != nil {
		types.MustDecode(sub.Consensus["para"], &subcfg)
	}
	s.block = blockchain.New(cfg.BlockChain)
	s.block.SetQueueClient(q.Client())

	s.exec = executor.New(cfg.Exec, sub.Exec)
	s.exec.SetQueueClient(q.Client())

	s.store = store.New(cfg.Store, sub.Store)
	s.store.SetQueueClient(q.Client())

	//cfg.Consensus.StartHeight = 0
	//add block by UT below
	subcfg.EmptyBlockInterval = 100

	s.para = New(cfg.Consensus, sub.Consensus["para"]).(*client)
	s.grpcCli = &typesmocks.Chain33Client{}

	s.createBlockMock()

	reply := &types.Reply{IsOk: true}
	s.grpcCli.On("IsSync", mock.Anything, mock.Anything).Return(reply, nil)
	result := &pt.ParacrossStatus{Height: -1}
	data := types.Encode(result)
	ret := &types.Reply{IsOk: true, Msg: data}
	s.grpcCli.On("QueryChain", mock.Anything, mock.Anything).Return(ret, nil).Maybe()
	s.grpcCli.On("SendTransaction", mock.Anything, mock.Anything).Return(reply, nil).Maybe()
	s.grpcCli.On("GetLastHeader", mock.Anything, mock.Anything).Return(&types.Header{Height: subcfg.StartHeight + minBlockNum}, nil).Maybe()
	s.grpcCli.On("GetBlockHash", mock.Anything, mock.Anything).Return(&types.ReplyHash{Hash: []byte("1")}, nil).Maybe()
	s.para.grpcClient = s.grpcCli
	s.para.SetQueueClient(q.Client())

	s.mem = mempool.New(cfg.Mempool, nil)
	s.mem.SetQueueClient(q.Client())
	s.mem.Wait()

	s.network = p2p.New(cfg.P2P)
	s.network.SetQueueClient(q.Client())

	//create block self
	s.createBlock()
}

func (s *suiteParaClient) createTempBlock() {
	var parentHash []byte
	for i := 0; i < 3; i++ {
		block := &types.Block{
			Height:     int64(i),
			ParentHash: parentHash,
		}
		hash := block.HashByForkHeight(1)
		s.mainBlockList = append(s.mainBlockList, block)
		parentHash = hash
	}
}

func (s *suiteParaClient) createBlockMock() {
	var i int64

	for i = 0; i < 3; i++ {
		blockSeq := &types.BlockSeq{
			Seq: &types.BlockSequence{
				Hash: s.mainBlockList[i].HashByForkHeight(1),
				Type: 1,
			},
			Detail: &types.BlockDetail{Block: s.mainBlockList[i]},
		}

		s.grpcCli.On("GetBlockBySeq", mock.Anything, &types.Int64{Data: i}).Return(blockSeq, nil)
	}

	// set block 3's parentHasn not equal, enter switch
	block3 := &types.Block{
		Height:     3,
		ParentHash: []byte(string(1)),
	}
	hash := block3.HashByForkHeight(1)
	blockSeq3 := &types.BlockSeq{
		Seq: &types.BlockSequence{
			Hash: hash,
			Type: 1,
		},
		Detail: &types.BlockDetail{Block: block3},
	}
	s.grpcCli.On("GetBlockBySeq", mock.Anything, &types.Int64{Data: 3}).Return(blockSeq3, nil)

	// RequestTx GetLastSeqOnMainChain
	seq := &types.Int64{Data: 1}
	s.grpcCli.On("GetLastBlockSequence", mock.Anything, mock.Anything).Return(seq, nil).Once()
	seq = &types.Int64{Data: 2}
	s.grpcCli.On("GetLastBlockSequence", mock.Anything, mock.Anything).Return(seq, nil).Once()
	seq = &types.Int64{Data: 3}
	s.grpcCli.On("GetLastBlockSequence", mock.Anything, mock.Anything).Return(seq, nil)

	// mock for switchHashMatchedBlock
	s.grpcCli.On("GetSequenceByHash", mock.Anything, &types.ReqHash{Hash: []byte(string(3))}).Return(nil, errors.New("hash err")).Once()
	s.grpcCli.On("GetSequenceByHash", mock.Anything, &types.ReqHash{Hash: []byte(string(2))}).Return(nil, errors.New("hash err")).Once()

	// mock for removeBlocks
	seq = &types.Int64{Data: 1}
	s.grpcCli.On("GetSequenceByHash", mock.Anything, mock.Anything).Return(seq, nil)
}

func (s *suiteParaClient) createBlock() {
	var i int64
	for i = 0; i < 3; i++ {
		lastBlock, err := s.para.RequestLastBlock()
		if err != nil {
			plog.Error("para test", "err", err.Error())
		}
		plog.Info("para test---------1", "last height", lastBlock.Height)
		s.para.createBlock(lastBlock, nil, i, s.getParaMainBlock(i+1, lastBlock.BlockTime+1))
	}
}

func (s *suiteParaClient) SetupSuite() {
	s.initEnv(types.InitCfg("../../../plugin/dapp/paracross/cmd/build/chain33.para.test.toml"))
}

func testRunSuiteParaClient(t *testing.T) {
	log := new(suiteParaClient)
	suite.Run(t, log)
}

func (s *suiteParaClient) TearDownSuite() {
	//time.Sleep(time.Second * 2)
	s.block.Close()
	s.para.Close()
	s.network.Close()
	s.exec.Close()
	s.store.Close()
	s.mem.Close()
	s.q.Close()

}

func (s *suiteParaClient) getParaMainBlock(height int64, BlockTime int64) *types.BlockSeq {

	return &types.BlockSeq{
		Num: height,
		Seq: &types.BlockSequence{Hash: s.mainBlockList[height-1].HashByForkHeight(1), Type: addAct},
		Detail: &types.BlockDetail{
			Block: &types.Block{
				ParentHash: s.mainBlockList[height-1].ParentHash,
				Height:     height,
				BlockTime:  BlockTime,
			},
		},
	}
}

func getPrivKey(t *testing.T) crypto.PrivKey {
	pk, err := common.FromHex("6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b")
	assert.Nil(t, err)

	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	assert.Nil(t, err)

	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)
	return priKey
}

func TestCalcCommitMsgTxs(t *testing.T) {
	priKey := getPrivKey(t)
	client := commitMsgClient{
		privateKey: priKey,
	}
	nt1 := &pt.ParacrossNodeStatus{
		Height: 1,
		Title:  "user.p.para",
	}
	nt2 := &pt.ParacrossNodeStatus{
		Height: 2,
		Title:  "user.p.para",
	}
	notify := []*pt.ParacrossNodeStatus{nt1}
	tx, count, err := client.calcCommitMsgTxs(notify)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)
	assert.NotNil(t, tx)

	notify = append(notify, nt2)
	tx, count, err = client.calcCommitMsgTxs(notify)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), count)
	assert.NotNil(t, tx)

	tx, err = client.singleCalcTx(nt2)
	assert.Nil(t, err)
	assert.NotNil(t, tx)

}

func TestGetConsensusStatus(t *testing.T) {
	para := new(client)
	grpcClient := &typesmocks.Chain33Client{}
	//grpcClient.On("GetFork", mock.Anything, &types.ReqKey{Key: []byte("ForkBlockHash")}).Return(&types.Int64{Data: 1}, errors.New("err")).Once()
	para.grpcClient = grpcClient
	commitCli := new(commitMsgClient)
	commitCli.paraClient = para

	block := &types.Block{
		Height:     1,
		MainHeight: 10,
	}

	status := &pt.ParacrossStatus{
		Height: 1,
	}
	reply := &types.Reply{
		IsOk: true,
		Msg:  types.Encode(status),
	}
	grpcClient.On("QueryChain", mock.Anything, mock.Anything).Return(reply, nil).Once()
	ret, err := commitCli.getConsensusStatus(block)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), ret.Height)
}
