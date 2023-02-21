// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"encoding/hex"
	"fmt"
	"sort"
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
	defaultWaitMinBlockNum = 100 //min block number startHeight before lastHeight in mainchain

	defaultGenesisBlockTime int64 = 1514533390
	//current miner tx take any privatekey for unify all nodes sign purpose, and para chain is free
	minerPrivateKey                      = "6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
	poolMainBlockMsec              int64 = 5000
	defaultEmptyBlockInterval      int64 = 50 //write empty block every interval blocks in mainchain
	defaultSearchMatchedBlockDepth int32 = 10000
)

var (
	plog     = log.New("module", "para")
	zeroHash [32]byte
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
	blockSyncClient *blockSyncClient
	multiDldCli     *multiDldClient
	jumpDldCli      *jumpDldClient
	minerPrivateKey crypto.PrivKey
	wg              sync.WaitGroup
	cfg             *types.Consensus
	subCfg          *subConfig
	dldCfg          *downloadClient
	blsSignCli      *blsClient
	isClosed        int32
	quit            chan struct{}
}

type subConfig struct {
	WriteBlockMsec          int64      `json:"writeBlockMsec,omitempty"`
	StartHeight             int64      `json:"startHeight,omitempty"`
	WaitMainBlockNum        int64      `json:"waitMainBlockNum,omitempty"`
	GenesisStartHeightSame  bool       `json:"genesisStartHeightSame,omitempty"`
	EmptyBlockInterval      []string   `json:"emptyBlockInterval,omitempty"`
	AuthAccount             string     `json:"authAccount,omitempty"`
	WaitBlocks4CommitMsg    int32      `json:"waitBlocks4CommitMsg,omitempty"`
	GenesisAmount           int64      `json:"genesisAmount,omitempty"`
	MainBlockHashForkHeight int64      `json:"mainBlockHashForkHeight,omitempty"`
	WaitConsensStopTimes    uint32     `json:"waitConsensStopTimes,omitempty"`
	MaxCacheCount           int64      `json:"maxCacheCount,omitempty"`
	MaxSyncErrCount         int32      `json:"maxSyncErrCount,omitempty"`
	BatchFetchBlockCount    int64      `json:"batchFetchBlockCount,omitempty"`
	ParaConsensStartHeight  int64      `json:"paraConsensStartHeight,omitempty"`
	MultiDownloadOpen       bool       `json:"multiDownloadOpen,omitempty"`
	MultiDownInvNumPerJob   int64      `json:"multiDownInvNumPerJob,omitempty"`
	MultiDownJobBuffNum     uint32     `json:"multiDownJobBuffNum,omitempty"`
	MultiDownServerRspTime  uint32     `json:"multiDownServerRspTime,omitempty"`
	RmCommitParamMainHeight int64      `json:"rmCommitParamMainHeight,omitempty"`
	JumpDownloadClose       bool       `json:"jumpDownloadClose,omitempty"`
	Bls                     *blsConfig `json:"bls,omitempty"`
}

// New function to init paracross env
func New(cfg *types.Consensus, sub []byte) queue.Module {
	c := drivers.NewBaseClient(cfg)
	var subcfg subConfig
	if sub != nil {
		types.MustDecode(sub, &subcfg)
	}
	//支持创世精度为0
	if subcfg.GenesisAmount < 0 {
		panic(fmt.Sprintf("genesis amount <0"))
	}

	if subcfg.WriteBlockMsec <= 0 {
		subcfg.WriteBlockMsec = poolMainBlockMsec
	}

	//WaitMainBlockNum 配置最小为1，因为genesis块是startHeight-1， wait=1和startHeight相等
	if subcfg.WaitMainBlockNum <= 0 {
		subcfg.WaitMainBlockNum = defaultWaitMinBlockNum
	}

	//最初平行链toml GenesisBlockTime=1514533394，但是未被使用，一直使用的内置的1514533390,最新版本开始适配cfg.GenesisBlockTime,并且
	//时间也缺省改为1514533390，支持修改时间， 如果有以前的旧的配置未修改，panic强制修改
	if cfg.GenesisBlockTime == 1514533394 {
		panic("para chain GenesisBlockTime need be modified to 1514533390 or other")
	}

	emptyInterval, err := parseEmptyBlockInterval(subcfg.EmptyBlockInterval)
	if err != nil {
		panic("para EmptyBlockInterval config not correct")
	}
	err = checkEmptyBlockInterval(emptyInterval)
	if err != nil {
		panic("para EmptyBlockInterval config not correct")
	}

	if subcfg.BatchFetchBlockCount <= 0 {
		subcfg.BatchFetchBlockCount = types.MaxBlockCountPerTime
	}
	if subcfg.BatchFetchBlockCount > types.MaxBlockCountPerTime {
		panic(fmt.Sprintf("BatchFetchBlockCount=%d should be <= %d ", subcfg.BatchFetchBlockCount, types.MaxBlockCountPerTime))
	}

	pk, err := hex.DecodeString(minerPrivateKey)
	if err != nil {
		panic(err)
	}
	secp, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		panic(err)
	}
	priKey, err := secp.PrivKeyFromBytes(pk)
	if err != nil {
		panic(err)
	}

	para := &client{
		BaseClient:      c,
		minerPrivateKey: priKey,
		cfg:             cfg,
		subCfg:          &subcfg,
		quit:            make(chan struct{}),
	}

	para.dldCfg = &downloadClient{}
	para.dldCfg.emptyInterval = append(para.dldCfg.emptyInterval, emptyInterval...)

	para.commitMsgClient = newCommitMsgCli(para, &subcfg)
	para.blockSyncClient = newBlockSyncCli(para, &subcfg)

	para.multiDldCli = newMultiDldCli(para, &subcfg)

	para.jumpDldCli = newJumpDldCli(para, &subcfg)

	para.blsSignCli = newBlsClient(para, &subcfg)

	c.SetChild(para)
	return para
}

//para 不检查任何的交易
func (client *client) CheckBlock(parent *types.Block, current *types.BlockDetail) error {
	err := checkMinerTx(current)
	return err
}

func (client *client) Close() {
	atomic.StoreInt32(&client.isClosed, 1)
	close(client.commitMsgClient.quit)
	close(client.quit)
	close(client.blockSyncClient.quitChan)
	close(client.blsSignCli.quit)

	client.wg.Wait()

	client.BaseClient.Close()

	plog.Info("consensus para closed")
}

func (client *client) isCancel() bool {
	return atomic.LoadInt32(&client.isClosed) == 1
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
	go client.blockSyncClient.syncBlocks()

	client.wg.Add(2)
	go client.blsSignCli.procAggregateTxs()
	go client.blsSignCli.procLeaderSync()

}

func (client *client) InitBlock() {
	var err error

	client.execAPI = api.New(client.BaseClient.GetAPI(), client.grpcClient)
	cfg := client.GetAPI().GetConfig()
	grpcCli, err := grpcclient.NewMainChainClient(cfg, "")
	if err != nil {
		panic(err)
	}
	client.grpcClient = grpcCli

	err = client.commitMsgClient.setSelfConsEnable()
	if err != nil {
		panic(err)
	}

	block, err := client.RequestLastBlock()
	if err != nil {
		panic(err)
	}

	if block == nil {
		if client.subCfg.StartHeight <= 0 {
			panic(fmt.Sprintf("startHeight(%d) should be more than 0 in mainchain", client.subCfg.StartHeight))
		}
		//平行链创世区块对应主链hash为startHeight-1的那个block的hash
		mainHash := client.GetStartMainHash(client.subCfg.StartHeight - 1)
		// 创世区块
		newblock := &types.Block{}
		newblock.Height = 0
		newblock.BlockTime = defaultGenesisBlockTime
		if client.cfg.GenesisBlockTime > 0 {
			newblock.BlockTime = client.cfg.GenesisBlockTime
		}

		newblock.ParentHash = zeroHash[:]
		newblock.MainHash = mainHash

		//缺省是减1,但有些特殊项目方6.2.0版本升级上来要求blockhash不变，则需与6.2.0保持一致，不减一
		newblock.MainHeight = client.subCfg.StartHeight - 1
		if client.subCfg.GenesisStartHeightSame {
			newblock.MainHeight = client.subCfg.StartHeight
		}
		tx := client.CreateGenesisTx()
		newblock.Txs = tx
		newblock.TxHash = merkle.CalcMerkleRoot(cfg, newblock.GetMainHeight(), newblock.Txs)
		err := client.blockSyncClient.createGenesisBlock(newblock)
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

	plog.Debug("para consensus init parameter", "mainBlockHashForkHeight", client.subCfg.MainBlockHashForkHeight)

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
		hint := time.NewTicker(time.Second)
		for lastHeight < height+client.subCfg.WaitMainBlockNum {
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
		plog.Info(fmt.Sprintf("lastHeight more than %d blocks after startHeight", client.subCfg.WaitMainBlockNum), "lastHeight", lastHeight, "startHeight", height)
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
	cfg := client.GetAPI().GetConfig()
	tx.Execer = []byte(cfg.ExecName(cfg.GetCoinExec()))
	tx.To = client.Cfg.Genesis
	//gen payload
	g := &cty.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{}
	g.Genesis.Amount = client.subCfg.GenesisAmount * cfg.GetCoinPrecision()
	tx.Payload = types.Encode(&cty.CoinsAction{Value: g, Ty: cty.CoinsActionGenesis})
	ret = append(ret, &tx)
	return
}

func (client *client) ProcEvent(msg *queue.Message) bool {
	if msg.Ty == types.EventReceiveSubData {
		if req, ok := msg.GetData().(*types.TopicData); ok {
			var sub pt.ParaP2PSubMsg
			err := types.Decode(req.Data, &sub)
			if err != nil {
				plog.Error("paracross ProcEvent decode", "ty", types.EventReceiveSubData)
				return true
			}
			plog.Info("paracross ProcEvent", "from", req.GetFrom(), "topic:", req.GetTopic(), "ty", sub.GetTy())
			if !client.blsSignCli.blsSignOn {
				plog.Error("paracross ProcEvent bls is closed")
				return true
			}
			switch sub.GetTy() {
			case P2pSubCommitTx:
				go client.blsSignCli.rcvCommitTx(sub.GetCommitTx())
			case P2pSubLeaderSyncMsg:
				err := client.blsSignCli.rcvLeaderSyncTx(sub.GetSyncMsg())
				if err != nil {
					plog.Error("bls.event.paracross ProcEvent leader sync msg", "err", err)
				}
			default:
				plog.Error("paracross ProcEvent not support", "ty", sub.GetTy())
			}

		} else {
			plog.Error("paracross ProcEvent topicData", "ty", types.EventReceiveSubData)
		}

		return true
	}

	return false
}

func (client *client) isCaughtUp() bool {
	return atomic.LoadInt32(&client.caughtUp) == 1
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
		plog.Error("checkMinerTx", "receiptTy", current.Receipts[0].Ty)
		return paracross.ErrParaMinerExecErr
	}
	return nil
}

//比较newBlock是不是最优区块
func (client *client) CmpBestBlock(newBlock *types.Block, cmpBlock *types.Block) bool {
	return false
}

//["0:50","100:20","500:30"]
func parseEmptyBlockInterval(cfg []string) ([]*emptyBlockInterval, error) {
	var emptyInter []*emptyBlockInterval
	if len(cfg) == 0 {
		interval := &emptyBlockInterval{startHeight: 0, interval: defaultEmptyBlockInterval}
		emptyInter = append(emptyInter, interval)
		return emptyInter, nil
	}

	list := make(map[int64]int64)
	var seq []int64
	for _, e := range cfg {
		ret, err := divideStr2Int64s(e, ":")
		if err != nil {
			plog.Error("parse empty block inter config", "str", e)
			return nil, err
		}
		seq = append(seq, ret[0])
		list[ret[0]] = ret[1]
	}
	sort.Slice(seq, func(i, j int) bool { return seq[i] < seq[j] })
	for _, h := range seq {
		emptyInter = append(emptyInter, &emptyBlockInterval{startHeight: h, interval: list[h]})
	}
	return emptyInter, nil
}

func checkEmptyBlockInterval(in []*emptyBlockInterval) error {
	for i := 0; i < len(in); i++ {
		if i == 0 && in[i].startHeight != 0 {
			plog.Error("EmptyBlockInterval,first blockHeight should be 0", "height", in[i].startHeight)
			return types.ErrInvalidParam
		}
		if i > 0 && in[i].startHeight <= in[i-1].startHeight {
			plog.Error("EmptyBlockInterval,blockHeight should be sequence", "preHeight", in[i-1].startHeight, "laterHeight", in[i].startHeight)
			return types.ErrInvalidParam
		}
		if in[i].interval <= 0 {
			plog.Error("EmptyBlockInterval,interval should > 0", "height", in[i].startHeight)
			return types.ErrInvalidParam
		}
	}
	return nil
}
