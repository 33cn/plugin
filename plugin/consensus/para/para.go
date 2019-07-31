// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"encoding/hex"
	"fmt"
	"sync"

	log "github.com/33cn/chain33/common/log/log15"

	"sync/atomic"

	"time"

	"github.com/33cn/chain33/client/api"
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
	batchFetchSeqEnable             bool
	batchFetchSeqNum                int64 = 128
)

func init() {
	drivers.Reg("para", New)
	drivers.QueryData.Register("para", &client{})
}

type client struct {
	*drivers.BaseClient
	grpcClient      types.Chain33Client
	execAPI         api.ExecutorAPI
	caughtUp        int32
	commitMsgClient *commitMsgClient
	blockSyncClient *BlockSyncClient
	authAccount     string
	privateKey      crypto.PrivKey
	wg              sync.WaitGroup
	subCfg          *subConfig
	quitCreate      chan struct{}
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
	MaxCacheCount                   int64  `json:"maxCacheCount,omitempty"`
	MaxSyncErrCount                 int32  `json:"maxSyncErrCount,omitempty"`
	BatchFetchSeqEnable             uint32 `json:"batchFetchSeqEnable,omitempty"`
	BatchFetchSeqNum                int64  `json:"batchFetchSeqNum,omitempty"`
	ParaConsensStartHeight          int64  `json:"paraConsensStartHeight,omitempty"`
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

	if subcfg.BatchFetchSeqEnable > 0 {
		batchFetchSeqEnable = true
	}

	if subcfg.BatchFetchSeqNum > 0 {
		batchFetchSeqNum = subcfg.BatchFetchSeqNum
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
		quitCreate:  make(chan struct{}),
	}

	waitBlocks := int32(3) //缺省是3
	if subcfg.WaitBlocks4CommitMsg > 0 {
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
		consensHeight:        -2,
		sendingHeight:        -1,
		consensStartHeight:   -1,
		resetCh:              make(chan interface{}, 1),
		quit:                 make(chan struct{}),
	}
	//note：只有在主链LoopCheckCommitTxDoneForkHeight之后才支持设置ParaConsensStartHeight
	if subcfg.ParaConsensStartHeight > 0 {
		para.commitMsgClient.consensStartHeight = subcfg.ParaConsensStartHeight - 1
	}

	para.blockSyncClient = &BlockSyncClient{
		paraClient:       para,
		notifyChan:       make(chan bool),
		quitChan:         make(chan struct{}),
		maxCacheCount:    1000,
		maxSyncErrCount:  100,
		isPrintDebugInfo: false,
	}
	if subcfg.MaxCacheCount > 0 {
		para.blockSyncClient.maxCacheCount = subcfg.MaxCacheCount
	}
	if subcfg.MaxSyncErrCount > 0 {
		para.blockSyncClient.maxSyncErrCount = subcfg.MaxSyncErrCount
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
	close(client.quitCreate)
	close(client.blockSyncClient.quitChan)
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
	client.wg.Add(1)
	go client.CreateBlock()
	client.wg.Add(1)
	go client.blockSyncClient.SyncBlocks()
}

func (client *client) InitBlock() {
	var err error

	client.execAPI = api.New(client.BaseClient.GetAPI(), client.grpcClient)

	block, err := client.RequestLastBlock()
	if err != nil {
		panic(err)
	}

	if block == nil {
		if startHeight <= 0 {
			panic(fmt.Sprintf("startHeight(%d) should be more than 0 in mainchain", startHeight))
		}
		//平行链创世区块对应主链hash为startHeight-1的那个block的hash
		mainHash := client.GetStartMainHash(startHeight - 1)
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
		err := client.blockSyncClient.CreateGenesisBlock(newblock)
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

// GetStartMainHash 获取start
func (client *client) GetStartMainHash(height int64) []byte {
	lastHeight, err := client.GetLastHeightOnMainChain()
	if err != nil {
		panic(err)
	}
	if lastHeight < height {
		panic(fmt.Sprintf("lastHeight(%d) less than startHeight(%d) in mainchain", lastHeight, height))
	}

	if height > 0 {
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
	}

	hash, err := client.GetHashByHeightOnMainChain(height)
	if err != nil {
		panic(err)
	}
	plog.Info("the start hash in mainchain", "height", height, "hash", hex.EncodeToString(hash))
	return hash
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

func (client *client) isCaughtUp() bool {
	return atomic.LoadInt32(&client.caughtUp) == 1
}

//IsCaughtUp 是否追上最新高度,
func (client *client) Query_IsCaughtUp(req *types.ReqNil) (types.Message, error) {
	if client == nil {
		return nil, fmt.Errorf("%s", "client not bind message queue.")
	}

	return &types.IsCaughtUp{Iscaughtup: client.isCaughtUp()}, nil
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
