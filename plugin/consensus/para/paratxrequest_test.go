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

	cfg.Consensus.StartHeight = 0
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
	s.para.grpcClient = s.grpcCli
	s.para.SetQueueClient(q.Client())

	s.mem = mempool.New(cfg.Mempool, nil)
	s.mem.SetQueueClient(q.Client())
	s.mem.Wait()

	s.network = p2p.New(cfg.P2P)
	s.network.SetQueueClient(q.Client())

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

//func newMockParaNode() *testnode.Chain33Mock {
//	//_, sub := testnode.GetDefaultConfig()
//	//cfg.Consensus.Minerstart = false
//	cfg, sub := types.InitCfg("../../../plugin/dapp/paracross/cmd/build/chain33.para.test.toml")
//	cfg.Consensus.StartHeight=0
//	mock33 := testnode.NewWithConfig(cfg, sub, nil)
//	return mock33
//}
//
//func TestSwitchHashMatchedBlock(t *testing.T) {
//	mockPara := newMockParaNode()
//	defer mockPara.Close()
//	mockPara.WaitHeight(0)
//	block := mockPara.GetBlock(0)
//	assert.Equal(t, block.Height, int64(0))
//
//	//consens:=mockPara.GetCfg().Consensus
//	//
//	//paraCli := New(mockPara.GetCfg().Consensus,nil).(*client)
//	////paraCli.BaseClient.SetQueueClient(mock33.GetClient())
//	//paraCli.SetQueueClient(mockPara.GetClient())
//	var  currSeq int64
//	var preMainBlockHash []byte
//	currSeq=2
//	cs := mockPara.GetConsensClient().(*client)
//	cs.switchHashMatchedBlock(&currSeq,&preMainBlockHash)
//	assert.Equal(t,1,currSeq)
//}
