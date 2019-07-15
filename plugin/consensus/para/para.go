// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"encoding/hex"

	log "github.com/33cn/chain33/common/log/log15"

	"github.com/33cn/chain33/client/api"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc/grpcclient"
	drivers "github.com/33cn/chain33/system/consensus"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	paracross "github.com/33cn/plugin/plugin/dapp/paracross/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

const (
	addAct int64 = 1 //add para block action
	delAct int64 = 2 //reference blockstore.go, del para block action

	minBlockNum = 100 //min block number startHeight before lastHeight in mainchain
)

var (
	plog                     = log.New("module", "para")
	grpcSite                 = "localhost:8802"
	genesisBlockTime   int64 = 1514533390
	startHeight        int64     //parachain sync from startHeight in mainchain
	blockSec           int64 = 5 //write block interval, second
	emptyBlockInterval int64 = 4 //write empty block every interval blocks in mainchain
	zeroHash           [32]byte
	//current miner tx take any privatekey for unify all nodes sign purpose, and para chain is free
	minerPrivateKey                       = "6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
	searchHashMatchDepth            int32 = 100
	mainBlockHashForkHeight         int64 = 209186          //calc block hash fork height in main chain
	mainParaSelfConsensusForkHeight int64 = types.MaxHeight //para chain self consensus height switch, must >= ForkParacrossCommitTx of main
	mainForkParacrossCommitTx       int64 = types.MaxHeight //support paracross commit tx fork height in main chain: ForkParacrossCommitTx
)

func init() {
	drivers.Reg("para", New)
	drivers.QueryData.Register("para", &client{})
}

type client struct {
	*drivers.BaseClient
	grpcClient      types.Chain33Client
	execAPI         api.ExecutorAPI
	isCaughtUp      bool
	commitMsgClient *commitMsgClient
	authAccount     string
	privateKey      crypto.PrivKey
	wg              sync.WaitGroup
	subCfg          *subConfig
	mtx             sync.Mutex
}

type subConfig struct {
	WriteBlockSeconds               int64  `json:"writeBlockSeconds,omitempty"`
	ParaRemoteGrpcClient            string `json:"paraRemoteGrpcClient,omitempty"`
	StartHeight                     int64  `json:"startHeight,omitempty"`
	EmptyBlockInterval              int64  `json:"emptyBlockInterval,omitempty"`
	AuthAccount                     string `json:"authAccount,omitempty"`
	WaitBlocks4CommitMsg            int32  `json:"waitBlocks4CommitMsg,omitempty"`
	SearchHashMatchedBlockDepth     int32  `json:"searchHashMatchedBlockDepth,omitempty"`
	GenesisAmount                   int64  `json:"genesisAmount,omitempty"`
	MainBlockHashForkHeight         int64  `json:"mainBlockHashForkHeight,omitempty"`
	MainParaSelfConsensusForkHeight int64  `json:"mainParaSelfConsensusForkHeight,omitempty"`
	MainForkParacrossCommitTx       int64  `json:"mainForkParacrossCommitTx,omitempty"`
	WaitConsensStopTimes            uint32 `json:"waitConsensStopTimes,omitempty"`
}

// New function to init paracross env
func New(cfg *types.Consensus, sub []byte) queue.Module {
	c := drivers.NewBaseClient(cfg)
	var subcfg subConfig
	if sub != nil {
		types.MustDecode(sub, &subcfg)
	}
	if subcfg.GenesisAmount <= 0 {
		subcfg.GenesisAmount = 1e8
	}
	if subcfg.ParaRemoteGrpcClient != "" {
		grpcSite = subcfg.ParaRemoteGrpcClient
	}
	if subcfg.StartHeight > 0 {
		startHeight = subcfg.StartHeight
	}
	if subcfg.WriteBlockSeconds > 0 {
		blockSec = subcfg.WriteBlockSeconds
	}
	if subcfg.EmptyBlockInterval > 0 {
		emptyBlockInterval = subcfg.EmptyBlockInterval
	}
	if subcfg.SearchHashMatchedBlockDepth > 0 {
		searchHashMatchDepth = subcfg.SearchHashMatchedBlockDepth
	}
	if subcfg.MainBlockHashForkHeight > 0 {
		mainBlockHashForkHeight = subcfg.MainBlockHashForkHeight
	}

	if subcfg.MainParaSelfConsensusForkHeight > 0 {
		mainParaSelfConsensusForkHeight = subcfg.MainParaSelfConsensusForkHeight
	}

	if subcfg.MainForkParacrossCommitTx > 0 {
		mainForkParacrossCommitTx = subcfg.MainForkParacrossCommitTx
	}

	pk, err := hex.DecodeString(minerPrivateKey)
	if err != nil {
		panic(err)
	}
	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		panic(err)
	}
	priKey, err := secp.PrivKeyFromBytes(pk)
	if err != nil {
		panic(err)
	}

	grpcCli, err := grpcclient.NewMainChainClient("")
	if err != nil {
		panic(err)
	}

	para := &client{
		BaseClient:  c,
		grpcClient:  grpcCli,
		authAccount: subcfg.AuthAccount,
		privateKey:  priKey,
		subCfg:      &subcfg,
	}

	waitBlocks := int32(2) //最小是2
	if subcfg.WaitBlocks4CommitMsg > 0 {
		if subcfg.WaitBlocks4CommitMsg < waitBlocks {
			panic("config WaitBlocks4CommitMsg should not less 2")
		}
		waitBlocks = subcfg.WaitBlocks4CommitMsg
	}

	waitConsensTimes := uint32(30) //30*10s = 5min
	if subcfg.WaitConsensStopTimes > 0 {
		waitConsensTimes = subcfg.WaitConsensStopTimes
	}

	para.commitMsgClient = &commitMsgClient{
		paraClient:           para,
		waitMainBlocks:       waitBlocks,
		waitConsensStopTimes: waitConsensTimes,
		commitCh:             make(chan int64, 1),
		resetCh:              make(chan int64, 1),
		consensHeight:        -2,
		sendingHeight:        -1,
		quit:                 make(chan struct{}),
	}
	c.SetChild(para)
	return para
}

//para 不检查任何的交易
func (client *client) CheckBlock(parent *types.Block, current *types.BlockDetail) error {
	err := checkMinerTx(current)
	return err
}

func (client *client) Close() {
	client.BaseClient.Close()
	close(client.commitMsgClient.quit)
	client.wg.Wait()
	plog.Info("consensus para closed")
}

func (client *client) SetQueueClient(c queue.Client) {
	plog.Info("Enter SetQueueClient method of Para consensus")
	client.InitClient(c, func() {
		client.InitBlock()
	})
	go client.EventLoop()
	client.wg.Add(1)
	go client.commitMsgClient.handler()
	go client.CreateBlock()
}

func (client *client) InitBlock() {
	var err error

	client.execAPI = api.New(client.BaseClient.GetAPI(), client.grpcClient)

	block, err := client.RequestLastBlock()
	if err != nil {
		panic(err)
	}

	if block == nil {
		startSeq, mainHash := client.GetStartSeq(startHeight)
		// 创世区块
		newblock := &types.Block{}
		newblock.Height = 0
		newblock.BlockTime = genesisBlockTime
		newblock.ParentHash = zeroHash[:]
		newblock.MainHash = mainHash
		newblock.MainHeight = startHeight - 1
		tx := client.CreateGenesisTx()
		newblock.Txs = tx
		newblock.TxHash = merkle.CalcMerkleRoot(newblock.Txs)
		err := client.WriteBlock(zeroHash[:], newblock, startSeq)
		if err != nil {
			panic(fmt.Sprintf("para chain create genesis block,err=%s", err.Error()))
		}
		err = client.createLocalGenesisBlock(newblock)
		if err != nil {
			panic(fmt.Sprintf("para chain create local genesis block,err=%s", err.Error()))
		}

	} else {
		client.SetCurrentBlock(block)
	}

	plog.Debug("para consensus init parameter", "mainBlockHashForkHeight", mainBlockHashForkHeight)
	plog.Debug("para consensus init parameter", "mainParaSelfConsensusForkHeight", mainParaSelfConsensusForkHeight)

}

// GetStartSeq get startSeq in mainchain
func (client *client) GetStartSeq(height int64) (int64, []byte) {
	if height <= 0 {
		panic(fmt.Sprintf("startHeight(%d) should be more than 0 in mainchain", height))
	}

	lastHeight, err := client.GetLastHeightOnMainChain()
	if err != nil {
		panic(err)
	}
	if lastHeight < height && lastHeight > 0 {
		panic(fmt.Sprintf("lastHeight(%d) less than startHeight(%d) in mainchain", lastHeight, height))
	}

	hint := time.NewTicker(5 * time.Second)
	for lastHeight < height+minBlockNum {
		select {
		case <-hint.C:
			plog.Info("Waiting lastHeight increase......", "lastHeight", lastHeight, "startHeight", height)
		default:
			lastHeight, err = client.GetLastHeightOnMainChain()
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Second)
		}
	}
	hint.Stop()
	plog.Info(fmt.Sprintf("lastHeight more than %d blocks after startHeight", minBlockNum), "lastHeight", lastHeight, "startHeight", height)

	seq, hash, err := client.GetSeqByHeightOnMainChain(height - 1)
	if err != nil {
		panic(err)
	}
	plog.Info("the start sequence in mainchain", "startHeight", height, "startSeq", seq)
	return seq, hash
}

func (client *client) CreateGenesisTx() (ret []*types.Transaction) {
	var tx types.Transaction
	tx.Execer = []byte(types.ExecName(cty.CoinsX))
	tx.To = client.Cfg.Genesis
	//gen payload
	g := &cty.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{}
	g.Genesis.Amount = client.subCfg.GenesisAmount * types.Coin
	tx.Payload = types.Encode(&cty.CoinsAction{Value: g, Ty: cty.CoinsActionGenesis})
	ret = append(ret, &tx)
	return
}

func (client *client) ProcEvent(msg *queue.Message) bool {
	return false
}

func (client *client) removeBlocks(endHeight int64) error {
	for {
		lastBlock, err := client.RequestLastBlock()
		if err != nil {
			plog.Error("Parachain RequestLastBlock fail", "err", err)
			return err
		}
		if lastBlock.Height == endHeight {
			return nil
		}

		blockedSeq, err := client.GetBlockedSeq(lastBlock.Hash())
		if err != nil {
			plog.Error("Parachain GetBlockedSeq fail", "err", err)
			return err
		}

		err = client.DelBlock(lastBlock, blockedSeq)
		if err != nil {
			plog.Error("Parachain GetBlockedSeq fail", "err", err)
			return err
		}
		plog.Info("Parachain removeBlocks succ", "localParaHeight", lastBlock.Height, "blockedSeq", blockedSeq)
	}
}

// miner tx need all para node create, but not all node has auth account, here just not sign to keep align
func (client *client) addMinerTx(preStateHash []byte, block *types.Block, main *types.BlockSeq, txs []*types.Transaction) error {
	status := &pt.ParacrossNodeStatus{
		Title:           types.GetTitle(),
		Height:          block.Height,
		MainBlockHash:   main.Seq.Hash,
		MainBlockHeight: main.Detail.Block.Height,
	}

	if !paracross.IsParaForkHeight(status.MainBlockHeight, pt.ForkLoopCheckCommitTxDone) {
		status.PreBlockHash = block.ParentHash
		status.PreStateHash = preStateHash

	}
	tx, err := pt.CreateRawMinerTx(&pt.ParacrossMinerAction{
		Status:          status,
		IsSelfConsensus: isParaSelfConsensusForked(status.MainBlockHeight),
	})
	if err != nil {
		return err
	}
	tx.Sign(types.SECP256K1, client.privateKey)
	block.Txs = append([]*types.Transaction{tx}, block.Txs...)
	return nil
}

func (client *client) createBlock(lastBlock *types.Block, txs []*types.Transaction, seq int64, mainBlock *types.BlockSeq) error {
	var newblock types.Block
	plog.Debug(fmt.Sprintf("the len txs is: %v", len(txs)))

	newblock.ParentHash = lastBlock.Hash()
	newblock.Height = lastBlock.Height + 1
	newblock.Txs = txs
	err := client.addMinerTx(lastBlock.StateHash, &newblock, mainBlock, txs)
	if err != nil {
		return err
	}
	//挖矿固定难度
	newblock.Difficulty = types.GetP(0).PowLimitBits
	newblock.TxHash = merkle.CalcMerkleRoot(newblock.Txs)
	newblock.BlockTime = mainBlock.Detail.Block.BlockTime
	newblock.MainHash = mainBlock.Seq.Hash
	newblock.MainHeight = mainBlock.Detail.Block.Height

	err = client.WriteBlock(lastBlock.StateHash, &newblock, seq)

	plog.Debug("para create new Block", "newblock.ParentHash", common.ToHex(newblock.ParentHash),
		"newblock.Height", newblock.Height, "newblock.TxHash", common.ToHex(newblock.TxHash),
		"newblock.BlockTime", newblock.BlockTime, "sequence", seq)
	return err
}

func (client *client) createBlockTemp(txs []*types.Transaction, mainBlock *types.BlockSeq) error {
	lastBlock, err := client.RequestLastBlock()
	if err != nil {
		plog.Error("Parachain RequestLastBlock fail", "err", err)
		return err
	}
	return client.createBlock(lastBlock, txs, 0, mainBlock)

}

// 向blockchain写区块
func (client *client) WriteBlock(prev []byte, paraBlock *types.Block, seq int64) error {
	//共识模块不执行block，统一由blockchain模块执行block并做去重的处理，返回执行后的blockdetail
	blockDetail := &types.BlockDetail{Block: paraBlock}

	parablockDetail := &types.ParaChainBlockDetail{Blockdetail: blockDetail, Sequence: seq}
	msg := client.GetQueueClient().NewMessage("blockchain", types.EventAddParaChainBlockDetail, parablockDetail)
	err := client.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}
	blkdetail := resp.GetData().(*types.BlockDetail)
	if blkdetail == nil {
		return errors.New("block detail is nil")
	}

	client.SetCurrentBlock(blkdetail.Block)

	if client.authAccount != "" {
		client.commitMsgClient.updateChainHeight(blockDetail.Block.Height, false)
	}

	return nil
}

// 向blockchain删区块
func (client *client) DelBlock(block *types.Block, seq int64) error {
	plog.Debug("delete block in parachain")
	start := block.Height
	if start == 0 {
		panic("Parachain attempt to Delete GenesisBlock !")
	}
	msg := client.GetQueueClient().NewMessage("blockchain", types.EventGetBlocks, &types.ReqBlocks{Start: start, End: start, IsDetail: true, Pid: []string{""}})
	err := client.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}
	blocks := resp.GetData().(*types.BlockDetails)

	parablockDetail := &types.ParaChainBlockDetail{Blockdetail: blocks.Items[0], Sequence: seq}
	msg = client.GetQueueClient().NewMessage("blockchain", types.EventDelParaChainBlockDetail, parablockDetail)
	err = client.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}
	resp, err = client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}

	if resp.GetData().(*types.Reply).IsOk {
		if client.authAccount != "" {
			client.commitMsgClient.updateChainHeight(blocks.Items[0].Block.Height, true)

		}
	} else {
		reply := resp.GetData().(*types.Reply)
		return errors.New(string(reply.GetMsg()))
	}
	return nil
}

//IsCaughtUp 是否追上最新高度,
func (client *client) Query_IsCaughtUp(req *types.ReqNil) (types.Message, error) {
	if client == nil {
		return nil, fmt.Errorf("%s", "client not bind message queue.")
	}
	client.mtx.Lock()
	caughtUp := client.isCaughtUp
	client.mtx.Unlock()

	return &types.IsCaughtUp{Iscaughtup: caughtUp}, nil
}

func checkMinerTx(current *types.BlockDetail) error {
	//检查第一个笔交易的execs, 以及执行状态
	if len(current.Block.Txs) == 0 {
		return types.ErrEmptyTx
	}
	baseTx := current.Block.Txs[0]
	//判断交易类型和执行情况
	var action paracross.ParacrossAction
	err := types.Decode(baseTx.GetPayload(), &action)
	if err != nil {
		return err
	}
	if action.GetTy() != pt.ParacrossActionMiner {
		return paracross.ErrParaMinerTxType
	}
	//判断交易执行是否OK
	if action.GetMiner() == nil {
		return paracross.ErrParaEmptyMinerTx
	}

	//判断exec 是否成功
	if current.Receipts[0].Ty != types.ExecOk {
		return paracross.ErrParaMinerExecErr
	}
	return nil
}

// Query_CreateNewAccount 通知para共识模块钱包创建了一个新的账户
func (client *client) Query_CreateNewAccount(acc *types.Account) (types.Message, error) {
	if acc == nil {
		return nil, types.ErrInvalidParam
	}
	plog.Info("Query_CreateNewAccount", "acc", acc.Addr)
	// 需要para共识这边处理新创建的账户是否是超级节点发送commit共识交易的账户
	client.commitMsgClient.onWalletAccount(acc)
	return &types.Reply{IsOk: true, Msg: []byte("OK")}, nil
}

// Query_WalletStatus 通知para共识模块钱包锁状态有变化
func (client *client) Query_WalletStatus(walletStatus *types.WalletStatus) (types.Message, error) {
	if walletStatus == nil {
		return nil, types.ErrInvalidParam
	}
	plog.Info("Query_WalletStatus", "walletStatus", walletStatus.IsWalletLock)
	// 需要para共识这边根据walletStatus.IsWalletLock锁的状态开启/关闭发送共识交易
	client.commitMsgClient.onWalletStatus(walletStatus)
	return &types.Reply{IsOk: true, Msg: []byte("OK")}, nil
}
