// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"math/rand"
	"testing"
	"time"

	"github.com/33cn/chain33/blockchain"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/mempool"
	"github.com/33cn/chain33/p2p"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	//_ "github.com/33cn/plugin/plugin/dapp/paracross"
	pp "github.com/33cn/plugin/plugin/dapp/paracross/executor"
	//"github.com/33cn/plugin/plugin/dapp/paracross/rpc"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/store"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	typesmocks "github.com/33cn/chain33/types/mocks"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

var random *rand.Rand

func init() {
	types.Init("user.p.para.", nil)
	pp.Init("paracross", nil)
	random = rand.New(rand.NewSource(types.Now().UnixNano()))
	consensusInterval = 2
	log.SetLogLevel("debug")
}

type suiteParaCommitMsg struct {
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

func initConfigFile() (*types.Config, *types.ConfigSubModule) {
	cfg, sub := types.InitCfg("../../../plugin/dapp/paracross/cmd/build/chain33.para.test.toml")
	return cfg, sub
}

func (s *suiteParaCommitMsg) initEnv(cfg *types.Config, sub *types.ConfigSubModule) {
	q := queue.New("channel")
	s.q = q
	//api, _ = client.New(q.Client(), nil)

	s.block = blockchain.New(cfg.BlockChain)
	s.block.SetQueueClient(q.Client())

	s.exec = executor.New(cfg.Exec, sub.Exec)
	s.exec.SetQueueClient(q.Client())

	s.store = store.New(cfg.Store, sub.Store)
	s.store.SetQueueClient(q.Client())
	s.para = New(cfg.Consensus, sub.Consensus["para"]).(*client)
	s.grpcCli = &typesmocks.Chain33Client{}
	blockHash := &types.BlockSequence{
		Hash: []byte("1"),
		Type: 1,
	}
	blockSeqs := &types.BlockSequences{Items: []*types.BlockSequence{blockHash}}

	s.grpcCli.On("GetBlockSequences", mock.Anything, mock.Anything).Return(blockSeqs, errors.New("nil"))
	block := &types.Block{}
	blockDetail := &types.BlockDetail{Block: block}
	blockDetails := &types.BlockDetails{Items: []*types.BlockDetail{blockDetail}}
	s.grpcCli.On("GetBlockByHashes", mock.Anything, mock.Anything).Return(blockDetails, errors.New("nil"))
	//data := &types.Int64{1}
	s.grpcCli.On("GetLastBlockSequence", mock.Anything, mock.Anything).Return(nil, errors.New("nil"))
	reply := &types.Reply{IsOk: true}
	s.grpcCli.On("IsSync", mock.Anything, mock.Anything).Return(reply, nil)
	result := &pt.ParacrossStatus{Height: -1}
	data := types.Encode(result)
	ret := &types.Reply{IsOk: true, Msg: data}
	s.grpcCli.On("QueryChain", mock.Anything, mock.Anything).Return(ret, nil).Maybe()
	s.grpcCli.On("SendTransaction", mock.Anything, mock.Anything).Return(reply, nil).Maybe()
	s.grpcCli.On("GetLastHeader", mock.Anything, mock.Anything).Return(&types.Header{Height:cfg.Consensus.StartHeight+minBlockNum}, nil).Maybe()
	s.grpcCli.On("GetBlockHash", mock.Anything, mock.Anything).Return(&types.ReplyHash{Hash:[]byte("1")}, nil).Maybe()
	s.grpcCli.On("GetSequenceByHash", mock.Anything, mock.Anything).Return(&types.Int64{Data:cfg.Consensus.StartHeight}, nil).Maybe()
	s.para.grpcClient = s.grpcCli
	s.para.SetQueueClient(q.Client())

	s.mem = mempool.New(cfg.Mempool, nil)
	s.mem.SetQueueClient(q.Client())
	s.mem.Wait()

	s.network = p2p.New(cfg.P2P)
	s.network.SetQueueClient(q.Client())

	s.para.wg.Add(1)
	go walletProcess(q, s.para)
}

func walletProcess(q queue.Queue, para *client) {
	defer para.wg.Done()

	client := q.Client()
	client.Sub("wallet")

	for {
		select {
		case <-para.commitMsgClient.quit:
			return
		case msg := <-client.Recv():
			if msg.Ty == types.EventDumpPrivkey {
				msg.Reply(client.NewMessage("", types.EventHeader, &types.ReplyString{Data: "6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"}))
			}
		}
	}

}

func (s *suiteParaCommitMsg) SetupSuite() {
	s.initEnv(initConfigFile())
}

func (s *suiteParaCommitMsg) TestRun_1() {
	//s.testGetBlock()
	lastBlock, err := s.para.RequestLastBlock()
	if err != nil {
		plog.Error("para test", "err", err.Error())
	}
	plog.Info("para test---------", "last height", lastBlock.Height)
	s.para.createBlock(lastBlock, nil, 0, getMainBlock(1, lastBlock.BlockTime+1))
	lastBlock, err = s.para.RequestLastBlock()
	if err != nil {
		plog.Error("para test--2", "err", err.Error())
	}
	plog.Info("para test---------", "last height", lastBlock.Height)
	s.para.createBlock(lastBlock, nil, 1, getMainBlock(2, lastBlock.BlockTime+1))
	time.Sleep(time.Second * 3)
	lastBlock, err = s.para.RequestLastBlock()
	s.para.DelBlock(lastBlock, 2)
	time.Sleep(time.Second * 3)
}

func TestRunSuiteParaCommitMsg(t *testing.T) {
	log := new(suiteParaCommitMsg)
	suite.Run(t, log)
}

func (s *suiteParaCommitMsg) TearDownSuite() {
	time.Sleep(time.Second * 5)
	s.block.Close()
	s.para.Close()
	s.exec.Close()
	s.store.Close()
	s.mem.Close()
	s.network.Close()
	s.q.Close()

}

func getMainBlock(height int64, BlockTime int64) *types.Block {
	return &types.Block{
		Height:    height,
		BlockTime: BlockTime,
	}
}
