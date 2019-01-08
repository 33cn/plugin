// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	"github.com/33cn/chain33/blockchain"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/mempool"
	"github.com/33cn/chain33/p2p"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	//"github.com/33cn/plugin/plugin/dapp/paracross/rpc"
	"time"

	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/store"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	typesmocks "github.com/33cn/chain33/types/mocks"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

func init() {
	//types.Init("user.p.para.", nil)
	log.SetLogLevel("debug")
}

type suiteParaClient struct {
	// Include our basic suite logic.
	suite.Suite
	para    *client
	grpcCli *typesmocks.Chain33Client
	q       queue.Queue
	block   *blockchain.BlockChain
	exec    *executor.Executor
	store   queue.Module
	mem     queue.Module
	network *p2p.P2p
}

func (s *suiteParaClient) initEnv(cfg *types.Config, sub *types.ConfigSubModule) {
	q := queue.New("channel")
	s.q = q
	//api, _ = client.New(q.Client(), nil)

	s.block = blockchain.New(cfg.BlockChain)
	s.block.SetQueueClient(q.Client())

	s.exec = executor.New(cfg.Exec, sub.Exec)
	s.exec.SetQueueClient(q.Client())

	s.store = store.New(cfg.Store, sub.Store)
	s.store.SetQueueClient(q.Client())

	//cfg.Consensus.StartHeight = 0
	cfg.Consensus.EmptyBlockInterval = 1
	s.para = New(cfg.Consensus, sub.Consensus["para"]).(*client)
	s.grpcCli = &typesmocks.Chain33Client{}
	blockHash := &types.BlockSequence{
		Hash: []byte("1"),
		Type: 1,
	}
	blockSeqs := &types.BlockSequences{Items: []*types.BlockSequence{blockHash}}

	s.grpcCli.On("GetBlockSequences", mock.Anything, mock.Anything).Return(blockSeqs, nil)
	block := &types.Block{Height: 0}
	blockDetail := &types.BlockDetail{Block: block}
	blockDetails := &types.BlockDetails{Items: []*types.BlockDetail{blockDetail}}
	s.grpcCli.On("GetBlockByHashes", mock.Anything, mock.Anything).Return(blockDetails, nil).Once()
	block = &types.Block{Height: 6, BlockTime: 8888888888}
	blockDetail = &types.BlockDetail{Block: block}
	blockDetails = &types.BlockDetails{Items: []*types.BlockDetail{blockDetail}}
	s.grpcCli.On("GetBlockByHashes", mock.Anything, mock.Anything).Return(blockDetails, nil).Once()
	block = &types.Block{Height: 0}
	blockDetail = &types.BlockDetail{Block: block}
	blockDetails = &types.BlockDetails{Items: []*types.BlockDetail{blockDetail}}
	s.grpcCli.On("GetBlockByHashes", mock.Anything, mock.Anything).Return(blockDetails, nil).Once()
	block = &types.Block{Height: 0}
	blockDetail = &types.BlockDetail{Block: block}
	blockDetails = &types.BlockDetails{Items: []*types.BlockDetail{blockDetail}}
	s.grpcCli.On("GetBlockByHashes", mock.Anything, mock.Anything).Return(blockDetails, nil)

	seq := &types.Int64{Data: 1}
	s.grpcCli.On("GetLastBlockSequence", mock.Anything, mock.Anything).Return(seq, nil).Once()
	seq = &types.Int64{Data: 2}
	s.grpcCli.On("GetLastBlockSequence", mock.Anything, mock.Anything).Return(seq, nil).Once()
	seq = &types.Int64{Data: 3}
	s.grpcCli.On("GetLastBlockSequence", mock.Anything, mock.Anything).Return(seq, nil)

	seq = &types.Int64{Data: 1}
	s.grpcCli.On("GetSequenceByHash", mock.Anything, mock.Anything).Return(seq, nil)

	reply := &types.Reply{IsOk: true}
	s.grpcCli.On("IsSync", mock.Anything, mock.Anything).Return(reply, nil)
	result := &pt.ParacrossStatus{Height: -1}
	data := types.Encode(result)
	ret := &types.Reply{IsOk: true, Msg: data}
	s.grpcCli.On("QueryChain", mock.Anything, mock.Anything).Return(ret, nil).Maybe()
	s.grpcCli.On("SendTransaction", mock.Anything, mock.Anything).Return(reply, nil).Maybe()
	s.grpcCli.On("GetLastHeader", mock.Anything, mock.Anything).Return(&types.Header{Height:cfg.Consensus.StartHeight+minBlockNum}, nil).Maybe()
	s.grpcCli.On("GetBlockHash", mock.Anything, mock.Anything).Return(&types.ReplyHash{Hash:[]byte("1")}, nil).Maybe()
	s.para.grpcClient = s.grpcCli
	s.para.SetQueueClient(q.Client())

	s.mem = mempool.New(cfg.Mempool, nil)
	s.mem.SetQueueClient(q.Client())
	s.mem.Wait()

	s.network = p2p.New(cfg.P2P)
	s.network.SetQueueClient(q.Client())

}

func (s *suiteParaClient) TestRun_Test() {
	//s.testGetBlock()
	lastBlock, err := s.para.RequestLastBlock()
	if err != nil {
		plog.Error("para test", "err", err.Error())
	}
	plog.Info("para test---------1", "last height", lastBlock.Height)
	s.para.createBlock(lastBlock, nil, 0, getMainBlock(2, lastBlock.BlockTime+1))
	lastBlock, err = s.para.RequestLastBlock()
	if err != nil {
		plog.Error("para test--2", "err", err.Error())
	}
	plog.Info("para test---------", "last height", lastBlock.Height)
	s.para.createBlock(lastBlock, nil, 1, getMainBlock(3, lastBlock.BlockTime+1))
	time.Sleep(time.Second * 1)

	s.testRunGetMinerTxInfo()
	s.testRunRmvBlock()

}

func (s *suiteParaClient) testRunGetMinerTxInfo() {
	lastBlock, err := s.para.RequestLastBlock()
	s.Nil(err)
	plog.Info("para test testRunGetMinerTxInfo", "last height", lastBlock.Height)
	s.True(lastBlock.Height > 1)
	status, err := getMinerTxInfo(lastBlock)
	s.Nil(err)
	s.Equal(int64(3), status.MainBlockHeight)

}

func (s *suiteParaClient) testRunRmvBlock() {
	lastBlock, err := s.para.RequestLastBlock()
	s.Nil(err)
	plog.Info("para test testRunGetMinerTxInfo", "last height", lastBlock.Height)
	s.True(lastBlock.Height > 1)
	s.para.removeBlocks(1)

	lastBlock, err = s.para.RequestLastBlock()
	s.Nil(err)
	plog.Info("para test testRunGetMinerTxInfo", "last height", lastBlock.Height)
	s.Equal(int64(1), lastBlock.Height)

}

func (s *suiteParaClient) SetupSuite() {
	s.initEnv(types.InitCfg("../../../plugin/dapp/paracross/cmd/build/chain33.para.test.toml"))
}

func TestRunSuiteParaClient(t *testing.T) {
	log := new(suiteParaClient)
	suite.Run(t, log)
}

func (s *suiteParaClient) TearDownSuite() {
	time.Sleep(time.Second * 5)
	s.block.Close()
	s.para.Close()
	s.network.Close()
	s.exec.Close()
	s.store.Close()
	s.mem.Close()
	s.q.Close()

}
